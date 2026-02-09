-- Security Operations Production Hardening - Expansion
-- Tables for Login Logs, Threat Detection, and MFA Dashboard

-- 1. Login Events (Strict match for requested fields)
CREATE TABLE IF NOT EXISTS login_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    email VARCHAR(255),
    ip_address VARCHAR(45),
    device TEXT,
    browser TEXT,
    location VARCHAR(255),
    status VARCHAR(20) CHECK (status IN ('SUCCESS', 'FAILED', 'CHALLENGED')),
    risk_level VARCHAR(20) DEFAULT 'low',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_login_events_org_id ON login_events(org_id);
CREATE INDEX IF NOT EXISTS idx_login_events_user_id ON login_events(user_id);
CREATE INDEX IF NOT EXISTS idx_login_events_status ON login_events(status);
CREATE INDEX IF NOT EXISTS idx_login_events_created_at ON login_events(created_at);

-- 2. Threat Alerts
CREATE TABLE IF NOT EXISTS threat_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID REFERENCES organizations(id),
    type VARCHAR(50) NOT NULL, -- e.g., 'BRUTE_FORCE', 'GEO_ANOMALY', 'SUSPICIOUS_IP'
    severity VARCHAR(20) NOT NULL, -- e.g., 'critical', 'warning', 'info'
    ip VARCHAR(45),
    user_id UUID REFERENCES users(id),
    description TEXT,
    status VARCHAR(20) DEFAULT 'open', -- 'open', 'acknowledged', 'resolved', 'blocked'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_threat_alerts_org_id ON threat_alerts(org_id);
CREATE INDEX IF NOT EXISTS idx_threat_alerts_status ON threat_alerts(status);

-- 3. MFA Users (Centralized status tracking)
CREATE TABLE IF NOT EXISTS mfa_users (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID REFERENCES organizations(id),
    status VARCHAR(20) DEFAULT 'DISABLED', -- 'ENABLED', 'DISABLED', 'ENFORCED'
    method VARCHAR(50), -- 'totp', 'sms'
    trusted_devices_count INTEGER DEFAULT 0,
    last_reset TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Session Enrichment (auth_tokens)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'auth_tokens' AND column_name = 'ip_address') THEN
        ALTER TABLE auth_tokens ADD COLUMN ip_address VARCHAR(45);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'auth_tokens' AND column_name = 'user_agent') THEN
        ALTER TABLE auth_tokens ADD COLUMN user_agent TEXT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'auth_tokens' AND column_name = 'last_active') THEN
        ALTER TABLE auth_tokens ADD COLUMN last_active TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'auth_tokens' AND column_name = 'is_active') THEN
        ALTER TABLE auth_tokens ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
    END IF;
END $$;
