-- Register fakesms as a system provider type (dev-only, used for E2E testing).
INSERT INTO providers (type, channel, name, capabilities)
VALUES ('fakesms', 'sms', 'Fake SMS Provider (Dev)', '{"send_sms": true}'::jsonb)
ON CONFLICT (type, channel) DO NOTHING;

-- Seed a dev SMTP provider config pointing at MailHog (localhost:1025).
-- This allows DB-first provider routing to work out of the box in development.
INSERT INTO provider_configs (tenant_id, provider_id, name, settings, enabled, priority, weight)
SELECT
    t.id,
    p.id,
    'Dev SMTP (MailHog)',
    '{"host":"localhost","port":1025,"from_email":"notifications@gonotify.local","from_name":"GoNotify","use_tls":false}'::jsonb,
    true,
    100,
    100
FROM tenants t
CROSS JOIN providers p
WHERE t.slug = 'default'
  AND p.type = 'smtp'
  AND p.channel = 'email'
ON CONFLICT DO NOTHING;

-- Seed a dev FakeSMS provider config.
INSERT INTO provider_configs (tenant_id, provider_id, name, settings, enabled, priority, weight)
SELECT
    t.id,
    p.id,
    'Dev Fake SMS',
    '{}'::jsonb,
    true,
    100,
    100
FROM tenants t
CROSS JOIN providers p
WHERE t.slug = 'default'
  AND p.type = 'fakesms'
  AND p.channel = 'sms'
ON CONFLICT DO NOTHING;
