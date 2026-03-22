#!/bin/bash
# RentFlow Database Migration Script
# Runs all SQL migrations for each service against its database

set -e

POSTGRES_CONTAINER="rentflow-postgres"
POSTGRES_USER="rentflow"

# Map of service-name -> database-name
declare -A SERVICES=(
  ["auth-service"]="auth_service"
  ["inventory-service"]="inventory_service"
  ["project-service"]="project_service"
  ["scanner-service"]="scanner_service"
  ["warehouse-service"]="warehouse_service"
  ["invoice-service"]="invoice_service"
  ["document-service"]="document_service"
  ["crew-service"]="crew_service"
  ["federation-service"]="federation_service"
  ["maintenance-service"]="maintenance_service"
  ["transport-service"]="transport_service"
  ["insurance-service"]="insurance_service"
  ["workflow-service"]="workflow_service"
  ["ai-service"]="ai_service"
  ["notification-service"]="notification_service"
  ["reporting-service"]="reporting_service"
  ["audit-service"]="audit_service"
  ["expense-service"]="expense_service"
)

for svc in "${!SERVICES[@]}"; do
  db="${SERVICES[$svc]}"
  migration_dir="services/$svc/migrations"

  if [ ! -d "$migration_dir" ] || [ -z "$(ls -A $migration_dir/*.sql 2>/dev/null)" ]; then
    echo "SKIP $svc (no migrations)"
    continue
  fi

  echo "MIGRATE $svc → $db"
  for f in $(ls $migration_dir/*.sql | grep -v '.down.sql' | sort); do
    echo "  Running $(basename $f)..."
    cat "$f" | docker exec -i $POSTGRES_CONTAINER psql -U $POSTGRES_USER -d $db -q 2>&1 | grep -v "already exists" | grep -v "NOTICE" || true
  done
  echo "  Done."
done

echo "All migrations complete."
