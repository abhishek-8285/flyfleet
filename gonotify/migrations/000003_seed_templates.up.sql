WITH tpl AS (
    INSERT INTO templates (
        tenant_id,
        key,
        name,
        channel,
        category
    )
    SELECT
        id,
        'order.created',
        'Order Created Email',
        'email',
        'transactional'
    FROM tenants
    WHERE slug = 'default'
    ON CONFLICT (tenant_id, key, channel)
    DO UPDATE SET
        updated_at = now()
    RETURNING id
)
INSERT INTO template_versions (
    template_id,
    version,
    subject,
    html_body,
    text_body,
    status
)
SELECT
    id,
    1,
    $$Your order {{.order_id}} has been created$$,
    $$
<p>Hello,</p>

<p>
  Your order <strong>{{.order_id}}</strong> has been created.
</p>

<p>
  Amount: {{.amount}} {{.currency}}
</p>

<p>
  This email was rendered from a database template.
</p>
$$,
    $$
Hello,

Your order {{.order_id}} has been created.

Amount: {{.amount}} {{.currency}}

This email was rendered from a database template.
$$,
    'published'
FROM tpl
ON CONFLICT (template_id, version) DO NOTHING;
