-- Add locked_at column to auth.users table
ALTER TABLE auth.users
ADD COLUMN locked_at TIMESTAMP WITH TIME ZONE NULL;

-- Create index on locked_at for efficient queries on locked accounts
CREATE INDEX idx_users_locked_at ON auth.users(locked_at);
