-- Inventory Service: Migration 002 — Price Rules (Preiskalkulation)

CREATE TABLE IF NOT EXISTS price_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    equipment_id UUID REFERENCES equipment(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE CASCADE,
    base_price_day BIGINT NOT NULL DEFAULT 0,
    base_price_week BIGINT,
    tier2_from_days INT,
    tier2_price_day BIGINT,
    tier3_from_days INT,
    tier3_price_day BIGINT,
    qty_discount_threshold INT,
    qty_discount_pct INT DEFAULT 0,
    season_start DATE,
    season_end DATE,
    season_surcharge_pct INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_price_rules_tenant ON price_rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_price_rules_equipment ON price_rules(equipment_id);
CREATE INDEX IF NOT EXISTS idx_price_rules_category ON price_rules(category_id);
