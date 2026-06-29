package fetcher

import "context"

// JobSource is anything that can fetch jobs from an external provider.
// Each provider (Remotive, Adzuna, Greenhouse, ...) implements this.

type JobSource interface {
	Name() string
	FetchJobs(ctx context.Context) ([]Job, error)
}