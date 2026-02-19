-- Telemetry Optimization Migration
-- Adds missing tables and optimizes JSONB fields

-- 1. GPS Locations (Hypertable compatible)
CREATE TABLE IF NOT EXISTS gps_locations (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    vehicle_id UUID NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    altitude DOUBLE PRECISION,
    speed DOUBLE PRECISION,
    heading DOUBLE PRECISION,
    accuracy DOUBLE PRECISION,
    ignition BOOLEAN,
    odometer DOUBLE PRECISION,
    battery_level DOUBLE PRECISION,
    satellites INTEGER,
    raw_data JSONB,
    PRIMARY KEY (time, vehicle_id)
);

-- 2. OBD Readings (Hypertable compatible)
CREATE TABLE IF NOT EXISTS obd_readings (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    vehicle_id UUID NOT NULL,
    rpm INTEGER,
    coolant_temp INTEGER,
    engine_load DECIMAL(5,2),
    throttle_pos DECIMAL(5,2),
    fuel_level DECIMAL(5,2),
    dtcs JSONB,
    vin VARCHAR(50),
    raw_data JSONB,
    PRIMARY KEY (time, vehicle_id)
);

-- 3. Fuel Readings (Hypertable compatible)
CREATE TABLE IF NOT EXISTS fuel_readings (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    vehicle_id UUID NOT NULL,
    level_percent DECIMAL(5,2),
    level_liters DECIMAL(10,2),
    refuel_event BOOLEAN DEFAULT FALSE,
    theft_event BOOLEAN DEFAULT FALSE,
    location GEOMETRY(Point, 4326),
    raw_data JSONB,
    PRIMARY KEY (time, vehicle_id)
);

-- 4. Vehicle Telemetry (Unified View or Table for Trip Service)
-- If it's a table used for aggregated data as seen in trip-service
CREATE TABLE IF NOT EXISTS vehicle_telemetry (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    vehicle_id UUID NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    speed_kph DOUBLE PRECISION,
    heading DOUBLE PRECISION,
    odometer_km DOUBLE PRECISION,
    fuel_level_percent DOUBLE PRECISION,
    engine_temp_celsius DOUBLE PRECISION,
    alerts JSONB DEFAULT '[]'::jsonb,
    PRIMARY KEY (time, vehicle_id)
);

-- 5. Optimization: GIN Indexes for JSONB fields
-- Allows fast searching within JSON data
CREATE INDEX IF NOT EXISTS idx_gps_raw_data_gin ON gps_locations USING GIN (raw_data);
CREATE INDEX IF NOT EXISTS idx_obd_raw_data_gin ON obd_readings USING GIN (raw_data);
CREATE INDEX IF NOT EXISTS idx_obd_dtcs_gin ON obd_readings USING GIN (dtcs);
CREATE INDEX IF NOT EXISTS idx_fuel_raw_data_gin ON fuel_readings USING GIN (raw_data);
CREATE INDEX IF NOT EXISTS idx_vehicle_telemetry_alerts_gin ON vehicle_telemetry USING GIN (alerts);

-- 6. Performance Indexes
CREATE INDEX IF NOT EXISTS idx_gps_vehicle_time ON gps_locations (vehicle_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_obd_vehicle_time ON obd_readings (vehicle_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_fuel_vehicle_time ON fuel_readings (vehicle_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_vehicle_telemetry_vehicle_time ON vehicle_telemetry (vehicle_id, time DESC);

-- 7. Add comments
COMMENT ON TABLE gps_locations IS 'Stores raw GPS telemetry from vehicles';
COMMENT ON TABLE obd_readings IS 'Stores On-Board Diagnostics data from vehicles';
COMMENT ON TABLE fuel_readings IS 'Stores dedicated fuel sensor data';
COMMENT ON TABLE vehicle_telemetry IS 'Aggregated telemetry data for trip reporting';
