package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binPath string

// TestMain compiles the gauntlet binary once so the CLI integration tests
// below exercise the real command wiring (cobra flags, exit codes, file
// side effects) end-to-end as a subprocess, the same way a user invokes it.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gauntlet-bin-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	name := "gauntlet"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binPath = filepath.Join(dir, name)

	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		panic("build gauntlet binary: " + err.Error() + "\n" + string(out))
	}

	os.Exit(m.Run())
}

type result struct {
	stdout, stderr string
	exitCode       int
}

func runCLI(t *testing.T, args ...string) result {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("run %v: %v", args, err)
		}
	}
	return result{stdout: stdout.String(), stderr: stderr.String(), exitCode: code}
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitCmd(t, dir, "init", "-b", "main")
	gitCmd(t, dir, "config", "user.email", "test@example.com")
	gitCmd(t, dir, "config", "user.name", "Test")
	gitCmd(t, dir, "commit", "--allow-empty", "-m", "init")
	return dir
}

func TestCLIInitWritesConfigHookAndSkill(t *testing.T) {
	dir := initTestRepo(t)

	res := runCLI(t, "--repo", dir, "init")
	if res.exitCode != 0 {
		t.Fatalf("init exit code = %d, want 0\nstdout: %s\nstderr: %s", res.exitCode, res.stdout, res.stderr)
	}
	for _, rel := range []string{".gauntlet/config.yaml", ".bob/settings.json", ".bob/skills/strengthen-tests/SKILL.md"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("init did not write %s: %v", rel, err)
		}
	}

	// Idempotent: second init must not fail or duplicate the hook.
	res2 := runCLI(t, "--repo", dir, "init")
	if res2.exitCode != 0 {
		t.Fatalf("second init exit code = %d, want 0\nstdout: %s\nstderr: %s", res2.exitCode, res2.stdout, res2.stderr)
	}
}

func TestCLIRunWithNoChangesWritesNullScoreArtifact(t *testing.T) {
	dir := initTestRepo(t)
	if res := runCLI(t, "--repo", dir, "init"); res.exitCode != 0 {
		t.Fatalf("init failed: %s", res.stderr)
	}

	res := runCLI(t, "--repo", dir, "run", "--trigger", "manual")
	if res.exitCode != 0 {
		t.Fatalf("run exit code = %d, want 0\nstdout: %s\nstderr: %s", res.exitCode, res.stdout, res.stderr)
	}
	if !contains(res.stdout, "trustScore: null") {
		t.Fatalf("run stdout missing null trustScore:\n%s", res.stdout)
	}

	reportRes := runCLI(t, "--repo", dir, "report")
	if reportRes.exitCode != 0 {
		t.Fatalf("report exit code = %d, want 0\nstderr: %s", reportRes.exitCode, reportRes.stderr)
	}
	if !contains(reportRes.stdout, "trustScore: null") {
		t.Fatalf("report stdout missing null trustScore:\n%s", reportRes.stdout)
	}
}

func TestCLIGateFailsClosedOnNullScore(t *testing.T) {
	dir := initTestRepo(t)
	runCLI(t, "--repo", dir, "init")
	runCLI(t, "--repo", dir, "run")

	res := runCLI(t, "--repo", dir, "gate", "--min-score", "0")
	if res.exitCode != ExitGateFailed {
		t.Fatalf("gate exit code = %d, want %d (fail closed on null score even at threshold 0)\nstdout: %s", res.exitCode, ExitGateFailed, res.stdout)
	}
}

func TestCLIGateNoRunsIsConfigError(t *testing.T) {
	dir := initTestRepo(t)
	runCLI(t, "--repo", dir, "init")

	res := runCLI(t, "--repo", dir, "gate", "--min-score", "80")
	if res.exitCode != ExitConfigOrEnvError {
		t.Fatalf("gate with no runs exit code = %d, want %d", res.exitCode, ExitConfigOrEnvError)
	}
}

func TestCLIRunInvalidTriggerIsConfigError(t *testing.T) {
	dir := initTestRepo(t)
	runCLI(t, "--repo", dir, "init")

	res := runCLI(t, "--repo", dir, "run", "--trigger", "bogus")
	if res.exitCode != ExitConfigOrEnvError {
		t.Fatalf("run --trigger bogus exit code = %d, want %d", res.exitCode, ExitConfigOrEnvError)
	}
}

func TestCLIStrengthenWithNoRunsIsConfigError(t *testing.T) {
	dir := initTestRepo(t)
	runCLI(t, "--repo", dir, "init")

	res := runCLI(t, "--repo", dir, "strengthen", "--prepare-only")
	if res.exitCode != ExitConfigOrEnvError {
		t.Fatalf("strengthen with no runs exit code = %d, want %d\nstdout: %s\nstderr: %s", res.exitCode, ExitConfigOrEnvError, res.stdout, res.stderr)
	}
}

func TestCLIStrengthenPrepareOnlyWithNoSurvivorsIsClean(t *testing.T) {
	dir := initTestRepo(t)
	runCLI(t, "--repo", dir, "init")
	runCLI(t, "--repo", dir, "run") // no changes -> zero mutants, zero survivors

	res := runCLI(t, "--repo", dir, "strengthen", "--prepare-only")
	if res.exitCode != 0 {
		t.Fatalf("strengthen --prepare-only exit code = %d, want 0\nstdout: %s\nstderr: %s", res.exitCode, res.stdout, res.stderr)
	}
	if !contains(res.stdout, "nothing to strengthen") {
		t.Fatalf("expected 'nothing to strengthen' with zero survivors, got:\n%s", res.stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gauntlet", "survivors.md")); err != nil {
		t.Fatalf("survivors.md not written: %v", err)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
