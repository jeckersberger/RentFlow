-- Item-level discount (basis points: 1000 = 10%, 500 = 5%)
ALTER TABLE invoice_items ADD COLUMN IF NOT EXISTS discount_pct BIGINT NOT NULL DEFAULT 0;
ALTER TABLE quote_items ADD COLUMN IF NOT EXISTS discount_pct BIGINT NOT NULL DEFAULT 0;
