-- Seed default module configuration for existing tenants
INSERT INTO auth.tenant_config (tenant_id, config_key, config_value)
SELECT id, 'modules.enabled', '["warehouse", "projects", "finance"]'::jsonb
FROM auth.tenants
ON CONFLICT DO NOTHING;
