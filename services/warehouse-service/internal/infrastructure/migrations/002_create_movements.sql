CREATE TABLE IF NOT EXISTS movements (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    equipment_id VARCHAR(255) NOT NULL,
    from_location_id VARCHAR(255),
    to_location_id VARCHAR(255) NOT NULL,
    movement_type VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    reason TEXT,
    user_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255),
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,

    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (from_location_id) REFERENCES locations(id) ON DELETE SET NULL,
    FOREIGN KEY (to_location_id) REFERENCES locations(id) ON DELETE RESTRICT
);

CREATE INDEX idx_movements_tenant_id ON movements(tenant_id);
CREATE INDEX idx_movements_equipment_id ON movements(tenant_id, equipment_id);
CREATE INDEX idx_movements_timestamp ON movements(timestamp DESC);
CREATE INDEX idx_movements_movement_type ON movements(movement_type);
CREATE INDEX idx_movements_from_location ON movements(from_location_id);
CREATE INDEX idx_movements_to_location ON movements(to_location_id);
