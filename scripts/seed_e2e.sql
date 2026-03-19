-- Comprehensive E2E Seeding Script
-- Standardizes Organizations, Roles, and Users for E2E Verification

-- 1. Ensure Organizations exist
INSERT INTO organizations (id, name, domain, status) VALUES 
('00000000-0000-0000-0000-000000000012', 'Delta Logistics', 'delta.io', 'active'),
('00000000-0000-0000-0000-000000000013', 'Echo Transport', 'echo.io', 'active'),
('00000000-0000-0000-0000-000000000001', 'FlyFleet Global', 'flyfleet.com', 'active')
ON CONFLICT (id) DO UPDATE SET status = 'active';

-- 2. Ensure all Roles exist
INSERT INTO roles (name, description, is_system) VALUES 
('super_admin', 'Full platform access', TRUE),
('org_admin', 'Organization level administrator', TRUE),
('fleet_manager', 'Manages vehicles and drivers', TRUE),
('dispatcher', 'Daily trip execution and route planning', TRUE),
('fleet_operator', 'Real-time monitoring and alert acknowledgment', TRUE),
('maintenance_manager', 'Vehicle health and service management', TRUE),
('analyst', 'Performance reporting and data analytics', TRUE),
('support_agent', 'Platform support and troubleshooting', TRUE),
('billing_admin', 'Invoices and subscription management', TRUE),
('driver', 'Driver app access only', TRUE)
ON CONFLICT (name) DO NOTHING;

-- 3. Cleanup existing test users to avoid conflicts
DELETE FROM user_organizations WHERE org_id IN ('00000000-0000-0000-0000-000000000012', '00000000-0000-0000-0000-000000000013', '00000000-0000-0000-0000-000000000001');
DELETE FROM users WHERE email IN (
    'admin@flyfleet.com', 'org_admin@delta.io', 'fleet_manager@delta.io', 
    'dispatcher@delta.io', 'operator@delta.io', 'maintenance@delta.io', 
    'analyst@delta.io', 'support_agent@delta.io', 'billing@delta.io', 
    'driver@delta.io'
);

-- 4. Create Users (Password: fleet_password)
WITH inserted_users AS (
    INSERT INTO users (email, password_hash, role_id, org_id)
    SELECT 
        email, '$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2', (SELECT id FROM roles WHERE name = r_name),
        CASE WHEN email = 'admin@flyfleet.com' THEN '00000000-0000-0000-0000-000000000001'::UUID ELSE '00000000-0000-0000-0000-000000000012'::UUID END
    FROM (
        VALUES 
            ('admin@flyfleet.com', 'super_admin'),
            ('org_admin@delta.io', 'org_admin'),
            ('fleet_manager@delta.io', 'fleet_manager'),
            ('dispatcher@delta.io', 'dispatcher'),
            ('operator@delta.io', 'fleet_operator'),
            ('maintenance@delta.io', 'maintenance_manager'),
            ('analyst@delta.io', 'analyst'),
            ('support_agent@delta.io', 'support_agent'),
            ('billing@delta.io', 'billing_admin'),
            ('driver@delta.io', 'driver')
    ) AS t(email, r_name)
    RETURNING id, email, role_id
)
-- 5. Link Users to Organizations
INSERT INTO user_organizations (user_id, org_id, role, status)
SELECT 
    iu.id, 
    CASE 
        WHEN iu.email = 'admin@flyfleet.com' THEN '00000000-0000-0000-0000-000000000001'::UUID
        ELSE '00000000-0000-0000-0000-000000000012'::UUID
    END,
    (SELECT name FROM roles WHERE id = iu.role_id),
    'active'
FROM inserted_users iu;
