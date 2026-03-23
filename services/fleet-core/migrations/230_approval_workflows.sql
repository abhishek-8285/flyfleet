-- Migration: 230_approval_workflows
-- Description: Create tables for multi-level approval workflows

CREATE TABLE IF NOT EXISTS approval_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    entity_type VARCHAR(50) NOT NULL, -- EXPENSE_REPORT, FUEL_LOG
    min_amount DECIMAL(15, 2) DEFAULT 0,
    max_amount DECIMAL(15, 2),
    levels JSONB NOT NULL, -- [{"role": "MANAGER", "index": 1}, {"role": "FINANCE", "index": 2}]
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS approval_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_id UUID NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    config_id UUID REFERENCES approval_configs(id),
    current_level_index INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED, CANCELLED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS approval_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID REFERENCES approval_instances(id),
    approver_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL, -- APPROVED, REJECTED
    level_index INTEGER NOT NULL,
    comments TEXT,
    action_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add column to tracking current approval state if needed, 
-- but we can use approval_instances to track it.

-- Indexes
CREATE INDEX idx_approval_configs_org ON approval_configs(organization_id);
CREATE INDEX idx_approval_instances_entity ON approval_instances(entity_id, entity_type);

-- Comments
COMMENT ON TABLE approval_configs IS 'Configuration for multi-level approval workflows based on organization and amount';
COMMENT ON TABLE approval_instances IS 'Running instances of approval workflows for specific entities';
COMMENT ON TABLE approval_actions IS 'Audit log of individual approval/rejection actions';
