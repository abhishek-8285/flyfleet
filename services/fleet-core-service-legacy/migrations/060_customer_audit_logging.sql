-- Migration: 060_customer_audit_logging
-- Description: Add audit logging for customer domain operations

-- Create customer audit log table
CREATE TABLE IF NOT EXISTS customer_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    
    -- Entity being audited
    entity_type VARCHAR(50) NOT NULL, -- customer, shipment, booking, invoice
    entity_id UUID NOT NULL,
    
    -- Action details
    action VARCHAR(50) NOT NULL, -- created, updated, deleted, status_changed
    old_value JSONB,
    new_value JSONB,
    
    -- User details
    user_id UUID,
    user_type VARCHAR(20), -- admin, customer, system
    user_email VARCHAR(255),
    
    -- Context
    ip_address INET,
    user_agent TEXT,
    reason TEXT, -- Reason for change (especially for status changes)
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_customer_audit_log_entity ON customer_audit_log(entity_type, entity_id);
CREATE INDEX idx_customer_audit_log_action ON customer_audit_log(action);
CREATE INDEX idx_customer_audit_log_user ON customer_audit_log(user_id);
CREATE INDEX idx_customer_audit_log_org ON customer_audit_log(org_id);
CREATE INDEX idx_customer_audit_log_created ON customer_audit_log(created_at);

-- Comments
COMMENT ON TABLE customer_audit_log IS 'Audit trail for customer domain operations';
COMMENT ON COLUMN customer_audit_log.entity_type IS 'Type of entity: customer, shipment, booking, invoice';
COMMENT ON COLUMN customer_audit_log.action IS 'Action performed: created, updated, deleted, status_changed';
COMMENT ON COLUMN customer_audit_log.old_value IS 'Previous state of entity (JSON)';
COMMENT ON COLUMN customer_audit_log.new_value IS 'New state of entity (JSON)';

-- Function to log customer changes
CREATE OR REPLACE FUNCTION log_customer_audit()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO customer_audit_log (
            org_id, entity_type, entity_id, action, new_value, user_type
        ) VALUES (
            NEW.org_id, 'customer', NEW.id, 'created', to_jsonb(NEW), 'system'
        );
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO customer_audit_log (
            org_id, entity_type, entity_id, action, old_value, new_value, user_type
        ) VALUES (
            NEW.org_id, 'customer', NEW.id, 'updated', to_jsonb(OLD), to_jsonb(NEW), 'system'
        );
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO customer_audit_log (
            org_id, entity_type, entity_id, action, old_value, user_type
        ) VALUES (
            OLD.org_id, 'customer', OLD.id, 'deleted', to_jsonb(OLD), 'system'
        );
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Add trigger for customer audit logging
DROP TRIGGER IF EXISTS trg_customer_audit ON customers;
CREATE TRIGGER trg_customer_audit
    AFTER INSERT OR UPDATE OR DELETE ON customers
    FOR EACH ROW
    EXECUTE FUNCTION log_customer_audit();

-- Function to log shipment status changes
CREATE OR REPLACE FUNCTION log_shipment_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO customer_audit_log (
            org_id, entity_type, entity_id, action, old_value, new_value, reason, user_type
        ) SELECT
            c.org_id, 'shipment', NEW.id, 'status_changed', 
            jsonb_build_object('status', OLD.status),
            jsonb_build_object('status', NEW.status),
            'Status changed from ' || OLD.status || ' to ' || NEW.status,
            'system'
        FROM customers c
        WHERE c.id = NEW.customer_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add trigger for shipment status audit
DROP TRIGGER IF EXISTS trg_shipment_status_audit ON shipments;
CREATE TRIGGER trg_shipment_status_audit
    AFTER UPDATE ON shipments
    FOR EACH ROW
    WHEN (OLD.status IS DISTINCT FROM NEW.status)
    EXECUTE FUNCTION log_shipment_status_change();

-- View for customer audit history
CREATE OR REPLACE VIEW customer_audit_history AS
SELECT 
    id,
    org_id,
    entity_type,
    entity_id,
    action,
    old_value,
    new_value,
    user_id,
    user_type,
    user_email,
    ip_address,
    reason,
    created_at
FROM customer_audit_log
ORDER BY created_at DESC;

-- View for shipment audit history
CREATE OR REPLACE VIEW shipment_audit_history AS
SELECT 
    cal.id,
    cal.org_id,
    cal.entity_id as shipment_id,
    cal.action,
    cal.old_value,
    cal.new_value,
    cal.user_id,
    cal.user_type,
    cal.reason,
    cal.created_at
FROM customer_audit_log cal
WHERE cal.entity_type = 'shipment'
ORDER BY cal.created_at DESC;
