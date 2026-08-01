UPDATE api_keys
SET scopes = ARRAY['events:write', 'notifications:read', 'templates:read', 'templates:write', 'templates:admin']
WHERE key_prefix = 'gn_dev_local_01';
