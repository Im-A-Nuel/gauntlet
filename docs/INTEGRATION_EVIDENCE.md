# Integration verification evidence

Verified on 2026-09-15. This record separates live results from bundled demo
artifacts so the submission does not overstate IBM Bob involvement.

## IBM Bob Shell

- Installed IBM Bob Shell `2.0.3` on Windows using IBM's official PowerShell
  installer from <https://bob.ibm.com/docs/shell/getting-started/install-and-setup>.
- The installer verified package SHA-256
  `c0f65ca2ad1166e5506d42781b474637796fd5f64cdd707d4c0eba7bc800da19`.
- `bob --version` reports version `2.0.3`, commit `9fa8a7ba2`.
- A bounded non-interactive authentication probe returned `BOB_AUTH_OK` with
  task ID `2776a622f07eaa7c267b78a2de6529b1`. The API key remained in the
  Windows user environment and was not printed, written to the repository, or
  passed as a command-line argument.
- The hook command was exercised directly as
  `gauntlet run --changed --base-ref 5b4762e --trigger hook`. It scoped Stryker
  to `src/cart.ts` and `src/pricing.ts`, generated 47 mutants, killed 35, left
  12 survivors, and recorded a `74.5%` Trust Score.
- `gauntlet strengthen` invoked Bob headlessly with a `$2` cost cap, a 25-turn
  cap, MCP and subagents disabled, and the generated survivor handoff. Bob task
  `31120d025cc7d443b941740a971f1606` completed in 80,292 ms with 14 tool calls
  and a recorded session cost of `$0.406046`.
- Bob added assertions only to `demo-repo/test/weak/cart.test.ts` and
  `demo-repo/test/weak/pricing.test.ts`. It did not modify either source file.
  The weak suite grew from 12 to 24 passing tests.
- Gauntlet's automatic post-Bob mutation run used the same 47-mutant scope. It
  killed 46, left 1 survivor, recorded `97.9%`, and linked the result to the
  `74.5%` run through `comparedTo`. `gauntlet gate --min-score 80` passed.

Evidence-only PR <https://github.com/Im-A-Nuel/gauntlet/pull/3> preserves the
Bob-authored test diff, the CLI-generated survivor handoff, and both generated
run artifacts. It is intentionally not merged because the demo on `main` keeps
the deliberately weak baseline suite.

Artifact integrity details:

| Run | Trigger | Result | SHA-256 |
|---|---|---|---|
| `1789485104500-2077f42` | `hook` | 35 killed, 12 survived, `74.5%` | `DD67834A9E05F30469A4FA97C0208A9EA598C3C5267437A0F56335F86887534A` |
| `1789485282017-2077f42` | `strengthen` | 46 killed, 1 survived, `97.9%` | `DABE8FF31AC98E036802ADB77F811284CEFBF70D6BF12D0465CA9C0ED7F526C2` |

This verifies headless Bob authentication, the Skill-guided survivor handoff,
additive test writing, Gauntlet's file-change guard, the post-Bob mutation
re-run, and the gate flip. Automatic firing of the configured Bob `Stop`
lifecycle hook was not separately observed; its command was run directly with
a `hook` trigger for this evidence.

## Required GitHub gate

Protection is active on `main` with these settings:

- required status check: `gauntlet-gate`;
- strict/up-to-date check: enabled;
- enforcement for administrators: enabled;
- force pushes: disabled;
- branch deletion: disabled.

The workflow runs on every pull request. Mutation steps run only when
`demo-repo/src/**` changed, so unrelated pull requests still receive a completed
required check instead of remaining pending.

### Controlled failing and passing proof

Evidence PR: <https://github.com/Im-A-Nuel/gauntlet/pull/1>

1. Failing revision `6148453` used a tiered shipping policy without its boundary
   assertions. The GitHub artifact recorded 69 mutants: 59 killed, 10 survived,
   Trust Score `85.5%`. The verification threshold was `95%`, the
   `gauntlet-gate` job failed, and GitHub reported the PR as blocked.
   Run: <https://github.com/Im-A-Nuel/gauntlet/actions/runs/34934864903>
2. Passing revision `ff4950e` added assertions at and immediately around every
   shipping tier boundary, plus a case distinguishing percent and flat coupons.
   The GitHub artifact recorded 69 mutants: 69 killed, 0 survived, Trust Score
   `100%`. The same `gauntlet-gate` job passed and GitHub reported the PR as
   clean/mergeable.
   Run: <https://github.com/Im-A-Nuel/gauntlet/actions/runs/34935098922>

The verification PR is evidence-only and is not intended for merge.
