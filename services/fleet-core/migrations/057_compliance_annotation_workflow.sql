-- Migration: 057_compliance_annotation_workflow
-- Description: Add editing and annotation workflow for duty status logs with correct column names.

-- 1. Resiliently update duty_status_logs
DO $$ 
BEGIN 
    -- Ensure tenant_id exists (critical for views and triggers)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'duty_status_logs' AND column_name = 'tenant_id') THEN
        ALTER TABLE duty_status_logs ADD COLUMN tenant_id UUID REFERENCES tenants(id);
    END IF;

    ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS is_edit BOOLEAN DEFAULT false;
    ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS edit_reason TEXT;
    ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS original_log_id UUID REFERENCES duty_status_logs(id);
    ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS edit_approved_by UUID;
    ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS edit_approved_at TIMESTAMPTZ;
END $$;

-- 2. Trigger for automatic annotation
CREATE OR REPLACE FUNCTION create_annotation_on_edit()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_edit = true AND NEW.edit_reason IS NOT NULL THEN
        INSERT INTO hos_annotations (
            tenant_id,
            duty_status_log_id,
            annotation_type,
            reason,
            original_data,
            new_data,
            requires_approval,
            created_by
        ) VALUES (
            NEW.tenant_id,
            NEW.id,
            'edit',
            NEW.edit_reason,
            to_jsonb(OLD),
            to_jsonb(NEW),
            false,
            NEW.edit_approved_by
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_annotation_on_edit ON duty_status_logs;
CREATE TRIGGER trg_annotation_on_edit
    AFTER UPDATE ON duty_status_logs
    FOR EACH ROW
    WHEN (OLD.status IS DISTINCT FROM NEW.status OR 
          OLD.started_at IS DISTINCT FROM NEW.started_at OR
          OLD.ended_at IS DISTINCT FROM NEW.ended_at)
    EXECUTE FUNCTION create_annotation_on_edit();

-- 3. Edit History View
CREATE OR REPLACE VIEW duty_status_edit_history AS
SELECT 
    dsl.id as log_id,
    dsl.tenant_id,
    dsl.driver_id,
    dsl.original_log_id,
    dsl.status,
    dsl.started_at,
    dsl.ended_at,
    dsl.is_edit,
    dsl.edit_reason,
    ha.annotation_type,
    ha.reason as annotation_reason,
    ha.created_at as edit_created_at
FROM duty_status_logs dsl
LEFT JOIN hos_annotations ha ON dsl.id = ha.duty_status_log_id
WHERE dsl.is_edit = true OR ha.id IS NOT NULL;
