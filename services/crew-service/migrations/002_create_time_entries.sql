CREATE TABLE crew.time_entries (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    crew_member_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255),
    date DATE NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    break_minutes INTEGER DEFAULT 0,
    total_hours DECIMAL(8,2),
    type VARCHAR(50) NOT NULL,
    notes TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_tenant_id FOREIGN KEY (tenant_id) REFERENCES public.tenants(id),
    CONSTRAINT fk_crew_member_id FOREIGN KEY (crew_member_id) REFERENCES crew.crew_members(id)
);

CREATE INDEX idx_time_entries_tenant_id ON crew.time_entries(tenant_id);
CREATE INDEX idx_time_entries_crew_member ON crew.time_entries(tenant_id, crew_member_id);
CREATE INDEX idx_time_entries_date ON crew.time_entries(date DESC);
CREATE INDEX idx_time_entries_status ON crew.time_entries(status);
