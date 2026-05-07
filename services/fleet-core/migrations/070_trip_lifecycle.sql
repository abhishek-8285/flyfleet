-- Migration: 070_trip_lifecycle
-- Description: Complete trip management schema with tenant isolation and India-compliance.

-- 1. Trips table
CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

DO $$ 
BEGIN 
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL REFERENCES tenants(id);
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS vehicle_id UUID NOT NULL;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS driver_id UUID;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'CREATED';
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS title VARCHAR(255) NOT NULL;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS description TEXT;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS pickup_lat DOUBLE PRECISION;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS pickup_lng DOUBLE PRECISION;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS drop_lat DOUBLE PRECISION;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS drop_lng DOUBLE PRECISION;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS planned_distance DECIMAL(10, 2) DEFAULT 0.00;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS total_distance DECIMAL(10, 2) DEFAULT 0.00;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS scheduled_start TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS start_time TIMESTAMP WITH TIME ZONE;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS end_time TIMESTAMP WITH TIME ZONE;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS start_odometer DECIMAL(10, 2);
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS end_odometer DECIMAL(10, 2);
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS eway_bill_number VARCHAR(100);
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS gst_classification VARCHAR(100);
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS legal_hold BOOLEAN NOT NULL DEFAULT FALSE;
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    ALTER TABLE trips ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
EXCEPTION WHEN OTHERS THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_trips_tenant ON trips(tenant_id);

-- 2. Immutable Trip Events
CREATE TABLE IF NOT EXISTS trip_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    lat DOUBLE PRECISION,
    lng DOUBLE PRECISION,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Telemetry Points (Partitioned)
DO $$ 
BEGIN 
    -- If telemetry_points exists but is NOT partitioned, rename it to back it up
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'telemetry_points') AND 
       NOT EXISTS (SELECT 1 FROM pg_partitioned_table pt JOIN pg_class c ON pt.partrelid = c.oid WHERE c.relname = 'telemetry_points') THEN
        ALTER TABLE telemetry_points RENAME TO telemetry_points_legacy;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS telemetry_points (
    tenant_id UUID REFERENCES tenants(id),
    trip_id UUID,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    speed DECIMAL(5, 2),
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL
) PARTITION BY RANGE (recorded_at);

-- Create a few default partitions
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'telemetry_points_default') THEN
        CREATE TABLE telemetry_points_default PARTITION OF telemetry_points DEFAULT;
    END IF;
END $$;
