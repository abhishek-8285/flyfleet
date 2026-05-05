-- Migration: 056_compliance_hos_limits
-- Description: Add HOS (Hours of Service) limits tracking for FMCSA compliance with tenant isolation (Harmonized).

-- 1. Create HOS daily limits tracking table
CREATE TABLE IF NOT EXISTS hos_daily_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    -- Harmonization
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'hos_daily_limits' AND column_name = 'org_id') 
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'hos_daily_limits' AND column_name = 'tenant_id') THEN
        ALTER TABLE hos_daily_limits RENAME COLUMN org_id TO tenant_id;
    END IF;

    -- Ensure Columns
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS driver_id UUID NOT NULL;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS date DATE NOT NULL;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS driving_used_minutes INTEGER DEFAULT 0;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS driving_limit_minutes INTEGER DEFAULT 660;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS on_duty_used_minutes INTEGER DEFAULT 0;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS on_duty_limit_minutes INTEGER DEFAULT 840;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS seventy_hour_used_minutes INTEGER DEFAULT 0;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS seventy_hour_limit_minutes INTEGER DEFAULT 4200;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS last_break_duration_minutes INTEGER DEFAULT 0;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS last_break_start TIMESTAMPTZ;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS hours_since_break INTEGER DEFAULT 0;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS break_required BOOLEAN DEFAULT false;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS restart_available BOOLEAN DEFAULT false;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS restart_taken_at TIMESTAMPTZ;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS restart_eligible_at TIMESTAMPTZ;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS is_violation BOOLEAN DEFAULT false;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS violation_details JSONB;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE hos_daily_limits ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;

    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'hos_daily_limits' AND column_name = 'tenant_id') IN ('character varying', 'text') THEN
        ALTER TABLE hos_daily_limits ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_hos_daily_limits_tenant ON hos_daily_limits(tenant_id);

-- 2. Create HOS violations table
CREATE TABLE IF NOT EXISTS hos_violations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    -- Harmonization
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'hos_violations' AND column_name = 'org_id') 
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'hos_violations' AND column_name = 'tenant_id') THEN
        ALTER TABLE hos_violations RENAME COLUMN org_id TO tenant_id;
    END IF;

    -- Ensure Columns
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS driver_id UUID NOT NULL;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS violation_type VARCHAR(50) NOT NULL;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS violation_date DATE NOT NULL;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS occurred_at TIMESTAMPTZ NOT NULL;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS detected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS limit_type VARCHAR(50);
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS limit_value INTEGER;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS actual_value INTEGER;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS excess_minutes INTEGER;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS duty_status_log_id UUID REFERENCES duty_status_logs(id);
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS vehicle_id UUID;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS location JSONB;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS severity VARCHAR(20) DEFAULT 'high';
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS notes TEXT;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS resolved BOOLEAN DEFAULT false;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS resolved_by UUID;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE hos_violations ADD COLUMN IF NOT EXISTS created_by UUID;

    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'hos_violations' AND column_name = 'tenant_id') IN ('character varying', 'text') THEN
        ALTER TABLE hos_violations ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_hos_violations_tenant ON hos_violations(tenant_id);

-- 3. Create HOS annotations table
CREATE TABLE IF NOT EXISTS hos_annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    -- Ensure Columns
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS duty_status_log_id UUID REFERENCES duty_status_logs(id) ON DELETE CASCADE;
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS annotation_type VARCHAR(20) NOT NULL;
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS reason TEXT NOT NULL;
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS original_data JSONB;
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS new_data JSONB;
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS requires_approval BOOLEAN DEFAULT false;
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS approved_by UUID;
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS approval_status VARCHAR(20) DEFAULT 'pending';
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE hos_annotations ADD COLUMN IF NOT EXISTS created_by UUID;

    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'hos_annotations' AND column_name = 'tenant_id') IN ('character varying', 'text') THEN
        ALTER TABLE hos_annotations ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_hos_annotations_tenant ON hos_annotations(tenant_id);

-- 4. Create ELD malfunctions table
CREATE TABLE IF NOT EXISTS eld_malfunctions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    -- Ensure Columns
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS vehicle_id UUID NOT NULL;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS driver_id UUID;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS malfunction_type VARCHAR(50) NOT NULL;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS malfunction_code VARCHAR(20);
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS description TEXT;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'open';
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS detected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS acknowledged_at TIMESTAMPTZ;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS acknowledged_by UUID;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS resolved_by UUID;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS reported_to_carrier BOOLEAN DEFAULT false;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS reported_at TIMESTAMPTZ;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS carrier_reference VARCHAR(100);
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS paper_logs_required BOOLEAN DEFAULT false;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS paper_logs_start TIMESTAMPTZ;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS paper_logs_end TIMESTAMPTZ;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE eld_malfunctions ADD COLUMN IF NOT EXISTS created_by UUID;

    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'eld_malfunctions' AND column_name = 'tenant_id') IN ('character varying', 'text') THEN
        ALTER TABLE eld_malfunctions ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_eld_malfunctions_tenant ON eld_malfunctions(tenant_id);

-- 5. Create ELD diagnostic events table
CREATE TABLE IF NOT EXISTS eld_diagnostic_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    -- Ensure Columns
    ALTER TABLE eld_diagnostic_events ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    ALTER TABLE eld_diagnostic_events ADD COLUMN IF NOT EXISTS vehicle_id UUID NOT NULL;
    ALTER TABLE eld_diagnostic_events ADD COLUMN IF NOT EXISTS event_type VARCHAR(50) NOT NULL;
    ALTER TABLE eld_diagnostic_events ADD COLUMN IF NOT EXISTS event_code VARCHAR(20);
    ALTER TABLE eld_diagnostic_events ADD COLUMN IF NOT EXISTS event_description TEXT;
    ALTER TABLE eld_diagnostic_events ADD COLUMN IF NOT EXISTS detected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE eld_diagnostic_events ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMPTZ;
    ALTER TABLE eld_diagnostic_events ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;

    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'eld_diagnostic_events' AND column_name = 'tenant_id') IN ('character varying', 'text') THEN
        ALTER TABLE eld_diagnostic_events ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_eld_diagnostic_events_tenant ON eld_diagnostic_events(tenant_id);

-- 6. Create driver certification table
CREATE TABLE IF NOT EXISTS driver_certifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    -- Ensure Columns
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id);
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS driver_id UUID NOT NULL;
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS certification_date DATE NOT NULL;
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS certification_period_start DATE NOT NULL;
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS certification_period_end DATE NOT NULL;
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS certification_text TEXT NOT NULL DEFAULT 'I certify that all entries I have made on this driver''s record of duty status for the past 7 days are true and correct.';
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS signature_image BYTEA;
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS signature_timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS reviewed_by UUID;
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS review_status VARCHAR(20) DEFAULT 'pending';
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS review_comments TEXT;
    ALTER TABLE driver_certifications ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;

    -- Type Safety
    IF (SELECT data_type FROM information_schema.columns WHERE table_name = 'driver_certifications' AND column_name = 'tenant_id') IN ('character varying', 'text') THEN
        ALTER TABLE driver_certifications ALTER COLUMN tenant_id TYPE UUID USING tenant_id::uuid;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_driver_certifications_tenant ON driver_certifications(tenant_id);
