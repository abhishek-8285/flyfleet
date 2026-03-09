-- Migration: 082_admin_session_management
-- Description: Add metadata columns to refresh_tokens for session tracking

ALTER TABLE refresh_tokens 
    ADD COLUMN IF NOT EXISTS ip_address VARCHAR(45),
    ADD COLUMN IF NOT EXISTS user_agent TEXT,
    ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP WITH TIME ZONE;

-- Backfill legacy tokens (set last_used_at to created_at)
UPDATE refresh_tokens SET last_used_at = created_at WHERE last_used_at IS NULL;

-- Create an index to quickly list user's active sessions sorted by recency
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_last_used ON refresh_tokens(user_id, last_used_at DESC);

COMMENT ON COLUMN refresh_tokens.ip_address IS 'IP address from which the refresh token was issued';
COMMENT ON COLUMN refresh_tokens.user_agent IS 'User agent string of the device issuing the token';
COMMENT ON COLUMN refresh_tokens.last_used_at IS 'Timestamp of the last time this refresh token was used to generate an access token';
