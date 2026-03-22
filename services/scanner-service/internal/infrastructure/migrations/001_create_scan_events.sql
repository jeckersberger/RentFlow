CREATE TABLE IF NOT EXISTS scan_events (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    barcode VARCHAR(255) NOT NULL,
    scan_type VARCHAR(50) NOT NULL,
    equipment_id VARCHAR(255),
    project_id VARCHAR(255),
    location_id VARCHAR(255),
    user_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    device_type VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    notes TEXT,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,

    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX idx_scan_events_tenant_id ON scan_events(tenant_id);
CREATE INDEX idx_scan_events_equipment_id ON scan_events(equipment_id);
CREATE INDEX idx_scan_events_status ON scan_events(status);
CREATE INDEX idx_scan_events_created_at ON scan_events(created_at DESC);
CREATE INDEX idx_scan_events_barcode ON scan_events(tenant_id, barcode);
