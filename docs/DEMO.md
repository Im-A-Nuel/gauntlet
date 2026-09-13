# Demo Script, Pitch Deck & Submission Checklist

## Video Script (target 3:00)

**0:00 to 0:25 — The lie.**
Screen: demo repo PR, CI green, coverage badge 91%.
Line: "This code was written by AI. So were the tests. Coverage says 91%, CI is green, and none of it proves these tests would catch a single real bug. This is a self-graded exam, and enterprises are merging thousands of these every day."

**0:25 to 0:50 — The reveal.**
Screen: terminal, Bob finishes a task, agent-stop hook fires, `gauntlet run` output scrolls.
Line: "Gauntlet is the adversary. The moment IBM Bob finishes a task, a Bob 2.0 hook triggers Gauntlet, which injects dozens of realistic bugs into exactly the files Bob changed, in parallel, and asks one question: do the tests notice?"

**0:50 to 1:30 — The number.**
Screen: dashboard overview, coverage 91% next to Trust Score 41%, then the kill matrix, then one surviving mutant diff.
Line: "They mostly don't. Trust Score: 41. More than half the injected bugs sailed through a 91%-coverage suite. Here's one: flip this greater-or-equal to greater-than, every test still passes. That's a pricing boundary nobody is actually testing."

**1:30 to 2:15 — The loop.**
Screen: `gauntlet strengthen`, headless Bob session writing tests via the strengthen-tests Skill, re-run, score animates to 95%.
Line: "Gauntlet doesn't just accuse, it fixes. Survivors are handed back to Bob through a custom Skill in headless Bob Shell. Bob writes the precise tests that kill each mutant, Gauntlet re-runs the gauntlet, and the score goes from 41 to 95. AI wrote the code, AI wrote the tests, and AI just proved the tests."

**2:15 to 2:45 — The gate.**
Screen: GitHub PR blocked by gauntlet-gate, then passing after strengthen.
Line: "In CI, Gauntlet is a merge gate: below the Trust Score policy, the PR does not land. This is the number coverage always pretended to be."

**2:45 to 3:00 — Close.**
Screen: logo + architecture strip (Hook, Skill, Subagents, headless Shell badges).
Line: "Gauntlet. Built on IBM Bob 2.0: Hooks for the trigger, Skills and Subagents for the fix, headless Shell for automation. Trust the code, because you tested the tests."

## Pitch Deck (10 slides, PDF)

1. Title: Gauntlet — AI wrote the tests. Who tests the tests?
2. Problem: self-graded exam; 42% of code AI-assisted; coverage measures execution, not verification.
3. Insight: mutation testing is the 40-year-old answer nobody could afford to run; Bob 2.0's parallel agents make it free.
4. Product: Trust Score in one screenshot (coverage vs trust gap).
5. How it works: loop diagram (Bob writes, hook fires, mutants attack, survivors return via Skill, Bob strengthens, gate enforces).
6. Deep Bob 2.0 usage: Hooks, custom Skill, Subagents, headless Shell, settings.json — one slide, explicit, because Application of Technology is the mandatory criterion.
7. Demo results: 41 to 95 on the demo repo, run time under 3 minutes, incremental scoping.
8. Business value: the merge gate that unblocks enterprise AI-coding adoption; fits IBM's governance narrative.
9. What's next: multi-language adapters, Bobalytics correlation, PR bot.
10. Ask/close: repo, live dashboard, one-line thesis.

## lablab Submission Checklist

- [ ] Working prototype accessible online (Vercel dashboard with sample runs + repo instructions for the CLI).
- [ ] Video uploaded (public link), within length rules, MP4.
- [ ] Pitch deck PDF.
- [ ] Public GitHub repo, MIT license, README with "How IBM Bob 2.0 is used" section.
- [ ] Project page: clear title, short/long descriptions within limits, correct tech/category tags, team details.
- [ ] Submitted well before the deadline (manual submission after deadline requires prior organizer approval; do not rely on it).

## Judge Q&A Prep

- "Isn't this just Stryker?" — Stryker is the mutation engine; Gauntlet is the agent loop around it: automatic triggering on agent-stop, AI-driven survivor remediation, and policy gating. Mutation testing failed for 40 years on operational cost; the agent loop is what makes it viable.
- "Why not whole-repo?" — Incremental scoping is the point: verify what the AI just changed, keep runs inside the loop's latency budget.
- "How is this different from Pedigree (Bob 1.0 winner)?" — Pedigree proves who wrote the code; Gauntlet proves the code is actually verified. Provenance vs correctness, complementary layers.
- "What if tests are flaky?" — Timeouts count as killed per Stryker convention; flaky detection is parking lot, and the demo repo is deterministic.
