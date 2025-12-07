package config

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

// Config holds the application configuration
type Config struct {
	HetznerToken  string
	ListenAddress string
	MetricsPath   string
	LogLevel      string
	ShowVersion   bool
}

// Load parses configuration from environment variables and command-line flags
func Load() (*Config, error) {
	cfg := &Config{}

	// Define command-line flags
	pflag.StringVar(&cfg.ListenAddress, "listen-address", getEnv("LISTEN_ADDRESS", ":9509"),
		"Address to listen on for HTTP requests")
	pflag.StringVar(&cfg.MetricsPath, "metrics-path", getEnv("METRICS_PATH", "/metrics"),
		"Path under which to expose metrics")
	pflag.StringVar(&cfg.LogLevel, "log-level", getEnv("LOG_LEVEL", "info"),
		"Log level (debug, info, warn, error)")
	pflag.StringVar(&cfg.HetznerToken, "hetzner-token", os.Getenv("HETZNER_TOKEN"),
		"Hetzner API token (can also be set via HETZNER_TOKEN env var)")
	pflag.BoolVar(&cfg.ShowVersion, "version", false,
		"Show version information and exit")

	pflag.Parse()

	// Validate required configuration
	if !cfg.ShowVersion && cfg.HetznerToken == "" {
		return nil, fmt.Errorf("HETZNER_TOKEN environment variable or --hetzner-token flag is required")
	}

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
