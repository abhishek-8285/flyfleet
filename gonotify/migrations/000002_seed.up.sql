INSERT INTO tenants (
    id,
    name,
    slug,
    status,
    settings
)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Default Tenant',
    'default',
    'active',
    '{}'::jsonb
)
ON CONFLICT (id) DO NOTHING;
