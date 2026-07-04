package fetcher

import (
	"context"
	"fmt"
	"testing"
)

// fakeSource is a test double implementing JobSource with canned data, no network, fully controllable

type fakeSource struct {
	name string
	jobs []Job
	err  error
}

func (f *fakeSource) Name() string {
	return f.name
}

func (f *fakeSource) FetchJobs(ctx context.Context) ([]Job, error){
	if f.err != nil {
		return nil, f.err 
	}

	return f.jobs, nil
}

func TestPoolFetchJobs(t *testing.T){
	sources := []JobSource{
		&fakeSource {
			name: "fake-a", 
			jobs: []Job{
				{Title: "Job 1", Company: "A", Hash: "h1"},
				{Title: "Job 2", Company: "A", Hash: "h2"},
			},

		},
		&fakeSource {
			name: "fake-b", 
			jobs: []Job{
				{Title: "Job 3", Company: "B", Hash: "h3"},
			},
		},
	}

	pool := NewPool(sources, 2)

	got, err := pool.FetchJobs(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	want := 3 
	if len(got) != want {
		t.Errorf("got %d jobs, want %d", len(got), want)
	}
}

func TestPoolGracefulDegradation(t *testing.T){
	sources := []JobSource{
		&fakeSource{
			name: "healthy",
			jobs: []Job{
				{Title: "Job 1", Company: "A", Hash: "h1"},
				{Title: "Job 2", Company: "A", Hash: "h2"},
			},
		},
		&fakeSource{
			name: "broken",
			err:  fmt.Errorf("simulated network failure"),
		},
	}

	pool := NewPool(sources, 2)

	got, err := pool.FetchJobs(context.Background())
	if err != nil {
		t.Fatalf("pool should not error when one source fails: %v", err)
	}

	// The boken source contributes nothing; the healthy one giive 2
	want := 2

	if len(got) != want {
		t.Errorf("Got %d jobs, want %d", len(got), want)
	}
}

func TestPoolNoSources (t *testing.T) {
	pool := NewPool([]JobSource{}, 2)

	got, err := pool.FetchJobs(context.Background())

	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if len(got) != 0 {
		t.Errorf("Got %d jobs, want 0", len(got))
	}
}