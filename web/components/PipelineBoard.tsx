"use client";

import { useState } from "react";
import { updateApplicationStatus } from "@/lib/api";
import type { Application, ApplicationStatus } from "@/lib/types";

const STAGES: { key: ApplicationStatus; label: string }[] = [
  { key: "new", label: "New" },
  { key: "applied", label: "Applied" },
  { key: "interview", label: "Interview" },
  { key: "offer", label: "Offer" },
  { key: "rejected", label: "Rejected" },
];

function ApplicationCard({
  app,
  onMove,
}: {
  app: Application;
  onMove: (id: number, status: ApplicationStatus) => void;
}) {
  const stageIndex = STAGES.findIndex((s) => s.key === app.status);
  const next = STAGES[stageIndex + 1];
  const prev = STAGES[stageIndex - 1];
  const canReject = app.status !== "rejected" && app.status !== "offer";

  return (
    <div className="rounded border border-slate-800 bg-slate-900/50 p-4">
      
      <a
        href={app.job_url}
        target="_blank"
        rel="noopener noreferrer"
        className="block text-sm font-medium leading-snug text-slate-100 hover:text-amber-400"
      >
        {app.job_title}
      </a>

      <div className="mt-1 text-xs text-slate-500">{app.job_company}</div>

      {app.applied_at && app.status !== "new" && (
        <div className="mt-2 font-mono text-xs text-slate-600">
            Applied {new Date(app.applied_at).toISOString().slice(0, 10)}
        </div>
        )}

      <div className="mt-3 flex items-center gap-3 border-t border-slate-800 pt-3">
        {prev && (
          <button
            onClick={() => onMove(app.id, prev.key)}
            className="-m-1 p-1 text-xs text-slate-600 transition-colors hover:text-slate-300"
            aria-label={`Move back to ${prev.label}`}
          >
            ←
          </button>
        )}

        {next && next.key !== "rejected" && (
          <button
            onClick={() => onMove(app.id, next.key)}
            className="-m-1 p-1 text-xs text-slate-500 transition-colors hover:text-amber-400"
          >
            → {next.label}
          </button>
        )}

        {canReject && (
          <button
            onClick={() => onMove(app.id, "rejected")}
            className="-m-1 ml-auto p-1 text-xs text-slate-600 transition-colors hover:text-red-400"
          >
            Reject
          </button>
        )}
      </div>
    </div>
  );
}

export function PipelineBoard({ initial }: { initial: Application[] }) {
  const [apps, setApps] = useState(initial);

  async function handleMove(id: number, status: ApplicationStatus) {
    const previous = apps;

    // Optimistic update — move it immediately, revert if the API fails.
    setApps((current) =>
      current.map((a) => (a.id === id ? { ...a, status } : a))
    );

    try {
      await updateApplicationStatus(id, status);
    } catch {
      setApps(previous);
    }
  }

  return (
    <div className="grid gap-4 md:grid-cols-5">
      {STAGES.map((stage) => {
        const inStage = apps.filter((a) => a.status === stage.key);
        return (
          <section key={stage.key}>
            <div className="mb-3 flex items-baseline justify-between border-b border-slate-800 pb-2">
              <h2 className="font-mono text-xs uppercase tracking-wide text-slate-400">
                {stage.label}
              </h2>
              <span className="font-mono text-xs tabular-nums text-slate-600">
                {inStage.length}
              </span>
            </div>

            <div className="space-y-3">
              {inStage.map((app) => (
                <ApplicationCard key={app.id} app={app} onMove={handleMove} />
              ))}
              {inStage.length === 0 && (
                <p className="py-4 text-xs text-slate-700">Nothing here yet.</p>
              )}
            </div>
          </section>
        );
      })}
    </div>
  );
}