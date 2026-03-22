-- Create setup_state table for tracking one-time setup
CREATE TABLE IF NOT EXISTS auth.setup_state (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMP WITH TIME ZONE,
    completed_by UUID,
    setup_token VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT single_row CHECK (id = 1)
);

-- Create index for quick lookup
CREATE INDEX idx_setup_state_completed ON auth.setup_state(is_completed);
