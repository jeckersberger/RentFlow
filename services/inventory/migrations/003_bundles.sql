-- Equipment Bundles/Kits: Predefined sets (e.g., "PA System" = 2x Speaker + 1x Amp + cables)
CREATE TABLE IF NOT EXISTS equipment_bundles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS equipment_bundle_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bundle_id UUID NOT NULL REFERENCES equipment_bundles(id) ON DELETE CASCADE,
    equipment_type_id UUID NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    is_optional BOOLEAN DEFAULT FALSE,
    notes TEXT,
    sort_order INT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_bundles_tenant ON equipment_bundles(tenant_id, is_active);
CREATE INDEX IF NOT EXISTS idx_bundle_items ON equipment_bundle_items(bundle_id);
