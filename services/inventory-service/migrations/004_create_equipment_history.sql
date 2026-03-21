CREATE TABLE IF NOT EXISTS inventory.equipment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    equipment_id VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    changed_by UUID,
    old_value JSONB,
    new_value JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_equipment_history_equipment ON inventory.equipment_history(equipment_id);
CREATE INDEX idx_equipment_history_tenant ON inventory.equipment_history(tenant_id);
CREATE INDEX idx_equipment_history_created ON inventory.equipment_history(created_at DESC);
