CREATE TABLE IF NOT EXISTS projects.contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'company',
    company_name VARCHAR(255) DEFAULT '',
    first_name VARCHAR(100) DEFAULT '',
    last_name VARCHAR(100) DEFAULT '',
    email VARCHAR(255) DEFAULT '',
    phone VARCHAR(50) DEFAULT '',
    mobile VARCHAR(50) DEFAULT '',
    website VARCHAR(255) DEFAULT '',
    street VARCHAR(255) DEFAULT '',
    house_number VARCHAR(20) DEFAULT '',
    zip VARCHAR(20) DEFAULT '',
    city VARCHAR(100) DEFAULT '',
    country VARCHAR(100) DEFAULT 'Deutschland',
    vat_id VARCHAR(50) DEFAULT '',
    notes TEXT DEFAULT '',
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID
);
CREATE INDEX IF NOT EXISTS idx_contacts_tenant ON projects.contacts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_contacts_type ON projects.contacts(tenant_id, type);
CREATE INDEX IF NOT EXISTS idx_contacts_name ON projects.contacts(tenant_id, company_name, last_name);
