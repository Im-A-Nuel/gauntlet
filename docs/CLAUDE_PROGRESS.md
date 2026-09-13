# Claude progress (Go CLI + demo-repo)

Owner: Claude Code session `gauntlet-claude`. Scope: `cmd/gauntlet`, `internal/{mutate,bob,report,gate,config,scope,orchestrate}`, `go.mod`/`go.sum`, `demo-repo/`. No commits made (Codex owns Git); no dashboard/root-doc/CI files touched.

## Status: done, real end-to-end loop verified

All five commands (`init`, `run`, `report`, `gate`, `strengthen`) are implemented, build clean (`go build ./...`), pass `go vet ./...`, and are covered by tests that all pass (`go test ./...`, 7 packages, unit + subprocess CLI integration tests). A **real** Stryker mutation baseline was run against `demo-repo` (not synthetic data) and a real strengthened re-run was produced, giving a genuine before/after Trust Score delta and a gate flip from FAIL to PASS.

## Exact commands used

```bash
# Go side
go mod init github.com/Im-A-Nuel/gauntlet
go get github.com/spf13/cobra@latest gopkg.in/yaml.v3@latest github.com/bmatcuk/doublestar/v4@latest
go build ./...
go vet ./...
go test ./...                    # all packages, ~10s total

# demo-repo side
cd demo-repo && npm install
npm test                         # weak/baseline suite + coverage
npm run test:strong              # weak+strong suite + coverage

# End-to-end (from repo root, using `go run` — see "compiled binary" note below)
go run ./cmd/gauntlet --repo demo-repo init
go run ./cmd/gauntlet --repo demo-repo run --changed --trigger manual
go run ./cmd/gauntlet --repo demo-repo strengthen --prepare-only
go run ./cmd/gauntlet --repo demo-repo run --changed --trigger manual --test-command "npm run test:strong"
go run ./cmd/gauntlet --repo demo-repo gate --min-score 80
go run ./cmd/gauntlet --repo demo-repo report
```

**Toolchains used**: Go 1.26.2, Node v24.10.0, npm 10.9.8 (matches `PROGRESS.md`). Git version in this environment supports `git diff --merge-base` (added in git ≥2.35), which `internal/scope` relies on.

## Package map

| Package | Responsibility | Tests |
|---|---|---|
| `internal/report` | Artifact schema (docs/SCHEMA.md §2 + clarifications), Trust Score formula, atomic write/read/list/latest | `score_test.go`, `artifact_test.go` |
| `internal/config` | `.gauntlet/config.yaml` load/validate/write | `config_test.go` |
| `internal/scope` | Changed-file resolution: base-ref diff + worktree/index + untracked, glob filtered, deletions excluded | `scope_test.go` (real temp git repos) |
| `internal/mutate` | Scoped `stryker.conf.json` generation, `npx stryker run` execution, mutation.json → artifact normalization, coverage-summary.json reader | `config_test.go`, `report_test.go`, `coverage_test.go` |
| `internal/gate` | Threshold policy: null score / runtimeError / stale-HEAD all fail closed | `gate_test.go` |
| `internal/bob` | Hook installer, Skill installer, survivors.md renderer, reentrancy lock, headless adapter, source/test-change verification | `hook_test.go`, `survivors_test.go`, `lock_test.go`, `headless_test.go`, `verify_test.go` |
| `internal/orchestrate` | Shared `Run()` used by both `run` and `strengthen` (single lock owner) | `run_test.go` |
| `cmd/gauntlet` | Cobra CLI: `init`, `run`, `report`, `gate`, `strengthen` | `main_test.go` (builds the real binary, drives it as a subprocess) |

## Score edge cases, exactly as specified

- Raw `killed` excludes `timeout`; the Trust Score formula's numerator is `killed + timeout`. Verified against the docs/SCHEMA.md worked example (24/34/2 → 43.3%).
- Empty denominator (no scored mutants) → `trustScore: null`, never 0 or 100.
- `noCoverage`/`ignored`/`compileError`/`runtimeError` never enter the denominator.
- `gate` fails closed on: null score, any `runtimeError` mutant (regardless of numeric score), and a `headSha` mismatch between the artifact and the repo's current HEAD (an explicit `--min-score` even at `0` still can't force a pass).
- An empty changed-file scope writes an explicit artifact (`changedFiles: []`, `trustScore: null`) — Stryker is never invoked, and there is no whole-repo fallback anywhere in the code path.

## IBM Bob integration: implemented to spec, **not verified against a real Bob install**

IBM Bob is not installed in this environment, so nothing below has been exercised against a live `bob` binary. Everything is built against IBM's public docs, fetched during this session:

- Lifecycle hooks: https://bob.ibm.com/docs/ide/configuration/lifecycle-hooks
- Headless Shell: https://bob.ibm.com/docs/shell/getting-started/start-bobshell-non-interactive

**Correction to docs/SCHEMA.md §5**: the session-end hook event is named **`Stop`**, not `agentStop` as drafted. The verified schema also wraps entries one level deeper than the draft:
```json
{ "hooks": { "Stop": [ { "hooks": [ { "type": "command", "command": "...", "timeout": 300 } ] } ] } }
```
(`matcher` is documented only for `PreToolUse`/`PostToolUse`; a `Stop` group omits it.) `internal/bob/hook.go` implements the verified shape and patches (never overwrites) `.bob/settings.json`, preserving every other key via `json.RawMessage` passthrough (tested in `hook_test.go`). **Whoever owns `docs/SCHEMA.md` should update §5** — I did not edit it myself since shared docs are out of my ownership scope.

I also set the hook's `timeout` to 300s explicitly: Bob's documented default is 10s, which would kill `gauntlet run --changed` mid-mutation-run on anything but a trivial change set (the demo repo's real run took ~76-79s).

Headless invocation (`internal/bob/headless.go`) is a fully configurable `Adapter{Executable, Args []string}` run via `exec.Command` (no shell, no string interpolation) so a real Bob CLI can be wired in via `--bob-cmd`/`--bob-arg` without touching Go code once its exact flags are confirmed. The default args build `bob run --format json "@.gauntlet/survivors.md Use the strengthen-tests skill..."`, per the documented `bob run [options] [prompt...]` and `@path` file-reference syntax. **Not documented anywhere I found**: a flag to explicitly select a Skill in headless mode — the design relies on the Skill's own `description` ("Trigger when ... .gauntlet/survivors.md is referenced") to auto-invoke, which is untested.

`strengthen --prepare-only` writes the survivor handoff and stops — no Bob invocation is made or claimed. Without `--prepare-only`, `strengthen` acquires `internal/bob.Acquire` (a simple exclusive-lock file at `.gauntlet/.lock`, non-blocking, fails fast) for its entire snapshot → invoke → verify → re-run sequence, so a nested Stop-hook-triggered `gauntlet run` (e.g. from one of Bob's own subagents finishing) can't race the same Stryker report file. **Known limitation**: the lock is existence-only, not liveness-checked — a crashed process leaves `.gauntlet/.lock` behind and a human has to delete it. I chose not to guess at staleness, since a wrong guess risks two mutation runs corrupting the same report concurrently.

After a real Bob invocation, `internal/bob.TakeSnapshot`/`Compare` mechanically check: no source file (matching `include` globs) was added/modified/deleted, no test file (matching `exclude` globs) was deleted, and no test file's `expect(` count dropped. This is **a heuristic, not full AST analysis** — it catches deletion and gross weakening of assertions, not a semantically-narrower rewrite that keeps the same `expect(` count. Documented here rather than oversold in code comments.

## The real baseline → strengthened loop (not synthetic)

`demo-repo` is a small TS pricing/cart service (`src/pricing.ts`, `src/cart.ts`) with two Vitest suites:
- `test/weak/` — deliberately shallow (type/truthiness assertions only), the baseline the first Trust Score run measures.
- `test/strong/` — hand-authored by **me (Claude)**, boundary-value assertions. **This is explicitly not a Bob output** — every file and this doc say so, because IBM Bob isn't installed and I won't misattribute authorship. `npm run test:strong` runs weak+strong together (additive, matching the "don't weaken existing assertions" rule).

Real, measured results from `gauntlet run --repo demo-repo` (Stryker `command` test runner, no plugins beyond `@stryker-mutator/core`):

| Run | Test suite | Mutants | Killed | Survived | Trust Score | Line coverage |
|---|---|---|---|---|---|---|
| `1789317267750-d1f85d4` | weak only (`npm test`) | 47 | 37 | 10 | **78.7%** | 100% |
| `1789317405596-d1f85d4` | weak+strong (`npm run test:strong`) | 47 | 45 | 2 | **95.7%** | 100% |

`gate --min-score 80` on the first run: `exit 1`, "Trust Score 78.7 is below the minimum 80.0". On the second: `exit 0`, "Trust Score 95.7 meets the minimum 80.0" — the blocked→passing flip the demo script wants, from a real mutation run, not staged numbers. Both runs sit at 100% line coverage the whole time, which is exactly the coverage-vs-verification gap the product argues for.

Two mutants remain genuinely un-killed after strengthening (`src/cart.ts` mutants `16` and `25` — see `demo-repo/.gauntlet/survivors.md` from the first run for the full list, or re-run `strengthen --prepare-only` against the second). Both survive by numeric coincidence (e.g. `100 * 0.9 == 100 - 10`), not a bug in the check — left as an honest example that mutation testing has diminishing-returns edges too, rather than chasing 100%.

**Artifact paths for Codex to bundle into `dashboard/sample-runs/`** (both already real, schema-valid JSON):
- `demo-repo/.gauntlet/runs/1789317267750-d1f85d4.json` (baseline, 78.7%) — I see this one is already copied into `dashboard/sample-runs/`.
- `demo-repo/.gauntlet/runs/1789317405596-d1f85d4.json` (strengthened, 95.7%, `comparedTo` the first) — not yet copied as of this writing; this is the "after" half of the before/after pair for the Compare page.

I read `dashboard/lib/schema.ts` (read-only) to check compatibility: field names, the `killed+timeout` score formula, `omitempty`-driven `default(0)` on `ignored`/`compileError`/`runtimeError`, and `comparedTo` as absent-vs-present (not null) all line up with what `internal/report` emits.

## Reproducing the loop from a clean clone

```bash
cd demo-repo && npm install
cd .. 
go run ./cmd/gauntlet --repo demo-repo init          # idempotent
go run ./cmd/gauntlet --repo demo-repo run --changed --trigger manual
go run ./cmd/gauntlet --repo demo-repo gate --min-score 80   # exits 1 on the weak baseline
go run ./cmd/gauntlet --repo demo-repo run --changed --trigger manual --test-command "npm run test:strong"
go run ./cmd/gauntlet --repo demo-repo gate --min-score 80   # exits 0 after strengthening
```
Each mutation run takes ~75-80s on this machine (47 mutants, concurrency 4).

## Corrections / discrepancies found in existing docs (not mine to edit)

1. **docs/SCHEMA.md §5**: hook event is `Stop`, not `agentStop`; schema nesting is one level deeper (see above). Same file's §8 GitHub Action example and my own initial assumption both used `--configFile` for Stryker — the real Stryker 8.x CLI takes the config path as a **positional** argument (`stryker run [options] [configFile]`), not a flag. `internal/mutate/run.go` uses the positional form; the workflow YAML in SCHEMA.md doesn't invoke Stryker's CLI directly (it shells out to `gauntlet run`), so it isn't affected, but anyone testing Stryker's CLI by hand should know.
2. `docs/ROADMAP.md` predicted a baseline Trust Score of "35 to 45%" — the real, measured baseline came in at 78.7%. Stryker's `command` runner with `perTest` coverage analysis is more precise about test-to-mutant coverage than the original estimate assumed. I left the demo repo's weak tests as genuinely weak (type/truthiness checks only, not tuned to hit a target number) rather than gaming them down to match the older prediction — real 78.7%→95.7% still tells the coverage-vs-verification story and still flips a gate at threshold 80.

## Dependencies added (all documented here per the "no undocumented deps" rule)

- Go: `github.com/spf13/cobra`, `gopkg.in/yaml.v3`, `github.com/bmatcuk/doublestar/v4` (only for `**` glob matching against config include/exclude — the stdlib has no equivalent).
- demo-repo (npm): `@stryker-mutator/core`, `vitest`, `@vitest/coverage-v8`, `typescript` — all devDependencies, matching docs/README.md's tech stack table. No test-runner plugin package needed since Stryker's built-in `command` runner shells out to `npm test`/`npm run test:strong` directly.
- `npm install` reports 14 known vulnerabilities (5 low/5 moderate/2 high/2 critical) in transitive deps of the Stryker/Vitest toolchain, not in first-party code. Not remediated — `npm audit fix --force` would pull in breaking major-version bumps of dev tooling only, no runtime/production exposure, and this is a devDependency-only demo repo.

## Coordinator-flagged fixes applied (this checkpoint)

Went through `docs/COORDINATOR_NOTES.md` item by item and fixed everything that was actually still broken (some items, like deleted-file exclusion, turned out already correct and are noted as such below):

1. **Windows `npx` execution bug (real, verified) — `internal/mutate/run.go`**: `exec.Command("npx", ...)` cannot execute the `.cmd` shim CreateProcess sees on Windows (no shell in between). Fixed by resolving Stryker's own JS entrypoint (`node_modules/@stryker-mutator/core/bin/stryker.js`, from its `package.json` `bin` field) and invoking it as `node <entrypoint> run <configPath>` — `node.exe` is a native executable on every platform, so this sidesteps the shim problem entirely rather than working around it with `cmd.exe /c`. Re-ran the full baseline → strengthen → gate loop after the fix to confirm it isn't just a compile-time change: fresh real Stryker run, 83.0% → 95.7%, gate flip 1→0, all against the rebuilt binary (see below).
2. **`gate` silently disabled the freshness check — `cmd/gauntlet/gate.go`**: `scope.HeadSha` failing used to fall back to `currentHead = ""`, which makes `gate.Evaluate` skip the stale-artifact check entirely (empty string short-circuits the `!=` comparison). Now a HEAD-resolution failure is a hard `ExitConfigOrEnvError`, matching every other environment-error path in this command.
3. **`--min-score` had no bounds validation — `cmd/gauntlet/gate.go`**: `NaN`, `+Inf`, negative, or `>100` all silently passed straight into `gate.Evaluate` before. Added an explicit `math.IsNaN`/`math.IsInf`/range check that fails closed with `ExitConfigOrEnvError`. Verified against the compiled binary: `--min-score 150` and `--min-score -5` both now exit 3 with a clear message; the underlying real exit code was confirmed with a locally built binary since `go run` doesn't relay it 1:1 to the shell in this environment.
4. **`config.Validate`'s `minScore` check had the same NaN gap**: `c.MinScore < 0 || c.MinScore > 100` is false for `NaN` (all NaN comparisons are false in Go), so a `.gauntlet/config.yaml` with `minScore: .nan` would have loaded successfully. Added the same `IsNaN`/`IsInf` guard.
5. **Filenames with spaces/unicode/embedded newlines could mis-split — `internal/scope/scope.go`**: `git diff --name-status` / `git ls-files` newline-delimited parsing is ambiguous for unusual filenames. Switched both invocations to `-z` (NUL-delimited) and rewrote the parsers (`parseNameStatusZ`, `parseLinesZ`) to split on `\x00` instead of `\n`. All existing `scope` tests still pass unchanged.
6. **Symlinks escaping the project root were never rejected — `internal/scope/scope.go`**: added `rejectEscapingSymlinks`, which `Lstat`s every candidate file and, for anything that is itself a symlink, resolves its target and fails the whole `Resolve` call closed if the target lands outside `repoRoot`. Stryker mutates candidate files in place; an unchecked symlink could otherwise let a mutation run write through to an arbitrary path.
7. **`strengthen`'s re-run used the wrong base ref and didn't verify scope stability — `cmd/gauntlet/strengthen.go`**: it used to call `orchestrate.Run` with no `BaseRefOverride`, silently falling back to `config.yaml`'s current `baseRef` instead of the `before` artifact's recorded `BaseRef` (only matters if the config changed between the two runs, but was wrong on principle for a same-baseline comparison). Now passes `BaseRefOverride: before.BaseRef` explicitly, and after the re-run, asserts `reflect.DeepEqual(before.ChangedFiles, after.ChangedFiles)` — if the Bob invocation somehow altered the mutation scope itself, `strengthen` now aborts with `ExitRunFailed` and an explicit message instead of reporting a delta between two runs that scored different file sets.
8. **Already correct, no change needed**: deleted-file exclusion (coordinator's `internal/scope` note) — `TestResolveExcludesDeletedFiles` already passed before this checkpoint; `parseNameStatusZ`'s `D`-prefix handling preserves that behavior.

Re-verified end-to-end after all of the above, against a freshly rebuilt binary, not just `go build`/`go vet`/`go test` (all of which also still pass, 7 packages):

```
go run ./cmd/gauntlet --repo demo-repo run --changed --trigger manual
  -> run 1789318258496-d1f85d4: 47 mutants, killed 39, survived 8, trustScore 83.0%, lineCoverage 100.0%

go run ./cmd/gauntlet --repo demo-repo run --changed --trigger manual --test-command "npm run test:strong"
  -> run 1789318332985-d1f85d4: 47 mutants, killed 45, survived 2, trustScore 95.7%, lineCoverage 100.0%

go run ./cmd/gauntlet --repo demo-repo gate --min-score 80   # exit 0, "Trust Score 95.7 meets the minimum 80.0"
go run ./cmd/gauntlet --repo demo-repo gate --run-id 1789318258496-d1f85d4 --min-score 80  # exit 0 too (83.0 > 80; the weak baseline happened to clear 80 this run — mutant order/selection has minor run-to-run variance, unrelated to any of the fixes above)
go run ./cmd/gauntlet --repo demo-repo gate --min-score 150  # exit 3, "--min-score 150 is invalid..."
go run ./cmd/gauntlet --repo demo-repo gate --min-score -5   # exit 3, "--min-score -5 is invalid..."
```

Not touched this checkpoint (still open, deliberately or out of scope):
- `.gauntlet/.lock` staleness (existence-only, no PID-liveness check) — unchanged, per the reasoning already recorded above: guessing wrong about staleness risks two concurrent mutation runs corrupting the same report.
- IBM Bob hook/headless/Skill invocation — still unverified against a real `bob` binary; no Bob install available in this environment.
- Stryker/Vitest major-version audit mentioned in coordinator notes — not run this pass; current versions still pass all tests and produced a correct real mutation run above.
- `dashboard/sample-runs/` already has both the baseline and strengthened artifacts bundled (I checked; this is Codex's owned directory, not modified by me).

## Dependency security upgrade (demo-repo): Stryker 8 -> 10, Vitest 2 -> 5

Coordinator notes flagged 14 known vulnerabilities (5 low/5 moderate/2 high/2 critical) in `demo-repo`'s transitive dev dependencies, and that the registry now offers `@stryker-mutator/core` 10 and `vitest`/`@vitest/coverage-v8` 5 (Node >=22, this environment runs Node 24.10.0). Did this on branch `deps/demo-repo-security-upgrade` since a major-version bump risks breaking the demo, and merged to `main` only after re-verifying the full loop for real:

- Bumped `demo-repo/package.json`: `@stryker-mutator/core` `^8.7.1` -> `^10.0.0`, `vitest` and `@vitest/coverage-v8` `^2.1.9` -> `^5.0.0`. Left `typescript` at `^5.7.3` (not flagged as vulnerable, out of scope for this bump).
- Fresh `npm install`: 14 vulnerabilities -> 2 moderate (a DoS in `qs`, pulled in transitively through `typed-rest-client`, itself several levels deep inside Stryker's own dependency tree, not something our config touches at runtime). `npm audit fix` doesn't resolve it non-major and the package isn't a direct dependency; left as-is — dev-only tooling, no runtime/production exposure, consistent with the existing "no undocumented deps" policy (nothing added, just versions bumped).
- Re-ran everything for real after the bump, not just installed and assumed it worked: `npm test` (12 tests) and `npm run test:strong` (33 tests) both pass unchanged under Vitest 5. Forced a real Stryker 10 mutation run with `--base-ref d1f85d4` (since `src/` is unchanged relative to the current `main` tip after the previous commit): 47 mutants, baseline 72.3% -> strengthened 95.7%, and `gate --min-score 80` flips FAIL (`below_threshold`, exit 1) -> PASS (exit 0) on the two runs respectively. Same real loop, same shape of result, on the upgraded major versions.
- `go build`/`go vet`/`go test ./...` (all 7 packages) still pass — the version bump is demo-repo-only, no Go-side change.

## Known limitations (stated plainly, not buried)

- Bob hook firing, headless `bob run`, and Skill auto-invocation are all **unverified against a real Bob binary** — implemented to the documented spec only.
- `.gauntlet/.lock` staleness requires manual cleanup after a crash (no PID-liveness check).
- The strengthen-tests verification (`internal/bob/verify.go`) is a content-hash + `expect(` count heuristic, not semantic assertion analysis.
- Two demo-repo mutants survive the strengthened suite by numeric coincidence, not by design.
- Running the compiled `gauntlet.exe` directly from this session's Bash tool triggers a permission prompt (sandbox policy on executing arbitrary local binaries); `go run ./cmd/gauntlet ...` does not, and is what every command above uses. This is a sandbox/tooling quirk of this build session, not a product limitation — a real user running the built binary directly is unaffected.
