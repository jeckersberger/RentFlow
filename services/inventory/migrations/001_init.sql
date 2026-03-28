-- Inventory Service: Initial schema
-- Categories, Equipment, Equipment Types, Flightcases, History

-- ============================================================
-- Categories (hierarchical tree)
-- ============================================================
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    icon VARCHAR(50),
    color VARCHAR(7),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, name, parent_id)
);

CREATE INDEX IF NOT EXISTS idx_categories_tenant ON categories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id);

-- ============================================================
-- Equipment Types (templates for bulk-creating equipment)
-- ============================================================
CREATE TABLE IF NOT EXISTS equipment_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    default_rental_price_day INTEGER DEFAULT 0,
    default_rental_price_week INTEGER DEFAULT 0,
    default_replacement_value INTEGER DEFAULT 0,
    specifications JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_equipment_types_tenant ON equipment_types(tenant_id);

-- ============================================================
-- Equipment (the core entity)
-- ============================================================
CREATE TABLE IF NOT EXISTS equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    equipment_type_id UUID REFERENCES equipment_types(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sku VARCHAR(100),
    barcode VARCHAR(255),
    qr_code VARCHAR(255),
    serial_number VARCHAR(255),
    rfid_tag VARCHAR(255),
    status VARCHAR(50) DEFAULT 'available',
    condition VARCHAR(50) DEFAULT 'operational',
    quantity_total INTEGER DEFAULT 1,
    quantity_available INTEGER DEFAULT 1,
    rental_price_day INTEGER DEFAULT 0,
    rental_price_week INTEGER DEFAULT 0,
    replacement_value INTEGER DEFAULT 0,
    weight_grams INTEGER,
    width_mm INTEGER,
    height_mm INTEGER,
    depth_mm INTEGER,
    location_id UUID,
    purchase_date DATE,
    purchase_price INTEGER DEFAULT 0,
    manufacturer VARCHAR(255),
    model VARCHAR(255),
    image_url VARCHAR(500),
    custom_fields JSONB DEFAULT '{}',
    notes TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_equipment_barcode ON equipment(tenant_id, barcode) WHERE barcode IS NOT NULL AND barcode != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_equipment_rfid ON equipment(tenant_id, rfid_tag) WHERE rfid_tag IS NOT NULL AND rfid_tag != '';
CREATE INDEX IF NOT EXISTS idx_equipment_tenant ON equipment(tenant_id);
CREATE INDEX IF NOT EXISTS idx_equipment_category ON equipment(category_id);
CREATE INDEX IF NOT EXISTS idx_equipment_status ON equipment(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_equipment_name_search ON equipment USING gin(to_tsvector('german', name || ' ' || COALESCE(description, '')));

-- ============================================================
-- Flightcases (containers that group equipment)
-- ============================================================
CREATE TABLE IF NOT EXISTS flightcases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    barcode VARCHAR(255),
    qr_code VARCHAR(255),
    description TEXT,
    weight_grams INTEGER,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flightcases_tenant ON flightcases(tenant_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_flightcases_barcode ON flightcases(tenant_id, barcode) WHERE barcode IS NOT NULL AND barcode != '';

CREATE TABLE IF NOT EXISTS flightcase_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flightcase_id UUID NOT NULL REFERENCES flightcases(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    quantity INTEGER DEFAULT 1,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(flightcase_id, equipment_id)
);

-- ============================================================
-- Equipment History (audit trail)
-- ============================================================
CREATE TABLE IF NOT EXISTS equipment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    equipment_id UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    user_id UUID,
    details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_equipment_history_equipment ON equipment_history(equipment_id);
CREATE INDEX IF NOT EXISTS idx_equipment_history_tenant ON equipment_history(tenant_id);

-- ============================================================
-- Auto-update updated_at trigger
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_categories_updated_at ON categories;
CREATE TRIGGER trg_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trg_equipment_updated_at ON equipment;
CREATE TRIGGER trg_equipment_updated_at BEFORE UPDATE ON equipment FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trg_equipment_types_updated_at ON equipment_types;
CREATE TRIGGER trg_equipment_types_updated_at BEFORE UPDATE ON equipment_types FOR EACH ROW EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS trg_flightcases_updated_at ON flightcases;
CREATE TRIGGER trg_flightcases_updated_at BEFORE UPDATE ON flightcases FOR EACH ROW EXECUTE FUNCTION update_updated_at();
