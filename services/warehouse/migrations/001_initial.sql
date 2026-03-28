-- Warehouse Service: Initial schema
-- Warehouses, Zones, Racks, Stock Locations, Movements, Inventory Checks

-- ============================================================
-- Warehouses
-- ============================================================
CREATE TABLE IF NOT EXISTS warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50),
    address TEXT,
    capacity_description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- Zones
-- ============================================================
CREATE TABLE IF NOT EXISTS zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50),
    climate_controlled BOOLEAN DEFAULT FALSE,
    max_weight_kg INTEGER,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- Racks
-- ============================================================
CREATE TABLE IF NOT EXISTS racks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id UUID NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50),
    levels INTEGER DEFAULT 4,
    bays_per_level INTEGER DEFAULT 6,
    max_weight_kg INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- Stock Locations
-- ============================================================
CREATE TABLE IF NOT EXISTS stock_locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rack_id UUID REFERENCES racks(id),
    zone_id UUID REFERENCES zones(id),
    tenant_id UUID NOT NULL,
    code VARCHAR(100) NOT NULL,
    barcode VARCHAR(255),
    level INTEGER,
    bay INTEGER,
    max_weight_kg INTEGER,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, code)
);

-- ============================================================
-- Movements
-- ============================================================
CREATE TABLE IF NOT EXISTS movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    equipment_id UUID NOT NULL,
    from_location_id UUID,
    to_location_id UUID,
    quantity INTEGER DEFAULT 1,
    reason VARCHAR(100),
    user_id UUID,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- Inventory Checks
-- ============================================================
CREATE TABLE IF NOT EXISTS inventory_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    zone_id UUID REFERENCES zones(id),
    status VARCHAR(50) DEFAULT 'in_progress',
    started_by UUID,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    expected_count INTEGER DEFAULT 0,
    actual_count INTEGER DEFAULT 0,
    discrepancy_count INTEGER DEFAULT 0,
    notes TEXT
);

-- ============================================================
-- Inventory Check Items
-- ============================================================
CREATE TABLE IF NOT EXISTS inventory_check_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    check_id UUID NOT NULL REFERENCES inventory_checks(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL,
    expected BOOLEAN DEFAULT TRUE,
    found BOOLEAN DEFAULT FALSE,
    scanned_at TIMESTAMPTZ,
    notes TEXT
);

-- ============================================================
-- Indexes
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_zones_warehouse ON zones(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_racks_zone ON racks(zone_id);
CREATE INDEX IF NOT EXISTS idx_stock_locations_tenant ON stock_locations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_movements_tenant ON movements(tenant_id);
CREATE INDEX IF NOT EXISTS idx_movements_equipment ON movements(equipment_id);
CREATE INDEX IF NOT EXISTS idx_inventory_checks_tenant ON inventory_checks(tenant_id);
