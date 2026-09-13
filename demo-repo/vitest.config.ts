import { defineConfig } from "vitest/config";

// Baseline config: only the deliberately shallow "weak" suite. This is what
// `.gauntlet/config.yaml`'s testCommand ("npm test") runs, and what the
// demo's first, low Trust Score measures against.
export default defineConfig({
  test: {
    include: ["test/weak/**/*.test.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json-summary"],
      include: ["src/**/*.ts"],
    },
  },
});
