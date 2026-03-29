ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS payload_kg INT DEFAULT 0;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS volume_m3 INT DEFAULT 0;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS fuel_type TEXT DEFAULT 'diesel';
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS fuel_consumption INT DEFAULT 0;

CREATE TABLE IF NOT EXISTS transport_costs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    order_id UUID REFERENCES transport_orders(id) ON DELETE CASCADE,
    cost_type TEXT NOT NULL,
    amount_cents BIGINT NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_transport_costs_order ON transport_costs(order_id);
CREATE INDEX IF NOT EXISTS idx_transport_costs_tenant ON transport_costs(tenant_id);
