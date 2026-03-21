-- Maintenance Service Schema

CREATE TABLE IF NOT EXISTS maintenance_records (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    equipment_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'scheduled',
    scheduled_date TIMESTAMP NOT NULL,
    completed_date TIMESTAMP,
    technician VARCHAR(255),
    cost DECIMAL(10, 2),
    notes TEXT,
    certificate_ref VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_maintenance_records_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE IF NOT EXISTS maintenance_schedules (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    equipment_id VARCHAR(255) NOT NULL,
    interval_days INTEGER NOT NULL,
    last_performed TIMESTAMP,
    next_due TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_maintenance_schedules_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT uk_maintenance_schedules_equipment UNIQUE(tenant_id, equipment_id)
);

CREATE INDEX idx_maintenance_records_tenant ON maintenance_records(tenant_id);
CREATE INDEX idx_maintenance_records_equipment ON maintenance_records(equipment_id);
CREATE INDEX idx_maintenance_records_status ON maintenance_records(status);
CREATE INDEX idx_maintenance_records_scheduled_date ON maintenance_records(scheduled_date);
CREATE INDEX idx_maintenance_schedules_tenant ON maintenance_schedules(tenant_id);
CREATE INDEX idx_maintenance_schedules_next_due ON maintenance_schedules(next_due);
