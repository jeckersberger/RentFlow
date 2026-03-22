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

CREATE TABLE IF NOT EXISTS maintenance_plans (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    equipment_id VARCHAR(255) NOT NULL,
    plan_type VARCHAR(50) NOT NULL,
    interval_days INTEGER,
    interval_hours INTEGER,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    checklist_template_id VARCHAR(255),
    last_executed_at TIMESTAMP,
    next_due_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_maintenance_plans_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE IF NOT EXISTS maintenance_tasks (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    plan_id VARCHAR(255) NOT NULL,
    equipment_id VARCHAR(255) NOT NULL,
    assigned_to VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'planned',
    priority VARCHAR(50) NOT NULL DEFAULT 'medium',
    scheduled_at TIMESTAMP NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    notes TEXT,
    checklist_data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_maintenance_tasks_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_maintenance_tasks_plan FOREIGN KEY (plan_id) REFERENCES maintenance_plans(id)
);

CREATE TABLE IF NOT EXISTS checklists (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    items JSONB NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_checklists_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE IF NOT EXISTS electrical_tests (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    task_id VARCHAR(255),
    equipment_id VARCHAR(255) NOT NULL,
    tester_id VARCHAR(255) NOT NULL,
    test_type VARCHAR(50) NOT NULL,
    test_date TIMESTAMP NOT NULL,
    next_test_date TIMESTAMP NOT NULL,
    result VARCHAR(50) NOT NULL,
    insulation_resistance_mohm DECIMAL(10, 2),
    protective_conductor_resistance_ohm DECIMAL(10, 2),
    leakage_current_ma DECIMAL(10, 2),
    visual_inspection_ok BOOLEAN NOT NULL DEFAULT true,
    functional_test_ok BOOLEAN NOT NULL DEFAULT true,
    test_device_id VARCHAR(255) NOT NULL,
    test_device_name VARCHAR(255) NOT NULL,
    certificate_number VARCHAR(255) NOT NULL,
    notes TEXT,
    izytron_import_id VARCHAR(255),
    raw_xml TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_electrical_tests_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_electrical_tests_task FOREIGN KEY (task_id) REFERENCES maintenance_tasks(id),
    CONSTRAINT uk_certificate_number UNIQUE(tenant_id, certificate_number)
);

CREATE INDEX idx_maintenance_records_tenant ON maintenance_records(tenant_id);
CREATE INDEX idx_maintenance_records_equipment ON maintenance_records(equipment_id);
CREATE INDEX idx_maintenance_records_status ON maintenance_records(status);
CREATE INDEX idx_maintenance_records_scheduled_date ON maintenance_records(scheduled_date);
CREATE INDEX idx_maintenance_schedules_tenant ON maintenance_schedules(tenant_id);
CREATE INDEX idx_maintenance_schedules_next_due ON maintenance_schedules(next_due);
CREATE INDEX idx_maintenance_plans_tenant ON maintenance_plans(tenant_id);
CREATE INDEX idx_maintenance_plans_equipment ON maintenance_plans(equipment_id);
CREATE INDEX idx_maintenance_plans_next_due ON maintenance_plans(next_due_at);
CREATE INDEX idx_maintenance_tasks_tenant ON maintenance_tasks(tenant_id);
CREATE INDEX idx_maintenance_tasks_plan ON maintenance_tasks(plan_id);
CREATE INDEX idx_maintenance_tasks_status ON maintenance_tasks(status);
CREATE INDEX idx_maintenance_tasks_scheduled ON maintenance_tasks(scheduled_at);
CREATE INDEX idx_checklists_tenant ON checklists(tenant_id);
CREATE INDEX idx_electrical_tests_tenant ON electrical_tests(tenant_id);
CREATE INDEX idx_electrical_tests_equipment ON electrical_tests(equipment_id);
CREATE INDEX idx_electrical_tests_test_date ON electrical_tests(test_date);
