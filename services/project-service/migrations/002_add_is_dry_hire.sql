-- Add is_dry_hire column to projects table
-- Dry Hire = customer picks up equipment themselves (requires signature on checkout)
ALTER TABLE projects.projects ADD COLUMN IF NOT EXISTS is_dry_hire BOOLEAN NOT NULL DEFAULT false;
