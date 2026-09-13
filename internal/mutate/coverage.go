package mutate

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// coverageSummaryPath is the standard location for Istanbul/V8
// json-summary-reporter output (e.g. `vitest run --coverage` with the
// json-summary reporter enabled), relative to a repo root.
const coverageSummaryPath = "coverage/coverage-summary.json"

type coverageSummary struct {
	Total struct {
		Lines struct {
			Pct float64 `json:"pct"`
		} `json:"lines"`
	} `json:"total"`
}

// ReadLineCoverage looks for a measured coverage-summary.json under
// repoRoot and returns its total line-coverage percentage. It returns nil
// (not 0, not an estimate) when no coverage report is present — line
// coverage must be measured, never derived from mutation outcomes
// (docs/SCHEMA.md §2).
func ReadLineCoverage(repoRoot string) (*float64, error) {
	path := filepath.Join(repoRoot, filepath.FromSlash(coverageSummaryPath))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var summary coverageSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, err
	}
	pct := summary.Total.Lines.Pct
	return &pct, nil
}
