-- Compliance-Erweiterungen für DATEV, Reverse Charge und Zahlungsbedingungen

-- USt-IdNr für Reverse Charge (§13b UStG)
ALTER TABLE customers ADD COLUMN IF NOT EXISTS vat_id VARCHAR(20);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS vat_country VARCHAR(2);

-- Persistente DATEV-Debitoren-Nummer (statt Hash-basierter Generierung)
ALTER TABLE customers ADD COLUMN IF NOT EXISTS datev_debitor_number VARCHAR(10);
CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_datev_debitor
    ON customers(tenant_id, datev_debitor_number) WHERE datev_debitor_number IS NOT NULL;

-- Zahlungsbedingungen
ALTER TABLE customers ADD COLUMN IF NOT EXISTS payment_terms_days INT DEFAULT 14;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS credit_limit BIGINT DEFAULT 0;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS is_reverse_charge_eligible BOOLEAN DEFAULT FALSE;

-- Sequence für automatische Debitoren-Nummern (10001, 10002, ...)
CREATE SEQUENCE IF NOT EXISTS datev_debitor_seq START WITH 10001;
