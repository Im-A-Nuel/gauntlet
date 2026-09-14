import { promises as fs } from "node:fs";
import path from "node:path";
import { parseRun, type Run } from "./schema";

export type RunSource = "sample" | "live";
export class ArtifactError extends Error {}
export async function loadRuns(
  directory?: string,
): Promise<{ runs: Run[]; source: RunSource }> {
  const configured = directory ?? process.env.GAUNTLET_RUNS_DIR;
  let folder = configured
    ? path.resolve(configured)
    : path.resolve(process.cwd(), "../.gauntlet/runs");
  let source: RunSource = "live";
  if (!configured) {
    try {
      await fs.access(folder);
    } catch {
      folder = path.resolve(process.cwd(), "sample-runs");
      source = "sample";
    }
  }
  let entries;
  try {
    entries = await fs.readdir(/* turbopackIgnore: true */ folder, {
      withFileTypes: true,
    });
  } catch (error) {
    if (
      source === "sample" &&
      (error as NodeJS.ErrnoException).code === "ENOENT"
    )
      return { runs: [], source };
    throw new ArtifactError(
      "Cannot read the run directory. Check GAUNTLET_RUNS_DIR and folder permissions.",
    );
  }
  const runs = await Promise.all(
    entries
      .filter((e) => e.isFile() && /^[a-zA-Z0-9_-]+\.json$/.test(e.name))
      .map(async (entry) => {
        try {
          // Sample files are explicitly bundled in next.config; live paths are runtime configuration.
          const full = path.join(
            /* turbopackIgnore: true */ folder,
            entry.name,
          );
          if (
            (await fs.stat(/* turbopackIgnore: true */ full)).size >
            20 * 1024 * 1024
          )
            throw new Error("Oversized artifact");
          const run = parseRun(
            JSON.parse(
              await fs.readFile(/* turbopackIgnore: true */ full, "utf8"),
            ),
          );
          if (entry.name !== `${run.runId}.json`)
            throw new Error("Filename does not match run ID");
          return run;
        } catch {
          throw new ArtifactError(
            `Invalid run artifact: ${entry.name}. Regenerate this report with the CLI.`,
          );
        }
      }),
  );
  return {
    runs: runs
      .sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt))
      .slice(0, 50),
    source,
  };
}
