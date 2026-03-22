-- Warehouses table - top-level warehouse facility
CREATE TABLE IF NOT EXISTS warehouses (
    id VARCHAR(100) PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    address TEXT,
    city VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    total_capacity INT DEFAULT 0,
    current_occupancy INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_tenant_code UNIQUE (tenant_id, code)
);

-- Zones table - areas within warehouse (e.g., "Receiving", "Shipping", "Cold Storage")
CREATE TABLE IF NOT EXISTS zones (
    id VARCHAR(100) PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL,
    warehouse_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    zone_type VARCHAR(50),
    description TEXT,
    temperature_min DECIMAL(5, 2),
    temperature_max DECIMAL(5, 2),
    humidity_min DECIMAL(5, 2),
    humidity_max DECIMAL(5, 2),
    capacity INT DEFAULT 0,
    current_occupancy INT DEFAULT 0,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_zone_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE,
    CONSTRAINT unique_tenant_warehouse_zone UNIQUE (tenant_id, warehouse_id, code)
);

-- Racks table - storage structures within zones
CREATE TABLE IF NOT EXISTS racks (
    id VARCHAR(100) PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL,
    zone_id VARCHAR(100) NOT NULL,
    warehouse_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    rack_type VARCHAR(50) DEFAULT 'standard',
    aisle VARCHAR(50),
    row_number INT,
    column_number INT,
    capacity INT DEFAULT 0,
    current_occupancy INT DEFAULT 0,
    height DECIMAL(10, 2),
    width DECIMAL(10, 2),
    depth DECIMAL(10, 2),
    weight_capacity DECIMAL(10, 2),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_rack_zone FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE CASCADE,
    CONSTRAINT fk_rack_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE,
    CONSTRAINT unique_tenant_zone_rack UNIQUE (tenant_id, zone_id, code)
);

-- Bays table - horizontal divisions on a rack (e.g., levels/shelves)
CREATE TABLE IF NOT EXISTS bays (
    id VARCHAR(100) PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL,
    rack_id VARCHAR(100) NOT NULL,
    zone_id VARCHAR(100) NOT NULL,
    warehouse_id VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    bay_number INT,
    bay_level INT,
    capacity INT DEFAULT 0,
    current_occupancy INT DEFAULT 0,
    weight_capacity DECIMAL(10, 2),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_bay_rack FOREIGN KEY (rack_id) REFERENCES racks(id) ON DELETE CASCADE,
    CONSTRAINT fk_bay_zone FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE CASCADE,
    CONSTRAINT fk_bay_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE,
    CONSTRAINT unique_tenant_rack_bay UNIQUE (tenant_id, rack_id, code)
);

-- Stock Locations table - individual storage slots
CREATE TABLE IF NOT EXISTS stock_locations (
    id VARCHAR(100) PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL,
    warehouse_id VARCHAR(100) NOT NULL,
    zone_id VARCHAR(100) NOT NULL,
    rack_id VARCHAR(100) NOT NULL,
    bay_id VARCHAR(100) NOT NULL,
    location_code VARCHAR(100) NOT NULL,
    barcode VARCHAR(255),
    qr_code_label TEXT,
    location_type VARCHAR(50) DEFAULT 'slot',
    capacity INT DEFAULT 1,
    current_occupancy INT DEFAULT 0,
    weight_capacity DECIMAL(10, 2),
    current_weight DECIMAL(10, 2) DEFAULT 0,
    equipment_id VARCHAR(100),
    is_available BOOLEAN DEFAULT true,
    access_level VARCHAR(50) DEFAULT 'standard',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_stock_location_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE,
    CONSTRAINT fk_stock_location_zone FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE CASCADE,
    CONSTRAINT fk_stock_location_rack FOREIGN KEY (rack_id) REFERENCES racks(id) ON DELETE CASCADE,
    CONSTRAINT fk_stock_location_bay FOREIGN KEY (bay_id) REFERENCES bays(id) ON DELETE CASCADE,
    CONSTRAINT unique_tenant_location_code UNIQUE (tenant_id, location_code)
);

-- Movements table - track all physical movements
CREATE TABLE IF NOT EXISTS movements (
    id VARCHAR(100) PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL,
    equipment_id VARCHAR(100) NOT NULL,
    from_location_id VARCHAR(100),
    to_location_id VARCHAR(100) NOT NULL,
    movement_type VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    reason TEXT,
    user_id VARCHAR(100) NOT NULL,
    project_id VARCHAR(100),
    reference_number VARCHAR(100),
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_movement_from_location FOREIGN KEY (from_location_id) REFERENCES stock_locations(id),
    CONSTRAINT fk_movement_to_location FOREIGN KEY (to_location_id) REFERENCES stock_locations(id)
);

-- Inventory checks/Inventur table - tracking stock counts
CREATE TABLE IF NOT EXISTS inventory_checks (
    id VARCHAR(100) PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL,
    warehouse_id VARCHAR(100),
    zone_id VARCHAR(100),
    check_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'planned',
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    completed_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_inventory_check_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE SET NULL,
    CONSTRAINT fk_inventory_check_zone FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE SET NULL
);

-- Inventory check items - individual items in a count
CREATE TABLE IF NOT EXISTS inventory_check_items (
    id VARCHAR(100) PRIMARY KEY,
    tenant_id VARCHAR(100) NOT NULL,
    check_id VARCHAR(100) NOT NULL,
    equipment_id VARCHAR(100) NOT NULL,
    location_id VARCHAR(100),
    expected_count INT DEFAULT 0,
    actual_count INT DEFAULT 0,
    variance INT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending',
    scanned_at TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_inventory_check_item_check FOREIGN KEY (check_id) REFERENCES inventory_checks(id) ON DELETE CASCADE,
    CONSTRAINT fk_inventory_check_item_location FOREIGN KEY (location_id) REFERENCES stock_locations(id) ON DELETE SET NULL
);

-- Create indexes for performance
CREATE INDEX idx_warehouses_tenant ON warehouses(tenant_id);
CREATE INDEX idx_zones_warehouse ON zones(warehouse_id);
CREATE INDEX idx_zones_tenant_warehouse ON zones(tenant_id, warehouse_id);
CREATE INDEX idx_racks_zone ON racks(zone_id);
CREATE INDEX idx_racks_tenant_warehouse ON racks(tenant_id, warehouse_id);
CREATE INDEX idx_bays_rack ON bays(rack_id);
CREATE INDEX idx_bays_zone ON bays(zone_id);
CREATE INDEX idx_bays_tenant_warehouse ON bays(tenant_id, warehouse_id);
CREATE INDEX idx_stock_locations_warehouse ON stock_locations(warehouse_id);
CREATE INDEX idx_stock_locations_zone ON stock_locations(zone_id);
CREATE INDEX idx_stock_locations_rack ON stock_locations(rack_id);
CREATE INDEX idx_stock_locations_bay ON stock_locations(bay_id);
CREATE INDEX idx_stock_locations_tenant_code ON stock_locations(tenant_id, location_code);
CREATE INDEX idx_stock_locations_equipment ON stock_locations(equipment_id);
CREATE INDEX idx_movements_tenant ON movements(tenant_id);
CREATE INDEX idx_movements_equipment ON movements(equipment_id);
CREATE INDEX idx_movements_from_location ON movements(from_location_id);
CREATE INDEX idx_movements_to_location ON movements(to_location_id);
CREATE INDEX idx_movements_timestamp ON movements(timestamp);
CREATE INDEX idx_inventory_checks_tenant ON inventory_checks(tenant_id);
CREATE INDEX idx_inventory_checks_warehouse ON inventory_checks(warehouse_id);
CREATE INDEX idx_inventory_checks_status ON inventory_checks(status);
CREATE INDEX idx_inventory_check_items_check ON inventory_check_items(check_id);
CREATE INDEX idx_inventory_check_items_equipment ON inventory_check_items(equipment_id);
