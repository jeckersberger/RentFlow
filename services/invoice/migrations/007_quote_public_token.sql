-- Customer Portal: Public token for quote viewing and response
ALTER TABLE quotes ADD COLUMN IF NOT EXISTS public_token VARCHAR(64);
ALTER TABLE quotes ADD COLUMN IF NOT EXISTS customer_response VARCHAR(20); -- 'accepted', 'declined', NULL
ALTER TABLE quotes ADD COLUMN IF NOT EXISTS customer_response_message TEXT;
ALTER TABLE quotes ADD COLUMN IF NOT EXISTS customer_response_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_quotes_public_token
    ON quotes(public_token) WHERE public_token IS NOT NULL;
