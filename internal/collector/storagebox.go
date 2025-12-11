package collector

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/crstian19/prometheus-storagebox-exporter/internal/cache"
	"github.com/crstian19/prometheus-storagebox-exporter/internal/hetzner"
	"github.com/prometheus/client_golang/prometheus"
)

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
	scrapeDuration *prometheus.Desc
	scrapeErrors   prometheus.Counter
	cacheHits      prometheus.Counter
	cacheMisses    prometheus.Counter
}

// NewStorageBoxCollector creates a new StorageBoxCollector
func NewStorageBoxCollector(client *hetzner.Client, cacheTTL time.Duration, cacheMaxSize int64, cacheCleanupInterval time.Duration) *StorageBoxCollector {
	cacheEnabled := cacheTTL > 0
	return &StorageBoxCollector{
		client:       client,
		cache:        cache.NewMetricsCache(cacheTTL, cacheMaxSize, cacheCleanupInterval),
		cacheEnabled: cacheEnabled,

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
			"Current status of storage box (1=active, 0=inactive)",
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
	ch <- c.snapshotPlan
	ch <- c.protectionDelete
	ch <- c.createdTimestamp
	ch <- c.scrapeDuration
	c.scrapeErrors.Describe(ch)
	c.cacheHits.Describe(ch)
	c.cacheMisses.Describe(ch)
}

// Collect implements prometheus.Collector
func (c *StorageBoxCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	var boxes []hetzner.StorageBox

	// Try to get data from cache first (only if cache is enabled)
	if c.cacheEnabled {
		if cachedData, found := c.cache.Get(); found {
			c.cacheHits.Inc()
			boxes = cachedData.([]hetzner.StorageBox)
		} else {
			// Cache miss - fetch from API
			c.cacheMisses.Inc()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			fetchedBoxes, err := c.client.ListStorageBoxes(ctx)
			if err != nil {
				log.Printf("Error fetching storage boxes: %v", err)
				c.scrapeErrors.Inc()
				c.scrapeErrors.Collect(ch)
				c.cacheHits.Collect(ch)
				c.cacheMisses.Collect(ch)
				return
			}

			boxes = fetchedBoxes
			// Store in cache
			c.cache.Set(boxes)
		}
	} else {
		// Cache disabled - always fetch from API
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		fetchedBoxes, err := c.client.ListStorageBoxes(ctx)
		if err != nil {
			log.Printf("Error fetching storage boxes: %v", err)
			c.scrapeErrors.Inc()
			c.scrapeErrors.Collect(ch)
			c.cacheHits.Collect(ch)
			c.cacheMisses.Collect(ch)
			return
		}

		boxes = fetchedBoxes
	}

	for _, box := range boxes {
		c.collectStorageBox(ch, &box)
	}

	// Record scrape duration
	duration := time.Since(start).Seconds()
	ch <- prometheus.MustNewConstMetric(
		c.scrapeDuration,
		prometheus.GaugeValue,
		duration,
	)

	c.scrapeErrors.Collect(ch)
	c.cacheHits.Collect(ch)
	c.cacheMisses.Collect(ch)
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

	// Status metric
	statusValue := float64(0)
	if box.Status == "active" {
		statusValue = 1
	}
	ch <- prometheus.MustNewConstMetric(
		c.status,
		prometheus.GaugeValue,
		statusValue,
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
