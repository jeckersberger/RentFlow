-- Add status workflow columns to packlist_items
-- status tracks the item through: planned -> packed -> loaded -> on_site -> returned
-- damaged is a parallel boolean flag (any status can be marked damaged)

ALTER TABLE packlist_items
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'planned',
    ADD COLUMN IF NOT EXISTS damaged BOOLEAN DEFAULT false;

-- Backfill existing items: derive status from quantity fields
UPDATE packlist_items
SET status = CASE
    WHEN quantity_returned > 0 THEN 'returned'
    WHEN quantity_packed > 0 THEN 'packed'
    ELSE 'planned'
END
WHERE status IS NULL OR status = 'planned';

-- Index for efficient summary aggregation queries
CREATE INDEX IF NOT EXISTS idx_packlist_items_status ON packlist_items(packlist_id, status);
