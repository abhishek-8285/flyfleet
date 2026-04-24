#!/bin/bash
psql -U abhishek -d fleet_db -c "CREATE TABLE IF NOT EXISTS telemetry_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(100) NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    speed DOUBLE PRECISION DEFAULT 0,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL,
    event_id UUID NOT NULL,
    tenant_id UUID NOT NULL
);"
