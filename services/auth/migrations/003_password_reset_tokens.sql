-- Password reset tokens for forgot-password flow.
-- Stores SHA256 hashes of reset tokens (plain token is sent via email/log).
-- Separate from the legacy password_resets table in 001_init.sql.

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    token_hash VARCHAR(64) NOT NULL, -- SHA256 hash of the token
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pwd_reset_tenant ON password_reset_tokens(tenant_id, token_hash);
