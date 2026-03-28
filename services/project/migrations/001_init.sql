-- Project Service: Initial schema
-- Projects, Project Equipment, Packlists, Reservations

-- ============================================================
-- Projects
-- ============================================================
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    project_number VARCHAR(50),
    description TEXT,
    status VARCHAR(50) DEFAULT 'draft',
    customer_id UUID,
    contact_name VARCHAR(255),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(100),
    venue_name VARCHAR(255),
    venue_address TEXT,
    venue_lat DECIMAL(10,7),
    venue_lng DECIMAL(10,7),
    start_date DATE,
    end_date DATE,
    setup_date DATE,
    teardown_date DATE,
    color VARCHAR(7) DEFAULT '#3b82f6',
    budget INTEGER DEFAULT 0,
    currency CHAR(3) DEFAULT 'EUR',
    manager_id UUID,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_tenant ON projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_projects_dates ON projects(tenant_id, start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_projects_customer ON projects(customer_id);

-- ============================================================
-- Project Equipment (equipment assigned to projects)
-- ============================================================
CREATE TABLE IF NOT EXISTS project_equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL,
    quantity INTEGER DEFAULT 1,
    allocated_from DATE,
    allocated_until DATE,
    status VARCHAR(50) DEFAULT 'planned',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_equipment_project ON project_equipment(project_id);
CREATE INDEX IF NOT EXISTS idx_project_equipment_equipment ON project_equipment(equipment_id);

-- ============================================================
-- Packlists
-- ============================================================
CREATE TABLE IF NOT EXISTS packlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'draft',
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_packlists_project ON packlists(project_id);
CREATE INDEX IF NOT EXISTS idx_packlists_tenant ON packlists(tenant_id);

-- ============================================================
-- Packlist Items
-- ============================================================
CREATE TABLE IF NOT EXISTS packlist_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    packlist_id UUID NOT NULL REFERENCES packlists(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL,
    quantity_planned INTEGER DEFAULT 1,
    quantity_packed INTEGER DEFAULT 0,
    quantity_returned INTEGER DEFAULT 0,
    packed_by UUID,
    packed_at TIMESTAMPTZ,
    notes TEXT
);

CREATE INDEX IF NOT EXISTS idx_packlist_items_packlist ON packlist_items(packlist_id);

-- ============================================================
-- Reservations
-- ============================================================
CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    equipment_id UUID NOT NULL,
    quantity INTEGER DEFAULT 1,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reservations_tenant ON reservations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_reservations_project ON reservations(project_id);
CREATE INDEX IF NOT EXISTS idx_reservations_equipment ON reservations(equipment_id);
CREATE INDEX IF NOT EXISTS idx_reservations_dates ON reservations(start_date, end_date);

-- ============================================================
-- Auto-update trigger
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_projects_updated_at ON projects;
CREATE TRIGGER trg_projects_updated_at BEFORE UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Fix: Ensure reservations FK has ON DELETE CASCADE
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'reservations_project_id_fkey'
        AND table_name = 'reservations'
    ) THEN
        ALTER TABLE reservations DROP CONSTRAINT reservations_project_id_fkey;
        ALTER TABLE reservations ADD CONSTRAINT reservations_project_id_fkey
            FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
    END IF;
END $$;
