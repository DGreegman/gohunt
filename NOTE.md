why the pool is bounded (control concurrent load on upstream APIs — 5 workers not 50 simultaneous calls), and why the pool could slot behind the handler untouched (it satisfies JobSource, so the handler depends on the abstraction, not the concrete source).


why launching order isn't execution order (goroutines start and run concurrently; blocking on an empty channel is how a worker waits), and where the API call actually lives (inside each source's FetchJobs, not the pool — pool orchestrates, sources fetch).


