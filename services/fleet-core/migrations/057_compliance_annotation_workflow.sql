-- Migration: 057_compliance_annotation_workflow
-- Description: Add editing and annotation workflow for duty status logs

-- Add columns to duty_status_logs for editing support
ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS is_edit BOOLEAN DEFAULT false;
ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS edit_reason TEXT;
ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS original_log_id UUID REFERENCES duty_status_logs(id);
ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS edit_approved_by UUID;
ALTER TABLE duty_status_logs ADD COLUMN IF NOT EXISTS edit_approved_at TIMESTAMPTZ;

-- Add index for edit tracking
CREATE INDEX IF NOT EXISTS idx_duty_status_logs_is_edit ON duty_status_logs(is_edit);
CREATE INDEX IF NOT EXISTS idx_duty_status_logs_original ON duty_status_logs(original_log_id);

-- Comments
COMMENT ON COLUMN duty_status_logs.is_edit IS 'True if this is an edited version of a log';
COMMENT ON COLUMN duty_status_logs.edit_reason IS 'FMCSA requires reason for all edits';
COMMENT ON COLUMN duty_status_logs.original_log_id IS 'Reference to original log being edited';

-- Function to create annotation when log is edited
CREATE OR REPLACE FUNCTION create_annotation_on_edit()
RETURNS TRIGGER AS $$
BEGIN
    -- Only create annotation if this is an edit
    IF NEW.is_edit = true AND NEW.edit_reason IS NOT NULL THEN
        INSERT INTO hos_annotations (
            duty_status_log_id,
            annotation_type,
            reason,
            original_data,
            new_data,
            requires_approval,
            created_by
        ) VALUES (
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

-- Add trigger for automatic annotation
DROP TRIGGER IF EXISTS trg_annotation_on_edit ON duty_status_logs;
CREATE TRIGGER trg_annotation_on_edit
    AFTER UPDATE ON duty_status_logs
    FOR EACH ROW
    WHEN (OLD.status IS DISTINCT FROM NEW.status OR 
          OLD.start_time IS DISTINCT FROM NEW.start_time OR
          OLD.ended_at IS DISTINCT FROM NEW.ended_at)
    EXECUTE FUNCTION create_annotation_on_edit();

-- View for audit trail of edits
CREATE OR REPLACE VIEW duty_status_edit_history AS
SELECT 
    dsl.id as log_id,
    dsl.driver_id,
    dsl.original_log_id,
    dsl.status,
    dsl.start_time,
    dsl.ended_at,
    dsl.is_edit,
    dsl.edit_reason,
    dsl.edit_approved_by,
    dsl.edit_approved_at,
    ha.annotation_type,
    ha.reason as annotation_reason,
    ha.original_data,
    ha.new_data,
    ha.created_at as edit_created_at,
    ha.created_by as edit_created_by
FROM duty_status_logs dsl
LEFT JOIN hos_annotations ha ON dsl.id = ha.duty_status_log_id
WHERE dsl.is_edit = true OR ha.id IS NOT NULL
ORDER BY dsl.start_time DESC;
