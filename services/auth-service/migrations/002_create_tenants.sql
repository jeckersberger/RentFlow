-- Create tenants table
CREATE TABLE IF NOT EXISTS auth.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    address_street VARCHAR(255),
    address_city VARCHAR(100),
    address_zip VARCHAR(20),
    address_country VARCHAR(3) DEFAULT 'DE',
    logo VARCHAR(500),
    default_language VARCHAR(5) DEFAULT 'de',
    currency VARCHAR(3) DEFAULT 'EUR',
    tax_rate DECIMAL(5,2) DEFAULT 19.00,
    invoice_prefix VARCHAR(20) DEFAULT 'RF',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_tenants_slug ON auth.tenants(slug);
CREATE INDEX idx_tenants_status ON auth.tenants(status);
