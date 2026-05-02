-- Migration: 056_compliance_hos_limits
-- Description: Add HOS (Hours of Service) limits tracking for FMCSA compliance with tenant isolation.

-- 1. Create HOS daily limits tracking table
CREATE TABLE IF NOT EXISTS hos_daily_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    driver_id UUID NOT NULL,
    date DATE NOT NULL,
    driving_used_minutes INTEGER DEFAULT 0,
    driving_limit_minutes INTEGER DEFAULT 660,
    driving_remaining_minutes INTEGER GENERATED ALWAYS AS (driving_limit_minutes - driving_used_minutes) STORED,
    on_duty_used_minutes INTEGER DEFAULT 0,
    on_duty_limit_minutes INTEGER DEFAULT 840,
    on_duty_remaining_minutes INTEGER GENERATED ALWAYS AS (on_duty_limit_minutes - on_duty_used_minutes) STORED,
    seventy_hour_used_minutes INTEGER DEFAULT 0,
    seventy_hour_limit_minutes INTEGER DEFAULT 4200,
    seventy_hour_remaining_minutes INTEGER GENERATED ALWAYS AS (seventy_hour_limit_minutes - seventy_hour_used_minutes) STORED,
    last_break_duration_minutes INTEGER DEFAULT 0,
    last_break_start TIMESTAMPTZ,
    hours_since_break INTEGER DEFAULT 0,
    break_required BOOLEAN DEFAULT false,
    restart_available BOOLEAN DEFAULT false,
    restart_taken_at TIMESTAMPTZ,
    restart_eligible_at TIMESTAMPTZ,
    is_violation BOOLEAN DEFAULT false,
    violation_details JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, driver_id, date),
    CONSTRAINT chk_driving_used CHECK (driving_used_minutes >= 0),
    CONSTRAINT chk_on_duty_used CHECK (on_duty_used_minutes >= 0),
    CONSTRAINT chk_seventy_hour_used CHECK (seventy_hour_used_minutes >= 0)
);

CREATE INDEX IF NOT EXISTS idx_hos_daily_limits_tenant ON hos_daily_limits(tenant_id);

-- 2. Create HOS violations table
CREATE TABLE IF NOT EXISTS hos_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    driver_id UUID NOT NULL,
    violation_type VARCHAR(50) NOT NULL,
    violation_date DATE NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    detected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    limit_type VARCHAR(50),
    limit_value INTEGER,
    actual_value INTEGER,
    excess_minutes INTEGER,
    duty_status_log_id UUID REFERENCES duty_status_logs(id),
    vehicle_id UUID,
    location JSONB,
    severity VARCHAR(20) DEFAULT 'high',
    notes TEXT,
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID
);

CREATE INDEX IF NOT EXISTS idx_hos_violations_tenant ON hos_violations(tenant_id);

-- 3. Create HOS annotations table
CREATE TABLE IF NOT EXISTS hos_annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    duty_status_log_id UUID NOT NULL REFERENCES duty_status_logs(id) ON DELETE CASCADE,
    annotation_type VARCHAR(20) NOT NULL,
    reason TEXT NOT NULL,
    original_data JSONB,
    new_data JSONB,
    requires_approval BOOLEAN DEFAULT false,
    approved_by UUID,
    approved_at TIMESTAMPTZ,
    approval_status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID
);

CREATE INDEX IF NOT EXISTS idx_hos_annotations_tenant ON hos_annotations(tenant_id);

-- 4. Create ELD malfunctions table
CREATE TABLE IF NOT EXISTS eld_malfunctions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vehicle_id UUID NOT NULL,
    driver_id UUID,
    malfunction_type VARCHAR(50) NOT NULL,
    malfunction_code VARCHAR(20),
    description TEXT,
    status VARCHAR(20) DEFAULT 'open',
    detected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMPTZ,
    acknowledged_by UUID,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID,
    reported_to_carrier BOOLEAN DEFAULT false,
    reported_at TIMESTAMPTZ,
    carrier_reference VARCHAR(100),
    paper_logs_required BOOLEAN DEFAULT false,
    paper_logs_start TIMESTAMPTZ,
    paper_logs_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID
);

CREATE INDEX IF NOT EXISTS idx_eld_malfunctions_tenant ON eld_malfunctions(tenant_id);

-- 5. Create ELD diagnostic events table
CREATE TABLE IF NOT EXISTS eld_diagnostic_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    vehicle_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_code VARCHAR(20),
    event_description TEXT,
    detected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_eld_diagnostic_events_tenant ON eld_diagnostic_events(tenant_id);

-- 6. Create driver certification table
CREATE TABLE IF NOT EXISTS driver_certifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    driver_id UUID NOT NULL,
    certification_date DATE NOT NULL,
    certification_period_start DATE NOT NULL,
    certification_period_end DATE NOT NULL,
    certification_text TEXT NOT NULL DEFAULT 'I certify that all entries I have made on this driver''s record of duty status for the past 7 days are true and correct.',
    signature_image BYTEA,
    signature_timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    review_status VARCHAR(20) DEFAULT 'pending',
    review_comments TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, driver_id, certification_date)
);

CREATE INDEX IF NOT EXISTS idx_driver_certifications_tenant ON driver_certifications(tenant_id);
