-- §14 UStG Pflichtangaben + GoBD Compliance
-- Steuernummer und USt-IdNr des Rechnungsstellers (Pflichtangaben)
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS issuer_tax_number VARCHAR(20);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS issuer_vat_id VARCHAR(20);

-- Leistungsdatum (kann vom Rechnungsdatum abweichen, §14 Abs. 4 UStG)
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS performance_date DATE;

-- Kunden-USt-IdNr (für Reverse Charge §13b UStG)
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS customer_vat_id VARCHAR(20);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS is_reverse_charge BOOLEAN NOT NULL DEFAULT FALSE;

-- Stornierung/Gutschrift: Referenz auf Originalrechnung (GoBD Audit Trail)
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS original_invoice_id UUID REFERENCES invoices(id);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS cancellation_reason TEXT;
