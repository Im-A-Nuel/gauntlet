package bob

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Im-A-Nuel/gauntlet/internal/report"
)

// SurvivorsPath is the survivor handoff file consumed by the
// strengthen-tests Skill (docs/SCHEMA.md §3).
const SurvivorsPath = ".gauntlet/survivors.md"

// RenderSurvivors formats every surviving mutant in a as Markdown, matching
// the docs/SCHEMA.md §3 layout. The per-mutant "meaning" line is generated
// mechanically from the original/mutated span (not an LLM-authored insight):
// it names the code delta so a human or Bob has the diff spelled out, but
// makes no claim about *why* a given test suite failed to notice it.
func RenderSurvivors(a report.Artifact, threshold float64) (body string, survivorCount int) {
	var b strings.Builder
	fmt.Fprintf(&b, "# Surviving mutants (run %s)\n", a.RunID)
	trustScoreStr := "null"
	if a.Totals.TrustScore != nil {
		trustScoreStr = fmt.Sprintf("%.1f%%", *a.Totals.TrustScore)
	}
	fmt.Fprintf(&b, "Trust Score: %s | Threshold: %.0f%%\n", trustScoreStr, threshold)

	for _, f := range a.Files {
		var survivors []report.Mutant
		for _, m := range f.Mutants {
			if m.Status == report.StatusSurvived {
				survivors = append(survivors, m)
			}
		}
		if len(survivors) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n## %s\n", f.Path)
		for _, m := range survivors {
			survivorCount++
			fmt.Fprintf(&b, "- [%s] line %d %s\n", m.ID, m.Line, m.Mutator)
			fmt.Fprintf(&b, "  original: `%s`\n", strings.TrimSpace(m.Original))
			fmt.Fprintf(&b, "  mutated:  `%s`\n", strings.TrimSpace(m.Mutated))
			fmt.Fprintf(&b, "  meaning: no test distinguished the mutated line from the original\n")
		}
	}
	return b.String(), survivorCount
}

// WriteSurvivors renders and writes the survivor handoff file, returning its
// absolute path and how many surviving mutants it lists.
func WriteSurvivors(repoRoot string, a report.Artifact, threshold float64) (path string, count int, err error) {
	body, count := RenderSurvivors(a, threshold)
	path = filepath.Join(repoRoot, filepath.FromSlash(SurvivorsPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", 0, err
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", 0, err
	}
	return path, count, nil
}
