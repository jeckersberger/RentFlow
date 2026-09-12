-- Stop storing the one-time setup bootstrap token in plaintext.
-- Pending legacy setups are rotated by auth-service on its next startup.
ALTER TABLE auth.setup_state
    ADD COLUMN IF NOT EXISTS setup_token_hash VARCHAR(64);

ALTER TABLE auth.setup_state
    ALTER COLUMN setup_token DROP NOT NULL;

-- Completed setup no longer needs any bootstrap secret material.
UPDATE auth.setup_state
SET setup_token = NULL,
    setup_token_hash = NULL
WHERE is_completed = TRUE;

COMMENT ON COLUMN auth.setup_state.setup_token_hash IS
    'SHA-256 hex digest of the current one-time setup token; plaintext is never persisted by current auth-service versions.';
