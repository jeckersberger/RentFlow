-- Idempotency inbox for offline scanner event deduplication.
-- Each scan event from the scanner app includes a client-generated event_id (UUID).
-- On replay/sync, duplicate events are detected and return the original result.

CREATE TABLE IF NOT EXISTS processed_events (
    event_id   UUID PRIMARY KEY,
    tenant_id  UUID NOT NULL,
    action     VARCHAR(50) NOT NULL,
    barcode    VARCHAR(255),
    result     JSONB,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_processed_events_tenant
    ON processed_events (tenant_id, event_id);

-- Automatically clean up old processed events (older than 30 days)
-- to prevent unbounded growth. Run via pg_cron or application-level cleanup.
