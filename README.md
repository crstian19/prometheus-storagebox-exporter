# Prometheus Hetzner Storage Box Exporter

[![CI](https://github.com/crstian/prometheus-storagebox-exporter/workflows/CI/badge.svg)](https://github.com/crstian/prometheus-storagebox-exporter/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/crstian/prometheus-storagebox-exporter)](https://goreportcard.com/report/github.com/crstian/prometheus-storagebox-exporter)
[![License](https://img.shields.io/github/license/crstian/prometheus-storagebox-exporter)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/crstian/prometheus-storagebox-exporter.svg)](https://github.com/crstian/prometheus-storagebox-exporter/releases/)

A Prometheus exporter for [Hetzner Storage Box](https://www.hetzner.com/storage/storage-box) metrics written in Go. This exporter uses the modern Hetzner API (`api.hetzner.com`) to collect and expose storage box statistics in Prometheus format.

## Features

- Modern Hetzner API integration (not the deprecated Robot API)
- Comprehensive metrics collection (12+ metrics)
- Docker support with multi-architecture images (amd64, arm64)
- Kubernetes-ready deployment
- Minimal resource footprint (<50MB memory)
- Graceful shutdown support
- Health check endpoint
- Built-in landing page with metrics overview

## Metrics

The exporter exposes the following metrics:

### Core Storage Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `storagebox_disk_usage_bytes` | Gauge | Total used diskspace in bytes | id, name, server, location |
| `storagebox_disk_usage_data_bytes` | Gauge | Diskspace used by files in bytes | id, name, server, location |
| `storagebox_disk_usage_snapshots_bytes` | Gauge | Diskspace used by snapshots in bytes | id, name, server, location |

### Information & Status Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `storagebox_info` | Info | Storage box information (value always 1) | id, name, username, server, location, storage_type, system |
| `storagebox_status` | Gauge | Current status (1=active, 0=inactive) | id, name, status |
| `storagebox_access_ssh_enabled` | Gauge | SSH access enabled (1=yes, 0=no) | id, name |
| `storagebox_access_samba_enabled` | Gauge | Samba/CIFS access enabled (1=yes, 0=no) | id, name |
| `storagebox_access_webdav_enabled` | Gauge | WebDAV access enabled (1=yes, 0=no) | id, name |
| `storagebox_snapshot_plan_enabled` | Gauge | Automatic snapshots configured (1=yes, 0=no) | id, name |
| `storagebox_protection_delete` | Gauge | Delete protection status (1=protected, 0=no) | id, name |
| `storagebox_created_timestamp` | Gauge | Unix timestamp of creation | id, name |

### Exporter Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `storagebox_exporter_scrape_duration_seconds` | Gauge | Duration of the scrape in seconds |
| `storagebox_exporter_scrape_errors_total` | Counter | Total number of scrape errors |

## Installation

### Prerequisites

You need a Hetzner API token with read permissions:

1. Log in to [Hetzner Cloud Console](https://console.hetzner.cloud/)
2. Navigate to **Security** → **API Tokens**
3. Create a new token with **Read** permissions
4. Copy the token (you'll need it for configuration)

### Binary

Download the latest release for your platform from the [releases page](https://github.com/crstian/prometheus-storagebox-exporter/releases):

```bash
# Linux amd64
wget https://github.com/crstian/prometheus-storagebox-exporter/releases/latest/download/prometheus-storagebox-exporter_linux_x86_64.tar.gz
tar xzf prometheus-storagebox-exporter_linux_x86_64.tar.gz
chmod +x prometheus-storagebox-exporter

# Run the exporter
export HETZNER_TOKEN="your-api-token"
./prometheus-storagebox-exporter
```

### Docker

```bash
docker run -d \
  --name storagebox-exporter \
  -p 9509:9509 \
  -e HETZNER_TOKEN="your-api-token" \
  ghcr.io/crstian/prometheus-storagebox-exporter:latest
```

### Docker Compose

Create a `.env` file (see [.env.example](.env.example)):

```bash
cp .env.example .env
# Edit .env and add your HETZNER_TOKEN
```

Then run:

```bash
docker-compose up -d
```

### From Source

```bash
# Clone the repository
git clone https://github.com/crstian/prometheus-storagebox-exporter.git
cd prometheus-storagebox-exporter

# Build
go build -o prometheus-storagebox-exporter .

# Run
export HETZNER_TOKEN="your-api-token"
./prometheus-storagebox-exporter
```

## Configuration

Configuration can be done via environment variables or command-line flags.

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HETZNER_TOKEN` | *required* | Hetzner API token |
| `LISTEN_ADDRESS` | `:9509` | Address to listen on |
| `METRICS_PATH` | `/metrics` | Path for metrics endpoint |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

### Command-line Flags

```bash
./prometheus-storagebox-exporter --help

Flags:
  --hetzner-token string     Hetzner API token (can also be set via HETZNER_TOKEN env var)
  --listen-address string    Address to listen on for HTTP requests (default ":9509")
  --metrics-path string      Path under which to expose metrics (default "/metrics")
  --log-level string         Log level (debug, info, warn, error) (default "info")
  --version                  Show version information and exit
```

## Prometheus Configuration

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'hetzner-storagebox'
    static_configs:
      - targets: ['localhost:9509']
    scrape_interval: 60s
    scrape_timeout: 30s
```

### Kubernetes Deployment

Example Kubernetes manifests:

<details>
<summary>Click to expand Kubernetes YAML</summary>

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: storagebox-exporter-secret
  namespace: monitoring
type: Opaque
stringData:
  hetzner-token: "your-api-token-here"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: storagebox-exporter
  namespace: monitoring
  labels:
    app: storagebox-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: storagebox-exporter
  template:
    metadata:
      labels:
        app: storagebox-exporter
    spec:
      containers:
      - name: storagebox-exporter
        image: ghcr.io/crstian/prometheus-storagebox-exporter:latest
        ports:
        - containerPort: 9509
          name: metrics
        env:
        - name: HETZNER_TOKEN
          valueFrom:
            secretKeyRef:
              name: storagebox-exporter-secret
              key: hetzner-token
        livenessProbe:
          httpGet:
            path: /health
            port: metrics
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: metrics
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "32Mi"
            cpu: "50m"
          limits:
            memory: "64Mi"
            cpu: "100m"
---
apiVersion: v1
kind: Service
metadata:
  name: storagebox-exporter
  namespace: monitoring
  labels:
    app: storagebox-exporter
spec:
  type: ClusterIP
  ports:
  - port: 9509
    targetPort: metrics
    name: metrics
  selector:
    app: storagebox-exporter
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: storagebox-exporter
  namespace: monitoring
  labels:
    app: storagebox-exporter
spec:
  selector:
    matchLabels:
      app: storagebox-exporter
  endpoints:
  - port: metrics
    interval: 60s
    scrapeTimeout: 30s
```

</details>

## Example Metrics Output

```prometheus
# HELP storagebox_disk_usage_bytes Total used diskspace in bytes
# TYPE storagebox_disk_usage_bytes gauge
storagebox_disk_usage_bytes{id="123456",name="backup-server",server="u123456.your-storagebox.de",location="fsn1"} 1.073741824e+11

# HELP storagebox_disk_usage_data_bytes Diskspace used by files in bytes
# TYPE storagebox_disk_usage_data_bytes gauge
storagebox_disk_usage_data_bytes{id="123456",name="backup-server",server="u123456.your-storagebox.de",location="fsn1"} 9.663676416e+10

# HELP storagebox_disk_usage_snapshots_bytes Diskspace used by snapshots in bytes
# TYPE storagebox_disk_usage_snapshots_bytes gauge
storagebox_disk_usage_snapshots_bytes{id="123456",name="backup-server",server="u123456.your-storagebox.de",location="fsn1"} 1.073741824e+10

# HELP storagebox_info Storage box information
# TYPE storagebox_info gauge
storagebox_info{id="123456",name="backup-server",username="u123456",server="u123456.your-storagebox.de",location="fsn1",storage_type="BX10",system="linux"} 1

# HELP storagebox_status Current status of storage box (1=active, 0=inactive)
# TYPE storagebox_status gauge
storagebox_status{id="123456",name="backup-server",status="active"} 1

# HELP storagebox_access_ssh_enabled SSH access enabled (1=enabled, 0=disabled)
# TYPE storagebox_access_ssh_enabled gauge
storagebox_access_ssh_enabled{id="123456",name="backup-server"} 1
```

## Grafana Dashboard

A pre-built Grafana dashboard is included in the repository: [grafana-dashboard.json](grafana-dashboard.json)

### Features

The dashboard includes:
- **Overview**: Total storage boxes count, status, scrape metrics
- **Disk Usage Graph**: Time series showing total, data, and snapshot usage
- **Usage Breakdown**: Individual panels for data and snapshot usage
- **Access Settings**: Visual status of SSH, Samba, and WebDAV access
- **Configuration**: Snapshot plan and delete protection status
- **Multi-box View**: Bar chart comparing all storage boxes
- **Variables**: Dropdown to select specific storage box

### Import Instructions

1. Open Grafana and navigate to **Dashboards** → **Import**
2. Click **Upload JSON file**
3. Select `grafana-dashboard.json` from this repository
4. Select your Prometheus data source
5. Click **Import**

Alternatively, copy the contents of `grafana-dashboard.json` and paste it into the import dialog.

### Screenshot

The dashboard displays:
- Storage usage trends over time
- Current usage statistics with thresholds
- Access configuration at a glance
- Exporter health metrics

### Customization

You can customize the dashboard by:
- Adjusting time ranges and refresh intervals
- Modifying thresholds for alerts
- Adding additional panels for specific metrics
- Creating alert rules based on usage thresholds

## Development

### Building

```bash
# Build binary
go build -o prometheus-storagebox-exporter .

# Build Docker image
docker build -t prometheus-storagebox-exporter .

# Run tests
go test -v ./...

# Run linter
golangci-lint run
```

### Project Structure

```
.
├── main.go                 # Application entry point
├── internal/
│   ├── collector/          # Prometheus collector implementation
│   ├── hetzner/           # Hetzner API client
│   └── config/            # Configuration handling
├── .github/workflows/     # CI/CD pipelines
├── Dockerfile             # Multi-stage Docker build
├── docker-compose.yml     # Docker Compose configuration
└── DESIGN.md             # Architecture documentation
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes (API changes, removed metrics)
- **MINOR**: New features (new metrics, new flags)
- **PATCH**: Bug fixes, documentation updates

### Creating a Release

```bash
# Tag a new version
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions will automatically:
# - Build binaries for all platforms
# - Create Docker images (amd64, arm64)
# - Publish to GitHub Container Registry
# - Create a GitHub release with artifacts
```

## Comparison with Reference Implementation

This exporter improves upon [fleaz/prometheus-storagebox-exporter](https://github.com/fleaz/prometheus-storagebox-exporter):

| Feature | This Exporter | Reference |
|---------|---------------|-----------|
| API | Modern Hetzner API (api.hetzner.com) | Deprecated Robot API |
| Authentication | Bearer token | Basic auth (user/pass) |
| Metrics Count | 12+ metrics | 4 metrics |
| Info Metrics | Yes (with rich labels) | No |
| Access Settings | Yes (SSH, Samba, WebDAV) | No |
| Protection Status | Yes | No |
| Snapshot Plan Info | Yes | No |
| Multi-arch Docker | Yes (amd64, arm64) | amd64 only |
| CI/CD | GitHub Actions + goreleaser | Not included |
| Future-proof | Yes (API sunset: never) | No (API sunset: July 2025) |

## Troubleshooting

### Error: "HETZNER_TOKEN environment variable or --hetzner-token flag is required"

Make sure you've set the `HETZNER_TOKEN` environment variable or passed it via the `--hetzner-token` flag.

### Error: "API request failed with status 401"

Your API token is invalid or has expired. Generate a new token from the Hetzner Cloud Console.

### Error: "API request failed with status 403"

Your API token doesn't have sufficient permissions. Make sure the token has at least **Read** permissions.

### No metrics appearing in Prometheus

1. Check that the exporter is running: `curl http://localhost:9509/health`
2. Check the metrics endpoint: `curl http://localhost:9509/metrics`
3. Verify your Prometheus scrape configuration
4. Check the exporter logs for errors

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Hetzner](https://www.hetzner.com/) for providing the Storage Box service and API
- [Prometheus](https://prometheus.io/) for the monitoring toolkit
- [fleaz/prometheus-storagebox-exporter](https://github.com/fleaz/prometheus-storagebox-exporter) for inspiration

## Support

- GitHub Issues: [Report a bug](https://github.com/crstian/prometheus-storagebox-exporter/issues)
- Documentation: See [DESIGN.md](DESIGN.md) for architecture details