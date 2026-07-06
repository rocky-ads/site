package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/testagent"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init("info", "text", ""); err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		os.Exit(1)
	}

	if err := pingSite(cfg.SiteURL); err != nil {
		logger.Fatal("site unreachable", "url", cfg.SiteURL, "error", err)
	}

	agentCfg := testagent.Config{
		SiteURL:  cfg.SiteURL,
		StateDir: cfg.StateDir,
		MinDelay: cfg.MinDelay,
		MaxDelay: cfg.MaxDelay,
	}

	reg, err := testagent.NewRegistry(cfg.AgentCount, agentCfg)
	if err != nil {
		logger.Fatal("registry", "error", err)
	}

	logger.Info("test agents ready (all stopped — start via control UI)",
		"count", cfg.AgentCount, "state_dir", cfg.StateDir)

	if err := runServer(cfg, reg); err != nil {
		logger.Fatal("server", "error", err)
	}
}

func pingSite(baseURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health status %d", resp.StatusCode)
	}
	return nil
}
