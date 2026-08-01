-- Local development API key.
--
-- Raw key:
-- gn_dev_local_0123456789abcdef0123456789abcdef
--
-- This key is for local development only.
-- Never use this key in production.

INSERT INTO api_keys (
    tenant_id,
    name,
    key_prefix,
    key_hash,
    scopes,
    environment
)
SELECT
    id,
    'Local Development Key',
    'gn_dev_local_01',
    encode(
        digest(
            'gn_dev_local_0123456789abcdef0123456789abcdef',
            'sha256'
        ),
        'hex'
    ),
    ARRAY[
        'events:write',
        'notifications:read'
    ],
    'development'
FROM tenants
WHERE slug = 'default'
ON CONFLICT (key_hash) DO NOTHING;
