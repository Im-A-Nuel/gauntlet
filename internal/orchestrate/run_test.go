package orchestrate

import (
	"bytes"
	"os/exec"
	"testing"
	"time"

	"github.com/Im-A-Nuel/gauntlet/internal/config"
	"github.com/Im-A-Nuel/gauntlet/internal/report"
)

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func initRepoWithConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitCmd(t, dir, "init", "-b", "main")
	gitCmd(t, dir, "config", "user.email", "test@example.com")
	gitCmd(t, dir, "config", "user.name", "Test")
	gitCmd(t, dir, "commit", "--allow-empty", "-m", "init")
	if err := config.Write(dir, config.Default()); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return dir
}

func TestRunMissingConfigIsEnvironmentError(t *testing.T) {
	dir := t.TempDir()
	gitCmd(t, dir, "init", "-b", "main")
	_, err := Run(Options{RepoRoot: dir, Trigger: report.TriggerManual, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err == nil {
		t.Fatalf("Run with no config = nil error, want config/environment error")
	}
}

func TestRunEmptyScopeWritesExplicitArtifactWithoutInvokingStryker(t *testing.T) {
	dir := initRepoWithConfig(t)
	a, err := Run(Options{RepoRoot: dir, Trigger: report.TriggerManual, Now: func() time.Time { return time.Unix(1000, 0) }, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(a.ChangedFiles) != 0 {
		t.Fatalf("ChangedFiles = %v, want empty", a.ChangedFiles)
	}
	if a.Totals.TrustScore != nil {
		t.Fatalf("TrustScore = %v, want nil for an empty change set", a.Totals.TrustScore)
	}
	if a.Totals.Mutants != 0 {
		t.Fatalf("Mutants = %d, want 0", a.Totals.Mutants)
	}

	got, ok, err := report.Latest(dir)
	if err != nil || !ok {
		t.Fatalf("expected a written artifact: ok=%v err=%v", ok, err)
	}
	if got.RunID != a.RunID {
		t.Fatalf("written artifact RunID = %q, want %q", got.RunID, a.RunID)
	}
}

func TestRunChainsComparedToAcrossRuns(t *testing.T) {
	dir := initRepoWithConfig(t)
	first, err := Run(Options{RepoRoot: dir, Trigger: report.TriggerManual, Now: func() time.Time { return time.Unix(1000, 0) }, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("first Run: %v", err)
	}
	if first.ComparedTo != nil {
		t.Fatalf("first.ComparedTo = %v, want nil (no prior run)", first.ComparedTo)
	}

	second, err := Run(Options{RepoRoot: dir, Trigger: report.TriggerManual, Now: func() time.Time { return time.Unix(2000, 0) }, Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if second.ComparedTo == nil || *second.ComparedTo != first.RunID {
		t.Fatalf("second.ComparedTo = %v, want %q", second.ComparedTo, first.RunID)
	}
}
