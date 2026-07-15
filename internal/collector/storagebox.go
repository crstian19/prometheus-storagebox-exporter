package collector

import (
	"context"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/crstian19/prometheus-storagebox-exporter/internal/cache"
	"github.com/crstian19/prometheus-storagebox-exporter/internal/hetzner"
	"github.com/prometheus/client_golang/prometheus"
)

// BuildInfo carries version metadata injected at build time, exposed as the
// storagebox_exporter_build_info metric.
type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

// StorageBoxCollector implements the prometheus.Collector interface
type StorageBoxCollector struct {
	client       *hetzner.Client
	cache        *cache.MetricsCache
	cacheEnabled bool

	// Core storage metrics
	diskQuota          *prometheus.Desc
	diskUsage          *prometheus.Desc
	diskUsageData      *prometheus.Desc
	diskUsageSnapshots *prometheus.Desc

	// Info and status metrics
	info              *prometheus.Desc
	status            *prometheus.Desc
	accessSSH         *prometheus.Desc
	accessSamba       *prometheus.Desc
	accessWebDAV      *prometheus.Desc
	accessZFS         *prometheus.Desc
	reachableExternal *prometheus.Desc
	snapshotPlan      *prometheus.Desc
	protectionDelete  *prometheus.Desc
	createdTimestamp  *prometheus.Desc

	// Exporter metrics
	up             *prometheus.Desc
	buildInfo      *prometheus.Desc
	buildInfoData  BuildInfo
	scrapeDuration *prometheus.Desc
	scrapeErrors   prometheus.Counter
	cacheHits      prometheus.Counter
	cacheMisses    prometheus.Counter

	// Error type metrics
	authErrors      prometheus.Counter
	rateLimitErrors prometheus.Counter
	serverErrors    prometheus.Counter
	clientErrors    prometheus.Counter
	networkErrors   prometheus.Counter
}

// NewStorageBoxCollector creates a new StorageBoxCollector
func NewStorageBoxCollector(client *hetzner.Client, cacheTTL time.Duration, cacheMaxSize int64, cacheCleanupInterval time.Duration, buildInfo BuildInfo) *StorageBoxCollector {
	cacheEnabled := cacheTTL > 0
	return &StorageBoxCollector{
		client:        client,
		cache:         cache.NewMetricsCache(cacheTTL, cacheMaxSize, cacheCleanupInterval),
		cacheEnabled:  cacheEnabled,
		buildInfoData: buildInfo,

		// Core storage metrics
		diskQuota: prometheus.NewDesc(
			"storagebox_disk_quota_bytes",
			"Total allocated diskspace in bytes",
			[]string{"id", "name", "server", "location"},
			nil,
		),
		diskUsage: prometheus.NewDesc(
			"storagebox_disk_usage_bytes",
			"Total used diskspace in bytes",
			[]string{"id", "name", "server", "location"},
			nil,
		),
		diskUsageData: prometheus.NewDesc(
			"storagebox_disk_usage_data_bytes",
			"Diskspace used by files in bytes",
			[]string{"id", "name", "server", "location"},
			nil,
		),
		diskUsageSnapshots: prometheus.NewDesc(
			"storagebox_disk_usage_snapshots_bytes",
			"Diskspace used by snapshots in bytes",
			[]string{"id", "name", "server", "location"},
			nil,
		),

		// Info and status metrics
		info: prometheus.NewDesc(
			"storagebox_info",
			"Storage box information",
			[]string{"id", "name", "username", "server", "location", "storage_type", "system"},
			nil,
		),
		status: prometheus.NewDesc(
			"storagebox_status",
			"Storage box status (always 1, status in label: active, initializing, locked)",
			[]string{"id", "name", "status"},
			nil,
		),
		accessSSH: prometheus.NewDesc(
			"storagebox_access_ssh_enabled",
			"SSH access enabled (1=enabled, 0=disabled)",
			[]string{"id", "name"},
			nil,
		),
		accessSamba: prometheus.NewDesc(
			"storagebox_access_samba_enabled",
			"Samba/CIFS access enabled (1=enabled, 0=disabled)",
			[]string{"id", "name"},
			nil,
		),
		accessWebDAV: prometheus.NewDesc(
			"storagebox_access_webdav_enabled",
			"WebDAV access enabled (1=enabled, 0=disabled)",
			[]string{"id", "name"},
			nil,
		),
		accessZFS: prometheus.NewDesc(
			"storagebox_access_zfs_enabled",
			"ZFS access enabled (1=enabled, 0=disabled)",
			[]string{"id", "name"},
			nil,
		),
		reachableExternal: prometheus.NewDesc(
			"storagebox_reachable_externally",
			"Storage box reachable from external networks (1=reachable, 0=not reachable)",
			[]string{"id", "name"},
			nil,
		),
		snapshotPlan: prometheus.NewDesc(
			"storagebox_snapshot_plan_enabled",
			"Automatic snapshot plan configured (1=enabled, 0=disabled)",
			[]string{"id", "name"},
			nil,
		),
		protectionDelete: prometheus.NewDesc(
			"storagebox_protection_delete",
			"Delete protection status (1=protected, 0=unprotected)",
			[]string{"id", "name"},
			nil,
		),
		createdTimestamp: prometheus.NewDesc(
			"storagebox_created_timestamp",
			"Unix timestamp of storage box creation",
			[]string{"id", "name"},
			nil,
		),

		// Exporter metrics
		up: prometheus.NewDesc(
			"storagebox_exporter_up",
			"Whether the last scrape of the Hetzner API succeeded (1=healthy, 0=unhealthy)",
			nil,
			nil,
		),
		buildInfo: prometheus.NewDesc(
			"storagebox_exporter_build_info",
			"Build information of the exporter (value always 1)",
			[]string{"version", "revision", "goversion", "build_date"},
			nil,
		),
		scrapeDuration: prometheus.NewDesc(
			"storagebox_exporter_scrape_duration_seconds",
			"Duration of the scrape in seconds",
			nil,
			nil,
		),
		scrapeErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "storagebox_exporter_scrape_errors_total",
			Help: "Total number of scrape errors",
		}),
		cacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "storagebox_exporter_cache_hits_total",
			Help: "Total number of cache hits",
		}),
		cacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "storagebox_exporter_cache_misses_total",
			Help: "Total number of cache misses",
		}),

		// Error type counters
		authErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "storagebox_exporter_auth_errors_total",
			Help: "Total number of authentication/authorization errors (401, 403)",
		}),
		rateLimitErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "storagebox_exporter_rate_limit_errors_total",
			Help: "Total number of rate limit errors (429)",
		}),
		serverErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "storagebox_exporter_server_errors_total",
			Help: "Total number of server errors (5xx)",
		}),
		clientErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "storagebox_exporter_client_errors_total",
			Help: "Total number of client errors (400, 404)",
		}),
		networkErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "storagebox_exporter_network_errors_total",
			Help: "Total number of network/connection errors",
		}),
	}
}

// Describe implements prometheus.Collector
func (c *StorageBoxCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.diskQuota
	ch <- c.diskUsage
	ch <- c.diskUsageData
	ch <- c.diskUsageSnapshots
	ch <- c.info
	ch <- c.status
	ch <- c.accessSSH
	ch <- c.accessSamba
	ch <- c.accessWebDAV
	ch <- c.accessZFS
	ch <- c.reachableExternal
	ch <- c.snapshotPlan
	ch <- c.protectionDelete
	ch <- c.createdTimestamp
	ch <- c.up
	ch <- c.buildInfo
	ch <- c.scrapeDuration
	c.scrapeErrors.Describe(ch)
	c.cacheHits.Describe(ch)
	c.cacheMisses.Describe(ch)
	c.authErrors.Describe(ch)
	c.rateLimitErrors.Describe(ch)
	c.serverErrors.Describe(ch)
	c.clientErrors.Describe(ch)
	c.networkErrors.Describe(ch)
}

// Collect implements prometheus.Collector
func (c *StorageBoxCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	// build_info is static and always emitted, regardless of scrape outcome.
	ch <- prometheus.MustNewConstMetric(
		c.buildInfo,
		prometheus.GaugeValue,
		1,
		c.buildInfoData.Version, c.buildInfoData.Commit, runtime.Version(), c.buildInfoData.BuildDate,
	)

	boxes, err := c.fetchBoxes()
	if err != nil {
		// Source unreachable/unparseable: report up=0 and omit storage box
		// metrics (no misleading zeros or stale values), per the exporter blueprint.
		c.emitExporterMetrics(ch, 0, time.Since(start).Seconds())
		return
	}

	for _, box := range boxes {
		c.collectStorageBox(ch, &box)
	}

	c.emitExporterMetrics(ch, 1, time.Since(start).Seconds())
}

// fetchBoxes returns the storage boxes, using the cache when enabled. On error
// it records the appropriate error counters via handleError.
func (c *StorageBoxCollector) fetchBoxes() ([]hetzner.StorageBox, error) {
	if c.cacheEnabled {
		if cachedData, found := c.cache.Get(); found {
			c.cacheHits.Inc()
			return cachedData.([]hetzner.StorageBox), nil
		}
		c.cacheMisses.Inc()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		boxes, err := c.client.ListStorageBoxes(ctx)
		if err != nil {
			c.handleError(err, "cache_miss")
			return nil, err
		}
		c.cache.Set(boxes)
		return boxes, nil
	}

	// Cache disabled - always fetch from API
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	boxes, err := c.client.ListStorageBoxes(ctx)
	if err != nil {
		c.handleError(err, "direct_api_call")
		return nil, err
	}
	return boxes, nil
}

// emitExporterMetrics emits the exporter-level metrics (up, scrape duration and
// all counters) shared by both the success and failure paths.
func (c *StorageBoxCollector) emitExporterMetrics(ch chan<- prometheus.Metric, up, duration float64) {
	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, up)
	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, duration)

	c.scrapeErrors.Collect(ch)
	c.cacheHits.Collect(ch)
	c.cacheMisses.Collect(ch)
	c.authErrors.Collect(ch)
	c.rateLimitErrors.Collect(ch)
	c.serverErrors.Collect(ch)
	c.clientErrors.Collect(ch)
	c.networkErrors.Collect(ch)
}

// collectStorageBox collects metrics for a single storage box
func (c *StorageBoxCollector) collectStorageBox(ch chan<- prometheus.Metric, box *hetzner.StorageBox) {
	id := formatInt64(box.ID)
	name := box.Name
	server := box.Server
	location := box.Location.Name

	// Core storage metrics
	// Quota from storage box type
	ch <- prometheus.MustNewConstMetric(
		c.diskQuota,
		prometheus.GaugeValue,
		float64(box.StorageBoxType.Size),
		id, name, server, location,
	)

	ch <- prometheus.MustNewConstMetric(
		c.diskUsage,
		prometheus.GaugeValue,
		float64(box.Stats.Size),
		id, name, server, location,
	)

	ch <- prometheus.MustNewConstMetric(
		c.diskUsageData,
		prometheus.GaugeValue,
		float64(box.Stats.SizeData),
		id, name, server, location,
	)

	ch <- prometheus.MustNewConstMetric(
		c.diskUsageSnapshots,
		prometheus.GaugeValue,
		float64(box.Stats.SizeSnapshots),
		id, name, server, location,
	)

	// Info metric
	ch <- prometheus.MustNewConstMetric(
		c.info,
		prometheus.GaugeValue,
		1,
		id, name, box.Username, server, location, box.StorageBoxType.Name, box.System,
	)

	// Status metric (always 1, status value in label)
	ch <- prometheus.MustNewConstMetric(
		c.status,
		prometheus.GaugeValue,
		1,
		id, name, box.Status,
	)

	// Access settings metrics
	ch <- prometheus.MustNewConstMetric(
		c.accessSSH,
		prometheus.GaugeValue,
		boolToFloat64(box.AccessSettings.SSH),
		id, name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessSamba,
		prometheus.GaugeValue,
		boolToFloat64(box.AccessSettings.Samba),
		id, name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessWebDAV,
		prometheus.GaugeValue,
		boolToFloat64(box.AccessSettings.WebDAV),
		id, name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.accessZFS,
		prometheus.GaugeValue,
		boolToFloat64(box.AccessSettings.ZFS),
		id, name,
	)

	ch <- prometheus.MustNewConstMetric(
		c.reachableExternal,
		prometheus.GaugeValue,
		boolToFloat64(box.AccessSettings.ReachableExternally),
		id, name,
	)

	// Snapshot plan metric
	snapshotEnabled := float64(0)
	if box.SnapshotPlan != nil && box.SnapshotPlan.Enabled {
		snapshotEnabled = 1
	}
	ch <- prometheus.MustNewConstMetric(
		c.snapshotPlan,
		prometheus.GaugeValue,
		snapshotEnabled,
		id, name,
	)

	// Protection metric
	ch <- prometheus.MustNewConstMetric(
		c.protectionDelete,
		prometheus.GaugeValue,
		boolToFloat64(box.Protection.Delete),
		id, name,
	)

	// Created timestamp metric
	ch <- prometheus.MustNewConstMetric(
		c.createdTimestamp,
		prometheus.GaugeValue,
		float64(box.Created.Unix()),
		id, name,
	)
}

// handleError processes an error and increments the appropriate error counter
func (c *StorageBoxCollector) handleError(err error, source string) {
	if hetzner.IsAPIError(err) {
		apiErr := hetzner.GetAPIError(err)

		// Increment specific error type counters
		if hetzner.IsAuthError(err) {
			c.authErrors.Inc()
		} else if apiErr.StatusCode == http.StatusTooManyRequests {
			c.rateLimitErrors.Inc()
		} else if hetzner.IsServerError(err) {
			c.serverErrors.Inc()
		} else if hetzner.IsClientError(err) {
			c.clientErrors.Inc()
		}

		// Log with structured information
		slog.Error("Hetzner API error occurred",
			"error", err,
			"error_type", http.StatusText(apiErr.StatusCode),
			"status_code", apiErr.StatusCode,
			"request_id", apiErr.RequestID,
			"source", source,
			"is_retryable", hetzner.IsRetryableError(err),
			"is_auth_error", hetzner.IsAuthError(err),
		)
	} else {
		// Non-API errors (network, timeouts, etc.)
		c.networkErrors.Inc()
		slog.Error("Network or system error occurred",
			"error", err,
			"error_type", "network",
			"source", source,
		)
	}

	// Always increment total errors counter
	c.scrapeErrors.Inc()
}

// Helper functions

func formatInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
