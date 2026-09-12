CREATE TABLE IF NOT EXISTS auth.signing_keys (
    id TEXT PRIMARY KEY,
    private_key_pem TEXT NOT NULL,
    public_key_pem TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE auth.signing_keys IS 'Persistent RS256 signing key material for auth tokens. The active key survives service restarts.';
