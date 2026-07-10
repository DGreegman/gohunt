package scorer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)



// Scorer calls the Claude Api to score jobs against profile

type Scorer struct {
	apiKey string
	baseURL string
	client *http.Client
}

// NewScorer constructs a Scorer with the given API key
func NewScorer(apiKey string) *Scorer {
	return &Scorer{
		apiKey: apiKey,
		baseURL: "https://api.anthropic.com/v1/messages",
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// --- Anthropic request/response shapes ---

type anthorpicRequest struct {
	Model string				`json:"model"`
	MaxTokens int				`json:"max_tokens"`
	Messages []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role string 		`json:"role"`
	Content string		`json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string		`json:"type"`
		Text string		`json:"text"`
	}`json:"content"`
}

// callClaude sends a prompt and returns the reaw text response
func (s *Scorer) callClaude(ctx context.Context, prompt string) (string, error) {
	reqBody := anthorpicRequest{
		Model: "claude-sonnet-4-6",
		MaxTokens: 1024,
		Messages: []anthropicMessage{
			{
				Role: "user", Content: prompt,
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("Marshaling Request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("Content-Type", "apllication/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling anthropic: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anthropic returned status %d: %s", resp.StatusCode, string(errBody))
	}

	var out anthropicResponse 
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decoding response %w", err)
	}

	if len(out.Content) == 0 {
		return "", fmt.Errorf("Empty Response from Anthropic")
	}

	return out.Content[0].Text, nil
}

// ScoreResult is the structured score we parse back from claude

type ScoreResult struct {
	RoleMatch 	int		`json:"role_match"`
	SkillOverlap int 	`json:"skill_overlap"`
	SeniorityFit int 	`json:"seniority_fit"`
	StackAlignment int `json:"stack_alignment"`
	LocationCompFit int `json:"location_comp_fit"`
	Rationale string	`json:"rationale"`
}

func (r ScoreResult) Total() int {
	return r.RoleMatch + r.SkillOverlap + r.SeniorityFit + r.StackAlignment + r.LocationCompFit
}

// buildPrompt constructs the rubric scoring prompt for a job + profile
func buildPrompt(jobTitle, jobDescription, jobLocation string, profile ProfileInput) string {
	return fmt.Sprintf(`You are scoring how well a job matches a candidate's profile.

CANDIDATE PROFILE:
- Target roles: %s
- Skills: %s
- Experience: %d years
- Preferred stack: %s
- Location preference: %s (remote only: %t)

JOB:
- Title: %s
- Location: %s
- Description: %s

Score the match on FIVE dimensions, each 0-20:
1. role_match: how well the job title/role aligns with the candidate's target roles
2. skill_overlap: how many of the candidate's skills the job requires
3. seniority_fit: whether the seniority level matches the candidate's experience
4. stack_alignment: how well the tech stack matches the candidate's preferred stack
5. location_comp_fit: whether location/remote arrangement suits the candidate

Respond with ONLY a valid JSON object, no other text, in exactly this format:
{"role_match": 0, "skill_overlap": 0, "seniority_fit": 0, "stack_alignment": 0, "location_comp_fit": 0, "rationale": "one sentence explaining the score"}`,
		strings.Join(profile.TargetRoles, ", "),
		strings.Join(profile.Skills, ", "),
		profile.ExperienceYears,
		strings.Join(profile.PreferredStack, ", "),
		profile.LocationPref,
		profile.RemoteOnly,
		jobTitle,
		jobLocation,
		truncate(jobDescription, 2000),
	)
}

type ProfileInput struct {
	TargetRoles []string 
	Skills      []string
	ExperienceYears int
	PreferredStack []string
	LocationPref	string
	RemoteOnly		bool
}

// truncate, caps a string to n characters to control prompt size/cost
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}


// Score builds the rubric prompt, calls Claude, and parses the result.
func (s *Scorer) Score(ctx context.Context, jobTitle, jobDescription, jobLocation string, profile ProfileInput) (ScoreResult, error) {
	prompt := buildPrompt(jobTitle, jobDescription, jobLocation, profile)

	raw, err := s.callClaude(ctx, prompt)

	if err != nil {
		return ScoreResult{}, fmt.Errorf("calling claude: %w", err)
	}

	jsonText := extractJSON(raw)

	var result ScoreResult 
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return ScoreResult{}, fmt.Errorf("Parsing Score JSON from %q: %w", raw, err)
	}

	return result, nil
}

// extractJSON pulls the JSON object out of a possibly-messy LLM response
// (strips markdown fences, leading/trailing prose).
func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")

	if start == -1 || end == -1 || end < start {
		return s // no braces found - return as-is, let Unmarsharl fail with a clear error
	}

	return s[start : end +1]
}