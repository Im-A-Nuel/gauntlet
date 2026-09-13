import { defineConfig } from "vitest/config";

// Strengthened config: weak + strong suites together, so the "after" run
// is additive (new tests kill survivors) rather than a replacement of the
// baseline (docs/SCHEMA.md clarification: strengthening must not delete or
// weaken existing assertions). Select this with
// `gauntlet run --changed --test-command "npm run test:strong"` to
// reproduce the before/after Trust Score delta without a real Bob install.
export default defineConfig({
  test: {
    include: ["test/weak/**/*.test.ts", "test/strong/**/*.test.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json-summary"],
      include: ["src/**/*.ts"],
    },
  },
});
