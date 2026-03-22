-- Number sequences table for GoBD-compliant sequential numbering
-- This ensures no gaps in invoice/quote numbers (German tax law requirement)
CREATE TABLE invoice.number_sequences (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    sequence_type VARCHAR(50) NOT NULL, -- 'invoice', 'quote'
    next_value INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Each tenant has one sequence per type
    UNIQUE(tenant_id, sequence_type),

    CONSTRAINT chk_sequence_type CHECK (sequence_type IN ('invoice', 'quote')),
    CONSTRAINT chk_next_value_positive CHECK (next_value > 0)
);

-- Indexes
CREATE INDEX idx_sequences_tenant_sequence ON invoice.number_sequences(tenant_id, sequence_type);

-- Trigger to update updated_at on sequence update
CREATE TRIGGER sequence_updated_at_trigger
BEFORE UPDATE ON invoice.number_sequences
FOR EACH ROW
EXECUTE FUNCTION invoice.update_invoice_updated_at();

-- Function to verify no gaps exist in invoice numbers
CREATE OR REPLACE FUNCTION invoice.verify_no_gaps(p_tenant_id VARCHAR, p_year INTEGER)
RETURNS TABLE(gap_found BOOLEAN, last_number INTEGER, expected_next INTEGER) AS $$
DECLARE
    v_last_number INTEGER;
    v_expected_next INTEGER;
BEGIN
    SELECT MAX(CAST(SUBSTRING(invoice_number FROM '-(\d+)$') AS INTEGER))
    INTO v_last_number
    FROM invoice.invoices
    WHERE tenant_id = p_tenant_id
    AND EXTRACT(YEAR FROM issue_date) = p_year;

    v_expected_next := COALESCE(v_last_number, 0) + 1;

    RETURN QUERY SELECT
        (COALESCE(v_last_number, 0) > 0 AND v_last_number != (
            SELECT COUNT(*) FROM invoice.invoices
            WHERE tenant_id = p_tenant_id
            AND EXTRACT(YEAR FROM issue_date) = p_year
        )),
        v_last_number,
        v_expected_next;
END;
$$ LANGUAGE plpgsql;
