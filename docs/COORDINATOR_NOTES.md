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

Both items below are resolved as of this checkpoint (Claude Code session `gauntlet-claude`); full detail, exact diffs touched, and verification evidence are in `docs/CLAUDE_PROGRESS.md`'s "CLI hardening checkpoint" section. Summary for whoever reviews next:

1. ~~`cmd/gauntlet/strengthen.go` acquires `.gauntlet/.lock` and defers cleanup, but several error paths call `die()`, which ultimately calls `os.Exit`.~~ **Resolved.** `strengthen`'s `RunE` now only acquires the lock, defers release, and returns a plain Go error from an extracted `runLockedStrengthen` function (never calls `die` after the lock). `cmd/gauntlet/root.go` gained an `*exitError`/`exitErrorf` mechanism so `main` can still map a returned error to a specific exit code (not just the generic config-error code) after `Execute()` returns — i.e. after every deferred cleanup in the call stack has already run. Regression test: `TestStrengthenMissingBobExecutableLeavesNoLock` in `cmd/gauntlet/main_test.go`.
2. ~~`internal/report.Read` currently establishes JSON syntax and shape through unmarshalling but does not fully validate count consistency, status values, score recomputation, path-safe run IDs, or timestamps.~~ **Resolved.** New `internal/report/validate.go` adds `(Artifact) Validate() error`, called by `Read` on every load: schema version, RFC3339 timestamp, enum checks (trigger, mutant status), path-safety on run IDs/`changedFiles`/`files[].path`/`comparedTo` (closes a `--run-id` path-traversal), per-file and per-run mutant-ID uniqueness, range checks (threshold/coverage/duration), and — the core fix — every score and count is recomputed from the artifact's own raw mutants via the existing `score.go` helpers and rejected on any mismatch, so a fabricated `trustScore` or count no longer passes. `Read` also now rejects a run ID whose filename and declared content disagree. 24 table-driven negative cases in `internal/report/validate_test.go` plus `TestReadRejectsOnDisk`; two pre-existing tests with under-specified fixtures were updated to construct schema-valid artifacts.

`go build`, `go vet`, and `go test ./...` all pass across all 8 Go packages after both fixes; real on-disk artifacts from prior runs still read correctly under the new validation (no regression), and a live path-traversal `--run-id` attempt against the compiled binary is rejected at exit code 3.

## Explicit limitations

- IBM Bob is not installed in this environment. Hook firing, live `bob run`, automatic Skill activation, and agent-written test changes remain unverified. `strengthen --prepare-only` is the honest local demonstration path.
- The existence-only lock has no PID/liveness recovery after a process crash.
- Test-strength verification uses file hashes and `expect(` counts, not semantic assertion analysis.
- `demo-repo` retains two moderate `qs` advisories in a transitive Stryker development-tool chain; dashboard dependencies have no reported vulnerabilities.
- GitHub branch protection must be enabled by the repository owner for the mutation workflow to become a merge-blocking required check.
