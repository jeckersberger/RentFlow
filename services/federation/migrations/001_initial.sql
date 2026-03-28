CREATE TABLE IF NOT EXISTS federation_partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    partner_name VARCHAR(255) NOT NULL,
    partner_url VARCHAR(500),
    api_key_hash VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS shared_listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    equipment_id UUID NOT NULL,
    daily_rate INTEGER DEFAULT 0,
    weekly_rate INTEGER DEFAULT 0,
    available_from DATE,
    available_until DATE,
    is_active BOOLEAN DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS federation_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    partner_id UUID REFERENCES federation_partners(id),
    listing_id UUID REFERENCES shared_listings(id),
    status VARCHAR(50) DEFAULT 'pending',
    start_date DATE,
    end_date DATE,
    total_cost INTEGER DEFAULT 0,
    notes TEXT,
    requested_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_federation_partners_tenant ON federation_partners(tenant_id);
CREATE INDEX IF NOT EXISTS idx_shared_listings_tenant ON shared_listings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_federation_requests_tenant ON federation_requests(tenant_id);
