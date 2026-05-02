-- 081_admin_support.sql
-- Admin Support Module Foundation

-- Organizations
DO $$ 
BEGIN 
    -- If organizations is a view, we need to handle it.
    -- However, for safety in this migration, we'll just ensure a table exists.
    -- If it's a view, we'll let it be for now and use tenants table instead if needed.
    -- But since this script expects a table, we check:
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'organizations' AND table_type = 'VIEW') THEN
        -- We can't easily drop a view if others depend on it.
        -- So we skip creating the table and rely on the view (which points to tenants).
        NULL;
    ELSE
        CREATE TABLE IF NOT EXISTS organizations (
            id UUID PRIMARY KEY,
            name TEXT NOT NULL,
            plan_type TEXT,
            status TEXT NOT NULL, -- ACTIVE, SUSPENDED, DELETED
            region TEXT,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        );
    END IF;
END $$;

-- Indexes for organizations (now on tenants table via the view)
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);
CREATE INDEX IF NOT EXISTS idx_tenants_created ON tenants(created_at);

-- Organization Users (Mapping users to tenants)
CREATE TABLE IF NOT EXISTS organization_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE organization_users ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE organization_users ADD COLUMN IF NOT EXISTS email TEXT NOT NULL;
    ALTER TABLE organization_users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL; -- OWNER, ADMIN, DISPATCHER, VIEWER
    ALTER TABLE organization_users ADD COLUMN IF NOT EXISTS status TEXT NOT NULL;
    ALTER TABLE organization_users ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'organization_users' AND constraint_name = 'uq_organization_users_tenant_email') THEN
        ALTER TABLE organization_users ADD CONSTRAINT uq_organization_users_tenant_email UNIQUE(tenant_id, email);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Admin Users (Platform-level)
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS email TEXT;
    ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL; -- SUPER_ADMIN, PLATFORM_ADMIN, SUPPORT_AGENT, FINANCE_ADMIN, READ_ONLY_AUDITOR
    ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN NOT NULL DEFAULT TRUE;
    ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_name = 'admin_users' AND constraint_name = 'admin_users_email_key') THEN
        ALTER TABLE admin_users ADD CONSTRAINT admin_users_email_key UNIQUE(email);
    END IF;
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

-- Admin Permissions
CREATE TABLE IF NOT EXISTS admin_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL -- org.manage, ticket.manage, incident.manage, system.override, dr.execute
);

-- Support Tickets
CREATE TABLE IF NOT EXISTS support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS subject TEXT NOT NULL DEFAULT '';
    ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS priority TEXT NOT NULL DEFAULT 'MEDIUM'; -- LOW, MEDIUM, HIGH, CRITICAL
    ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'OPEN'; -- OPEN, IN_PROGRESS, WAITING_CUSTOMER, RESOLVED, CLOSED, ESCALATED
    ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS created_by UUID; -- Could be org user or admin
    ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS assigned_to UUID; -- Admin user
    ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS sla_deadline TIMESTAMP;
    ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();
    ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_ticket_tenant ON support_tickets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ticket_status ON support_tickets(status);
CREATE INDEX IF NOT EXISTS idx_ticket_priority ON support_tickets(priority);

-- Ticket Messages
CREATE TABLE IF NOT EXISTS support_ticket_messages (
    id UUID PRIMARY KEY,
    ticket_id UUID NOT NULL REFERENCES support_tickets(id),
    sender_id UUID NOT NULL, -- admin_id or org_user_id
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Incidents
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY,
    severity TEXT NOT NULL, -- SEV1, SEV2, SEV3, SEV4
    status TEXT NOT NULL, -- OPEN, MITIGATING, RESOLVED, POSTMORTEM_PENDING
    description TEXT NOT NULL,
    detected_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Incident Updates
CREATE TABLE IF NOT EXISTS incident_updates (
    id UUID PRIMARY KEY,
    incident_id UUID NOT NULL REFERENCES incidents(id),
    update_message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- System Flags
CREATE TABLE IF NOT EXISTS system_flags (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Maintenance Windows
CREATE TABLE IF NOT EXISTS maintenance_windows (
    id UUID PRIMARY KEY,
    service TEXT NOT NULL, -- tracking, video, integrations, etc.
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Impersonation Sessions
CREATE TABLE IF NOT EXISTS impersonation_sessions (
    id UUID PRIMARY KEY,
    admin_id UUID NOT NULL REFERENCES admin_users(id),
    user_id UUID NOT NULL, -- The user being impersonated
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL, -- Max 30 mins
    audit_log_id UUID -- Link to the audit log entry that started this
);

-- Admin Audit Logs (Append-only, immutable)
CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id UUID PRIMARY KEY,
    actor_id UUID NOT NULL, -- admin_id
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    before_state JSONB,
    after_state JSONB,
    ip_address TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Ensure immutability via Trigger (soft measure)
CREATE OR REPLACE FUNCTION protect_audit_logs()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'UPDATE' OR TG_OP = 'DELETE') THEN
        RAISE EXCEPTION 'Admin Audit Logs are immutable';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_protect_audit_logs') THEN
        CREATE TRIGGER trg_protect_audit_logs
        BEFORE UPDATE OR DELETE ON admin_audit_logs
        FOR EACH ROW EXECUTE FUNCTION protect_audit_logs();
    END IF;
END $$;
