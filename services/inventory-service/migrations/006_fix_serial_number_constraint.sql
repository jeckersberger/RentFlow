-- Drop the unique constraint on serial_number that prevents
-- multiple equipment items with empty serial numbers
ALTER TABLE inventory.equipment DROP CONSTRAINT IF EXISTS equipment_tenant_id_serial_number_key;
