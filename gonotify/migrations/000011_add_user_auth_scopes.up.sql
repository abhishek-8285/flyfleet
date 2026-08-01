UPDATE api_keys
SET scopes = ARRAY[
    'events:write',
    'notifications:read',
    'templates:read',
    'templates:write',
    'providers:read',
    'providers:write',
    'users:write',
    'inbox:read',
    'inbox:write'
]
WHERE key_hash = encode(
    digest(
        'gn_dev_local_0123456789abcdef0123456789abcdef',
        'sha256'
    ),
    'hex'
);
