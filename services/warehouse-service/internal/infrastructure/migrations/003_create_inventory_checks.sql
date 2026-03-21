CREATE TABLE IF NOT EXISTS inventory_checks (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    location_id VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    items JSONB NOT NULL DEFAULT '[]'::jsonb,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE SET NULL
);

CREATE INDEX idx_inventory_checks_tenant_id ON inventory_checks(tenant_id);
CREATE INDEX idx_inventory_checks_status ON inventory_checks(status);
CREATE INDEX idx_inventory_checks_location_id ON inventory_checks(location_id);
CREATE INDEX idx_inventory_checks_created_at ON inventory_checks(created_at DESC);
