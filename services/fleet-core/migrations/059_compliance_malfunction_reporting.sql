-- Migration: 059_compliance_malfunction_reporting
-- Description: Add ELD malfunction and diagnostic event reporting with tenant isolation.

-- 1. Resiliently update eld_malfunctions
DO $$ 
BEGIN 
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS malfunction_category VARCHAR(50);
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS is_power_failure BOOLEAN DEFAULT false;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS is_engine_sync_failure BOOLEAN DEFAULT false;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS is_timing_failure BOOLEAN DEFAULT false;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS is_position_failure BOOLEAN DEFAULT false;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS is_data_transfer_failure BOOLEAN DEFAULT false;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS reported_to_fmcsa BOOLEAN DEFAULT false;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS fmcsa_reference VARCHAR(100);
END $$;

-- 2. Trigger to auto-create malfunction from diagnostic event
CREATE OR REPLACE FUNCTION create_malfunction_from_diagnostic()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.event_type IN ('POWER_DATA_DIAGNOSTIC', 'ENGINE_SYNC_DIAGNOSTIC', 'TIMING_DIAGNOSTIC', 'POSITION_DIAGNOSTIC', 'DATA_TRANSFER_DIAGNOSTIC') THEN
        INSERT INTO eld_malfunctions (
            tenant_id,
            vehicle_id,
            malfunction_type,
            malfunction_code,
            description,
            detected_at
        ) VALUES (
            NEW.tenant_id,
            NEW.vehicle_id,
            CASE NEW.event_type
                WHEN 'POWER_DATA_DIAGNOSTIC' THEN 'POWER'
                WHEN 'ENGINE_SYNC_DIAGNOSTIC' THEN 'ENGINE_SYNC'
                WHEN 'TIMING_DIAGNOSTIC' THEN 'TIMING'
                WHEN 'POSITION_DIAGNOSTIC' THEN 'POSITION'
                WHEN 'DATA_TRANSFER_DIAGNOSTIC' THEN 'DATA_TRANSFER'
                ELSE 'UNKNOWN'
            END,
            NEW.event_code,
            NEW.event_description,
            NEW.detected_at
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_malfunction_from_diagnostic ON eld_diagnostic_events;
CREATE TRIGGER trg_malfunction_from_diagnostic
    AFTER INSERT ON eld_diagnostic_events
    FOR EACH ROW
    EXECUTE FUNCTION create_malfunction_from_diagnostic();

-- 3. Compliance View
CREATE OR REPLACE VIEW malfunction_compliance_status AS
SELECT 
    em.id,
    em.tenant_id,
    em.vehicle_id,
    em.malfunction_type,
    em.status,
    em.detected_at,
    em.reported_to_fmcsa,
    CASE 
        WHEN em.status = 'resolved' THEN 'Compliant'
        WHEN em.reported_to_fmcsa = true THEN 'Reported'
        WHEN CURRENT_TIMESTAMP - em.detected_at > INTERVAL '24 hours' THEN 'Non-Compliant (Overdue)'
        ELSE 'Compliant (Within 24h)'
    END as compliance_status
FROM eld_malfunctions em;
