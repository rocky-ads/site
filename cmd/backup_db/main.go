package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/imagestore"
	"github.com/rocky-ads/site/internal/logger"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	if err := logger.Init("info", "text", ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	databaseURL := config.DatabaseURL
	if databaseURL == "" {
		logger.Fatal("DATABASE_URL must be set")
	}
	if err := db.Init(databaseURL); err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}
	defer db.Close()

	switch os.Args[1] {
	case "backup":
		runBackupCmd(os.Args[2:])
	case "restore":
		runRestoreCmd(os.Args[2:])
	case "migrate-schema":
		runMigrateSchemaCmd(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func runBackupCmd(args []string) {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	outDir := fs.String("out", "", "Backup output directory")
	dryRun := fs.Bool("dry-run", false, "Preview without writing")
	verbose := fs.Bool("verbose", false, "Log per-ad image progress")
	fs.Parse(args)
	if *outDir == "" {
		logger.Fatal("-out is required")
	}
	store, err := imagestore.NewDefault()
	if err != nil {
		logger.Fatal("Failed to initialize image store", "error", err)
	}
	if err := runBackup(*outDir, store, *dryRun, *verbose); err != nil {
		logger.Fatal("Backup failed", "error", err)
	}
}

func runRestoreCmd(args []string) {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	fromDir := fs.String("from", "", "Backup directory to restore")
	dryRun := fs.Bool("dry-run", false, "Preview without writing")
	verbose := fs.Bool("verbose", false, "Log per-ad image progress")
	fs.Parse(args)
	if *fromDir == "" {
		logger.Fatal("-from is required")
	}
	store, err := imagestore.NewDefault()
	if err != nil {
		logger.Fatal("Failed to initialize image store", "error", err)
	}
	if err := runRestore(*fromDir, store, *dryRun, *verbose); err != nil {
		logger.Fatal("Restore failed", "error", err)
	}
}

func runMigrateSchemaCmd(args []string) {
	fs := flag.NewFlagSet("migrate-schema", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Preview without altering")
	fs.Parse(args)
	if err := migratePhoneLifecycleSchema(*dryRun); err != nil {
		logger.Fatal("Schema migration failed", "error", err)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  backup_db backup         -out <dir> [-dry-run] [-verbose]
  backup_db restore        -from <dir> [-dry-run] [-verbose]
  backup_db migrate-schema [-dry-run]

migrate-schema is a one-time upgrade from the pre-phone-lifecycle schema
(global unique phone_hash; phone_verification without purpose) to the current
schema. restore also runs it idempotently before importing.
`)
}
