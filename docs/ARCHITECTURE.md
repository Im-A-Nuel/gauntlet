# System Architecture

## Overview

Gauntlet is a local-first toolchain: a single Go CLI orchestrates a mutation engine, the IBM Bob 2.0 agent, and JSON artifacts on disk. A Next.js dashboard renders the artifacts. There is no server, database, or auth in the MVP; the file system is the contract between components. This keeps the 48-hour build honest and makes the demo impossible to break by network failure.

## System Diagram

```
                +--------------------------------------+
                |            IBM Bob 2.0               |
                |  Agent mode: writes code + tests     |
                +-----+----------------------+---------+
                      | agent-stop Hook      | headless invocation
                      | (.bob/settings.json) | (strengthen-tests Skill)
                      v                      |
            +---------------------+          |
            |    gauntlet CLI     |<---------+
            |        (Go)         |
            +--+-------+-------+--+
               |       |       |
     changed   |       |       |  gate --min-score
     file set  |       |       |  (exit code)
               v       v       v
        +----------+ +-----------------+ +------------------+
        | git diff | | StrykerJS       | | GitHub Action    |
        | vs base  | | (mutants, runs  | | (PR merge gate)  |
        +----------+ |  tests in       | +------------------+
                     |  parallel x4)   |
                     +--------+--------+
                              v
                   .gauntlet/runs/<id>.json
                              |
                              v
                  +------------------------+
                  |  Next.js dashboard     |
                  |  Trust Score, matrix,  |
                  |  survivor diffs, delta |
                  +------------------------+
```

## Component Responsibilities

### 1. `gauntlet` CLI (Go)
- `init`: writes `.gauntlet/config.yaml`, installs the agent-stop hook into `.bob/settings.json`, copies the `strengthen-tests` Skill into `.bob/skills/`.
- `run [--changed]`: resolves changed files via `git diff --name-only <base>...HEAD` plus working tree, generates a scoped Stryker config, executes Stryker as a child process, normalizes its JSON report into a Gauntlet run artifact.
- `strengthen`: reads survivors from the latest artifact, renders them into a task file, invokes Bob Shell headlessly with the Skill, waits, then re-runs `run --changed` and prints the delta.
- `gate --min-score <n>`: reads latest artifact, exits non-zero below threshold.
- `report`: prints the latest run summary as a terminal table.

### 2. Mutation core (StrykerJS)
Stryker supplies mutant generation, sandboxing, and parallel test execution. Gauntlet never implements mutation operators itself; it generates a per-run `stryker.conf.json` with `mutate` limited to the changed source files and `concurrency: 4`, then parses `reports/mutation/mutation.json`.

### 3. Bob 2.0 integration layer
- **Hook (auto-trigger)**: agent-stop entry in `.bob/settings.json` runs `gauntlet run --changed`, so every Bob session ends with a verdict on its own work.
- **Skill (strengthen-tests)**: a `.bob/skills/strengthen-tests/SKILL.md` instructing Bob to read `.gauntlet/survivors.md`, write the minimal tests that kill each surviving mutant without weakening existing assertions, and run the test suite before finishing.
- **Headless Shell**: `gauntlet strengthen` shells out to non-interactive Bob (Bob Shell v2 automation mode) with a prompt that activates the Skill. Exact flags are confirmed in the hour-1 spike; the fallback is instructing the judge-visible flow through interactive `bob chat` with the same Skill.
- **Subagents**: Bob's own subagent mechanism is exercised inside the strengthen task (the Skill directs Bob to analyze survivors in parallel per file). Gauntlet additionally achieves parallel adversarial execution through Stryker workers; the pitch presents both honestly.

### 4. Dashboard (Next.js 14)
App Router, one API route (`/api/runs`) that reads `.gauntlet/runs/` (path configurable via env), pages: Overview (Trust Score gauge vs coverage bar), Matrix (files x mutants, green/red cells), Survivors (mutation diff viewer), Compare (run A vs run B delta). Recharts for the gauge and history; Tailwind for layout. Demo artifacts are committed under `dashboard/sample-runs/` so the Vercel deployment works standalone.

## Key Design Decisions

- **Wrap Stryker instead of building a mutation engine.** Reason: mutation operators, sandboxing, and test-runner integration are years of work; the 48 hours must go to the Bob loop and the demo. Alternative rejected: AST-level custom mutator in Go (too risky, no time).
- **Changed-files scoping as the core primitive.** Reason: makes runs fast enough to sit inside an agent loop, and it is the honest framing: verify what the AI just did. Alternative rejected: whole-repo runs (slow, kills the demo).
- **File-system artifacts instead of a database.** Reason: zero infra, judge can inspect raw JSON, dashboard trivially portable. Alternative rejected: SQLite (adds nothing at this scale).
- **Go for the CLI.** Reason: single static binary, robust child-process control, the developer's strongest backend language. Alternative rejected: Node CLI (would tangle with Stryker's own Node runtime).
- **Trust Score is always shown next to line coverage.** Reason: the product's entire argument is the gap between the two numbers; the UI must make that gap the hero.

## Security Considerations

- Mutated code executes only inside Stryker's sandbox directory with the project's own test command; Gauntlet never executes mutants against a live environment.
- The Bob hook runs a fixed binary with fixed flags, no user-interpolated shell strings.
- The GitHub Action needs no secrets; it operates on the checked-out repo only.

## Scalability Plan (post-MVP)

Per-language adapters behind the same artifact schema (mutmut for Python, go-mutesting for Go), a queue for monorepo-scale runs, and Bobalytics correlation (Trust Score vs model/cost per session) as the enterprise story.
