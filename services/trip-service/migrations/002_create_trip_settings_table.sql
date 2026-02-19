-- Create trip_settings table
CREATE TABLE IF NOT EXISTS trip_settings (
    org_id VARCHAR(36) PRIMARY KEY,
    settings JSONB NOT NULL DEFAULT '{}',
    version VARCHAR(20) DEFAULT '1.0',
    last_modified TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    modified_by VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Note: In this environment, we use VARCHAR(36) for UUIDs to match the trips table.
-- The org_id is the primary key as settings are global per organization.
