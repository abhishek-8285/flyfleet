-- Migration 107: Add missing columns to users table
-- These columns are referenced by IAM auth handlers but were never created

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS avatar_url TEXT;

ALTER TABLE users
ADD COLUMN IF NOT EXISTS mfa_backup_codes JSONB DEFAULT '[]'::jsonb;

COMMENT ON COLUMN users.avatar_url IS 'URL or path to user avatar/profile image';
COMMENT ON COLUMN users.mfa_backup_codes IS 'Encrypted MFA backup codes';
