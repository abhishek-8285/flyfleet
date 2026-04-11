-- Add user preference columns to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS sidebar_collapsed BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS language VARCHAR(10) NOT NULL DEFAULT 'en';
