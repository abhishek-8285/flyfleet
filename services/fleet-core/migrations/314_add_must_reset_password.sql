-- Add must_reset_password flag to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS must_reset_password BOOLEAN DEFAULT FALSE;

-- Add index to status for better filtering
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
