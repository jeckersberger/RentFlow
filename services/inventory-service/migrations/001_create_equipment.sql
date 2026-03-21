CREATE SCHEMA IF NOT EXISTS inventory;

CREATE TABLE inventory.equipment (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id VARCHAR(255) NOT NULL,
    sku VARCHAR(100),
    serial_number VARCHAR(100),
    barcode VARCHAR(255) NOT NULL,
    status VARCHAR(30) DEFAULT 'available',
    condition VARCHAR(30) DEFAULT 'good',
    purchase_date DATE,
    purchase_price DECIMAL(12,2),
    rental_price_day DECIMAL(10,2),
    rental_price_week DECIMAL(10,2),
    weight DECIMAL(10,3),
    dim_length DECIMAL(10,2),
    dim_width DECIMAL(10,2),
    dim_height DECIMAL(10,2),
    dim_unit VARCHAR(5) DEFAULT 'cm',
    location_id VARCHAR(255),
    image_refs TEXT[] DEFAULT '{}',
    tags TEXT[] DEFAULT '{}',
    custom_fields JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by_user_id UUID NOT NULL,
    UNIQUE(tenant_id, barcode),
    UNIQUE(tenant_id, serial_number)
);

CREATE INDEX idx_equip_tenant ON inventory.equipment(tenant_id);
CREATE INDEX idx_equip_barcode ON inventory.equipment(tenant_id, barcode);
CREATE INDEX idx_equip_status ON inventory.equipment(tenant_id, status);
CREATE INDEX idx_equip_category ON inventory.equipment(tenant_id, category_id);
CREATE INDEX idx_equip_location ON inventory.equipment(tenant_id, location_id);
CREATE INDEX idx_equip_sku ON inventory.equipment(tenant_id, sku);
CREATE INDEX idx_equip_serial ON inventory.equipment(tenant_id, serial_number);
