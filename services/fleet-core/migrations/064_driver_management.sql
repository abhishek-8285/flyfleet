-- Migration: 064_driver_management
-- Description: Complete driver management schema with behavior scoring

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create drivers table
CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    license_number VARCHAR(50) NOT NULL,
    license_type VARCHAR(20), -- LMV, HMV, Transport
    license_expiry DATE NOT NULL,
    license_authority VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    address JSONB NOT NULL,
    emergency_contact JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'resigned')),
    current_vehicle_id UUID,
    current_trip_id UUID,
    joining_date DATE NOT NULL DEFAULT CURRENT_DATE,
    resignation_date DATE,
    experience_years INTEGER DEFAULT 0 CHECK (experience_years >= 0),
    rating DECIMAL(3, 2) DEFAULT 0.00 CHECK (rating >= 0 AND rating <= 5),
    total_trips INTEGER DEFAULT 0,
    total_distance_km DECIMAL(10, 2) DEFAULT 0.00,
    total_hours_driven DECIMAL(10, 2) DEFAULT 0.00,
    behavior_score INTEGER DEFAULT 100 CHECK (behavior_score >= 0 AND behavior_score <= 100),
    documents JSONB DEFAULT '[]'::jsonb,
    attendance_records JSONB DEFAULT '[]'::jsonb,
    schedules JSONB DEFAULT '[]'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT uq_drivers_email UNIQUE (email),
    CONSTRAINT uq_drivers_license UNIQUE (license_number),
    CONSTRAINT uq_drivers_org_email UNIQUE (org_id, email),
    CONSTRAINT uq_drivers_org_license UNIQUE (org_id, license_number)
);

-- Create driver_behavior table for behavior analytics
CREATE TABLE IF NOT EXISTS driver_behavior (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    org_id UUID NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    score DECIMAL(5, 2) DEFAULT 0.00 CHECK (score >= 0 AND score <= 100),
    safety_score DECIMAL(5, 2) DEFAULT 0.00,
    efficiency_score DECIMAL(5, 2) DEFAULT 0.00,
    compliance_score DECIMAL(5, 2) DEFAULT 0.00,
    harsh_accelerations INTEGER DEFAULT 0,
    harsh_brakings INTEGER DEFAULT 0,
    harsh_cornerings INTEGER DEFAULT 0,
    speeding_incidents INTEGER DEFAULT 0,
    phone_usage INTEGER DEFAULT 0,
    seatbelt_violations INTEGER DEFAULT 0,
    fuel_consumption DECIMAL(10, 2) DEFAULT 0.00,
    distance_driven DECIMAL(10, 2) DEFAULT 0.00,
    idle_time INTEGER DEFAULT 0,
    eco_driving_score DECIMAL(5, 2) DEFAULT 0.00,
    hours_of_service DECIMAL(10, 2) DEFAULT 0.00,
    rest_periods INTEGER DEFAULT 0,
    document_compliance BOOLEAN DEFAULT true,
    maintenance_adherence BOOLEAN DEFAULT true,
    night_driving_hours DECIMAL(10, 2) DEFAULT 0.00,
    highway_percentage DECIMAL(5, 2) DEFAULT 0.00,
    city_driving_percentage DECIMAL(5, 2) DEFAULT 0.00,
    risk_level VARCHAR(20) DEFAULT 'low' CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    risk_factors TEXT[] DEFAULT '{}',
    improvement_areas TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_behavior_period CHECK (period_end >= period_start)
);

-- Create driver_incidents table
CREATE TABLE IF NOT EXISTS driver_incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    org_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL, -- speeding, harsh_braking, harsh_acceleration, etc.
    severity VARCHAR(20) NOT NULL DEFAULT 'low' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    timestamp TIMESTAMPTZ NOT NULL,
    location JSONB,
    speed DECIMAL(5, 2),
    details JSONB,
    resolved BOOLEAN DEFAULT false,
    resolution_notes TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create behavior_score_updates table for audit trail
CREATE TABLE IF NOT EXISTS behavior_score_updates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    score_change INTEGER NOT NULL,
    reason TEXT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    details JSONB,
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create driver_messages table
CREATE TABLE IF NOT EXISTS driver_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    trip_id UUID,
    direction VARCHAR(20) NOT NULL CHECK (direction IN ('INBOUND', 'OUTBOUND')),
    type VARCHAR(20) NOT NULL DEFAULT 'TEXT' CHECK (type IN ('TEXT', 'AUDIO', 'IMAGE')),
    content TEXT,
    attachment_url TEXT,
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create driver_expenses table
CREATE TABLE IF NOT EXISTS driver_expenses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    trip_id UUID,
    type VARCHAR(50) NOT NULL CHECK (type IN ('FUEL', 'TOLL', 'MAINTENANCE', 'FOOD', 'LODGING', 'OTHER')),
    amount DECIMAL(10, 2) NOT NULL CHECK (amount >= 0),
    currency VARCHAR(3) DEFAULT 'USD',
    receipt_url TEXT,
    notes TEXT,
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    transaction_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create duty_status_logs table (HOOS)
CREATE TABLE IF NOT EXISTS duty_status_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('OFF_DUTY', 'SLEEPER_BERTH', 'DRIVING', 'ON_DUTY')),
    location JSONB,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    duration_minutes INTEGER,
    remarks TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create driver_audit_logs table for comprehensive audit trail
CREATE TABLE IF NOT EXISTS driver_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    org_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    user_id UUID NOT NULL,
    user_type VARCHAR(20) NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    reason TEXT,
    metadata JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_drivers_org ON drivers(org_id);
CREATE INDEX idx_drivers_status ON drivers(status);
CREATE INDEX idx_drivers_org_status ON drivers(org_id, status);
CREATE INDEX idx_drivers_email ON drivers(email);
CREATE INDEX idx_drivers_license ON drivers(license_number);
CREATE INDEX idx_drivers_current_vehicle ON drivers(current_vehicle_id);
CREATE INDEX idx_drivers_created ON drivers(created_at DESC);
CREATE INDEX idx_drivers_deleted ON drivers(deleted_at);

CREATE INDEX idx_driver_behavior_driver ON driver_behavior(driver_id);
CREATE INDEX idx_driver_behavior_period ON driver_behavior(period_start, period_end);
CREATE INDEX idx_driver_behavior_risk ON driver_behavior(risk_level);

CREATE INDEX idx_driver_incidents_driver ON driver_incidents(driver_id);
CREATE INDEX idx_driver_incidents_timestamp ON driver_incidents(timestamp DESC);
CREATE INDEX idx_driver_incidents_type ON driver_incidents(type);

CREATE INDEX idx_behavior_score_updates_driver ON behavior_score_updates(driver_id);
CREATE INDEX idx_behavior_score_updates_created ON behavior_score_updates(created_at DESC);

CREATE INDEX idx_driver_messages_driver ON driver_messages(driver_id);
CREATE INDEX idx_driver_messages_trip ON driver_messages(trip_id);
CREATE INDEX idx_driver_messages_created ON driver_messages(created_at DESC);

CREATE INDEX idx_driver_expenses_driver ON driver_expenses(driver_id);
CREATE INDEX idx_driver_expenses_trip ON driver_expenses(trip_id);
CREATE INDEX idx_driver_expenses_type ON driver_expenses(type);
CREATE INDEX idx_driver_expenses_status ON driver_expenses(status);

CREATE INDEX idx_duty_status_logs_driver ON duty_status_logs(driver_id);
CREATE INDEX idx_duty_status_logs_started ON duty_status_logs(started_at DESC);
CREATE INDEX idx_duty_status_logs_status ON duty_status_logs(status);

CREATE INDEX idx_driver_audit_logs_driver ON driver_audit_logs(driver_id);
CREATE INDEX idx_driver_audit_logs_org ON driver_audit_logs(org_id);
CREATE INDEX idx_driver_audit_logs_action ON driver_audit_logs(action);
CREATE INDEX idx_driver_audit_logs_created ON driver_audit_logs(created_at DESC);
CREATE INDEX idx_driver_audit_logs_user ON driver_audit_logs(user_id);

-- Comments
COMMENT ON TABLE drivers IS 'Driver profiles with license, behavior, and assignment tracking';
COMMENT ON TABLE driver_behavior IS 'Driver behavior analytics and scoring';
COMMENT ON TABLE driver_incidents IS 'Specific driving incidents (speeding, harsh braking, etc.)';
COMMENT ON TABLE behavior_score_updates IS 'Audit trail for behavior score changes';
COMMENT ON TABLE driver_messages IS 'Messages between driver and dispatch';
COMMENT ON TABLE driver_expenses IS 'Driver-incurred expenses';
COMMENT ON TABLE duty_status_logs IS 'Hours of Service (HOOS) logs';

COMMENT ON COLUMN drivers.behavior_score IS 'Overall behavior score 0-100';
COMMENT ON COLUMN drivers.rating IS 'Driver rating 0-5 stars';
COMMENT ON COLUMN driver_behavior.risk_level IS 'Risk assessment: low, medium, high, critical';
COMMENT ON COLUMN duty_status_logs.status IS 'FMCSA duty status: OFF_DUTY, SLEEPER_BERTH, DRIVING, ON_DUTY';

-- Create view for driver performance summary
CREATE OR REPLACE VIEW driver_performance_summary AS
SELECT 
    d.id,
    d.full_name,
    d.email,
    d.status,
    d.rating,
    d.behavior_score,
    d.total_trips,
    d.total_distance_km,
    d.total_hours_driven,
    d.current_vehicle_id,
    d.license_expiry,
    CASE 
        WHEN d.license_expiry < CURRENT_DATE THEN false
        ELSE true
    END as license_valid,
    CASE 
        WHEN d.current_vehicle_id IS NOT NULL THEN true
        ELSE false
    END as vehicle_assigned,
    COALESCE(MAX(dbr.created_at), d.created_at) as last_active,
    COALESCE(AVG(dbr.score), 0) as avg_behavior_score_30d
FROM drivers d
LEFT JOIN driver_behavior dbr ON d.id = dbr.driver_id 
    AND dbr.period_end >= CURRENT_DATE - INTERVAL '30 days'
WHERE d.deleted_at IS NULL
GROUP BY d.id, d.full_name, d.email, d.status, d.rating, d.behavior_score,
         d.total_trips, d.total_distance_km, d.total_hours_driven,
         d.current_vehicle_id, d.license_expiry, d.created_at;

-- Create view for drivers with expiring documents
CREATE OR REPLACE VIEW drivers_expiring_documents AS
SELECT 
    d.id,
    d.full_name,
    d.email,
    d.license_number,
    d.license_expiry,
    d.status,
    d.phone,
    CASE 
        WHEN d.license_expiry < CURRENT_DATE THEN 'expired'
        WHEN d.license_expiry < CURRENT_DATE + INTERVAL '30 days' THEN 'expiring_soon'
        ELSE 'valid'
    END as license_status,
    d.license_expiry - CURRENT_DATE as days_until_expiry
FROM drivers d
WHERE d.deleted_at IS NULL
  AND d.license_expiry < CURRENT_DATE + INTERVAL '90 days'
ORDER BY d.license_expiry ASC;

-- Create view for behavior score trends
CREATE OR REPLACE VIEW behavior_score_trends AS
SELECT 
    driver_id,
    DATE_TRUNC('week', period_end) as week,
    AVG(score) as avg_score,
    AVG(safety_score) as avg_safety_score,
    AVG(efficiency_score) as avg_efficiency_score,
    AVG(compliance_score) as avg_compliance_score,
    SUM(speeding_incidents) as total_speeding,
    SUM(harsh_brakings) as total_harsh_brakings,
    risk_level
FROM driver_behavior
WHERE period_end >= CURRENT_DATE - INTERVAL '90 days'
GROUP BY driver_id, DATE_TRUNC('week', period_end), risk_level
ORDER BY driver_id, week DESC;

-- Function to update driver statistics after trip completion
CREATE OR REPLACE FUNCTION update_driver_stats_after_trip()
RETURNS TRIGGER AS $$
BEGIN
    -- This would be called when a trip is completed
    -- Updates driver statistics
    UPDATE drivers SET
        total_trips = total_trips + 1,
        total_distance_km = total_distance_km + COALESCE(NEW.distance_km, 0),
        total_hours_driven = total_hours_driven + COALESCE(NEW.duration_hours, 0),
        updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.driver_id AND deleted_at IS NULL;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to check license expiry daily
CREATE OR REPLACE FUNCTION check_license_expiry()
RETURNS void AS $$
DECLARE
    driver_record RECORD;
BEGIN
    FOR driver_record IN 
        SELECT id, email, full_name, license_expiry
        FROM drivers
        WHERE deleted_at IS NULL
          AND status = 'active'
          AND license_expiry < CURRENT_DATE
    LOOP
        -- Log expired license
        RAISE NOTICE 'Driver % (%) has expired license: %', 
            driver_record.full_name, driver_record.email, driver_record.license_expiry;
        
        -- Could send notification here
        -- Could auto-suspend driver here
    END LOOP;
END;
$$ LANGUAGE plpgsql;
