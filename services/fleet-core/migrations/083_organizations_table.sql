-- 083_organizations_table.sql
-- Organization Management for Multi-Tenancy

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    tax_id TEXT UNIQUE,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add foreign key constraints to existing tables
ALTER TABLE users ADD CONSTRAINT fk_user_organization FOREIGN KEY (organization_id) REFERENCES organizations(id);
ALTER TABLE subscriptions ADD CONSTRAINT fk_subscription_organization FOREIGN KEY (organization_id) REFERENCES organizations(id);
ALTER TABLE roles ADD CONSTRAINT fk_role_organization FOREIGN KEY (organization_id) REFERENCES organizations(id);

-- Seed System Organization
INSERT INTO organizations (id, name, tax_id, status) VALUES 
    ('00000000-0000-0000-0000-000000000000', 'System', 'SYSTEM-001', 'active')
ON CONFLICT DO NOTHING;
