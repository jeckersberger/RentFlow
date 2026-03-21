CREATE SCHEMA IF NOT EXISTS crew;

CREATE TABLE crew.crew_members (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    type VARCHAR(50) NOT NULL,
    skills TEXT[] DEFAULT '{}',
    hourly_rate DECIMAL(10,2),
    daily_rate DECIMAL(10,2),
    status VARCHAR(50) DEFAULT 'active',
    availability_calendar TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_tenant_id FOREIGN KEY (tenant_id) REFERENCES public.tenants(id)
);

CREATE INDEX idx_crew_members_tenant_id ON crew.crew_members(tenant_id);
CREATE INDEX idx_crew_members_email ON crew.crew_members(tenant_id, email);
CREATE INDEX idx_crew_members_created_at ON crew.crew_members(created_at DESC);
