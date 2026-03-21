-- Quotes table (Angebot - Estimates)
CREATE TABLE invoice.quotes (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    quote_number VARCHAR(50) NOT NULL UNIQUE,
    project_id VARCHAR(255),
    client_name VARCHAR(255) NOT NULL,
    client_address_street VARCHAR(255),
    client_address_city VARCHAR(255),
    client_address_postcode VARCHAR(20),
    client_address_country VARCHAR(100),
    client_email VARCHAR(255),
    sub_total DECIMAL(12, 2) NOT NULL DEFAULT 0,
    tax_rate DECIMAL(5, 2) NOT NULL DEFAULT 19,
    tax_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'EUR',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    valid_until TIMESTAMP NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_quote_status CHECK (status IN ('draft', 'sent', 'accepted', 'rejected', 'expired')),
    CONSTRAINT chk_quote_tax_rate CHECK (tax_rate IN (0, 7, 19)),
    CONSTRAINT chk_quote_total_positive CHECK (total > 0 OR status = 'draft')
);

-- Quote items table
CREATE TABLE invoice.quote_items (
    id VARCHAR(255) PRIMARY KEY,
    quote_id VARCHAR(255) NOT NULL REFERENCES invoice.quotes(id) ON DELETE CASCADE,
    description VARCHAR(500) NOT NULL,
    quantity DECIMAL(10, 3) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    unit_price DECIMAL(12, 2) NOT NULL,
    total_price DECIMAL(12, 2) NOT NULL,
    tax_rate DECIMAL(5, 2) NOT NULL,
    equipment_id VARCHAR(255),

    CONSTRAINT chk_quote_quantity_positive CHECK (quantity > 0)
);

-- Indexes
CREATE INDEX idx_quotes_tenant_id ON invoice.quotes(tenant_id);
CREATE INDEX idx_quotes_quote_number ON invoice.quotes(quote_number);
CREATE INDEX idx_quotes_status ON invoice.quotes(status);
CREATE INDEX idx_quotes_valid_until ON invoice.quotes(valid_until);
CREATE INDEX idx_quotes_client_email ON invoice.quotes(client_email);
CREATE INDEX idx_quote_items_quote_id ON invoice.quote_items(quote_id);

-- Trigger to update updated_at on quote update
CREATE TRIGGER quote_updated_at_trigger
BEFORE UPDATE ON invoice.quotes
FOR EACH ROW
EXECUTE FUNCTION invoice.update_invoice_updated_at();
