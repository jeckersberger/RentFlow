-- report_definitions table
CREATE TABLE report_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    report_type VARCHAR(50) NOT NULL CHECK (report_type IN ('revenue', 'utilization', 'dunning', 'inventory', 'maintenance', 'crew_hours')),
    description TEXT,
    parameters JSONB,
    schedule_cron VARCHAR(100),
    email_recipients JSONB,
    format VARCHAR(50) NOT NULL DEFAULT 'pdf' CHECK (format IN ('pdf', 'csv', 'both')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_report_name_per_tenant UNIQUE (tenant_id, name)
);

CREATE INDEX idx_report_definitions_tenant_id ON report_definitions(tenant_id);
CREATE INDEX idx_report_definitions_report_type ON report_definitions(report_type);
CREATE INDEX idx_report_definitions_is_active ON report_definitions(is_active);

-- report_runs table
CREATE TABLE report_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    report_definition_id UUID NOT NULL REFERENCES report_definitions(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL CHECK (status IN ('queued', 'running', 'completed', 'failed')),
    parameters_used JSONB,
    period_start TIMESTAMP WITH TIME ZONE,
    period_end TIMESTAMP WITH TIME ZONE,
    file_path VARCHAR(500),
    file_size BIGINT,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_report_runs_tenant_id ON report_runs(tenant_id);
CREATE INDEX idx_report_runs_report_definition_id ON report_runs(report_definition_id);
CREATE INDEX idx_report_runs_status ON report_runs(status);
CREATE INDEX idx_report_runs_created_at ON report_runs(created_at);

-- kpi_snapshots table
CREATE TABLE kpi_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    kpi_type VARCHAR(50) NOT NULL CHECK (kpi_type IN ('revenue', 'utilization', 'equipment_count', 'active_projects', 'overdue_invoices', 'avg_rental_days')),
    value NUMERIC(15, 2),
    previous_value NUMERIC(15, 2),
    change_percentage NUMERIC(10, 4),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_kpi_snapshots_tenant_id ON kpi_snapshots(tenant_id);
CREATE INDEX idx_kpi_snapshots_kpi_type ON kpi_snapshots(kpi_type);
CREATE INDEX idx_kpi_snapshots_snapshot_date ON kpi_snapshots(snapshot_date);
CREATE UNIQUE INDEX idx_kpi_snapshots_unique_daily ON kpi_snapshots(tenant_id, kpi_type, snapshot_date);
