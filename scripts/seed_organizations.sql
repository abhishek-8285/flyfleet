-- Seed Additional Organizations for Demo (v2)

-- Org 2: Delta Logistics (New unique entries)
INSERT INTO organizations (id, name, domain, status, default_theme)
VALUES ('00000000-0000-0000-0000-000000000012', 'Delta Logistics', 'delta.flyfleet.io', 'active', 'light')
ON CONFLICT (domain) DO NOTHING;

-- Org 3: Echo Transport
INSERT INTO organizations (id, name, domain, status, default_theme)
VALUES ('00000000-0000-0000-0000-000000000013', 'Echo Transport', 'echo.transport.com', 'active', 'dark')
ON CONFLICT (domain) DO NOTHING;

-- Org 4: Foxtrot Systems
INSERT INTO organizations (id, name, domain, status, default_theme)
VALUES ('00000000-0000-0000-0000-000000000014', 'Foxtrot Systems', 'foxtrot.systems', 'pending', 'light')
ON CONFLICT (domain) DO NOTHING;

-- Seed Users for Delta
INSERT INTO users (id, email, name, password_hash, status, verified)
VALUES 
(gen_random_uuid(), 'admin@delta.io', 'Delta Admin', '$2a$10$X7.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1', 'active', true),
(gen_random_uuid(), 'driver1@delta.io', 'Delta Driver 1', '$2a$10$X7.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1', 'active', true)
ON CONFLICT (email) DO NOTHING;

-- Link users to Delta using subquery to get Org ID by domain to be safe, or just use the fixed ID if insert succeeded
-- Using fixed ID as we expect it to succeed now.
INSERT INTO user_organizations (user_id, org_id, role)
SELECT id, '00000000-0000-0000-0000-000000000012', 'org_admin' FROM users WHERE email = 'admin@delta.io'
ON CONFLICT (user_id, org_id) DO NOTHING;

INSERT INTO user_organizations (user_id, org_id, role)
SELECT id, '00000000-0000-0000-0000-000000000012', 'driver' FROM users WHERE email = 'driver1@delta.io'
ON CONFLICT (user_id, org_id) DO NOTHING;

-- Seed Vehicles for Delta
INSERT INTO vehicles (id, org_id, license_plate, make, model, year, status, vin)
VALUES 
(gen_random_uuid(), '00000000-0000-0000-0000-000000000012', 'DLT-001', 'Tesla', 'Semi', 2024, 'active', 'VIN-DLT-001'),
(gen_random_uuid(), '00000000-0000-0000-0000-000000000012', 'DLT-002', 'Nikola', 'Tre', 2023, 'maintenance', 'VIN-DLT-002')
ON CONFLICT (license_plate) DO NOTHING;

-- Seed Vehicles for Echo
INSERT INTO vehicles (id, org_id, license_plate, make, model, year, status, vin)
VALUES 
(gen_random_uuid(), '00000000-0000-0000-0000-000000000013', 'ECHO-101', 'Rivian', 'EDV', 2024, 'active', 'VIN-ECHO-101')
ON CONFLICT (license_plate) DO NOTHING;
