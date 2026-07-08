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


	sources := []fetcher.JobSource{
		fetcher.NewRemotiveSource(),
		fetcher.NewGreenhouseSource("gitlab"),
	}

	source := fetcher.NewPool(sources, 5)
	
	store := fetcher.NewStore(pool)
	handler := api.NewHandler(pool, source, store, sc)

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
	app.Post("/api/score/trigger", handler.TriggerScoring)
	app.Get("/api/profile", handler.GetProfile)
	app.Put("/api/profile", handler.UpsertProfile)
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
	log.Fatal(app.Listen(":8080"))
}