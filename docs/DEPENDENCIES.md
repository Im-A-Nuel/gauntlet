# Dependencies

## Dashboard

- Next.js 16.3.5, React/React DOM 19.3.0: App Router, local artifact API, interactive report views. Updated from the planning-stage Next.js 14, which is outside [official support](https://nextjs.org/support-policy).
- TypeScript 5.x, Node type definitions and React type definitions: strict type checks.
- Zod 4: validate artifacts before showing scores or policy status.
- Recharts 3.10.1: accessible before/after score chart.
- Tailwind CSS / PostCSS plugin 4.3.3: styling pipeline; semantic project tokens and CSS own the visual system.
- IBM Plex Sans / Mono via Fontsource: bundled typography, no remote font calls.
- tsx: execute focused artifact tests with Node's test runner.
- Playwright test and axe-core Playwright integration: browser interaction, responsive overflow checks, screenshots, and accessibility testing. Installed browser channel is Microsoft Edge locally.

Exact resolutions are locked in dashboard/package-lock.json. CLI/demo dependencies and verification results are recorded by Claude in docs/CLAUDE_PROGRESS.md and go.mod/demo-repo/package.json.
