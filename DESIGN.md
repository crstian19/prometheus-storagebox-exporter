# Prometheus Hetzner Storage Box Exporter - Design Document

## Overview

A Prometheus exporter for Hetzner Storage Box metrics written in Go. This exporter uses the modern Hetzner API (api.hetzner.com) to collect and expose storage box statistics.

## API Information

### Hetzner Storage Box API

- **Base URL:** `https://api.hetzner.com/v1`
- **Endpoint:** `/storage_boxes`
- **Authentication:** Bearer token (API token from Hetzner Cloud Console)
- **Format:** JSON over HTTPS (RESTful)

### Key Differences from Reference Implementation

The reference implementation (fleaz/prometheus-storagebox-exporter) uses the **deprecated Robot API** (robot-ws.your-server.de) which will be removed on **July 30, 2025**. Our implementation uses the modern API with:
- Better authentication (Bearer tokens instead of Basic Auth)
- Richer data structures
- More metadata and configuration details
- Future-proof API endpoint

## Planned Metrics

### Core Storage Metrics

These match the reference implementation but use the new API:

| Metric Name | Type | Description | Unit |
|------------|------|-------------|------|
| `storagebox_disk_quota_bytes` | Gauge | Total allocated diskspace | bytes |
| `storagebox_disk_usage_bytes` | Gauge | Total used diskspace | bytes |
| `storagebox_disk_usage_data_bytes` | Gauge | Diskspace used by files | bytes |
| `storagebox_disk_usage_snapshots_bytes` | Gauge | Diskspace used by snapshots | bytes |

**Labels:** `id`, `name`, `server`, `location`

### Additional Enhanced Metrics

New metrics leveraging the richer API data:

| Metric Name | Type | Description | Values |
|------------|------|-------------|--------|
| `storagebox_info` | Info | Storage box information | Labels: id, name, username, server, location, storage_type, system |
| `storagebox_status` | Gauge | Current status of storage box | 1=active, 0=inactive/locked |
| `storagebox_access_ssh_enabled` | Gauge | SSH access enabled | 1=enabled, 0=disabled |
| `storagebox_access_samba_enabled` | Gauge | Samba/CIFS access enabled | 1=enabled, 0=disabled |
| `storagebox_access_webdav_enabled` | Gauge | WebDAV access enabled | 1=enabled, 0=disabled |
| `storagebox_snapshot_plan_enabled` | Gauge | Automatic snapshot plan configured | 1=enabled, 0=disabled |
| `storagebox_protection_delete` | Gauge | Delete protection status | 1=protected, 0=unprotected |
| `storagebox_created_timestamp` | Gauge | Unix timestamp of creation | seconds |

**Labels:** Varies by metric (all include at minimum: `id`, `name`)

## Architecture

### Component Structure

```
prometheus-storagebox-exporter/
├── main.go                 # Entry point, HTTP server, CLI flags
├── internal/
│   ├── collector/
│   │   └── storagebox.go  # Prometheus collector implementation
│   ├── hetzner/
│   │   └── client.go      # Hetzner API client
│   └── config/
│       └── config.go      # Configuration handling
├── go.mod                  # Go module definition
├── go.sum                  # Go dependencies
├── Dockerfile              # Multi-stage Docker build
├── docker-compose.yml      # Docker Compose example
├── .github/
│   └── workflows/
│       ├── ci.yml         # CI pipeline (test, build, lint)
│       └── release.yml    # Release automation with goreleaser
├── .goreleaser.yml        # Goreleaser configuration
├── README.md              # User documentation
└── DESIGN.md              # This document
```

### Data Flow

```
┌─────────────────┐
│  Prometheus     │
│  Scrapes        │
│  /metrics       │
└────────┬────────┘
         │ HTTP GET
         ▼
┌─────────────────────────────┐
│  Exporter HTTP Server       │
│  (Port 9509)                │
└────────┬────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│  Prometheus Collector       │
│  - Collect() implementation │
│  - Metric definitions       │
└────────┬────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│  Hetzner API Client         │
│  - ListStorageBoxes()       │
│  - HTTP client with retry   │
│  - Bearer token auth        │
└────────┬────────────────────┘
         │ HTTPS GET
         ▼
┌─────────────────────────────┐
│  Hetzner API                │
│  api.hetzner.com/v1         │
│  /storage_boxes             │
└─────────────────────────────┘
```

### Core Components

#### 1. Hetzner API Client (`internal/hetzner/client.go`)

**Responsibilities:**
- HTTP client configuration with timeouts
- Bearer token authentication
- API request/response handling
- Error handling and retries
- Rate limiting (if needed)
- Data structure definitions

**Key Types:**
```go
type Client struct {
    httpClient *http.Client
    token      string
    baseURL    string
}

type StorageBox struct {
    ID       int64
    Name     string
    Username string
    Status   string
    Server   string
    Location Location
    Stats    Stats
    AccessSettings AccessSettings
    // ... other fields
}

type Stats struct {
    Size          int64  // Total usage in bytes
    SizeData      int64  // Data usage in bytes
    SizeSnapshots int64  // Snapshot usage in bytes
}
```

**Main Methods:**
- `NewClient(token string) *Client`
- `ListStorageBoxes(ctx context.Context) ([]StorageBox, error)`

#### 2. Prometheus Collector (`internal/collector/storagebox.go`)

**Responsibilities:**
- Implement `prometheus.Collector` interface
- Define all metrics with appropriate types
- Fetch data from Hetzner API client
- Update metrics with current values
- Handle collection errors gracefully

**Implementation:**
- `Describe()` - Register metric descriptors
- `Collect()` - Fetch data and update metrics

#### 3. Configuration (`internal/config/config.go`)

**Environment Variables:**
- `HETZNER_TOKEN` - API token (required)
- `LISTEN_ADDRESS` - HTTP server address (default: `:9509`)
- `METRICS_PATH` - Metrics endpoint path (default: `/metrics`)
- `LOG_LEVEL` - Logging level (default: `info`)

**Command-line Flags:**
- `--listen-address` - Override listen address
- `--metrics-path` - Override metrics path
- `--log-level` - Override log level
- `--version` - Show version and exit

#### 4. Main Application (`main.go`)

**Responsibilities:**
- Parse configuration
- Initialize Hetzner API client
- Register Prometheus collector
- Start HTTP server
- Graceful shutdown handling
- Health check endpoint

**Endpoints:**
- `GET /metrics` - Prometheus metrics
- `GET /health` - Health check (returns 200 OK)
- `GET /` - Landing page with info

## Configuration

### Authentication

Users must generate an API token from the Hetzner Cloud Console:
1. Log in to Hetzner Cloud Console
2. Navigate to Security > API Tokens
3. Create a new token with **read** permissions
4. Set the token as `HETZNER_TOKEN` environment variable

### Running the Exporter

**Binary:**
```bash
export HETZNER_TOKEN="your-api-token"
./prometheus-storagebox-exporter
```

**Docker:**
```bash
docker run -p 9509:9509 -e HETZNER_TOKEN="your-token" prometheus-storagebox-exporter
```

**Docker Compose:**
```yaml
version: '3.8'
services:
  storagebox-exporter:
    image: prometheus-storagebox-exporter:latest
    environment:
      HETZNER_TOKEN: ${HETZNER_TOKEN}
    ports:
      - "9509:9509"
```

### Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'hetzner-storagebox'
    static_configs:
      - targets: ['localhost:9509']
    scrape_interval: 60s
```

## CI/CD Pipeline

### GitHub Actions Workflows

#### CI Pipeline (`.github/workflows/ci.yml`)
Triggers: Push, Pull Request

**Steps:**
1. Checkout code
2. Setup Go (latest stable)
3. Run `go mod download`
4. Run `go vet ./...`
5. Run `golangci-lint`
6. Run `go test -v -race -coverprofile=coverage.out ./...`
7. Upload coverage to codecov (optional)
8. Build binary for multiple platforms

#### Release Pipeline (`.github/workflows/release.yml`)
Triggers: Git tag (v*)

**Steps:**
1. Checkout code
2. Setup Go
3. Run goreleaser with GitHub token
4. Build and push Docker images
5. Create GitHub release with binaries

### Semantic Versioning

- Format: `vMAJOR.MINOR.PATCH` (e.g., `v1.0.0`)
- Tags trigger automated releases
- Goreleaser generates:
  - Linux binaries (amd64, arm64, armv7)
  - macOS binaries (amd64, arm64)
  - Windows binaries (amd64)
  - Docker images (multi-arch)
  - Checksums and signatures
  - GitHub release notes

### Versioning Strategy

- **MAJOR:** Breaking changes (API changes, removed metrics)
- **MINOR:** New features (new metrics, new flags)
- **PATCH:** Bug fixes, documentation updates

## Docker Strategy

### Multi-stage Build

```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder
# Build binary with CGO_ENABLED=0 for static linking

# Stage 2: Runtime
FROM scratch
# Copy binary and CA certificates
# Minimal attack surface
```

### Image Tags

- `latest` - Latest release
- `vX.Y.Z` - Specific version
- `vX.Y` - Minor version (updates with patches)
- `vX` - Major version (updates with minors)

## Dependencies

### Core Dependencies

- `github.com/prometheus/client_golang` - Prometheus client library
- `github.com/spf13/pflag` - Command-line flag parsing (POSIX/GNU-style)

### Development Dependencies

- `golangci-lint` - Linting
- `goreleaser` - Release automation
- GitHub Actions - CI/CD

## Error Handling

### API Client
- Network errors: Retry with exponential backoff
- Authentication errors: Log and exit (invalid token)
- Rate limiting: Respect rate limits, add delays if needed
- Timeout: Configurable timeout per request

### Collector
- API failures: Log error, return stale metrics (if available)
- Partial failures: Continue collecting available metrics
- Invalid data: Log warning, skip invalid entries

## Monitoring and Observability

### Internal Metrics

The exporter exposes its own metrics:
- `storagebox_exporter_scrape_duration_seconds` - Scrape duration
- `storagebox_exporter_scrape_errors_total` - Total scrape errors
- `storagebox_exporter_api_requests_total` - API request count
- `storagebox_exporter_api_request_duration_seconds` - API request duration

### Logging

Structured logging with levels:
- `debug` - Detailed API requests/responses
- `info` - Startup, config, regular operations
- `warn` - Recoverable errors, retries
- `error` - Critical failures

## Security Considerations

1. **API Token Protection:**
   - Never log the token
   - Use environment variables
   - Document secure practices

2. **Minimal Container:**
   - Use `scratch` base image
   - No shell or utilities
   - Reduces attack surface

3. **Read-only Token:**
   - Exporter only needs read permissions
   - Document creating read-only tokens

4. **Network Security:**
   - HTTPS only for API calls
   - Validate TLS certificates

## Performance Considerations

- **Caching:** Consider caching API responses (60s TTL)
- **Concurrency:** Use goroutines for parallel API calls if multiple boxes
- **Memory:** Minimal memory footprint (<50MB)
- **API Rate Limits:** Respect Hetzner API rate limits

## Future Enhancements

Potential future additions:
- Support for Storage Box subaccounts metrics
- Snapshot list and metrics
- Alert rules examples
- Grafana dashboard JSON
- Health checks for individual storage boxes
- Metrics about snapshot schedules
