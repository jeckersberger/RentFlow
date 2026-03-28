CREATE TABLE IF NOT EXISTS insurance_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(255),
    policy_number VARCHAR(100),
    coverage_type VARCHAR(100) DEFAULT 'all_risk',
    coverage_amount INTEGER DEFAULT 0,
    deductible INTEGER DEFAULT 0,
    premium INTEGER DEFAULT 0,
    start_date DATE,
    end_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS insured_equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES insurance_policies(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL,
    insured_value INTEGER DEFAULT 0,
    added_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS insurance_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES insurance_policies(id),
    tenant_id UUID NOT NULL,
    equipment_id UUID,
    claim_number VARCHAR(100),
    description TEXT NOT NULL,
    damage_amount INTEGER DEFAULT 0,
    claim_amount INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'submitted',
    incident_date DATE,
    filed_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_insurance_policies_tenant ON insurance_policies(tenant_id);
CREATE INDEX IF NOT EXISTS idx_insured_equipment_policy ON insured_equipment(policy_id);
CREATE INDEX IF NOT EXISTS idx_insurance_claims_tenant ON insurance_claims(tenant_id);
CREATE INDEX IF NOT EXISTS idx_insurance_claims_policy ON insurance_claims(policy_id);
