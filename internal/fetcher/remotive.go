package fetcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)


// RemotiveSource fetches jobs from the remotive public API

type RemotiveSource struct {
	baseURL string
	client *http.Client
}

// NewRemotiveSource constructs a RemotiveSource with a sane timeout.

func NewRemotiveSource() *RemotiveSource {
	return &RemotiveSource{
		baseURL: "https://remotive.com/api/remote-jobs",
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *RemotiveSource) Name() string {
	return "Remotive"
}

// remotiveResponse mirrors the shape of the Remotive API payload.
type remotiveResponse struct {
	Jobs []remotiveJob `json:"jobs"`
}
type remotiveJob struct {
	Title           string `json:"title"`
	CompanyName     string `json:"company_name"`
	URL             string `json:"url"`
	Location        string `json:"location"`
	PublicationDate string `json:"publication_date"`
	Description     string `json:"description"`
}

func (s *RemotiveSource) FetchJobs(ctx context.Context) ([]Job, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL, nil)

	if err != nil {
		return nil, fmt.Errorf("Building Remotive request: %w", err)
	}

	resp, err := s.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Calling Remotive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Remotive returnd status %d", resp.StatusCode)
	}

	var payload remotiveResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("Decoding remoteive response: %w", err)
	}

	jobs := make([]Job, 0, len(payload.Jobs))

	for _, rj := range payload.Jobs {
		jobs = append(jobs, s.toJob(rj))
	}

	return jobs, nil
}

// toJob nomalises a Remotive job into our domain Job.
func (s *RemotiveSource) toJob(rj remotiveJob) Job {
	postedAt, err := time.Parse("2006-01-02T15:04:05", rj.PublicationDate)
	if err != nil {
		log.Printf("DEBUG date parse failed: raw=%q err=%v", rj.PublicationDate, err)
	}

	return Job{
		Title:       rj.Title,
		Company:     rj.CompanyName,
		Source:      s.Name(),
		URL:         rj.URL,
		Description: rj.CompanyName,
		Location:    rj.Location,
		Remote:      true,
		PostedAt:    postedAt,
		Hash:        hashJob(rj.Title, rj.CompanyName, rj.Location),
		RawJSON:     nil,
	}
}

// hashJob builds a dedub hash from nomalised fields

func hashJob(title, company, location string) string {
	normalise := func(s string) string {
		return strings.ToLower(strings.TrimSpace(s))
	}
	key := normalise(title) + "|" + normalise(company) + "|" + normalise(location)
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
