-- Barcode Range Allocation: Auto-generate sequential barcodes
CREATE TABLE IF NOT EXISTS barcode_ranges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    pattern VARCHAR(100) NOT NULL DEFAULT 'EQP-{YEAR}-{SEQ:05d}',
    next_sequence INT NOT NULL DEFAULT 1,
    last_allocated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_barcode_ranges_tenant ON barcode_ranges(tenant_id, pattern);
