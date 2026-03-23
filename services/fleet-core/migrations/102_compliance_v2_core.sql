-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS duty_status_logs (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    odometer_start DOUBLE PRECISION,
    odometer_end DOUBLE PRECISION,
    engine_hours_start DOUBLE PRECISION,
    engine_hours_end DOUBLE PRECISION,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    location_description TEXT,
    annotation TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS hos_violations (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    vehicle_id UUID,
    violation_type VARCHAR(100) NOT NULL,
    violation_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration INTEGER,
    resolved BOOLEAN DEFAULT false NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS driver_certifications (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    log_date DATE NOT NULL,
    certified BOOLEAN DEFAULT false NOT NULL,
    signature TEXT,
    certified_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(driver_id, log_date)
);

CREATE TABLE IF NOT EXISTS eld_events (
    event_id UUID PRIMARY KEY,
    sequence_id BIGSERIAL NOT NULL,
    driver_id UUID,
    vehicle_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    location TEXT,
    odometer DOUBLE PRECISION,
    engine_hours DOUBLE PRECISION,
    origin VARCHAR(50),
    checksum TEXT,
    UNIQUE(vehicle_id, sequence_id)
);

CREATE TABLE IF NOT EXISTS eld_malfunctions (
    id UUID PRIMARY KEY,
    device_id UUID NOT NULL,
    malfunction_type VARCHAR(100) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    resolved_time TIMESTAMP WITH TIME ZONE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS unidentified_driver_events (
    id UUID PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    event_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location TEXT,
    odometer DOUBLE PRECISION,
    assigned_driver_id UUID
);

CREATE TABLE IF NOT EXISTS dvir_reports (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    vehicle_id UUID NOT NULL,
    inspection_type VARCHAR(50) NOT NULL,
    condition VARCHAR(50) NOT NULL,
    notes TEXT,
    signature TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS dvir_defects (
    id UUID PRIMARY KEY,
    dvir_id UUID NOT NULL REFERENCES dvir_reports(id) ON DELETE CASCADE,
    defect_type VARCHAR(100) NOT NULL,
    description TEXT,
    resolved BOOLEAN DEFAULT false NOT NULL,
    resolved_by UUID,
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS dvir_defects;
DROP TABLE IF EXISTS dvir_reports;
DROP TABLE IF EXISTS unidentified_driver_events;
DROP TABLE IF EXISTS eld_malfunctions;
DROP TABLE IF EXISTS eld_events;
DROP TABLE IF EXISTS driver_certifications;
DROP TABLE IF EXISTS hos_violations;
DROP TABLE IF EXISTS duty_status_logs;

-- +goose StatementEnd
