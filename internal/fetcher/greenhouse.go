package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GreenhouseSource fetches jobs from a single company's Greenhouse board.
type GreenhouseSource struct {

	slug string
	baseURL string
	client *http.Client
}

// NewGreenhouseSource constructs a GreenhouseSource with a sane timeout and a source for one company board.
func NewGreenhouseSource(slug string) *GreenhouseSource {
	return &GreenhouseSource{
		slug: slug,
		baseURL: "https://boards-api.greenhouse.io/v1/boards",
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *GreenhouseSource) Name() string {
	return "Greenhouse: " + s.slug
}

type greenhouseResponse struct {
	Jobs []greenhouseJob `json:"jobs"`
}

type greenhouseJob struct {
	Title	   	string `json:"title"`
	AbsoluteURL string `json:"absolute_url"`
	UpdatedAt  	string `json:"updated_at"`
	Content 	string `json:"content"`
	CompanyName	string `json:"company_name"`
	Location   struct {
		Name string `json:"name"`
	} `json:"location"`
}

func (s *GreenhouseSource) FetchJobs(ctx context.Context) ([]Job, error) {
	url := fmt.Sprintf("%s/%s/jobs?content=true", s.baseURL, s.slug)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, fmt.Errorf("Building greenhouse request: %w", err)
	}

	resp, err := s.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Calling Greenhouse %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Greenhouse returned status %d", resp.StatusCode)
	}

	var payload greenhouseResponse

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("Decoding Greenhouse response: %w", err)
	}

	jobs := make([]Job, 0, len(payload.Jobs))
	for _, gj := range payload.Jobs {
		postedAt, _ := time.Parse(time.RFC3339, gj.UpdatedAt)
		company := gj.CompanyName 
		if company == "" {
			company = s.slug
		}
		jobs = append(jobs, Job{
			Title: gj.Title,
			Company: company,
			Source: s.Name(),
			URL: gj.AbsoluteURL,
			Description: gj.Content,
			Location: gj.Location.Name,
			Remote: false,
			PostedAt: postedAt,
			Hash: hashJob(gj.Title, s.slug, gj.Location.Name),
		})
	}

	return jobs, nil

}
