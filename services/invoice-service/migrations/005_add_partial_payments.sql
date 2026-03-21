ALTER TABLE invoices.invoices ADD COLUMN IF NOT EXISTS paid_amount DECIMAL(12,2) DEFAULT 0;
ALTER TABLE invoices.invoices ADD COLUMN IF NOT EXISTS remaining_amount DECIMAL(12,2) DEFAULT 0;

-- Update existing: set remaining_amount = total for non-paid invoices
UPDATE invoices.invoices SET remaining_amount = total WHERE status != 'paid' AND remaining_amount = 0;

-- Update constraint to include partially_paid
ALTER TABLE invoices.invoices DROP CONSTRAINT IF EXISTS chk_status;
ALTER TABLE invoices.invoices ADD CONSTRAINT chk_status CHECK (status IN ('draft', 'sent', 'overdue', 'paid', 'partially_paid', 'cancelled', 'credited'));
