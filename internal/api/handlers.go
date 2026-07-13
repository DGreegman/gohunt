package api

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/DGreegman/gohunt/internal/db"
	"github.com/DGreegman/gohunt/internal/fetcher"
	"github.com/DGreegman/gohunt/internal/scorer"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// JobResponse is the API shape for a job (decoupled from the DB row).
type JobResponse struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	Company    string  `json:"company"`
	Source     string  `json:"source"`
	URL        string  `json:"url"`
	Location   string  `json:"location"`
	Remote     bool    `json:"remote"`
	PostedAt   *string `json:"posted_at"`
	LinkStatus string  `json:"link_status"`
	CreatedAt  string  `json:"created_at"`
	FitScore  *int32 	`json:"fit_score"`
	Rationale *string	`json:"rationale"`
	Dimensions map[string]int `json:"dimensions,omitempty"`
}

// Handler holds dependencies shared across HTTP handlers.
type Handler struct {
	queries *db.Queries
	source  fetcher.JobSource
	store   *fetcher.Store
	scorer  *scorer.Scorer
}

// NewHandler wires handlers to the connection pool.
func NewHandler(pool *pgxpool.Pool, source fetcher.JobSource, store *fetcher.Store, sc *scorer.Scorer) *Handler {
	return &Handler{
		queries: db.New(pool),
		source:  source,
		store:   store,
		scorer:  sc,
	}
}

// ListJobs godoc
// @Summary      List jobs
// @Description  Returns a paginated list of jobs. Use sort=score to rank by AI fit score.
// @Tags         jobs
// @Produce      json
// @Param        sort    query     string  false  "Sort order: 'score' for fit ranking, otherwise by date"
// @Param        limit   query     int     false  "Max results (default 20)"
// @Param        offset  query     int     false  "Rows to skip (default 0)"
// @Success      200     {object}  map[string]interface{}
// @Failure      500     {object}  map[string]interface{}
// @Router       /api/jobs [get]
func (h *Handler) ListJobs(c *fiber.Ctx) error {
	limit := parseIntDefault(c.Query("limit"), 20)
	offset := parseIntDefault(c.Query("offset"), 0)
	sortByScore := c.Query("sort") == "score"

	jobs, err := h.queries.ListJobs(c.Context(), db.ListJobsParams{
		SortByScore: sortByScore,
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to List Jobs",
		})
	}
	resp := make([]JobResponse, 0, len(jobs))
	for _, j := range jobs {
		resp = append(resp, toJobResponse(j))
	}

	return c.JSON(fiber.Map{
		"count": len(resp),
		"jobs":  resp,
	})
}

// TriggerFetch godoc
// @Summary      Trigger a job fetch
// @Description  Fetches jobs from the configured source and stores new ones
// @Tags         fetch
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/fetch/trigger [post]
func (h *Handler) TriggerFetch(c *fiber.Ctx) error {
	
	fetched, inserted, err := fetcher.FetchAndStore(c.Context(), h.source, h.store)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "fetch failed...",
		})
	}

	return c.JSON(fiber.Map{
		"source":   h.source.Name(),
		"fetched":  fetched,
		"inserted": inserted,
	})
}

func toJobResponse(j db.ListJobsRow) JobResponse {
	r := JobResponse{
		ID:         j.ID,
		Title:      j.Title,
		Company:    j.Company,
		Source:     j.Source,
		URL:        j.Url,
		Location:   j.Location,
		Remote:     j.Remote,
		LinkStatus: j.LinkStatus,
		CreatedAt:  j.CreatedAt.Time.Format(time.RFC3339),
	}
	if j.PostedAt.Valid {
		s := j.PostedAt.Time.Format(time.RFC3339)
		r.PostedAt = &s
	}
	if j.FitScore.Valid {
		v := j.FitScore.Int32
		r.FitScore = &v
	}

	if j.Rationale.Valid {
		r.Rationale = &j.Rationale.String
	}
	if len(j.DimensionScores) > 0 {
		var dims map[string]int
		if err := json.Unmarshal(j.DimensionScores, &dims); err == nil {
			r.Dimensions = dims
		}
	}
	return r
}

func parseIntDefault(s string, fallback int) int {
	if s == "" {
		return fallback
	}

	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return fallback
	}

	return n
}

// ProfileRequest is the shape a client sends to create/update the profile
type ProfileRequest struct {
	Name             string   `json:"name"`
	CurrentRoleTitle string   `json:"current_role_title"`
	Skills           []string `json:"skills"`
	ExperienceYears  int32    `json:"experience_years"`
	TargetRoles      []string `json:"target_roles"`
	PreferredStack   []string `json:"preferred_stack"`
	SalaryMin        int32    `json:"salary_min"`
	SalaryMax        int32    `json:"salary_max"`
	LocationPref     string   `json:"location_pref"`
	RemoteOnly       bool     `json:"remote_only"`
	Timezone         string   `json:"timezone"`
}

// GetProfile godoc
// @Summary      Get user profile
// @Tags         profile
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /api/profile [get]
func (h *Handler) GetProfile(c *fiber.Ctx) error {
	profile, err := h.queries.GetProfile(c.Context())
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "no profile found - create one with PUT /api/profile",
		})
	}
	return c.JSON(profile)
}

// UpsertProfile godoc
// @Summary      Create or update user profile
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        profile  body      ProfileRequest  true  "Profile data"
// @Success      200      {object}  map[string]interface{}
// @Failure      400      {object}  map[string]interface{}
// @Router       /api/profile [put]
func (h *Handler) UpsertProfile(c *fiber.Ctx) error {
	var req ProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	profile, err := h.queries.UpsertProfile(c.Context(), db.UpsertProfileParams{
		Name:             req.Name,
		CurrentRoleTitle: pgtype.Text{String: req.CurrentRoleTitle, Valid: req.CurrentRoleTitle != ""},
		Skills:           req.Skills,
		ExperienceYears:  pgtype.Int4{Int32: req.ExperienceYears, Valid: true},
		TargetRoles:      req.TargetRoles,
		PreferredStack:   req.PreferredStack,
		SalaryMin:        pgtype.Int4{Int32: req.SalaryMin, Valid: true},
		SalaryMax:        pgtype.Int4{Int32: req.SalaryMax, Valid: true},
		LocationPref:     pgtype.Text{String: req.LocationPref, Valid: req.LocationPref != ""},
		RemoteOnly:       req.RemoteOnly,
		Timezone:         req.Timezone,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save Profile",
		})
	}

	return c.JSON(profile)
}

// TriggerScoring godoc
// @Summary      Score unscored jobs
// @Description  Loads the profile, filters unscored jobs, scores them via Claude, stores results
// @Tags         scoring
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/score/trigger [post]
func (h *Handler) TriggerScoring(c *fiber.Ctx) error {
	ctx := c.Context()

	// 1. Load the profile to score against
	profile, err := h.queries.GetProfile(ctx)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No Profile Found - Create one with and try again later",
		})
	}

	// 2. Get unscored jobs (cap the batch to control cost).
	jobs, err := h.queries.ListUnscoredJobs(ctx, 500)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list UnScored Jobs",
		})
	}

	// 3. Build filter criteria from the profile.
	criteria := fetcher.FilterCriteria{
		Keywords: buildKeywords(profile),
		MaxAge:   90 * 24 * time.Hour,
	}

	// 4. Map the db profile into the scorer's ProfileInput.
	profileInput := scorer.ProfileInput{
		TargetRoles:     profile.TargetRoles,
		Skills:          profile.Skills,
		ExperienceYears: int(profile.ExperienceYears.Int32),
		PreferredStack:  profile.PreferredStack,
		LocationPref:    profile.LocationPref.String,
		RemoteOnly:      profile.RemoteOnly,
	}

	const maxToScore = 100
	scored := 0
	skipped := 0

	// 5. Score each job that passes the cheap filter.

	for i, j := range jobs {
		log.Printf("LOOP iteration %d: job %d %q", i, j.ID, j.Title)

		if scored >= maxToScore {
			break
		}
		job := fetcher.Job{
			Title:       j.Title,
			Description: j.Description,
			Location:    j.Location,
			PostedAt:    j.PostedAt.Time,
		}

		if !criteria.Relevant(job) {
			skipped++
			continue
		}

		

		result, err := h.scorer.Score(ctx, j.Title, j.Description, j.Location, profileInput)

		if err != nil {
			log.Printf("SCORE FAILED job %d: %v", j.ID, err)
			continue
		}


		dimensions, _ := json.Marshal(map[string]int{
			"role_match" : result.RoleMatch,
			"skill_overlap": result.SkillOverlap,
			"seniority_fit": result.SeniorityFit,
			"stack_alignment": result.StackAlignment,
			"location_comp_fit": result.LocationCompFit,
		})


		_, err = h.queries.CreateJobScore(ctx, db.CreateJobScoreParams{
			JobID: j.ID,
			FitScore: int32(result.Total()),
			DimensionScores: dimensions,
			Rationale: pgtype.Text{String: result.Rationale, Valid: result.Rationale != ""},
		})

		if err != nil {
			log.Printf("Saving Score for job %d failed: %v", j.ID, err)
			continue
			
		}
		log.Printf("SAVE OK job %d", j.ID)
		scored++
		
	}

	return c.JSON(fiber.Map{
		"unscored_found" : len(jobs),
		"scored" : scored,
		"skipped_filter" : skipped,
	})

}

//  buildKeywords flattens target roles + skills into lowercase filter keywords.
func buildKeywords(profile db.UserProfile) []string {
	var kws []string

	for _, r := range profile.TargetRoles {
		kws = append(kws, strings.ToLower(r))
	}

	for _, s := range profile.Skills {
		kws = append(kws, strings.ToLower(s))
	}

	return kws
}

// ApplicationResponse is the API shape for an application (with job details).
type ApplicationResponse struct {
	ID             int64   `json:"id"`
	JobID          int64   `json:"job_id"`
	Status         string  `json:"status"`
	AppliedAt      *string `json:"applied_at"`
	Notes          *string `json:"notes"`
	NextAction     *string `json:"next_action"`
	NextActionDate *string `json:"next_action_date"`
	CreatedAt      string  `json:"created_at"`

	// Joined job details
	JobTitle   string `json:"job_title,omitempty"`
	JobCompany string `json:"job_company,omitempty"`
	JobURL     string `json:"job_url,omitempty"`
}

type CreateApplicationRequest struct {
	JobID int64		`json:"job_id"`
	Notes string	`json:"notes"`
}

// CreateApplication godoc
// @Summary      Create an application from a job
// @Tags         applications
// @Accept       json
// @Produce      json
// @Param        application  body      CreateApplicationRequest  true  "Application data"
// @Success      201          {object}  map[string]interface{}
// @Failure      400          {object}  map[string]interface{}
// @Router       /api/applications [post]
func (h *Handler) CreateApplication(c *fiber.Ctx) error {
	var req CreateApplicationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error" : "invalid request body",
		})
	}
	if req.JobID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error" : "job_id is required...",
		})
	}

	app, err := h.queries.CreateApplication(c.Context(), db.CreateApplicationParams{
		JobID: req.JobID,
		Notes: pgtype.Text{String:req.Notes, Valid: req.Notes != ""},
	})

	if err != nil {
		log.Printf("DEBUG create application error: %v (type %T)", err, err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error" : "This is already being tracked...",
			})

		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error" : "failed to create application",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(app)
}


// ListApplications godoc
// @Summary      List all tracked applications
// @Tags         applications
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/applications [get]
func (h *Handler) ListApplications(c *fiber.Ctx) error {
	apps, err := h.queries.ListApplications(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error" : "failed to list applications",
		})
	}

	resp := make([]ApplicationResponse, 0, len(apps))

	for _, a := range apps {
		resp = append(resp, toApllicationResponse(a))
	}
	return c.JSON(fiber.Map{
		"count": len(apps),
		"applications": resp,
	})

}

// UpdateApplicationRequest is the body for updating an application.
// Fields are pointers so we can tell "not provided" from "set to empty".
type UpdateApplicationRequest struct {
	Status     *string `json:"status"`
	Notes      *string `json:"notes"`
	NextAction *string `json:"next_action"`
}

// validStatuses are the allowed pipeline stages.
var validStatuses = map[string]bool{
	"new":       true,
	"applied":   true,
	"interview": true,
	"offer":     true,
	"rejected":  true,
}


// UpdateApplication godoc
// @Summary      Update an application's status or notes
// @Tags         applications
// @Accept       json
// @Produce      json
// @Param        id           path      int                       true  "Application ID"
// @Param        application  body      UpdateApplicationRequest  true  "Fields to update"
// @Success      200          {object}  map[string]interface{}
// @Failure      400          {object}  map[string]interface{}
// @Failure      404          {object}  map[string]interface{}
// @Router       /api/applications/{id} [patch]
func (h *Handler) UpdateApplication(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Application ID",
		})
	}

	var req UpdateApplicationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error" : "Invalid Request Body",
		})
	}

	// Validate status if one was provided.
	if req.Status != nil && !validStatuses[*req.Status] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error" : "Invalid Status - must be one of: new, applied, interview, offer, rejected",
		})
	}

	params := db.UpdateApplicationParams{
		ID: 		id,
		Status : 	toPgText(req.Status),
		Notes: 		toPgText(req.Notes),
		NextAction: toPgText(req.NextAction),
	}


	app, err := h.queries.UpdateApplication(c.Context(), params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Application not Found...",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error" : "Failed to Update application",
		})
	}

	return c.JSON(app)
}

// toPgText converts an optional string into a nullable pgtype.Text.
// nil pointer → NULL (field not provided, COALESCE keeps the old value).
func toPgText(s *string) pgtype.Text{
	if s == nil{
		return pgtype.Text{Valid: false}
	}

	return pgtype.Text{String: *s, Valid: true}
}

func toApllicationResponse(a db.ListApplicationsRow) ApplicationResponse {
	r := ApplicationResponse {
		ID: 			a.ID,
		JobID: 			a.JobID,
		Status:  		a.Status,
		CreatedAt: 		a.CreatedAt.Time.Format(time.RFC3339),
		JobTitle: 		a.Title,
		JobCompany: 	a.Company,
		JobURL: 		a.Url,
	}

	if a.AppliedAt.Valid {
		s := a.AppliedAt.Time.Format(time.RFC3339)
		r.AppliedAt = &s
	}
	if a.Notes.Valid {
		r.Notes = &a.Notes.String
		
	}
	if a.NextAction.Valid {
		r.NextAction = &a.NextAction.String
	}
	if a.NextActionDate.Valid {
		s := a.NextActionDate.Time.Format("2006-01-02")
		r.NextActionDate = &s
	}
	return r
}