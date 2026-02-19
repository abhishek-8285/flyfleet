-- Create drivers table
CREATE TABLE IF NOT EXISTS drivers (
    id VARCHAR(36) PRIMARY KEY,
    org_id VARCHAR(36) NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    full_name VARCHAR(101) GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) NOT NULL,
    license_number VARCHAR(50) UNIQUE NOT NULL,
    license_type VARCHAR(20),
    license_expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    license_authority VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    address JSONB,
    emergency_contact JSONB,
    status VARCHAR(15) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'resigned')),
    current_vehicle_id VARCHAR(36),
    current_trip_id VARCHAR(36),
    joining_date DATE NOT NULL,
    resignation_date DATE,
    experience_years INTEGER DEFAULT 0 CHECK (experience_years >= 0),
    rating DECIMAL(3,2) DEFAULT 0 CHECK (rating >= 0 AND rating <= 5),
    total_trips INTEGER DEFAULT 0 CHECK (total_trips >= 0),
    total_distance_km DECIMAL(10,2) DEFAULT 0 CHECK (total_distance_km >= 0),
    total_hours_driven DECIMAL(10,2) DEFAULT 0 CHECK (total_hours_driven >= 0),
    behavior_score INTEGER DEFAULT 50 CHECK (behavior_score >= 0 AND behavior_score <= 100),
    documents JSONB DEFAULT '[]'::jsonb,
    attendance_records JSONB DEFAULT '[]'::jsonb,
    schedules JSONB DEFAULT '[]'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_drivers_org_id ON drivers(org_id);
CREATE INDEX IF NOT EXISTS idx_drivers_email ON drivers(email);
CREATE INDEX IF NOT EXISTS idx_drivers_license_number ON drivers(license_number);
CREATE INDEX IF NOT EXISTS idx_drivers_status ON drivers(status);
CREATE INDEX IF NOT EXISTS idx_drivers_current_vehicle_id ON drivers(current_vehicle_id);

-- Create updated_at trigger function if it doesn't exist
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for drivers table
CREATE TRIGGER update_drivers_updated_at
    BEFORE UPDATE ON drivers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
