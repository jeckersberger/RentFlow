-- Credit Notes table for GoBD-compliant credit note documents
CREATE TABLE IF NOT EXISTS invoice.credit_notes (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    credit_note_number VARCHAR(50) NOT NULL,
    original_invoice_id VARCHAR(255) NOT NULL REFERENCES invoice.invoices(id),
    original_invoice_number VARCHAR(50) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    client_email VARCHAR(255) DEFAULT '',
    items JSONB NOT NULL DEFAULT '[]',
    subtotal DECIMAL(12,2) NOT NULL DEFAULT 0,
    tax_rate DECIMAL(5,4) NOT NULL DEFAULT 0.19,
    tax_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    total DECIMAL(12,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'EUR',
    reason TEXT DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    issued_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_credit_note_status CHECK (status IN ('draft', 'issued')),
    UNIQUE(tenant_id, credit_note_number)
);

CREATE INDEX IF NOT EXISTS idx_credit_notes_tenant ON invoice.credit_notes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_credit_notes_invoice ON invoice.credit_notes(original_invoice_id);
