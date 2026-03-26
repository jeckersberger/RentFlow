-- Add username field to users (login with username OR email)
ALTER TABLE auth.users ADD COLUMN IF NOT EXISTS username VARCHAR(50);

-- Make email nullable (optional)
ALTER TABLE auth.users ALTER COLUMN email DROP NOT NULL;

-- Unique username per tenant
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_tenant ON auth.users(tenant_id, username) WHERE username IS NOT NULL AND username != '';

-- Auto-generate username from email for existing users
UPDATE auth.users SET username = SPLIT_PART(email, '@', 1) WHERE username IS NULL OR username = '';
