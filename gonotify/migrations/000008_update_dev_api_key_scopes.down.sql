UPDATE api_keys
SET scopes = ARRAY['events:write', 'notifications:read']
WHERE key_prefix = 'gn_dev_local_01';
