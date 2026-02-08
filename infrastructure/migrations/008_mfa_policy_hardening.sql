-- MFA Policy Hardening & Versioning
-- Created: 2026-02-07
-- Description: Adds enterprise-grade policy controls, overrides, and history snapshoting

-- 1. Extend MFA Policies
ALTER TABLE mfa_policies 
ADD COLUMN IF NOT EXISTS grace_period_days INTEGER DEFAULT 7,
ADD COLUMN IF NOT EXISTS remember_device_days INTEGER DEFAULT 30,
ADD COLUMN IF NOT EXISTS enforce_at_login BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS admin_only_mode BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS remember_device_allowed BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS mfa_challenge_timeout INTEGER DEFAULT 60,
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;

-- 2. Organization Overrides
CREATE TABLE IF NOT EXISTS mfa_org_overrides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT TRUE,
    allowed_methods JSONB DEFAULT '["totp", "sms"]',
    max_attempts INTEGER DEFAULT 5,
    lockout_minutes INTEGER DEFAULT 30,
    remember_device_days INTEGER DEFAULT 30,
    grace_period_days INTEGER DEFAULT 7,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(org_id)
);

-- 3. Policy History (Snapshots)
CREATE TABLE IF NOT EXISTS mfa_policy_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID, -- Can be NULL for global policy history if preferred, or linked to mfa_policies.id
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE, -- Scoped to org
    snapshot JSONB NOT NULL,
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    change_reason TEXT,
    version INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Extend Exceptions
ALTER TABLE mfa_exceptions 
ADD COLUMN IF NOT EXISTS start_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- 5. Audit Logging for Policy Changes
-- Existing audit_logs table will be used via the repository layer.

-- Indexes
CREATE INDEX IF NOT EXISTS idx_mfa_policy_history_org_id ON mfa_policy_history(org_id);
CREATE INDEX IF NOT EXISTS idx_mfa_org_overrides_org_id ON mfa_org_overrides(org_id);
