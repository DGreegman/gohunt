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
