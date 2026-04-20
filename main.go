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

	// Create and register the storage box collector with cache
	collector := collector.NewStorageBoxCollector(hetznerClient, cfg.CacheTTL, cfg.CacheMaxSize, cfg.CacheCleanupInterval)
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
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Prometheus Hetzner Storage Box Exporter</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        .header {
            background: white;
            color: #dc2626;
            padding: 40px;
            text-align: center;
        }
        .logo {
            margin-bottom: 20px;
        }
        .logo img {
            width: 100px;
            height: 100px;
        }
        h1 {
            font-size: 28px;
            font-weight: 700;
            margin-bottom: 8px;
        }
        .subtitle {
            font-size: 16px;
            opacity: 0.8;
            font-weight: 400;
            color: #991b1b;
        }
        .content {
            padding: 40px;
        }
        .info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .info-card {
            background: linear-gradient(135deg, #f3f4f6 0%%, #e5e7eb 100%%);
            padding: 20px;
            border-radius: 12px;
            border-left: 4px solid #dc2626;
        }
        .info-card h3 {
            font-size: 12px;
            text-transform: uppercase;
            color: #6b7280;
            margin-bottom: 8px;
            font-weight: 600;
            letter-spacing: 0.5px;
        }
        .info-card p {
            font-size: 16px;
            color: #1f2937;
            font-weight: 500;
            word-break: break-all;
        }
        .links {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-bottom: 30px;
        }
        .btn {
            display: inline-block;
            padding: 14px 24px;
            border-radius: 8px;
            text-decoration: none;
            font-weight: 600;
            text-align: center;
            transition: all 0.3s ease;
            font-size: 14px;
        }
        .btn-primary {
            background: linear-gradient(135deg, #dc2626 0%%, #b91c1c 100%%);
            color: white;
            box-shadow: 0 4px 6px rgba(220, 38, 38, 0.3);
        }
        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 12px rgba(220, 38, 38, 0.4);
        }
        .btn-secondary {
            background: white;
            color: #dc2626;
            border: 2px solid #dc2626;
        }
        .btn-secondary:hover {
            background: #fee2e2;
            transform: translateY(-2px);
        }
        .btn-github {
            background: linear-gradient(135deg, #24292e 0%%, #1a1f23 100%%);
            color: white;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.2);
        }
        .btn-github:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 12px rgba(0, 0, 0, 0.3);
        }
        .btn svg {
            width: 18px;
            height: 18px;
            margin-right: 8px;
            vertical-align: middle;
        }
        .section {
            margin-bottom: 30px;
        }
        .section h2 {
            font-size: 20px;
            color: #1f2937;
            margin-bottom: 16px;
            padding-bottom: 8px;
            border-bottom: 2px solid #e5e7eb;
        }
        .metrics-list {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
            gap: 12px;
        }
        .metric-item {
            background: #f9fafb;
            padding: 12px 16px;
            border-radius: 8px;
            border-left: 3px solid #10b981;
            font-size: 14px;
            color: #374151;
        }
        .metric-name {
            font-family: 'Courier New', monospace;
            font-weight: 600;
            color: #059669;
        }
        .footer {
            background: #f9fafb;
            padding: 20px 40px;
            text-align: center;
            color: #6b7280;
            font-size: 14px;
            border-top: 1px solid #e5e7eb;
        }
        .footer a {
            color: #dc2626;
            text-decoration: none;
        }
        .footer a:hover {
            text-decoration: underline;
        }
        @media (max-width: 768px) {
            .header {
                padding: 30px 20px;
            }
            .content {
                padding: 30px 20px;
            }
            h1 {
                font-size: 24px;
            }
            .info-grid {
                grid-template-columns: 1fr;
            }
            .links {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">
                <img src="https://cdn.crstian.me/storage-box-exporter-logo.png" alt="Storage Box Exporter Logo">
            </div>
            <h1>Prometheus Storage Box Exporter</h1>
            <p class="subtitle">Hetzner Storage Box Metrics for Prometheus</p>
        </div>
        <div class="content">
            <div class="info-grid">
                <div class="info-card">
                    <h3>Version</h3>
                    <p>%s</p>
                </div>
                <div class="info-card">
                    <h3>Git Commit</h3>
                    <p>%s</p>
                </div>
                <div class="info-card">
                    <h3>Build Date</h3>
                    <p>%s</p>
                </div>
            </div>
            <div class="links">
                <a href="%s" class="btn btn-primary">
                    <svg fill="currentColor" viewBox="0 0 20 20">
                        <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z"/>
                    </svg>
                    View Metrics
                </a>
                <a href="/health" class="btn btn-secondary">
                    <svg fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                    </svg>
                    Health Check
                </a>
                <a href="https://github.com/crstian19/prometheus-storagebox-exporter" class="btn btn-github" target="_blank" rel="noopener">
                    <svg fill="currentColor" viewBox="0 0 24 24">
                        <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/>
                    </svg>
                    GitHub Repository
                </a>
            </div>
            <div class="section">
                <h2>About</h2>
                <p style="color: #4b5563; line-height: 1.6;">
                    This exporter collects comprehensive metrics from Hetzner Storage Boxes via the Hetzner Cloud API
                    and exposes them in Prometheus format. Monitor disk usage, access settings, protection status, and more.
                </p>
            </div>
            <div class="section">
                <h2>Exported Metrics</h2>
                <div class="metrics-list">
                    <div class="metric-item">
                        <span class="metric-name">storagebox_disk_quota_bytes</span><br>
                        Total storage quota
                    </div>
                    <div class="metric-item">
                        <span class="metric-name">storagebox_disk_usage_bytes</span><br>
                        Total used disk space
                    </div>
                    <div class="metric-item">
                        <span class="metric-name">storagebox_disk_usage_data_bytes</span><br>
                        Disk space used by files
                    </div>
                    <div class="metric-item">
                        <span class="metric-name">storagebox_disk_usage_snapshots_bytes</span><br>
                        Disk space used by snapshots
                    </div>
                    <div class="metric-item">
                        <span class="metric-name">storagebox_info</span><br>
                        Storage box information
                    </div>
                    <div class="metric-item">
                        <span class="metric-name">storagebox_status</span><br>
                        Current operational status
                    </div>
                    <div class="metric-item">
                        <span class="metric-name">storagebox_access_*_enabled</span><br>
                        Access settings (SSH, Samba, WebDAV, ZFS)
                    </div>
                    <div class="metric-item">
                        <span class="metric-name">storagebox_snapshot_plan_enabled</span><br>
                        Snapshot plan status
                    </div>
                    <div class="metric-item">
                        <span class="metric-name">storagebox_protection_delete</span><br>
                        Delete protection status
                    </div>
                    <div class="metric-item">
                        <span class="metric-name">storagebox_created_timestamp</span><br>
                        Creation timestamp
                    </div>
                    <div class="metric-item">
                        <span class="metric-name">storagebox_exporter_*</span><br>
                        Exporter performance & error metrics
                    </div>
                </div>
            </div>
        </div>
        <div class="footer">
            Made with ❤️ by <a href="https://github.com/crstian19" target="_blank" rel="noopener">@crstian19</a> •
            Built with <a href="https://claude.ai" target="_blank" rel="noopener">Claude.ai</a> •
            <a href="https://prometheus.io" target="_blank" rel="noopener">Prometheus</a> •
            <a href="https://www.hetzner.com" target="_blank" rel="noopener">Hetzner</a>
        </div>
    </div>
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
