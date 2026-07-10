package scorer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

)

func TestScore(t *testing.T) {
	// A fake Anthropic server that returns a canned, well-formed response.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert the request looks right (proves our client builds it correctly).
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("missing or wrong x-api-key header")
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Respond in Anthropic's shape, with our scorer's expected JSON in the text.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"content": [
				{"type": "text", "text": "{\"role_match\": 18, \"skill_overlap\": 16, \"seniority_fit\": 14, \"stack_alignment\": 20, \"location_comp_fit\": 12, \"rationale\": \"strong go match\"}"}
			]
		}`))
	}))
	defer server.Close()

	// Point a scorer at the fake server instead of the real API.
	sc := &Scorer{
		apiKey:  "test-key",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}

	profile := ProfileInput{
		TargetRoles:     []string{"Backend Engineer"},
		Skills:          []string{"Go"},
		ExperienceYears: 3,
	}

	result, err := sc.Score(context.Background(), "Backend Engineer", "Go role", "Remote", profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert the five dimensions parsed correctly.
	if result.RoleMatch != 18 {
		t.Errorf("RoleMatch = %d, want 18", result.RoleMatch)
	}
	if result.StackAlignment != 20 {
		t.Errorf("StackAlignment = %d, want 20", result.StackAlignment)
	}
	if result.Rationale != "strong go match" {
		t.Errorf("Rationale = %q, want %q", result.Rationale, "strong go match")
	}

	// Assert Total() sums correctly: 18+16+14+20+12 = 80.
	if result.Total() != 80 {
		t.Errorf("Total() = %d, want 80", result.Total())
	}
}


func TestScoreMalformedJSON(t *testing.T) {
	// server returns a 200 but with text that isn't valid score JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(
			`{
			"content": [
				{"type": "text", "text": "I'm sorry, I can't score this job right now."}
			]
		}`))
	}))
	defer server.Close()

	sc := &Scorer{
		apiKey: "test-key",
		baseURL: server.URL,
		client: &http.Client{Timeout: 5 * time.Second},
	}

	_, err := sc.Score(context.Background(), "Backend Engineer", "Go role", "Remote", ProfileInput{})
	if err == nil {
		t.Fatal("Expected an error for unparsable score text, got nil")
	}
}

func TestScoreServerError(t *testing.T) {
	// Server returns a 500 - Simulating the API being down or rejecting us

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	sc := &Scorer{
		apiKey: "test-key",
		baseURL: server.URL,
		client: &http.Client{Timeout: 5 * time.Second},
	}

	_, err := sc.Score(context.Background(), "Bacdend Engineer", "Go role", "Remote", ProfileInput{})

	if err == nil {
		t.Fatal("Expected an error for status, got nil")
	}
	t.Logf("got expected error: %v", err) 
}