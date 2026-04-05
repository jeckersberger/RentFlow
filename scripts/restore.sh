#!/bin/bash
set -euo pipefail

# CrateDesk Database Restore Script
# Usage: ./restore.sh <backup-file.sql.gz>

BACKUP_FILE="${1:?Usage: restore.sh <backup.sql.gz>}"
CONTAINER_NAME="${CONTAINER_NAME:-cratedesk-postgres}"
DB_USER="${DB_USER:-cratedesk}"

if [ ! -f "${BACKUP_FILE}" ]; then
    echo "ERROR: File not found: ${BACKUP_FILE}" >&2
    exit 1
fi

echo "========================================"
echo "  CrateDesk Database Restore"
echo "========================================"
echo "File:      ${BACKUP_FILE}"
echo "Container: ${CONTAINER_NAME}"
echo "User:      ${DB_USER}"
echo ""
echo "ACHTUNG: Alle Datenbanken werden ueberschrieben!"
echo "Druecke Ctrl+C innerhalb von 5 Sekunden zum Abbrechen."
echo ""

sleep 5

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting restore..."

if gunzip -c "${BACKUP_FILE}" | docker exec -i "${CONTAINER_NAME}" psql -U "${DB_USER}" 2>&1; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Restore completed successfully"
else
    echo "ERROR: Restore failed" >&2
    exit 1
fi

echo ""
echo "Bitte alle Services neustarten: docker compose restart"
