-- Create projects schema and table
CREATE SCHEMA IF NOT EXISTS projects;

CREATE TABLE IF NOT EXISTS projects.projects (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    client_name VARCHAR(255) NOT NULL,
    client_email VARCHAR(255),
    client_phone VARCHAR(20),
    client_street VARCHAR(255),
    client_city VARCHAR(100),
    client_state VARCHAR(100),
    client_postal_code VARCHAR(20),
    client_country VARCHAR(100),
    client_coordinates VARCHAR(255),
    venue_street VARCHAR(255),
    venue_city VARCHAR(100),
    venue_state VARCHAR(100),
    venue_postal_code VARCHAR(20),
    venue_country VARCHAR(100),
    venue_coordinates VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    setup_date TIMESTAMP WITH TIME ZONE,
    teardown_date TIMESTAMP WITH TIME ZONE,
    project_manager VARCHAR(255),
    budget NUMERIC(12, 2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    notes TEXT,
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_by_user_id VARCHAR(255),

    CONSTRAINT projects_tenant_idx UNIQUE (tenant_id, id)
);

CREATE INDEX IF NOT EXISTS projects_tenant_id_idx ON projects.projects(tenant_id);
CREATE INDEX IF NOT EXISTS projects_status_idx ON projects.projects(status);
CREATE INDEX IF NOT EXISTS projects_client_name_idx ON projects.projects USING GIN(
    to_tsvector('english', client_name || ' ' || COALESCE(name, ''))
);
CREATE INDEX IF NOT EXISTS projects_created_at_idx ON projects.projects(created_at DESC);
