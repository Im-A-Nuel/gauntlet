// Package gate implements the CI merge-gate policy: exit 0 iff the latest
// run artifact is fresh, fully scored, and its Trust Score clears the
// configured threshold (docs/SCHEMA.md §7).
package gate

import (
	"fmt"

	"github.com/Im-A-Nuel/gauntlet/internal/report"
)

// Reason identifies why a gate check failed, independent of the freeform
// message, so callers (CLI, tests) can branch on it.
type Reason string

const (
	ReasonPass           Reason = "pass"
	ReasonBelowThreshold Reason = "below_threshold"
	ReasonNullScore      Reason = "null_score"
	ReasonRuntimeError   Reason = "runtime_error"
	ReasonStaleArtifact  Reason = "stale_artifact"
)

// Result is the outcome of a gate evaluation.
type Result struct {
	Pass    bool
	Reason  Reason
	Message string
}

// Evaluate applies the gate policy to a run artifact against minScore and
// the repo's current HEAD sha. It fails closed: a null Trust Score, any
// runtimeError mutants, or an artifact that does not match currentHeadSha
// (docs/SCHEMA.md §2: "artifacts from a different HEAD") never pass,
// regardless of the numeric score.
func Evaluate(a report.Artifact, minScore float64, currentHeadSha string) Result {
	if currentHeadSha != "" && a.HeadSha != currentHeadSha {
		return Result{
			Pass:   false,
			Reason: ReasonStaleArtifact,
			Message: fmt.Sprintf(
				"latest run artifact is for HEAD %s but the repo is now at %s; run `gauntlet run` again before gating",
				shortSha(a.HeadSha), shortSha(currentHeadSha),
			),
		}
	}

	if a.Totals.RuntimeError > 0 {
		return Result{
			Pass:    false,
			Reason:  ReasonRuntimeError,
			Message: fmt.Sprintf("%d mutant(s) produced a runtime error; a run with runtime errors never passes the gate", a.Totals.RuntimeError),
		}
	}

	if a.Totals.TrustScore == nil {
		return Result{
			Pass:    false,
			Reason:  ReasonNullScore,
			Message: "Trust Score is null (no scored mutants: empty change set, no-coverage-only run, or malformed artifact); the gate fails closed",
		}
	}

	if *a.Totals.TrustScore < minScore {
		return Result{
			Pass:    false,
			Reason:  ReasonBelowThreshold,
			Message: fmt.Sprintf("Trust Score %.1f is below the minimum %.1f", *a.Totals.TrustScore, minScore),
		}
	}

	return Result{
		Pass:    true,
		Reason:  ReasonPass,
		Message: fmt.Sprintf("Trust Score %.1f meets the minimum %.1f", *a.Totals.TrustScore, minScore),
	}
}

func shortSha(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	if sha == "" {
		return "(none)"
	}
	return sha
}
