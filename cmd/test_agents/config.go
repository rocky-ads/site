package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type envConfig struct {
	SiteURL    string
	Port       string
	AgentCount int
	StateDir   string
	MinDelay   time.Duration
	MaxDelay   time.Duration
}

func loadConfig() (envConfig, error) {
	siteURL := os.Getenv("SITE_URL")
	if siteURL == "" {
		return envConfig{}, fmt.Errorf("SITE_URL is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "10002"
	}
	count := 10
	if v := os.Getenv("AGENT_COUNT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return envConfig{}, fmt.Errorf("AGENT_COUNT: %w", err)
		}
		count = n
	}
	stateDir := os.Getenv("STATE_DIR")
	if stateDir == "" {
		stateDir = "./.test-agents-state"
	}
	minDelay := 30 * time.Second
	if v := os.Getenv("AGENT_MIN_DELAY"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return envConfig{}, fmt.Errorf("AGENT_MIN_DELAY: %w", err)
		}
		minDelay = d
	}
	maxDelay := 3 * time.Minute
	if v := os.Getenv("AGENT_MAX_DELAY"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return envConfig{}, fmt.Errorf("AGENT_MAX_DELAY: %w", err)
		}
		maxDelay = d
	}
	return envConfig{
		SiteURL:    siteURL,
		Port:       port,
		AgentCount: count,
		StateDir:   stateDir,
		MinDelay:   minDelay,
		MaxDelay:   maxDelay,
	}, nil
}
