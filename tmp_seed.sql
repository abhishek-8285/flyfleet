-- 1. Default organization
INSERT INTO organizations (id, name)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    'Default Fleet'
)
ON CONFLICT (id) DO NOTHING;

-- 2. Default RBAC roles
INSERT INTO roles (id, name, description, is_system)
VALUES
    ('e349ce46-2c59-4c43-8338-afc37d01b88e', 'Super Admin', 'Full system access', true),
    ('819ac656-6648-41bf-9dec-c38906264adc', 'Fleet Manager', 'Manage vehicles, drivers, trips', true),
    ('215779f2-b515-4bca-9c73-de2ac1e3f047', 'Dispatcher', 'Manage trips and routing', true),
    ('857ad16f-1021-415b-b141-9bdc5b08b942', 'Driver', 'View assigned trips and vehicles', true),
    ('d3a2b1c0-0000-0000-0000-000000000005', 'Viewer', 'Read-only access', true)
ON CONFLICT (id) DO NOTHING;

-- 3. Default super-admin user (password: admin123, bcrypt cost 10)
INSERT INTO users (id, email, password_hash, name, status, email_verified, role_id, permission_mask)
VALUES (
    '550e8400-e29b-41d4-a716-446655440001',
    'admin@flyfleet.io',
    '$2a$10$B2/XZniBnqEmecEhQGzUn.9ZU5sNpthotIAmxEJcWqGEkodD7Jp9i',
    'Admin User',
    'active',
    true,
    'e349ce46-2c59-4c43-8338-afc37d01b88e',
    9223372036854775807
)
ON CONFLICT (id) DO NOTHING;

-- 4. Link admin to organization
INSERT INTO user_organizations (user_id, org_id, role, status)
VALUES (
    '550e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440000',
    'admin',
    'active'
)
ON CONFLICT DO NOTHING;
