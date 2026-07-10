package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRemotiveFetchJobs(t *testing.T) {
	// Fake Remotive server returning a canned response - note the T-separator date.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"jobs": [
				{
					"title": "Senior Backend Engineer",
					"company_name": "Acme Corp",
					"url": "https://example.com/job/1",
					"candidate_required_location": "Remote",
					"publication_date": "2026-06-26T14:55:18",
					"description": "Build things with Go"
				},
				{
					"title": "Frontend Developer",
					"company_name": "Beta Inc",
					"url": "https://example.com/job/2",
					"candidate_required_location": "Worldwide",
					"publication_date": "2026-06-20T09:30:00",
					"description": "React work"
				}
			]
		}`))

	}))

	defer server.Close()

	src := &RemotiveSource{
		baseURL: server.URL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}

	jobs, err := src.FetchJobs(context.Background())

	if err != nil {
		t.Fatalf("Unexpected Error: %v", err)
	}

	// Two Jobs in the canned response
	if len(jobs) != 2 {
		t.Fatalf("got %d jobs, want 2", len(jobs))
	}

	// Check the first job parsed correctly.
	j := jobs[0]
	if j.Title != "Senior Backend Engineer" {
		t.Errorf("Title = %q, want %q", j.Title, "Senior Backend Engineer")
	}
	if j.Company != "Acme Corp" {
		t.Errorf("Company = %q, want %q", j.Company, "Acme Corp")
	}
	if j.Source != "Remotive" {
		t.Errorf("Source = %q, want %q", j.Source, "Remotive")
	}

	// The critical one: the T-separator date must parse (not be zero).
	if j.PostedAt.IsZero(){
		t.Error("Post at zero - the zero T-separator date failed to parse")
	}
	if j.PostedAt.Year() != 2026 || j.PostedAt.Month() != 6 || j.PostedAt.Day() != 26 {
		t.Errorf("PostedAt = %v, want 2026-06-26", j.PostedAt)
	}

	// Hash should be non-empty (dedup depends on it).
	if j.Hash == "" {
		t.Error("Hash is empty")
	}
}
