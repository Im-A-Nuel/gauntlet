package report

import (
	"testing"
	"time"
)

func TestNewRunIDShortensSha(t *testing.T) {
	now := time.Unix(1758945600, 0)
	got := NewRunID(now, "a1b2c3d4e5f6")
	want := "1758945600000-a1b2c3d"
	if got != want {
		t.Fatalf("NewRunID = %q, want %q", got, want)
	}
}

func TestNewRunIDHandlesNoGit(t *testing.T) {
	got := NewRunID(time.Unix(1, 0), "")
	if got != "1000-nogit" {
		t.Fatalf("NewRunID with empty sha = %q, want %q", got, "1000-nogit")
	}
}

// validArtifact returns a minimal but fully schema-valid, self-consistent
// artifact (real Files/Totals built via the same score.go helpers the CLI
// itself uses), suitable as a base for Write/Read round-trip tests now that
// Read fail-closed validates everything it loads.
func validArtifact(runID string) Artifact {
	files := []FileResult{BuildFileResult("src/pricing.ts", []Mutant{
		{ID: "1", Mutator: "ConditionalExpression", Line: 42, Status: StatusKilled, Original: "a", Mutated: "b"},
	})}
	return Artifact{
		SchemaVersion: SchemaVersion,
		RunID:         runID,
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

func TestWriteReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	a := validArtifact("1758945600000-a1b2c3d")
	path, err := Write(dir, a)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	got, err := Read(dir, a.RunID)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.RunID != a.RunID || got.Totals.TrustScore == nil || *got.Totals.TrustScore != 100 {
		t.Fatalf("round-trip mismatch: %+v (from %s)", got, path)
	}
}

func TestWriteLeavesNoTempFilesBehind(t *testing.T) {
	dir := t.TempDir()
	a := Artifact{RunID: "run-1", SchemaVersion: SchemaVersion}
	if _, err := Write(dir, a); err != nil {
		t.Fatalf("Write: %v", err)
	}
	ids, err := List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ids) != 1 || ids[0] != "run-1" {
		t.Fatalf("List = %v, want exactly [run-1] (no leaked .tmp- files)", ids)
	}
}

func TestLatestPicksMostRecentByRunID(t *testing.T) {
	dir := t.TempDir()
	for _, id := range []string{"100-abc", "300-abc", "200-abc"} {
		if _, err := Write(dir, validArtifact(id)); err != nil {
			t.Fatalf("Write %s: %v", id, err)
		}
	}
	latest, ok, err := Latest(dir)
	if err != nil || !ok {
		t.Fatalf("Latest: ok=%v err=%v", ok, err)
	}
	if latest.RunID != "300-abc" {
		t.Fatalf("Latest.RunID = %q, want 300-abc", latest.RunID)
	}
}

func TestLatestNoRunsIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	_, ok, err := Latest(dir)
	if err != nil {
		t.Fatalf("Latest on empty dir: %v", err)
	}
	if ok {
		t.Fatalf("Latest on empty dir: ok = true, want false")
	}
}
