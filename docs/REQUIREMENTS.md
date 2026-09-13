# Requirements

Context: IBM Bob 2.0 Hackathon (lablab.ai), Sep 25 to 27 2026, 48-hour build, solo. Judging criteria assumed from Bob 1.0: Application of Technology (meaningful IBM Bob use, mandatory), Originality, Business Value, Presentation.

## Problem Statement

AI coding agents now produce a large share of committed code and, critically, the tests that guard it. Line coverage and green CI cannot distinguish a rigorous test from a trivial assertion that never fails. Teams therefore merge AI-generated changes on evidence that measures execution, not verification. Mutation testing solves exactly this (inject bugs, check whether tests notice) but has been shelved for decades because it is computationally expensive and painful to operate. Bob 2.0's parallel Subagents, Hooks, and headless Shell remove precisely those barriers.

## Goals & Non-Goals

### Goals (In Scope)
- Compute a Trust Score (mutation kill rate) for the files touched in a Bob session, in under 3 minutes on the demo repo.
- Trigger runs automatically when a Bob agent session ends (agent-stop Hook), with a manual CLI fallback.
- Close the loop: feed surviving mutants back to Bob via a custom Skill in headless mode and demonstrate the score rising after Bob strengthens the tests.
- Enforce a minimum Trust Score in CI via a GitHub Action exit-code gate.
- Present results in a dashboard a judge understands in under 30 seconds.

### Non-Goals (Out of Scope for MVP)
- Multi-language support. TypeScript/JavaScript only (Stryker). Others go to the parking lot.
- Whole-repo mutation runs. Scope is always the changed-file set.
- Persistent server-side storage, auth, multi-tenant SaaS.
- Editing or forking Bob itself. Gauntlet only uses public extension points (Hooks, Skills, Shell, settings).
- Cryptographic attestation of results (Pedigree territory, deliberately avoided).

## Functional Requirements

### FR-01: Incremental mutation run
- Description: The system shall mutate only files changed relative to a base ref (default `main`) and run the project's test command against each mutant.
- Input: repo path, base ref, test command (from `.gauntlet/config.yaml`).
- Output: `.gauntlet/runs/<run-id>.json` artifact with per-mutant status (killed, survived, timeout, no-coverage).
- Priority: High.

### FR-02: Trust Score computation
- Description: The system shall compute Trust Score = killed / (killed + survived) as a percentage, overall and per file, and shall report it alongside line coverage to expose the gap.
- Input: run artifact.
- Output: score fields in the artifact; non-zero CLI summary.
- Priority: High.

### FR-03: Bob agent-stop trigger
- Description: `gauntlet init` shall install an agent-stop hook in `.bob/settings.json` that executes `gauntlet run --changed` when a Bob session ends.
- Input: none (installer).
- Output: patched `.bob/settings.json`; hook fires on session end.
- Priority: High. Fallback: manual `gauntlet run --changed` preserves the full demo if hook behavior differs in the hackathon build.

### FR-04: Strengthen loop
- Description: `gauntlet strengthen` shall serialize surviving mutants (file, line, mutation diff) into a prompt context and invoke headless Bob with the `strengthen-tests` Skill; on completion it shall re-run FR-01 and report the score delta.
- Input: latest run artifact with survivors.
- Output: new tests written by Bob; follow-up run artifact; before/after delta.
- Priority: High.

### FR-05: CI gate
- Description: `gauntlet gate --min-score <n>` shall exit 1 when the latest run's Trust Score is below `n`, and a provided GitHub Actions workflow shall run it on pull requests.
- Input: threshold, latest artifact.
- Output: exit code, annotated summary in the Action log.
- Priority: Medium.

### FR-06: Dashboard
- Description: A Next.js dashboard shall display the latest Trust Score, coverage-vs-trust comparison, a file-by-mutant kill matrix, surviving-mutant diffs, and a before/after comparison between two runs.
- Input: `.gauntlet/runs/*.json` (served via a local API route).
- Output: web UI at `localhost:3000`, deployable to Vercel with bundled demo artifacts.
- Priority: High (this is the demo).

## Non-Functional Requirements

- Performance: a full changed-files run on the demo repo (about 8 source files, about 60 mutants) completes in under 3 minutes on a laptop; mutant test runs execute with parallelism 4.
- Reliability: every CLI command is idempotent and re-runnable; a crashed run never corrupts prior artifacts.
- Portability: single Go binary plus npx Stryker; no Docker required for the demo.
- Demo safety: dashboard renders fully from committed sample artifacts even if no live run has executed.

## Constraints

- 48-hour build window, solo developer, TypeScript/Go stack.
- Bob 2.0 hackathon access is provisioned at kickoff; hook and headless behavior must be validated in hour 1.
- lablab submission rules: public GitHub repo, MIT license, working prototype, video, pitch deck.

## Assumptions

- Bob 2.0 Hooks fire shell commands on agent stop as documented in the changelog, and Bob Shell v2 supports non-interactive invocation.
- The official challenge statement (published at kickoff) stays within "build what's next in AI-assisted development"; Gauntlet fits any reasonable variant. If the statement adds a mandatory element (e.g. watsonx), it will be attached as a summary generator for run reports.
