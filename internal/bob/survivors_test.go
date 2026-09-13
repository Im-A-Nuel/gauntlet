package bob

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Im-A-Nuel/gauntlet/internal/report"
)

func scoreOf(v float64) *float64 { return &v }

func sampleArtifact() report.Artifact {
	return report.Artifact{
		RunID: "1758945600000-a1b2c3d",
		Totals: report.Totals{
			TrustScore: scoreOf(41.4),
		},
		Files: []report.FileResult{
			{
				Path: "src/pricing.ts",
				Mutants: []report.Mutant{
					{ID: "m-017", Line: 42, Mutator: "ConditionalExpression", Status: report.StatusSurvived,
						Original: "if (qty >= bulkThreshold)", Mutated: "if (qty > bulkThreshold)"},
					{ID: "m-018", Line: 50, Mutator: "BooleanLiteral", Status: report.StatusKilled},
				},
			},
			{
				Path: "src/cart.ts",
				Mutants: []report.Mutant{
					{ID: "m-002", Line: 5, Mutator: "ArithmeticOperator", Status: report.StatusSurvived,
						Original: "return a + b;", Mutated: "return a - b;"},
				},
			},
		},
	}
}

func TestRenderSurvivorsListsOnlySurvived(t *testing.T) {
	body, count := RenderSurvivors(sampleArtifact(), 80)
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
	if !strings.Contains(body, "m-017") || !strings.Contains(body, "m-002") {
		t.Fatalf("body missing survivor ids:\n%s", body)
	}
	if strings.Contains(body, "m-018") {
		t.Fatalf("body must not list killed mutant m-018:\n%s", body)
	}
	if !strings.Contains(body, "Trust Score: 41.4%") || !strings.Contains(body, "Threshold: 80%") {
		t.Fatalf("body missing header line:\n%s", body)
	}
}

func TestRenderSurvivorsHandlesNullScore(t *testing.T) {
	a := sampleArtifact()
	a.Totals.TrustScore = nil
	body, _ := RenderSurvivors(a, 80)
	if !strings.Contains(body, "Trust Score: null") {
		t.Fatalf("body must render null score literally, got:\n%s", body)
	}
}

func TestWriteSurvivorsWritesFile(t *testing.T) {
	dir := t.TempDir()
	path, count, err := WriteSurvivors(dir, sampleArtifact(), 80)
	if err != nil {
		t.Fatalf("WriteSurvivors: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("survivors file not written: %v", err)
	}
	want := filepath.Join(dir, filepath.FromSlash(SurvivorsPath))
	if path != want {
		t.Fatalf("path = %q, want %q", path, want)
	}
}

func TestRenderSurvivorsNoSurvivorsIsEmpty(t *testing.T) {
	a := sampleArtifact()
	for i := range a.Files {
		for j := range a.Files[i].Mutants {
			a.Files[i].Mutants[j].Status = report.StatusKilled
		}
	}
	_, count := RenderSurvivors(a, 80)
	if count != 0 {
		t.Fatalf("count = %d, want 0 when nothing survived", count)
	}
}
