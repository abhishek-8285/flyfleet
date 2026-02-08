-- Create trip_settings table for administrative configuration
CREATE TABLE IF NOT EXISTS trip_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID UNIQUE NOT NULL,
    retention_period INT DEFAULT 36,
    policies JSONB DEFAULT '{}',
    updated_at TIMESTAMP DEFAULT NOW()
);
