package main

import (
	"context"
	"log"
	"os"

	_ "github.com/DGreegman/gohunt/docs"
	"github.com/DGreegman/gohunt/internal/api"
	"github.com/DGreegman/gohunt/internal/database"
	"github.com/DGreegman/gohunt/internal/fetcher"
	"github.com/DGreegman/gohunt/internal/scorer"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title           GoHunt API
// @version         1.1
// @description     AI-powered job aggregation and application assistant.
// @host            localhost:8080
// @BasePath        /
func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, databaseURL)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Connected to Database")

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == ""{
		log.Fatal("ANTHROPIC_API_KEY is not set")
	}
	sc := scorer.NewScorer(apiKey)

	testProfile := scorer.ProfileInput {
		TargetRoles:     []string{"Backend Engineer", "Golang Developer"},
		Skills:          []string{"Go", "PostgreSQL", "Docker", "Fiber"},
		ExperienceYears: 3,
		PreferredStack:  []string{"Go", "PostgreSQL"},
		LocationPref:    "Remote",
		RemoteOnly:      true,
	}
	result, err := sc.Score(ctx, "Senior Backend Engineer (Go)",
		"We're hiring a Go engineer to build microservices with PostgreSQL and Docker. Remote-friendly.",
		"Remote",
		testProfile,)
	if err != nil {
		log.Fatalf("Score failed: %v", err)
	}
	log.Printf("SCORE: total=%d role=%d skill=%d seniority=%d stack=%d location=%d",
    result.Total(), result.RoleMatch, result.SkillOverlap, result.SeniorityFit,
    result.StackAlignment, result.LocationCompFit)
	log.Printf("RATIONALE: %s", result.Rationale)


	sources := []fetcher.JobSource{
		fetcher.NewRemotiveSource(),
		fetcher.NewGreenhouseSource("gitlab"),
	}

	source := fetcher.NewPool(sources, 5)
	
	store := fetcher.NewStore(pool)
	handler := api.NewHandler(pool, source, store)

	app := fiber.New(fiber.Config{
		AppName: "GoHunt v1.1",
	})

	
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})
	app.Get("/api/jobs", handler.ListJobs)
	app.Post("/api/fetch/trigger", handler.TriggerFetch)
	app.Get("/api/profile", handler.GetProfile)
	app.Put("/api/profile", handler.UpsertProfile)
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	log.Fatal(app.Listen(":8080"))
}