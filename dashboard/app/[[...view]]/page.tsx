import { loadRuns } from "@/lib/runs";
import { Dashboard } from "@/components/dashboard";
import { notFound } from "next/navigation";
export const dynamic = "force-dynamic";
export default async function Page({
  params,
  searchParams,
}: {
  params: Promise<{ view?: string[] }>;
  searchParams: Promise<{ run?: string }>;
}) {
  const { view } = await params,
    query = await searchParams;
  const selected = view?.[0] ?? "overview";
  if (
    (view?.length ?? 0) > 1 ||
    !["overview", "matrix", "survivors", "compare"].includes(selected)
  )
    notFound();
  try {
    const data = await loadRuns();
    return <Dashboard {...data} view={selected} selectedId={query.run} />;
  } catch (e) {
    return (
      <Dashboard
        runs={[]}
        source="live"
        view={selected}
        initialError={(e as Error).message}
      />
    );
  }
}
