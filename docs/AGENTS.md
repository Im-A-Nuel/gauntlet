# AGENTS.md

Context file for AI coding agents (IBM Bob, Claude Code) working on this repo. Keep under 200 lines. Update "Current Focus" when switching blocks.

## Project
Gauntlet — adversarial verification for AI-written code: mutation-based Trust Score, auto-triggered by IBM Bob 2.0 hooks, self-healing via a strengthen-tests Skill, enforced as a CI gate.

## Stack
- CLI: Go 1.22, cobra. Packages under `internal/` (mutate, bob, report, gate).
- Mutation engine: StrykerJS via child process. We NEVER implement mutation operators ourselves.
- Dashboard: Next.js 16 App Router, TypeScript, Tailwind, Recharts. Reads JSON artifacts only. Framework updated to a supported release during implementation.
- No database. Artifacts in `.gauntlet/runs/`. Schemas in `docs/SCHEMA.md` are the source of truth.

## Commands
```bash
go build ./cmd/gauntlet          # build CLI
go test ./...                    # CLI tests
gauntlet run --changed           # mutation run on changed files
gauntlet strengthen              # Bob loop on survivors
cd dashboard && npm run dev      # dashboard on :3000
cd demo-repo && npm test         # demo repo test suite
```

## Project Structure
```
cmd/gauntlet/        CLI entrypoint (cobra)
internal/mutate/     changed-file scoping, Stryker config gen + report parsing
internal/bob/        hook installer, Skill template, headless Bob invocation
internal/report/     artifact read/write, Trust Score math
internal/gate/       threshold policy
dashboard/           Next.js app (+ sample-runs/ committed demo artifacts)
demo-repo/           curated TS service for the demo (weak tests on purpose)
docs/                REQUIREMENTS, ARCHITECTURE, SCHEMA, ROADMAP, DEMO
```

## Key Conventions
- Artifact fields and exit codes exactly as `docs/SCHEMA.md`; change the doc first, then code.
- Go: table-driven tests, wrap errors with `%w`, no global state outside `internal/report` cache.
- TypeScript: strict mode, no `any` in dashboard code.
- Commits: `feat|fix|docs(scope): message`.
- Trust Score formula: killed / (killed + survived); timeouts count as killed; noCoverage excluded from denominator. Never redefine.

## Architecture Notes
- The dashboard must render fully from `dashboard/sample-runs/` with no CLI present. Never fetch anything remote.
- `gauntlet init` PATCHES `.bob/settings.json`; never overwrite user hooks.
- Stryker runs are always scoped: pass explicit `mutate` file lists; never mutate the whole repo.
- Headless Bob invocation lives in `internal/bob/headless.go` behind a configurable executable and argument array.

## Do NOT
- Do not add a database, auth, or Docker.
- Do not modify files in `demo-repo/src` when strengthening tests; test files only.
- Do not delete or weaken existing assertions to raise the kill rate.
- Do not install new dependencies without listing them in the PR/commit description.
- Do not touch `.gauntlet/runs/` by hand; artifacts are written by the CLI only.

## Current Focus
Codex-owned dashboard, shared docs, integration QA, samples, and CI definitions are complete. Remaining CLI hardening and live IBM Bob verification are listed in `PROGRESS.md` and `docs/COORDINATOR_NOTES.md`.
