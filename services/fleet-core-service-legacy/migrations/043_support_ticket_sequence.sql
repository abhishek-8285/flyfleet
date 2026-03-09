-- Migration: 043_support_ticket_sequence
-- Description: Add sequence for unique ticket number generation

-- Create sequence for support ticket numbers
CREATE SEQUENCE IF NOT EXISTS support_tickets_ticket_id_seq
    START WITH 1000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

-- Add ticket_id column to support_tickets if it doesn't exist
-- This is the human-readable ticket ID (e.g., INC-1000)
ALTER TABLE support_tickets 
    ADD COLUMN IF NOT EXISTS ticket_id VARCHAR(32);

-- Add index for ticket_id lookups
CREATE INDEX IF NOT EXISTS idx_support_tickets_ticket_id 
    ON support_tickets(ticket_id);

-- Add unique constraint on ticket_id
ALTER TABLE support_tickets 
    ADD CONSTRAINT support_tickets_ticket_id_key 
    UNIQUE (ticket_id);

-- Add tags and metadata columns if they don't exist
ALTER TABLE support_tickets 
    ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';

ALTER TABLE support_tickets 
    ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';

-- Add sla_deadline for SLA tracking
ALTER TABLE support_tickets 
    ADD COLUMN IF NOT EXISTS sla_deadline TIMESTAMP WITH TIME ZONE;

-- Create incident_comments table if it doesn't exist
CREATE TABLE IF NOT EXISTS incident_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_id UUID,
    author_name TEXT NOT NULL,
    comment_type TEXT NOT NULL DEFAULT 'INTERNAL', -- INTERNAL, CUSTOMER, SYSTEM
    content TEXT NOT NULL,
    is_private BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_incident_comments_ticket_id 
    ON incident_comments(ticket_id);

-- Create org_support_metrics table for caching metrics
CREATE TABLE IF NOT EXISTS org_support_metrics (
    org_id UUID PRIMARY KEY,
    open_tickets INTEGER DEFAULT 0,
    resolved_tickets INTEGER DEFAULT 0,
    avg_resolution_time DOUBLE PRECISION DEFAULT 0, -- in hours
    sla_breaches INTEGER DEFAULT 0,
    critical_issues INTEGER DEFAULT 0,
    last_calculated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create support_audit_log table for audit trail
CREATE TABLE IF NOT EXISTS support_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    user_id UUID,
    action TEXT NOT NULL, -- CREATE, UPDATE, STATUS_CHANGE, ASSIGNEE_CHANGE, COMMENT_ADD
    old_value JSONB,
    new_value JSONB,
    details TEXT,
    ip_address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_support_audit_log_ticket_id 
    ON support_audit_log(ticket_id);

CREATE INDEX IF NOT EXISTS idx_support_audit_log_user_id 
    ON support_audit_log(user_id);

CREATE INDEX IF NOT EXISTS idx_support_audit_log_created_at 
    ON support_audit_log(created_at);
