CREATE TABLE IF NOT EXISTS scan_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    device_id VARCHAR(255),
    barcode VARCHAR(255),
    rfid_tag VARCHAR(255),
    equipment_id UUID,
    action VARCHAR(50) NOT NULL,
    project_id UUID,
    location_id UUID,
    condition_rating INTEGER,
    condition_notes TEXT,
    gps_lat DECIMAL(10,7),
    gps_lng DECIMAL(10,7),
    timestamp TIMESTAMPTZ NOT NULL,
    synced_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS scanner_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    device_name VARCHAR(255) DEFAULT '',
    device_type VARCHAR(50) DEFAULT 'cf-h906',
    fcm_token TEXT,
    ring_requested BOOLEAN DEFAULT FALSE,
    last_seen TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(device_id, tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_scan_events_tenant ON scan_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_scan_events_equipment ON scan_events(equipment_id);
CREATE INDEX IF NOT EXISTS idx_scan_events_project ON scan_events(project_id);
CREATE INDEX IF NOT EXISTS idx_scan_events_timestamp ON scan_events(tenant_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_scanner_devices_tenant ON scanner_devices(tenant_id);
