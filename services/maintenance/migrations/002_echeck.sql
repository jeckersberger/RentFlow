CREATE TABLE IF NOT EXISTS echeck_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    equipment_id UUID NOT NULL,
    check_date DATE NOT NULL,
    next_check_date DATE,
    result TEXT NOT NULL DEFAULT 'pending',
    performed_by TEXT,
    measuring_device TEXT,
    protection_conductor_resistance DECIMAL(10,4),
    insulation_resistance DECIMAL(10,4),
    leakage_current DECIMAL(10,4),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_echeck_tenant ON echeck_records(tenant_id);
CREATE INDEX IF NOT EXISTS idx_echeck_equipment ON echeck_records(equipment_id);
CREATE INDEX IF NOT EXISTS idx_echeck_next ON echeck_records(next_check_date);
