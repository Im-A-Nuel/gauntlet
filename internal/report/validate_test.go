package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// baseValidArtifact returns a fully valid artifact (same shape as
// validArtifact in artifact_test.go) that each negative test case mutates
// to introduce exactly one defect.
func baseValidArtifact() Artifact {
	files := []FileResult{BuildFileResult("src/pricing.ts", []Mutant{
		{ID: "1", Mutator: "ConditionalExpression", Line: 42, Status: StatusKilled, Original: "a", Mutated: "b"},
		{ID: "2", Mutator: "EqualityOperator", Line: 43, Status: StatusSurvived, Original: "c", Mutated: "d"},
	})}
	return Artifact{
		SchemaVersion: SchemaVersion,
		RunID:         "1758945600000-a1b2c3d",
		CreatedAt:     "2026-09-26T14:00:00Z",
		BaseRef:       "main",
		HeadSha:       "a1b2c3d",
		Trigger:       TriggerManual,
		Threshold:     80,
		ChangedFiles:  []string{"src/pricing.ts"},
		Files:         files,
		Totals:        BuildTotals(files, nil, 0),
	}
}

func TestValidateAcceptsAWellFormedArtifact(t *testing.T) {
	if err := baseValidArtifact().Validate(); err != nil {
		t.Fatalf("Validate on a well-formed artifact = %v, want nil", err)
	}
}

func TestValidateRejectsKnownDefects(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(a *Artifact)
		wantSub string
	}{
		{
			name:    "unsupported schema version",
			mutate:  func(a *Artifact) { a.SchemaVersion = 2 },
			wantSub: "schemaVersion",
		},
		{
			name:    "unknown mutant status",
			mutate:  func(a *Artifact) { a.Files[0].Mutants[0].Status = "haunted" },
			wantSub: "unknown status",
		},
		{
			name:    "invalid createdAt timestamp",
			mutate:  func(a *Artifact) { a.CreatedAt = "not-a-timestamp" },
			wantSub: "RFC3339",
		},
		{
			name:    "createdAt missing timezone",
			mutate:  func(a *Artifact) { a.CreatedAt = "2026-09-26 14:00:00" },
			wantSub: "RFC3339",
		},
		{
			name:    "empty baseRef",
			mutate:  func(a *Artifact) { a.BaseRef = "" },
			wantSub: "baseRef",
		},
		{
			name:    "empty headSha",
			mutate:  func(a *Artifact) { a.HeadSha = "" },
			wantSub: "headSha",
		},
		{
			name:    "unknown trigger",
			mutate:  func(a *Artifact) { a.Trigger = "cron" },
			wantSub: "trigger",
		},
		{
			name:    "threshold below zero",
			mutate:  func(a *Artifact) { a.Threshold = -1 },
			wantSub: "threshold",
		},
		{
			name:    "threshold above 100",
			mutate:  func(a *Artifact) { a.Threshold = 150 },
			wantSub: "threshold",
		},
		{
			name:    "threshold NaN",
			mutate:  func(a *Artifact) { a.Threshold = nan() },
			wantSub: "threshold",
		},
		{
			name:    "lineCoverage above 100",
			mutate:  func(a *Artifact) { v := 250.0; a.Totals.LineCoverage = &v },
			wantSub: "lineCoverage",
		},
		{
			name:    "lineCoverage negative",
			mutate:  func(a *Artifact) { v := -5.0; a.Totals.LineCoverage = &v },
			wantSub: "lineCoverage",
		},
		{
			name:    "negative duration",
			mutate:  func(a *Artifact) { a.Totals.DurationMs = -1 },
			wantSub: "durationMs",
		},
		{
			name:    "forged totals.trustScore",
			mutate:  func(a *Artifact) { v := 99.9; a.Totals.TrustScore = &v },
			wantSub: "trustScore",
		},
		{
			name:    "forged totals.killed count",
			mutate:  func(a *Artifact) { a.Totals.Killed = 999 },
			wantSub: "killed",
		},
		{
			name:    "forged totals.survived count (hides a survivor)",
			mutate:  func(a *Artifact) { a.Totals.Survived = 0 },
			wantSub: "survived",
		},
		{
			name:    "forged per-file trustScore",
			mutate:  func(a *Artifact) { v := 100.0; a.Files[0].TrustScore = &v },
			wantSub: "trustScore",
		},
		{
			name: "changedFiles path traversal",
			mutate: func(a *Artifact) {
				a.ChangedFiles = append(a.ChangedFiles, "../../etc/passwd")
			},
			wantSub: "changedFiles",
		},
		{
			name: "file path not listed in changedFiles",
			mutate: func(a *Artifact) {
				a.ChangedFiles = []string{"src/other.ts"}
			},
			wantSub: "not listed in changedFiles",
		},
		{
			name: "duplicate mutant id within a file",
			mutate: func(a *Artifact) {
				a.Files[0].Mutants = append(a.Files[0].Mutants, Mutant{ID: "1", Status: StatusKilled})
				a.Files[0] = BuildFileResult(a.Files[0].Path, a.Files[0].Mutants)
				// Force the stored score to still match the (still-invalid)
				// mutant list so the duplicate-ID check — not the
				// score-mismatch check — is what fails.
			},
			wantSub: "duplicate mutant id",
		},
		{
			name:    "empty mutant id",
			mutate:  func(a *Artifact) { a.Files[0].Mutants[0].ID = "" },
			wantSub: "empty id",
		},
		{
			name:    "negative mutant line",
			mutate:  func(a *Artifact) { a.Files[0].Mutants[0].Line = -1 },
			wantSub: "negative line",
		},
		{
			name:    "invalid run ID characters (would traverse a path)",
			mutate:  func(a *Artifact) { a.RunID = "../../../etc/passwd" },
			wantSub: "runId",
		},
		{
			name:    "comparedTo path traversal",
			mutate:  func(a *Artifact) { id := "../elsewhere"; a.ComparedTo = &id },
			wantSub: "comparedTo",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := baseValidArtifact()
			c.mutate(&a)
			err := a.Validate()
			if err == nil {
				t.Fatalf("Validate() = nil, want an error containing %q", c.wantSub)
			}
			if !strings.Contains(err.Error(), c.wantSub) {
				t.Fatalf("Validate() = %q, want it to contain %q", err.Error(), c.wantSub)
			}
		})
	}
}

// TestReadRejectsOnDisk exercises the same defects through the real Read
// path (temp dir + Write + hand-edited file), covering the two checks that
// live in Read itself rather than Validate: path-traversal run IDs supplied
// by a caller (e.g. --run-id) and a run ID that disagrees with the
// artifact's own recorded RunID field.
func TestReadRejectsOnDisk(t *testing.T) {
	t.Run("traversal run ID is rejected before touching the filesystem", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := Write(dir, baseValidArtifact()); err != nil {
			t.Fatalf("Write: %v", err)
		}
		_, err := Read(dir, "../../../etc/passwd")
		if err == nil {
			t.Fatalf("Read with a traversal run ID = nil error, want rejection")
		}
	})

	t.Run("run ID with path separator is rejected", func(t *testing.T) {
		dir := t.TempDir()
		_, err := Read(dir, "sub/dir")
		if err == nil {
			t.Fatalf("Read with a path-separator run ID = nil error, want rejection")
		}
	})

	t.Run("mismatched run ID between filename and content is rejected", func(t *testing.T) {
		dir := t.TempDir()
		a := baseValidArtifact() // a.RunID == "1758945600000-a1b2c3d"
		data, err := json.MarshalIndent(a, "", "  ")
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		// Save valid content under a filename that disagrees with the
		// artifact's own declared runId field — e.g. a copy-pasted or
		// hand-renamed artifact file.
		runsDir := RunsDir(dir)
		if err := os.MkdirAll(runsDir, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(runsDir, "different-name.json"), data, 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}

		if _, err := Read(dir, "different-name"); err == nil {
			t.Fatalf("Read(%q) = nil error for content whose runId is %q, want mismatch rejection", "different-name", a.RunID)
		}
	})

	t.Run("unknown run ID is a plain not-found error, not a validation panic", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := Read(dir, "does-not-exist"); err == nil {
			t.Fatalf("Read of a missing run ID = nil error, want a not-found error")
		}
	})
}

func nan() float64 {
	var zero float64
	return zero / zero
}
