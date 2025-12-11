package cache

import (
	"sync"
	"time"
)

// MetricsCache is a thread-safe cache for storing metrics data with TTL
type MetricsCache struct {
	mu              sync.RWMutex
	data            interface{}
	expiration      time.Time
	ttl             time.Duration
	maxSize         int64
	currentSize     int64
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

// NewMetricsCache creates a new cache instance with the specified configuration
func NewMetricsCache(ttl time.Duration, maxSize int64, cleanupInterval time.Duration) *MetricsCache {
	return &MetricsCache{
		ttl:             ttl,
		maxSize:         maxSize,
		cleanupInterval: cleanupInterval,
		lastCleanup:     time.Now(),
	}
}

// Get retrieves data from the cache if it exists and hasn't expired
// Returns (data, true) if cache hit, (nil, false) if cache miss or expired
func (c *MetricsCache) Get() (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if cache is empty or expired
	if c.data == nil || time.Now().After(c.expiration) {
		return nil, false
	}

	return c.data, true
}

// Set stores data in the cache with the configured TTL
func (c *MetricsCache) Set(data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = data
	c.expiration = time.Now().Add(c.ttl)
}

// IsExpired checks if the cache has expired without retrieving data
func (c *MetricsCache) IsExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.data == nil || time.Now().After(c.expiration)
}

// Clear removes all data from the cache
func (c *MetricsCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = nil
	c.expiration = time.Time{}
}

// TTL returns the configured time-to-live duration
func (c *MetricsCache) TTL() time.Duration {
	return c.ttl
}

// MaxSize returns the configured maximum cache size in bytes
func (c *MetricsCache) MaxSize() int64 {
	return c.maxSize
}

// CurrentSize returns the estimated current cache size in bytes
func (c *MetricsCache) CurrentSize() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentSize
}

// CleanupInterval returns the configured cleanup interval
func (c *MetricsCache) CleanupInterval() time.Duration {
	return c.cleanupInterval
}

// Size returns the configured maximum cache size in bytes (alias for MaxSize for backward compatibility)
func (c *MetricsCache) Size() int64 {
	return c.maxSize
}

// Cleanup performs cache cleanup if the interval has passed
// Returns true if cleanup was performed, false otherwise
func (c *MetricsCache) Cleanup() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if time.Since(c.lastCleanup) < c.cleanupInterval {
		return false
	}

	// Check if cache has expired
	if c.data != nil && now.After(c.expiration) {
		c.data = nil
		c.expiration = time.Time{}
		c.currentSize = 0
	}

	c.lastCleanup = now
	return true
}

// ShouldCleanup returns true if cleanup should be performed based on the interval
func (c *MetricsCache) ShouldCleanup() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return time.Since(c.lastCleanup) >= c.cleanupInterval
}
