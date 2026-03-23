-- Migration: 056_compliance_hos_limits
-- Description: Add HOS (Hours of Service) limits tracking for FMCSA compliance

-- Create HOS daily limits tracking table
CREATE TABLE IF NOT EXISTS hos_daily_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL,
    date DATE NOT NULL,
    
    -- 11-Hour Driving Limit
    driving_used_minutes INTEGER DEFAULT 0,
    driving_limit_minutes INTEGER DEFAULT 660, -- 11 hours = 660 minutes
    driving_remaining_minutes INTEGER GENERATED ALWAYS AS (driving_limit_minutes - driving_used_minutes) STORED,
    
    -- 14-Hour On-Duty Limit
    on_duty_used_minutes INTEGER DEFAULT 0,
    on_duty_limit_minutes INTEGER DEFAULT 840, -- 14 hours = 840 minutes
    on_duty_remaining_minutes INTEGER GENERATED ALWAYS AS (on_duty_limit_minutes - on_duty_used_minutes) STORED,
    
    -- 60/70-Hour Limit (7/8 consecutive days)
    seventy_hour_used_minutes INTEGER DEFAULT 0,
    seventy_hour_limit_minutes INTEGER DEFAULT 4200, -- 70 hours = 4200 minutes
    seventy_hour_remaining_minutes INTEGER GENERATED ALWAYS AS (seventy_hour_limit_minutes - seventy_hour_used_minutes) STORED,
    
    -- 30-Minute Break Rule
    last_break_duration_minutes INTEGER DEFAULT 0,
    last_break_start TIMESTAMPTZ,
    hours_since_break INTEGER DEFAULT 0,
    break_required BOOLEAN DEFAULT false,
    
    -- 34-Hour Restart
    restart_available BOOLEAN DEFAULT false,
    restart_taken_at TIMESTAMPTZ,
    restart_eligible_at TIMESTAMPTZ,
    
    -- Status
    is_violation BOOLEAN DEFAULT false,
    violation_details JSONB,
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(driver_id, date),
    
    -- Constraints
    CONSTRAINT chk_driving_used CHECK (driving_used_minutes >= 0),
    CONSTRAINT chk_on_duty_used CHECK (on_duty_used_minutes >= 0),
    CONSTRAINT chk_seventy_hour_used CHECK (seventy_hour_used_minutes >= 0)
);

-- Create HOS violations table
CREATE TABLE IF NOT EXISTS hos_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL,
    violation_type VARCHAR(50) NOT NULL, -- DRIVING_LIMIT, ON_DUTY_LIMIT, SEVENTY_HOUR, BREAK_REQUIRED
    violation_date DATE NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    detected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    -- Violation details
    limit_type VARCHAR(50), -- 11_HOUR, 14_HOUR, 70_HOUR, 30_MIN_BREAK
    limit_value INTEGER, -- The limit that was exceeded
    actual_value INTEGER, -- The actual value that exceeded the limit
    excess_minutes INTEGER, -- How many minutes over the limit
    
    -- Context
    duty_status_log_id UUID REFERENCES duty_status_logs(id),
    vehicle_id UUID,
    location JSONB,
    
    -- Resolution
    severity VARCHAR(20) DEFAULT 'high', -- low, medium, high, critical
    notes TEXT,
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID,
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    
    -- Constraints
    CONSTRAINT chk_violation_type CHECK (violation_type IN ('DRIVING_LIMIT', 'ON_DUTY_LIMIT', 'SEVENTY_HOUR', 'BREAK_REQUIRED', 'SLEEPER_BERTH')),
    CONSTRAINT chk_severity CHECK (severity IN ('low', 'medium', 'high', 'critical'))
);

-- Create HOS annotations table (for editing logs)
CREATE TABLE IF NOT EXISTS hos_annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    duty_status_log_id UUID NOT NULL REFERENCES duty_status_logs(id) ON DELETE CASCADE,
    
    -- Annotation details
    annotation_type VARCHAR(20) NOT NULL, -- edit, unassigned_driving, co_driver_assignment
    reason TEXT NOT NULL,
    original_data JSONB, -- Snapshot of original data before edit
    new_data JSONB, -- New data after edit
    
    -- Approval workflow
    requires_approval BOOLEAN DEFAULT false,
    approved_by UUID,
    approved_at TIMESTAMPTZ,
    approval_status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID, -- User who made the annotation
    
    CONSTRAINT chk_annotation_type CHECK (annotation_type IN ('edit', 'unassigned_driving', 'co_driver_assignment', 'duty_status_change'))
);

-- Create ELD malfunctions table
CREATE TABLE IF NOT EXISTS eld_malfunctions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL,
    driver_id UUID,
    
    -- Malfunction details
    malfunction_type VARCHAR(50) NOT NULL, -- POWER, ENGINE_SYNC, TIMING,_POSITION, DATA_TRANSFER
    malfunction_code VARCHAR(20),
    description TEXT,
    
    -- Status
    status VARCHAR(20) DEFAULT 'open', -- open, acknowledged, resolved
    detected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMPTZ,
    acknowledged_by UUID,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID,
    
    -- Reporting
    reported_to_carrier BOOLEAN DEFAULT false,
    reported_at TIMESTAMPTZ,
    carrier_reference VARCHAR(100),
    
    -- FMCSA requirements
    paper_logs_required BOOLEAN DEFAULT false,
    paper_logs_start TIMESTAMPTZ,
    paper_logs_end TIMESTAMPTZ,
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    
    CONSTRAINT chk_malfunction_type CHECK (malfunction_type IN ('POWER', 'ENGINE_SYNC', 'TIMING', 'POSITION', 'DATA_TRANSFER', 'UNKNOWN')),
    CONSTRAINT chk_malfunction_status CHECK (status IN ('open', 'acknowledged', 'resolved'))
);

-- Create ELD diagnostic events table
CREATE TABLE IF NOT EXISTS eld_diagnostic_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- POWER_DATA_DIAGNOSTIC, ENGINE_SYNC_DIAGNOSTIC, etc.
    event_code VARCHAR(20),
    event_description TEXT,
    
    detected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_diagnostic_event_type CHECK (event_type IN ('POWER_DATA_DIAGNOSTIC', 'ENGINE_SYNC_DIAGNOSTIC', 'TIMING_DIAGNOSTIC', 'POSITION_DIAGNOSTIC', 'DATA_TRANSFER_DIAGNOSTIC'))
);

-- Create driver certification table (for ROC)
CREATE TABLE IF NOT EXISTS driver_certifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL,
    
    -- Certification period
    certification_date DATE NOT NULL,
    certification_period_start DATE NOT NULL,
    certification_period_end DATE NOT NULL,
    
    -- Certification statement
    certification_text TEXT NOT NULL DEFAULT 'I certify that all entries I have made on this driver''s record of duty status for the past 7 days are true and correct.',
    
    -- Signature
    signature_image BYTEA, -- Signature image
    signature_timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    -- Review status
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    review_status VARCHAR(20) DEFAULT 'pending', -- pending, reviewed, rejected
    review_comments TEXT,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(driver_id, certification_date)
);

-- Indexes
CREATE INDEX idx_hos_daily_limits_driver ON hos_daily_limits(driver_id);
CREATE INDEX idx_hos_daily_limits_date ON hos_daily_limits(date);
CREATE INDEX idx_hos_daily_limits_violation ON hos_daily_limits(is_violation);
CREATE INDEX idx_hos_violations_driver ON hos_violations(driver_id);
CREATE INDEX idx_hos_violations_date ON hos_violations(violation_date);
CREATE INDEX idx_hos_violations_type ON hos_violations(violation_type);
CREATE INDEX idx_hos_violations_severity ON hos_violations(severity);
CREATE INDEX idx_hos_annotations_log ON hos_annotations(duty_status_log_id);
CREATE INDEX idx_eld_malfunctions_vehicle ON eld_malfunctions(vehicle_id);
CREATE INDEX idx_eld_malfunctions_status ON eld_malfunctions(status);
CREATE INDEX idx_driver_certifications_driver ON driver_certifications(driver_id);
CREATE INDEX idx_driver_certifications_date ON driver_certifications(certification_date);

-- Comments
COMMENT ON TABLE hos_daily_limits IS 'Tracks daily HOS limits for each driver (11-hour, 14-hour, 60/70-hour)';
COMMENT ON TABLE hos_violations IS 'Records HOS violations for compliance reporting';
COMMENT ON TABLE hos_annotations IS 'Annotations and edits to duty status logs (FMCSA requires annotations for all edits)';
COMMENT ON TABLE eld_malfunctions IS 'ELD malfunctions and diagnostics (FMCSA requires reporting within 24 hours)';
COMMENT ON TABLE driver_certifications IS 'Driver certifications for ROC (Record of Duty Status)';

COMMENT ON COLUMN hos_daily_limits.driving_limit_minutes IS '11 hours = 660 minutes (FMCSA limit)';
COMMENT ON COLUMN hos_daily_limits.on_duty_limit_minutes IS '14 hours = 840 minutes (FMCSA limit)';
COMMENT ON COLUMN hos_daily_limits.seventy_hour_limit_minutes IS '70 hours = 4200 minutes (FMCSA 8-day limit)';
COMMENT ON COLUMN hos_violations.excess_minutes IS 'Minutes over the legal limit';
COMMENT ON COLUMN eld_malfunctions.paper_logs_required IS 'If true, driver must use paper logs for 8 days';

-- Function to check if driver has HOS violations
CREATE OR REPLACE FUNCTION check_hos_violations(
    p_driver_id UUID,
    p_driving_used INTEGER,
    p_on_duty_used INTEGER,
    p_seventy_hour_used INTEGER,
    p_hours_since_break INTEGER
) RETURNS TABLE(
    violation_type VARCHAR,
    limit_type VARCHAR,
    limit_value INTEGER,
    actual_value INTEGER,
    excess_minutes INTEGER,
    severity VARCHAR
) AS $$
BEGIN
    -- Check 11-hour driving limit
    IF p_driving_used > 660 THEN
        RETURN QUERY SELECT 
            'DRIVING_LIMIT'::VARCHAR,
            '11_HOUR'::VARCHAR,
            660,
            p_driving_used,
            p_driving_used - 660,
            'critical'::VARCHAR;
    END IF;
    
    -- Check 14-hour on-duty limit
    IF p_on_duty_used > 840 THEN
        RETURN QUERY SELECT 
            'ON_DUTY_LIMIT'::VARCHAR,
            '14_HOUR'::VARCHAR,
            840,
            p_on_duty_used,
            p_on_duty_used - 840,
            'critical'::VARCHAR;
    END IF;
    
    -- Check 70-hour limit
    IF p_seventy_hour_used > 4200 THEN
        RETURN QUERY SELECT 
            'SEVENTY_HOUR'::VARCHAR,
            '70_HOUR'::VARCHAR,
            4200,
            p_seventy_hour_used,
            p_seventy_hour_used - 4200,
            'critical'::VARCHAR;
    END IF;
    
    -- Check 30-minute break rule (8 hours since break)
    IF p_hours_since_break > 480 THEN
        RETURN QUERY SELECT 
            'BREAK_REQUIRED'::VARCHAR,
            '30_MIN_BREAK'::VARCHAR,
            480,
            p_hours_since_break,
            p_hours_since_break - 480,
            'high'::VARCHAR;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate remaining HOS limits
CREATE OR REPLACE FUNCTION calculate_hos_remaining(
    p_driver_id UUID,
    p_date DATE
) RETURNS TABLE(
    driving_remaining INTEGER,
    on_duty_remaining INTEGER,
    seventy_hour_remaining INTEGER,
    break_required BOOLEAN
) AS $$
DECLARE
    v_limits hos_daily_limits%ROWTYPE;
BEGIN
    SELECT * INTO v_limits
    FROM hos_daily_limits
    WHERE driver_id = p_driver_id AND date = p_date;
    
    IF v_limits.id IS NULL THEN
        -- No limits tracked yet, return defaults
        RETURN QUERY SELECT 660, 840, 4200, false;
        RETURN;
    END IF;
    
    RETURN QUERY SELECT 
        v_limits.driving_remaining_minutes,
        v_limits.on_duty_remaining_minutes,
        v_limits.seventy_hour_remaining_minutes,
        v_limits.break_required;
END;
$$ LANGUAGE plpgsql;
