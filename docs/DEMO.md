# Demo Script, Pitch Deck & Submission Checklist

Implementation note: the numerical results below now have two independent sources. The bundled Stryker 10 samples remain reproducible without Bob. A live Bob Shell 2.0.3 run also produced the same 74.5% to 97.9% result; PR #3 preserves Bob's test diff and both generated artifacts.

## Video Script (target 3:00)

**0:00 to 0:25 — The lie.**
Screen: weak demo tests passing with 100% measured line coverage.
Line: "The tests are green and every line ran. That still does not prove the assertions would catch a behavioral regression. Coverage measures execution, not verification."

**0:25 to 0:50 — The reveal.**
Screen: installed `.bob/settings.json` Stop hook, then the manual fallback `gauntlet run --changed` output.
Line: "Gauntlet installs a Bob Stop hook and also exposes the same run manually. It asks Stryker to mutate only the changed source files, then records which changes the tests detect."

**0:50 to 1:30 — The number.**
Screen: dashboard overview, 100% line coverage next to Trust Score 74.5%, then the kill matrix and one surviving mutant diff.
Line: "Line coverage is 100%, but Trust Score is only 74.5 and the merge policy fails. Twelve of 47 changes went unnoticed. Each survivor points to a behavior the assertions did not pin down."

**1:30 to 2:15 — The loop.**
Screen: the generated `survivors.md`, live `gauntlet strengthen` output, Bob's additive test diff, and Compare changing from 74.5% to 97.9%.
Line: "Gauntlet sends twelve survivors to its Bob Skill. Bob adds precise boundary and exact-value assertions without touching source, then Gauntlet attacks the same 47 mutants again. Forty-six are killed, one survives, and Trust Score reaches 97.9%."

**2:15 to 2:45 — The gate.**
Screen: baseline gate returning exit 1, follow-up gate returning exit 0, then the GitHub workflow definition.
Line: "The policy is executable: 74.5 fails at the 80 threshold and 97.9 passes. The required GitHub check is active on main; PR #1 records a blocked revision and the passing revision after its tests were strengthened."

**2:45 to 3:00 — Close.**
Screen: logo and architecture strip: Bob Stop hook, Skill handoff, headless adapter, Stryker workers, artifact, dashboard, gate.
Line: "Gauntlet connects Bob's extension points to adversarial evidence: trigger, strengthen, re-run, inspect, enforce. Trust the tests because you tested the tests."

## Pitch Deck (10 slides, PDF)

1. Title: Gauntlet — AI wrote the tests. Who tests the tests?
2. Problem: self-graded exam; 42% of code AI-assisted; coverage measures execution, not verification.
3. Insight: mutation testing is the 40-year-old answer nobody could afford to run; Bob 2.0's parallel agents make it free.
4. Product: Trust Score in one screenshot (coverage vs trust gap).
5. How it works: loop diagram (Bob writes, hook fires, mutants attack, survivors return via Skill, Bob strengthens, gate enforces).
6. Deep Bob 2.0 usage: Hooks, custom Skill, Subagents, headless Shell, settings.json — one slide, explicit, because Application of Technology is the mandatory criterion.
7. Demo results: 74.5 to 97.9 on the same 47 mutants, 100% line coverage throughout, each run under one minute locally.
8. Business value: the merge gate that unblocks enterprise AI-coding adoption; fits IBM's governance narrative.
9. What's next: multi-language adapters, Bobalytics correlation, PR bot.
10. Ask/close: repo, live dashboard, one-line thesis.

## lablab Submission Checklist

- [ ] Working prototype accessible online (Vercel dashboard with sample runs + repo instructions for the CLI).
- [ ] Video uploaded (public link), within length rules, MP4.
- [ ] Pitch deck PDF.
- [x] Public GitHub repo, MIT license, README with "How IBM Bob 2.0 is used" section.
- [ ] Project page: clear title, short/long descriptions within limits, correct tech/category tags, team details.
- [ ] Submitted well before the deadline (manual submission after deadline requires prior organizer approval; do not rely on it).

## Judge Q&A Prep

- "Isn't this just Stryker?" — Stryker is the mutation engine; Gauntlet is the agent loop around it: automatic triggering on agent-stop, AI-driven survivor remediation, and policy gating. Mutation testing failed for 40 years on operational cost; the agent loop is what makes it viable.
- "Why not whole-repo?" — Incremental scoping is the point: verify what the AI just changed, keep runs inside the loop's latency budget.
- "How is this different from Pedigree (Bob 1.0 winner)?" — Pedigree proves who wrote the code; Gauntlet proves the code is actually verified. Provenance vs correctness, complementary layers.
- "What if tests are flaky?" — Timeouts count as killed per Stryker convention; flaky detection is parking lot, and the demo repo is deterministic.
