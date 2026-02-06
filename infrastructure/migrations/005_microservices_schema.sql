-- Enable UUID extension if not enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- VEHICLES (Re-defining for Microservice Compatibility)
DROP TABLE IF EXISTS vehicles CASCADE;
CREATE TABLE vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    license_plate VARCHAR(20) NOT NULL UNIQUE,
    model VARCHAR(50) NOT NULL,
    make VARCHAR(50) NOT NULL,
    year INTEGER NOT NULL,
    color VARCHAR(30),
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, maintenance, retired
    capacity INTEGER DEFAULT 1,
    fuel_type VARCHAR(20), -- petrol, diesel, electric, hybrid
    vin VARCHAR(50),
    registration VARCHAR(50),
    insurance VARCHAR(50),
    odometer BIGINT DEFAULT 0,
    last_service TIMESTAMP WITH TIME ZONE,
    current_driver_id UUID, -- For direct assignment
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_vehicles_org_id ON vehicles(org_id);
CREATE INDEX idx_vehicles_status ON vehicles(status);
CREATE INDEX idx_vehicles_license ON vehicles(license_plate);

-- DRIVERS
DROP TABLE IF EXISTS drivers CASCADE;
CREATE TABLE drivers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    full_name VARCHAR(100) GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED,
    email VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    license_number VARCHAR(50) NOT NULL,
    license_type VARCHAR(20), -- LMV, HMV, Transport
    license_expiry TIMESTAMP WITH TIME ZONE,
    license_authority VARCHAR(100),
    date_of_birth TIMESTAMP WITH TIME ZONE,
    address JSONB,          -- Struct: Street, City, State, Pincode...
    emergency_contact JSONB, -- Struct: Name, Phone, Relation
    status VARCHAR(20) DEFAULT 'active',
    current_vehicle_id UUID REFERENCES vehicles(id),
    current_trip_id UUID,
    joining_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resignation_date TIMESTAMP WITH TIME ZONE,
    experience_years INTEGER DEFAULT 0,
    rating DOUBLE PRECISION DEFAULT 5.0,
    total_trips INTEGER DEFAULT 0,
    total_distance_km DOUBLE PRECISION DEFAULT 0,
    total_hours_driven DOUBLE PRECISION DEFAULT 0,
    behavior_score INTEGER DEFAULT 100,
    documents JSONB,         -- Array of Document objects
    attendance_records JSONB, -- Array of Attendance objects (or create separate table if normalized)
    schedules JSONB,         -- Array of Schedule objects
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_drivers_org_id ON drivers(org_id);
CREATE INDEX idx_drivers_status ON drivers(status);
CREATE INDEX idx_drivers_vehicle ON drivers(current_vehicle_id);

-- DRIVER BEHAVIOR (Analytics)
CREATE TABLE IF NOT EXISTS driver_behavior (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id),
    org_id UUID NOT NULL,
    period_start TIMESTAMP WITH TIME ZONE NOT NULL,
    period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    score DOUBLE PRECISION,
    safety_score DOUBLE PRECISION,
    efficiency_score DOUBLE PRECISION,
    compliance_score DOUBLE PRECISION,
    harsh_accelerations INTEGER DEFAULT 0,
    harsh_brakings INTEGER DEFAULT 0,
    harsh_cornerings INTEGER DEFAULT 0,
    speeding_incidents INTEGER DEFAULT 0,
    phone_usage INTEGER DEFAULT 0,
    seatbelt_violations INTEGER DEFAULT 0,
    fuel_consumption DOUBLE PRECISION DEFAULT 0,
    distance_driven DOUBLE PRECISION DEFAULT 0,
    idle_time INTEGER DEFAULT 0,
    eco_driving_score DOUBLE PRECISION,
    hours_of_service DOUBLE PRECISION,
    rest_periods INTEGER,
    document_compliance BOOLEAN DEFAULT TRUE,
    maintenance_adherence BOOLEAN DEFAULT TRUE,
    night_driving_hours DOUBLE PRECISION DEFAULT 0,
    highway_percentage DOUBLE PRECISION DEFAULT 0,
    city_driving_percentage DOUBLE PRECISION DEFAULT 0,
    risk_level VARCHAR(20),
    risk_factors JSONB,
    improvement_areas JSONB,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_behavior_driver_period ON driver_behavior(driver_id, period_start);

-- DRIVER INCIDENTS
CREATE TABLE IF NOT EXISTS driver_incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL REFERENCES drivers(id),
    org_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL, -- speeding, harsh_braking
    severity VARCHAR(20) NOT NULL, -- low, medium, high, critical
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    location JSONB, -- Lat, Lng, Address
    speed DOUBLE PRECISION,
    details JSONB,
    resolved BOOLEAN DEFAULT FALSE,
    resolution_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_incidents_driver ON driver_incidents(driver_id);
