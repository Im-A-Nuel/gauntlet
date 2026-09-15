# System Architecture

## Overview

Gauntlet is a local-first toolchain: a single Go CLI orchestrates a mutation engine, the IBM Bob agent, and JSON artifacts on disk. A Next.js server serves the dashboard and artifact API; there is no separate backend service, database, or auth. The file system is the contract between components. Bundled reports can be viewed without a live agent or network connection, while live IBM Bob execution still depends on its installed environment and service access.

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
- **Hook (configured trigger)**: a documented `Stop` entry in `.bob/settings.json` runs `gauntlet run --changed`. The installer preserves existing settings. The exact command and `hook` artifact trigger have been exercised directly; automatic lifecycle firing remains a separate observation.
- **Skill (strengthen-tests)**: a `.bob/skills/strengthen-tests/SKILL.md` instructing Bob to read `.gauntlet/survivors.md`, write the minimal tests that kill each surviving mutant without weakening existing assertions, and run the test suite before finishing.
- **Headless Shell**: `gauntlet strengthen` invokes a configurable executable and argument array without shell interpolation. A bounded Bob Shell 2.0.3 run authenticated through `BOB_API_KEY`, received the installed Skill name and survivor reference in its prompt, edited test files, and completed the automatic mutation re-run from 74.5% to 97.9%. `--prepare-only` remains available for environments without Bob service access.
- **Subagents**: the Skill instructs Bob to delegate per-file analysis when the survivor set spans more than three files. The live evidence covered two files and explicitly disabled subagents, so this path remains unexercised. Stryker worker concurrency is separate and verified.

### 4. Dashboard (Next.js 16)
Next.js App Router with three same-origin APIs: `/api/runs`, `/api/runs/:runId`, and `/api/compare`. The reader uses `GAUNTLET_RUNS_DIR` when explicitly configured, otherwise local `.gauntlet/runs`, then bundled recorded samples. Views are Overview (policy and score evidence), Matrix (file mutation strips and dialog inspector), Survivors (searchable outcome list), and Compare (before/after score and per-file delta). Recharts renders the comparison chart; semantic CSS tokens and the Tailwind pipeline own layout and visual styling. Strict schema checks prevent inconsistent artifacts from being displayed.

## Key Design Decisions

- **Wrap Stryker instead of building a mutation engine.** Reason: mutation operators, sandboxing, and test-runner integration are years of work; the 48 hours must go to the Bob loop and the demo. Alternative rejected: AST-level custom mutator in Go (too risky, no time).
- **Changed-files scoping as the core primitive.** Reason: makes runs fast enough to sit inside an agent loop, and it is the honest framing: verify what the AI just did. Alternative rejected: whole-repo runs (slow, kills the demo).
- **File-system artifacts instead of a database.** Reason: zero infra, judge can inspect raw JSON, dashboard trivially portable. Alternative rejected: SQLite (adds nothing at this scale).
- **Go for the CLI.** Reason: single static binary, robust child-process control, the developer's strongest backend language. Alternative rejected: Node CLI (would tangle with Stryker's own Node runtime).
- **Trust Score is always shown next to line coverage.** Reason: the product's entire argument is the gap between the two numbers; the UI must make that gap the hero.

## Security Considerations

- Mutated code executes only inside Stryker's sandbox directory with the project's own test command; Gauntlet never executes mutants against a live environment.
- Bob and Stryker child processes use executable-plus-argument arrays without user-interpolated shell strings.
- The GitHub Action needs no secrets; it operates on the checked-out repo only.

## Scalability Plan (post-MVP)

Per-language adapters behind the same artifact schema (mutmut for Python, go-mutesting for Go), a queue for monorepo-scale runs, and Bobalytics correlation (Trust Score vs model/cost per session) as the enterprise story.
