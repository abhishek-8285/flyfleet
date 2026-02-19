-- Organization Enhancements Migration
-- Added: contact_email, default_theme columns

ALTER TABLE organizations 
ADD COLUMN IF NOT EXISTS contact_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS default_theme VARCHAR(50) DEFAULT 'system',
ADD COLUMN IF NOT EXISTS plan VARCHAR(50) DEFAULT 'enterprise';
