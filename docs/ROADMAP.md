# MVP Roadmap

Planning timeline retained for the hackathon. Implementation truth and measured results live in `PROGRESS.md`; unchecked items below include external submission/deployment work and are not evidence that the corresponding code is absent.

## MVP Definition

The MVP is one complete, reproducible loop shown live: Bob finishes a feature with green tests and high coverage; the agent-stop hook fires Gauntlet; the dashboard shows a low Trust Score with surviving mutants; `gauntlet strengthen` sends survivors to Bob; the re-run shows the score above threshold and the CI gate flipping from blocked to pass. Everything else is garnish.

## Phase 0: Pre-Hackathon Prep (Sep 13 to 24, before kickoff)

**Goal**: Zero cold-start work inside the 48 hours.

- [ ] Register on lablab, join Discord, confirm team page (solo).
- [ ] Scaffold repo: Go module, cobra CLI skeleton, Next.js dashboard shell with Tailwind, MIT license, CI stub.
- [ ] Build the demo repo: a small TypeScript "pricing/cart" service (about 8 source files) with deliberately shallow tests (high coverage, weak assertions) so the first Trust Score lands around 35 to 45%.
- [ ] Dry-run StrykerJS on the demo repo locally; record baseline timings; tune to under 3 minutes.
- [ ] Write the Skill and hook JSON as drafts from public Bob docs.
- [ ] Prepare pitch deck skeleton (10 slides) and video shot list.
- [ ] Read Bob 2.0 docs: Hooks, Skills, Shell headless flags, settings.json format.

## 48-Hour Timeline (kickoff Fri Sep 25, 22:00 WIB assumed; adjust to official clock)

### Block A: Hour 0 to 2 — Validation spike (highest risk first)
- [ ] Read the official challenge statement; confirm Gauntlet's framing, adjust pitch wording only.
- [ ] Install Bob, verify: (1) agent-stop hook executes a shell command, (2) headless Shell invocation works, (3) Skill triggers. Record exact flags/keys in NOTES.md.
- [ ] Decision gate: if hooks are unavailable in the hackathon build, switch trigger to manual CLI and rewrite FR-03 language in the pitch. Do not burn more than 2 hours here.

### Block B: Hour 2 to 8 — Mutation runner
- [ ] `gauntlet init` + config loader.
- [ ] Changed-file resolution (git diff vs base + working tree).
- [ ] Stryker config generation, child-process execution, report parsing into the run artifact.
- [ ] `gauntlet report` terminal summary. Milestone: real artifact from the demo repo.

### Block C: Hour 8 to 14 — Bob loop
- [ ] Hook installer writing `.bob/settings.json`.
- [ ] `survivors.md` renderer.
- [ ] `strengthen` command: headless Bob call with Skill, wait, re-run, delta output.
- [ ] End-to-end rehearsal 1: full loop on demo repo. Milestone: score demonstrably rises after Bob strengthens.

### Sleep: Hour 14 to 19. Non-negotiable. A rested demo beats one more feature.

### Block D: Hour 19 to 28 — Dashboard
- [ ] /api/runs routes reading artifacts.
- [ ] Overview page: Trust Score gauge vs coverage bar (the hero visual).
- [ ] Kill matrix page (files x mutants, green/red).
- [ ] Survivor diff viewer.
- [ ] Compare page (before/after strengthen).

### Block E: Hour 28 to 34 — Gate + hardening
- [ ] `gauntlet gate` + GitHub Action; capture a blocked-PR screenshot and a passing one.
- [ ] Error handling: no git repo, empty changed set, Stryker failure, artifact corruption.
- [ ] Commit sample artifacts into `dashboard/sample-runs/`; deploy dashboard to Vercel.

### Sleep: Hour 34 to 38.

### Block F: Hour 38 to 44 — Presentation
- [ ] Record video (script in DEMO.md), max per lablab rules, target 3 minutes.
- [ ] Finish deck (PDF), polish README with architecture diagram and "How IBM Bob 2.0 is used" section listing Hook, Skill, Subagents, headless Shell explicitly.
- [ ] End-to-end rehearsal 2 on a clean clone.

### Block G: Hour 44 to 48 — Submission buffer
- [ ] Submit on lablab: title, descriptions within limits, tech tags, GitHub link, video link, deck.
- [ ] Manual re-check on another device. Never submit in the final 30 minutes.

## Parking Lot (Post-MVP)
- Python (mutmut) and Go (go-mutesting) adapters.
- Bobalytics correlation: Trust Score per model/session cost.
- PR comment bot posting the kill matrix inline.
- Historical trend view and org-level policy files.

## Definition of Done
- Command runs on a clean clone following README only.
- The full loop rehearsed twice without manual intervention beyond documented commands.
- Dashboard deployed and rendering from sample artifacts with no local dependency.
- Repo public, MIT, no secrets, no dead code paths referenced in docs.
