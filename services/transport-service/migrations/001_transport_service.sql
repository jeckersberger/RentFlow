-- Transport Service Schema

CREATE TABLE IF NOT EXISTS vehicles (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    license_plate VARCHAR(50) NOT NULL UNIQUE,
    capacity_kg DECIMAL(10, 2) NOT NULL,
    capacity_m3 DECIMAL(10, 2) NOT NULL,
    vehicle_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    dguv_last_check TIMESTAMP,
    dguv_next_check TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_vehicles_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE IF NOT EXISTS tours (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255),
    vehicle_id VARCHAR(255) NOT NULL,
    driver_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'planned',
    departure_at TIMESTAMP,
    arrival_at TIMESTAMP,
    km_start DECIMAL(10, 2),
    km_end DECIMAL(10, 2),
    total_cost DECIMAL(12, 2),
    fuel_cost DECIMAL(12, 2),
    delivery_note_number VARCHAR(255),
    delivery_note_generated_at TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_tours_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    CONSTRAINT fk_tours_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles(id)
);

CREATE TABLE IF NOT EXISTS tour_equipment (
    id VARCHAR(255) PRIMARY KEY,
    tour_id VARCHAR(255) NOT NULL,
    equipment_id VARCHAR(255) NOT NULL,
    weight_kg DECIMAL(10, 2) NOT NULL,
    volume_m3 DECIMAL(10, 2) NOT NULL,
    loaded_at TIMESTAMP,
    unloaded_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_tour_equipment_tour FOREIGN KEY (tour_id) REFERENCES tours(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS driver_logs (
    id VARCHAR(255) PRIMARY KEY,
    tour_id VARCHAR(255) NOT NULL,
    driver_id VARCHAR(255) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    break_minutes INTEGER DEFAULT 0,
    km_driven DECIMAL(10, 2),
    activity_type VARCHAR(50),
    rest_minutes INTEGER DEFAULT 0,
    location_start VARCHAR(255),
    location_end VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_driver_logs_tour FOREIGN KEY (tour_id) REFERENCES tours(id) ON DELETE CASCADE
);

CREATE INDEX idx_vehicles_tenant ON vehicles(tenant_id);
CREATE INDEX idx_vehicles_status ON vehicles(status);
CREATE INDEX idx_vehicles_license_plate ON vehicles(license_plate);
CREATE INDEX idx_tours_tenant ON tours(tenant_id);
CREATE INDEX idx_tours_vehicle ON tours(vehicle_id);
CREATE INDEX idx_tours_status ON tours(status);
CREATE INDEX idx_tours_departure ON tours(departure_at);
CREATE INDEX idx_tour_equipment_tour ON tour_equipment(tour_id);
CREATE INDEX idx_driver_logs_tour ON driver_logs(tour_id);
CREATE INDEX idx_driver_logs_driver ON driver_logs(driver_id);
