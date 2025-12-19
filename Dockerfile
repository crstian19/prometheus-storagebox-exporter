# Stage 1: Build
FROM golang:1.25.5-alpine AS builder

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build args for version info
ARG VERSION=dev
ARG GIT_COMMIT=none
ARG BUILD_DATE=unknown
ARG TARGETARCH

# Build static binary for multi-arch support
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE}" \
    -o prometheus-storagebox-exporter .

# Stage 2: Runtime
FROM scratch

# CA certs for HTTPS and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /build/prometheus-storagebox-exporter /prometheus-storagebox-exporter

EXPOSE 9509

USER 65534:65534

ENTRYPOINT ["/prometheus-storagebox-exporter"]
