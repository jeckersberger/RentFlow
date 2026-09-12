#!/bin/bash
# Integration smoke test for migrate-internal.sh. Requires a reachable PostgreSQL
# instance and permission to create/drop the auth_service test database.

set -euo pipefail

export PGHOST="${PGHOST:-127.0.0.1}"
export PGPORT="${PGPORT:-5432}"
export PGUSER="${PGUSER:-postgres}"
export PGPASSWORD="${PGPASSWORD:-postgres}"

fixture_root="$(mktemp -d)"
cleanup() {
  rm -rf "$fixture_root"
  dropdb --if-exists auth_service >/dev/null 2>&1 || true
}
trap cleanup EXIT

createdb auth_service
mkdir -p "$fixture_root/auth-service/migrations"

cat >"$fixture_root/auth-service/migrations/001_create_probe.sql" <<'SQL'
CREATE TABLE migration_probe (
  id INTEGER PRIMARY KEY,
  note TEXT NOT NULL
);
INSERT INTO migration_probe(id, note) VALUES (1, 'first apply');
SQL

MIGRATIONS_ROOT="$fixture_root" MIGRATION_LEGACY_DUPLICATES="scripts/migration-legacy-duplicates.txt" \
  bash scripts/migrate-internal.sh auth-service

# A second run must be a no-op, not a second insert/application.
MIGRATIONS_ROOT="$fixture_root" MIGRATION_LEGACY_DUPLICATES="scripts/migration-legacy-duplicates.txt" \
  bash scripts/migrate-internal.sh auth-service

probe_count="$(psql -X -At -d auth_service -c 'SELECT COUNT(*) FROM migration_probe')"
ledger_count="$(psql -X -At -d auth_service -c "SELECT COUNT(*) FROM public.cratedesk_schema_migrations WHERE service = 'auth-service' AND filename = '001_create_probe.sql'")"

[[ "$probe_count" == "1" ]] || {
  echo "ERROR: migration was applied more than once (probe_count=$probe_count)" >&2
  exit 1
}
[[ "$ledger_count" == "1" ]] || {
  echo "ERROR: migration ledger does not contain exactly one entry (ledger_count=$ledger_count)" >&2
  exit 1
}

# Historical drift must be rejected once the file has been recorded.
printf '\n-- modified after apply\n' >>"$fixture_root/auth-service/migrations/001_create_probe.sql"
if MIGRATIONS_ROOT="$fixture_root" MIGRATION_LEGACY_DUPLICATES="scripts/migration-legacy-duplicates.txt" \
  bash scripts/migrate-internal.sh auth-service; then
  echo "ERROR: checksum drift was accepted" >&2
  exit 1
fi

echo "Migration runner integration test passed"
