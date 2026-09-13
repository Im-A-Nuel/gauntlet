// Package report defines the Gauntlet run-artifact schema (docs/SCHEMA.md,
// section 2, including the "MVP implementation clarifications") and the
// Trust Score math derived from it.
package report

// MutantStatus enumerates the statuses Stryker's mutation report can produce.
// killed/survived/timeout are "scored"; noCoverage/ignored/compileError/
// runtimeError are reported but excluded from the Trust Score denominator.
type MutantStatus string

const (
	StatusKilled       MutantStatus = "killed"
	StatusSurvived     MutantStatus = "survived"
	StatusTimeout      MutantStatus = "timeout"
	StatusNoCoverage   MutantStatus = "noCoverage"
	StatusIgnored      MutantStatus = "ignored"
	StatusCompileError MutantStatus = "compileError"
	StatusRuntimeError MutantStatus = "runtimeError"
)

// Trigger enumerates what caused a run.
type Trigger string

const (
	TriggerHook       Trigger = "hook"
	TriggerManual     Trigger = "manual"
	TriggerStrengthen Trigger = "strengthen"
	TriggerCI         Trigger = "ci"
)

// SchemaVersion identifies the run-artifact contract (docs/SCHEMA.md §2).
const SchemaVersion = 1

// Mutant is one mutation applied to one file.
type Mutant struct {
	ID       string       `json:"id"`
	Mutator  string       `json:"mutator"`
	Line     int          `json:"line"`
	Status   MutantStatus `json:"status"`
	Original string       `json:"original"`
	Mutated  string       `json:"mutated"`
}

// FileResult is the per-file rollup of mutants and Trust Score.
type FileResult struct {
	Path       string   `json:"path"`
	TrustScore *float64 `json:"trustScore"`
	Mutants    []Mutant `json:"mutants"`
}

// Totals is the run-level rollup. Score-related fields are pointers so a
// null in JSON round-trips as nil rather than a misleading zero value.
type Totals struct {
	Mutants      int      `json:"mutants"`
	Killed       int      `json:"killed"`
	Survived     int      `json:"survived"`
	Timeout      int      `json:"timeout"`
	NoCoverage   int      `json:"noCoverage"`
	Ignored      int      `json:"ignored,omitempty"`
	CompileError int      `json:"compileError,omitempty"`
	RuntimeError int      `json:"runtimeError,omitempty"`
	TrustScore   *float64 `json:"trustScore"`
	LineCoverage *float64 `json:"lineCoverage"`
	DurationMs   int64    `json:"durationMs"`
}

// Artifact is the full `.gauntlet/runs/<runId>.json` document.
type Artifact struct {
	SchemaVersion int          `json:"schemaVersion"`
	RunID         string       `json:"runId"`
	CreatedAt     string       `json:"createdAt"` // RFC3339
	BaseRef       string       `json:"baseRef"`
	HeadSha       string       `json:"headSha"`
	Trigger       Trigger      `json:"trigger"`
	Threshold     float64      `json:"threshold"`
	ChangedFiles  []string     `json:"changedFiles"`
	Totals        Totals       `json:"totals"`
	Files         []FileResult `json:"files"`
	ComparedTo    *string      `json:"comparedTo,omitempty"`
}
