CREATE TABLE expenses.expense_categories (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    skr03_code VARCHAR(50),
    skr04_code VARCHAR(50),
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_tenant_id FOREIGN KEY (tenant_id) REFERENCES public.tenants(id)
);

CREATE INDEX idx_categories_tenant_id ON expenses.expense_categories(tenant_id);
CREATE INDEX idx_categories_is_default ON expenses.expense_categories(tenant_id, is_default);
