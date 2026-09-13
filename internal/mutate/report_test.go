package mutate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Im-A-Nuel/gauntlet/internal/report"
)

const fixtureReport = `{
  "schemaVersion": "1.0",
  "files": {
    "src/pricing.ts": {
      "language": "typescript",
      "source": "export function ok(qty: number, bulkThreshold: number) {\n  if (qty >= bulkThreshold) {\n    return true;\n  }\n  return false;\n}\n",
      "mutants": [
        {
          "id": "1",
          "mutatorName": "ConditionalExpression",
          "replacement": ">",
          "status": "Survived",
          "location": { "start": { "line": 2, "column": 11 }, "end": { "line": 2, "column": 13 } }
        },
        {
          "id": "2",
          "mutatorName": "BooleanLiteral",
          "replacement": "false",
          "status": "Killed",
          "location": { "start": { "line": 3, "column": 12 }, "end": { "line": 3, "column": 16 } }
        },
        {
          "id": "3",
          "mutatorName": "BooleanLiteral",
          "replacement": "true",
          "status": "NoCoverage",
          "location": { "start": { "line": 5, "column": 10 }, "end": { "line": 5, "column": 15 } }
        }
      ]
    }
  }
}`

func writeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "mutation.json")
	if err := os.WriteFile(path, []byte(fixtureReport), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseReportNormalizesStatusesAndScore(t *testing.T) {
	path := writeFixture(t)
	files, err := ParseReport(path)
	if err != nil {
		t.Fatalf("ParseReport: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("len(files) = %d, want 1", len(files))
	}
	f := files[0]
	if f.Path != "src/pricing.ts" {
		t.Fatalf("f.Path = %q, want src/pricing.ts", f.Path)
	}
	if len(f.Mutants) != 3 {
		t.Fatalf("len(f.Mutants) = %d, want 3", len(f.Mutants))
	}

	byID := map[string]report.Mutant{}
	for _, m := range f.Mutants {
		byID[m.ID] = m
	}

	if byID["1"].Status != report.StatusSurvived {
		t.Fatalf("mutant 1 status = %q, want survived", byID["1"].Status)
	}
	if byID["2"].Status != report.StatusKilled {
		t.Fatalf("mutant 2 status = %q, want killed", byID["2"].Status)
	}
	if byID["3"].Status != report.StatusNoCoverage {
		t.Fatalf("mutant 3 status = %q, want noCoverage", byID["3"].Status)
	}

	// killed=1, survived=1, noCoverage excluded from denominator -> 50.0
	if f.TrustScore == nil || *f.TrustScore != 50.0 {
		t.Fatalf("f.TrustScore = %v, want 50.0 (noCoverage must not enter the denominator)", f.TrustScore)
	}
}

func TestParseReportReconstructsOriginalAndMutatedLines(t *testing.T) {
	path := writeFixture(t)
	files, err := ParseReport(path)
	if err != nil {
		t.Fatalf("ParseReport: %v", err)
	}
	var m1 report.Mutant
	for _, m := range files[0].Mutants {
		if m.ID == "1" {
			m1 = m
		}
	}
	if m1.Original != "  if (qty >= bulkThreshold) {" {
		t.Fatalf("Original = %q, want the full source line", m1.Original)
	}
	if m1.Mutated != "  if (qty > bulkThreshold) {" {
		t.Fatalf("Mutated = %q, want >= replaced by >", m1.Mutated)
	}
	if m1.Line != 2 {
		t.Fatalf("Line = %d, want 2", m1.Line)
	}
}

func TestParseReportMissingFile(t *testing.T) {
	if _, err := ParseReport(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatalf("ParseReport on missing file = nil error, want error")
	}
}
