-- Migration 007: Add per-item tax and Kleinunternehmerregelung (§19 UStG)
--
-- Adds per-item tax_rate and tax_amount to invoice_items.
-- Adds Kleinunternehmer fields to invoices.
-- Creates VAT classes table for tenant-specific tax configuration.

-- Add tax_rate and tax_amount to invoice items
ALTER TABLE invoice.invoice_items ADD COLUMN IF NOT EXISTS tax_rate DECIMAL(5,2) NOT NULL DEFAULT 19.0;
ALTER TABLE invoice.invoice_items ADD COLUMN IF NOT EXISTS tax_amount DECIMAL(12,2) NOT NULL DEFAULT 0;

-- Add kleinunternehmer fields to invoices
ALTER TABLE invoice.invoices ADD COLUMN IF NOT EXISTS is_kleinunternehmer BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE invoice.invoices ADD COLUMN IF NOT EXISTS kleinunternehmer_text TEXT NOT NULL DEFAULT '';

-- VAT classes table
CREATE TABLE IF NOT EXISTS invoice.vat_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    rate DECIMAL(5,2) NOT NULL,
    description TEXT DEFAULT '',
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Seed default VAT classes for all existing tenants
DO $$
DECLARE
    tid UUID;
BEGIN
    FOR tid IN SELECT DISTINCT tenant_id FROM invoice.invoices
    UNION
    SELECT id FROM auth.tenants
    LOOP
        INSERT INTO invoice.vat_classes (tenant_id, name, rate, description, is_default)
        VALUES
            (tid, 'Standardsatz', 19.0, 'Regulärer Mehrwertsteuersatz', true),
            (tid, 'Ermäßigter Satz', 7.0, 'Ermäßigter Mehrwertsteuersatz', false),
            (tid, 'Steuerfrei', 0.0, 'Keine Mehrwertsteuer', false)
        ON CONFLICT DO NOTHING;
    END LOOP;
END $$;
