-- Warehouse Transfer Orders: Move equipment between warehouses
CREATE TABLE IF NOT EXISTS warehouse_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    from_warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    to_warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    status VARCHAR(50) NOT NULL DEFAULT 'planned', -- planned, in_transit, completed, cancelled
    notes TEXT,
    scheduled_at TIMESTAMPTZ,
    shipped_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT different_warehouses CHECK (from_warehouse_id != to_warehouse_id)
);

CREATE TABLE IF NOT EXISTS warehouse_transfer_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id UUID NOT NULL REFERENCES warehouse_transfers(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    notes TEXT
);

CREATE INDEX IF NOT EXISTS idx_warehouse_transfers_tenant ON warehouse_transfers(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_warehouse_transfer_items ON warehouse_transfer_items(transfer_id);
