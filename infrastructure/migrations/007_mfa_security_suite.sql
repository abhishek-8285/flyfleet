-- MFA Security Suite Migration
-- Created: 2026-02-07
-- Description: Advanced MFA management, policies, and reporting

-- MFA Policies
CREATE TABLE IF NOT EXISTS mfa_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    require_mfa BOOLEAN DEFAULT FALSE,
    allowed_methods JSONB DEFAULT '["totp", "sms"]',
    max_attempts INTEGER DEFAULT 5,
    lockout_minutes INTEGER DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(org_id)
);

-- MFA Events (Detailed logging for audit feed)
CREATE TABLE IF NOT EXISTS mfa_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL, -- CHALLENGE_ISSUED, VERIFY_SUCCESS, VERIFY_FAILED, BYPASS_USED
    status VARCHAR(20) NOT NULL, -- success, failure, pending
    method VARCHAR(20), -- totp, sms, recovery_code
    ip_address VARCHAR(45),
    user_agent TEXT,
    location VARCHAR(255),
    details JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- MFA Failures (Tracking for "Top Failure Users/Locations")
CREATE TABLE IF NOT EXISTS mfa_failures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    method VARCHAR(20),
    failure_reason TEXT,
    ip_address VARCHAR(45),
    location VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- MFA Exceptions (Waiving MFA for specific users)
CREATE TABLE IF NOT EXISTS mfa_exceptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    granted_by UUID REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- MFA Lockouts (Specific tracking for account lockouts)
CREATE TABLE IF NOT EXISTS mfa_lockouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    reason VARCHAR(255),
    locked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    unlocked_at TIMESTAMP WITH TIME ZONE,
    unlocked_by UUID REFERENCES users(id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_mfa_events_user_id ON mfa_events(user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_events_org_id ON mfa_events(org_id);
CREATE INDEX IF NOT EXISTS idx_mfa_events_created_at ON mfa_events(created_at);
CREATE INDEX IF NOT EXISTS idx_mfa_failures_location ON mfa_failures(location);
CREATE INDEX IF NOT EXISTS idx_mfa_failures_user_id ON mfa_failures(user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_lockouts_user_id ON mfa_lockouts(user_id);
CREATE INDEX IF NOT EXISTS idx_mfa_exceptions_user_id ON mfa_exceptions(user_id);
