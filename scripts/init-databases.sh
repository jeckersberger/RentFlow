#!/bin/bash
# =============================================================================
# CrateDesk — PostgreSQL Init Script
# Erstellt eine separate Datenbank pro Microservice.
# Wird automatisch beim ersten Start von PostgreSQL ausgefuehrt.
# =============================================================================

set -euo pipefail

DATABASES=(
  auth_service
  inventory_service
  project_service
  scanner_service
  warehouse_service
  invoice_service
  document_service
  crew_service
  federation_service
  maintenance_service
  transport_service
  insurance_service
  workflow_service
  ai_service
  notification_service
  reporting_service
  audit_service
  expense_service
  customer_service
)

echo "=== CrateDesk: Erstelle ${#DATABASES[@]} Service-Datenbanken ==="

for db in "${DATABASES[@]}"; do
  echo "  -> Erstelle Datenbank: $db"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-SQL
    SELECT 'CREATE DATABASE $db OWNER $POSTGRES_USER'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$db')\gexec
SQL
done

echo "=== CrateDesk: Alle Datenbanken erstellt ==="
