-- Migration: 059_compliance_malfunction_reporting
-- Description: Add ELD malfunction and diagnostic event reporting (FMCSA required)

-- Add malfunction reporting columns to eld_malfunctions
ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS malfunction_category VARCHAR(50);
ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS is_power_failure BOOLEAN DEFAULT false;
ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS is_engine_sync_failure BOOLEAN DEFAULT false;
ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS is_timing_failure BOOLEAN DEFAULT false;
ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS is_position_failure BOOLEAN DEFAULT false;
ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS is_data_transfer_failure BOOLEAN DEFAULT false;

-- FMCSA requires malfunctions to be reported within 24 hours
ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS reported_to_fmcsa BOOLEAN DEFAULT false;
ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS reported_at TIMESTAMPTZ;
ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS fmcsa_reference VARCHAR(100);

-- Add index for malfunction reporting
CREATE INDEX IF NOT EXISTS idx_eld_malfunctions_reported ON eld_malfunctions(reported_to_fmcsa);
CREATE INDEX IF NOT EXISTS idx_eld_malfunctions_detected ON eld_malfunctions(detected_at);

-- Comments
COMMENT ON COLUMN eld_malfunctions.reported_to_fmcsa IS 'FMCSA requires reporting within 24 hours';
COMMENT ON COLUMN eld_malfunctions.malfunction_category IS 'POWER, ENGINE_SYNC, TIMING, POSITION, DATA_TRANSFER';

-- Function to check for unreported malfunctions
CREATE OR REPLACE FUNCTION get_unreported_malfunctions()
RETURNS TABLE(
    malfunction_id UUID,
    vehicle_id UUID,
    driver_id UUID,
    malfunction_type VARCHAR,
    detected_at TIMESTAMPTZ,
    hours_since_detection INTEGER,
    is_overdue BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        em.id,
        em.vehicle_id,
        em.driver_id,
        em.malfunction_type,
        em.detected_at,
        EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - em.detected_at)) / 3600::INTEGER as hours_since_detection,
        CASE 
            WHEN CURRENT_TIMESTAMP - em.detected_at > INTERVAL '24 hours' THEN true
            ELSE false
        END as is_overdue
    FROM eld_malfunctions em
    WHERE em.status = 'open'
      AND em.reported_to_fmcsa = false
    ORDER BY em.detected_at ASC;
END;
$$ LANGUAGE plpgsql;

-- Function to auto-detect malfunctions from diagnostic events
CREATE OR REPLACE FUNCTION detect_malfunction_from_event(
    p_vehicle_id UUID,
    p_event_type VARCHAR,
    p_event_code VARCHAR
) RETURNS VARCHAR AS $$
DECLARE
    v_malfunction_type VARCHAR;
BEGIN
    -- Map diagnostic events to malfunction types per FMCSA §395.34
    CASE p_event_type
        WHEN 'POWER_DATA_DIAGNOSTIC' THEN
            v_malfunction_type := 'POWER';
        WHEN 'ENGINE_SYNC_DIAGNOSTIC' THEN
            v_malfunction_type := 'ENGINE_SYNC';
        WHEN 'TIMING_DIAGNOSTIC' THEN
            v_malfunction_type := 'TIMING';
        WHEN 'POSITION_DIAGNOSTIC' THEN
            v_malfunction_type := 'POSITION';
        WHEN 'DATA_TRANSFER_DIAGNOSTIC' THEN
            v_malfunction_type := 'DATA_TRANSFER';
        ELSE
            v_malfunction_type := 'UNKNOWN';
    END CASE;
    
    RETURN v_malfunction_type;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-create malfunction from diagnostic event
CREATE OR REPLACE FUNCTION create_malfunction_from_diagnostic()
RETURNS TRIGGER AS $$
BEGIN
    -- Only create malfunction for certain event types
    IF NEW.event_type IN ('POWER_DATA_DIAGNOSTIC', 'ENGINE_SYNC_DIAGNOSTIC', 'TIMING_DIAGNOSTIC', 'POSITION_DIAGNOSTIC', 'DATA_TRANSFER_DIAGNOSTIC') THEN
        INSERT INTO eld_malfunctions (
            vehicle_id,
            malfunction_type,
            malfunction_code,
            event_description,
            detected_at
        ) VALUES (
            NEW.vehicle_id,
            detect_malfunction_from_event(NEW.vehicle_id, NEW.event_type, NEW.event_code),
            NEW.event_code,
            NEW.event_description,
            NEW.detected_at
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add trigger for automatic malfunction creation
DROP TRIGGER IF EXISTS trg_malfunction_from_diagnostic ON eld_diagnostic_events;
CREATE TRIGGER trg_malfunction_from_diagnostic
    AFTER INSERT ON eld_diagnostic_events
    FOR EACH ROW
    EXECUTE FUNCTION create_malfunction_from_diagnostic();

-- View for malfunction compliance status
CREATE OR REPLACE VIEW malfunction_compliance_status AS
SELECT 
    em.id,
    em.vehicle_id,
    em.driver_id,
    em.malfunction_type,
    em.status,
    em.detected_at,
    em.acknowledged_at,
    em.reported_to_fmcsa,
    em.reported_at,
    EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - em.detected_at)) / 3600 as hours_since_detection,
    CASE 
        WHEN em.status = 'resolved' THEN 'Compliant'
        WHEN em.reported_to_fmcsa = true THEN 'Reported'
        WHEN CURRENT_TIMESTAMP - em.detected_at > INTERVAL '24 hours' THEN 'Non-Compliant (Overdue)'
        WHEN em.status = 'open' AND CURRENT_TIMESTAMP - em.detected_at <= INTERVAL '24 hours' THEN 'Compliant (Within 24h)'
        ELSE 'Unknown'
    END as compliance_status
FROM eld_malfunctions em
WHERE em.created_at >= CURRENT_TIMESTAMP - INTERVAL '30 days'
ORDER BY em.detected_at DESC;
