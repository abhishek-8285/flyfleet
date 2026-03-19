ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_methods text[];
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_password_change timestamp with time zone;
ALTER TABLE users ADD COLUMN IF NOT EXISTS permissions text[];
