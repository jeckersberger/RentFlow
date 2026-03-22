CREATE TABLE IF NOT EXISTS projects.customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    address_street VARCHAR(255),
    address_city VARCHAR(100),
    address_postcode VARCHAR(20),
    address_country VARCHAR(3) DEFAULT 'DE',
    tax_id VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_customers_tenant ON projects.customers(tenant_id);
CREATE INDEX idx_customers_name ON projects.customers(tenant_id, name);
