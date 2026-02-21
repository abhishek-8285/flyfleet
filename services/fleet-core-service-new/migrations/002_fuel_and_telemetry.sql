-- Migration for Fuel and Telemetry tables
-- This ensures the new fleet-core-service-new has the necessary persistence layer

-- Fuel Logs Table
CREATE TABLE IF NOT EXISTS fuel_logs (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    station_name VARCHAR(255),
    fuel_type VARCHAR(50),
    quantity DECIMAL(10,2),
    unit_price DECIMAL(10,2),
    total_cost DECIMAL(10,2),
    odometer DECIMAL(15,2),
    transaction_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'PENDING',
    receipt_url TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fuel_logs_org_id ON fuel_logs(org_id);
CREATE INDEX IF NOT EXISTS idx_fuel_logs_vehicle_id ON fuel_logs(vehicle_id);

-- Vehicle Telemetry Table
-- Using time as the first part of the primary key for TimescaleDB optimization
CREATE TABLE IF NOT EXISTS vehicle_telemetry (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    vehicle_id UUID NOT NULL,
    org_id UUID NOT NULL,
    device_id VARCHAR(100),
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    speed DECIMAL(6,2),
    fuel_level DECIMAL(5,2),
    battery_voltage DECIMAL(5,2),
    coolant_temp DECIMAL(5,2),
    odometer BIGINT,
    metadata JSONB DEFAULT '{}'::jsonb,
    PRIMARY KEY (time, vehicle_id)
);

CREATE INDEX IF NOT EXISTS idx_vehicle_telemetry_vehicle_id ON vehicle_telemetry(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_telemetry_org_id ON vehicle_telemetry(org_id);
