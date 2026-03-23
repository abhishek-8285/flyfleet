-- 041_support_persistence.sql
-- Adds real persistence for Support and Monitoring domains

CREATE TABLE IF NOT EXISTS support_tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    subject TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'OPEN', -- OPEN, IN_PROGRESS, RESOLVED, CLOSED
    priority TEXT NOT NULL DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, CRITICAL
    assigned_to UUID, -- User ID
    reporter_id UUID, -- User ID or Vehicle ID
    reporter_name TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS ticket_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_id UUID,
    author_name TEXT,
    event_type TEXT NOT NULL, -- COMMENT, STATUS_CHANGE, ASSIGNEE_CHANGE, PRIORITY_CHANGE
    message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS system_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    action TEXT NOT NULL, -- e.g., 'PASSWORD_RESET', 'MFA_RESET', 'ACCOUNT_UNLOCK'
    target_user_id UUID,
    ip_address TEXT,
    user_agent TEXT,
    status TEXT NOT NULL, -- SUCCESS, FAILURE
    error_message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_support_tickets_organization ON support_tickets(organization_id);
CREATE INDEX idx_support_tickets_status ON support_tickets(status);
CREATE INDEX idx_support_tickets_assigned ON support_tickets(assigned_to);
CREATE INDEX idx_ticket_events_ticket ON ticket_events(ticket_id);
CREATE INDEX idx_system_audit_logs_user ON system_audit_logs(user_id);
CREATE INDEX idx_system_audit_logs_target ON system_audit_logs(target_user_id);
CREATE INDEX idx_system_audit_logs_org ON system_audit_logs(organization_id);
