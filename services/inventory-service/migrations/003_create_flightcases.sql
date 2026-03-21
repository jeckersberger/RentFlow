CREATE TABLE inventory.flightcases (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    barcode VARCHAR(255) NOT NULL,
    weight DECIMAL(10,3),
    location_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by_user_id UUID NOT NULL,
    UNIQUE(tenant_id, barcode)
);

CREATE TABLE inventory.flightcase_items (
    flightcase_id VARCHAR(255) NOT NULL REFERENCES inventory.flightcases(id) ON DELETE CASCADE,
    equipment_id VARCHAR(255) NOT NULL,
    quantity INT DEFAULT 1,
    added_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (flightcase_id, equipment_id)
);

CREATE INDEX idx_fc_tenant ON inventory.flightcases(tenant_id);
CREATE INDEX idx_fc_barcode ON inventory.flightcases(tenant_id, barcode);
CREATE INDEX idx_fc_location ON inventory.flightcases(location_id);
CREATE INDEX idx_fc_items_equipment ON inventory.flightcase_items(equipment_id);
