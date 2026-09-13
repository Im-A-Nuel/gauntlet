package bob

import (
	"os"
	"path/filepath"
	"testing"
)

var (
	include = []string{"src/**/*.ts"}
	exclude = []string{"**/*.test.ts"}
)

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCompareCleanRunHasNoViolations(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "src/pricing.ts", "export const x = 1;\n")
	writeFile(t, dir, "src/pricing.test.ts", "expect(1).toBe(1);\n")

	before, err := TakeSnapshot(dir, include, exclude)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}
	// Simulate Bob adding a new assertion to the existing test file only.
	writeFile(t, dir, "src/pricing.test.ts", "expect(1).toBe(1);\nexpect(2).toBe(2);\n")
	after, err := TakeSnapshot(dir, include, exclude)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}

	res := Compare(before, after)
	if !res.OK {
		t.Fatalf("Compare = %+v, want OK (adding assertions to a test file is allowed)", res)
	}
}

func TestCompareFlagsSourceModification(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "src/pricing.ts", "export const x = 1;\n")
	before, _ := TakeSnapshot(dir, include, exclude)

	writeFile(t, dir, "src/pricing.ts", "export const x = 2;\n")
	after, _ := TakeSnapshot(dir, include, exclude)

	res := Compare(before, after)
	if res.OK {
		t.Fatalf("Compare = %+v, want a violation for modified source file", res)
	}
	if len(res.Violations) != 1 || res.Violations[0] != "source file modified: src/pricing.ts" {
		t.Fatalf("Violations = %v, want exactly one source-modified violation", res.Violations)
	}
}

func TestCompareFlagsDeletedTestFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "src/pricing.test.ts", "expect(1).toBe(1);\n")
	before, _ := TakeSnapshot(dir, include, exclude)

	if err := os.Remove(filepath.Join(dir, "src", "pricing.test.ts")); err != nil {
		t.Fatal(err)
	}
	after, _ := TakeSnapshot(dir, include, exclude)

	res := Compare(before, after)
	if res.OK {
		t.Fatalf("Compare = %+v, want a violation for deleted test file", res)
	}
	if len(res.Violations) != 1 || res.Violations[0] != "test file deleted: src/pricing.test.ts" {
		t.Fatalf("Violations = %v, want exactly one test-deleted violation", res.Violations)
	}
}

func TestCompareFlagsWeakenedAssertions(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "src/pricing.test.ts", "expect(1).toBe(1);\nexpect(2).toBe(2);\n")
	before, _ := TakeSnapshot(dir, include, exclude)

	writeFile(t, dir, "src/pricing.test.ts", "expect(1).toBe(1);\n")
	after, _ := TakeSnapshot(dir, include, exclude)

	res := Compare(before, after)
	if res.OK {
		t.Fatalf("Compare = %+v, want a violation for a dropped assertion", res)
	}
	if len(res.Violations) != 1 || res.Violations[0] != "assertion count dropped in src/pricing.test.ts: 2 -> 1" {
		t.Fatalf("Violations = %v, want exactly one assertion-drop violation", res.Violations)
	}
}

func TestCompareFlagsNewSourceFile(t *testing.T) {
	dir := t.TempDir()
	before, _ := TakeSnapshot(dir, include, exclude)
	writeFile(t, dir, "src/sneaky.ts", "export const y = 1;\n")
	after, _ := TakeSnapshot(dir, include, exclude)

	res := Compare(before, after)
	if res.OK {
		t.Fatalf("Compare = %+v, want a violation for an unexpected new source file", res)
	}
}
