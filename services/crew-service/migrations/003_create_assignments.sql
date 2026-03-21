CREATE TABLE crew.assignments (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    crew_member_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    role VARCHAR(255),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_tenant_id FOREIGN KEY (tenant_id) REFERENCES public.tenants(id),
    CONSTRAINT fk_crew_member_id FOREIGN KEY (crew_member_id) REFERENCES crew.crew_members(id)
);

CREATE INDEX idx_assignments_tenant_id ON crew.assignments(tenant_id);
CREATE INDEX idx_assignments_crew_member ON crew.assignments(crew_member_id);
CREATE INDEX idx_assignments_project ON crew.assignments(tenant_id, project_id);
CREATE INDEX idx_assignments_status ON crew.assignments(status);
CREATE INDEX idx_assignments_dates ON crew.assignments(start_date, end_date);
