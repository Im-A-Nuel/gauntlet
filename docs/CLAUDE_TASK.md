# Claude implementation task

Proceed with implementation now. User authorized full build and collaboration. Codex initialized Git main and an initial docs commit. Read AGENTS.md, PROGRESS.md, and docs/SCHEMA.md including implementation clarifications before coding.

You own cmd/gauntlet, internal/{mutate,bob,report,gate}, go.mod, go.sum, and demo-repo. Write progress in docs/CLAUDE_PROGRESS.md. Do not edit dashboard, shared docs, root README, CI, or Git configuration/history. Codex owns those. No commits, pushes, or AI coauthors.

Toolchains: Go 1.26.2, Node 24, npm 10. Module: github.com/Im-A-Nuel/gauntlet. Implement usable cobra init/run/report/gate/strengthen commands, --repo target, config validation, scoped Git files including worktree/untracked, atomic artifacts, gate checking artifact freshness, Stryker report normalization, score edge cases. Raw killed excludes timeout, score includes both, empty denominator yields null.

IBM Bob is not installed. Implement configurable executable + argument-array headless adapter, honest --prepare-only handoff, source change checks and recursion/reentrancy guard. Public official docs are available: https://bob.ibm.com/docs/ide/configuration/lifecycle-hooks (event is Stop, not agentStop) and https://bob.ibm.com/docs/shell/getting-started/start-bobshell-non-interactive . Verify formats with official documentation. Do not represent untested integration as verified.

Build demo TS pricing/cart service with Vitest and actual Stryker mutation baseline, measured nullable line coverage. Provide separately selectable strengthened tests for reproducible before/after validation without falsely attributing those tests to Bob. Run actual mutations. CLI must write sample artifacts; report paths for Codex to bundle. Keep source file set identical between comparison runs. Never whole-repo mutation fallback.

Write meaningful CLI tests and run all tests. Finish your owned work autonomously, document dependencies and exact commands in your progress, and report paths/results/limitations. Avoid large destructive operations. Restrict work to this project.
