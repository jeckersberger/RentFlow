-- Packing Manifests: Load sequence and stowage positions for transport orders
CREATE TABLE IF NOT EXISTS packing_manifests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    transport_order_id UUID NOT NULL REFERENCES transport_orders(id) ON DELETE CASCADE,
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS packing_manifest_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    manifest_id UUID NOT NULL REFERENCES packing_manifests(id) ON DELETE CASCADE,
    transport_item_id UUID NOT NULL REFERENCES transport_items(id),
    load_sequence INT NOT NULL DEFAULT 0,
    stowage_position VARCHAR(50), -- 'top-left', 'bottom-center', etc.
    is_fragile BOOLEAN DEFAULT FALSE,
    notes TEXT
);

CREATE INDEX IF NOT EXISTS idx_packing_manifests_tenant ON packing_manifests(tenant_id);
CREATE INDEX IF NOT EXISTS idx_packing_manifest_items ON packing_manifest_items(manifest_id);

-- Proof of Delivery: Signature + photos captured by driver on delivery
CREATE TABLE IF NOT EXISTS delivery_proofs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    transport_order_id UUID NOT NULL REFERENCES transport_orders(id),
    delivered_by UUID,
    delivered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    signature_svg TEXT,
    photo_urls TEXT[], -- Array of uploaded photo URLs
    recipient_name VARCHAR(255),
    damage_notes TEXT,
    gps_lat DECIMAL(10, 7),
    gps_lng DECIMAL(10, 7),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_delivery_proofs_tenant ON delivery_proofs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_delivery_proofs_order ON delivery_proofs(transport_order_id);
