-- Add RFID tag support for UHF PDA scanners (e.g. CF-H906)
ALTER TABLE inventory.equipment ADD COLUMN IF NOT EXISTS rfid_tag VARCHAR(64);
CREATE INDEX IF NOT EXISTS idx_equipment_rfid ON inventory.equipment(tenant_id, rfid_tag) WHERE rfid_tag IS NOT NULL;
