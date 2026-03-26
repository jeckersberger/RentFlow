-- Make category_id optional on equipment_types
ALTER TABLE inventory.equipment_types ALTER COLUMN category_id DROP NOT NULL;
ALTER TABLE inventory.equipment_types ALTER COLUMN category_id SET DEFAULT '';
