-- Migration: Erweiterte Felder fuer KI-Belegauswertung + GoBD-Konformitaet
-- Stand: 2026-03-25

-- Neue Spalten fuer expenses
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS type VARCHAR(50) DEFAULT 'invoice';
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS vendor_address TEXT;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS vendor_vat_id VARCHAR(50);
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS vendor_iban VARCHAR(50);
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS booking_account VARCHAR(20);
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS service_period_from DATE;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS service_period_to DATE;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS due_date DATE;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS discount_percent DECIMAL(5,2) DEFAULT 0;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS discount_days INTEGER DEFAULT 0;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS discount_deadline DATE;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS payment_status VARCHAR(20) DEFAULT 'open';
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS paid_at TIMESTAMP;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS invoice_number VARCHAR(100);
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS receipt_checksum VARCHAR(64);
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS receipt_nas_path VARCHAR(500);
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS ocr_confidence DECIMAL(3,2) DEFAULT 0;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS source VARCHAR(30) DEFAULT 'manual';
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS email_ref VARCHAR(255);
-- Bewirtungsbeleg-Felder
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS entertainment_location VARCHAR(255);
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS entertainment_reason TEXT;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS entertainment_guests TEXT;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS entertainment_tip DECIMAL(12,2) DEFAULT 0;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS entertainment_deductible DECIMAL(12,2) DEFAULT 0;
ALTER TABLE expenses.expenses ADD COLUMN IF NOT EXISTS entertainment_non_deductible DECIMAL(12,2) DEFAULT 0;

-- Indices fuer neue Felder
CREATE INDEX IF NOT EXISTS idx_expenses_payment_status ON expenses.expenses(payment_status);
CREATE INDEX IF NOT EXISTS idx_expenses_due_date ON expenses.expenses(due_date);
CREATE INDEX IF NOT EXISTS idx_expenses_type ON expenses.expenses(type);
CREATE INDEX IF NOT EXISTS idx_expenses_source ON expenses.expenses(source);
CREATE INDEX IF NOT EXISTS idx_expenses_vendor_vat_id ON expenses.expenses(vendor_vat_id);
CREATE INDEX IF NOT EXISTS idx_expenses_invoice_number ON expenses.expenses(tenant_id, invoice_number);

-- Lieferanten-Zuordnung (lernt ueber Zeit)
CREATE TABLE IF NOT EXISTS expenses.vendor_mappings (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    vendor_name VARCHAR(255) NOT NULL,
    vendor_name_normalized VARCHAR(255) NOT NULL,
    vendor_vat_id VARCHAR(50),
    vendor_iban VARCHAR(50),
    vendor_address TEXT,
    default_category_code VARCHAR(100),
    default_booking_account VARCHAR(20),
    default_tax_rate DECIMAL(5,2) DEFAULT 19.00,
    usage_count INTEGER DEFAULT 1,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_vm_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id),
    UNIQUE(tenant_id, vendor_name_normalized)
);

CREATE INDEX IF NOT EXISTS idx_vendor_mappings_tenant ON expenses.vendor_mappings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vendor_mappings_name ON expenses.vendor_mappings(tenant_id, vendor_name_normalized);
