-- ============================================================================
-- Seed: Essential data for fresh database setup
-- Run after ALL migrations (001-108) to bootstrap the system with default data
-- ============================================================================

-- 1. Default organization
INSERT INTO organizations (id, name, theme_color, theme_logo_url)
VALUES (
    '550e8400-e29b-41d4-a716-446655440000',
    'Default Fleet',
    '#137fec',
    ''
)
ON CONFLICT (id) DO NOTHING;

-- 2. Default RBAC roles (tied to the default org)
INSERT INTO roles (id, org_id, name, description, is_system, permission_mask)
VALUES
    ('e349ce46-2c59-4c43-8338-afc37d01b88e', '550e8400-e29b-41d4-a716-446655440000', 'Super Admin', 'Full system access', true, 9223372036854775807),
    ('819ac656-6648-41bf-9dec-c38906264adc', '550e8400-e29b-41d4-a716-446655440000', 'Fleet Manager', 'Manage vehicles, drivers, trips', true, 281474976710655),
    ('215779f2-b515-4bca-9c73-de2ac1e3f047', '550e8400-e29b-41d4-a716-446655440000', 'Dispatcher', 'Manage trips and routing', true, 65535),
    ('857ad16f-1021-415b-b141-9bdc5b08b942', '550e8400-e29b-41d4-a716-446655440000', 'Driver', 'View assigned trips and vehicles', true, 1023),
    ('d3a2b1c0-0000-0000-0000-000000000005', '550e8400-e29b-41d4-a716-446655440000', 'Viewer', 'Read-only access', true, 1)
ON CONFLICT (id) DO NOTHING;

-- 3. Default super-admin user (password: admin123, bcrypt cost 10)
-- Must come after org + roles since org_id is NOT NULL
INSERT INTO users (id, org_id, email, password_hash, name, role, status, is_active,
                   email_verified, permission_mask, mfa_enabled, mfa_secret,
                   mfa_backup_codes, failed_logins, phone, avatar_url, last_login_ip,
                   sidebar_collapsed, language)
VALUES (
    '550e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440000',
    'admin@flyfleet.io',
    '$2a$10$B2/XZniBnqEmecEhQGzUn.9ZU5sNpthotIAmxEJcWqGEkodD7Jp9i',
    'Admin User',
    'admin',
    'active',
    true,
    true,
    9223372036854775807,
    false, '',
    '[]'::jsonb,
    0, '', '', '',
    false, 'en'
)
ON CONFLICT (id) DO NOTHING;

-- 4. Link admin to organization
INSERT INTO user_organizations (user_id, org_id, role, is_active)
VALUES (
    '550e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440000',
    'admin',
    true
)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Done! Default bootstrap data is ready.
-- Admin login: admin@flyfleet.io / admin123
-- Change password immediately after first login!
-- ============================================================================
