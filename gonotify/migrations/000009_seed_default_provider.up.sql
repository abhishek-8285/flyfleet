INSERT INTO providers (type, channel, name, capabilities)
VALUES
    ('smtp', 'email', 'Standard SMTP Email Provider', '{"send_email": true}'::jsonb),
    ('ses', 'email', 'AWS Simple Email Service', '{"send_email": true, "batch": true}'::jsonb),
    ('twilio', 'sms', 'Twilio SMS Gateway', '{"send_sms": true}'::jsonb)
ON CONFLICT (type, channel) DO NOTHING;

UPDATE api_keys
SET scopes = ARRAY['events:write', 'notifications:read', 'templates:read', 'templates:write', 'templates:admin', 'providers:read', 'providers:write', 'providers:admin']
WHERE key_prefix = 'gn_dev_local_01';
