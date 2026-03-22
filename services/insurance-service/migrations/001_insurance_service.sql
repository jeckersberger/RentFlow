-- Create policies table
CREATE TABLE IF NOT EXISTS policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    policy_number VARCHAR(50) NOT NULL UNIQUE,
    policy_type VARCHAR(20) NOT NULL CHECK (policy_type IN ('liability', 'comprehensive', 'transport', 'renter')),
    provider VARCHAR(100) NOT NULL,
    coverage_amount DECIMAL(15, 2) NOT NULL,
    deductible DECIMAL(15, 2) NOT NULL,
    premium_annual DECIMAL(12, 2) NOT NULL,
    premium_monthly DECIMAL(12, 2) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'expired', 'cancelled')) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_policies_tenant_id ON policies(tenant_id);
CREATE INDEX idx_policies_status ON policies(status);
CREATE INDEX idx_policies_end_date ON policies(end_date);

-- Create claims table
CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE RESTRICT,
    claim_number VARCHAR(50) NOT NULL UNIQUE,
    equipment_id UUID,
    project_id UUID,
    incident_date DATE NOT NULL,
    reported_date DATE NOT NULL,
    description TEXT NOT NULL,
    damage_type VARCHAR(20) NOT NULL CHECK (damage_type IN ('theft', 'breakage', 'water', 'fire', 'transport', 'other')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('reported', 'documented', 'submitted', 'in_review', 'approved', 'rejected', 'settled')) DEFAULT 'reported',
    claimed_amount DECIMAL(15, 2) NOT NULL,
    approved_amount DECIMAL(15, 2),
    settled_amount DECIMAL(15, 2),
    adjuster_notes TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_claims_tenant_id ON claims(tenant_id);
CREATE INDEX idx_claims_policy_id ON claims(policy_id);
CREATE INDEX idx_claims_status ON claims(status);
CREATE INDEX idx_claims_incident_date ON claims(incident_date);

-- Create claim_items table
CREATE TABLE IF NOT EXISTS claim_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_id UUID NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL,
    description TEXT NOT NULL,
    replacement_value DECIMAL(15, 2) NOT NULL,
    repair_cost DECIMAL(15, 2),
    photo_urls JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_claim_items_claim_id ON claim_items(claim_id);
CREATE INDEX idx_claim_items_equipment_id ON claim_items(equipment_id);
