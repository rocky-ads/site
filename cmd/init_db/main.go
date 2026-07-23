package main

import (
	"flag"
	"time"

	"github.com/rocky-ads/site/cmd/init_db/seed"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/logger"
)

func main() {
	loadSeed := flag.Bool(
		"load-seed", false,
		"Also load seed users and ads (categories always load)",
	)
	flag.Parse()

	startTime := time.Now()

	if err := logger.Init("info", "text", ""); err != nil {
		logger.Fatal("Failed to initialize logger", "error", err)
	}

	databaseURL := config.DatabaseURL
	if databaseURL == "" {
		logger.Fatal("DATABASE_URL must be set")
	}

	stepStart := time.Now()
	logger.Info("Initializing database...")
	if err := db.Init(databaseURL); err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}
	defer db.Close()
	logger.Info("Init database step", "duration", time.Since(stepStart))

	stepStart = time.Now()
	if err := db.ResetSchema(); err != nil {
		logger.Fatal("Failed to setup database", "error", err)
	}
	logger.Info("Setup database step", "duration", time.Since(stepStart))

	stepStart = time.Now()
	if *loadSeed {
		logger.Info("Loading seed data...")
		if err := seed.LoadAll(); err != nil {
			logger.Fatal("Failed to load seed data", "error", err)
		}
	} else {
		logger.Info("Loading categories only (skipping users and ads)...")
		if err := seed.LoadCategories(); err != nil {
			logger.Fatal("Failed to load categories", "error", err)
		}
	}
	logger.Info("Load seed data step", "duration", time.Since(stepStart))
	logger.Info("Seed data loaded successfully")

	logger.Info("Database rebuild complete!", "total_duration", time.Since(startTime))
}
