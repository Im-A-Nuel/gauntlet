# Gauntlet

Historical product brief. For executable setup, measured results, and current integration limitations, use the repository-root README.md and PROGRESS.md. Statistics and target scores in this brief are unverified planning assumptions.

**AI writes the code. AI writes the tests. Gauntlet tests the tests.**

## Overview

Roughly 42% of committed code today is AI-assisted, and in most workflows the test suite guarding that code was written by the same AI in the same session. That is a self-graded exam: CI is green, coverage reads 90%, and none of it proves the tests would actually catch a real bug. Enterprises are merging AI-generated pull requests on the strength of signals that measure nothing.

Gauntlet is an adversarial verification layer for AI-written code, built on IBM Bob 2.0. When Bob finishes a task, Gauntlet automatically unleashes parallel mutant runs against the changed files: each mutant injects one realistic bug (a flipped conditional, an off-by-one, a deleted null check) and re-runs the test suite. Good tests kill mutants. Mutants that survive expose tests that lie. The result is a single honest number, the **Trust Score** (mutation kill rate), which is what coverage pretends to be.

Gauntlet then closes the loop: surviving mutants are handed back to Bob through a custom `strengthen-tests` Skill running in headless Bob Shell, Bob writes tests that kill them, and Gauntlet re-runs until the score clears policy. A GitHub Action gate blocks merges below the configured Trust Score threshold.

Target users: engineering teams adopting AI coding agents who need a merge gate they can defend to security, QA, and compliance.

## Tech Stack

| Layer | Technology |
|---|---|
| Orchestrator CLI | Go 1.22 (`gauntlet` binary) |
| Mutation core | StrykerJS (TypeScript/JavaScript targets) |
| AI agent | IBM Bob 2.0 (IDE + Bob Shell v2 headless), Hooks, Skills, Subagents |
| Dashboard | Next.js 16 (App Router), TypeScript, Tailwind CSS, Recharts |
| Artifacts | JSON files under `.gauntlet/runs/` (no external DB) |
| CI gate | GitHub Actions |
| Hosting (dashboard demo) | Vercel |

## Quick Start

```bash
# 1. Install
go install github.com/<you>/gauntlet/cmd/gauntlet@latest
npm i -D @stryker-mutator/core

# 2. Initialize in a repo (writes .gauntlet/config.yaml, .bob/settings.json hook, Skill)
gauntlet init

# 3. Run against files changed since main
gauntlet run --changed

# 4. Ask Bob to kill the survivors, then re-run
gauntlet strengthen

# 5. Open the dashboard
cd dashboard && npm run dev
```

## Project Structure

```
gauntlet/
  cmd/gauntlet/        # CLI entrypoint
  internal/
    mutate/            # Stryker wrapper + changed-file scoping
    bob/               # Hook installer, headless Bob invocation, Skill templates
    report/            # Trust Score computation, artifact writer
    gate/              # Exit-code policy for CI
  dashboard/           # Next.js app (reads .gauntlet/runs/*.json)
  .bob/
    settings.json      # agent-stop hook -> gauntlet run --changed
    skills/strengthen-tests/
  .github/workflows/gauntlet.yml
  demo-repo/           # Curated TS service used for the demo
```

## Features (MVP)

- Incremental mutation testing scoped to AI-changed files only
- Trust Score with per-file kill matrix and surviving-mutant diffs
- Auto-trigger via Bob 2.0 agent-stop Hook
- Self-healing loop: `strengthen` sends survivors to headless Bob via a custom Skill
- CI merge gate (`gauntlet gate --min-score`)
- Dashboard: score, matrix, mutant diff viewer, before/after run comparison

## License

MIT (lablab.ai submission requirement).
