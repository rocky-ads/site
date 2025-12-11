package main

import (
	"time"

	"github.com/rocky-ads/site/config"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/handlers"
	"github.com/rocky-ads/site/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
)

// setupApp configures the Fiber app with middleware and routes
func setupApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: handlers.ErrorHandler,
		BodyLimit:    config.ServerBodyLimit, // Total request body size (for multiple files)
		ReadTimeout:  30 * time.Second,       // Prevent long-running requests
		WriteTimeout: 30 * time.Second,       // Prevent long-running responses
	})

	// Must be early in middleware chain
	app.Use(handlers.ConfigureHelmet())

	app.Use(limiter.New(limiter.Config{
		Max:        config.ServerRateLimitMax,
		Expiration: config.ServerRateLimitExp,
	}))

	app.Use(handlers.JWTMiddleware)
	app.Use(handlers.CSRFMiddleware)

	app.Use(flogger.New(flogger.Config{
		Output: logger.Writer(),
		Format: "${status} | ${latency} | ${ip} | ${method} | ${path}\n",
	}))

	api := app.Group("/api")

	categoryRouter := api.Group("/categories/:category")
	categoryRouter.Get("/values/:field", handlers.GetAllValuesHandler)
	categoryRouter.Get("/any-values/:field", handlers.GetAnyValuesHandler)
	categoryRouter.Post("/ad-values/:field", handlers.GetAdValuesHandler)
	categoryRouter.Get("/chains", handlers.GetChainsHandler)
	categoryRouter.Get("/first-spec-fields", handlers.GetFirstSpecFieldsHandler)
	categoryRouter.Get("/last-spec-field", handlers.GetLastSpecFieldHandler)
	categoryRouter.Post("/search", handlers.SearchHandler)

	api.Get("/ads/:id/filter-values", handlers.GetAdFilterValuesHandler)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	return app
}

func main() {
	if err := logger.Init(config.LogLevel, config.LogFormat,
		config.LogFile); err != nil {
		logger.Fatal("Failed to initialize logger", "error", err)
	}

	// Assumes database is already built with cmd/rebuild_db
	if err := db.Init("project.db"); err != nil {
		logger.Fatal("Failed to open database", "error", err)
	}
	defer db.Close()

	app := setupApp()

	port := ":" + config.ServerPort
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
