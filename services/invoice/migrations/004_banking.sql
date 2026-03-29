-- Invoice Service: Migration 004 — Bank Import & Auto-Matching

CREATE TABLE IF NOT EXISTS bank_transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL,
    booking_date        DATE NOT NULL,
    value_date          DATE,
    amount              BIGINT NOT NULL,
    currency            TEXT DEFAULT 'EUR',
    reference           TEXT,
    counterparty_name   TEXT,
    counterparty_iban   TEXT,
    matched_invoice_id  UUID REFERENCES invoices(id),
    match_confidence    TEXT DEFAULT 'none',
    import_source       TEXT,
    import_batch_id     UUID,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bank_tx_tenant ON bank_transactions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bank_tx_matched ON bank_transactions(matched_invoice_id);
CREATE INDEX IF NOT EXISTS idx_bank_tx_batch ON bank_transactions(import_batch_id);
