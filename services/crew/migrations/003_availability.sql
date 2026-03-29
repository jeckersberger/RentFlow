CREATE TABLE IF NOT EXISTS crew_availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    crew_member_id UUID NOT NULL REFERENCES crew_members(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    status TEXT NOT NULL DEFAULT 'available',
    note TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, crew_member_id, date)
);
CREATE INDEX IF NOT EXISTS idx_availability_member ON crew_availability(crew_member_id);
CREATE INDEX IF NOT EXISTS idx_availability_date ON crew_availability(tenant_id, date);
