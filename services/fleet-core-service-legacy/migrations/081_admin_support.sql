-- 081_admin_support.sql
-- Admin Support Module Foundation

-- Organizations
CREATE TABLE organizations (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    plan_type TEXT,
    status TEXT NOT NULL, -- ACTIVE, SUSPENDED, DELETED
    region TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_org_status ON organizations(status);
CREATE INDEX idx_org_created ON organizations(created_at);

-- Organization Users (Mapping users to tenants)
CREATE TABLE organization_users (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id),
    email TEXT NOT NULL,
    role TEXT NOT NULL, -- OWNER, ADMIN, DISPATCHER, VIEWER
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, email)
);

-- Admin Users (Platform-level)
CREATE TABLE admin_users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    role TEXT NOT NULL, -- SUPER_ADMIN, PLATFORM_ADMIN, SUPPORT_AGENT, FINANCE_ADMIN, READ_ONLY_AUDITOR
    mfa_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Admin Permissions
CREATE TABLE admin_permissions (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL -- org.manage, ticket.manage, incident.manage, system.override, dr.execute
);

-- Admin RBAC (Admin User roles -> Permissions)
-- Since roles are currently just strings in admin_users, we might want a table later.
-- For now, we'll use a direct mapping or just check roles in code.
-- But the spec implies permissions are manageable.

-- Support Tickets
CREATE TABLE support_tickets (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id),
    subject TEXT NOT NULL,
    priority TEXT NOT NULL, -- LOW, MEDIUM, HIGH, CRITICAL
    status TEXT NOT NULL, -- OPEN, IN_PROGRESS, WAITING_CUSTOMER, RESOLVED, CLOSED, ESCALATED
    created_by UUID NOT NULL, -- Could be org user or admin
    assigned_to UUID, -- Admin user
    sla_deadline TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticket_org ON support_tickets(organization_id);
CREATE INDEX idx_ticket_status ON support_tickets(status);
CREATE INDEX idx_ticket_priority ON support_tickets(priority);

-- Ticket Messages
CREATE TABLE support_ticket_messages (
    id UUID PRIMARY KEY,
    ticket_id UUID NOT NULL REFERENCES support_tickets(id),
    sender_id UUID NOT NULL, -- admin_id or org_user_id
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Incidents
CREATE TABLE incidents (
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
CREATE TABLE incident_updates (
    id UUID PRIMARY KEY,
    incident_id UUID NOT NULL REFERENCES incidents(id),
    update_message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- System Flags
CREATE TABLE system_flags (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Maintenance Windows
CREATE TABLE maintenance_windows (
    id UUID PRIMARY KEY,
    service TEXT NOT NULL, -- tracking, video, integrations, etc.
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Impersonation Sessions
CREATE TABLE impersonation_sessions (
    id UUID PRIMARY KEY,
    admin_id UUID NOT NULL REFERENCES admin_users(id),
    user_id UUID NOT NULL, -- The user being impersonated
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL, -- Max 30 mins
    audit_log_id UUID -- Link to the audit log entry that started this
);

-- Admin Audit Logs (Append-only, immutable)
CREATE TABLE admin_audit_logs (
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

CREATE TRIGGER trg_protect_audit_logs
BEFORE UPDATE OR DELETE ON admin_audit_logs
FOR EACH ROW EXECUTE FUNCTION protect_audit_logs();
