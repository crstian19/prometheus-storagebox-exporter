# Multi-stage build for minimal final image

# Stage 1: Build the Go binary
FROM golang:1.23.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 for static linking (required for scratch image)
# -ldflags for embedding version information
ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE}" \
    -o prometheus-storagebox-exporter .

# Stage 2: Create minimal runtime image
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /build/prometheus-storagebox-exporter /prometheus-storagebox-exporter

# Expose metrics port
EXPOSE 9509

# Run as non-root user
USER 65534:65534

# Set entry point
ENTRYPOINT ["/prometheus-storagebox-exporter"]
