-- Scanner Service: Create all tables
-- Migration 001

CREATE SCHEMA IF NOT EXISTS scanner;

-- Scan events table
CREATE TABLE IF NOT EXISTS scanner.scan_events (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    barcode VARCHAR(255) NOT NULL,
    scan_type VARCHAR(50) NOT NULL,  -- check_in, check_out, inventory, movement, return
    equipment_id VARCHAR(255) DEFAULT '',
    project_id VARCHAR(255),
    location_id VARCHAR(255),
    session_id VARCHAR(255),
    user_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    device_type VARCHAR(50) NOT NULL,  -- usb, handheld, camera, rfid
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    notes TEXT DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, processed, failed, synced
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scan_events_tenant ON scanner.scan_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_scan_events_barcode ON scanner.scan_events(tenant_id, barcode);
CREATE INDEX IF NOT EXISTS idx_scan_events_session ON scanner.scan_events(session_id);
CREATE INDEX IF NOT EXISTS idx_scan_events_created ON scanner.scan_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_scan_events_equipment ON scanner.scan_events(tenant_id, equipment_id);

-- Scan sessions table
CREATE TABLE IF NOT EXISTS scanner.scan_sessions (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    context VARCHAR(50) NOT NULL,  -- check_out, check_in, lager_einraeumen, inventur
    project_id VARCHAR(255),
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE,
    device_type VARCHAR(50) NOT NULL DEFAULT 'handheld',
    device_id VARCHAR(255) NOT NULL DEFAULT '',
    total_scans INTEGER NOT NULL DEFAULT 0,
    successful_scans INTEGER NOT NULL DEFAULT 0,
    failed_scans INTEGER NOT NULL DEFAULT 0,
    signature_data TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scan_sessions_tenant ON scanner.scan_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_scan_sessions_user ON scanner.scan_sessions(tenant_id, user_id);

-- Legacy devices table
CREATE TABLE IF NOT EXISTS scanner.devices (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    serial VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    location VARCHAR(255) DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, serial)
);

CREATE INDEX IF NOT EXISTS idx_devices_tenant ON scanner.devices(tenant_id);

-- Offline queue table
CREATE TABLE IF NOT EXISTS scanner.offline_queue (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    sync_status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, syncing, synced, failed
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    synced_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_offline_queue_tenant ON scanner.offline_queue(tenant_id, sync_status);

-- Scanner devices table (Find My Scanner)
CREATE TABLE IF NOT EXISTS scanner.scanner_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    device_name VARCHAR(255) NOT NULL DEFAULT '',
    device_type VARCHAR(50) NOT NULL DEFAULT 'handheld',
    fcm_token TEXT,
    ring_requested BOOLEAN NOT NULL DEFAULT false,
    last_seen TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_scanner_devices_tenant ON scanner.scanner_devices(tenant_id);
