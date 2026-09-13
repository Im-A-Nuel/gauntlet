package mutate

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// strykerEntrypoint returns the path to Stryker's own JS CLI entrypoint
// inside the target repo's node_modules, as declared by its package.json
// "bin" field (node_modules/@stryker-mutator/core/bin/stryker.js).
//
// We invoke this file directly via `node <entrypoint>` rather than `npx
// stryker` or the node_modules/.bin/stryker(.cmd) shim: on Windows, the npx
// and .bin shims are .cmd files, and exec.Command cannot execute a .cmd
// directly (CreateProcess only runs native executables; there is no shell in
// between). `node.exe` is a native executable on every platform, so this
// path works identically on Windows, macOS, and Linux with no shell
// involved anywhere in the invocation.
func strykerEntrypoint(repoRoot string) (string, error) {
	entry := filepath.Join(repoRoot, "node_modules", "@stryker-mutator", "core", "bin", "stryker.js")
	if _, err := os.Stat(entry); err != nil {
		return "", fmt.Errorf("stryker not found at %s: run `npm install` in %s first", entry, repoRoot)
	}
	return entry, nil
}

// Execute runs Stryker's mutation run against configPath with cwd=repoRoot,
// streaming Stryker's own output to stdout/stderr so a human running the CLI
// sees live progress. It returns the wall-clock duration of the mutation run.
//
// A non-zero exit here means Stryker itself failed to complete (bad config,
// no mutants found, a crashed test runner) — the caller should treat that as
// a run failure (exit code 2, docs/SCHEMA.md §7), distinct from a completed
// run that simply scored low.
func Execute(repoRoot, configPath string, stdout, stderr io.Writer) (durationMs int64, err error) {
	entry, err := strykerEntrypoint(repoRoot)
	if err != nil {
		return 0, err
	}

	// Stryker's CLI takes the config file as a positional argument
	// (`stryker run [options] [configFile]`), not a --configFile flag.
	cmd := exec.Command("node", entry, "run", configPath)
	cmd.Dir = repoRoot
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	err = cmd.Run()
	durationMs = time.Since(start).Milliseconds()
	if err != nil {
		return durationMs, fmt.Errorf("stryker run failed: %w", err)
	}
	return durationMs, nil
}

// ReportPath returns the absolute path to the Stryker JSON report Execute's
// invocation would have produced (Stryker's default reporters:["json"] output
// location, relative to the cwd it ran in).
func ReportPath(repoRoot string) string {
	return filepath.Join(repoRoot, filepath.FromSlash(DefaultReportPath))
}
