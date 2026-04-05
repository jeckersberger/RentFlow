-- Prevent double-booking of equipment for overlapping periods within a tenant.
-- Uses an exclusion constraint with daterange overlap operator (&&).
-- Requires btree_gist extension for combining UUID equality with range overlap.

CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Add exclusion constraint: same equipment + same tenant cannot have overlapping periods.
-- Only applies when both allocated_from and allocated_until are set.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'no_overlapping_equipment_allocation'
    ) THEN
        ALTER TABLE project_equipment
            ADD CONSTRAINT no_overlapping_equipment_allocation
            EXCLUDE USING gist (
                equipment_id WITH =,
                tstzrange(allocated_from, allocated_until) WITH &&
            )
            WHERE (allocated_from IS NOT NULL AND allocated_until IS NOT NULL);
    END IF;
END $$;
