-- Equipment Types (Katalog/Typen)
CREATE TABLE IF NOT EXISTS inventory.equipment_types (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id VARCHAR(255) NOT NULL,
    manufacturer VARCHAR(255),
    model VARCHAR(255),
    sku_prefix VARCHAR(50),
    rental_price_day DECIMAL(10,2) DEFAULT 0,
    rental_price_week DECIMAL(10,2) DEFAULT 0,
    replacement_value DECIMAL(12,2) DEFAULT 0,
    weight DECIMAL(10,3) DEFAULT 0,
    dim_length DECIMAL(10,2) DEFAULT 0,
    dim_width DECIMAL(10,2) DEFAULT 0,
    dim_height DECIMAL(10,2) DEFAULT 0,
    dim_unit VARCHAR(5) DEFAULT 'cm',
    image_url TEXT,
    tags TEXT[] DEFAULT '{}',
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_equip_types_tenant ON inventory.equipment_types(tenant_id);
CREATE INDEX IF NOT EXISTS idx_equip_types_category ON inventory.equipment_types(tenant_id, category_id);

-- Add equipment_type_id to existing equipment table
ALTER TABLE inventory.equipment ADD COLUMN IF NOT EXISTS equipment_type_id VARCHAR(255);
-- Auto-serial counter per type
ALTER TABLE inventory.equipment ADD COLUMN IF NOT EXISTS item_number INTEGER DEFAULT 0;
