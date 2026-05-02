-- Migration: 044_emergency_commands
-- Description: Add persistence for emergency halt commands and audit trail with tenant isolation.

-- 1. Create emergency_commands table
CREATE TABLE IF NOT EXISTS emergency_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    command_id VARCHAR(64) NOT NULL UNIQUE,
    command_type TEXT NOT NULL,
    reason TEXT,
    issued_by TEXT NOT NULL,
    cleared_by TEXT,
    issued_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    cleared_at TIMESTAMP WITH TIME ZONE,
    status TEXT NOT NULL DEFAULT 'ACTIVE',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_emergency_commands_tenant_id ON emergency_commands(tenant_id);
CREATE INDEX IF NOT EXISTS idx_emergency_commands_command_id ON emergency_commands(command_id);

-- 2. Create emergency_audit_log
CREATE TABLE IF NOT EXISTS emergency_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    command_id UUID NOT NULL REFERENCES emergency_commands(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    actor TEXT NOT NULL,
    details TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_emergency_audit_log_tenant_id ON emergency_audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_emergency_audit_log_command_id ON emergency_audit_log(command_id);

-- 3. Create emergency_status table (Multi-tenant)
CREATE TABLE IF NOT EXISTS emergency_status (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id),
    is_active BOOLEAN DEFAULT FALSE,
    current_command_id UUID REFERENCES emergency_commands(id),
    last_halt_at TIMESTAMP WITH TIME ZONE,
    last_clear_at TIMESTAMP WITH TIME ZONE,
    halt_reason TEXT,
    halt_issued_by TEXT,
    clear_issued_by TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed System Tenant status
INSERT INTO emergency_status (tenant_id, is_active) 
VALUES ('00000000-0000-0000-0000-000000000000', FALSE)
ON CONFLICT (tenant_id) DO NOTHING;
