-- ═══════════════════════════════════════════════════════════
-- RentFlow PostgreSQL Initialisierung
-- Separate Datenbanken pro Service
-- ═══════════════════════════════════════════════════════════

-- ─── Datenbanken anlegen ─────────────────────────────────
CREATE DATABASE auth_service;
CREATE DATABASE inventory_service;
CREATE DATABASE project_service;
CREATE DATABASE scanner_service;
CREATE DATABASE warehouse_service;
CREATE DATABASE invoice_service;
CREATE DATABASE document_service;
CREATE DATABASE crew_service;
CREATE DATABASE federation_service;
CREATE DATABASE maintenance_service;
CREATE DATABASE transport_service;
CREATE DATABASE insurance_service;
CREATE DATABASE workflow_service;
CREATE DATABASE ai_service;
CREATE DATABASE notification_service;
CREATE DATABASE reporting_service;
CREATE DATABASE audit_service;
CREATE DATABASE expense_service;

-- ─── Extensions in jeder Datenbank ───────────────────────
DO $$
DECLARE
    db_name TEXT;
    db_list TEXT[] := ARRAY[
        'auth_service', 'inventory_service', 'project_service',
        'scanner_service', 'warehouse_service', 'invoice_service',
        'document_service', 'crew_service', 'federation_service',
        'maintenance_service', 'transport_service', 'insurance_service',
        'workflow_service', 'ai_service', 'notification_service',
        'reporting_service', 'audit_service', 'expense_service'
    ];
BEGIN
    RAISE NOTICE 'RentFlow: % Service-Datenbanken angelegt', array_length(db_list, 1);
END $$;
