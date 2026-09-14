# Gauntlet build progress

Updated: 2026-09-14. Owners: Codex + Claude Code (`gauntlet-claude`).

## Working agreement

- Build the documented local-first MVP: scoped Stryker runs, Go CLI, Bob handoff, artifact dashboard, CI gate.
- Codex owns schema coordination, dashboard, integration QA, root documentation, and Git.
- Claude owns `cmd/`, `internal/`, `go.mod`, `go.sum`, `demo-repo/`, and `docs/CLAUDE_PROGRESS.md`.
- Keep ownership boundaries; communicate interface changes before editing another owner's files.
- All commits use Im-A-Nuel and the existing configured email. No AI coauthor trailers. Push only to a verified project remote.
- Dependencies must be documented. No database, auth, or Docker.

## Milestones

- [x] Inspect the seven planning documents and available toolchains (Go 1.26.2, Node 24.10.0, npm 10.9.8).
- [x] Resume the named Claude session and agree on ownership.
- [x] Clarify score/artifact contract and initialize Git (initial commit 5b4762e, branch feat/mvp).
- [x] Build and test CLI + demo repository (Claude): five CLI commands, eight Go packages, weak and strong Vitest suites.
- [x] Build artifact-backed dashboard (Codex): Overview, Matrix, Survivors, Compare, three API routes, explicit loading/error/empty states, and strict artifact validation.
- [x] Generate real Stryker 10 sample runs at revision `450d975`: 74.5% baseline and 97.9% strengthened.
- [x] Exercise CLI, APIs, dashboard interactions, accessibility, and responsive layouts locally; verify the policy gate flips from exit 1 to exit 0.
- [x] Finish shared documentation, dependency inventory, local QA, and CI workflow definitions.

## Open constraints

- IBM Bob executable is not currently on PATH; live Bob hook/headless integration needs actual installation and verified invocation. Do not represent a stub as verified IBM integration.
- User supplied origin https://github.com/Im-A-Nuel/gauntlet.git; read-only remote check found no existing refs. Working branch renamed to main as requested.
- Requested UI skills `impeccable`, `design-taste-frontend`, and `high-end-visual-design` were not found in local Codex/Claude skill directories. Available requested skills: ui-ux-pro-max, antislop, antislop-ui.
- Dashboard uses the dark engineering inspection-report direction in DESIGN.md and the supplied monochrome visual references.
- Two CLI hardening findings remain in Claude's ownership: `strengthen` error exits can bypass deferred lock cleanup, and gate-side artifact reads need integrity validation before policy evaluation. See `docs/COORDINATOR_NOTES.md`.
- `demo-repo` has two moderate transitive `qs` advisories through Stryker's development-only dependency chain. Dashboard audit is clean.

## Verification log

- Initial workspace: documentation only, no executable project and no Git repository.
- Planning documents are UTF-8; the previous encoding warning was terminal decoding, not established file corruption.
- Dashboard strict typecheck, production build, four artifact tests, two Playwright suites, and axe WCAG A/AA checks pass; npm audit reports zero vulnerabilities.
- Updated Next.js to supported 16.3.5; dependencies documented in docs/DEPENDENCIES.md.
- Official IBM docs confirm Stop event/nested hooks and bob run stdin invocation. Live executable absent.
- In-app browser connection failed with runtime metadata error (missing sandboxPolicy); browser QA uses local Playwright/Edge fallback.
- Local Edge exercised report selection, clipboard feedback, mutant dialog/Escape handling, filters, comparisons, and routes at 375, 768, 1024, and 1440 px with no horizontal overflow.
- CLI Go tests and go vet passed at the intermediate checkpoint. Integration review findings are recorded in docs/COORDINATOR_NOTES.md, including lock cleanup and stale-artifact policy.
- Fresh Stryker 10 runs produced 47 mutants against the same two source files: weak tests killed 35 (74.5%, gate FAIL), strong tests killed 46 (97.9%, gate PASS), while line coverage remained 100%.
- Dashboard checkpoint `d1f85d4` and CLI/dependency checkpoints through `450d975` are on `origin/main`; all commits use only Im-A-Nuel as author and committer.

## Next checkpoint

Codex-owned implementation is complete in this checkpoint. When Claude work resumes, close the two CLI hardening findings above and verify the Bob hook/headless flow on a machine with IBM Bob installed.
