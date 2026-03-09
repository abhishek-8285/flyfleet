-- 082_admin_support_phase2.sql
-- Admin Support Module Phase 2: RBAC Refinement and Attachments

-- Admin Role Permissions (Role -> Permission mapping)
CREATE TABLE admin_role_permissions (
    role TEXT NOT NULL,
    permission TEXT NOT NULL REFERENCES admin_permissions(name),
    PRIMARY KEY (role, permission)
);

-- Support Ticket Attachments
CREATE TABLE support_ticket_attachments (
    id UUID PRIMARY KEY,
    ticket_id UUID NOT NULL REFERENCES support_tickets(id),
    file_url TEXT NOT NULL,
    uploaded_by UUID NOT NULL, -- admin_id or org_user_id
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed Initial Permissions
INSERT INTO admin_permissions (id, name) VALUES 
    (gen_random_uuid(), 'org.manage'),
    (gen_random_uuid(), 'ticket.manage'),
    (gen_random_uuid(), 'incident.manage'),
    (gen_random_uuid(), 'system.override'),
    (gen_random_uuid(), 'dr.execute')
ON CONFLICT (name) DO NOTHING;

-- Seed Default Role Mappings
-- SUPER_ADMIN: All permissions
INSERT INTO admin_role_permissions (role, permission) VALUES 
    ('SUPER_ADMIN', 'org.manage'),
    ('SUPER_ADMIN', 'ticket.manage'),
    ('SUPER_ADMIN', 'incident.manage'),
    ('SUPER_ADMIN', 'system.override'),
    ('SUPER_ADMIN', 'dr.execute');

-- PLATFORM_ADMIN: Orgs, Tickets, Incidents
INSERT INTO admin_role_permissions (role, permission) VALUES 
    ('PLATFORM_ADMIN', 'org.manage'),
    ('PLATFORM_ADMIN', 'ticket.manage'),
    ('PLATFORM_ADMIN', 'incident.manage');

-- SUPPORT_AGENT: Tickets
INSERT INTO admin_role_permissions (role, permission) VALUES 
    ('SUPPORT_AGENT', 'ticket.manage');

-- READ_ONLY_AUDITOR: (None for now, they can just read via handlers)
