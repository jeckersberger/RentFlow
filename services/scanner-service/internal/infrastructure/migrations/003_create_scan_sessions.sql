CREATE TABLE IF NOT EXISTS scan_sessions (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    context VARCHAR(50) NOT NULL,
    project_id VARCHAR(255),
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    device_type VARCHAR(50) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    total_scans INT DEFAULT 0,
    successful_scans INT DEFAULT 0,
    failed_scans INT DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX idx_scan_sessions_tenant_id ON scan_sessions(tenant_id);
CREATE INDEX idx_scan_sessions_user_id ON scan_sessions(user_id);
CREATE INDEX idx_scan_sessions_context ON scan_sessions(context);
CREATE INDEX idx_scan_sessions_active ON scan_sessions(tenant_id, ended_at) WHERE ended_at IS NULL;
CREATE INDEX idx_scan_sessions_created_at ON scan_sessions(created_at DESC);
