-- Outbox Pattern: Reliable side-effect delivery (email, PDF, DATEV export)
-- Events are written in the same transaction as the business operation,
-- then processed asynchronously by a dispatcher.

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL, -- 'email.invoice_sent', 'pdf.invoice_generated', 'datev.export'
    aggregate_type VARCHAR(50) NOT NULL, -- 'invoice', 'quote', 'dunning'
    aggregate_id UUID NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_pending ON outbox_events(status, next_retry_at)
    WHERE status IN ('pending', 'failed');
CREATE INDEX IF NOT EXISTS idx_outbox_tenant ON outbox_events(tenant_id, event_type);
