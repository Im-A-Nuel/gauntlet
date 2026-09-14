import { z } from "zod";

const percent = z.number().finite().min(0).max(100).nullable();
const count = z.number().int().nonnegative();
export const statuses = [
  "killed",
  "survived",
  "timeout",
  "noCoverage",
  "ignored",
  "compileError",
  "runtimeError",
] as const;
export const mutantSchema = z.object({
  id: z.string(),
  mutator: z.string(),
  line: z.number().int().positive(),
  status: z.enum(statuses),
  original: z.string(),
  mutated: z.string(),
});
export const runSchema = z.object({
  schemaVersion: z.literal(1).default(1),
  runId: z.string().regex(/^[a-zA-Z0-9_-]+$/),
  createdAt: z.string().datetime({ offset: true }),
  baseRef: z.string(),
  headSha: z.string(),
  trigger: z.enum(["hook", "manual", "strengthen", "ci"]),
  changedFiles: z.array(z.string()).default([]),
  threshold: z.number().min(0).max(100).default(80),
  totals: z.object({
    mutants: count,
    killed: count,
    survived: count,
    timeout: count,
    noCoverage: count,
    trustScore: percent,
    lineCoverage: percent,
    durationMs: z.number().nonnegative(),
    ignored: count.default(0),
    compileError: count.default(0),
    runtimeError: count.default(0),
  }),
  files: z.array(
    z.object({
      path: z.string(),
      trustScore: percent,
      mutants: z.array(mutantSchema),
    }),
  ),
  comparedTo: z.string().optional(),
});
export type Run = z.infer<typeof runSchema>;
export type Mutant = z.infer<typeof mutantSchema>;
export type Status = Mutant["status"];
export type LocatedMutant = Mutant & { path: string };
export const labels: Record<Status, string> = {
  killed: "Killed",
  survived: "Survived",
  timeout: "Timed out",
  noCoverage: "No coverage",
  ignored: "Ignored",
  compileError: "Compile error",
  runtimeError: "Runtime error",
};
export function score(
  killed: number,
  timeout: number,
  survived: number,
): number | null {
  const total = killed + timeout + survived;
  return total ? Math.round(((killed + timeout) / total) * 1000) / 10 : null;
}
export function parseRun(input: unknown): Run {
  const run = runSchema.parse(input);
  const counts = Object.fromEntries(statuses.map((s) => [s, 0])) as Record<
    Status,
    number
  >;
  const paths = new Set<string>();
  for (const file of run.files) {
    if (paths.has(file.path)) throw new Error("Duplicate file path");
    paths.add(file.path);
    const local = { killed: 0, timeout: 0, survived: 0 };
    const ids = new Set<string>();
    for (const mutant of file.mutants) {
      if (ids.has(mutant.id)) throw new Error("Duplicate mutant ID in file");
      ids.add(mutant.id);
      counts[mutant.status]++;
      if (
        mutant.status === "killed" ||
        mutant.status === "timeout" ||
        mutant.status === "survived"
      )
        local[mutant.status]++;
    }
    if (file.trustScore !== score(local.killed, local.timeout, local.survived))
      throw new Error("Inconsistent file score");
  }
  for (const status of statuses)
    if (run.totals[status] !== counts[status])
      throw new Error("Inconsistent status totals");
  if (run.totals.mutants !== Object.values(counts).reduce((a, b) => a + b, 0))
    throw new Error("Inconsistent mutant total");
  if (
    run.totals.trustScore !==
    score(counts.killed, counts.timeout, counts.survived)
  )
    throw new Error("Inconsistent trust score");
  return run;
}
export function verdict(run: Run): "pass" | "fail" | "unscored" {
  if (run.totals.trustScore === null) return "unscored";
  return run.totals.trustScore >= run.threshold && run.totals.runtimeError === 0
    ? "pass"
    : "fail";
}
export const formatScore = (value: number | null) =>
  value === null ? "N/A" : value.toFixed(1);
export const mutantsOf = (run: Run): LocatedMutant[] =>
  run.files.flatMap((file) =>
    file.mutants.map((m) => ({ ...m, path: file.path })),
  );
export function compare(a: Run, b: Run) {
  return {
    a,
    b,
    comparable:
      a.headSha === b.headSha &&
      [...a.changedFiles].sort().join("\n") ===
        [...b.changedFiles].sort().join("\n"),
    delta: {
      trustScore:
        a.totals.trustScore === null || b.totals.trustScore === null
          ? null
          : Math.round((b.totals.trustScore - a.totals.trustScore) * 10) / 10,
      killed: b.totals.killed - a.totals.killed,
      survived: b.totals.survived - a.totals.survived,
    },
    perFile: [...new Set([...a.files, ...b.files].map((f) => f.path))].map(
      (path) => ({
        path,
        a: a.files.find((f) => f.path === path)?.trustScore ?? null,
        b: b.files.find((f) => f.path === path)?.trustScore ?? null,
      }),
    ),
  };
}
