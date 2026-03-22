-- Create sessions table for refresh token tracking
CREATE TABLE IF NOT EXISTS auth.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(45),
    user_agent_hash VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    used_at TIMESTAMP WITH TIME ZONE,
    rotation_count INT DEFAULT 0,
    max_rotations INT DEFAULT 5,
    status VARCHAR(20) DEFAULT 'active',
    FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES auth.tenants(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_sessions_user ON auth.sessions(user_id);
CREATE INDEX idx_sessions_tenant ON auth.sessions(tenant_id);
CREATE INDEX idx_sessions_token_hash ON auth.sessions(refresh_token_hash);
CREATE INDEX idx_sessions_expires ON auth.sessions(expires_at);
CREATE INDEX idx_sessions_status ON auth.sessions(status);
