-- Seed Full RBAC Roles and Test Users (V2 - with user_organizations)
-- This script ensures all 10 roles exist and creates test users for them.

-- 1. Insert Missing Roles
INSERT INTO roles (name, description, is_system) VALUES 
('dispatcher', 'Daily trip execution and route planning', TRUE),
('fleet_operator', 'Real-time monitoring and alert acknowledgment', TRUE),
('maintenance_manager', 'Vehicle health and service management', TRUE),
('analyst', 'Performance reporting and data analytics', TRUE),
('support_agent', 'Platform support and troubleshooting', TRUE),
('billing_admin', 'Invoices and subscription management', TRUE)
ON CONFLICT (name) DO NOTHING;

-- 2. Create Test Users
-- Delete existing test users from both tables to ensure fresh start
DELETE FROM user_organizations WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@delta.io' OR email LIKE '%@echo.io' OR email = 'support_agent@delta.io');
DELETE FROM users WHERE email IN (
    'org_admin@delta.io', 'fleet_manager@delta.io', 'dispatcher@delta.io', 
    'operator@delta.io', 'maintenance@delta.io', 'analyst@delta.io', 
    'support_agent@delta.io', 'billing@delta.io', 'driver@delta.io', 
    'org_admin@echo.io'
);

-- Insert Users into 'users' table
WITH inserted_users AS (
    INSERT INTO users (email, password_hash, role_id, status)
    SELECT 
        email, '$2b$12$Z3TNc9LZ5P0jOmnW0.O1T.CfrQ5eucQrm3Ty6RNggRMjoinOLRLd2', (SELECT id FROM roles WHERE name = r_name), 'active'
    FROM (
        VALUES 
            ('org_admin@delta.io', 'org_admin'),
            ('fleet_manager@delta.io', 'fleet_manager'),
            ('dispatcher@delta.io', 'dispatcher'),
            ('operator@delta.io', 'fleet_operator'),
            ('maintenance@delta.io', 'maintenance_manager'),
            ('analyst@delta.io', 'analyst'),
            ('support_agent@delta.io', 'support_agent'),
            ('billing@delta.io', 'billing_admin'),
            ('driver@delta.io', 'driver'),
            ('org_admin@echo.io', 'org_admin')
    ) AS t(email, r_name)
    RETURNING id, email, role_id
)
-- 3. Insert into user_organizations
INSERT INTO user_organizations (user_id, org_id, role, status)
SELECT 
    iu.id, 
    CASE 
        WHEN iu.email = 'org_admin@echo.io' THEN '00000000-0000-0000-0000-000000000013'::UUID
        ELSE '00000000-0000-0000-0000-000000000012'::UUID
    END,
    r.name,
    'active'
FROM inserted_users iu
JOIN roles r ON iu.role_id = r.id;
