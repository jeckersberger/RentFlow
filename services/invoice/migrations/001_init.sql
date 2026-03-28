-- Invoice Service: Initial Schema
-- GoBD-compliant: sequential numbering, hash chain, immutability after finalize

CREATE TABLE IF NOT EXISTS number_sequences (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    prefix      TEXT NOT NULL DEFAULT 'RE',
    year        INT  NOT NULL,
    last_number BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, prefix, year)
);

CREATE TABLE IF NOT EXISTS invoices (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL,
    invoice_number    TEXT NOT NULL,
    invoice_type      TEXT NOT NULL DEFAULT 'invoice',
    status            TEXT NOT NULL DEFAULT 'draft',
    customer_name     TEXT NOT NULL,
    customer_email    TEXT,
    customer_address  TEXT,
    invoice_date      DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date          DATE,
    vat_rate          BIGINT NOT NULL DEFAULT 1900,
    kleinunternehmer  BOOLEAN NOT NULL DEFAULT FALSE,
    total_net         BIGINT NOT NULL DEFAULT 0,
    total_vat         BIGINT NOT NULL DEFAULT 0,
    total_gross       BIGINT NOT NULL DEFAULT 0,
    amount_paid       BIGINT NOT NULL DEFAULT 0,
    notes             TEXT,
    hash              TEXT,
    previous_hash     TEXT,
    finalized_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, invoice_number)
);

CREATE INDEX IF NOT EXISTS idx_invoices_tenant ON invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_invoices_date   ON invoices(tenant_id, invoice_date);

CREATE TABLE IF NOT EXISTS invoice_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    invoice_id  UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    quantity    BIGINT NOT NULL DEFAULT 1,
    unit        TEXT NOT NULL DEFAULT 'Stueck',
    unit_price  BIGINT NOT NULL DEFAULT 0,
    position    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_items_invoice ON invoice_items(invoice_id);

CREATE TABLE IF NOT EXISTS payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    invoice_id      UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount          BIGINT NOT NULL,
    payment_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    payment_method  TEXT NOT NULL DEFAULT 'bank_transfer',
    reference       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_invoice ON payments(invoice_id);

CREATE OR REPLACE FUNCTION update_invoice_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_invoices_updated_at') THEN
        CREATE TRIGGER trg_invoices_updated_at
            BEFORE UPDATE ON invoices
            FOR EACH ROW EXECUTE FUNCTION update_invoice_updated_at();
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_invoice_items_updated_at') THEN
        CREATE TRIGGER trg_invoice_items_updated_at
            BEFORE UPDATE ON invoice_items
            FOR EACH ROW EXECUTE FUNCTION update_invoice_updated_at();
    END IF;
END $$;
