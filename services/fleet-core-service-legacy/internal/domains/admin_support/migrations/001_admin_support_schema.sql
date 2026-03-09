-- Admin Support Module Schema
-- FlyFleet Platform - Tenant Governance and Support Operations

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ─────────────────────────────────────────────────────────────────────────────
-- ORGANIZATIONS (Tenant Management)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    plan_type TEXT DEFAULT 'standard',
    region TEXT DEFAULT 'us-east-1',
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED', 'DELETED')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_org_status ON organizations(status);
CREATE INDEX idx_org_created ON organizations(created_at);
CREATE INDEX idx_org_name ON organizations(name);

-- Organization Users
CREATE TABLE organization_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('OWNER', 'ADMIN', 'DISPATCHER', 'VIEWER')),
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(organization_id, email)
);

CREATE INDEX idx_org_users_org ON organization_users(organization_id);
CREATE INDEX idx_org_users_email ON organization_users(email);

-- ─────────────────────────────────────────────────────────────────────────────
-- ADMIN USERS & RBAC
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE admin_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('SUPER_ADMIN', 'PLATFORM_ADMIN', 'SUPPORT_AGENT', 'FINANCE_ADMIN', 'READ_ONLY_AUDITOR')),
    mfa_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_users_email ON admin_users(email);
CREATE INDEX idx_admin_users_role ON admin_users(role);

-- Admin Permissions
CREATE TABLE admin_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT UNIQUE NOT NULL
);

-- Seed default permissions
INSERT INTO admin_permissions (name) VALUES
    ('org.manage'),
    ('org.suspend'),
    ('org.delete'),
    ('ticket.manage'),
    ('ticket.assign'),
    ('ticket.close'),
    ('incident.manage'),
    ('incident.declare'),
    ('incident.resolve'),
    ('system.override'),
    ('system.flag.manage'),
    ('system.maintenance'),
    ('dr.execute'),
    ('impersonation.start'),
    ('audit.view'),
    ('admin.manage');

-- Role-Permission Mapping
CREATE TABLE admin_role_permissions (
    role TEXT NOT NULL,
    permission TEXT NOT NULL REFERENCES admin_permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role, permission)
);

-- Seed default role permissions
INSERT INTO admin_role_permissions (role, permission) VALUES
    -- SUPER_ADMIN has all permissions
    ('SUPER_ADMIN', 'org.manage'),
    ('SUPER_ADMIN', 'org.suspend'),
    ('SUPER_ADMIN', 'org.delete'),
    ('SUPER_ADMIN', 'ticket.manage'),
    ('SUPER_ADMIN', 'ticket.assign'),
    ('SUPER_ADMIN', 'ticket.close'),
    ('SUPER_ADMIN', 'incident.manage'),
    ('SUPER_ADMIN', 'incident.declare'),
    ('SUPER_ADMIN', 'incident.resolve'),
    ('SUPER_ADMIN', 'system.override'),
    ('SUPER_ADMIN', 'system.flag.manage'),
    ('SUPER_ADMIN', 'system.maintenance'),
    ('SUPER_ADMIN', 'dr.execute'),
    ('SUPER_ADMIN', 'impersonation.start'),
    ('SUPER_ADMIN', 'audit.view'),
    ('SUPER_ADMIN', 'admin.manage'),
    -- PLATFORM_ADMIN
    ('PLATFORM_ADMIN', 'org.manage'),
    ('PLATFORM_ADMIN', 'org.suspend'),
    ('PLATFORM_ADMIN', 'ticket.manage'),
    ('PLATFORM_ADMIN', 'ticket.assign'),
    ('PLATFORM_ADMIN', 'incident.manage'),
    ('PLATFORM_ADMIN', 'incident.declare'),
    ('PLATFORM_ADMIN', 'incident.resolve'),
    ('PLATFORM_ADMIN', 'system.flag.manage'),
    ('PLATFORM_ADMIN', 'system.maintenance'),
    ('PLATFORM_ADMIN', 'impersonation.start'),
    ('PLATFORM_ADMIN', 'audit.view'),
    -- SUPPORT_AGENT
    ('SUPPORT_AGENT', 'ticket.manage'),
    ('SUPPORT_AGENT', 'ticket.assign'),
    ('SUPPORT_AGENT', 'ticket.close'),
    ('SUPPORT_AGENT', 'incident.declare'),
    ('SUPPORT_AGENT', 'audit.view'),
    -- FINANCE_ADMIN
    ('FINANCE_ADMIN', 'org.manage'),
    ('FINANCE_ADMIN', 'audit.view'),
    -- READ_ONLY_AUDITOR
    ('READ_ONLY_AUDITOR', 'audit.view');

-- ─────────────────────────────────────────────────────────────────────────────
-- SUPPORT TICKETS
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE support_tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    subject TEXT NOT NULL,
    priority TEXT NOT NULL CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'IN_PROGRESS', 'WAITING_CUSTOMER', 'RESOLVED', 'CLOSED', 'ESCALATED')),
    created_by UUID NOT NULL,
    assigned_to UUID REFERENCES admin_users(id),
    sla_deadline TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticket_org ON support_tickets(organization_id);
CREATE INDEX idx_ticket_status ON support_tickets(status);
CREATE INDEX idx_ticket_priority ON support_tickets(priority);
CREATE INDEX idx_ticket_assigned ON support_tickets(assigned_to);
CREATE INDEX idx_ticket_sla ON support_tickets(sla_deadline);

-- Ticket Messages
CREATE TABLE support_ticket_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticket_msg_ticket ON support_ticket_messages(ticket_id);
CREATE INDEX idx_ticket_msg_created ON support_ticket_messages(created_at);

-- Ticket Attachments
CREATE TABLE support_ticket_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    file_url TEXT NOT NULL,
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticket_att_ticket ON support_ticket_attachments(ticket_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- INCIDENT MANAGEMENT
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    severity TEXT NOT NULL CHECK (severity IN ('SEV1', 'SEV2', 'SEV3', 'SEV4')),
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MITIGATING', 'RESOLVED', 'POSTMORTEM_PENDING')),
    description TEXT NOT NULL,
    detected_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_incident_status ON incidents(status);
CREATE INDEX idx_incident_severity ON incidents(severity);
CREATE INDEX idx_incident_detected ON incidents(detected_at);

-- Incident Timeline Updates
CREATE TABLE incident_updates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    update_message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_incident_upd_incident ON incident_updates(incident_id);
CREATE INDEX idx_incident_upd_created ON incident_updates(created_at);

-- ─────────────────────────────────────────────────────────────────────────────
-- SYSTEM FLAGS & MAINTENANCE
-- ─────────────────────────────────────────────────────────────────────────────

-- System Flags (Feature Flags, Global Toggles)
CREATE TABLE system_flags (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed default system flags
INSERT INTO system_flags (key, value) VALUES
    ('tracking_pause', 'false'),
    ('video_pause', 'false'),
    ('integrations_pause', 'false'),
    ('billing_pause', 'false'),
    ('platform_maintenance', 'false');

-- Maintenance Windows
CREATE TABLE maintenance_windows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service TEXT NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_maint_service ON maintenance_windows(service);
CREATE INDEX idx_maint_start ON maintenance_windows(start_time);
CREATE INDEX idx_maint_end ON maintenance_windows(end_time);

-- ─────────────────────────────────────────────────────────────────────────────
-- IMPERSONATION SESSIONS
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE impersonation_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES admin_users(id),
    user_id UUID NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    audit_log_id UUID,
    CHECK (expires_at > started_at)
);

CREATE INDEX idx_impersonation_admin ON impersonation_sessions(admin_id);
CREATE INDEX idx_impersonation_user ON impersonation_sessions(user_id);
CREATE INDEX idx_impersonation_expires ON impersonation_sessions(expires_at);

-- ─────────────────────────────────────────────────────────────────────────────
-- AUDIT LOGS (Immutable, Append-Only)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE admin_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_id UUID NOT NULL,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    before_state JSONB,
    after_state JSONB,
    ip_address TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_actor ON admin_audit_logs(actor_id);
CREATE INDEX idx_audit_action ON admin_audit_logs(action);
CREATE INDEX idx_audit_resource ON admin_audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_created ON admin_audit_logs(created_at);

-- ─────────────────────────────────────────────────────────────────────────────
-- TRIGGERS FOR UPDATED_AT
-- ─────────────────────────────────────────────────────────────────────────────

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_org_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ticket_updated_at
    BEFORE UPDATE ON support_tickets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_incident_updated_at
    BEFORE UPDATE ON incidents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_flag_updated_at
    BEFORE UPDATE ON system_flags
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
