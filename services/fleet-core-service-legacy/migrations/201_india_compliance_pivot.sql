-- +goose Up
-- +goose StatementBegin

-- 1. Vehicle Mandatory Documents (India specific)
CREATE TABLE IF NOT EXISTS vehicle_documents (
    id UUID PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    document_type VARCHAR(100) NOT NULL, -- RC, INSURANCE, PUC, FITNESS, PERMIT
    document_number TEXT NOT NULL,
    issue_date DATE,
    expiry_date DATE NOT NULL,
    file_url TEXT,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    verified BOOLEAN DEFAULT false,
    verified_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 2. Driver Documents
CREATE TABLE IF NOT EXISTS driver_documents (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    document_type VARCHAR(100) NOT NULL, -- DRIVING_LICENSE, AADHAAR, PAN, MEDICAL_CERTIFICATE
    document_number TEXT NOT NULL,
    issue_date DATE,
    expiry_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    file_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 3. Trip Sheets (Indian fleet workflow)
CREATE TABLE IF NOT EXISTS trip_sheets (
    id UUID PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    trip_start TIMESTAMP WITH TIME ZONE,
    trip_end TIMESTAMP WITH TIME ZONE,
    origin TEXT,
    destination TEXT,
    status VARCHAR(50) DEFAULT 'CREATED', -- CREATED, ASSIGNED, IN_PROGRESS, COMPLETED, CLOSED
    distance_km DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS trip_events (
    id UUID PRIMARY KEY,
    trip_id UUID NOT NULL REFERENCES trip_sheets(id),
    event_type VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    location TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 4. AIS-140 Device Management
CREATE TABLE IF NOT EXISTS ais140_devices (
    id UUID PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    device_imei TEXT NOT NULL UNIQUE,
    sim_number TEXT NOT NULL,
    device_model TEXT,
    firmware_version TEXT,
    status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, OFFLINE, TAMPERED, DECOMMISSIONED
    last_seen TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS device_health_logs (
    id UUID PRIMARY KEY,
    device_id UUID NOT NULL REFERENCES ais140_devices(id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    battery_voltage DOUBLE PRECISION,
    gps_signal VARCHAR(50),
    network_signal VARCHAR(50),
    temperature DOUBLE PRECISION,
    status VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 5. Panic Alerts (AIS-140)
CREATE TABLE IF NOT EXISTS panic_alerts (
    id UUID PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    driver_id UUID,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    status VARCHAR(50) DEFAULT 'ACTIVE', -- ACTIVE, IN_PROGRESS, RESOLVED
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS panic_alert_actions (
    id UUID PRIMARY KEY,
    alert_id UUID NOT NULL REFERENCES panic_alerts(id),
    action TEXT NOT NULL,
    actor UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 6. Indian Driver Duty Logs (MoRTH 8h/12h rules)
-- Replaces FMCSA duty_status_logs
CREATE TABLE IF NOT EXISTS driver_duty_logs (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    vehicle_id UUID,
    status VARCHAR(50) NOT NULL, -- OFF_DUTY, DRIVING, ON_DUTY, BREAK, REST
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    location TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS driver_break_logs (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration_minutes INTEGER,
    break_type VARCHAR(50) NOT NULL, -- REST, MEAL, MANDATORY_BREAK
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS driver_compliance_violations (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL,
    vehicle_id UUID,
    violation_type VARCHAR(100) NOT NULL, -- DRIVING_LIMIT_EXCEEDED, DUTY_LIMIT_EXCEEDED, BREAK_REQUIRED, WEEKLY_LIMIT_EXCEEDED
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    location TEXT,
    severity VARCHAR(50) NOT NULL, -- LOW, MEDIUM, HIGH, CRITICAL
    resolved BOOLEAN DEFAULT false,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 7. FASTag Toll Tracking
CREATE TABLE IF NOT EXISTS toll_transactions (
    id UUID PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    fastag_id TEXT NOT NULL,
    toll_plaza TEXT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    location TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS fastag_reconciliation (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES toll_transactions(id),
    vehicle_id UUID NOT NULL,
    expected_amount DOUBLE PRECISION,
    actual_amount DOUBLE PRECISION,
    status VARCHAR(50) DEFAULT 'PENDING', -- MATCHED, MISMATCH, MISSING
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 8. State Border Monitoring
CREATE TABLE IF NOT EXISTS state_border_crossings (
    id UUID PRIMARY KEY,
    vehicle_id UUID NOT NULL,
    from_state TEXT NOT NULL,
    to_state TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    gps_location TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 9. Compliance Alerts & Audit
CREATE TABLE IF NOT EXISTS compliance_alerts (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    alert_type VARCHAR(100) NOT NULL, -- DRIVER_VIOLATION, DOCUMENT_EXPIRING, DOCUMENT_EXPIRED, PANIC_ALERT, DEVICE_TAMPER, DEVICE_OFFLINE
    entity_id UUID NOT NULL,
    severity VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS compliance_audit_logs (
    id UUID PRIMARY KEY,
    actor_id UUID NOT NULL,
    action TEXT NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS compliance_audit_logs;
DROP TABLE IF EXISTS compliance_alerts;
DROP TABLE IF EXISTS state_border_crossings;
DROP TABLE IF EXISTS fastag_reconciliation;
DROP TABLE IF EXISTS toll_transactions;
DROP TABLE IF EXISTS driver_compliance_violations;
DROP TABLE IF EXISTS driver_break_logs;
DROP TABLE IF EXISTS driver_duty_logs;
DROP TABLE IF EXISTS panic_alert_actions;
DROP TABLE IF EXISTS panic_alerts;
DROP TABLE IF EXISTS device_health_logs;
DROP TABLE IF EXISTS ais140_devices;
DROP TABLE IF EXISTS trip_events;
DROP TABLE IF EXISTS trip_sheets;
DROP TABLE IF EXISTS driver_documents;
DROP TABLE IF EXISTS vehicle_documents;


-- +goose StatementEnd
