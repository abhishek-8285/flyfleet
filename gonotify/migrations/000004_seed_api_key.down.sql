DELETE FROM api_keys
WHERE key_hash = encode(
    digest(
        'gn_dev_local_0123456789abcdef0123456789abcdef',
        'sha256'
    ),
    'hex'
);
