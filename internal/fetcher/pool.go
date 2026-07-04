package fetcher

import (
	"context"
	"log"
	"sync"
	"time"
)

// Pool runs multiple JobSources concurrently with a bounded number of workers

type Pool struct {
	sources []JobSource
	numWorkers int
}

// NewPool constructs a pool over the given sources with the given number of workers.
func NewPool(sources []JobSource, numWorkers int) *Pool{
	return &Pool{
		sources: sources,
		numWorkers: numWorkers,
	}
}
// Name satisfies Jobsorce so the Pool can be used anywhere a source is can
func (p *Pool) Name() string {
	return "Pool"
}
// FetchJobs urns every source concurrently and returns all jobs combined.
func (p *Pool) FetchJobs(ctx context.Context) ([]Job, error) {
	// Give the whole fetch a hard deadline. if it isn't done in time,
	// the context cancels and every source's HTTP call aborts
	ctx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()

	work := make(chan JobSource) //queue: sources going IN to workers
	results := make(chan Job) //belt: jobs coming OUT of workers

	var wg sync.WaitGroup 

	// Start a fixed number of workers
	for i :=0; i< p.numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for source := range work { // pull a source, fetch, repeat
				jobs, err := source.FetchJobs(ctx)
				if err != nil {
					log.Printf("Source %s failed: %v", source.Name(), err)
					continue // graceful degradation: skip, don't abort
				}

				for _, j := range jobs {
					results <- j // send each job to the results belts
				}
			}
		}()
	}


	// Feed the sources into the work queue, then close it
	go func() {
		for _, s := range p.sources {
			work <- s
		}
		close(work)
	}()

	// Close results once all workers are done ()
	go func(){
		wg.Wait()
		close(results)
	}()

	// Main goroutine colleccts everything off the results belts.
	var all []Job
	for j := range results {
		all = append(all, j)
	}

	return all, nil 
}