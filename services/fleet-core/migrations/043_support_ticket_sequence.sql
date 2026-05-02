-- Migration: 043_support_ticket_sequence
-- Description: Add sequence for unique ticket number generation and supporting tables with tenant isolation.

-- 1. Create sequence for support ticket numbers
CREATE SEQUENCE IF NOT EXISTS support_tickets_ticket_id_seq
    START WITH 1000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

-- 2. Update support_tickets human-readable ID
ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS ticket_id VARCHAR(32);
CREATE INDEX IF NOT EXISTS idx_support_tickets_ticket_id ON support_tickets(ticket_id);

-- Add constraint safely
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM information_schema.constraint_column_usage WHERE table_name = 'support_tickets' AND column_name = 'ticket_id') THEN
        ALTER TABLE support_tickets ADD CONSTRAINT support_tickets_ticket_id_key UNIQUE (ticket_id);
    END IF;
END $$;

ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';
ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';
ALTER TABLE support_tickets ADD COLUMN IF NOT EXISTS sla_deadline TIMESTAMP WITH TIME ZONE;

-- 3. Create incident_comments table
CREATE TABLE IF NOT EXISTS incident_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_id UUID,
    author_name TEXT NOT NULL,
    comment_type TEXT NOT NULL DEFAULT 'INTERNAL',
    content TEXT NOT NULL,
    is_private BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_incident_comments_ticket_id ON incident_comments(ticket_id);

-- 4. Create tenant_support_metrics table (renamed from org_support_metrics)
CREATE TABLE IF NOT EXISTS tenant_support_metrics (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id),
    open_tickets INTEGER DEFAULT 0,
    resolved_tickets INTEGER DEFAULT 0,
    avg_resolution_time DOUBLE PRECISION DEFAULT 0,
    sla_breaches INTEGER DEFAULT 0,
    critical_issues INTEGER DEFAULT 0,
    last_calculated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Create support_audit_log table
CREATE TABLE IF NOT EXISTS support_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    user_id UUID,
    action TEXT NOT NULL,
    old_value JSONB,
    new_value JSONB,
    details TEXT,
    ip_address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_support_audit_log_ticket_id ON support_audit_log(ticket_id);
CREATE INDEX IF NOT EXISTS idx_support_audit_log_user_id ON support_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_support_audit_log_created_at ON support_audit_log(created_at);
