ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS sequence_number BIGSERIAL;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS checksum TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS prev_checksum TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS aggregate_type TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS aggregate_id UUID;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS event_type TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS event_payload JSONB DEFAULT '{}';
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS service_name TEXT;

CREATE INDEX IF NOT EXISTS idx_audit_sequence ON audit_logs(sequence_number);
CREATE INDEX IF NOT EXISTS idx_audit_aggregate ON audit_logs(aggregate_type, aggregate_id);
