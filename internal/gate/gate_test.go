package gate

import (
	"testing"

	"github.com/Im-A-Nuel/gauntlet/internal/report"
)

func scoreOf(v float64) *float64 { return &v }

func baseArtifact() report.Artifact {
	return report.Artifact{
		HeadSha: "a1b2c3d",
		Totals:  report.Totals{TrustScore: scoreOf(85.0)},
	}
}

func TestEvaluatePassesAboveThreshold(t *testing.T) {
	res := Evaluate(baseArtifact(), 80, "a1b2c3d")
	if !res.Pass || res.Reason != ReasonPass {
		t.Fatalf("Evaluate = %+v, want pass", res)
	}
}

func TestEvaluateFailsBelowThreshold(t *testing.T) {
	a := baseArtifact()
	a.Totals.TrustScore = scoreOf(41.4)
	res := Evaluate(a, 80, "a1b2c3d")
	if res.Pass || res.Reason != ReasonBelowThreshold {
		t.Fatalf("Evaluate = %+v, want below_threshold", res)
	}
}

func TestEvaluateFailsClosedOnNullScore(t *testing.T) {
	a := baseArtifact()
	a.Totals.TrustScore = nil
	res := Evaluate(a, 0, "a1b2c3d") // even a 0 threshold must not pass a null score
	if res.Pass || res.Reason != ReasonNullScore {
		t.Fatalf("Evaluate = %+v, want null_score (fail closed even at minScore=0)", res)
	}
}

func TestEvaluateFailsOnRuntimeError(t *testing.T) {
	a := baseArtifact()
	a.Totals.RuntimeError = 1
	res := Evaluate(a, 0, "a1b2c3d")
	if res.Pass || res.Reason != ReasonRuntimeError {
		t.Fatalf("Evaluate = %+v, want runtime_error (never passes regardless of score)", res)
	}
}

func TestEvaluateFailsOnStaleArtifact(t *testing.T) {
	res := Evaluate(baseArtifact(), 0, "deadbee")
	if res.Pass || res.Reason != ReasonStaleArtifact {
		t.Fatalf("Evaluate = %+v, want stale_artifact", res)
	}
}

func TestEvaluateSkipsFreshnessCheckWhenHeadShaOmitted(t *testing.T) {
	res := Evaluate(baseArtifact(), 80, "")
	if !res.Pass {
		t.Fatalf("Evaluate with empty currentHeadSha = %+v, want pass (freshness check opt-out)", res)
	}
}
