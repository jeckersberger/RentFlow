CREATE TABLE IF NOT EXISTS kpi_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    active_projects INT DEFAULT 0,
    equipment_out_count INT DEFAULT 0,
    total_equipment INT DEFAULT 0,
    open_invoices_amount BIGINT DEFAULT 0,
    overdue_invoices_amount BIGINT DEFAULT 0,
    monthly_revenue BIGINT DEFAULT 0,
    customer_count INT DEFAULT 0,
    utilization_pct INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, snapshot_date)
);
CREATE INDEX IF NOT EXISTS idx_kpi_tenant_date ON kpi_snapshots(tenant_id, snapshot_date);
