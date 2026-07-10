package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGreenhouseFetchJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(
			`{
			"jobs": [
				{
					"title": "Backend Engineer (Go)",
					"absolute_url": "https://example.com/gh/1",
					"updated_at": "2026-06-26T14:55:18-04:00",
					"location": {"name": "Remote, US"}
				}
			]
		}`))
	}))

	defer server.Close()

	src := &GreenhouseSource{
		slug : "acme",
		baseURL: server.URL,
		client: &http.Client{Timeout: 5 * time.Second},
	}
	jobs, err := src.FetchJobs(context.Background())

	if err != nil {
		t.Fatalf("Unexpecteed error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("got %d jobs, want 1", len(jobs))
	}

	j := jobs[0]

	if j.Title != "Backend Engineer (Go)" {
		t.Errorf("Title = %q, want %q", j.Title, "Backend Engineer (Go)")
	}
	if j.Company != "acme" {
		t.Errorf("Company = %q, want %q", j.Company, "acme")
	}
	if j.Source != "Greenhouse: acme" {
		t.Errorf("Source = %q, want %q", j.Source, "Greenhouse: acme")
	}
	if j.Location != "Remote, US" {
		t.Errorf("Location = %q, want %q", j.Location, "Remote, US")
	}
	// Greenhouse uses RFC3339 dates (with timezone) — verify it parsed.
	if j.PostedAt.IsZero() {
		t.Error("PostedAt is zero — RFC3339 date failed to parse")
	}
	if j.Hash == "" {
		t.Error("Hash is empty")
	}
}