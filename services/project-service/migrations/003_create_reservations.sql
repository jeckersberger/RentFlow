-- Create reservations table
CREATE TABLE IF NOT EXISTS projects.reservations (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL REFERENCES projects.projects(id) ON DELETE CASCADE,
    equipment_id VARCHAR(255) NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT reservations_tenant_idx UNIQUE (tenant_id, id),
    CONSTRAINT end_after_start CHECK (end_date > start_date)
);

CREATE INDEX IF NOT EXISTS reservations_tenant_id_idx ON projects.reservations(tenant_id);
CREATE INDEX IF NOT EXISTS reservations_project_id_idx ON projects.reservations(project_id);
CREATE INDEX IF NOT EXISTS reservations_equipment_id_idx ON projects.reservations(equipment_id);
CREATE INDEX IF NOT EXISTS reservations_status_idx ON projects.reservations(status);

-- Index for conflict checking (overlapping date ranges)
CREATE INDEX IF NOT EXISTS reservations_equipment_dates_idx
    ON projects.reservations(equipment_id, start_date, end_date)
    WHERE status != 'cancelled';
