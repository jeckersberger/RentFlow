CREATE TABLE expenses.budgets (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    category_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255),
    period VARCHAR(50) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    spent DECIMAL(12,2) DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_tenant_id FOREIGN KEY (tenant_id) REFERENCES public.tenants(id),
    CONSTRAINT fk_category_id FOREIGN KEY (category_id) REFERENCES expenses.expense_categories(id)
);

CREATE INDEX idx_budgets_tenant_id ON expenses.budgets(tenant_id);
CREATE INDEX idx_budgets_category ON expenses.budgets(category_id);
CREATE INDEX idx_budgets_project ON expenses.budgets(project_id);
CREATE INDEX idx_budgets_period ON expenses.budgets(tenant_id, period);
