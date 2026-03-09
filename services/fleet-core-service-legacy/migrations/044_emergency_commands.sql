-- Migration: 044_emergency_commands
-- Description: Add persistence for emergency halt commands and audit trail

-- Create emergency_commands table for tracking halt/clear commands
CREATE TABLE IF NOT EXISTS emergency_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    command_id VARCHAR(64) NOT NULL UNIQUE,
    command_type TEXT NOT NULL, -- GLOBAL_HALT, CLEAR_HALT
    reason TEXT,
    issued_by TEXT NOT NULL,
    cleared_by TEXT,
    issued_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    cleared_at TIMESTAMP WITH TIME ZONE,
    status TEXT NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, CLEARED, EXPIRED
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_emergency_commands_command_id 
    ON emergency_commands(command_id);

CREATE INDEX IF NOT EXISTS idx_emergency_commands_type 
    ON emergency_commands(command_type);

CREATE INDEX IF NOT EXISTS idx_emergency_commands_status 
    ON emergency_commands(status);

CREATE INDEX IF NOT EXISTS idx_emergency_commands_issued_at 
    ON emergency_commands(issued_at);

-- Create emergency_audit_log for detailed audit trail
CREATE TABLE IF NOT EXISTS emergency_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    command_id UUID NOT NULL REFERENCES emergency_commands(id) ON DELETE CASCADE,
    action TEXT NOT NULL, -- HALT_ISSUED, HALT_CLEARED, HALT_ACKNOWLEDGED
    actor TEXT NOT NULL,
    details TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_emergency_audit_log_command_id 
    ON emergency_audit_log(command_id);

CREATE INDEX IF NOT EXISTS idx_emergency_audit_log_created_at 
    ON emergency_audit_log(created_at);

-- Create emergency_status table for current state (single row)
CREATE TABLE IF NOT EXISTS emergency_status (
    id INTEGER PRIMARY KEY CHECK (id = 1), -- Enforce single row
    is_active BOOLEAN DEFAULT FALSE,
    current_command_id UUID REFERENCES emergency_commands(id),
    last_halt_at TIMESTAMP WITH TIME ZONE,
    last_clear_at TIMESTAMP WITH TIME ZONE,
    halt_reason TEXT,
    halt_issued_by TEXT,
    clear_issued_by TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Initialize with default status
INSERT INTO emergency_status (id, is_active) 
VALUES (1, FALSE)
ON CONFLICT (id) DO NOTHING;
