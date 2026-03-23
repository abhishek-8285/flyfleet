-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    driver_id UUID,
    status VARCHAR(50) NOT NULL CHECK (status IN ('CREATED', 'SCHEDULED', 'STARTED', 'PAUSED', 'RESUMED', 'COMPLETED', 'CANCELLED', 'FAILED')),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- PostGIS Geospatial fields
    pickup_geo GEOGRAPHY(POINT, 4326),
    drop_geo GEOGRAPHY(POINT, 4326),
    
    planned_route JSONB, -- list of waypoints
    actual_route JSONB, -- recording of points
    waypoints JSONB,
    
    planned_distance DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_distance DOUBLE PRECISION,
    estimated_duration_min INTEGER NOT NULL DEFAULT 0,
    actual_duration_min INTEGER,
    
    cargo_details JSONB,
    scheduled_start TIMESTAMP WITH TIME ZONE NOT NULL,
    planned_eta TIMESTAMP WITH TIME ZONE,
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    
    start_location JSONB, -- Snapshot for non-GIS search
    end_location JSONB,
    
    start_odometer DOUBLE PRECISION,
    end_odometer DOUBLE PRECISION,
    fuel_level_snapshot DOUBLE PRECISION,
    fuel_used DOUBLE PRECISION,
    idling_time_minutes INTEGER,
    
    -- India-Specific Compliance
    eway_bill_no VARCHAR(100),
    gst_classification VARCHAR(100),
    permit_verified BOOLEAN DEFAULT FALSE,
    ais140_compliant BOOLEAN DEFAULT FALSE,

    -- Compliance & Audit (Section 9)
    legal_hold BOOLEAN NOT NULL DEFAULT FALSE,        -- 7-year archival, CERT-In compliance
    correlation_id UUID DEFAULT gen_random_uuid(),    -- Per-trip distributed tracing ID

    customer_name VARCHAR(255),
    customer_phone VARCHAR(50),
    customer_email VARCHAR(255),
    cancellation_reason TEXT,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trips_vehicle_status ON trips(vehicle_id, status);
CREATE INDEX IF NOT EXISTS idx_trips_driver_status ON trips(driver_id, status);
CREATE INDEX IF NOT EXISTS idx_trips_tenant_start ON trips(tenant_id, start_time);
CREATE INDEX IF NOT EXISTS idx_trips_pickup_geo ON trips USING GIST(pickup_geo);
CREATE INDEX IF NOT EXISTS idx_trips_drop_geo ON trips USING GIST(drop_geo);

-- Immutable Trip Events with Hash Chaining for Audit (Section 9)
CREATE TABLE IF NOT EXISTS trip_events (
    id UUID PRIMARY KEY,
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    location GEOGRAPHY(POINT, 4326),                 -- AIS-140 event geo-location
    payload JSONB NOT NULL DEFAULT '{}',
    hash TEXT NOT NULL,
    prev_hash TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trip_events_trip_id ON trip_events(trip_id);
CREATE INDEX IF NOT EXISTS idx_trip_events_location ON trip_events USING GIST(location);
CREATE INDEX IF NOT EXISTS idx_trips_legal_hold ON trips(legal_hold) WHERE legal_hold = TRUE;
CREATE INDEX IF NOT EXISTS idx_trips_correlation_id ON trips(correlation_id);

-- Telemetry Points (Partitioned Monthly)
CREATE TABLE IF NOT EXISTS telemetry_points (
    trip_id UUID NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    speed DOUBLE PRECISION,
    heading DOUBLE PRECISION,
    fuel_level DOUBLE PRECISION,
    ignition_status BOOLEAN,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL
) PARTITION BY RANGE (recorded_at);

-- Example partition for March 2026
CREATE TABLE IF NOT EXISTS telemetry_points_2026_03 PARTITION OF telemetry_points
    FOR VALUES FROM ('2026-03-01 00:00:00') TO ('2026-04-01 00:00:00');

CREATE INDEX IF NOT EXISTS idx_telemetry_trip_time ON telemetry_points(trip_id, recorded_at);
CREATE INDEX IF NOT EXISTS idx_telemetry_geo ON telemetry_points USING GIST (ST_SetSRID(ST_MakePoint(lng, lat), 4326));

CREATE TABLE IF NOT EXISTS trip_offers (
    id UUID PRIMARY KEY,
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    driver_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    rejection_reason TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    responded_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trip_offers_trip_id ON trip_offers(trip_id);
CREATE INDEX IF NOT EXISTS idx_trip_offers_driver_id ON trip_offers(driver_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS trip_offers;
DROP TABLE IF EXISTS telemetry_points;
DROP TABLE IF EXISTS trip_events;
DROP TABLE IF EXISTS trips;
-- +goose StatementEnd
