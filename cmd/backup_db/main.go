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
	if isHelp(os.Args[1]) {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "backup":
		runBackupCmd(os.Args[2:])
	case "restore":
		runRestoreCmd(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func runBackupCmd(args []string) {
	fs := flag.NewFlagSet("backup", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	out := fs.String("out", "", "Output .tar.gz path (.tar.gz appended if omitted)")
	dryRun := fs.Bool("dry-run", false, "Preview without writing")
	verbose := fs.Bool("verbose", false, "Log per-ad image progress")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage:\n  backup_db backup -out <path.tar.gz> [-dry-run] [-verbose]\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err == flag.ErrHelp {
		os.Exit(0)
	} else if err != nil {
		os.Exit(2)
	}
	if *out == "" {
		fs.Usage()
		os.Exit(2)
	}

	mustInit()
	defer db.Close()

	store, err := imagestore.NewDefault()
	if err != nil {
		logger.Fatal("Failed to initialize image store", "error", err)
	}
	if err := backupToArchive(*out, store, *dryRun, *verbose); err != nil {
		logger.Fatal("Backup failed", "error", err)
	}
}

func runRestoreCmd(args []string) {
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	from := fs.String("from", "", "Input .tar.gz path (.tar.gz appended if omitted)")
	dryRun := fs.Bool("dry-run", false, "Preview without writing")
	verbose := fs.Bool("verbose", false, "Log per-ad image progress")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage:\n  backup_db restore -from <path.tar.gz> [-dry-run] [-verbose]\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err == flag.ErrHelp {
		os.Exit(0)
	} else if err != nil {
		os.Exit(2)
	}
	if *from == "" {
		fs.Usage()
		os.Exit(2)
	}

	mustInit()
	defer db.Close()

	store, err := imagestore.NewDefault()
	if err != nil {
		logger.Fatal("Failed to initialize image store", "error", err)
	}
	if err := restoreFromArchive(*from, store, *dryRun, *verbose); err != nil {
		logger.Fatal("Restore failed", "error", err)
	}
}

func mustInit() {
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
}

func isHelp(arg string) bool {
	return arg == "-h" || arg == "-help" || arg == "--help" || arg == "help"
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  backup_db backup  -out <path.tar.gz> [-dry-run] [-verbose]
  backup_db restore -from <path.tar.gz> [-dry-run] [-verbose]
`)
}

func backupToArchive(out string, store imagestore.Store, dryRun, verbose bool) error {
	if dryRun {
		return runBackup("", store, true, verbose)
	}
	archivePath := resolveArchivePath(out)
	staging, err := os.MkdirTemp("", "backup_db-*")
	if err != nil {
		return fmt.Errorf("create staging dir: %w", err)
	}
	defer os.RemoveAll(staging)

	if err := runBackup(staging, store, false, verbose); err != nil {
		return err
	}
	if err := createTarGz(archivePath, staging); err != nil {
		return err
	}
	logger.Info("Wrote archive", "path", archivePath)
	return nil
}

func restoreFromArchive(from string, store imagestore.Store, dryRun, verbose bool) error {
	archivePath := resolveArchivePath(from)
	staging, err := os.MkdirTemp("", "backup_db-restore-*")
	if err != nil {
		return fmt.Errorf("create staging dir: %w", err)
	}
	defer os.RemoveAll(staging)

	if err := extractTarGz(archivePath, staging); err != nil {
		return err
	}
	return runRestore(staging, store, dryRun, verbose)
}
