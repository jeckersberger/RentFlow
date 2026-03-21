CREATE TABLE IF NOT EXISTS locations (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    parent_id VARCHAR(255),
    path VARCHAR(1024) NOT NULL,
    capacity INT NOT NULL DEFAULT 100,
    current_count INT NOT NULL DEFAULT 0,
    sort_order INT DEFAULT 0,
    barcode VARCHAR(255),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES locations(id) ON DELETE SET NULL
);

CREATE INDEX idx_locations_tenant_id ON locations(tenant_id);
CREATE INDEX idx_locations_parent_id ON locations(tenant_id, parent_id);
CREATE INDEX idx_locations_path ON locations(path);
CREATE INDEX idx_locations_type ON locations(tenant_id, type);
