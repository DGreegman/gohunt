import { fetchApplications } from "@/lib/api";
import { PipelineBoard } from "@/components/PipelineBoard";

export default async function Pipeline() {
  const data = await fetchApplications();

  return (
    <main className="min-h-screen bg-[#0F1419] text-slate-100">
      <div className="mx-auto max-w-6xl px-6 py-12">
        <header className="mb-12">
          <h1 className="font-mono text-sm uppercase tracking-widest text-amber-400">
            Pipeline
          </h1>
          <p className="mt-4 text-3xl font-light text-slate-200">
            {data.count} {data.count === 1 ? "application" : "applications"} in flight.
          </p>
        </header>

        <PipelineBoard initial={data.applications} />
      </div>
    </main>
  );
}