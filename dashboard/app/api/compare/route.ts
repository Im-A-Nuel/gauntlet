import { loadRuns } from "@/lib/runs";
import { compare } from "@/lib/schema";
export const dynamic = "force-dynamic";
export async function GET(request: Request) {
  const url = new URL(request.url);
  if (!url.searchParams.get("a") || !url.searchParams.get("b"))
    return Response.json(
      { error: "Select both run IDs with a and b." },
      { status: 400 },
    );
  try {
    const { runs } = await loadRuns();
    const a = runs.find((r) => r.runId === url.searchParams.get("a")),
      b = runs.find((r) => r.runId === url.searchParams.get("b"));
    return a && b
      ? Response.json(compare(a, b))
      : Response.json({ error: "run not found" }, { status: 404 });
  } catch (e) {
    return Response.json({ error: (e as Error).message }, { status: 500 });
  }
}
