-- Create crew_members table
CREATE TABLE crew_members (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  first_name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  phone VARCHAR(20),
  role VARCHAR(50) NOT NULL CHECK (role IN ('technician', 'rigger', 'driver', 'sound_engineer', 'lighting_tech', 'freelancer')),
  status VARCHAR(50) NOT NULL CHECK (status IN ('active', 'inactive', 'on_leave')) DEFAULT 'active',
  hourly_rate DECIMAL(10, 2),
  daily_rate DECIMAL(10, 2),
  preferred_vehicle_id UUID,
  emergency_contact VARCHAR(255),
  notes TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_crew_members_tenant_id ON crew_members(tenant_id);
CREATE INDEX idx_crew_members_email ON crew_members(email);
CREATE UNIQUE INDEX idx_crew_members_email_tenant ON crew_members(tenant_id, email);

-- Create qualifications table
CREATE TABLE qualifications (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  crew_member_id UUID NOT NULL REFERENCES crew_members(id) ON DELETE CASCADE,
  qualification_type VARCHAR(50) NOT NULL CHECK (qualification_type IN ('IPAF', 'rigger_cert', 'sound_engineer', 'lighting_tech', 'electrical_cert', 'driver_license_c', 'driver_license_ce', 'first_aid', 'forklift')),
  issued_at TIMESTAMP WITH TIME ZONE NOT NULL,
  expires_at TIMESTAMP WITH TIME ZONE,
  certificate_number VARCHAR(255),
  issuing_authority VARCHAR(255),
  status VARCHAR(50) NOT NULL CHECK (status IN ('valid', 'expired', 'pending_renewal')) DEFAULT 'valid',
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_qualifications_tenant_id ON qualifications(tenant_id);
CREATE INDEX idx_qualifications_crew_member_id ON qualifications(crew_member_id);
CREATE INDEX idx_qualifications_type ON qualifications(qualification_type);

-- Create crew_assignments table
CREATE TABLE crew_assignments (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  crew_member_id UUID NOT NULL REFERENCES crew_members(id) ON DELETE CASCADE,
  project_id UUID,
  tour_id UUID,
  role VARCHAR(255) NOT NULL,
  start_date TIMESTAMP WITH TIME ZONE NOT NULL,
  end_date TIMESTAMP WITH TIME ZONE NOT NULL,
  status VARCHAR(50) NOT NULL CHECK (status IN ('planned', 'confirmed', 'active', 'completed', 'cancelled')) DEFAULT 'planned',
  notes TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_crew_assignments_tenant_id ON crew_assignments(tenant_id);
CREATE INDEX idx_crew_assignments_crew_member_id ON crew_assignments(crew_member_id);
CREATE INDEX idx_crew_assignments_project_id ON crew_assignments(project_id);
CREATE INDEX idx_crew_assignments_tour_id ON crew_assignments(tour_id);
CREATE INDEX idx_crew_assignments_status ON crew_assignments(status);
CREATE INDEX idx_crew_assignments_start_date ON crew_assignments(start_date);

-- Create time_records table
CREATE TABLE time_records (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL,
  crew_member_id UUID NOT NULL REFERENCES crew_members(id) ON DELETE CASCADE,
  assignment_id UUID REFERENCES crew_assignments(id) ON DELETE SET NULL,
  date DATE NOT NULL,
  start_time TIMESTAMP WITH TIME ZONE NOT NULL,
  end_time TIMESTAMP WITH TIME ZONE,
  break_minutes INTEGER DEFAULT 0,
  overtime_minutes INTEGER DEFAULT 0,
  status VARCHAR(50) NOT NULL CHECK (status IN ('running', 'completed', 'approved')) DEFAULT 'running',
  notes TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_time_records_tenant_id ON time_records(tenant_id);
CREATE INDEX idx_time_records_crew_member_id ON time_records(crew_member_id);
CREATE INDEX idx_time_records_assignment_id ON time_records(assignment_id);
CREATE INDEX idx_time_records_date ON time_records(date);
CREATE INDEX idx_time_records_status ON time_records(status);
