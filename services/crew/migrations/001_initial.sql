CREATE TABLE IF NOT EXISTS crew_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(100),
    role VARCHAR(100) DEFAULT 'technician',
    hourly_rate INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS crew_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    crew_member_id UUID NOT NULL REFERENCES crew_members(id) ON DELETE CASCADE,
    project_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    role VARCHAR(100),
    start_date DATE,
    end_date DATE,
    hours_planned DECIMAL(8,2) DEFAULT 0,
    hours_actual DECIMAL(8,2) DEFAULT 0,
    status VARCHAR(50) DEFAULT 'planned',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS crew_qualifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    crew_member_id UUID NOT NULL REFERENCES crew_members(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    issued_at DATE,
    expires_at DATE,
    certificate_number VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_crew_members_tenant ON crew_members(tenant_id);
CREATE INDEX IF NOT EXISTS idx_crew_assignments_tenant ON crew_assignments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_crew_assignments_member ON crew_assignments(crew_member_id);
CREATE INDEX IF NOT EXISTS idx_crew_assignments_project ON crew_assignments(project_id);
CREATE INDEX IF NOT EXISTS idx_crew_qualifications_member ON crew_qualifications(crew_member_id);
