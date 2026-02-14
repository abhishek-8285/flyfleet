-- Migration: Add DR Storage Pools and Backup Verification tables
-- Date: 2026-02-14

-- Storage Pools table
CREATE TABLE IF NOT EXISTS storage_pools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    total_bytes BIGINT NOT NULL DEFAULT 0,
    used_bytes BIGINT NOT NULL DEFAULT 0,
    available_bytes BIGINT NOT NULL DEFAULT 0,
    location VARCHAR(255),
    backup_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Backup Verifications table
CREATE TABLE IF NOT EXISTS backup_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    backup_id UUID NOT NULL UNIQUE REFERENCES backups(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    checksum VARCHAR(128),
    verified_at TIMESTAMP WITH TIME ZONE,
    verified_by UUID REFERENCES users(id),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_storage_pools_status ON storage_pools(status);
CREATE INDEX IF NOT EXISTS idx_backup_verifications_backup_id ON backup_verifications(backup_id);
CREATE INDEX IF NOT EXISTS idx_backup_verifications_status ON backup_verifications(status);

-- Insert demo storage pools data
INSERT INTO storage_pools (id, name, status, total_bytes, used_bytes, available_bytes, location, backup_count, created_at, updated_at) VALUES
    ('a1b2c3d4-0001-0001-0001-000000000001', 'Primary Storage', 'ACTIVE', 1073741824000, 375809638400, 697932185600, 'us-east-1', 24, NOW() - INTERVAL '90 days', NOW()),
    ('a1b2c3d4-0002-0002-0002-000000000002', 'DR Storage', 'ACTIVE', 2147483648000, 858993459200, 1288490188800, 'us-west-2', 48, NOW() - INTERVAL '180 days', NOW()),
    ('a1b2c3d4-0003-0003-0003-000000000003', 'Archive Storage', 'WARNING', 5368709120000, 4509715660800, 858993459200, 'eu-central-1', 120, NOW() - INTERVAL '365 days', NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;
