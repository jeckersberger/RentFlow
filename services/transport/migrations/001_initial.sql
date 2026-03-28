CREATE TABLE IF NOT EXISTS vehicles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    license_plate VARCHAR(50),
    type VARCHAR(50) DEFAULT 'van',
    capacity_kg INTEGER,
    capacity_description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transport_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    project_id UUID,
    vehicle_id UUID REFERENCES vehicles(id),
    driver_id UUID,
    type VARCHAR(50) DEFAULT 'delivery',
    status VARCHAR(50) DEFAULT 'planned',
    pickup_address TEXT,
    delivery_address TEXT,
    scheduled_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    notes TEXT,
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transport_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES transport_orders(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL,
    quantity INTEGER DEFAULT 1,
    notes TEXT
);

CREATE INDEX IF NOT EXISTS idx_vehicles_tenant ON vehicles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_transport_orders_tenant ON transport_orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_transport_orders_project ON transport_orders(project_id);
CREATE INDEX IF NOT EXISTS idx_transport_items_order ON transport_items(order_id);
