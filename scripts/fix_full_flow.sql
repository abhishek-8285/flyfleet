-- Fix Missing Driver and Link to Trips/Alerts

-- 1. Insert Driver for Delta (using WHERE NOT EXISTS to be safe)
INSERT INTO drivers (id, org_id, first_name, last_name, email, phone, license_number, status, experience_years, rating)
SELECT
    gen_random_uuid(), 
    '00000000-0000-0000-0000-000000000012', 
    'Delta', 'Driver 1', 
    'driver1@delta.io', 
    '+15550202', 
    'DL-DLT-001', 
    'active', 
    5, 
    4.5
WHERE NOT EXISTS (SELECT 1 FROM drivers WHERE email = 'driver1@delta.io');

-- 2. Update Active Trip for DLT-001 with this driver
UPDATE trips
SET driver_id = (SELECT id FROM drivers WHERE email = 'driver1@delta.io')
WHERE vehicle_id = (SELECT id FROM vehicles WHERE license_plate = 'DLT-001')
AND status = 'in_progress';

-- 3. Update Alerts for DLT-001/driver1
UPDATE alerts
SET driver_id = (SELECT id FROM drivers WHERE email = 'driver1@delta.io')
WHERE vehicle_id = (SELECT id FROM vehicles WHERE license_plate = 'DLT-001')
AND driver_id IS NULL;
