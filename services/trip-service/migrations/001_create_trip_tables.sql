-- Create trips table
CREATE TABLE IF NOT EXISTS trips (
    id VARCHAR(36) PRIMARY KEY,
    org_id VARCHAR(36) NOT NULL,
    vehicle_id VARCHAR(36) NOT NULL,
    driver_id VARCHAR(36),
    status VARCHAR(15) NOT NULL DEFAULT 'planned' CHECK (status IN ('planned', 'in_progress', 'paused', 'completed', 'cancelled')),
    title VARCHAR(100) NOT NULL,
    description TEXT,
    planned_route JSONB DEFAULT '[]'::jsonb,
    actual_route JSONB DEFAULT '[]'::jsonb,
    waypoints JSONB DEFAULT '[]'::jsonb,
    estimated_distance_km DECIMAL(10,2) DEFAULT 0 CHECK (estimated_distance_km >= 0),
    actual_distance_km DECIMAL(10,2) CHECK (actual_distance_km >= 0),
    estimated_duration_min INTEGER DEFAULT 0 CHECK (estimated_duration_min >= 0),
    actual_duration_min INTEGER CHECK (actual_duration_min >= 0),
    cargo_details JSONB,
    scheduled_start TIMESTAMP WITH TIME ZONE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    start_location JSONB,
    end_location JSONB,
    fuel_consumed_liters DECIMAL(8,2) CHECK (fuel_consumed_liters >= 0),
    idling_time_minutes INTEGER DEFAULT 0 CHECK (idling_time_minutes >= 0),
    events JSONB DEFAULT '[]'::jsonb,
    pod_documents JSONB DEFAULT '[]'::jsonb,
    customer_name VARCHAR(100) NOT NULL,
    customer_phone VARCHAR(20) NOT NULL,
    customer_email VARCHAR(255),
    cancellation_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_trips_org_id ON trips(org_id);
CREATE INDEX IF NOT EXISTS idx_trips_vehicle_id ON trips(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_trips_driver_id ON trips(driver_id);
CREATE INDEX IF NOT EXISTS idx_trips_status ON trips(status);
CREATE INDEX IF NOT EXISTS idx_trips_scheduled_start ON trips(scheduled_start);
CREATE INDEX IF NOT EXISTS idx_trips_started_at ON trips(started_at);

-- Create updated_at trigger function if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for trips table
CREATE TRIGGER update_trips_updated_at
    BEFORE UPDATE ON trips
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
