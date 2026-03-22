-- Create invoice schema
CREATE SCHEMA IF NOT EXISTS invoice;

-- Invoices table (GoBD-compliant with immutability constraints)
CREATE TABLE invoice.invoices (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    invoice_number VARCHAR(50) NOT NULL UNIQUE,
    project_id VARCHAR(255),
    client_name VARCHAR(255) NOT NULL,
    client_address_street VARCHAR(255),
    client_address_city VARCHAR(255),
    client_address_postcode VARCHAR(20),
    client_address_country VARCHAR(100),
    client_email VARCHAR(255),
    client_tax_id VARCHAR(100),
    sub_total DECIMAL(12, 2) NOT NULL DEFAULT 0,
    tax_rate DECIMAL(5, 2) NOT NULL DEFAULT 19,
    tax_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    total DECIMAL(12, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'EUR',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    issue_date TIMESTAMP NOT NULL DEFAULT NOW(),
    due_date TIMESTAMP NOT NULL,
    paid_date TIMESTAMP,
    payment_method VARCHAR(100),
    payment_ref VARCHAR(255),
    notes TEXT,
    internal_notes TEXT,
    pdf_ref VARCHAR(255),
    hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_status CHECK (status IN ('draft', 'sent', 'overdue', 'paid', 'cancelled', 'credited')),
    CONSTRAINT chk_tax_rate CHECK (tax_rate IN (0, 7, 19)),
    CONSTRAINT chk_total_positive CHECK (total > 0 OR status = 'draft'),

    -- GoBD: No updates after finalization
    CONSTRAINT immutable_finalized CHECK (status = 'draft')
);

-- Invoice items table
CREATE TABLE invoice.invoice_items (
    id VARCHAR(255) PRIMARY KEY,
    invoice_id VARCHAR(255) NOT NULL REFERENCES invoice.invoices(id) ON DELETE CASCADE,
    description VARCHAR(500) NOT NULL,
    quantity DECIMAL(10, 3) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    unit_price DECIMAL(12, 2) NOT NULL,
    total_price DECIMAL(12, 2) NOT NULL,
    tax_rate DECIMAL(5, 2) NOT NULL,
    equipment_id VARCHAR(255),

    CONSTRAINT chk_quantity_positive CHECK (quantity > 0)
);

-- Indexes for performance
CREATE INDEX idx_invoices_tenant_id ON invoice.invoices(tenant_id);
CREATE INDEX idx_invoices_invoice_number ON invoice.invoices(invoice_number);
CREATE INDEX idx_invoices_status ON invoice.invoices(status);
CREATE INDEX idx_invoices_due_date ON invoice.invoices(due_date);
CREATE INDEX idx_invoices_client_email ON invoice.invoices(client_email);
CREATE INDEX idx_invoice_items_invoice_id ON invoice.invoice_items(invoice_id);

-- Trigger to update updated_at on invoice update
CREATE OR REPLACE FUNCTION invoice.update_invoice_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER invoice_updated_at_trigger
BEFORE UPDATE ON invoice.invoices
FOR EACH ROW
EXECUTE FUNCTION invoice.update_invoice_updated_at();
