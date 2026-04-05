-- Minimum stock level alerts per equipment type per warehouse
CREATE TABLE IF NOT EXISTS stock_level_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    equipment_type_id UUID NOT NULL,
    warehouse_id UUID,
    min_quantity INT NOT NULL DEFAULT 1,
    alert_email VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    last_triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stock_alerts_tenant ON stock_level_alerts(tenant_id, is_active);
CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_alerts_unique ON stock_level_alerts(tenant_id, equipment_type_id, warehouse_id)
    WHERE warehouse_id IS NOT NULL;
