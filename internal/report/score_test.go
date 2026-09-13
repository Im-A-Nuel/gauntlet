package report

import "testing"

func float64Ptr(v float64) *float64 { return &v }

func TestTrustScore(t *testing.T) {
	cases := []struct {
		name                      string
		killed, survived, timeout int
		want                      *float64
	}{
		{"schema example 43.3%", 24, 34, 2, float64Ptr(43.3)},
		{"all killed is 100", 10, 0, 0, float64Ptr(100)},
		{"all survived is 0", 0, 10, 0, float64Ptr(0)},
		{"timeout only still scores", 0, 0, 3, float64Ptr(100)},
		{"no scored mutants is null", 0, 0, 0, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := TrustScore(c.killed, c.survived, c.timeout)
			if (got == nil) != (c.want == nil) {
				t.Fatalf("TrustScore(%d,%d,%d) = %v, want %v", c.killed, c.survived, c.timeout, got, c.want)
			}
			if got != nil && *got != *c.want {
				t.Fatalf("TrustScore(%d,%d,%d) = %v, want %v", c.killed, c.survived, c.timeout, *got, *c.want)
			}
		})
	}
}

func TestTrustScoreExcludesUnscoredStatuses(t *testing.T) {
	mutants := []Mutant{
		{ID: "1", Status: StatusKilled},
		{ID: "2", Status: StatusSurvived},
		{ID: "3", Status: StatusNoCoverage},
		{ID: "4", Status: StatusIgnored},
		{ID: "5", Status: StatusCompileError},
		{ID: "6", Status: StatusRuntimeError},
	}
	killed, survived, timeout, noCoverage, ignored, compileError, runtimeError := CountsFromMutants(mutants)
	if killed != 1 || survived != 1 || timeout != 0 {
		t.Fatalf("scored counts = %d/%d/%d, want 1/1/0", killed, survived, timeout)
	}
	if noCoverage != 1 || ignored != 1 || compileError != 1 || runtimeError != 1 {
		t.Fatalf("unscored counts = %d/%d/%d/%d, want 1/1/1/1", noCoverage, ignored, compileError, runtimeError)
	}
	got := TrustScore(killed, survived, timeout)
	want := 50.0 // 1 killed / (1 killed + 1 survived), unscored statuses never enter the denominator
	if got == nil || *got != want {
		t.Fatalf("TrustScore = %v, want %v (unscored statuses must not affect denominator)", got, want)
	}
}

func TestBuildTotalsAggregatesAcrossFiles(t *testing.T) {
	files := []FileResult{
		BuildFileResult("a.ts", []Mutant{{ID: "1", Status: StatusKilled}, {ID: "2", Status: StatusSurvived}}),
		BuildFileResult("b.ts", []Mutant{{ID: "3", Status: StatusTimeout}, {ID: "4", Status: StatusNoCoverage}}),
	}
	cov := float64Ptr(91.0)
	totals := BuildTotals(files, cov, 1234)
	if totals.Mutants != 4 || totals.Killed != 1 || totals.Survived != 1 || totals.Timeout != 1 || totals.NoCoverage != 1 {
		t.Fatalf("unexpected totals: %+v", totals)
	}
	// effective killed = killed(1) + timeout(1) = 2; denom = 2 + survived(1) = 3 -> 66.7
	if totals.TrustScore == nil || *totals.TrustScore != 66.7 {
		t.Fatalf("totals.TrustScore = %v, want 66.7", totals.TrustScore)
	}
	if totals.LineCoverage == nil || *totals.LineCoverage != 91.0 {
		t.Fatalf("totals.LineCoverage = %v, want 91.0", totals.LineCoverage)
	}
	if totals.DurationMs != 1234 {
		t.Fatalf("totals.DurationMs = %d, want 1234", totals.DurationMs)
	}
}
