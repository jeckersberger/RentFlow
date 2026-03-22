CREATE SCHEMA IF NOT EXISTS expenses;

CREATE TABLE expenses.expenses (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    vendor VARCHAR(255) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'EUR',
    tax_rate DECIMAL(5,2) DEFAULT 19.00,
    tax_amount DECIMAL(12,2),
    net_amount DECIMAL(12,2),
    category_code VARCHAR(100),
    date DATE NOT NULL,
    payment_method VARCHAR(50),
    receipt_ref VARCHAR(500),
    ocr_data JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'draft',
    project_id VARCHAR(255),
    approved_by VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_tenant_id FOREIGN KEY (tenant_id) REFERENCES public.tenants(id)
);

CREATE INDEX idx_expenses_tenant_id ON expenses.expenses(tenant_id);
CREATE INDEX idx_expenses_date ON expenses.expenses(date DESC);
CREATE INDEX idx_expenses_status ON expenses.expenses(status);
CREATE INDEX idx_expenses_category ON expenses.expenses(tenant_id, category_code);
CREATE INDEX idx_expenses_project ON expenses.expenses(project_id);
