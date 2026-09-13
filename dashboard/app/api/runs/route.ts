import { loadRuns } from '@/lib/runs';
export const dynamic = 'force-dynamic';
export async function GET() {
  try {
    const { runs, source } = await loadRuns();
    return Response.json(runs.map(({runId, createdAt, trigger, totals}) => ({runId, createdAt, trigger, totals})),
      { headers: { 'X-Gauntlet-Source': source, 'Cache-Control': 'no-store' } });
  } catch (e) { return Response.json({error: (e as Error).message}, {status: 500}); }
}
