package main

import (
	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/handlers"
	"github.com/rocky-ads/site/logger"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initialize logger
	if err := logger.Init(config.LogLevel, config.LogFormat,
		config.LogFile); err != nil {
		logger.Fatal("Failed to initialize logger", "error", err)
	}

	// Open existing database (assumes it's already been built with cmd/rebuild_db)
	if err := db.Init("project.db"); err != nil {
		logger.Fatal("Failed to open database", "error", err)
	}
	defer db.Close()

	// Set up Fiber app
	app := fiber.New()

	// API routes
	api := app.Group("/api")

	// Category-specific routes
	categoryRouter := api.Group("/categories/:category")
	categoryRouter.Get("/values/:field", handlers.GetAllValuesHandler)
	categoryRouter.Get("/any-values/:field", handlers.GetAnyValuesHandler)
	categoryRouter.Post("/ad-values/:field", handlers.GetAdValuesHandler)
	categoryRouter.Get("/chains", handlers.GetChainsHandler)
	categoryRouter.Get("/first-spec-fields", handlers.GetFirstSpecFieldsHandler)
	categoryRouter.Get("/last-spec-field", handlers.GetLastSpecFieldHandler)
	categoryRouter.Post("/search", handlers.SearchHandler)

	// Ad routes
	api.Get("/ads/:id/filter-values", handlers.GetAdFilterValuesHandler)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Start server
	port := ":8080"
	logger.Info("Server starting", "port", port)
	logger.Info("API endpoints:")
	logger.Info("  GET  /api/categories/:category/values/:field")
	logger.Info("  GET  /api/categories/:category/any-values/:field")
	logger.Info("  POST /api/categories/:category/ad-values/:field")
	logger.Info("  GET  /api/categories/:category/chains")
	logger.Info("  GET  /api/categories/:category/first-spec-fields")
	logger.Info("  GET  /api/categories/:category/last-spec-field")
	logger.Info("  POST /api/categories/:category/search")
	logger.Info("  GET  /api/ads/:id/filter-values")

	if err := app.Listen(port); err != nil {
		logger.Fatal("Server failed to start", "error", err)
	}
}
