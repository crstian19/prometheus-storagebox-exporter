package hetzner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.hetzner.com/v1"
	defaultTimeout = 30 * time.Second
)

// Client is a Hetzner API client for Storage Boxes
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

// NewClient creates a new Hetzner API client
func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		token:   token,
		baseURL: defaultBaseURL,
	}
}

// StorageBox represents a Hetzner Storage Box
type StorageBox struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name"`
	Username       string         `json:"username"`
	Status         string         `json:"status"`
	Server         string         `json:"server"`
	System         string         `json:"system"`
	StorageBoxType StorageBoxType `json:"storage_box_type"`
	Location       Location       `json:"location"`
	Stats          Stats          `json:"stats"`
	AccessSettings AccessSettings `json:"access_settings"`
	SnapshotPlan   *SnapshotPlan  `json:"snapshot_plan"`
	Protection     Protection     `json:"protection"`
	Labels         map[string]string `json:"labels"`
	Created        time.Time      `json:"created"`
}

// Location represents the data center location
type Location struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Country     string `json:"country"`
	City        string `json:"city"`
}

// StorageBoxType represents the type of storage box
type StorageBoxType struct {
	Name string `json:"name"`
	Size int64  `json:"size"` // Total quota/capacity in bytes
}

// Stats represents storage usage statistics
type Stats struct {
	Size          int64 `json:"size"`           // Total usage in bytes
	SizeData      int64 `json:"size_data"`      // Data usage in bytes
	SizeSnapshots int64 `json:"size_snapshots"` // Snapshot usage in bytes
}

// AccessSettings represents the access configuration
type AccessSettings struct {
	SSH               bool `json:"ssh_enabled"`        // SSH access enabled
	Samba             bool `json:"samba_enabled"`      // Samba access enabled
	WebDAV            bool `json:"webdav_enabled"`     // WebDAV access enabled
	ZFS               bool `json:"zfs_enabled"`        // ZFS access enabled
	ReachableExternally bool `json:"reachable_externally"` // Storage box reachable externally
}

// SnapshotPlan represents the automatic snapshot configuration
type SnapshotPlan struct {
	Enabled bool `json:"enabled"`
}

// Protection represents the protection settings
type Protection struct {
	Delete bool `json:"delete"`
}

// storageBoxesResponse represents the API response for listing storage boxes
type storageBoxesResponse struct {
	StorageBoxes []StorageBox `json:"storage_boxes"`
}

// ListStorageBoxes retrieves all storage boxes from the Hetzner API
func (c *Client) ListStorageBoxes(ctx context.Context) ([]StorageBox, error) {
	url := fmt.Sprintf("%s/storage_boxes", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API request failed with status %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result storageBoxesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.StorageBoxes, nil
}
