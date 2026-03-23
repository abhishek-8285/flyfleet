-- Driver Enterprise (India Logistics) Final Implementation Migration

-- 1. Driver Onboarding Workflow
CREATE TABLE IF NOT EXISTS driver_onboarding_steps (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    step VARCHAR(50) NOT NULL, -- APPLICATION, DOCUMENT_VERIFICATION, etc.
    status VARCHAR(30) NOT NULL, -- PENDING, COMPLETED, REJECTED
    remarks TEXT,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Driver Shift Scheduling
CREATE TABLE IF NOT EXISTS driver_shifts (
    id UUID PRIMARY KEY,
    driver_id UUID,
    shift_start TIMESTAMP WITH TIME ZONE NOT NULL,
    shift_end TIMESTAMP WITH TIME ZONE NOT NULL,
    shift_type VARCHAR(20), -- DAY, NIGHT, ROTATIONAL
    status VARCHAR(20), -- SCHEDULED, ACTIVE, COMPLETED, CANCELLED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS driver_shift_assignments (
    id UUID PRIMARY KEY,
    shift_id UUID REFERENCES driver_shifts(id) ON DELETE CASCADE,
    vehicle_id UUID,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Driver Training System
CREATE TABLE IF NOT EXISTS driver_training_modules (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS driver_training_records (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    module_id UUID REFERENCES driver_training_modules(id),
    completed_at TIMESTAMP WITH TIME ZONE,
    certificate_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Driver Payroll System
CREATE TABLE IF NOT EXISTS driver_salary (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    base_salary NUMERIC(12, 2) DEFAULT 0,
    per_km_rate NUMERIC(12, 2) DEFAULT 0,
    daily_allowance NUMERIC(12, 2) DEFAULT 0,
    night_allowance NUMERIC(12, 2) DEFAULT 0,
    effective_from DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS driver_trip_earnings (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    trip_id UUID NOT NULL,
    distance_km NUMERIC(12, 2) DEFAULT 0,
    earnings NUMERIC(12, 2) DEFAULT 0,
    bonus NUMERIC(12, 2) DEFAULT 0,
    penalty NUMERIC(12, 2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS driver_payouts (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    total_amount NUMERIC(12, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, PAID, CANCELLED
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Driver Rating System
CREATE TABLE IF NOT EXISTS driver_ratings (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    trip_id UUID NOT NULL,
    punctuality_score INT CHECK (punctuality_score BETWEEN 1 AND 5),
    behavior_score INT CHECK (behavior_score BETWEEN 1 AND 5),
    driving_score INT CHECK (driving_score BETWEEN 1 AND 5),
    overall_rating NUMERIC(3, 2) NOT NULL,
    comments TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 6. Driver Safety Event System
CREATE TABLE IF NOT EXISTS driver_safety_events (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    trip_id UUID,
    event_type VARCHAR(50) NOT NULL, -- ACCIDENT, FATIGUE, OVERSPEED, etc.
    severity VARCHAR(20) NOT NULL, -- LOW, MEDIUM, HIGH, CRITICAL
    gps_lat NUMERIC(9, 6),
    gps_lng NUMERIC(9, 6),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 7. Driver Mobile Offline Sync
CREATE TABLE IF NOT EXISTS driver_mobile_sync_queue (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, SYNCED, FAILED
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 8. Analytics Views
CREATE OR REPLACE VIEW driver_performance_summary AS
SELECT
    driver_id,
    SUM(distance_km) AS total_distance,
    AVG(earnings) AS avg_earnings_per_trip,
    COUNT(trip_id) AS trip_count
FROM driver_trip_earnings
GROUP BY driver_id;

-- Note: driver_behavior_scores table might come from previous migrations or is assumed as current state
-- We will use the safety_score from the driver_v2_behavior_scores logic if it exists, or drivers table if renamed
CREATE OR REPLACE VIEW driver_safety_leaderboard AS
SELECT
    id AS driver_id,
    behavior_score AS avg_score -- Using behavior_score from drivers table as primary
FROM drivers
ORDER BY avg_score DESC;

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_onboarding_driver ON driver_onboarding_steps(driver_id);
CREATE INDEX IF NOT EXISTS idx_shifts_driver ON driver_shifts(driver_id);
CREATE INDEX IF NOT EXISTS idx_shifts_start ON driver_shifts(shift_start);
CREATE INDEX IF NOT EXISTS idx_payroll_driver ON driver_salary(driver_id);
CREATE INDEX IF NOT EXISTS idx_earnings_driver ON driver_trip_earnings(driver_id);
CREATE INDEX IF NOT EXISTS idx_safety_driver ON driver_safety_events(driver_id);
CREATE INDEX IF NOT EXISTS idx_sync_queue_status ON driver_mobile_sync_queue(status);
