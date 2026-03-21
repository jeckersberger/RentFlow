CREATE TABLE documents.templates (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL,
    type VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    variables TEXT[] DEFAULT '{}',
    is_default BOOLEAN DEFAULT FALSE,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_tenant_id FOREIGN KEY (tenant_id) REFERENCES public.tenants(id)
);

CREATE INDEX idx_templates_tenant_id ON documents.templates(tenant_id);
CREATE INDEX idx_templates_is_default ON documents.templates(tenant_id, is_default);
CREATE INDEX idx_templates_created_at ON documents.templates(created_at DESC);
