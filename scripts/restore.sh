#!/bin/bash
set -e

# RentFlow Restore Script
# Usage: ./scripts/restore.sh <backup_file>
# Example: ./scripts/restore.sh backups/postgres-20260321-030000.dump

BACKUP_FILE="${1:-}"
POSTGRES_USER="${POSTGRES_USER:-rentflow_user}"
POSTGRES_DB="${POSTGRES_DB:-rentflow}"
ENCRYPTION_KEY="${BACKUP_ENCRYPTION_KEY}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
  echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
  echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1" >&2
  exit 1
}

if [ -z "$BACKUP_FILE" ]; then
  echo "Usage: $0 <backup_file>"
  echo "Available backups:"
  ls -lh /opt/rentflow/backups/*.dump 2>/dev/null || echo "No backups found"
  exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
  error "Backup file not found: $BACKUP_FILE"
fi

log "Restoring from: $BACKUP_FILE"

# Decrypt if needed
if [[ "$BACKUP_FILE" == *.enc ]]; then
  log "Decrypting backup..."
  DECRYPTED_FILE="${BACKUP_FILE%.enc}"
  if ! openssl enc -d -aes-256-cbc -in "$BACKUP_FILE" \
    -out "$DECRYPTED_FILE" -k "$ENCRYPTION_KEY"; then
    error "Decryption failed"
  fi
  BACKUP_FILE="$DECRYPTED_FILE"
fi

# Restore PostgreSQL
if [[ "$BACKUP_FILE" == *postgres* ]]; then
  log "Restoring PostgreSQL database..."
  if docker exec rentflow-postgres pg_restore \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    --verbose \
    "$BACKUP_FILE"; then
    log "PostgreSQL restore successful"
  else
    error "PostgreSQL restore failed"
  fi
fi

log "Restore completed successfully!"
log "Services will auto-reconnect. Monitor health with: docker compose ps"
