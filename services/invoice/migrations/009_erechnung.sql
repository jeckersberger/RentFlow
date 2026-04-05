-- E-Rechnung Readiness (ZUGFeRD/XRechnung Basis)
-- Vorbereitung für B2B E-Rechnungspflicht (BMF/EU)

-- Leitweg-ID für öffentliche Auftraggeber (XRechnung)
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS leitweg_id VARCHAR(50);

-- E-Rechnung Format und Status
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS e_invoice_format VARCHAR(20); -- 'zugferd', 'xrechnung', NULL
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS e_invoice_xml TEXT; -- Generated XML embedded in PDF
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS e_invoice_sent_at TIMESTAMPTZ;

-- Kunden-Referenz (Bestellnummer des Kunden, §14 Abs. 4 Nr. 4 UStG)
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS customer_reference VARCHAR(100);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS purchase_order_number VARCHAR(100);
