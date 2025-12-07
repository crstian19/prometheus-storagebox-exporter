#!/bin/bash

# Quick start script for local testing with Grafana

set -e

echo "=========================================="
echo "Hetzner Storage Box Exporter - Local Test"
echo "=========================================="
echo ""

# Check if HETZNER_TOKEN is set
if [ -z "$HETZNER_TOKEN" ]; then
    echo "ERROR: HETZNER_TOKEN environment variable is not set!"
    echo ""
    echo "Please set it before running:"
    echo "  export HETZNER_TOKEN=your-api-token"
    echo ""
    exit 1
fi

echo "✓ HETZNER_TOKEN is set"
echo ""

# Build and start services
echo "Building and starting services..."
docker-compose -f docker-compose.dev.yml up -d --build

echo ""
echo "=========================================="
echo "Services started successfully!"
echo "=========================================="
echo ""
echo "📊 Grafana:   http://localhost:3000"
echo "   Username:  admin"
echo "   Password:  admin"
echo ""
echo "📈 Prometheus: http://localhost:9090"
echo "🔧 Exporter:   http://localhost:9509/metrics"
echo ""
echo "The Grafana dashboard 'Hetzner Storage Box Overview' is automatically loaded."
echo ""
echo "To stop the services, run:"
echo "  docker-compose -f docker-compose.dev.yml down"
echo ""
echo "To view logs:"
echo "  docker-compose -f docker-compose.dev.yml logs -f"
echo ""
