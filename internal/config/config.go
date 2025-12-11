package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/pflag"
)

// Config holds the application configuration
type Config struct {
	HetznerToken          string
	ListenAddress         string
	MetricsPath           string
	LogLevel              string
	CacheTTL              time.Duration
	CacheMaxSize          int64
	CacheCleanupInterval  time.Duration
	CacheStorageType      string
	ShowVersion           bool
}

// Load parses configuration from environment variables and command-line flags
func Load() (*Config, error) {
	cfg := &Config{}

	// Parse cache configuration (default: 0 = disabled, following Prometheus best practices)
	cacheTTLSeconds := 0
	var cacheTTLFlag int
	var cacheMaxSizeFlag int64
	var cacheCleanupIntervalFlag int

	// Define command-line flags
	pflag.StringVar(&cfg.ListenAddress, "listen-address", getEnv("LISTEN_ADDRESS", ":9509"),
		"Address to listen on for HTTP requests")
	pflag.StringVar(&cfg.MetricsPath, "metrics-path", getEnv("METRICS_PATH", "/metrics"),
		"Path under which to expose metrics")
	pflag.StringVar(&cfg.LogLevel, "log-level", getEnv("LOG_LEVEL", "info"),
		"Log level (debug, info, warn, error)")
	pflag.IntVar(&cacheTTLFlag, "cache-ttl", 0,
		"Cache TTL in seconds, 0 to disable (can also be set via CACHE_TTL env var, default: 0 - disabled)")
	pflag.Int64Var(&cacheMaxSizeFlag, "cache-max-size", 0,
		"Cache maximum size in bytes, 0 for unlimited (can also be set via CACHE_MAX_SIZE env var, default: 0 - unlimited)")
	pflag.IntVar(&cacheCleanupIntervalFlag, "cache-cleanup-interval", 0,
		"Cache cleanup interval in seconds, 0 for default (can also be set via CACHE_CLEANUP_INTERVAL env var, default: 0 - 10s)")
	pflag.StringVar(&cfg.CacheStorageType, "cache-storage-type", getEnv("CACHE_STORAGE_TYPE", "memory"),
		"Cache storage type (memory, redis) (can also be set via CACHE_STORAGE_TYPE env var, default: memory)")
	pflag.StringVar(&cfg.HetznerToken, "hetzner-token", os.Getenv("HETZNER_TOKEN"),
		"Hetzner API token (can also be set via HETZNER_TOKEN env var)")
	pflag.BoolVar(&cfg.ShowVersion, "version", false,
		"Show version information and exit")

	pflag.Parse()

	// Determine cache TTL: flag > env var > default (0 = disabled)
	if cacheTTLFlag > 0 {
		cacheTTLSeconds = cacheTTLFlag
	} else if envTTL := os.Getenv("CACHE_TTL"); envTTL != "" {
		if parsed, err := strconv.Atoi(envTTL); err == nil && parsed >= 0 {
			cacheTTLSeconds = parsed
		}
	}
	cfg.CacheTTL = time.Duration(cacheTTLSeconds) * time.Second

	// Determine cache max size: flag > env var > default (0 = unlimited)
	if cacheMaxSizeFlag > 0 {
		cfg.CacheMaxSize = cacheMaxSizeFlag
	} else if envSize := os.Getenv("CACHE_MAX_SIZE"); envSize != "" {
		if parsed, err := strconv.ParseInt(envSize, 10, 64); err == nil && parsed >= 0 {
			cfg.CacheMaxSize = parsed
		}
	} else {
		cfg.CacheMaxSize = 0 // 0 means unlimited
	}

	// Determine cache cleanup interval: flag > env var > default (10s)
	cleanupSeconds := 10 // default
	if cacheCleanupIntervalFlag > 0 {
		cleanupSeconds = cacheCleanupIntervalFlag
	} else if envCleanup := os.Getenv("CACHE_CLEANUP_INTERVAL"); envCleanup != "" {
		if parsed, err := strconv.Atoi(envCleanup); err == nil && parsed > 0 {
			cleanupSeconds = parsed
		}
	}
	cfg.CacheCleanupInterval = time.Duration(cleanupSeconds) * time.Second

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
