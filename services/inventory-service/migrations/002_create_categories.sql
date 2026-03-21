CREATE TABLE inventory.categories (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id VARCHAR(255) REFERENCES inventory.categories(id),
    icon VARCHAR(50),
    color VARCHAR(50),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by_user_id UUID NOT NULL,
    UNIQUE(tenant_id, name, parent_id)
);

CREATE INDEX idx_cat_tenant ON inventory.categories(tenant_id);
CREATE INDEX idx_cat_parent ON inventory.categories(parent_id);
CREATE INDEX idx_cat_sort ON inventory.categories(tenant_id, sort_order);
