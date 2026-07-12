"use client";

import { useState } from "react";
import type { Job } from "@/lib/types";

const DIMENSION_LABELS: Record<string, string> = {
  role_match: "Role match",
  skill_overlap: "Skill overlap",
  seniority_fit: "Seniority fit",
  stack_alignment: "Stack alignment",
  location_comp_fit: "Location & comp",
};

const DIMENSION_ORDER = [
  "role_match",
  "skill_overlap",
  "seniority_fit",
  "stack_alignment",
  "location_comp_fit",
];

function scoreColor(score: number | null) {
  if (score === null) return "text-slate-700";
  if (score >= 70) return "text-amber-400";
  if (score >= 50) return "text-amber-400/60";
  return "text-slate-500";
}

function DimensionBar({ label, value }: { label: string; value: number }) {
  const pct = (value / 20) * 100;
  return (
    <div className="flex items-center gap-3">
      <div className="w-32 shrink-0 text-xs text-slate-500">{label}</div>
      <div className="h-1 flex-1 overflow-hidden rounded-full bg-slate-800">
        <div
          className="h-full rounded-full bg-amber-400/70"
          style={{ width: `${pct}%` }}
        />
      </div>
      <div className="w-8 shrink-0 text-right font-mono text-xs tabular-nums text-slate-400">
        {value}
      </div>
    </div>
  );
}

export function JobCard({ job }: { job: Job }) {
  const [expanded, setExpanded] = useState(false);
  const hasDimensions = job.dimensions && Object.keys(job.dimensions).length > 0;

  return (
    <article className="-mx-4 border-b border-slate-800 px-4 py-8 transition-colors hover:bg-slate-900/40">
      <div className="flex gap-6">
        <div className="w-16 shrink-0 pt-1">
          {job.fit_score === null ? (
            <div className="font-mono text-2xl tabular-nums text-slate-700">—</div>
          ) : (
            <button
              onClick={() => hasDimensions && setExpanded(!expanded)}
              disabled={!hasDimensions}
              className={`font-mono text-3xl font-medium tabular-nums transition-opacity ${scoreColor(
                job.fit_score
              )} ${hasDimensions ? "cursor-pointer hover:opacity-70" : "cursor-default"}`}
              aria-expanded={expanded}
              aria-label={`Fit score ${job.fit_score}. ${
                hasDimensions ? "Click to see breakdown." : ""
              }`}
            >
              {job.fit_score}
            </button>
          )}
        </div>

        <div className="min-w-0 flex-1">
          
          <a
            href={job.url}
            target="_blank"
            rel="noopener noreferrer"
            className="block text-lg font-medium text-slate-100 hover:text-amber-400">
            {job.title}
          </a>

          <div className="mt-1 text-sm text-slate-400">
            {job.company}
            {job.location && (
              <span className="text-slate-600"> · {job.location}</span>
            )}
          </div>

          {job.rationale && (
            <p className="mt-3 text-sm leading-relaxed text-slate-400">
              {job.rationale}
            </p>
          )}

          {expanded && hasDimensions && (
            <div className="mt-5 space-y-2 rounded border border-slate-800 bg-slate-900/50 p-4">
              {DIMENSION_ORDER.map((key) =>
                job.dimensions?.[key] !== undefined ? (
                  <DimensionBar
                    key={key}
                    label={DIMENSION_LABELS[key] ?? key}
                    value={job.dimensions[key]}
                  />
                ) : null
              )}
            </div>
          )}

          <div className="mt-3 font-mono text-xs uppercase tracking-wide text-slate-600">
            {job.source}
          </div>
        </div>
      </div>
    </article>
  );
}