# Coordinator review notes for Claude

Please read before final verification (Codex will also send as a follow-up if this turn finishes first).

- Git now on feat/mvp, main points to the initial documentation commit. No source files have been committed yet, so your changed-source demo baseline is still valid. Do not commit/push.
- Dashboard production build and four artifact tests passed. It strictly recomputes scores/counts when reading sample JSON. Use arrays [] rather than null for files, mutants and changedFiles. Raw statuses as SCHEMA.
- I updated SCHEMA hook example to documented nested Stop/hooks/type:command format, with timeout 240 seconds. Real IBM integration remains unverified locally.
- Review current mutate.Execute on Windows: exec.Command("npx", ...) cannot directly execute npm .cmd shim. Prefer running the installed Stryker JS entrypoint using node, no shell interpolation and no automatic network install. Check actual Stryker CLI config argument format rather than assuming --configFile.
- Ensure report.Read validates values/counts/formula/HEAD; a syntactically valid JSON object with a fabricated score must not pass the gate. Reject missing/unknown statuses and NaN thresholds.
- A failed mutation run must not leave an old successful report eligible for the gate. At minimum record/track latest attempt status and use unique per-run engine report paths. Check locking and cleanup.
- Check --repo demo-repo with parent Git root: config paths and changed-file paths must be relative to target project, not accidentally repository root.
- Please write docs/CLAUDE_PROGRESS.md soon with running status and blockers so progress is visible to both agents.
