-- Seed for Runtime Gate
-- Organization
INSERT INTO organizations (id, name, domain, status)
VALUES ('00000000-0000-0000-0000-000000000001', 'FlyFleet Global', 'flyfleet.com', 'active')
ON CONFLICT (id) DO NOTHING;

-- User: admin@flyfleet.com / admin123
-- Just update the fields we need, don't try to change the ID if it already exists
INSERT INTO users (id, email, password_hash, name, role, status, org_id)
VALUES ('00000000-0000-0000-0000-000000000001', 'admin@flyfleet.com', '$2b$12$9k5kXkgHFTeLVYowLaX7Je4HdOQ/FnW0tQNEnz6bls.xqV0q4R4FK', 'Platform Admin', 'admin', 'active', '00000000-0000-0000-0000-000000000001')
ON CONFLICT (email) DO UPDATE SET 
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    org_id = EXCLUDED.org_id;

-- Mapping
-- Get the ID of the user we just ensured exists (it might not be the one we provided if ON CONFLICT handled it)
INSERT INTO user_organizations (user_id, org_id, role)
SELECT id, '00000000-0000-0000-0000-000000000001', 'admin'
FROM users WHERE email = 'admin@flyfleet.com'
ON CONFLICT (user_id, org_id) DO NOTHING;
