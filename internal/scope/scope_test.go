package scope

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

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

// initRepo creates a fresh git repo (local-only identity, no global config
// touched) on branch "main" with one commit.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, "init", "-b", "main")
	run(t, dir, "config", "user.email", "test@example.com")
	run(t, dir, "config", "user.name", "Test")
	writeFile(t, dir, "src/a.ts", "export const a = 1;\n")
	writeFile(t, dir, "README.md", "demo\n")
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-m", "init")
	return dir
}

func TestResolveNotGitRepo(t *testing.T) {
	dir := t.TempDir()
	if _, err := Resolve(dir, "main", []string{"src/**/*.ts"}, nil); err == nil {
		t.Fatalf("Resolve on non-git dir = nil error, want error")
	}
}

func TestResolveEmptyWhenNoChanges(t *testing.T) {
	dir := initRepo(t)
	res, err := Resolve(dir, "main", []string{"src/**/*.ts"}, nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res.Files) != 0 {
		t.Fatalf("Files = %v, want empty (no changes yet)", res.Files)
	}
	if res.HeadSha == "" {
		t.Fatalf("HeadSha is empty")
	}
}

func TestResolveDetectsWorktreeAndUntracked(t *testing.T) {
	dir := initRepo(t)
	// Modify a tracked file (worktree change, unstaged).
	writeFile(t, dir, "src/a.ts", "export const a = 2;\n")
	// Add an untracked file.
	writeFile(t, dir, "src/b.ts", "export const b = 1;\n")

	res, err := Resolve(dir, "main", []string{"src/**/*.ts"}, nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := []string{"src/a.ts", "src/b.ts"}
	if !equalSlices(res.Files, want) {
		t.Fatalf("Files = %v, want %v", res.Files, want)
	}
}

func TestResolveExcludesDeletedFiles(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "src/c.ts", "export const c = 1;\n")
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-m", "add c")

	if err := os.Remove(filepath.Join(dir, "src", "c.ts")); err != nil {
		t.Fatal(err)
	}

	res, err := Resolve(dir, "main", []string{"src/**/*.ts"}, nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	for _, f := range res.Files {
		if f == "src/c.ts" {
			t.Fatalf("Files = %v, deleted file src/c.ts must not be a mutation candidate", res.Files)
		}
	}
}

func TestResolveAppliesExcludeGlob(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "src/d.ts", "export const d = 1;\n")
	writeFile(t, dir, "src/d.test.ts", "test('x', () => {});\n")

	res, err := Resolve(dir, "main", []string{"src/**/*.ts"}, []string{"**/*.test.ts"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := []string{"src/d.ts"}
	if !equalSlices(res.Files, want) {
		t.Fatalf("Files = %v, want %v (test file must be excluded)", res.Files, want)
	}
}

func TestResolveDetectsCommittedDivergenceFromBaseRef(t *testing.T) {
	dir := initRepo(t)
	run(t, dir, "checkout", "-b", "feature")
	writeFile(t, dir, "src/feature.ts", "export const f = 1;\n")
	run(t, dir, "add", ".")
	run(t, dir, "commit", "-m", "feature work")

	res, err := Resolve(dir, "main", []string{"src/**/*.ts"}, nil)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	want := []string{"src/feature.ts"}
	if !equalSlices(res.Files, want) {
		t.Fatalf("Files = %v, want %v (fully committed change vs base ref)", res.Files, want)
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
