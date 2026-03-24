#!/bin/bash
set -e

# =============================================================================
# RentFlow Update Script
# Pulls latest images and restarts services with zero-downtime approach
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()   { echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"; }
warn()  { echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] WARN:${NC} $1"; }
error() { echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1" >&2; }

cd "$PROJECT_DIR"

# Determine compose command
if docker compose version > /dev/null 2>&1; then
    COMPOSE="docker compose"
elif command -v docker-compose > /dev/null 2>&1; then
    COMPOSE="docker-compose"
else
    error "Docker Compose nicht gefunden. Bitte Docker Compose v2 installieren."
    exit 1
fi

# Determine compose files
COMPOSE_FILES="-f docker-compose.yml"
if [ -f docker-compose.prod.yml ]; then
    COMPOSE_FILES="$COMPOSE_FILES -f docker-compose.prod.yml"
fi

# Check if specific services were passed as arguments
SERVICES="$*"

log "========================================="
log "  RentFlow Update"
log "========================================="

# 1. Pre-flight check
log "Pruefe laufende Services..."
$COMPOSE $COMPOSE_FILES ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null || true

# 2. Create backup before update
if [ -x "$SCRIPT_DIR/backup.sh" ]; then
    log "Erstelle Backup vor dem Update..."
    "$SCRIPT_DIR/backup.sh" || warn "Backup fehlgeschlagen -- Update wird trotzdem fortgesetzt"
else
    warn "Backup-Script nicht gefunden, ueberspringe Backup"
fi

# 3. Pull latest images
log "Lade neueste Images herunter..."
if [ -n "$SERVICES" ]; then
    $COMPOSE $COMPOSE_FILES pull $SERVICES
else
    $COMPOSE $COMPOSE_FILES pull
fi

# 4. Rebuild local images if needed
log "Baue lokale Images neu..."
if [ -n "$SERVICES" ]; then
    $COMPOSE $COMPOSE_FILES build --no-cache $SERVICES
else
    $COMPOSE $COMPOSE_FILES build --no-cache
fi

# 5. Restart services
log "Starte Services neu..."
if [ -n "$SERVICES" ]; then
    $COMPOSE $COMPOSE_FILES up -d --remove-orphans $SERVICES
else
    $COMPOSE $COMPOSE_FILES up -d --remove-orphans
fi

# 6. Wait for health checks
log "Warte auf Health-Checks (max 120s)..."
TIMEOUT=120
ELAPSED=0
while [ $ELAPSED -lt $TIMEOUT ]; do
    UNHEALTHY=$($COMPOSE $COMPOSE_FILES ps --format "{{.Name}} {{.Status}}" 2>/dev/null | grep -c "starting\|unhealthy" || true)
    if [ "$UNHEALTHY" -eq 0 ]; then
        break
    fi
    sleep 5
    ELAPSED=$((ELAPSED + 5))
    echo -n "."
done
echo ""

if [ $ELAPSED -ge $TIMEOUT ]; then
    warn "Timeout: Einige Services sind noch nicht healthy"
    $COMPOSE $COMPOSE_FILES ps
    echo ""
    warn "Pruefen mit: docker compose ps"
    warn "Rollback mit: ./scripts/rollback.sh"
    exit 1
fi

# 7. Cleanup old images
log "Raeume alte Images auf..."
docker image prune -f > /dev/null 2>&1 || true

# 8. Final status
log "========================================="
log "  Update abgeschlossen!"
log "========================================="
$COMPOSE $COMPOSE_FILES ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null || $COMPOSE $COMPOSE_FILES ps
