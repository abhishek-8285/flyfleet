DELETE FROM template_versions
WHERE template_id IN (
    SELECT id
    FROM templates
    WHERE tenant_id = (
        SELECT id
        FROM tenants
        WHERE slug = 'default'
    )
      AND key = 'order.created'
      AND channel = 'email'
);

DELETE FROM templates
WHERE tenant_id = (
    SELECT id
    FROM tenants
    WHERE slug = 'default'
)
  AND key = 'order.created'
  AND channel = 'email';
