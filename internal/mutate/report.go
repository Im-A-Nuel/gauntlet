package mutate

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Im-A-Nuel/gauntlet/internal/report"
)

// DefaultReportPath is where Stryker writes its JSON report by default
// (relative to the cwd it was run in) when configured with reporters: ["json"].
const DefaultReportPath = "reports/mutation/mutation.json"

// Raw Stryker mutation-testing-report-schema types (only the fields Gauntlet
// consumes; see https://github.com/stryker-mutator/mutation-testing-elements).
type strykerReport struct {
	Files map[string]strykerFile `json:"files"`
}

type strykerFile struct {
	Source  string          `json:"source"`
	Mutants []strykerMutant `json:"mutants"`
}

type strykerMutant struct {
	ID          string     `json:"id"`
	MutatorName string     `json:"mutatorName"`
	Replacement string     `json:"replacement"`
	Status      string     `json:"status"`
	Location    strykerLoc `json:"location"`
}

type strykerLoc struct {
	Start strykerPos `json:"start"`
	End   strykerPos `json:"end"`
}

type strykerPos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// statusMap converts Stryker's PascalCase status strings to Gauntlet's
// MutantStatus enum (docs/SCHEMA.md §2).
var statusMap = map[string]report.MutantStatus{
	"Killed":       report.StatusKilled,
	"Survived":     report.StatusSurvived,
	"NoCoverage":   report.StatusNoCoverage,
	"Timeout":      report.StatusTimeout,
	"Ignored":      report.StatusIgnored,
	"CompileError": report.StatusCompileError,
	"RuntimeError": report.StatusRuntimeError,
}

// ParseReport reads and normalizes a Stryker mutation.json report at path
// into per-file Gauntlet FileResults, sorted by path.
func ParseReport(path string) ([]report.FileResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read stryker report %s: %w", path, err)
	}
	var raw strykerReport
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse stryker report %s: %w", path, err)
	}

	paths := make([]string, 0, len(raw.Files))
	for p := range raw.Files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	results := make([]report.FileResult, 0, len(paths))
	for _, p := range paths {
		f := raw.Files[p]
		lines := strings.Split(f.Source, "\n")
		mutants := make([]report.Mutant, 0, len(f.Mutants))
		for _, m := range f.Mutants {
			status, ok := statusMap[m.Status]
			if !ok {
				status = report.MutantStatus(strings.ToLower(m.Status[:1]) + m.Status[1:])
			}
			original, mutated := spliceMutation(lines, m.Location, m.Replacement)
			mutants = append(mutants, report.Mutant{
				ID:       m.ID,
				Mutator:  m.MutatorName,
				Line:     m.Location.Start.Line,
				Status:   status,
				Original: original,
				Mutated:  mutated,
			})
		}
		results = append(results, report.BuildFileResult(p, mutants))
	}
	return results, nil
}

// spliceMutation reconstructs the human-readable "original" and "mutated"
// line text (docs/SCHEMA.md §2 example) from Stryker's 1-based line/column
// location and replacement string, since the raw report only stores the
// mutated span, not rendered before/after lines.
func spliceMutation(lines []string, loc strykerLoc, replacement string) (original, mutated string) {
	startIdx := loc.Start.Line - 1
	endIdx := loc.End.Line - 1
	if startIdx < 0 || startIdx >= len(lines) {
		return "", replacement
	}
	if endIdx < 0 || endIdx >= len(lines) {
		endIdx = startIdx
	}

	if startIdx == endIdx {
		line := lines[startIdx]
		sc := clampCol(loc.Start.Column-1, len(line))
		ec := clampCol(loc.End.Column-1, len(line))
		if ec < sc {
			ec = sc
		}
		mutated = line[:sc] + replacement + line[ec:]
		return line, mutated
	}

	// Rare multi-line mutation span: report the joined source lines as
	// "original" and a best-effort single collapsed "mutated" line.
	original = strings.Join(lines[startIdx:endIdx+1], "\n")
	first := lines[startIdx]
	last := lines[endIdx]
	sc := clampCol(loc.Start.Column-1, len(first))
	ec := clampCol(loc.End.Column-1, len(last))
	mutated = first[:sc] + replacement + last[ec:]
	return original, mutated
}

func clampCol(col, lineLen int) int {
	if col < 0 {
		return 0
	}
	if col > lineLen {
		return lineLen
	}
	return col
}
