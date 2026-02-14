-- Add verified column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS verified BOOLEAN DEFAULT true;
UPDATE users SET verified = true WHERE verified IS NULL;
