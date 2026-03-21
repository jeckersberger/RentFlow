-- Enable pg_trgm extension if not already enabled
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create trigram index for full-text search
CREATE INDEX IF NOT EXISTS idx_equipment_search_trgm 
ON inventory.equipment USING gin (
    (COALESCE(name, '') || ' ' || COALESCE(description, '') || ' ' || COALESCE(sku, '') || ' ' || COALESCE(serial_number, '') || ' ' || COALESCE(barcode, '')) gin_trgm_ops
);
