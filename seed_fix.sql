BEGIN;

-- 1. Insert Drivers (linked to 'Default Fleet': 550e8400-e29b-41d4-a716-446655440000)
INSERT INTO drivers (id, tenant_id, first_name, last_name, email, phone, license_number, status) VALUES
    ('550e8400-e29b-41d4-a716-446655440010', '550e8400-e29b-41d4-a716-446655440000', 'Rajesh', 'Kumar', 'rajesh@delta.io', '+91 98765 43210', 'DL-12345', 'active'),
    ('550e8400-e29b-41d4-a716-446655440011', '550e8400-e29b-41d4-a716-446655440000', 'Amit', 'Sharma', 'amit@delta.io', '+91 98765 43211', 'DL-23456', 'active'),
    ('550e8400-e29b-41d4-a716-446655440012', '550e8400-e29b-41d4-a716-446655440000', 'Suresh', 'Patel', 'suresh@delta.io', '+91 98765 43212', 'DL-34567', 'active')
ON CONFLICT (id) DO NOTHING;

-- 2. Insert Vehicles
INSERT INTO vehicles (id, tenant_id, license_plate, make, model, year, status, current_driver_id) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440000', 'DL01AB1234', 'Tata', 'Ace Gold', 2023, 'active', '550e8400-e29b-41d4-a716-446655440010'),
    ('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440000', 'MH02CD5678', 'Ashok Leyland', 'Partner', 2022, 'active', '550e8400-e29b-41d4-a716-446655440011'),
    ('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440000', 'KA03EF9012', 'Eicher', 'Pro 1000', 2023, 'active', '550e8400-e29b-41d4-a716-446655440012')
ON CONFLICT (id) DO NOTHING;

-- 3. Telemetry Points (vehicle_telemetry)
INSERT INTO vehicle_telemetry (time, vehicle_id, latitude, longitude, speed_kph, heading, fuel_level_percent)
VALUES
    (NOW() - INTERVAL '1 minute', '550e8400-e29b-41d4-a716-446655440001', 28.6139, 77.2090, 45.0, 180, 75.5),
    (NOW() - INTERVAL '1 minute', '550e8400-e29b-41d4-a716-446655440002', 19.0760, 72.8777, 0.0, 90, 85.0),
    (NOW() - INTERVAL '1 minute', '550e8400-e29b-41d4-a716-446655440003', 12.9716, 77.5946, 12.5, 45, 60.0)
ON CONFLICT DO NOTHING;

-- 4. Latest GPS Locations
INSERT INTO gps_locations (time, vehicle_id, latitude, longitude, speed, heading, ignition)
VALUES
    (NOW(), '550e8400-e29b-41d4-a716-446655440001', 28.6139, 77.2090, 45.0, 180, true),
    (NOW(), '550e8400-e29b-41d4-a716-446655440002', 19.0760, 72.8777, 0.0, 90, true),
    (NOW(), '550e8400-e29b-41d4-a716-446655440003', 12.9716, 77.5946, 12.5, 45, true)
ON CONFLICT DO NOTHING;

COMMIT;
