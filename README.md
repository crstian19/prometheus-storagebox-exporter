# Prometheus Hetzner Storage Box Exporter

<div align="center">

<img src="https://cdn.crstian.me/storage-box-exporter-logo.png" alt="Storage Box Exporter Logo" width="200">

**Modern Prometheus exporter for Hetzner Storage Box with comprehensive metrics**

![CI](https://img.shields.io/github/actions/workflow/status/crstian19/prometheus-storagebox-exporter/ci.yml?label=CI&logo=github&logoColor=white&style=for-the-badge&branch=main)
![Go Report Card](https://goreportcard.com/badge/github.com/crstian19/prometheus-storagebox-exporter?style=for-the-badge&logo=go&logoColor=white)
![License](https://img.shields.io/github/license/crstian19/prometheus-storagebox-exporter?style=for-the-badge&logo=unlicense&logoColor=white)
![Docker](https://img.shields.io/badge/docker-ghcr.io-blue?style=for-the-badge&logo=docker&logoColor=white)
![GitHub Release](https://img.shields.io/github/release/crstian19/prometheus-storagebox-exporter?style=for-the-badge&logo=github&logoColor=white&label=release)
![Go Version](https://img.shields.io/github/go-mod/go-version/crstian19/prometheus-storagebox-exporter?style=for-the-badge&logo=go&logoColor=white&label=go)

[Quick Start](#-quick-start) • [Metrics](#-metrics) • [Installation](#-installation) • [Grafana Dashboard](#-grafana-dashboard) • [Configuration](#%EF%B8%8F-configuration)

</div>

---

## 📋 Overview

A Prometheus exporter for [Hetzner Storage Box](https://www.hetzner.com/storage/storage-box) that uses the modern Hetzner API (`api.hetzner.com`) instead of the deprecated Robot API.

- ✅ **Modern API integration** - Uses current Hetzner API (no sunset deadline)
- 📊 **Comprehensive metrics** - 15+ metrics covering usage, access, and configuration
- 🐳 **Docker ready** - Multi-architecture images (amd64, arm64)
- ☸️ **Kubernetes ready** - Includes manifests and Helm chart
- 🎯 **Minimal footprint** - Less than 50MB memory usage
- 📈 **Grafana dashboard** - Pre-built dashboard with visualizations
- 🔒 **Secure** - Bearer token authentication

### Why this exporter?

This exporter:

- Uses the modern Hetzner API that won't be deprecated
- Provides 4x more metrics including access settings and protection status
- Offers multi-architecture Docker images
- Includes a comprehensive Grafana dashboard
- Is actively maintained with CI/CD automation

### 📊 Dashboard Preview

<div align="center">

<img src="https://cdn.crstian.me/storage-exporter-dashboard-1.png" alt="Dashboard Overview" width="800">

<img src="https://cdn.crstian.me/storage-exporter-dashboard-2.png" alt="Dashboard Details" width="800">

</div>

---

## 🚀 Quick Start

### Prerequisites

You need a Hetzner API token with read permissions:

1. Log in to [Hetzner Cloud Console](https://console.hetzner.cloud/)
2. Navigate to **Security** → **API Tokens**
3. Create a new token with **Read** permissions
4. Copy the token for configuration

### Docker Compose (Recommended)


```bash
docker-compose up -d
```

### Docker

```bash
docker run -d \
  --name storagebox-exporter \
  -p 9509:9509 \
  -e HETZNER_TOKEN="your-api-token" \
  ghcr.io/crstian19/prometheus-storagebox-exporter:latest
```

### Binary

```bash
# Linux amd64
wget https://github.com/crstian19/prometheus-storagebox-exporter/releases/latest/download/prometheus-storagebox-exporter_linux_x86_64.tar.gz
tar xzf prometheus-storagebox-exporter_linux_x86_64.tar.gz

# Run
export HETZNER_TOKEN="your-api-token"
./prometheus-storagebox-exporter
```

### Access Metrics

Open http://localhost:9509/metrics to view the exported metrics.

---

---

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HETZNER_TOKEN` | *required* | Hetzner API token |
| `LISTEN_ADDRESS` | `:9509` | Address to listen on |
| `METRICS_PATH` | `/metrics` | Path for metrics endpoint |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `CACHE_TTL` | `0` | Cache TTL in seconds, 0 to disable (default: disabled) |
| `CACHE_MAX_SIZE` | `0` | Cache maximum size in bytes, 0 for unlimited |
| `CACHE_CLEANUP_INTERVAL` | `0` | Cache cleanup interval in seconds, 0 for 10s default |
| `CACHE_STORAGE_TYPE` | `memory` | Cache storage type (memory, redis) |

### Command-line Flags

```bash
./prometheus-storagebox-exporter --help

Flags:
  --hetzner-token string           Hetzner API token (can also be set via HETZNER_TOKEN env var)
  --listen-address string          Address to listen on for HTTP requests (default ":9509")
  --metrics-path string            Path under which to expose metrics (default "/metrics")
  --log-level string               Log level (debug, info, warn, error) (default "info")
  --cache-ttl int                  Cache TTL in seconds, 0 to disable (can also be set via CACHE_TTL env var, default: 0 - disabled)
  --cache-max-size int64           Cache maximum size in bytes, 0 for unlimited (can also be set via CACHE_MAX_SIZE env var, default: 0 - unlimited)
  --cache-cleanup-interval int     Cache cleanup interval in seconds, 0 for default (can also be set via CACHE_CLEANUP_INTERVAL env var, default: 0 - 10s)
  --cache-storage-type string      Cache storage type (memory, redis) (can also be set via CACHE_STORAGE_TYPE env var, default: memory)
  --version                        Show version information and exit
```

### Cache Configuration (Optional)

> ⚠️ **Cache is disabled by default** following Prometheus best practices. Use `scrape_interval` in Prometheus instead of caching for most use cases.

#### When to Enable Cache

- **Multiple Prometheus instances**: Prevent duplicate API calls
- **Rate limiting concerns**: Reduce API request frequency
- **Development/Testing**: Minimize API calls during testing

#### Cache Configuration Examples

```bash
# Enable cache with 60s TTL, 1MB size limit
export CACHE_TTL=60
export CACHE_MAX_SIZE=1048576
export CACHE_CLEANUP_INTERVAL=30

# Enable with memory limit only (10s cleanup interval)
export CACHE_TTL=120
export CACHE_MAX_SIZE=5242880

# Enable with custom cleanup interval
export CACHE_TTL=300
export CACHE_CLEANUP_INTERVAL=60
```

#### Prometheus Configuration (Recommended Alternative)

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'storagebox-exporter'
    scrape_interval: 5m    # Control scraping frequency
    scrape_timeout: 30s
    static_configs:
      - targets: ['localhost:9509']
```

---

## 📊 Metrics

The exporter exposes 15+ metrics organized in 4 categories:

### Core Storage Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `storagebox_disk_quota_bytes` | Gauge | Total storage quota in bytes | id, name, server, location |
| `storagebox_disk_usage_bytes` | Gauge | Total used diskspace in bytes | id, name, server, location |
| `storagebox_disk_usage_data_bytes` | Gauge | Diskspace used by files in bytes | id, name, server, location |
| `storagebox_disk_usage_snapshots_bytes` | Gauge | Diskspace used by snapshots in bytes | id, name, server, location |

### Information & Status Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `storagebox_info` | Info | Storage box information (value always 1) | id, name, username, server, location, storage_type, system |
| `storagebox_status` | Gauge | Current status (1=active, 0=inactive) | id, name, status |
| `storagebox_created_timestamp` | Gauge | Unix timestamp of creation | id, name |

### Access Settings Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `storagebox_access_ssh_enabled` | Gauge | SSH access enabled (1=yes, 0=no) | id, name |
| `storagebox_access_samba_enabled` | Gauge | Samba/CIFS access enabled (1=yes, 0=no) | id, name |
| `storagebox_access_webdav_enabled` | Gauge | WebDAV access enabled (1=yes, 0=no) | id, name |
| `storagebox_access_zfs_enabled` | Gauge | ZFS access enabled (1=yes, 0=no) | id, name |
| `storagebox_reachable_externally` | Gauge | External reachability (1=yes, 0=no) | id, name |

### Protection & Snapshot Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `storagebox_snapshot_plan_enabled` | Gauge | Automatic snapshots configured (1=yes, 0=no) | id, name |
| `storagebox_protection_delete` | Gauge | Delete protection status (1=protected, 0=no) | id, name |

### Exporter Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `storagebox_exporter_scrape_duration_seconds` | Gauge | Duration of the scrape in seconds |
| `storagebox_exporter_scrape_errors_total` | Counter | Total number of scrape errors |
| `storagebox_exporter_up` | Gauge | Exporter health status (1=healthy, 0=unhealthy) |
| `storagebox_exporter_cache_hits_total` | Counter | Total number of cache hits (0 when cache disabled) |
| `storagebox_exporter_cache_misses_total` | Counter | Total number of cache misses (increments every scrape when cache disabled) |

---

## 📈 Grafana Dashboard

A comprehensive Grafana dashboard is included with 21 panels:



### Dashboard Features

- **📊 Overview Section**: Gauges for disk usage percentage and disk space distribution
- **📈 Time Series Graphs**:
  - Disk usage over time with quota visualization
  - Usage breakdown (Data vs Snapshots) with dual Y-axes
  - Disk usage percentage trends
  - Storage growth rate analysis (1h intervals)
- **📋 Detailed Table**: Complete storage box details with all metrics
- **🔧 Access Status**: Visual indicators for SSH, Samba, WebDAV, and ZFS access
- **🛡️ Configuration Info**: Snapshot plan and delete protection status
- **📊 Multi-box Support**: Variable to filter by specific storage box or view all

### Quick test with Docker Compose

The repository includes a complete Docker Compose test with:

```bash
# Start all services (Exporter + Prometheus + Grafana)
./test-env.sh

# Or manually:
docker-compose -f docker-compose.dev.yml up -d
```

Access points:
- 🎯 **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- 📈 **Prometheus**: http://localhost:9090
- 🔧 **Exporter**: http://localhost:9509/metrics

### Dashboard Panels

The dashboard includes:

<table>
<tr>
<td width="50%">

- Disk Usage Percentage (Gauge)
- Disk Space Distribution (Pie)
- Total Quota (Stat)
- Total Used (Stat)
- Free Space (Stat)
- Data Files Usage (Stat)

</td>
<td width="50%">

- Snapshots Usage (Stat)
- Snapshot Overhead (Stat)
- Disk Usage Over Time
- Usage Breakdown Over Time
- Disk Usage Percentage Over Time
- Storage Growth Rate (1h)

</td>
</tr>
</table>

**Access Settings Panels:**
- SSH Access Status
- Samba Access Status
- WebDAV Access Status
- ZFS Access Status
- External Reachability Status

**Configuration Panels:**
- Storage Box Status
- Snapshot Plan Status
- Delete Protection Status

**Details Table:**
- Storage Box Details (comprehensive table)



## 🐳 Docker Deployment

### Docker Compose

Complete `docker-compose.yml` example with Prometheus and Grafana:

<details>
<summary>Click to expand Docker Compose</summary>

```yaml
version: '3.8'

services:
  # Storage Box Exporter
  storagebox-exporter:
    image: ghcr.io/crstian19/prometheus-storagebox-exporter:latest
    container_name: storagebox-exporter
    restart: unless-stopped
    ports:
      - "9509:9509"
    environment:
      - HETZNER_TOKEN=${HETZNER_TOKEN}
      # Optional cache configuration (uncomment to enable)
      # - CACHE_TTL=60
      # - CACHE_MAX_SIZE=1048576
      # - CACHE_CLEANUP_INTERVAL=30
    networks:
      - monitoring

  # Prometheus
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
      - '--web.enable-lifecycle'
    networks:
      - monitoring
    depends_on:
      - storagebox-exporter

  # Grafana
  grafana:
    image: grafana/grafana:10.2.0
    container_name: grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana-provisioning:/etc/grafana/provisioning:ro
    networks:
      - monitoring
    depends_on:
      - prometheus

networks:
  monitoring:
    driver: bridge

volumes:
  prometheus-data:
  grafana-data:
```

</details>

### Prometheus Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'hetzner-storagebox'
    static_configs:
      - targets: ['storagebox-exporter:9509']
    scrape_interval: 60s
    scrape_timeout: 30s
```

---

## ☸️ Kubernetes Deployment

### Quick Deploy

```bash
# Apply all manifests
kubectl apply -f k8s/

# Check status
kubectl get pods -n monitoring
```

### Manifests

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
        image: ghcr.io/crstian19/prometheus-storagebox-exporter:latest
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

---

## 🏗️ Development

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
├── grafana-provisioning/  # Grafana dashboard provisioning
├── k8s/                   # Kubernetes manifests
├── .github/workflows/     # CI/CD pipelines
├── Dockerfile             # Multi-stage Docker build
├── docker-compose.yml     # Docker Compose configuration
├── docker-compose.dev.yml # Development environment
└── DESIGN.md             # Architecture documentation
```


## 🐛 Troubleshooting

### Common Issues

<details>
<summary><strong>Error: "HETZNER_TOKEN environment variable or --hetzner-token flag is required"</strong></summary>

Make sure you've set the `HETZNER_TOKEN` environment variable or passed it via the `--hetzner-token` flag.

```bash
export HETZNER_TOKEN="your-token-here"
./prometheus-storagebox-exporter
```

</details>

<details>
<summary><strong>Error: "API request failed with status 401"</strong></summary>

Your API token is invalid or has expired. Generate a new token from the Hetzner Cloud Console with **Read** permissions.

</details>

<details>
<summary><strong>Error: "API request failed with status 403"</strong></summary>

Your API token doesn't have sufficient permissions. Ensure the token has at least **Read** permissions.

</details>

<details>
<summary><strong>No metrics appearing in Prometheus</strong></summary>

1. Check exporter health: `curl http://localhost:9509/health`
2. Check metrics endpoint: `curl http://localhost:9509/metrics`
3. Verify Prometheus configuration
4. Check exporter logs: `docker logs storagebox-exporter`

</details>

<details>
<summary><strong>Grafana dashboard shows "No data"</strong></summary>

1. Verify Prometheus is scraping the exporter
2. Check the data source URL in Grafana
3. Ensure the storage box variable has values
4. Check the time range in Grafana

</details>

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Adding New Metrics

1. Update the collector in `internal/collector/storagebox.go`
2. Add metric definitions
3. Update the Grafana dashboard if needed
4. Update this README

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Credits

### Technologies
- [Hetzner Storage Box](https://www.hetzner.com/storage/storage-box) - Backup storage solution
- [Hetzner Cloud API](https://docs.hetzner.cloud/) - Modern API infrastructure
- [Prometheus](https://prometheus.io/) - Monitoring toolkit and time series database
- [Grafana](https://grafana.com/) - Analytics and monitoring platform
- [Go](https://golang.org/) - Programming language

### Inspiration
- [fleaz/prometheus-storagebox-exporter](https://github.com/fleaz/prometheus-storagebox-exporter) - Original implementation
- [Prometheus exporter best practices](https://github.com/prometheus/client_golang)

---

## 📞 Support

- **Issues**: [Report a bug](https://github.com/crstian19/prometheus-storagebox-exporter/issues)
- **Features**: [Request a feature](https://github.com/crstian19/prometheus-storagebox-exporter/issues/new?template=feature_request.md)
- **Security**: [Report a vulnerability](https://github.com/crstian19/prometheus-storagebox-exporter/security)
- **Documentation**: [DESIGN.md](DESIGN.md) for architecture details

## 💝 Donate

If you find this project useful and want to support its development, you can donate via PayPal:

[![Donate with PayPal](https://raw.githubusercontent.com/stefan-niedermann/paypal-donate-button/master/paypal-donate-button.png)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=MUW2XFMQB2782)

Your support helps keep this project maintained and improved! 🙏

---

<div align="center">

**⭐ If this project helped you, consider giving it a star!**

Made with ❤️ from 🇪🇸 for the Prometheus and Hetzner communities

</div>