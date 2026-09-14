# Integration closeout notes

Reviewed by Codex on 2026-09-14 after the CLI/demo implementation and dependency upgrade were merged into `main`.

## Verified

- `go test ./...`, `go vet ./...`, and `go build ./cmd/gauntlet` pass across eight Go packages.
- The demo weak suite passes 12 tests; the additive strong suite passes 33 tests. Both report 100% line coverage.
- Fresh Stryker 10 runs against the same 47 mutants produced a real 74.5% baseline and 97.9% follow-up. `gate --min-score 80` returned exit 1 for the baseline and exit 0 for the follow-up.
- Windows mutation execution uses the installed Stryker JavaScript entrypoint through Node instead of the `npx.cmd` shim.
- Scope resolution handles deleted files, NUL-delimited unusual filenames, invalid base refs, and symlinks escaping the project.
- Dashboard artifact parsing recomputes counts and scores, rejects malformed measurements, and never silently replaces a broken configured live directory with samples.
- Dashboard unit tests, strict type generation/checking, production build, Playwright interaction/API tests, axe checks, and responsive overflow checks pass.

## Remaining CLI hardening

1. `cmd/gauntlet/strengthen.go` acquires `.gauntlet/.lock` and defers cleanup, but several error paths call `die()`, which ultimately calls `os.Exit`. Go does not run deferred cleanup after `os.Exit`, so a failed Bob invocation can leave the lock behind. Return typed errors from the locked operation and map them to an exit code only after the deferred release has run; add a missing-Bob regression test.
2. `internal/report.Read` currently establishes JSON syntax and shape through unmarshalling but does not fully validate count consistency, status values, score recomputation, path-safe run IDs, or timestamps. The gate must reject fabricated or inconsistent artifacts before evaluating policy.

## Explicit limitations

- IBM Bob is not installed in this environment. Hook firing, live `bob run`, automatic Skill activation, and agent-written test changes remain unverified. `strengthen --prepare-only` is the honest local demonstration path.
- The existence-only lock has no PID/liveness recovery after a process crash.
- Test-strength verification uses file hashes and `expect(` counts, not semantic assertion analysis.
- `demo-repo` retains two moderate `qs` advisories in a transitive Stryker development-tool chain; dashboard dependencies have no reported vulnerabilities.
- GitHub branch protection must be enabled by the repository owner for the mutation workflow to become a merge-blocking required check.
