-- Support Service Database Schema
-- Migration: 022_support_service_schema.sql
-- Description: Core tables for Fleet Support/Operations Portal

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Support Tickets Table
CREATE TABLE IF NOT EXISTS support_tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id VARCHAR(50) UNIQUE NOT NULL, -- Human-readable ID like INC-12345
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN', -- OPEN, IN_PROGRESS, RESOLVED, CLOSED, ESCALATED
    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, CRITICAL
    category VARCHAR(100), -- Technical, Billing, Account, Feature Request
    assigned_to UUID REFERENCES users(id), -- Support agent
    reporter_id UUID REFERENCES users(id), -- Who reported the issue
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    sla_deadline TIMESTAMP WITH TIME ZONE,
    tags JSONB DEFAULT '[]'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Incident History/Comments Table
CREATE TABLE IF NOT EXISTS incident_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    author_id UUID REFERENCES users(id),
    author_name VARCHAR(100), -- For external users
    comment_type VARCHAR(20) DEFAULT 'INTERNAL', -- INTERNAL, EXTERNAL, SYSTEM
    content TEXT NOT NULL,
    is_private BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User Sessions Tracking
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(100) UNIQUE NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_info JSONB,
    login_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    logout_time TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    location JSONB -- {country, city, coordinates}
);

-- Support Audit Trail
CREATE TABLE IF NOT EXISTS support_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID REFERENCES support_tickets(id),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL, -- ASSIGN, STATUS_CHANGE, PRIORITY_CHANGE, COMMENT
    old_value JSONB,
    new_value JSONB,
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address VARCHAR(45)
);

-- Organization Support Metrics
CREATE TABLE IF NOT EXISTS org_support_metrics (
    org_id UUID PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    open_tickets INTEGER DEFAULT 0,
    resolved_tickets INTEGER DEFAULT 0,
    avg_resolution_time INTERVAL,
    sla_breaches INTEGER DEFAULT 0,
    critical_issues INTEGER DEFAULT 0,
    last_calculated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for Performance
CREATE INDEX IF NOT EXISTS idx_support_tickets_org_id ON support_tickets(org_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_status ON support_tickets(status);
CREATE INDEX IF NOT EXISTS idx_support_tickets_priority ON support_tickets(priority);
CREATE INDEX IF NOT EXISTS idx_support_tickets_assigned_to ON support_tickets(assigned_to);
CREATE INDEX IF NOT EXISTS idx_support_tickets_created_at ON support_tickets(created_at);
CREATE INDEX IF NOT EXISTS idx_incident_comments_ticket_id ON incident_comments(ticket_id);
CREATE INDEX IF NOT EXISTS idx_incident_comments_created_at ON incident_comments(created_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_is_active ON user_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_support_audit_ticket_id ON support_audit_log(ticket_id);
CREATE INDEX IF NOT EXISTS idx_support_audit_created_at ON support_audit_log(created_at);

-- Triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_support_tickets_updated_at 
    BEFORE UPDATE ON support_tickets 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_incident_comments_updated_at 
    BEFORE UPDATE ON incident_comments 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_org_support_metrics_updated_at 
    BEFORE UPDATE ON org_support_metrics 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Initial Seed Data
INSERT INTO support_tickets (ticket_id, org_id, title, description, status, priority, category, reporter_id) VALUES
('INC-1001', '00000000-0000-0000-0000-000000000001', 'Vehicle telemetry sync failure', 'Missing heartbeats from units in the southern corridor.', 'IN_PROGRESS', 'HIGH', 'Technical', '00000000-0000-0000-0000-000000000002'),
('INC-1002', '00000000-0000-0000-0000-000000000001', 'Driver app login issue', 'New drivers unable to complete initial registration flow.', 'PENDING', 'MEDIUM', 'Account', '00000000-0000-0000-0000-000000000003');

INSERT INTO incident_comments (ticket_id, author_name, content, comment_type) VALUES
((SELECT id FROM support_tickets WHERE ticket_id = 'INC-1001'), 'John D.', 'Initial investigation started. Checking network connectivity to southern region devices.', 'INTERNAL'),
((SELECT id FROM support_tickets WHERE ticket_id = 'INC-1001'), 'System Monitor', 'Automated alert triggered at 13:30 UTC', 'SYSTEM');

-- Initialize organization metrics
INSERT INTO org_support_metrics (org_id) 
SELECT id FROM organizations 
ON CONFLICT (org_id) DO NOTHING;

COMMENT ON TABLE support_tickets IS 'Core support ticket records';
COMMENT ON TABLE incident_comments IS 'Comments and history for support tickets';
COMMENT ON TABLE user_sessions IS 'User session tracking for support investigations';
COMMENT ON TABLE support_audit_log IS 'Audit trail for all support operations';
COMMENT ON TABLE org_support_metrics IS 'Pre-calculated organization support metrics';