package api

import (
	"strconv"
	"time"

	"github.com/DGreegman/gohunt/internal/db"
	"github.com/DGreegman/gohunt/internal/fetcher"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// JobResponse is the API shape for a job (decoupled from the DB row).
type JobResponse struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Company   string  `json:"company"`
	Source    string  `json:"source"`
	URL       string  `json:"url"`
	Location  string  `json:"location"`
	Remote    bool    `json:"remote"`
	PostedAt  *string `json:"posted_at"`
	LinkStatus string `json:"link_status"`
	CreatedAt string  `json:"created_at"`
}

// Handler holds dependencies shared across HTTP handlers.
type Handler struct {
	queries *db.Queries
	source fetcher.JobSource
	store  *fetcher.Store
}

// NewHandler wires handlers to the connection pool.
func NewHandler(pool *pgxpool.Pool, source fetcher.JobSource, store *fetcher.Store) *Handler {
	return &Handler{
		queries: db.New(pool),
		source: source,
		store: store,
	}
}

// ListJobs handles GET /api/jobs?limit=&offset=
// ListJobs godoc
// @Summary      List jobs
// @Description  Returns a paginated list of aggregated jobs
// @Tags         jobs
// @Produce      json
// @Param        limit   query     int  false  "Max results (default 20)"
// @Param        offset  query     int  false  "Rows to skip (default 0)"
// @Success      200     {object}  map[string]interface{}
// @Failure      500     {object}  map[string]interface{}
// @Router       /api/jobs [get]
func (h *Handler) ListJobs(c *fiber.Ctx) error {
	limit := parseIntDefault(c.Query("limit"), 20)
	offset := parseIntDefault(c.Query("offset"), 0)

	jobs, err := h.queries.ListJobs(c.Context(), db.ListJobsParams{
		Limit: int32(limit),
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
	jobs, err := h.source.FetchJobs(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Fetch Failed...",
		})
	}

	inserted, err := h.store.SaveJobs(c.Context(), jobs)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error" : "Failed to Save Jobs",
		})
	}

	return c.JSON(fiber.Map{
		"source": h.source.Name(),
		"fetched": len(jobs),
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
			"error" : "no profile found - create one with PUT /api/profile",
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
			"error" : "invalid request body",
		})
	}

	profile, err := h.queries.UpsertProfile(c.Context(), db.UpsertProfileParams{
		Name: 				req.Name,
		CurrentRoleTitle: 	pgtype.Text{String: req.CurrentRoleTitle, Valid: req.CurrentRoleTitle != ""},
		Skills: 			req.Skills,	
		ExperienceYears:    pgtype.Int4{Int32: req.ExperienceYears, Valid: true},
		TargetRoles:        req.TargetRoles,
		PreferredStack:     req.PreferredStack,
		SalaryMin:          pgtype.Int4{Int32: req.SalaryMin, Valid: true},
		SalaryMax:          pgtype.Int4{Int32: req.SalaryMax, Valid: true},
		LocationPref:       pgtype.Text{String: req.LocationPref, Valid: req.LocationPref != ""},
		RemoteOnly:         req.RemoteOnly,
		Timezone:           req.Timezone,
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save Profile",
		})
	}

	return c.JSON(profile)
}