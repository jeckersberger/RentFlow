#!/bin/bash
# Versioned CrateDesk database migration runner.
# Records each applied migration with a SHA-256 checksum and never suppresses
# SQL errors. Existing migration files are treated as immutable.

set -euo pipefail

export PGHOST="${PGHOST:-postgres}"
export PGPORT="${PGPORT:-5432}"
export PGUSER="${PGUSER:-rentflow}"
export PGPASSWORD="${PGPASSWORD:-rentflow_dev}"
MIGRATIONS_ROOT="${MIGRATIONS_ROOT:-/services}"

ALL_SERVICES="auth-service inventory-service project-service scanner-service warehouse-service invoice-service document-service crew-service federation-service maintenance-service transport-service insurance-service workflow-service ai-service notification-service reporting-service audit-service expense-service"

service_database() {
  case "$1" in
    auth-service) echo auth_service ;;
    inventory-service) echo inventory_service ;;
    project-service) echo project_service ;;
    scanner-service) echo scanner_service ;;
    warehouse-service) echo warehouse_service ;;
    invoice-service) echo invoice_service ;;
    document-service) echo document_service ;;
    crew-service) echo crew_service ;;
    federation-service) echo federation_service ;;
    maintenance-service) echo maintenance_service ;;
    transport-service) echo transport_service ;;
    insurance-service) echo insurance_service ;;
    workflow-service) echo workflow_service ;;
    ai-service) echo ai_service ;;
    notification-service) echo notification_service ;;
    reporting-service) echo reporting_service ;;
    audit-service) echo audit_service ;;
    expense-service) echo expense_service ;;
    *) return 1 ;;
  esac
}

checksum_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  else
    echo "ERROR: no SHA-256 utility available" >&2
    return 1
  fi
}

collect_migrations() {
  find "$1" -maxdepth 1 -type f -name '*.sql' ! -name '*.down.sql' -print | LC_ALL=C sort
}

migrate_service() {
  local svc="$1"
  local db migration_dir file base checksum existing version normalized_version script
  local -A seen_versions=()
  local -a files=()

  db="$(service_database "$svc")" || {
    echo "ERROR: unknown service: $svc" >&2
    return 1
  }
  migration_dir="$MIGRATIONS_ROOT/$svc/migrations"

  if [[ ! -d "$migration_dir" ]]; then
    echo "SKIP $svc (no migration directory)"
    return 0
  fi

  mapfile -t files < <(collect_migrations "$migration_dir")
  if (( ${#files[@]} == 0 )); then
    echo "SKIP $svc (no migrations)"
    return 0
  fi

  for file in "${files[@]}"; do
    base="$(basename "$file")"
    if [[ ! "$base" =~ ^[0-9]+_[A-Za-z0-9._-]+\.sql$ ]]; then
      echo "ERROR: invalid migration filename in $svc: $base" >&2
      return 1
    fi
    version="${base%%_*}"
    normalized_version="$((10#$version))"
    if [[ -n "${seen_versions[$normalized_version]:-}" ]]; then
      echo "ERROR: duplicate migration version $version in $svc: ${seen_versions[$normalized_version]} and $base" >&2
      return 1
    fi
    seen_versions[$normalized_version]="$base"
  done

  echo "MIGRATE $svc -> $db"

  psql -X -v ON_ERROR_STOP=1 -d "$db" <<SQL
CREATE TABLE IF NOT EXISTS public.cratedesk_schema_migrations (
  service TEXT NOT NULL,
  filename TEXT NOT NULL,
  checksum CHAR(64) NOT NULL,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (service, filename)
);
SQL

  for file in "${files[@]}"; do
    base="$(basename "$file")"
    checksum="$(checksum_file "$file")"
    existing="$(psql -X -v ON_ERROR_STOP=1 -At -d "$db" -c "SELECT checksum FROM public.cratedesk_schema_migrations WHERE service = '$svc' AND filename = '$base'")"

    if [[ -n "$existing" ]]; then
      if [[ "$existing" != "$checksum" ]]; then
        echo "ERROR: checksum mismatch for already-applied migration $svc/$base" >&2
        return 1
      fi
      echo "SKIP $svc/$base (already applied)"
      continue
    fi

    script="$(mktemp)"
    cat >"$script" <<SQL
\\set ON_ERROR_STOP on
BEGIN;
SELECT pg_advisory_xact_lock(hashtext('cratedesk:migrations:$svc'));
SELECT EXISTS (
  SELECT 1 FROM public.cratedesk_schema_migrations
  WHERE service = '$svc' AND filename = '$base'
) AS migration_exists \\gset
\\if :migration_exists
  \\echo 'SKIP $svc/$base (applied by another runner)'
\\else
  \\echo 'APPLY $svc/$base'
  \\i '$file'
  INSERT INTO public.cratedesk_schema_migrations(service, filename, checksum)
  VALUES ('$svc', '$base', '$checksum');
\\endif
COMMIT;
SQL

    if ! psql -X -d "$db" -f "$script"; then
      rm -f "$script"
      return 1
    fi
    rm -f "$script"
  done

  echo "DONE $svc"
}

echo "=== CrateDesk Versioned Migrations ==="
echo "Connecting to $PGHOST:$PGPORT as $PGUSER"

if (( $# > 1 )); then
  echo "Usage: $0 [service-name]" >&2
  exit 2
fi

if (( $# == 1 )); then
  migrate_service "$1"
else
  for svc in $ALL_SERVICES; do
    migrate_service "$svc"
  done
fi

echo "=== All requested migrations complete. ==="
