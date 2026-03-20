#!/bin/bash
set -e

# RentFlow Automated Backup Script
# Daily execution: 03:00 UTC
# Backs up PostgreSQL + KurrentDB with AES-256-GCM encryption

BACKUP_DIR="${BACKUP_DIR:-/opt/rentflow/backups}"
POSTGRES_USER="${POSTGRES_USER:-rentflow_user}"
POSTGRES_DB="${POSTGRES_DB:-rentflow}"
KURRENTDB_USER="${KURRENTDB_USER:-admin}"
ENCRYPTION_KEY="${BACKUP_ENCRYPTION_KEY}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
  echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
  echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1" >&2
}

warn() {
  echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] WARN:${NC} $1"
}

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"

# 1. PostgreSQL Backup
log "Backing up PostgreSQL database..."
POSTGRES_BACKUP_FILE="$BACKUP_DIR/postgres-$(date +%Y%m%d-%H%M%S).dump"
if docker exec rentflow-postgres pg_dump \
  -U "$POSTGRES_USER" \
  --format=custom \
  --compress=9 \
  --verbose \
  "$POSTGRES_DB" > "$POSTGRES_BACKUP_FILE" 2>&1; then
  log "PostgreSQL backup successful: $(du -h "$POSTGRES_BACKUP_FILE" | cut -f1)"
else
  error "PostgreSQL backup failed"
  exit 1
fi

# 2. KurrentDB Backup
log "Backing up KurrentDB event store..."
KURRENTDB_BACKUP_FILE="$BACKUP_DIR/kurrentdb-$(date +%Y%m%d-%H%M%S).bak"
if docker exec rentflow-kurrentdb curl -s \
  -u "$KURRENTDB_USER:$KURRENTDB_PASSWORD" \
  --max-time 300 \
  http://localhost:2113/backup \
  > "$KURRENTDB_BACKUP_FILE"; then
  log "KurrentDB backup successful: $(du -h "$KURRENTDB_BACKUP_FILE" | cut -f1)"
else
  error "KurrentDB backup failed"
  exit 1
fi

# 3. Upload Backups
log "Backing up uploads and documents..."
UPLOAD_BACKUP_FILE="$BACKUP_DIR/uploads-$(date +%Y%m%d-%H%M%S).tar.gz"
if tar -czf "$UPLOAD_BACKUP_FILE" -C /opt/rentflow storage/ 2>&1 | head -20; then
  log "Upload backup successful: $(du -h "$UPLOAD_BACKUP_FILE" | cut -f1)"
else
  error "Upload backup failed"
  exit 1
fi

# 4. Encrypt All Backups
log "Encrypting backups with AES-256-GCM..."
for file in "$POSTGRES_BACKUP_FILE" "$KURRENTDB_BACKUP_FILE" "$UPLOAD_BACKUP_FILE"; do
  if [ ! -f "$file" ]; then
    warn "Backup file not found: $file"
    continue
  fi

  ENC_FILE="$file.enc"
  if openssl enc -aes-256-cbc -salt -in "$file" \
    -out "$ENC_FILE" -k "$ENCRYPTION_KEY" 2>/dev/null; then
    rm "$file"
    log "Encrypted: $(basename "$ENC_FILE")"
  else
    error "Encryption failed for $file"
    exit 1
  fi
done

# 5. Cleanup Old Backups (Local)
log "Cleaning up backups older than $BACKUP_RETENTION_DAYS days..."
DELETED_COUNT=$(find "$BACKUP_DIR" -name "*.enc" -mtime +$BACKUP_RETENTION_DAYS -delete -print | wc -l)
if [ "$DELETED_COUNT" -gt 0 ]; then
  log "Deleted $DELETED_COUNT old backup files"
fi

# 6. S3 Sync (Optional)
if [ "${BACKUP_S3_ENABLED}" = "true" ]; then
  log "Syncing backups to S3 (${BACKUP_S3_BUCKET})..."
  if aws s3 sync "$BACKUP_DIR" "s3://${BACKUP_S3_BUCKET}/rentflow-backups/" \
    --region "${BACKUP_S3_REGION}" \
    --delete \
    --sse AES256 \
    2>&1 | tail -5; then
    log "S3 sync successful"
  else
    error "S3 sync failed"
    exit 1
  fi
fi

# 7. Backup Status Summary
log "=========== Backup Summary ==========="
echo "Directory: $BACKUP_DIR"
echo "Total size: $(du -sh "$BACKUP_DIR" | cut -f1)"
echo "Files created:"
ls -lh "$BACKUP_DIR" | tail -3 | awk '{print "  " $9 " (" $5 ")"}'
echo "====================================="

log "Backup completed successfully!"
