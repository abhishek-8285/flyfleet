-- Migration: 063_diagnostics_audit_logging.sql
-- Description: Create audit logging table and triggers for diagnostics changes

CREATE TABLE IF NOT EXISTS diagnostics_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    entity_type TEXT,
    entity_id UUID,
    action TEXT,
    actor_id UUID,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Generic Audit Trigger Function
CREATE OR REPLACE FUNCTION trigger_insert_diagnostics_audit_log()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO diagnostics_audit_logs (entity_type, entity_id, action, metadata)
        VALUES (TG_TABLE_NAME, NEW.id, 'INSERT', row_to_json(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO diagnostics_audit_logs (entity_type, entity_id, action, metadata)
        VALUES (TG_TABLE_NAME, NEW.id, 'UPDATE', jsonb_build_object('old', row_to_json(OLD), 'new', row_to_json(NEW)));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO diagnostics_audit_logs (entity_type, entity_id, action, metadata)
        VALUES (TG_TABLE_NAME, OLD.id, 'DELETE', row_to_json(OLD));
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger for devices
DROP TRIGGER IF EXISTS trg_audit_devices ON devices;
CREATE TRIGGER trg_audit_devices
AFTER INSERT OR UPDATE OR DELETE ON devices
FOR EACH ROW EXECUTE FUNCTION trigger_insert_diagnostics_audit_log();

-- Trigger for firmware_packages
DROP TRIGGER IF EXISTS trg_audit_firmware_packages ON firmware_packages;
CREATE TRIGGER trg_audit_firmware_packages
AFTER INSERT OR UPDATE OR DELETE ON firmware_packages
FOR EACH ROW EXECUTE FUNCTION trigger_insert_diagnostics_audit_log();

-- Trigger for firmware_rollouts
DROP TRIGGER IF EXISTS trg_audit_firmware_rollouts ON firmware_rollouts;
CREATE TRIGGER trg_audit_firmware_rollouts
AFTER INSERT OR UPDATE OR DELETE ON firmware_rollouts
FOR EACH ROW EXECUTE FUNCTION trigger_insert_diagnostics_audit_log();

-- Trigger for device_commands
DROP TRIGGER IF EXISTS trg_audit_device_commands ON device_commands;
CREATE TRIGGER trg_audit_device_commands
AFTER INSERT OR UPDATE OR DELETE ON device_commands
FOR EACH ROW EXECUTE FUNCTION trigger_insert_diagnostics_audit_log();
