-- ============================================================================
-- FlyFleet - Complete Dev/Test Seed Data
-- ============================================================================
-- Purpose: Seed a complete test organization with user, vehicles, drivers,
--          telemetry, geofences, and events for development & testing.
-- Usage:   psql -U <user> -d fleet_db -f seed_complete_dev_data.sql
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. ORGANIZATION
-- ============================================================================
INSERT INTO organizations (id, name)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'Delta Logistics Pvt Ltd')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 2. USERS (IAM — create table if missing)
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organizations(id),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(255),
    role          VARCHAR(50) DEFAULT 'dispatcher',
    is_active     BOOLEAN DEFAULT true,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Admin user (password: "Admin@123" — bcrypt hash)
INSERT INTO users (id, org_id, email, password_hash, name, role)
VALUES (
    '550e8400-e29b-41d4-a716-446655440001'::uuid,
    '550e8400-e29b-41d4-a716-446655440000',
    'admin@delta.io',
    '$2a$04$WOojoGDL.4DzIGNd3opDV.phzq0f8tFT8pEQXk2StJ7YvyGPDBahq',
    'Admin User',
    'admin'
) ON CONFLICT (id) DO NOTHING;

-- Dispatcher user
INSERT INTO users (id, org_id, email, password_hash, name, role)
VALUES (
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    '550e8400-e29b-41d4-a716-446655440000',
    'dispatcher@delta.io',
    '$2a$04$WOojoGDL.4DzIGNd3opDV.phzq0f8tFT8pEQXk2StJ7YvyGPDBahq',
    'Rajesh Dispatcher',
    'dispatcher'
) ON CONFLICT (id) DO NOTHING;

-- Fleet manager user
INSERT INTO users (id, org_id, email, password_hash, name, role)
VALUES (
    '550e8400-e29b-41d4-a716-446655440003'::uuid,
    '550e8400-e29b-41d4-a716-446655440000',
    'manager@delta.io',
    '$2a$04$WOojoGDL.4DzIGNd3opDV.phzq0f8tFT8pEQXk2StJ7YvyGPDBahq',
    'Priya Manager',
    'fleet_manager'
) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 3. DRIVERS
-- ============================================================================
INSERT INTO drivers (id, name, phone) VALUES
    ('550e8400-e29b-41d4-a716-446655440010', 'Rajesh Kumar',    '+91 98765 43210'),
    ('550e8400-e29b-41d4-a716-446655440011', 'Amit Sharma',     '+91 98765 43211'),
    ('550e8400-e29b-41d4-a716-446655440012', 'Suresh Patel',    '+91 98765 43212'),
    ('550e8400-e29b-41d4-a716-446655440013', 'Vikram Singh',    '+91 98765 43213'),
    ('550e8400-e29b-41d4-a716-446655440014', 'Deepak Yadav',    '+91 98765 43214')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 4. VEHICLES (5 vehicles, different statuses)
-- ============================================================================
INSERT INTO vehicles (id, tenant_id, org_id, registration, make, model, status, current_driver_id) VALUES
    -- Active, moving vehicle
    ('550e8400-e29b-41d4-a716-446655440001'::uuid, '550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'DL01AB1234', 'Tata',       'Ace Gold',    'active', '550e8400-e29b-41d4-a716-446655440010'),
    -- Active, idle vehicle
    ('550e8400-e29b-41d4-a716-446655440002'::uuid, '550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'MH02CD5678', 'Ashok Leyland','Partner',    'active', '550e8400-e29b-41d4-a716-446655440011'),
    -- Active, stopped vehicle
    ('550e8400-e29b-41d4-a716-446655440003'::uuid, '550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'KA03EF9012', 'Eicher',     'Pro 1000',    'active', '550e8400-e29b-41d4-a716-446655440012'),
    -- Active, no driver
    ('550e8400-e29b-41d4-a716-446655440004'::uuid, '550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'TN04GH3456', 'BharatBenz', '1212R',      'active', NULL),
    -- Active, offline vehicle
    ('550e8400-e29b-41d4-a716-446655440005'::uuid, '550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'GJ05JK7890', 'Force',      'Traveller',  'active', '550e8400-e29b-41d4-a716-446655440014')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 5. TELEMETRY DATA (historical + recent)
-- ============================================================================

-- Clear existing telemetry for clean seed
DELETE FROM telemetry_points WHERE vehicle_id IN (
    '550e8400-e29b-41d4-a716-446655440001'::uuid,
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    '550e8400-e29b-41d4-a716-446655440003'::uuid,
    '550e8400-e29b-41d4-a716-446655440004'::uuid,
    '550e8400-e29b-41d4-a716-446655440005'::uuid
);

-- Telemetry: Vehicle 1 (DL01AB1234) — MOVING: Delhi → Mumbai route (last 2 hours)
INSERT INTO telemetry_points (vehicle_id, latitude, longitude, speed_kmh, heading, ignition, fuel_level, time)
SELECT
    '550e8400-e29b-41d4-a716-446655440001'::uuid::uuid,
    28.6139 + (gs * 0.001),          -- lat drifting south
    77.2090 - (gs * 0.0008),         -- lng drifting west
    40 + (gs % 20),                   -- speed 40-60 km/h
    135 + (gs % 10),                  -- heading SE
    true,                             -- ignition on
    72.5 - (gs * 0.05),              -- fuel slowly decreasing
    NOW() - INTERVAL '2 hours' + (gs * INTERVAL '1 minute')
FROM generate_series(0, 119) AS gs;

-- Vehicle 2 (MH02CD5678) — IDLE: parked near Mumbai depot (last 30 min)
INSERT INTO telemetry_points (vehicle_id, latitude, longitude, speed_kmh, heading, ignition, fuel_level, time)
SELECT
    '550e8400-e29b-41d4-a716-446655440002'::uuid::uuid,
    19.0760 + (random() * 0.0001),   -- slight GPS jitter
    72.8777 + (random() * 0.0001),
    CASE WHEN gs < 10 THEN 5 ELSE 0 END,  -- slowing to idle
    90,
    true,                             -- ignition on but not moving
    85.0 - (gs * 0.01),
    NOW() - INTERVAL '30 minutes' + (gs * INTERVAL '1 minute')
FROM generate_series(0, 29) AS gs;

-- Vehicle 3 (KA03EF9012) — STOPPED: engine off in Bangalore (last 1 hour)
INSERT INTO telemetry_points (vehicle_id, latitude, longitude, speed_kmh, heading, ignition, fuel_level, time)
SELECT
    '550e8400-e29b-41d4-a716-446655440003'::uuid::uuid,
    12.9716,
    77.5946,
    0,
    45,
    false,                            -- ignition off
    60.0,
    NOW() - INTERVAL '1 hour' + (gs * INTERVAL '5 minutes')
FROM generate_series(0, 11) AS gs;

-- Vehicle 4 (TN04GH3456) — No recent telemetry (driverless, last seen 2 days ago)
INSERT INTO telemetry_points (vehicle_id, latitude, longitude, speed_kmh, heading, ignition, fuel_level, time)
VALUES (
    '550e8400-e29b-41d4-a716-446655440004'::uuid::uuid,
    13.0827,
    80.2707,
    0,
    0,
    false,
    45.0,
    NOW() - INTERVAL '2 days'
);

-- Vehicle 5 (GJ05JK7890) — MOVING: Gujarat highway (last 45 min)
INSERT INTO telemetry_points (vehicle_id, latitude, longitude, speed_kmh, heading, ignition, fuel_level, time)
SELECT
    '550e8400-e29b-41d4-a716-446655440005'::uuid::uuid,
    23.0225 + (gs * 0.0005),
    72.5714 + (gs * 0.0003),
    65 + (gs % 15),
    270 + (gs % 5),
    true,
    40.0 - (gs * 0.1),
    NOW() - INTERVAL '45 minutes' + (gs * INTERVAL '1 minute')
FROM generate_series(0, 44) AS gs;

-- ============================================================================
-- 6. GEOFENCES (4 geofences across India)
-- ============================================================================
DELETE FROM geofences WHERE org_id = '550e8400-e29b-41d4-a716-446655440000';

INSERT INTO geofences (id, org_id, name, description, type, center_lat, center_lng, radius_m, enabled) VALUES
    ('550e8400-e29b-41d4-a716-446655440100', '550e8400-e29b-41d4-a716-446655440000', 'Delhi Warehouse',   'Main distribution center',  'circle', 28.6139, 77.2090, 500,   true),
    ('550e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440000', 'Mumbai Depot',       'Western hub',                'circle', 19.0760, 72.8777, 300,   true),
    ('550e8400-e29b-41d4-a716-446655440102', '550e8400-e29b-41d4-a716-446655440000', 'Bangalore Office',   'Southern office',            'circle', 12.9716, 77.5946, 200,   true),
    ('550e8400-e29b-41d4-a716-446655440103', '550e8400-e29b-41d4-a716-446655440000', 'Restricted Zone',    'No-entry area',              'circle', 23.0225, 72.5714, 1000,  true)
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 7. GEOFENCE ASSIGNMENTS
-- ============================================================================
DELETE FROM geofence_assignments WHERE vehicle_id IN (
    '550e8400-e29b-41d4-a716-446655440001'::uuid,
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    '550e8400-e29b-41d4-a716-446655440003'::uuid,
    '550e8400-e29b-41d4-a716-446655440004'::uuid,
    '550e8400-e29b-41d4-a716-446655440005'::uuid
);

INSERT INTO geofence_assignments (geofence_id, vehicle_id, assignment_type, dwell_minutes, is_active) VALUES
    -- Vehicle 1: Delhi Warehouse (enter/exit alerts)
    ('550e8400-e29b-41d4-a716-446655440100', '550e8400-e29b-41d4-a716-446655440001'::uuid, 'both', NULL, true),
    -- Vehicle 1: Mumbai Depot (enter/exit alerts)
    ('550e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440001'::uuid, 'both', NULL, true),
    -- Vehicle 2: Mumbai Depot (enter/exit + dwell after 15 min)
    ('550e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440002'::uuid, 'both', 15, true),
    -- Vehicle 3: Bangalore Office (enter/exit)
    ('550e8400-e29b-41d4-a716-446655440102', '550e8400-e29b-41d4-a716-446655440003'::uuid, 'both', NULL, true),
    -- Vehicle 5: Restricted Zone (dwell alert after 10 min)
    ('550e8400-e29b-41d4-a716-446655440103', '550e8400-e29b-41d4-a716-446655440005'::uuid, 'dwell', 10, true)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 8. GEOFENCE EVENTS (historical events for testing alerts)
-- ============================================================================
DELETE FROM geofence_events WHERE vehicle_id IN (
    '550e8400-e29b-41d4-a716-446655440001'::uuid,
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    '550e8400-e29b-41d4-a716-446655440003'::uuid
);

INSERT INTO geofence_events (time, vehicle_id, geofence_id, event_type, latitude, longitude, duration_minutes, alert_sent, alert_channels) VALUES
    -- Vehicle 1 entered Delhi Warehouse 3 hours ago
    (NOW() - INTERVAL '3 hours', '550e8400-e29b-41d4-a716-446655440001'::uuid, '550e8400-e29b-41d4-a716-446655440100', 'enter',  28.6139, 77.2090, NULL, true, '{"dashboard","email"}'),
    -- Vehicle 1 exited Delhi Warehouse 2.5 hours ago
    (NOW() - INTERVAL '2.5 hours', '550e8400-e29b-41d4-a716-446655440001'::uuid, '550e8400-e29b-41d4-a716-446655440100', 'exit',   28.6145, 77.2085, 30, true, '{"dashboard"}'),
    -- Vehicle 2 entered Mumbai Depot 1 hour ago
    (NOW() - INTERVAL '1 hour', '550e8400-e29b-41d4-a716-446655440002'::uuid, '550e8400-e29b-41d4-a716-446655440101', 'enter',  19.0760, 72.8777, NULL, true, '{"dashboard","sms"}'),
    -- Vehicle 2 dwell alert at Mumbai Depot (45 min)
    (NOW() - INTERVAL '15 minutes', '550e8400-e29b-41d4-a716-446655440002'::uuid, '550e8400-e29b-41d4-a716-446655440101', 'dwell',  19.0761, 72.8778, 45, true, '{"dashboard"}'),
    -- Vehicle 3 entered Bangalore Office 2 hours ago
    (NOW() - INTERVAL '2 hours', '550e8400-e29b-41d4-a716-446655440003'::uuid, '550e8400-e29b-41d4-a716-446655440102', 'enter',  12.9716, 77.5946, NULL, true, '{"dashboard"}')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 9. GPS LOCATIONS (upsert table for latest positions)
-- ============================================================================
INSERT INTO gps_locations (vehicle_id, latitude, longitude, speed_kmh, heading, ignition, fuel_level, odometer_km, last_seen, status, tenant_id)
VALUES
    ('550e8400-e29b-41d4-a716-446655440001'::uuid, 28.7339, 77.1130, 55.2, 140, true,  70.0, 45300.5, NOW() - INTERVAL '2 minutes',  'online', '550e8400-e29b-41d4-a716-446655440000'),
    ('550e8400-e29b-41d4-a716-446655440002'::uuid, 19.0761, 72.8778, 0.0,   90,  true,  84.5, 32100.0, NOW() - INTERVAL '5 minutes',  'online', '550e8400-e29b-41d4-a716-446655440000'),
    ('550e8400-e29b-41d4-a716-446655440003'::uuid, 12.9716, 77.5946, 0.0,   45,  false, 60.0, 28900.0, NOW() - INTERVAL '60 minutes', 'online', '550e8400-e29b-41d4-a716-446655440000'),
    ('550e8400-e29b-41d4-a716-446655440005'::uuid, 23.0445, 72.5848, 72.5,  275, true,  38.0, 51200.0, NOW() - INTERVAL '1 minute',   'online', '550e8400-e29b-41d4-a716-446655440000')
ON CONFLICT (vehicle_id) DO UPDATE SET
    latitude     = EXCLUDED.latitude,
    longitude    = EXCLUDED.longitude,
    speed_kmh    = EXCLUDED.speed_kmh,
    heading      = EXCLUDED.heading,
    ignition     = EXCLUDED.ignition,
    fuel_level   = EXCLUDED.fuel_level,
    odometer_km  = EXCLUDED.odometer_km,
    last_seen    = EXCLUDED.last_seen,
    status       = EXCLUDED.status,
    updated_at   = NOW();

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================
\echo ''
\echo '============================================'
\echo '  SEED DATA SUMMARY'
\echo '============================================'
\echo ''
\echo 'Organization:'
SELECT id, name FROM organizations;
\echo ''
\echo 'Users (3):'
SELECT id, email, name, role FROM users ORDER BY role;
\echo ''
\echo 'Drivers (5):'
SELECT id, name, phone FROM drivers;
\echo ''
\echo 'Vehicles (5):'
SELECT v.id, v.registration, v.make || ' ' || v.model AS vehicle,
       v.status, d.name AS driver
FROM vehicles v
LEFT JOIN drivers d ON v.current_driver_id = d.id
WHERE v.tenant_id = '550e8400-e29b-41d4-a716-446655440000'
ORDER BY v.registration;
\echo ''
\echo 'Telemetry Points:'
SELECT vehicle_id, COUNT(*) AS points,
       MIN(time) AS earliest, MAX(time) AS latest
FROM telemetry_points
GROUP BY vehicle_id
ORDER BY vehicle_id;
\echo ''
\echo 'Geofences (4):'
SELECT id, name, type, center_lat, center_lng, radius_m
FROM geofences WHERE org_id = '550e8400-e29b-41d4-a716-446655440000';
\echo ''
\echo 'Geofence Assignments:'
SELECT ga.vehicle_id, g.name AS geofence, ga.assignment_type, ga.dwell_minutes
FROM geofence_assignments ga
JOIN geofences g ON ga.geofence_id = g.id
ORDER BY ga.vehicle_id;
\echo ''
\echo 'Geofence Events (5):'
SELECT time, vehicle_id, g.name AS geofence, event_type, duration_minutes, alert_sent
FROM geofence_events ge
JOIN geofences g ON ge.geofence_id = g.id
ORDER BY time DESC;
\echo ''
\echo 'GPS Locations (latest positions):'
SELECT vehicle_id, latitude, longitude, speed_kmh,
       CASE WHEN last_seen > NOW() - INTERVAL '5 minutes' THEN 'ONLINE'
            ELSE 'OFFLINE' END AS status
FROM gps_locations
ORDER BY vehicle_id;
\echo ''
\echo '============================================'
\echo '  JWT Token for Testing'
\echo '============================================'
\echo ''
\echo 'Run this Node.js command to generate a valid JWT:'
\echo ''
\echo 'node -e "'
\echo 'const c=require('\''crypto'\'');'
\echo 'const b=d=>Buffer.from(JSON.stringify(d)).toString('\''base64url'\'');'
\echo 'const h=b({alg:'\''HS256'\'',type:'\''JWT'\''});'
\echo 'const p=b({sub:'\''550e8400-e29b-41d4-a716-446655440001'::uuid\'',org_id:'\''550e8400-e29b-41d4-a716-446655440000'\'',roles:['\''admin'\''],exp:Math.floor(Date.now()/1000)+3600});'
\echo 'const s=c.createHmac('\''sha256'\'','\''flyfleet-super-secret-jwt-key-32chars'\'').update(h+'\''.'\'+'+p).digest('\''base64url'\'');'
\echo 'console.log(h+'\''.'\'+'+p+'\''.'\'+'+s);"'
\echo ''
\echo '============================================'

COMMIT;
