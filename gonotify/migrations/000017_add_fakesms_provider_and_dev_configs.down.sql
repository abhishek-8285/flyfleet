-- Remove seeded dev provider configs.
DELETE FROM provider_configs
WHERE name IN ('Dev SMTP (MailHog)', 'Dev Fake SMS')
  AND tenant_id = (SELECT id FROM tenants WHERE slug = 'default');

-- Remove the fakesms system provider type.
DELETE FROM providers WHERE type = 'fakesms' AND channel = 'sms';
