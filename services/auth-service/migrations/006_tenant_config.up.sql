CREATE TABLE IF NOT EXISTS auth.tenant_config (
    tenant_id UUID NOT NULL,
    config_key TEXT NOT NULL,
    config_value JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID,
    PRIMARY KEY (tenant_id, config_key)
);

-- Seed default configs for existing tenants
INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'finance.kleinunternehmer', '{"enabled": false, "threshold_previous_year": 22000, "threshold_current_year": 100000}'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;

INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'finance.bank_details', '{"bank_name": "", "iban": "", "bic": "", "account_holder": ""}'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;

INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'finance.payment_terms', '{"default_days": 14, "reminder1_days": 7, "reminder2_days": 14, "reminder3_days": 21}'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;

INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'finance.vat', '{"default_rate": 19.0, "reduced_rate": 7.0, "default_scheme": "standard"}'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;

INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'company.details', '{"trade_register": "", "tax_number": "", "vat_id": "", "managing_director": ""}'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;

INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'email.smtp', '{"host": "", "port": 587, "username": "", "password": "", "sender_email": "", "sender_name": "", "tls": true, "configured": false}'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;

INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'email.signature', '{"html": "", "text": ""}'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;

INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'locale', '{"timezone": "Europe/Berlin", "date_format": "DD.MM.YYYY", "currency": "EUR", "currency_symbol": "€", "language": "de"}'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;

INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'number_sequences', '{"invoice_prefix": "RE-", "invoice_next": 1, "quote_prefix": "AN-", "quote_next": 1, "project_prefix": "P-", "project_next": 1, "credit_note_prefix": "GS-", "credit_note_next": 1}'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;

INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'labels', '{"default_size": "50x25", "include_qr": true, "include_barcode": true, "include_name": true, "include_sku": true, "printer_type": "browser"}'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;
