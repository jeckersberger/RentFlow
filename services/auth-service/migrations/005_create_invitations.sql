-- Invitation system for employee onboarding
CREATE TABLE IF NOT EXISTS auth.invitations (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES auth.tenants(id),
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'readonly',
    token VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    invited_by VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    claimed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_invitations_tenant ON auth.invitations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_invitations_token ON auth.invitations(token);
CREATE INDEX IF NOT EXISTS idx_invitations_email ON auth.invitations(tenant_id, email);
