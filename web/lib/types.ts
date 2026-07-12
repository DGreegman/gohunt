export type Job = {
  id: number;
  title: string;
  company: string;
  source: string;
  url: string;
  location: string;
  remote: boolean;
  posted_at: string | null;
  link_status: string;
  created_at: string;
  fit_score: number | null;
  rationale: string | null;
  dimensions?: Record<string, number>;
};

export type JobsResponse = {
  count: number;
  jobs: Job[];
};


export type ApplicationStatus =
  | "new"
  | "applied"
  | "interview"
  | "offer"
  | "rejected";

export type Application = {
  id: number;
  job_id: number;
  status: ApplicationStatus;
  applied_at: string | null;
  notes: string | null;
  next_action: string | null;
  next_action_date: string | null;
  created_at: string;
  job_title: string;
  job_company: string;
  job_url: string;
};

export type ApplicationsResponse = {
  count: number;
  applications: Application[];
};