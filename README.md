# GoHunt

**An AI-powered job aggregator that fetches real listings, scores them against your profile, and ranks your feed by fit.**

GoHunt pulls job postings from multiple sources, filters out the noise, and uses the Claude API to score how well each role matches a candidate's skills, target roles, and experience — then surfaces the best matches first, each with a one-line rationale explaining the score.

Built in Go with a focus on production practices: concurrent fetching, type-safe database access, structured AI scoring, and a tested concurrency layer.

> **Status:** MVP in active development. Core loop (fetch → dedupe → filter → score → ranked feed) is working. See [Roadmap](#roadmap).

---

## Why this exists

I'm a Node.js/Python developer learning production-grade Go by building something real rather than following tutorials. GoHunt is both a genuine tool (I use it to find and rank roles worth applying to) and a portfolio project demonstrating idiomatic Go engineering.

Every architectural decision is one I can explain — the goal was defensible fluency, not just working code.

---

## How it works

```
                    ┌─────────────┐
   Job sources ────▶│ Worker pool │───▶ dedupe ───▶ Postgres
   (Remotive,       │ (concurrent,│     (hash)
    Greenhouse)     │  bounded)   │
                    └─────────────┘
                                              │
   GET /api/jobs?sort=score ◀── ranked feed ◀─┤
                                              │
   POST /api/score/trigger ──▶ filter ──▶ Claude scorer ──▶ job_scores
                            (cheap, in Go)  (rubric, 5 dims)
```

1. **Fetch** — A bounded worker pool concurrently pulls listings from multiple sources. Each source implements a common `JobSource` interface, so adding a source (or swapping a single source for the whole pool) requires no changes to the calling code.
2. **Dedupe** — Jobs are hashed on normalised title/company/location. A `UNIQUE` constraint plus `ON CONFLICT DO NOTHING` enforces deduplication at the database level, across all sources.
3. **Filter** — A cheap, in-Go relevance pass (keyword + staleness) decides which jobs are worth scoring, so the paid AI step only runs on plausible matches.
4. **Score** — Survivors are scored by the Claude API against the user's profile using a five-dimension rubric (role match, skill overlap, seniority fit, stack alignment, location/comp), each 0–20, summed to a 0–100 fit score with a rationale.
5. **Rank** — `GET /api/jobs?sort=score` returns the feed ordered by fit, best matches first.

---

## Tech stack

| Layer | Technology |
|---|---|
| Language | Go 1.22+ |
| Web framework | [Fiber](https://gofiber.io/) v2 |
| Database | PostgreSQL (via Docker Compose) |
| DB access | [sqlc](https://sqlc.dev/) (type-safe codegen) + [pgx](https://github.com/jackc/pgx) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| AI | Anthropic Claude API (hand-rolled HTTP client) |
| API docs | [swaggo](https://github.com/swaggo/swag) (OpenAPI / Swagger UI) |
| CI | GitHub Actions (`go vet`, `golangci-lint`, `go test -race`) |

---

## Engineering highlights

Things I deliberately built to demonstrate production instincts:

- **Concurrent, bounded worker pool** — Sources are fetched in parallel through a fixed pool of workers (not one goroutine per source), with a work queue and results channel, `WaitGroup` coordination, per-source rate limiting via HTTP client timeouts, and **graceful degradation** (one source failing never aborts the batch).
- **Context-based cancellation** — A timeout context propagates from the pool down through each source's HTTP request, so the whole fetch has a hard deadline and cancels cleanly.
- **Race-tested concurrency** — The worker pool has a table-driven test suite (happy path, graceful degradation, empty input) run under `go test -race` in CI.
- **Interface-driven design** — Handlers depend on the `JobSource` interface, not concrete types. The worker pool itself satisfies `JobSource`, so it drops in wherever a single source would (composite pattern).
- **Type-safe SQL** — All queries are hand-written SQL, with sqlc generating type-safe Go. Mistakes surface at build time, not in production. Parameterised queries throughout (injection-safe by construction).
- **DTOs at every boundary** — Separate types for fetched jobs, stored rows, API responses, and scorer inputs, so the database schema never leaks into the API contract or business logic.
- **Structured LLM output** — The Claude scorer returns parseable JSON via a rubric prompt, with defensive extraction to handle real-world LLM output variance.

---

## Getting started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- An Anthropic API key ([console.anthropic.com](https://console.anthropic.com))

### Setup

```bash
# 1. Clone
git clone https://github.com/DGreegman/gohunt.git
cd gohunt

# 2. Start Postgres
docker compose up -d

# 3. Configure environment
cp .env.example .env
# edit .env — set DATABASE_URL and ANTHROPIC_API_KEY

# 4. Run migrations
migrate -path migrations -database "$DATABASE_URL" up

# 5. Run the server
go run ./cmd/server
```

The API runs on `http://localhost:8080`. Swagger docs at `http://localhost:8080/swagger/index.html`.

### Quick start

```bash
# Set your profile (what jobs get scored against)
curl -X PUT localhost:8080/api/profile \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane","skills":["Go","PostgreSQL","Docker"],"target_roles":["Backend Engineer"],"experience_years":3,"remote_only":true,"timezone":"UTC"}'

# Fetch jobs from all sources
curl -X POST localhost:8080/api/fetch/trigger

# Score the fetched jobs against your profile
curl -X POST localhost:8080/api/score/trigger

# Get your feed, ranked by fit
curl "localhost:8080/api/jobs?sort=score&limit=10"
```

---

## API

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Liveness check |
| `GET` | `/api/jobs` | List jobs (`?sort=score` to rank by fit, `?limit=` / `?offset=` to paginate) |
| `POST` | `/api/fetch/trigger` | Fetch jobs from all sources concurrently |
| `POST` | `/api/score/trigger` | Filter and score unscored jobs against the profile |
| `GET` | `/api/profile` | Get the user profile |
| `PUT` | `/api/profile` | Create or update the profile |
| `GET` | `/swagger/*` | Interactive API documentation |

---

## Roadmap

- [x] Concurrent multi-source fetcher with dedupe
- [x] AI fit scoring (Claude rubric) + ranked feed
- [x] User profile
- [ ] Scheduled fetching (cron) to replace manual triggers
- [ ] Application tracker (Kanban pipeline)
- [ ] AI-drafted cover letters
- [ ] Daily digest of top matches

---

## Notes

This is a single-user MVP by design — multi-user support is deliberately out of scope until the core loop is complete. Job-board dates from ATS sources (e.g. Greenhouse) reflect last-updated rather than true posting time, so the staleness filter treats them as approximate.

Built in public. Follow along.
