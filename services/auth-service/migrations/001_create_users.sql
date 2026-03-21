-- Create auth schema
CREATE SCHEMA IF NOT EXISTS auth;

-- Create users table
CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    roles TEXT DEFAULT '',
    status VARCHAR(20) DEFAULT 'active',
    failed_logins INT DEFAULT 0,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

-- Create indexes
CREATE INDEX idx_users_tenant ON auth.users(tenant_id);
CREATE INDEX idx_users_email ON auth.users(tenant_id, email);
CREATE INDEX idx_users_status ON auth.users(status);
