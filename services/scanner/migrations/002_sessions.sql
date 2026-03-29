CREATE TABLE IF NOT EXISTS scan_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    session_type TEXT NOT NULL DEFAULT 'checkout',
    project_id UUID,
    started_by UUID NOT NULL,
    started_at TIMESTAMPTZ DEFAULT NOW(),
    ended_at TIMESTAMPTZ,
    items_count INT DEFAULT 0,
    signature_data TEXT,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scan_sessions_tenant ON scan_sessions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_scan_sessions_status ON scan_sessions(tenant_id, status);

-- Add session_id column to scan_events for linking events to sessions.
ALTER TABLE scan_events ADD COLUMN IF NOT EXISTS session_id UUID REFERENCES scan_sessions(id);
CREATE INDEX IF NOT EXISTS idx_scan_events_session ON scan_events(session_id);
