-- Migration: 060_customer_audit_logging
-- Description: Add audit logging for customer domain operations with tenant isolation.

-- 1. Create customer audit log table
CREATE TABLE IF NOT EXISTS customer_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    user_id UUID,
    user_type VARCHAR(20),
    user_email VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_customer_audit_log_tenant ON customer_audit_log(tenant_id);

-- 2. Function to log customer changes
CREATE OR REPLACE FUNCTION log_customer_audit()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO customer_audit_log (tenant_id, entity_type, entity_id, action, new_value, user_type)
        VALUES (NEW.tenant_id, 'customer', NEW.id, 'created', to_jsonb(NEW), 'system');
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO customer_audit_log (tenant_id, entity_type, entity_id, action, old_value, new_value, user_type)
        VALUES (NEW.tenant_id, 'customer', NEW.id, 'updated', to_jsonb(OLD), to_jsonb(NEW), 'system');
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO customer_audit_log (tenant_id, entity_type, entity_id, action, old_value, user_type)
        VALUES (OLD.tenant_id, 'customer', OLD.id, 'deleted', to_jsonb(OLD), 'system');
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_customer_audit ON customers;
CREATE TRIGGER trg_customer_audit
    AFTER INSERT OR UPDATE OR DELETE ON customers
    FOR EACH ROW
    EXECUTE FUNCTION log_customer_audit();

-- 3. Views
CREATE OR REPLACE VIEW customer_audit_history AS
SELECT * FROM customer_audit_log ORDER BY created_at DESC;
