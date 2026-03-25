-- Migration: Create booking_requests table for freelancer booking workflow
-- Stand: 2026-03-25

CREATE TABLE IF NOT EXISTS booking_requests (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  assignment_id UUID NOT NULL REFERENCES crew_assignments(id) ON DELETE CASCADE,
  crew_member_id UUID NOT NULL REFERENCES crew_members(id) ON DELETE CASCADE,
  project_id UUID,
  token VARCHAR(255) NOT NULL UNIQUE,
  status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'accepted', 'declined', 'alternative')) DEFAULT 'pending',
  response_message TEXT,
  responded_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_booking_requests_tenant_id ON booking_requests(tenant_id);
CREATE INDEX IF NOT EXISTS idx_booking_requests_token ON booking_requests(token);
CREATE INDEX IF NOT EXISTS idx_booking_requests_crew_member_id ON booking_requests(crew_member_id);
CREATE INDEX IF NOT EXISTS idx_booking_requests_status ON booking_requests(status);
