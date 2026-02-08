-- Create telemetry table for real-time trip tracking
CREATE TABLE IF NOT EXISTS telemetry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID REFERENCES trips(id) ON DELETE CASCADE,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    speed DECIMAL(6,2),
    heading DECIMAL(5,2),
    event_type VARCHAR(50),
    event_time TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_telemetry_trip_id ON telemetry(trip_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_event_time ON telemetry(event_time);
