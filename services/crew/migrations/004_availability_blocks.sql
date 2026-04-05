CREATE TABLE IF NOT EXISTS crew_availability_blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    crew_member_id UUID NOT NULL REFERENCES crew_members(id) ON DELETE CASCADE,
    block_type VARCHAR(50) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_crew_availability_blocks_tenant ON crew_availability_blocks(tenant_id, crew_member_id);
CREATE INDEX IF NOT EXISTS idx_crew_availability_blocks_dates ON crew_availability_blocks(tenant_id, start_date, end_date);
