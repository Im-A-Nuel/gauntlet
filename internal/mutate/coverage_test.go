package mutate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadLineCoverageMissingReportIsNil(t *testing.T) {
	pct, err := ReadLineCoverage(t.TempDir())
	if err != nil {
		t.Fatalf("ReadLineCoverage: %v", err)
	}
	if pct != nil {
		t.Fatalf("pct = %v, want nil when no coverage report exists (never estimate)", pct)
	}
}

func TestReadLineCoverageParsesSummary(t *testing.T) {
	dir := t.TempDir()
	covDir := filepath.Join(dir, "coverage")
	if err := os.MkdirAll(covDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"total":{"lines":{"total":100,"covered":91,"skipped":0,"pct":91.0}}}`
	if err := os.WriteFile(filepath.Join(covDir, "coverage-summary.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	pct, err := ReadLineCoverage(dir)
	if err != nil {
		t.Fatalf("ReadLineCoverage: %v", err)
	}
	if pct == nil || *pct != 91.0 {
		t.Fatalf("pct = %v, want 91.0", pct)
	}
}
