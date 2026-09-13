# Gauntlet build progress

Updated: 2026-09-13. Owners: Codex + Claude Code (`gauntlet-claude`).

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
- [ ] Build and test CLI + demo repository (Claude).
- [x] Build artifact-backed dashboard (Codex): Overview, Matrix, Survivors, Compare and three API routes; awaiting real samples and browser QA.
- [ ] Generate real baseline and strengthened sample runs.
- [ ] Exercise CLI, APIs, dashboard desktop/mobile, and CI gate.
- [ ] Final documentation, dependency inventory, verified commits.

## Open constraints

- IBM Bob executable is not currently on PATH; live Bob hook/headless integration needs actual installation and verified invocation. Do not represent a stub as verified IBM integration.
- User supplied origin https://github.com/Im-A-Nuel/gauntlet.git; read-only remote check found no existing refs. Working branch renamed to main as requested.
- Requested UI skills `impeccable`, `design-taste-frontend`, and `high-end-visual-design` were not found in local Codex/Claude skill directories. Available requested skills: ui-ux-pro-max, antislop, antislop-ui.
- Dashboard direction proposed: dark engineering console, evidence-first layout; awaiting optional user preference.

## Verification log

- Initial workspace: documentation only, no executable project and no Git repository.
- Planning documents are UTF-8; the previous encoding warning was terminal decoding, not established file corruption.
- Dashboard strict typecheck, production build, four artifact tests passed; npm audit reports zero vulnerabilities.
- Updated Next.js to supported 16.3.5; dependencies documented in docs/DEPENDENCIES.md.
- Official IBM docs confirm Stop event/nested hooks and bob run stdin invocation. Live executable absent.
- In-app browser connection failed with runtime metadata error (missing sandboxPolicy); browser QA uses local Playwright/Edge fallback.

## Next checkpoint

Freeze the schema, begin the CLI/demo work in Claude, and build the dashboard against the shared contract.
