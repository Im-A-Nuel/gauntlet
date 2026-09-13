// Package orchestrate composes config, scope, mutate, bob (locking), and
// report into the single `doRun` critical section shared by the `run`
// command and `strengthen`'s post-invocation re-run, so the two never
// duplicate (or race) the mutation-run logic.
package orchestrate

import (
	"fmt"
	"io"
	"time"

	"github.com/Im-A-Nuel/gauntlet/internal/bob"
	"github.com/Im-A-Nuel/gauntlet/internal/config"
	"github.com/Im-A-Nuel/gauntlet/internal/mutate"
	"github.com/Im-A-Nuel/gauntlet/internal/report"
	"github.com/Im-A-Nuel/gauntlet/internal/scope"
)

// Options configures one Run invocation.
type Options struct {
	RepoRoot            string
	Trigger             report.Trigger
	BaseRefOverride     string // "" uses config.BaseRef
	TestCommandOverride string // "" uses config.TestCommand; lets `strengthen`/demos select an alternate test suite (e.g. the strengthened tests) without editing config.yaml
	Now                 func() time.Time
	Stdout, Stderr      io.Writer
	// SkipLock is set by `strengthen`, which holds bob.Acquire for its whole
	// snapshot -> invoke Bob -> verify -> re-run sequence itself (a nested
	// hook-triggered run must not slip in between those steps); Run must not
	// then try to acquire the same lock again from the same process.
	SkipLock bool
}

// RunFailedError distinguishes a completed-but-failed mutation run (Stryker
// or the test command itself errored — docs/SCHEMA.md §7 exit code 2) from a
// config/environment error (exit code 3). Callers (cmd/gauntlet) map this to
// the right process exit code.
type RunFailedError struct{ Err error }

func (e *RunFailedError) Error() string { return e.Err.Error() }
func (e *RunFailedError) Unwrap() error { return e.Err }

// Run resolves changed-file scope, and — unless the scope is empty — runs a
// scoped Stryker mutation pass, normalizes the result, and writes a new run
// artifact. An empty scope still writes an explicit artifact (zero mutants,
// null Trust Score) rather than silently doing nothing or falling back to a
// whole-repo run.
func Run(opts Options) (report.Artifact, error) {
	now := opts.Now
	if now == nil {
		now = time.Now
	}

	cfg, err := config.Load(opts.RepoRoot)
	if err != nil {
		return report.Artifact{}, err // config/environment error, exit 3
	}

	baseRef := cfg.BaseRef
	if opts.BaseRefOverride != "" {
		baseRef = opts.BaseRefOverride
	}
	testCommand := cfg.TestCommand
	if opts.TestCommandOverride != "" {
		testCommand = opts.TestCommandOverride
	}

	scopeRes, err := scope.Resolve(opts.RepoRoot, baseRef, cfg.Include, cfg.Exclude)
	if err != nil {
		return report.Artifact{}, fmt.Errorf("resolve changed-file scope: %w", err) // exit 3
	}

	prev, hasPrev, err := report.Latest(opts.RepoRoot)
	if err != nil {
		return report.Artifact{}, fmt.Errorf("read previous run artifact: %w", err)
	}
	var comparedTo *string
	if hasPrev {
		id := prev.RunID
		comparedTo = &id
	}

	runID := report.NewRunID(now(), scopeRes.HeadSha)
	base := report.Artifact{
		SchemaVersion: report.SchemaVersion,
		RunID:         runID,
		CreatedAt:     now().UTC().Format(time.RFC3339),
		BaseRef:       baseRef,
		HeadSha:       scopeRes.HeadSha,
		Trigger:       opts.Trigger,
		Threshold:     cfg.MinScore,
		ChangedFiles:  scopeRes.Files,
		ComparedTo:    comparedTo,
	}

	if len(scopeRes.Files) == 0 {
		base.Files = []report.FileResult{}
		base.Totals = report.BuildTotals(base.Files, nil, 0)
		if _, err := report.Write(opts.RepoRoot, base); err != nil {
			return report.Artifact{}, fmt.Errorf("write run artifact: %w", err)
		}
		return base, nil
	}

	if !opts.SkipLock {
		release, err := bob.Acquire(opts.RepoRoot)
		if err != nil {
			return report.Artifact{}, err // exit 3: reentrancy/contention
		}
		defer release()
	}

	configPath, cleanup, err := mutate.GenerateConfig(scopeRes.Files, testCommand, cfg.Concurrency)
	if err != nil {
		return report.Artifact{}, fmt.Errorf("generate stryker config: %w", err)
	}
	defer cleanup()

	durationMs, err := mutate.Execute(opts.RepoRoot, configPath, opts.Stdout, opts.Stderr)
	if err != nil {
		return report.Artifact{}, &RunFailedError{Err: err}
	}

	files, err := mutate.ParseReport(mutate.ReportPath(opts.RepoRoot))
	if err != nil {
		return report.Artifact{}, &RunFailedError{Err: fmt.Errorf("parse stryker report: %w", err)}
	}
	lineCoverage, _ := mutate.ReadLineCoverage(opts.RepoRoot) // best-effort; nil if unmeasured

	base.Files = files
	base.Totals = report.BuildTotals(files, lineCoverage, durationMs)

	if _, err := report.Write(opts.RepoRoot, base); err != nil {
		return report.Artifact{}, fmt.Errorf("write run artifact: %w", err)
	}
	return base, nil
}
