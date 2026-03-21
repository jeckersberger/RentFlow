-- Dunning table (Mahnwesen - Payment reminders)
CREATE TABLE invoice.dunning (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    invoice_id VARCHAR(255) NOT NULL REFERENCES invoice.invoices(id) ON DELETE CASCADE,
    invoice_number VARCHAR(50) NOT NULL,
    level INTEGER NOT NULL,
    sent_at TIMESTAMP,
    due_date TIMESTAMP NOT NULL,
    fee DECIMAL(12, 2) NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_dunning_level CHECK (level BETWEEN 1 AND 3),
    CONSTRAINT chk_dunning_fee_positive CHECK (fee >= 0)
);

-- Dunning levels:
-- 1 = Zahlungserinnerung (Payment reminder - no fee)
-- 2 = 1. Mahnung (First notice - usually €5-10)
-- 3 = 2. Mahnung (Second notice - usually €10-20)

-- Indexes
CREATE INDEX idx_dunning_tenant_id ON invoice.dunning(tenant_id);
CREATE INDEX idx_dunning_invoice_id ON invoice.dunning(invoice_id);
CREATE INDEX idx_dunning_level ON invoice.dunning(level);
CREATE INDEX idx_dunning_sent_at ON invoice.dunning(sent_at);
CREATE INDEX idx_dunning_due_date ON invoice.dunning(due_date);

-- Trigger to update updated_at on dunning update
CREATE TRIGGER dunning_updated_at_trigger
BEFORE UPDATE ON invoice.dunning
FOR EACH ROW
EXECUTE FUNCTION invoice.update_invoice_updated_at();
