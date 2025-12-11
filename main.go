package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crstian19/prometheus-storagebox-exporter/internal/collector"
	"github.com/crstian19/prometheus-storagebox-exporter/internal/config"
	"github.com/crstian19/prometheus-storagebox-exporter/internal/hetzner"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Version information (set by build flags)
	Version   = "dev"
	GitCommit = "none"
	BuildDate = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize structured logger with JSON output
	logLevel := parseLogLevel(cfg.LogLevel)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Show version and exit if requested
	if cfg.ShowVersion {
		fmt.Printf("prometheus-storagebox-exporter\n")
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Date: %s\n", BuildDate)
		os.Exit(0)
	}

	// Initialize Hetzner API client
	hetznerClient := hetzner.NewClient(cfg.HetznerToken)

	// Create and register the storage box collector
	collector := collector.NewStorageBoxCollector(hetznerClient)
	prometheus.MustRegister(collector)

	// Set up HTTP server
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.Handle(cfg.MetricsPath, promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			// Log the error but don't fail the health check
			slog.Warn("Failed to write health check response", "error", err)
		}
	})

	// Landing page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
	<title>Prometheus Hetzner Storage Box Exporter</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 40px; }
		h1 { color: #333; }
		a { color: #0066cc; text-decoration: none; }
		a:hover { text-decoration: underline; }
		.info { background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0; }
	</style>
</head>
<body>
	<h1>Prometheus Hetzner Storage Box Exporter</h1>
	<div class="info">
		<p><strong>Version:</strong> %s</p>
		<p><strong>Git Commit:</strong> %s</p>
		<p><strong>Build Date:</strong> %s</p>
	</div>
	<p><a href="%s">Metrics</a></p>
	<p><a href="/health">Health Check</a></p>
	<h2>About</h2>
	<p>This exporter collects metrics from Hetzner Storage Boxes and exposes them in Prometheus format.</p>
	<h3>Metrics Exposed:</h3>
	<ul>
		<li>storagebox_disk_usage_bytes - Total used diskspace</li>
		<li>storagebox_disk_usage_data_bytes - Diskspace used by files</li>
		<li>storagebox_disk_usage_snapshots_bytes - Diskspace used by snapshots</li>
		<li>storagebox_info - Storage box information</li>
		<li>storagebox_status - Current status</li>
		<li>storagebox_access_*_enabled - Access settings (SSH, Samba, WebDAV)</li>
		<li>storagebox_snapshot_plan_enabled - Snapshot plan status</li>
		<li>storagebox_protection_delete - Delete protection status</li>
		<li>storagebox_created_timestamp - Creation timestamp</li>
	</ul>
</body>
</html>
`, Version, GitCommit, BuildDate, cfg.MetricsPath)
	})

	server := &http.Server{
		Addr:         cfg.ListenAddress,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Starting prometheus-storagebox-exporter",
			"version", Version,
			"git_commit", GitCommit,
			"build_date", BuildDate,
			"listen_address", cfg.ListenAddress,
			"metrics_path", cfg.MetricsPath,
			"log_level", cfg.LogLevel,
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-stop

	slog.Info("Shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Error during shutdown", "error", err)
	}

	slog.Info("Exporter stopped")
}

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
