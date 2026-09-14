# Dependencies

## Dashboard

- Next.js 16.3.5, React/React DOM 19.3.0: App Router, local artifact API, interactive report views. Updated from the planning-stage Next.js 14, which is outside [official support](https://nextjs.org/support-policy).
- TypeScript 5.x, Node type definitions and React type definitions: strict type checks.
- Zod 4: validate artifacts before showing scores or policy status.
- Recharts 3.10.1: accessible before/after score chart.
- Tailwind CSS / PostCSS plugin 4.3.3: styling pipeline; semantic project tokens and CSS own the visual system.
- IBM Plex Sans / Mono and Manrope Variable via Fontsource: locally bundled typography with Manrope reserved for the wordmark, and no remote font calls.
- tsx: execute focused artifact tests with Node's test runner.
- Playwright test and axe-core Playwright integration: browser interaction, responsive overflow checks, screenshots, and accessibility testing. Installed browser channel is Microsoft Edge locally.
- Prettier: consistent formatting for dashboard source and tests.

Exact resolutions are locked in `dashboard/package-lock.json`. `npm audit` reports zero dashboard vulnerabilities.

## CLI

- Cobra: command and flag handling.
- yaml.v3: `.gauntlet/config.yaml` parsing and writing.
- doublestar/v4: `**` include/exclude matching for mutation scope.

Exact Go module versions are locked in `go.sum`.

## Demo repository

- StrykerJS 10: mutation engine invoked through its installed JavaScript entrypoint.
- Vitest 5 and `@vitest/coverage-v8` 5: weak and additive strong test suites with measured coverage.
- TypeScript 5: demo source and tests.

All demo dependencies are development-only and locked in `demo-repo/package-lock.json`. `npm audit` reports two moderate `qs` advisories inherited through Stryker's `typed-rest-client` chain; there is no application runtime dependency or dashboard exposure. Full CLI/demo history remains in `docs/CLAUDE_PROGRESS.md`.
