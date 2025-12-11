package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rocky-ads/site/cache"
	"github.com/rocky-ads/site/cmd/rebuild_db/seed"
	"github.com/rocky-ads/site/db"
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

	// Remove existing database if it exists
	if _, err := os.Stat(*dbPath); err == nil {
		fmt.Printf("Removing existing database: %s\n", *dbPath)
		if err := os.Remove(*dbPath); err != nil {
			log.Fatalf("Failed to remove existing database: %v", err)
		}
	}

	// Initialize database connection and schema
	fmt.Println("Initializing database...")
	if err := initDatabaseWithSchema(*dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Load seed data
	fmt.Println("Loading seed data...")
	if err := seed.LoadAll(*includeTestAds); err != nil {
		log.Fatalf("Failed to load seed data: %v", err)
	}
	if *includeTestAds {
		fmt.Println("Seed data (including test ads) loaded successfully")
	} else {
		fmt.Println("Seed data (schema and spec data only) loaded successfully")
	}

	// Initialize caches
	fmt.Println("Initializing caches...")
	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to initialize caches: %v", err)
	}
	fmt.Println("Caches initialized successfully")

	fmt.Println("Database rebuild complete!")
}
