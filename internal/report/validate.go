package report

import (
	"fmt"
	"math"
	"strings"
	"time"
)

var validTriggers = map[Trigger]bool{
	TriggerHook:       true,
	TriggerManual:     true,
	TriggerStrengthen: true,
	TriggerCI:         true,
}

var validStatuses = map[MutantStatus]bool{
	StatusKilled:       true,
	StatusSurvived:     true,
	StatusTimeout:      true,
	StatusNoCoverage:   true,
	StatusIgnored:      true,
	StatusCompileError: true,
	StatusRuntimeError: true,
}

// Validate checks a fully-parsed Artifact against the docs/SCHEMA.md
// contract before gate or report may treat it as evidence of anything
// (docs/COORDINATOR_NOTES.md: "a syntactically valid JSON object with a
// fabricated score must not pass the gate"). It never trusts a stored count
// or score at face value: every total and per-file Trust Score is
// recomputed from the raw per-mutant statuses using the same
// killed+timeout-vs-survived formula the CLI itself writes with (score.go),
// and rejected if the stored value disagrees.
//
// Read calls this on every load; gate.Evaluate and the `report` command
// never see an artifact that failed it.
func (a Artifact) Validate() error {
	if a.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schemaVersion %d is not supported (want %d)", a.SchemaVersion, SchemaVersion)
	}
	if err := validateRunID(a.RunID); err != nil {
		return fmt.Errorf("runId: %w", err)
	}
	if _, err := time.Parse(time.RFC3339, a.CreatedAt); err != nil {
		return fmt.Errorf("createdAt %q is not a valid RFC3339 timestamp: %w", a.CreatedAt, err)
	}
	if a.BaseRef == "" {
		return fmt.Errorf("baseRef must not be empty")
	}
	if a.HeadSha == "" {
		return fmt.Errorf("headSha must not be empty")
	}
	if !validTriggers[a.Trigger] {
		return fmt.Errorf("trigger %q is not one of hook|manual|strengthen|ci", a.Trigger)
	}
	if math.IsNaN(a.Threshold) || math.IsInf(a.Threshold, 0) || a.Threshold < 0 || a.Threshold > 100 {
		return fmt.Errorf("threshold %v must be a finite number between 0 and 100", a.Threshold)
	}
	if err := validatePercent("totals.lineCoverage", a.Totals.LineCoverage); err != nil {
		return err
	}
	if err := validatePercent("totals.trustScore", a.Totals.TrustScore); err != nil {
		return err
	}
	if a.Totals.DurationMs < 0 {
		return fmt.Errorf("totals.durationMs must not be negative, got %d", a.Totals.DurationMs)
	}
	if a.ComparedTo != nil {
		if err := validateRunID(*a.ComparedTo); err != nil {
			return fmt.Errorf("comparedTo: %w", err)
		}
	}

	changed := make(map[string]bool, len(a.ChangedFiles))
	for _, f := range a.ChangedFiles {
		if err := validateRelPath(f); err != nil {
			return fmt.Errorf("changedFiles: %w", err)
		}
		if changed[f] {
			return fmt.Errorf("changedFiles contains duplicate path %q", f)
		}
		changed[f] = true
	}

	seenFilePaths := make(map[string]bool, len(a.Files))
	for _, f := range a.Files {
		if err := validateRelPath(f.Path); err != nil {
			return fmt.Errorf("files: %w", err)
		}
		if seenFilePaths[f.Path] {
			return fmt.Errorf("files contains duplicate path %q", f.Path)
		}
		seenFilePaths[f.Path] = true
		if !changed[f.Path] {
			return fmt.Errorf("files entry %q is not listed in changedFiles", f.Path)
		}

		seenMutantIDs := make(map[string]bool, len(f.Mutants))
		for _, m := range f.Mutants {
			if m.ID == "" {
				return fmt.Errorf("file %q has a mutant with an empty id", f.Path)
			}
			if seenMutantIDs[m.ID] {
				// Mutant IDs need only be unique within a file (docs/SCHEMA.md §2).
				return fmt.Errorf("file %q has duplicate mutant id %q", f.Path, m.ID)
			}
			seenMutantIDs[m.ID] = true
			if !validStatuses[m.Status] {
				return fmt.Errorf("file %q mutant %q has unknown status %q", f.Path, m.ID, m.Status)
			}
			if m.Line < 0 {
				return fmt.Errorf("file %q mutant %q has negative line %d", f.Path, m.ID, m.Line)
			}
		}

		want := BuildFileResult(f.Path, f.Mutants)
		if !floatPtrEqual(f.TrustScore, want.TrustScore) {
			return fmt.Errorf("file %q trustScore %s does not match %s recomputed from its own mutants",
				f.Path, formatFloatPtr(f.TrustScore), formatFloatPtr(want.TrustScore))
		}
	}

	// Recompute totals purely from the (already individually validated)
	// per-file mutants — never from the stored integer fields — and require
	// an exact match. LineCoverage and DurationMs are externally measured,
	// not derivable from mutant outcomes, so they are passed through and
	// checked only for range above, not recomputed here.
	want := BuildTotals(a.Files, a.Totals.LineCoverage, a.Totals.DurationMs)
	switch {
	case a.Totals.Mutants != want.Mutants:
		return fmt.Errorf("totals.mutants %d does not match %d recomputed from files", a.Totals.Mutants, want.Mutants)
	case a.Totals.Killed != want.Killed:
		return fmt.Errorf("totals.killed %d does not match %d recomputed from files", a.Totals.Killed, want.Killed)
	case a.Totals.Survived != want.Survived:
		return fmt.Errorf("totals.survived %d does not match %d recomputed from files", a.Totals.Survived, want.Survived)
	case a.Totals.Timeout != want.Timeout:
		return fmt.Errorf("totals.timeout %d does not match %d recomputed from files", a.Totals.Timeout, want.Timeout)
	case a.Totals.NoCoverage != want.NoCoverage:
		return fmt.Errorf("totals.noCoverage %d does not match %d recomputed from files", a.Totals.NoCoverage, want.NoCoverage)
	case a.Totals.Ignored != want.Ignored:
		return fmt.Errorf("totals.ignored %d does not match %d recomputed from files", a.Totals.Ignored, want.Ignored)
	case a.Totals.CompileError != want.CompileError:
		return fmt.Errorf("totals.compileError %d does not match %d recomputed from files", a.Totals.CompileError, want.CompileError)
	case a.Totals.RuntimeError != want.RuntimeError:
		return fmt.Errorf("totals.runtimeError %d does not match %d recomputed from files", a.Totals.RuntimeError, want.RuntimeError)
	}
	if !floatPtrEqual(a.Totals.TrustScore, want.TrustScore) {
		return fmt.Errorf("totals.trustScore %s does not match %s recomputed from files (killed+timeout vs survived)",
			formatFloatPtr(a.Totals.TrustScore), formatFloatPtr(want.TrustScore))
	}

	return nil
}

func validatePercent(field string, v *float64) error {
	if v == nil {
		return nil
	}
	if math.IsNaN(*v) || math.IsInf(*v, 0) || *v < 0 || *v > 100 {
		return fmt.Errorf("%s %v must be a finite number between 0 and 100, or null", field, *v)
	}
	return nil
}

func floatPtrEqual(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func formatFloatPtr(v *float64) string {
	if v == nil {
		return "null"
	}
	return fmt.Sprintf("%v", *v)
}

// validateRelPath rejects anything that isn't a plain repo-relative,
// forward-slash path: empty strings, backslashes, absolute paths (leading
// "/" or a Windows drive letter), and any ".." path segment.
func validateRelPath(p string) error {
	if p == "" {
		return fmt.Errorf("path must not be empty")
	}
	if strings.Contains(p, "\\") {
		return fmt.Errorf("path %q must use forward slashes", p)
	}
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("path %q must be repo-relative, not absolute", p)
	}
	if len(p) >= 2 && p[1] == ':' {
		return fmt.Errorf("path %q must be repo-relative, not absolute", p)
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return fmt.Errorf("path %q must not contain \"..\"", p)
		}
	}
	return nil
}
