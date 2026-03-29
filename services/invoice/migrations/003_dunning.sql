-- Invoice Service: Migration 003 — Dunning (Mahnwesen)

CREATE TABLE IF NOT EXISTS dunning_entries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    invoice_id  UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    level       TEXT NOT NULL DEFAULT 'reminder',
    fee_cents   BIGINT NOT NULL DEFAULT 0,
    sent_at     TIMESTAMPTZ,
    status      TEXT NOT NULL DEFAULT 'pending',
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dunning_invoice ON dunning_entries(invoice_id);
CREATE INDEX IF NOT EXISTS idx_dunning_tenant ON dunning_entries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_dunning_status ON dunning_entries(tenant_id, status);

-- Dunning config per tenant
CREATE TABLE IF NOT EXISTS dunning_config (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL UNIQUE,
    reminder_days   INT NOT NULL DEFAULT 7,
    dunning1_days   INT NOT NULL DEFAULT 14,
    dunning2_days   INT NOT NULL DEFAULT 28,
    reminder_fee    BIGINT NOT NULL DEFAULT 0,
    dunning1_fee    BIGINT NOT NULL DEFAULT 500,
    dunning2_fee    BIGINT NOT NULL DEFAULT 1000,
    auto_send       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
