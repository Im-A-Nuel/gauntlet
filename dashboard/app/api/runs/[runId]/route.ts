import { loadRuns } from "@/lib/runs";
export const dynamic = "force-dynamic";
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ runId: string }> },
) {
  const { runId } = await params;
  if (!/^[a-zA-Z0-9_-]+$/.test(runId))
    return Response.json({ error: "run not found" }, { status: 404 });
  try {
    const { runs, source } = await loadRuns();
    const run = runs.find((r) => r.runId === runId);
    return run
      ? Response.json(run, {
          headers: { "X-Gauntlet-Source": source, "Cache-Control": "no-store" },
        })
      : Response.json({ error: "run not found" }, { status: 404 });
  } catch (e) {
    return Response.json({ error: (e as Error).message }, { status: 500 });
  }
}
