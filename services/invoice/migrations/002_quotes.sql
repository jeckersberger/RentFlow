-- Invoice Service: Migration 002 — Quotes (Angebote)
-- Adds quote/offer support with versioning and status tracking

CREATE TABLE IF NOT EXISTS quotes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL,
    quote_number    TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft',
    customer_name   TEXT NOT NULL,
    customer_email  TEXT,
    customer_address TEXT,
    project_id      UUID,
    subject         TEXT,
    intro_text      TEXT DEFAULT 'Gerne unterbreiten wir Ihnen folgendes Angebot:',
    outro_text      TEXT DEFAULT 'Wir freuen uns auf Ihre Rueckmeldung.',
    quote_date      DATE NOT NULL DEFAULT CURRENT_DATE,
    valid_until     DATE,
    vat_rate        BIGINT NOT NULL DEFAULT 1900,
    kleinunternehmer BOOLEAN NOT NULL DEFAULT FALSE,
    total_net       BIGINT NOT NULL DEFAULT 0,
    total_vat       BIGINT NOT NULL DEFAULT 0,
    total_gross     BIGINT NOT NULL DEFAULT 0,
    payment_terms_days INTEGER DEFAULT 14,
    discount_pct    BIGINT DEFAULT 0,
    notes           TEXT,
    converted_invoice_id UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, quote_number)
);

CREATE INDEX IF NOT EXISTS idx_quotes_tenant ON quotes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_quotes_status ON quotes(tenant_id, status);

CREATE TABLE IF NOT EXISTS quote_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    quote_id    UUID NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    quantity    BIGINT NOT NULL DEFAULT 1,
    unit        TEXT NOT NULL DEFAULT 'Stueck',
    unit_price  BIGINT NOT NULL DEFAULT 0,
    discount_pct BIGINT DEFAULT 0,
    position    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_quote_items_quote ON quote_items(quote_id);

-- Add AN (Angebot) number sequence prefix
INSERT INTO number_sequences (tenant_id, prefix, year, last_number)
SELECT tenant_id, 'AN', 2026, 0 FROM number_sequences WHERE prefix = 'RE'
ON CONFLICT (tenant_id, prefix, year) DO NOTHING;

-- Trigger for updated_at
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_quotes_updated_at') THEN
        CREATE TRIGGER trg_quotes_updated_at
            BEFORE UPDATE ON quotes
            FOR EACH ROW EXECUTE FUNCTION update_invoice_updated_at();
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_quote_items_updated_at') THEN
        CREATE TRIGGER trg_quote_items_updated_at
            BEFORE UPDATE ON quote_items
            FOR EACH ROW EXECUTE FUNCTION update_invoice_updated_at();
    END IF;
END $$;
