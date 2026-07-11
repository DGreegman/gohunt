why the pool is bounded (control concurrent load on upstream APIs — 5 workers not 50 simultaneous calls), and why the pool could slot behind the handler untouched (it satisfies JobSource, so the handler depends on the abstraction, not the concrete source).


why launching order isn't execution order (goroutines start and run concurrently; blocking on an empty channel is how a worker waits), and where the API call actually lives (inside each source's FetchJobs, not the pool — pool orchestrates, sources fetch).


contexts form a tree and cancellation flows down it (a child from WithTimeout cancels on timeout OR when its parent cancels, so you get both deadline and caller-cancellation from one line), and why defer cancel() is mandatory (releases the timer even on the happy path; go vet enforces it


why the fake source pattern works (the pool depends on the JobSource interface, so tests inject fakes with canned data and no network — dependency injection is what makes it testable), and what -race actually does (instruments the binary to detect data races at runtime, which is why the test takes ~1s not ~0s, and why running it in CI continuously guards the concurrent code).


why the fake-source pattern makes the pool testable (depends on the interface, inject fakes), and what -race actually does. And genuinely — take the win. You closed out concurrency and tests, and you can defend every line of both.


## One Honest limitations about the filter flag
>> One honest limitation to flag: substring matching is crude. "Go" would match "goal" or "category" (contains "go"). For a cheap pre-filter that's acceptable — a few false positives just means a few extra jobs get scored, which the AI then correctly rates low. But it's worth knowing the limitation so you can defend it: "the keyword filter is deliberately loose — it's a cheap first pass to cut obvious non-matches, and the AI scorer handles precision. Being slightly permissive here is fine; being too strict would drop real jobs." That framing turns a limitation into a reasoned tradeoff.

## Node→Go talking points
>> Go handlers need no async/await — Fiber runs each request in its own goroutine, so blocking DB/fetch calls only park that goroutine while the runtime runs other requests. Async behavior, no async syntax. Separately, the pool's explicit go goroutines add parallelism within a single request. Two levels of concurrency, one automatic (per request) and one you wrote (per fetch). In your own words, though 


## Decoupling fetching of jobs and scoring
>> I decoupled fetching from scoring because they have different cost and latency profiles; fetching is free and fast, scoring is paid and slow, so forcing them together would be poor design.

## The Architecture
>> Standard Go layered architecture with dependency injection and interface-based abstractions at the boundaries. Handlers depend on interfaces, not concrete types, so the worker pool could replace a single source without touching the HTTP layer. DTOs at each layer keep the database schema from leaking into the API contract or the scoring logic

>> Layered — HTTP handlers on top, domain packages for fetching and scoring in the middle, sqlc-generated queries for persistence. Dependencies are injected from main as constructor arguments, and handlers depend on interfaces rather than concrete types — which is how I swapped a single source for a concurrent pool without touching the HTTP layer.

## -race flag for test
>> -race catches data races — two goroutines touching the same memory unsafely — at runtime, which is the only reliable way to find these non-deterministic concurrency bugs. And running it in CI means your concurrent code is continuously guarded.