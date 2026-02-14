-- Seed Full Flow Data (Trips, Alerts, Rules) - Fixed Schema

-- Target Org: Delta Logistics (00000000-0000-0000-0000-000000000012)
-- Vehicle: DLT-001
-- Driver: driver1@delta.io

-- 1. Create Alert Rules for Delta
INSERT INTO alert_rules (id, org_id, name, type, description, conditions, severity, actions, enabled, created_at, updated_at, created_by)
SELECT 
    gen_random_uuid(),
    '00000000-0000-0000-0000-000000000012',
    'Max Speed Limit',
    'overspeed',
    'Alert when speed exceeds 80 km/h',
    '{"threshold": 80}',
    'critical',
    '[{"type": "email", "target": "admin@delta.io", "delay": 0}]',
    true,
    NOW(),
    NOW(),
    (SELECT id FROM users WHERE email = 'admin@delta.io')
WHERE EXISTS (SELECT 1 FROM organizations WHERE id = '00000000-0000-0000-0000-000000000012')
AND NOT EXISTS (SELECT 1 FROM alert_rules WHERE org_id = '00000000-0000-0000-0000-000000000012' AND type = 'overspeed');

-- 2. Create Alerts for Delta Vehicles
INSERT INTO alerts (id, org_id, rule_id, type, severity, vehicle_id, driver_id, location, timestamp, details, status, created_at, updated_at)
SELECT
    gen_random_uuid(),
    '00000000-0000-0000-0000-000000000012',
    (SELECT id FROM alert_rules WHERE org_id = '00000000-0000-0000-0000-000000000012' LIMIT 1),
    'overspeed',
    'critical',
    (SELECT id FROM vehicles WHERE license_plate = 'DLT-001'),
    (SELECT id FROM drivers WHERE email = 'driver1@delta.io'), -- drivers table must have this email
    '{"latitude": 12.9716, "longitude": 77.5946, "address": "Hosur Road, Bangalore"}',
    NOW() - INTERVAL '1 hour',
    '{"message": "Vehicle DLT-001 exceeded speed limit (85 km/h)", "speed": 85, "limit": 80}',
    'new',
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '1 hour'
WHERE EXISTS (SELECT 1 FROM vehicles WHERE license_plate = 'DLT-001');

-- Ensure Driver Exists in 'drivers' table for Delta
INSERT INTO drivers (id, org_id, first_name, last_name, email, phone, license_number, license_type, license_expiry, license_authority, date_of_birth, status, experience_years, rating, address, emergency_contact)
VALUES (
    gen_random_uuid(), 
    '00000000-0000-0000-0000-000000000012', 
    'Delta', 'Driver 1', 
    'driver1@delta.io', 
    '+15550202', 
    'DL-DLT-001',
    'LMV',
    NOW() + INTERVAL '5 years',
    'RTO Bangalore',
    '1990-01-01',
    'active', 
    5, 
    4.5,
    '{}',
    '{}'
);

-- 3. Create Active Trip for DLT-001
INSERT INTO trips (
    id, organization_id, vehicle_id, driver_id, status, title, description, 
    planned_route, actual_route, waypoints, 
    estimated_distance_km, estimated_duration_min, 
    cargo_details, scheduled_start, start_time, 
    start_location, end_location, 
    fuel_consumed_liters, idling_time_minutes, 
    customer_name, customer_phone, 
    created_at, updated_at
)
SELECT
    gen_random_uuid(),
    '00000000-0000-0000-0000-000000000012',
    (SELECT id FROM vehicles WHERE license_plate = 'DLT-001'),
    (SELECT id FROM drivers WHERE email = 'driver1@delta.io'),
    'in_progress',
    'Bangalore to Mysore Logistics',
    'Auto parts delivery',
    '[
        {"sequence": 1, "location": {"latitude": 12.9716, "longitude": 77.5946, "address": "Bangalore Hub"}, "type": "pickup"},
        {"sequence": 2, "location": {"latitude": 12.2958, "longitude": 76.6394, "address": "Mysore Factory"}, "type": "dropoff"}
    ]',
    '[
         {"sequence": 1, "location": {"latitude": 12.9716, "longitude": 77.5946, "address": "Bangalore Hub"}, "type": "pickup", "status": "departed", "arrival_time": "2024-02-12T08:00:00Z"}
    ]',
    '[]', -- waypoints
    150.5,
    180, -- 3 hours
    '{"type": "Auto Parts", "weight_kg": 800, "fragile": false}',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '2 hours',
    '{"latitude": 12.9716, "longitude": 77.5946, "address": "Bangalore Hub"}',
    '{"latitude": 12.2958, "longitude": 76.6394, "address": "Mysore Factory"}',
    25.5,
    10,
    'AutoComponents Inc',
    '+919876543212',
    NOW() - INTERVAL '3 hours',
    NOW()
WHERE EXISTS (SELECT 1 FROM vehicles WHERE license_plate = 'DLT-001');

-- 4. Create Scheduled Trip for DLT-002
INSERT INTO trips (
    id, organization_id, vehicle_id, driver_id, status, title, description,
    estimated_distance_km, estimated_duration_min,
    cargo_details, scheduled_start,
    start_location, end_location,
    fuel_consumed_liters,
    customer_name, customer_phone,
    created_at, updated_at
)
SELECT
    gen_random_uuid(),
    '00000000-0000-0000-0000-000000000012',
    (SELECT id FROM vehicles WHERE license_plate = 'DLT-002'),
    NULL, -- No driver assigned yet
    'planned',
    'Inter-city Transfer',
    'Stock transfer',
    300.0,
    300,
    '{"type": "General", "weight_kg": 500}',
    NOW() + INTERVAL '2 days',
    '{"latitude": 12.9716, "longitude": 77.5946, "address": "Bangalore Hub"}',
    '{"latitude": 13.0827, "longitude": 80.2707, "address": "Chennai Hub"}',
    0.0,
    'Internal',
    '+919876543213',
    NOW(),
    NOW()
WHERE EXISTS (SELECT 1 FROM vehicles WHERE license_plate = 'DLT-002');
