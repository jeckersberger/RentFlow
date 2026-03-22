CREATE TABLE federation_partners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    partner_name VARCHAR(200) NOT NULL,
    partner_endpoint VARCHAR(500) NOT NULL,
    status VARCHAR(30) DEFAULT 'pending',
    trust_level VARCHAR(20) DEFAULT 'basic',
    cert_fingerprint VARCHAR(128),
    cert_expires_at TIMESTAMPTZ,
    shared_categories TEXT[],
    data_policy JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_federation_partners_tenant_id ON federation_partners(tenant_id);
CREATE INDEX idx_federation_partners_status ON federation_partners(status);

CREATE TABLE sub_rental_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    partner_id UUID REFERENCES federation_partners(id) ON DELETE CASCADE,
    direction VARCHAR(10) NOT NULL,
    status VARCHAR(30) DEFAULT 'pending',
    equipment_category VARCHAR(100),
    equipment_description TEXT,
    quantity INT DEFAULT 1,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    daily_rate DECIMAL(10,2),
    total_amount DECIMAL(10,2),
    handover_document_id UUID,
    invoice_id UUID,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sub_rental_requests_tenant_id ON sub_rental_requests(tenant_id);
CREATE INDEX idx_sub_rental_requests_partner_id ON sub_rental_requests(partner_id);
CREATE INDEX idx_sub_rental_requests_status ON sub_rental_requests(status);

CREATE TABLE partner_equipment_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id UUID REFERENCES federation_partners(id) ON DELETE CASCADE,
    category VARCHAR(100),
    item_name VARCHAR(200),
    quantity_available INT,
    daily_rate DECIMAL(10,2),
    last_synced_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_partner_equipment_cache_partner_id ON partner_equipment_cache(partner_id);
CREATE INDEX idx_partner_equipment_cache_category ON partner_equipment_cache(category);

CREATE TABLE federation_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    partner_id UUID REFERENCES federation_partners(id) ON DELETE SET NULL,
    cert_type VARCHAR(20) NOT NULL,
    cert_pem TEXT NOT NULL,
    key_pem_encrypted TEXT,
    fingerprint VARCHAR(128) NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_federation_certificates_tenant_id ON federation_certificates(tenant_id);
CREATE INDEX idx_federation_certificates_partner_id ON federation_certificates(partner_id);
CREATE INDEX idx_federation_certificates_fingerprint ON federation_certificates(fingerprint);
CREATE INDEX idx_federation_certificates_expires_at ON federation_certificates(expires_at);
