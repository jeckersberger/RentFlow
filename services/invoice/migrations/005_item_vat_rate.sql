-- Add vat_rate column to invoice_items for per-item tax rates.
ALTER TABLE invoice_items ADD COLUMN IF NOT EXISTS vat_rate BIGINT DEFAULT 0;
