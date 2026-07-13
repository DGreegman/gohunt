import { fetchJobs } from "@/lib/api";
import { JobCard } from "@/components/JobCard";


export default async function Home() {
  const data = await fetchJobs({ sort: "score", limit: 25 });

  return (
    <main className="min-h-screen bg-[#0F1419] text-slate-100">
      <div className="mx-auto max-w-3xl px-6 py-12">
        <header className="mb-12">
          <h1 className="font-mono text-sm uppercase tracking-widest text-amber-400">
            GoHunt
          </h1>
          <p className="mt-4 text-3xl font-light text-slate-200">
            {data.count} roles, ranked by fit.
          </p>
          <p className="mt-2 text-sm text-slate-500">
            Scored 0–100 against your profile. Click a score to see the breakdown.
          </p>
        </header>

        <div>
          {data.jobs.map((job) => (
            <JobCard key={job.id} job={job} />
          ))}
        </div>
      </div>
    </main>
  );
}