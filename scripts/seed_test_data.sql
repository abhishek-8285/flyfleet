-- Seed Data for FlyFleet Testing
-- Targeted Org: FlyFleet Global (00000000-0000-0000-0000-000000000001)

-- 1. Ensure Organization exists
INSERT INTO organizations (id, name, domain, status)
VALUES ('00000000-0000-0000-0000-000000000001', 'FlyFleet Global', 'test.com', 'active')
ON CONFLICT (id) DO NOTHING;

-- 2. Seed Vehicles (Idempotent using license_plate unique constraint)
INSERT INTO vehicles (id, org_id, license_plate, make, model, year, status, fuel_type, odometer, vin)
VALUES 
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'FLD-101', 'Freightliner', 'Cascadia', 2022, 'active', 'diesel', 45000, '1F9E-101'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'FLD-102', 'Freightliner', 'M2 106', 2021, 'active', 'diesel', 12000, '1F9E-102'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'FLD-103', 'Volvo', 'VNL 860', 2023, 'maintenance', 'diesel', 5000, 'VNL-860-103'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'FLD-104', 'Kenworth', 'T680', 2022, 'active', 'diesel', 89000, 'KW-T680-104'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'FLD-105', 'Peterbilt', '579', 2022, 'active', 'diesel', 34200, 'PB-579-105'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'FLD-106', 'Mack', 'Anthem', 2021, 'inactive', 'diesel', 150000, 'MACK-ANT-106')
ON CONFLICT (license_plate) DO NOTHING;

-- 3. Seed Drivers (Idempotent using NOT EXISTS on email)
INSERT INTO drivers (id, org_id, first_name, last_name, email, phone, license_number, status, experience_years, rating)
SELECT gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'John', 'Doe', 'john.doe@test.com', '+15550101', 'DL-77221', 'active', 10, 4.8
WHERE NOT EXISTS (SELECT 1 FROM drivers WHERE email = 'john.doe@test.com');

INSERT INTO drivers (id, org_id, first_name, last_name, email, phone, license_number, status, experience_years, rating)
SELECT gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'Jane', 'Smith', 'jane.smith@test.com', '+15550102', 'DL-88332', 'active', 5, 4.9
WHERE NOT EXISTS (SELECT 1 FROM drivers WHERE email = 'jane.smith@test.com');

INSERT INTO drivers (id, org_id, first_name, last_name, email, phone, license_number, status, experience_years, rating)
SELECT gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'Robert', 'Brown', 'bob.brown@test.com', '+15550103', 'DL-99443', 'active', 12, 4.5
WHERE NOT EXISTS (SELECT 1 FROM drivers WHERE email = 'bob.brown@test.com');

INSERT INTO drivers (id, org_id, first_name, last_name, email, phone, license_number, status, experience_years, rating)
SELECT gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'Alice', 'Wilson', 'alice.w@test.com', '+15550104', 'DL-11554', 'active', 3, 5.0
WHERE NOT EXISTS (SELECT 1 FROM drivers WHERE email = 'alice.w@test.com');

-- 4. Map Users to Orgs
INSERT INTO user_organizations (user_id, org_id, role)
SELECT id, '00000000-0000-0000-0000-000000000001', 'admin'
FROM users WHERE email IN ('admin@flyfleet.com', 'test@example.com')
ON CONFLICT (user_id, org_id) DO NOTHING;

-- 5. Cross-link some drivers to vehicles
-- Assign Driver 1 to FLD-101
UPDATE vehicles v 
SET current_driver_id = (SELECT id FROM drivers d WHERE d.email = 'john.doe@test.com' LIMIT 1)
WHERE v.license_plate = 'FLD-101' AND v.current_driver_id IS NULL;

-- Assign Driver 2 to FLD-102
UPDATE vehicles v 
SET current_driver_id = (SELECT id FROM drivers d WHERE d.email = 'jane.smith@test.com' LIMIT 1)
WHERE v.license_plate = 'FLD-102' AND v.current_driver_id IS NULL;
