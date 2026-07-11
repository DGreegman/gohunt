package fetcher

import "context"

// FetchAndStore runs a fetch from the source and saves new jobs.
// Shared by the HTTP trigger and the schedular

func FetchAndStore(ctx context.Context, source JobSource, store *Store) (fetched, inserted int, err error) {
	jobs, err := source.FetchJobs(ctx)

	if err != nil {
		return 0, 0, err
	}

	inserted, err = store.SaveJobs(ctx, jobs)

	if err != nil {
		return len(jobs), 0, err
	}

	return len(jobs), inserted, nil 
}