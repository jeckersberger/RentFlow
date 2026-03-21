-- Add storage_location column to packlist_items table
ALTER TABLE projects.packlist_items ADD COLUMN IF NOT EXISTS storage_location TEXT DEFAULT '';

-- Create index for storage_location sorting
CREATE INDEX IF NOT EXISTS idx_packlist_items_storage_location ON projects.packlist_items(storage_location);
