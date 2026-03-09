-- Migration: 049_invoice_audit_log
-- Description: Add audit logging for invoice changes

-- Create invoice audit log table
CREATE TABLE IF NOT EXISTS invoice_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL, -- created, updated, issued, cancelled, payment_recorded, etc.
    old_value JSONB,
    new_value JSONB,
    changed_by UUID,
    changed_by_email VARCHAR(255),
    changed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    reason TEXT -- Optional reason for change (e.g., cancellation reason)
);

-- Indexes for performance
CREATE INDEX idx_invoice_audit_log_invoice_id ON invoice_audit_log(invoice_id);
CREATE INDEX idx_invoice_audit_log_action ON invoice_audit_log(action);
CREATE INDEX idx_invoice_audit_log_changed_at ON invoice_audit_log(changed_at);
CREATE INDEX idx_invoice_audit_log_changed_by ON invoice_audit_log(changed_by);

-- Comments
COMMENT ON TABLE invoice_audit_log IS 'Audit trail for all invoice changes';
COMMENT ON COLUMN invoice_audit_log.action IS 'Type of action performed';
COMMENT ON COLUMN invoice_audit_log.old_value IS 'Previous state of invoice (JSON)';
COMMENT ON COLUMN invoice_audit_log.new_value IS 'New state of invoice (JSON)';
COMMENT ON COLUMN invoice_audit_log.ip_address IS 'IP address of user who made change';
COMMENT ON COLUMN invoice_audit_log.reason IS 'Reason for change (required for cancellations)';

-- Function to automatically log invoice changes
CREATE OR REPLACE FUNCTION log_invoice_change()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO invoice_audit_log (invoice_id, action, new_value, changed_by, changed_at)
        VALUES (NEW.id, 'created', to_jsonb(NEW), NEW.created_by, NEW.created_at);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO invoice_audit_log (invoice_id, action, old_value, new_value, changed_by, changed_at)
        VALUES (NEW.id, 'updated', to_jsonb(OLD), to_jsonb(NEW), NEW.updated_by, NEW.updated_at);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO invoice_audit_log (invoice_id, action, old_value, changed_at)
        VALUES (OLD.id, 'deleted', to_jsonb(OLD), CURRENT_TIMESTAMP);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Add trigger to invoices table
DROP TRIGGER IF EXISTS trg_invoice_audit_log ON invoices;
CREATE TRIGGER trg_invoice_audit_log
    AFTER INSERT OR UPDATE OR DELETE ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION log_invoice_change();

-- Create view for easy audit log查询
CREATE OR REPLACE VIEW invoice_audit_log_summary AS
SELECT 
    invoice_id,
    COUNT(*) as total_changes,
    MAX(changed_at) as last_changed_at,
    array_agg(DISTINCT action) as actions_performed
FROM invoice_audit_log
GROUP BY invoice_id;
