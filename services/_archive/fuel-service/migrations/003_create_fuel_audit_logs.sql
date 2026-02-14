-- Create fuel audit logs table
CREATE TABLE IF NOT EXISTS fuel_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fuel_log_id UUID NOT NULL REFERENCES fuel_logs(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    changed_by UUID NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    old_values JSONB,
    new_values JSONB,
    diff JSONB
);

-- Index for searching audit logs by fuel log
CREATE INDEX IF NOT EXISTS idx_fuel_audit_logs_fuel_log ON fuel_audit_logs(fuel_log_id);
CREATE INDEX IF NOT EXISTS idx_fuel_audit_logs_action ON fuel_audit_logs(action);
