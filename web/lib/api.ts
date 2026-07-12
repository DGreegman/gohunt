import type { JobsResponse } from "./types";
import type { ApplicationsResponse, ApplicationStatus } from "./types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function fetchJobs(params?: {
  sort?: "score" | "date";
  limit?: number;
  offset?: number;
}): Promise<JobsResponse> {
  const query = new URLSearchParams();
  if (params?.sort) query.set("sort", params.sort);
  if (params?.limit) query.set("limit", String(params.limit));
  if (params?.offset) query.set("offset", String(params.offset));

  const res = await fetch(`${API_URL}/api/jobs?${query.toString()}`, {
    cache: "no-store",
  });

  if (!res.ok) {
    throw new Error(`Failed to fetch jobs: ${res.status}`);
  }

  return res.json();
}


export async function trackJob(jobId: number, notes?: string): Promise<void> {
  const res = await fetch(`${API_URL}/api/applications`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ job_id: jobId, notes: notes ?? "" }),
  });

  if (res.status === 409) {
    throw new Error("ALREADY_TRACKED");
  }
  if (!res.ok) {
    throw new Error(`Failed to track job: ${res.status}`);
  }
}



export async function fetchApplications(): Promise<ApplicationsResponse> {
  const res = await fetch(`${API_URL}/api/applications`, {
    cache: "no-store",
  });
  if (!res.ok) throw new Error(`Failed to fetch applications: ${res.status}`);
  return res.json();
}

export async function updateApplicationStatus(
  id: number,
  status: ApplicationStatus
): Promise<void> {
  const res = await fetch(`${API_URL}/api/applications/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ status }),
  });
  if (!res.ok) throw new Error(`Failed to update: ${res.status}`);
}