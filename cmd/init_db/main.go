package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rocky-ads/site/cmd/init_db/seed"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/logger"
)

func setupDatabase(databaseURL string) error {
	logger.Info("Dropping all existing tables")
	var dropSQL sql.NullString
	err := db.QueryRow(`
		SELECT COALESCE('DROP TABLE IF EXISTS ' || string_agg('"' || tablename || '"', ', ') || ' CASCADE', '')
		FROM pg_tables
		WHERE schemaname = 'public'
	`).Scan(&dropSQL)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("generating DROP statement: %w", err)
	}

	if dropSQL.Valid && dropSQL.String != "" {
		if _, err := db.Exec(dropSQL.String); err != nil {
			return fmt.Errorf("dropping tables: %w", err)
		}
		logger.Info("All existing tables dropped")
	}

	schemaPath := "internal/db/schema.sql"
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		schemaPath = cwd + "/internal/db/schema.sql"
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("reading schema file: %w", err)
	}

	logger.Info("Executing schema.sql using pgx")
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return fmt.Errorf("connecting with pgx: %w", err)
	}
	defer conn.Close(context.Background())

	if _, err := conn.Exec(context.Background(), string(schema)); err != nil {
		return fmt.Errorf("executing schema: %w", err)
	}

	return nil
}

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
	if err := setupDatabase(databaseURL); err != nil {
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
