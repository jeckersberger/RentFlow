CREATE SCHEMA IF NOT EXISTS documents;

CREATE TABLE documents.documents (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL,
    type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(100),
    entity_id VARCHAR(255),
    file_ref VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100),
    size BIGINT,
    checksum VARCHAR(500),
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_tenant_id FOREIGN KEY (tenant_id) REFERENCES public.tenants(id)
);

CREATE INDEX idx_documents_tenant_id ON documents.documents(tenant_id);
CREATE INDEX idx_documents_entity ON documents.documents(tenant_id, entity_type, entity_id);
CREATE INDEX idx_documents_created_at ON documents.documents(created_at DESC);
