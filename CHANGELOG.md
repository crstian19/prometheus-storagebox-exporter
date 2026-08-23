# Changelog

## [v0.7.0](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.7.0)

[Compare to previous version](https://github.com/crstian19/prometheus-storagebox-exporter/compare/v0.6.0...v0.7.0)

### Features

- **deps**: update Go toolchain to 1.27.0 (#91) ([8dc5336](https://github.com/crstian19/prometheus-storagebox-exporter/commit/8dc533655961fde8ff4da26687a840b837ea3a92))

### Bug Fixes

- **collector**: derive snapshot_plan_enabled from plan presence (#90) ([23935f6](https://github.com/crstian19/prometheus-storagebox-exporter/commit/23935f67fca6480979ed9ea9cf846fb3c8c26cf6))
- **deps**: update module github.com/prometheus/client_golang to v1.24.1 (#86) ([28b410f](https://github.com/crstian19/prometheus-storagebox-exporter/commit/28b410f84058dacf2c76546b9168b1ce44056ddb))

## [v0.6.0](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.6.0)

[Compare to previous version](https://github.com/crstian19/prometheus-storagebox-exporter/compare/v0.5.6...v0.6.0)

### Features

- **collector**: expose up and build_info metrics ([31194ae](https://github.com/crstian19/prometheus-storagebox-exporter/commit/31194ae539678dc1e03584a6fd577fff438e51c3))

## [v0.5.6](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.5.6)

### Bug Fixes

- **release**: document releasable commit types

## [v0.5.5](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.5.5)

### Bug Fixes

- **deps**: update go to v1.26.2 (#66)

## [v0.5.4](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.5.4)

### Bug Fixes

- **ci**: use PAT for tag creation and clean up pipeline triggers (#56)

## [v0.5.3](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.5.3)

### Bug Fixes

- **ci**: fix docker build trigger, double release and goreleaser deprecations (#53)

## [v0.5.2](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.5.2)

### Bug Fixes

- **ci**: trigger docker build on release published event (#50)

## [v0.5.1](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.5.1)

### Bug Fixes

- **grafana-dashboard**: unify stat panel colors and fix label consistency (#48)

## [v0.5.0](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.5.0)

### Features

- **deps**: update Go to v1.26.1 (#46)

## [v0.4.0](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.4.0)

### Features

- **main.go**: enhance HTML and CSS for Prometheus Hetzner Storage Box Exporter

### Bug Fixes

- **dockerfile**: use TARGETARCH for multi-arch builds
- **dockerfile**: use TARGETARCH for multi-arch builds

## [v0.2.0](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.2.0)

### Features

- merge metrics cache functionality
- **hetzner**: add detailed error handling and metrics

## [v0.1.0](https://github.com/crstian19/prometheus-storagebox-exporter/releases/tag/v0.1.0)

### Features

- **logging**: replace log with structured logging using slog

### Bug Fixes

- **grafana-dashboard**: update and clean dashboard JSON
