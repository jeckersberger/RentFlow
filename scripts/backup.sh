#!/bin/bash
set -euo pipefail

# CrateDesk Database Backup Script
# Runs daily via cron: 0 3 * * * /opt/cratedesk/scripts/backup.sh >> /var/log/cratedesk-backup.log 2>&1

BACKUP_DIR="${BACKUP_DIR:-/opt/cratedesk/backups}"
CONTAINER_NAME="${CONTAINER_NAME:-cratedesk-postgres}"
DB_USER="${DB_USER:-cratedesk}"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/cratedesk_${DATE}.sql.gz"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

# Optional NAS copy
NAS_USER="${NAS_USER:-janis}"
NAS_HOST="${NAS_HOST:-192.168.178.148}"
NAS_PATH="${NAS_PATH:-/volume3/backups/cratedesk/}"

mkdir -p "${BACKUP_DIR}"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting backup..."

# Dump all databases via Docker
if ! docker exec "${CONTAINER_NAME}" pg_dumpall -U "${DB_USER}" | gzip > "${BACKUP_FILE}"; then
    echo "ERROR: pg_dumpall failed" >&2
    rm -f "${BACKUP_FILE}"
    exit 1
fi

# Verify backup is non-empty (at least 1KB)
FILESIZE=$(stat -c%s "${BACKUP_FILE}" 2>/dev/null || stat -f%z "${BACKUP_FILE}" 2>/dev/null || echo 0)
if [ "${FILESIZE}" -lt 1024 ]; then
    echo "ERROR: Backup too small (${FILESIZE} bytes)" >&2
    exit 1
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup created: ${BACKUP_FILE} (${FILESIZE} bytes)"

# Delete backups older than retention period
DELETED=$(find "${BACKUP_DIR}" -name "cratedesk_*.sql.gz" -mtime +${RETENTION_DAYS} -delete -print | wc -l)
if [ "${DELETED}" -gt 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaned up ${DELETED} old backup(s)"
fi

# Copy to NAS (optional, non-fatal)
if [ -n "${NAS_HOST}" ]; then
    if scp -o ConnectTimeout=10 "${BACKUP_FILE}" "${NAS_USER}@${NAS_HOST}:${NAS_PATH}" 2>/dev/null; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] NAS copy OK"
    else
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] WARN: NAS copy failed (non-fatal)"
    fi
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup complete"
