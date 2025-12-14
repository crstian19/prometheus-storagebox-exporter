# ============================================================================
# Multi-stage Dockerfile for Prometheus Storage Box Exporter
# ============================================================================

# ============================================================================
# Stage 1: Builder - Compile the Go application
# ============================================================================
FROM golang:1.25.5-alpine AS builder

# Install build dependencies (git for go modules, ca-certificates for HTTPS)
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /build

# ============================================================================
# Dependency layer caching optimization
# Copy go.mod and go.sum first, then download dependencies
# This allows Docker to cache dependencies if source code changes
# ============================================================================
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ============================================================================
# Copy source code
# This is done AFTER dependency download to maximize cache hits
# ============================================================================
COPY . .

# ============================================================================
# Build the static binary
# ============================================================================
# Build arguments for version information
ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_DATE=unknown

# Compile with optimization flags:
# - CGO_ENABLED=0: Disable CGO for static linking (required for scratch)
# - GOOS=linux: Target Linux OS
# - GOARCH=amd64: Target AMD64 architecture (change for multi-arch)
# - -ldflags="-w -s": Strip debug info and symbol table (reduces size ~30%)
# - -ldflags="-X ...": Embed version information at compile time
# - -trimpath: Remove file system paths from binary for reproducibility
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE}" \
    -o prometheus-storagebox-exporter .

# ============================================================================
# Stage 2: Runtime - Minimal final image
# ============================================================================
FROM scratch

# Copy CA certificates from builder (required for HTTPS requests to Hetzner API)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data from builder (required for proper time handling)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the compiled static binary
COPY --from=builder /build/prometheus-storagebox-exporter /prometheus-storagebox-exporter

# Expose the metrics port
EXPOSE 9509

USER 65534:65534

ENTRYPOINT ["/prometheus-storagebox-exporter"]
