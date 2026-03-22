-- Audit log table
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    sequence_number BIGSERIAL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    service_name VARCHAR(100) NOT NULL,
    operation VARCHAR(50) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID,
    user_id UUID,
    user_name VARCHAR(200),
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    checksum VARCHAR(64) NOT NULL,
    previous_checksum VARCHAR(64),
    is_pseudonymized BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_log_chain ON audit_log(tenant_id, sequence_number);
CREATE INDEX idx_audit_log_entity ON audit_log(tenant_id, entity_type, entity_id);
CREATE INDEX idx_audit_log_timestamp ON audit_log(tenant_id, timestamp);
CREATE INDEX idx_audit_log_user ON audit_log(tenant_id, user_id);

-- Audit exports table
CREATE TABLE audit_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    export_type VARCHAR(30) NOT NULL,
    date_from TIMESTAMPTZ NOT NULL,
    date_to TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    file_path VARCHAR(500),
    file_size_bytes BIGINT,
    checksum VARCHAR(64),
    requested_by UUID,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_exports_tenant ON audit_exports(tenant_id);
CREATE INDEX idx_audit_exports_status ON audit_exports(status);
