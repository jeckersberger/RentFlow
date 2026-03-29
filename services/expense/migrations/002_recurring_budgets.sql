CREATE TABLE IF NOT EXISTS recurring_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    category_id UUID,
    amount BIGINT NOT NULL,
    frequency TEXT NOT NULL DEFAULT 'monthly',
    next_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_recurring_tenant ON recurring_expenses(tenant_id);

CREATE TABLE IF NOT EXISTS budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    scope_type TEXT NOT NULL DEFAULT 'global',
    scope_id UUID,
    period_type TEXT NOT NULL DEFAULT 'monthly',
    amount BIGINT NOT NULL,
    alert_threshold_pct INT DEFAULT 80,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_budgets_tenant ON budgets(tenant_id);
