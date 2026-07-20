package main

import (
	"fmt"
	"os"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/user"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	name := os.Args[2]
	if cmd != "promote" && cmd != "demote" {
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

	u, err := user.GetByName(name)
	if err != nil {
		logger.Fatal("User not found", "name", name, "error", err)
	}

	switch cmd {
	case "promote":
		if u.IsAdmin {
			logger.Info("User is already an admin", "name", name, "id", u.ID)
			return
		}
		if err := user.PromoteToAdmin(u.ID); err != nil {
			logger.Fatal("Failed to promote user", "name", name, "error", err)
		}
		logger.Info("Promoted user to admin", "name", name, "id", u.ID)
	case "demote":
		if !u.IsAdmin {
			logger.Info("User is not an admin", "name", name, "id", u.ID)
			return
		}
		if err := user.DemoteFromAdmin(u.ID); err != nil {
			logger.Fatal("Failed to demote user", "name", name, "error", err)
		}
		logger.Info("Demoted user from admin", "name", name, "id", u.ID)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  set_admin promote <name>
  set_admin demote <name>
`)
}
