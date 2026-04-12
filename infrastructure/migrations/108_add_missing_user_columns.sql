-- Migration 108: Add all missing columns to users table for IAM domain
-- These columns are defined in the User model but missing from the database

ALTER TABLE users
ADD COLUMN IF NOT EXISTS phone TEXT,
ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS mfa_secret TEXT,
ADD COLUMN IF NOT EXISTS email_verified BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS failed_logins INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP,
ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS last_login_ip TEXT,
ADD COLUMN IF NOT EXISTS permission_mask BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS sidebar_collapsed BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS language TEXT DEFAULT 'en',
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

COMMENT ON COLUMN users.phone IS 'User phone number';
COMMENT ON COLUMN users.mfa_enabled IS 'Whether MFA is enabled';
COMMENT ON COLUMN users.mfa_secret IS 'Encrypted MFA TOTP secret';
COMMENT ON COLUMN users.email_verified IS 'Whether email has been verified';
COMMENT ON COLUMN users.failed_logins IS 'Count of consecutive failed login attempts';
COMMENT ON COLUMN users.locked_until IS 'Account lockout expiration time';
COMMENT ON COLUMN users.last_login_at IS 'Timestamp of last successful login';
COMMENT ON COLUMN users.last_login_ip IS 'IP address of last successful login';
COMMENT ON COLUMN users.permission_mask IS 'Bitmask for user permissions';
COMMENT ON COLUMN users.sidebar_collapsed IS 'UI preference for sidebar state';
COMMENT ON COLUMN users.language IS 'User preferred language';
