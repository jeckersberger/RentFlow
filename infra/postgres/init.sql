-- ═══════════════════════════════════════════════════════════
-- RentFlow PostgreSQL Initialisierung
-- Schema-per-Service Isolation (alle in einer DB: rentflow)
-- ═══════════════════════════════════════════════════════════

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- ─── Schemas anlegen ──────────────────────────────────────
CREATE SCHEMA IF NOT EXISTS auth_schema;
CREATE SCHEMA IF NOT EXISTS inventory_schema;
CREATE SCHEMA IF NOT EXISTS project_schema;
CREATE SCHEMA IF NOT EXISTS scanner_schema;
CREATE SCHEMA IF NOT EXISTS warehouse_schema;
CREATE SCHEMA IF NOT EXISTS invoice_schema;
CREATE SCHEMA IF NOT EXISTS document_schema;
CREATE SCHEMA IF NOT EXISTS crew_schema;
CREATE SCHEMA IF NOT EXISTS federation_schema;
CREATE SCHEMA IF NOT EXISTS maintenance_schema;
CREATE SCHEMA IF NOT EXISTS transport_schema;
CREATE SCHEMA IF NOT EXISTS insurance_schema;
CREATE SCHEMA IF NOT EXISTS workflow_schema;
CREATE SCHEMA IF NOT EXISTS ai_schema;
CREATE SCHEMA IF NOT EXISTS notification_schema;
CREATE SCHEMA IF NOT EXISTS reporting_schema;
CREATE SCHEMA IF NOT EXISTS audit_schema;
CREATE SCHEMA IF NOT EXISTS expense_schema;

-- ─── Bestätigung ──────────────────────────────────────────
DO $$
DECLARE
    schema_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO schema_count
    FROM information_schema.schemata
    WHERE schema_name LIKE '%_schema';
    RAISE NOTICE 'RentFlow: % Service-Schemas angelegt', schema_count;
END $$;
