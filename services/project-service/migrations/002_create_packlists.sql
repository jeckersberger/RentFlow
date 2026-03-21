-- Create packlists and packlist_items tables
CREATE TABLE IF NOT EXISTS projects.packlists (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL REFERENCES projects.projects(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT packlists_project_idx UNIQUE (project_id, id),
    CONSTRAINT packlists_tenant_idx UNIQUE (tenant_id, id)
);

CREATE INDEX IF NOT EXISTS packlists_tenant_id_idx ON projects.packlists(tenant_id);
CREATE INDEX IF NOT EXISTS packlists_project_id_idx ON projects.packlists(project_id);
CREATE INDEX IF NOT EXISTS packlists_status_idx ON projects.packlists(status);

CREATE TABLE IF NOT EXISTS projects.packlist_items (
    id VARCHAR(255) PRIMARY KEY,
    packlist_id VARCHAR(255) NOT NULL REFERENCES projects.packlists(id) ON DELETE CASCADE,
    equipment_id VARCHAR(255) NOT NULL,
    equipment_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    quantity_packed INTEGER NOT NULL DEFAULT 0,
    quantity_returned INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS packlist_items_packlist_id_idx ON projects.packlist_items(packlist_id);
CREATE INDEX IF NOT EXISTS packlist_items_equipment_id_idx ON projects.packlist_items(equipment_id);
CREATE INDEX IF NOT EXISTS packlist_items_status_idx ON projects.packlist_items(status);
