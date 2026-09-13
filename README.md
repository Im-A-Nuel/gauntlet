# Gauntlet

Test the tests guarding AI-written code.

Gauntlet runs StrykerJS mutations against changed TypeScript/JavaScript files, records which changes the tests detect, and presents a Trust Score beside measured line coverage. Surviving mutations become a handoff for IBM Bob to strengthen tests. A CLI gate checks the resulting evidence against a threshold.

## Dashboard

Requires Node.js 24 and npm. The dashboard reads local JSON artifacts and supports bundled recorded demo runs without an agent or external service.

```sh
cd dashboard
npm ci
npm run dev
```

Open http://127.0.0.1:3000. Overview shows the policy result; Kill matrix opens individual mutant diffs; Survivors filters outcomes; Compare runs shows before/after evidence.

To read a project outside this repository, set `GAUNTLET_RUNS_DIR` to its absolute `.gauntlet/runs` directory. An explicitly configured missing or corrupt directory produces an error rather than substituting demo data.

```powershell
$env:GAUNTLET_RUNS_DIR = 'F:\path\to\project\.gauntlet\runs'
npm run dev
```

## Score semantics

Raw killed, timeout and survived counts are disjoint. Trust Score is `(killed + timeout) / (killed + timeout + survived) × 100`, rounded to one decimal. No-coverage mutants are reported separately and excluded from this denominator. No scored mutants means no score, never an automatic pass. A high score does not establish correctness; review uncaught behavior, no-coverage cases, timeouts and equivalent mutants.

Line coverage comes from a measured coverage report or is displayed as unavailable. It is never inferred from the mutation score.

## IBM Bob integration

Gauntlet uses the documented `Stop` lifecycle hook, a `strengthen-tests` Skill, and a configurable non-interactive Bob command. Bob is not bundled. The actual IBM executable was unavailable on the development machine; live agent execution remains a separate verification step. Preparing a survivor handoff does not claim the agent ran.

Official references: [Bob lifecycle hooks](https://bob.ibm.com/docs/ide/configuration/lifecycle-hooks), [Bob non-interactive sessions](https://bob.ibm.com/docs/shell/getting-started/start-bobshell-non-interactive), [Stryker configuration](https://stryker-mutator.io/docs/stryker-js/configuration/).

## Development

```sh
cd dashboard
npm test
npm run typecheck
npm run build
```

Implementation status and integration results are maintained in [PROGRESS.md](PROGRESS.md). Shared interfaces are in [docs/SCHEMA.md](docs/SCHEMA.md); dependencies are recorded in [docs/DEPENDENCIES.md](docs/DEPENDENCIES.md). Dashboard framework was updated from the early Next.js 14 plan to supported Next.js 16.

MIT © Im-A-Nuel.
