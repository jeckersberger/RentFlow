#!/bin/bash
# =============================================================================
# CrateDesk — Production Deployment Script
# Builds services SEQUENTIALLY to avoid OOM on small servers (CX33 = 4GB RAM)
# Usage: ssh root@46.224.105.10 "cd /opt/cratedesk && bash scripts/deploy.sh"
# =============================================================================

set -euo pipefail

echo "=== CrateDesk Deployment ==="
echo "  Host: $(hostname)"
echo "  Date: $(date)"
echo ""

# Pull latest code
echo "=== Step 1: Pull latest code ==="
git fetch origin develop
git reset --hard origin/develop

# Build infrastructure first
echo "=== Step 2: Build infrastructure ==="
docker compose pull postgres redis traefik 2>/dev/null || true

# Build frontend (small, fast)
echo "=== Step 3: Build frontend ==="
docker compose build frontend

# Build Go services one at a time to avoid OOM
SERVICES=(auth inventory project scanner warehouse invoice document crew federation maintenance transport insurance workflow ai notification reporting audit expense customer)

echo "=== Step 4: Build ${#SERVICES[@]} Go services (sequential) ==="
for svc in "${SERVICES[@]}"; do
  echo "  Building: $svc ..."
  docker compose build "$svc" 2>&1 | tail -1
done

echo ""
echo "=== Step 5: Start all services ==="
docker compose up -d

echo ""
echo "=== Step 6: Wait for health checks ==="
sleep 15

echo ""
echo "=== Step 7: Service status ==="
docker compose ps --format "table {{.Name}}\t{{.Status}}"

echo ""
echo "=== Deployment complete ==="
