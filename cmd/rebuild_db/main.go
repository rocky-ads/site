package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rocky-ads/site/cmd/rebuild_db/seed"
	"github.com/rocky-ads/site/db"
	"github.com/rocky-ads/site/logger"
)

// initDatabaseWithSchema initializes the database and loads the schema
func initDatabaseWithSchema(dbPath string) error {
	// Open database connection
	if err := db.Init(dbPath); err != nil {
		return fmt.Errorf("opening database: %w", err)
	}

	// Read and execute schema
	schemaPath := "db/schema.sql"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// Try from current working directory
		cwd, _ := os.Getwd()
		schemaPath = cwd + "/db/schema.sql"
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("reading schema file: %w", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("executing schema: %w", err)
	}

	return nil
}

func main() {
	dbPath := flag.String("db", "project.db", "Path to database file")
	includeTestAds := flag.Bool("test-ads", false, "Include test ads in seed data")
	flag.Parse()

	startTime := time.Now()

	// Initialize logger
	if err := logger.Init("info", "text", ""); err != nil {
		logger.Fatal("Failed to initialize logger", "error", err)
	}

	// Remove existing database if it exists
	stepStart := time.Now()
	if _, err := os.Stat(*dbPath); err == nil {
		logger.Info("Removing existing database", "path", *dbPath)
		if err := os.Remove(*dbPath); err != nil {
			logger.Fatal("Failed to remove existing database", "error", err)
		}
	}
	logger.Info("Remove database step", "duration", time.Since(stepStart))

	// Initialize database connection and schema
	stepStart = time.Now()
	logger.Info("Initializing database...")
	if err := initDatabaseWithSchema(*dbPath); err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}
	defer db.Close()
	logger.Info("Init database step", "duration", time.Since(stepStart))

	// Load seed data
	stepStart = time.Now()
	logger.Info("Loading seed data...")
	if err := seed.LoadAll(*includeTestAds); err != nil {
		logger.Fatal("Failed to load seed data", "error", err)
	}
	logger.Info("Load seed data step", "duration", time.Since(stepStart))
	if *includeTestAds {
		logger.Info("Seed data (including test ads) loaded successfully")
	} else {
		logger.Info("Seed data (schema and spec data only) loaded successfully")
	}

	totalDuration := time.Since(startTime)
	logger.Info("Database rebuild complete!", "total_duration", totalDuration)
}
