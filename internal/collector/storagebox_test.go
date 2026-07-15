package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/crstian19/prometheus-storagebox-exporter/internal/hetzner"
	"github.com/prometheus/client_golang/prometheus"
)

// mockStorageBoxResponse creates a mock API response with storage boxes
func mockStorageBoxResponse() map[string]interface{} {
	return map[string]interface{}{
		"storage_boxes": []map[string]interface{}{
			{
				"id":       12345,
				"name":     "test-storagebox",
				"username": "u123456",
				"status":   "active",
				"server":   "u123456.your-storagebox.de",
				"system":   "storagebox",
				"storage_box_type": map[string]interface{}{
					"name": "BX10",
					"size": int64(1099511627776), // 1TB in bytes
				},
				"location": map[string]interface{}{
					"name":        "fsn1",
					"description": "Falkenstein DC Park 1",
					"country":     "DE",
					"city":        "Falkenstein",
				},
				"stats": map[string]interface{}{
					"size":           int64(536870912000), // 500GB
					"size_data":      int64(429496729600), // 400GB
					"size_snapshots": int64(107374182400), // 100GB
				},
				"access_settings": map[string]interface{}{
					"ssh_enabled":          true,
					"samba_enabled":        true,
					"webdav_enabled":       false,
					"zfs_enabled":          false,
					"reachable_externally": true,
				},
				"snapshot_plan": map[string]interface{}{
					"enabled": true,
				},
				"protection": map[string]interface{}{
					"delete": true,
				},
				"labels":  map[string]string{},
				"created": "2024-01-15T10:30:00Z",
			},
			{
				"id":       12346,
				"name":     "inactive-storagebox",
				"username": "u123457",
				"status":   "inactive",
				"server":   "u123457.your-storagebox.de",
				"system":   "storagebox",
				"storage_box_type": map[string]interface{}{
					"name": "BX20",
					"size": int64(2199023255552), // 2TB in bytes
				},
				"location": map[string]interface{}{
					"name":        "nbg1",
					"description": "Nuremberg DC Park 1",
					"country":     "DE",
					"city":        "Nuremberg",
				},
				"stats": map[string]interface{}{
					"size":           int64(0),
					"size_data":      int64(0),
					"size_snapshots": int64(0),
				},
				"access_settings": map[string]interface{}{
					"ssh_enabled":          false,
					"samba_enabled":        false,
					"webdav_enabled":       false,
					"zfs_enabled":          false,
					"reachable_externally": false,
				},
				"snapshot_plan": nil,
				"protection": map[string]interface{}{
					"delete": false,
				},
				"labels":  map[string]string{},
				"created": "2024-06-01T08:00:00Z",
			},
		},
	}
}

// setupMockServer creates a mock HTTP server for testing
func setupMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *hetzner.Client) {
	server := httptest.NewServer(handler)
	client := hetzner.NewClient("test-token")
	client.SetBaseURL(server.URL)
	return server, client
}

func TestNewStorageBoxCollector(t *testing.T) {
	client := hetzner.NewClient("test-token")

	tests := []struct {
		name                 string
		cacheTTL             time.Duration
		cacheMaxSize         int64
		cacheCleanupInterval time.Duration
		expectCacheEnabled   bool
	}{
		{
			name:                 "cache disabled",
			cacheTTL:             0,
			cacheMaxSize:         0,
			cacheCleanupInterval: 0,
			expectCacheEnabled:   false,
		},
		{
			name:                 "cache enabled",
			cacheTTL:             time.Minute,
			cacheMaxSize:         1024,
			cacheCleanupInterval: time.Minute,
			expectCacheEnabled:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewStorageBoxCollector(client, tt.cacheTTL, tt.cacheMaxSize, tt.cacheCleanupInterval, BuildInfo{})
			if collector == nil {
				t.Fatal("expected collector to be non-nil")
			}
			if collector.client != client {
				t.Error("expected client to be set")
			}
			if collector.cacheEnabled != tt.expectCacheEnabled {
				t.Errorf("expected cacheEnabled=%v, got %v", tt.expectCacheEnabled, collector.cacheEnabled)
			}
		})
	}
}

func TestDescribe(t *testing.T) {
	client := hetzner.NewClient("test-token")
	collector := NewStorageBoxCollector(client, 0, 0, 0, BuildInfo{})

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We should have at least the core metrics described
	expectedMinDescs := 10
	if len(descs) < expectedMinDescs {
		t.Errorf("expected at least %d descriptors, got %d", expectedMinDescs, len(descs))
	}
}

func TestCollectSuccess(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockStorageBoxResponse()); err != nil {
			t.Errorf("Failed to encode mock response: %v", err)
		}
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	collector := NewStorageBoxCollector(client, 0, 0, 0, BuildInfo{})

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// We should have metrics for 2 storage boxes + exporter metrics
	if len(metrics) < 20 {
		t.Errorf("expected at least 20 metrics, got %d", len(metrics))
	}
}

func TestCollectWithCacheHit(t *testing.T) {
	callCount := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockStorageBoxResponse()); err != nil {
			t.Errorf("Failed to encode mock response: %v", err)
		}
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	// Enable cache with 1 minute TTL
	collector := NewStorageBoxCollector(client, time.Minute, 0, time.Minute, BuildInfo{})

	// First collection - should hit API
	ch1 := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch1)
		close(ch1)
	}()
	for range ch1 {
	}

	if callCount != 1 {
		t.Errorf("expected 1 API call after first collect, got %d", callCount)
	}

	// Second collection - should use cache
	ch2 := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch2)
		close(ch2)
	}()
	for range ch2 {
	}

	if callCount != 1 {
		t.Errorf("expected still 1 API call after second collect (cache hit), got %d", callCount)
	}
}

func TestCollectAPIError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"error": {"message": "Invalid token", "code": "unauthorized"}}`,
		},
		{
			name:       "forbidden",
			statusCode: http.StatusForbidden,
			body:       `{"error": {"message": "Forbidden", "code": "forbidden"}}`,
		},
		{
			name:       "rate limited",
			statusCode: http.StatusTooManyRequests,
			body:       `{"error": {"message": "Rate limited", "code": "rate_limit_exceeded"}}`,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			body:       `{"error": {"message": "Internal server error", "code": "server_error"}}`,
		},
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			body:       `{"error": {"message": "Bad request", "code": "bad_request"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-Id", "test-request-123")
				w.WriteHeader(tt.statusCode)
				if _, err := w.Write([]byte(tt.body)); err != nil {
					t.Errorf("Failed to write response body: %v", err)
				}
			}

			server, client := setupMockServer(t, handler)
			defer server.Close()

			collector := NewStorageBoxCollector(client, 0, 0, 0, BuildInfo{})

			ch := make(chan prometheus.Metric, 100)
			go func() {
				collector.Collect(ch)
				close(ch)
			}()

			var metrics []prometheus.Metric
			for metric := range ch {
				metrics = append(metrics, metric)
			}

			// Should still get error counter metrics
			if len(metrics) == 0 {
				t.Error("expected at least error counter metrics")
			}
		})
	}
}

func TestCollectWithCacheAndAPIError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"error": {"message": "Server error"}}`)); err != nil {
			t.Errorf("Failed to write response body: %v", err)
		}
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	// Enable cache
	collector := NewStorageBoxCollector(client, time.Minute, 0, time.Minute, BuildInfo{})

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Should get error counter metrics even with cache enabled
	if len(metrics) == 0 {
		t.Error("expected at least error counter metrics")
	}
}

func TestFormatInt64(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{12345, "12345"},
		{-12345, "-12345"},
		{9223372036854775807, "9223372036854775807"},
	}

	for _, tt := range tests {
		result := formatInt64(tt.input)
		if result != tt.expected {
			t.Errorf("formatInt64(%d) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestBoolToFloat64(t *testing.T) {
	tests := []struct {
		input    bool
		expected float64
	}{
		{true, 1},
		{false, 0},
	}

	for _, tt := range tests {
		result := boolToFloat64(tt.input)
		if result != tt.expected {
			t.Errorf("boolToFloat64(%v) = %f, expected %f", tt.input, result, tt.expected)
		}
	}
}

func TestCollectStorageBoxMetrics(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockStorageBoxResponse()); err != nil {
			t.Errorf("Failed to encode mock response: %v", err)
		}
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	collector := NewStorageBoxCollector(client, 0, 0, 0, BuildInfo{})

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metricNames := make(map[string]int)
	for metric := range ch {
		desc := metric.Desc()
		metricNames[desc.String()]++
	}

	// Verify we got metrics for multiple storage boxes
	if len(metricNames) == 0 {
		t.Error("expected some metrics to be collected")
	}
}

func TestCollectEmptyStorageBoxes(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"storage_boxes": []interface{}{},
		}); err != nil {
			t.Errorf("Failed to encode mock response: %v", err)
		}
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	collector := NewStorageBoxCollector(client, 0, 0, 0, BuildInfo{})

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Should still get exporter metrics even with no storage boxes
	if len(metrics) == 0 {
		t.Error("expected at least exporter metrics")
	}
}

func TestHandleErrorAPIError(t *testing.T) {
	client := hetzner.NewClient("test-token")
	collector := NewStorageBoxCollector(client, 0, 0, 0, BuildInfo{})

	tests := []struct {
		name   string
		err    error
		source string
	}{
		{
			name:   "auth error 401",
			err:    hetzner.NewAPIError(http.StatusUnauthorized, "Unauthorized", "req-123"),
			source: "test",
		},
		{
			name:   "auth error 403",
			err:    hetzner.NewAPIError(http.StatusForbidden, "Forbidden", "req-123"),
			source: "test",
		},
		{
			name:   "rate limit error",
			err:    hetzner.NewAPIError(http.StatusTooManyRequests, "Rate limited", "req-123"),
			source: "test",
		},
		{
			name:   "server error",
			err:    hetzner.NewAPIError(http.StatusInternalServerError, "Server error", "req-123"),
			source: "test",
		},
		{
			name:   "client error",
			err:    hetzner.NewAPIError(http.StatusBadRequest, "Bad request", "req-123"),
			source: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset collector for each test
			collector = NewStorageBoxCollector(client, 0, 0, 0, BuildInfo{})
			// This should not panic
			collector.handleError(tt.err, tt.source)
		})
	}
}

func TestHandleErrorNetworkError(t *testing.T) {
	client := hetzner.NewClient("test-token")
	collector := NewStorageBoxCollector(client, 0, 0, 0, BuildInfo{})

	// Simulate a network error (non-API error)
	networkErr := &testNetworkError{message: "connection refused"}
	collector.handleError(networkErr, "test")
	// Should not panic
}

type testNetworkError struct {
	message string
}

func (e *testNetworkError) Error() string {
	return e.message
}

func TestCollectorRegistration(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockStorageBoxResponse()); err != nil {
			t.Errorf("Failed to encode mock response: %v", err)
		}
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	collector := NewStorageBoxCollector(client, 0, 0, 0, BuildInfo{})

	// Create a new registry and register the collector
	registry := prometheus.NewRegistry()
	err := registry.Register(collector)
	if err != nil {
		t.Fatalf("failed to register collector: %v", err)
	}

	// Gather metrics
	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	if len(metrics) == 0 {
		t.Error("expected at least some metrics after gathering")
	}
}

// gaugeValue gathers metrics from the registry and returns the value of the
// first sample of the named metric family, or -1 if the family is absent.
func gaugeValue(t *testing.T, reg *prometheus.Registry, name string) float64 {
	t.Helper()
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	for _, mf := range families {
		if mf.GetName() == name {
			return mf.GetMetric()[0].GetGauge().GetValue()
		}
	}
	return -1
}

func TestCollectUpAndBuildInfo(t *testing.T) {
	build := BuildInfo{Version: "1.2.3", Commit: "abc123", BuildDate: "2026-07-15"}

	t.Run("up=1 and build_info on success", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(mockStorageBoxResponse()); err != nil {
				t.Errorf("Failed to encode mock response: %v", err)
			}
		}
		server, client := setupMockServer(t, handler)
		defer server.Close()

		reg := prometheus.NewRegistry()
		if err := reg.Register(NewStorageBoxCollector(client, 0, 0, 0, build)); err != nil {
			t.Fatalf("failed to register collector: %v", err)
		}

		if got := gaugeValue(t, reg, "storagebox_exporter_up"); got != 1 {
			t.Errorf("expected storagebox_exporter_up=1, got %v", got)
		}

		// build_info must always be present with value 1.
		families, _ := reg.Gather()
		var found bool
		for _, mf := range families {
			if mf.GetName() != "storagebox_exporter_build_info" {
				continue
			}
			found = true
			m := mf.GetMetric()[0]
			if m.GetGauge().GetValue() != 1 {
				t.Errorf("expected build_info value 1, got %v", m.GetGauge().GetValue())
			}
			labels := map[string]string{}
			for _, lp := range m.GetLabel() {
				labels[lp.GetName()] = lp.GetValue()
			}
			if labels["version"] != "1.2.3" || labels["revision"] != "abc123" || labels["build_date"] != "2026-07-15" {
				t.Errorf("unexpected build_info labels: %v", labels)
			}
			if labels["goversion"] == "" {
				t.Error("expected goversion label to be set")
			}
		}
		if !found {
			t.Error("expected storagebox_exporter_build_info metric")
		}
	})

	t.Run("up=0 and no storage box metrics on API error", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}
		server, client := setupMockServer(t, handler)
		defer server.Close()

		reg := prometheus.NewRegistry()
		if err := reg.Register(NewStorageBoxCollector(client, 0, 0, 0, build)); err != nil {
			t.Fatalf("failed to register collector: %v", err)
		}

		if got := gaugeValue(t, reg, "storagebox_exporter_up"); got != 0 {
			t.Errorf("expected storagebox_exporter_up=0, got %v", got)
		}
		// Storage box metrics must be omitted on failure.
		if got := gaugeValue(t, reg, "storagebox_disk_usage_bytes"); got != -1 {
			t.Errorf("expected no storagebox_disk_usage_bytes on failure, got %v", got)
		}
		// build_info must still be present even when the scrape fails.
		if got := gaugeValue(t, reg, "storagebox_exporter_build_info"); got != 1 {
			t.Errorf("expected build_info present on failure, got %v", got)
		}
	})
}

func TestCollectWithNilSnapshotPlan(t *testing.T) {
	response := map[string]interface{}{
		"storage_boxes": []map[string]interface{}{
			{
				"id":       12345,
				"name":     "test-storagebox",
				"username": "u123456",
				"status":   "active",
				"server":   "u123456.your-storagebox.de",
				"system":   "storagebox",
				"storage_box_type": map[string]interface{}{
					"name": "BX10",
					"size": int64(1099511627776),
				},
				"location": map[string]interface{}{
					"name":        "fsn1",
					"description": "Falkenstein",
					"country":     "DE",
					"city":        "Falkenstein",
				},
				"stats": map[string]interface{}{
					"size":           int64(0),
					"size_data":      int64(0),
					"size_snapshots": int64(0),
				},
				"access_settings": map[string]interface{}{
					"ssh_enabled":          false,
					"samba_enabled":        false,
					"webdav_enabled":       false,
					"zfs_enabled":          false,
					"reachable_externally": false,
				},
				"snapshot_plan": nil, // nil snapshot plan
				"protection": map[string]interface{}{
					"delete": false,
				},
				"labels":  map[string]string{},
				"created": "2024-01-15T10:30:00Z",
			},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode mock response: %v", err)
		}
	}

	server, client := setupMockServer(t, handler)
	defer server.Close()

	collector := NewStorageBoxCollector(client, 0, 0, 0, BuildInfo{})

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Should handle nil snapshot_plan without panicking
	if len(metrics) < 10 {
		t.Errorf("expected at least 10 metrics, got %d", len(metrics))
	}
}
