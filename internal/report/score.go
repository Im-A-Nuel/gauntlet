package report

import "math"

// TrustScore computes killed/(killed+survived)*100 per docs/SCHEMA.md,
// where the effective "killed" count folds in timeouts (Stryker convention:
// a mutant that times out is treated as caught). Raw status counts stay
// disjoint (killed excludes timeout) so callers can still report them
// separately; only the formula's numerator merges them.
//
// noCoverage/ignored/compileError/runtimeError never enter the denominator.
// An empty denominator (no scored mutants at all) yields nil, not 100 or 0:
// a score with nothing behind it is not information.
func TrustScore(killed, survived, timeout int) *float64 {
	effectiveKilled := killed + timeout
	denom := effectiveKilled + survived
	if denom == 0 {
		return nil
	}
	pct := math.Round(float64(effectiveKilled)/float64(denom)*1000) / 10
	return &pct
}

// CountsFromMutants tallies raw status counts for one mutant slice.
func CountsFromMutants(mutants []Mutant) (killed, survived, timeout, noCoverage, ignored, compileError, runtimeError int) {
	for _, m := range mutants {
		switch m.Status {
		case StatusKilled:
			killed++
		case StatusSurvived:
			survived++
		case StatusTimeout:
			timeout++
		case StatusNoCoverage:
			noCoverage++
		case StatusIgnored:
			ignored++
		case StatusCompileError:
			compileError++
		case StatusRuntimeError:
			runtimeError++
		}
	}
	return
}

// BuildFileResult rolls up one file's mutants into a FileResult, computing
// its own per-file Trust Score with the same formula as the run total.
func BuildFileResult(path string, mutants []Mutant) FileResult {
	killed, survived, timeout, _, _, _, _ := CountsFromMutants(mutants)
	return FileResult{
		Path:       path,
		TrustScore: TrustScore(killed, survived, timeout),
		Mutants:    mutants,
	}
}

// BuildTotals rolls up every file's mutants into the run-level Totals.
// lineCoverage is passed through as measured externally (never derived from
// mutation outcomes) and durationMs is the wall-clock time of the mutation
// run itself.
func BuildTotals(files []FileResult, lineCoverage *float64, durationMs int64) Totals {
	var killed, survived, timeout, noCoverage, ignored, compileError, runtimeError, total int
	for _, f := range files {
		k, s, t, n, i, c, r := CountsFromMutants(f.Mutants)
		killed += k
		survived += s
		timeout += t
		noCoverage += n
		ignored += i
		compileError += c
		runtimeError += r
		total += len(f.Mutants)
	}
	return Totals{
		Mutants:      total,
		Killed:       killed,
		Survived:     survived,
		Timeout:      timeout,
		NoCoverage:   noCoverage,
		Ignored:      ignored,
		CompileError: compileError,
		RuntimeError: runtimeError,
		TrustScore:   TrustScore(killed, survived, timeout),
		LineCoverage: lineCoverage,
		DurationMs:   durationMs,
	}
}
