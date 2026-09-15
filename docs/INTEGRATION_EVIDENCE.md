# Integration verification evidence

Verified on 2026-09-15. This record separates live results from bundled demo
artifacts so the submission does not overstate IBM Bob involvement.

## IBM Bob Shell

- Installed IBM Bob Shell `2.0.3` on Windows using IBM's official PowerShell
  installer from <https://bob.ibm.com/docs/shell/getting-started/install-and-setup>.
- The installer verified package SHA-256
  `c0f65ca2ad1166e5506d42781b474637796fd5f64cdd707d4c0eba7bc800da19`.
- `bob --version` reports version `2.0.3`, commit `9fa8a7ba2`.
- A non-interactive read-only probe was attempted with `bob run`, an explicit
  cost/turn limit, and the repository as its workspace. Bob returned
  `Bob API key is required. Set BOB_API_KEY environment variable.`
- IBM documents API-key authentication as required for headless automation:
  <https://bob.ibm.com/docs/shell/getting-started/start-bobshell-non-interactive>.

Live Bob strengthening is therefore **not yet verified**. The committed
`74.5% -> 97.9%` demo artifacts remain real Stryker results produced by the weak
and strengthened test suites, but they must not be described as output from a
live Bob session until `BOB_API_KEY` is provided and the flow is rerun.

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
