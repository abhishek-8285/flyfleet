-- Migration: 304_org_hierarchy
-- Description: Add branches and cost_centers tables for organizational hierarchy

CREATE TABLE IF NOT EXISTS branches (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    code        VARCHAR(50),
    location    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS cost_centers (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    code        VARCHAR(50),
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_branches_tenant ON branches(tenant_id);
CREATE INDEX IF NOT EXISTS idx_cost_centers_tenant ON cost_centers(tenant_id);

-- Seed some initial data for the system organization (00000000-0000-0000-0000-000000000000)
-- and our development organization if it exists.
INSERT INTO branches (id, tenant_id, name, code)
VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', '00000000-0000-0000-0000-000000000000', 'Gurugram Main Depot', 'GRG-01'),
    ('550e8400-e29b-41d4-a716-446655440002', '00000000-0000-0000-0000-000000000000', 'Delhi Central', 'DEL-01')
ON CONFLICT (id) DO NOTHING;

INSERT INTO cost_centers (id, tenant_id, name, code)
VALUES 
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000000', 'Operations - North', 'CC-OPS-N'),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000000', 'Supply Chain Mumbai', 'CC-SCM-MUM')
ON CONFLICT DO NOTHING;
