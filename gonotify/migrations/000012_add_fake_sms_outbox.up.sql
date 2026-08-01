CREATE TABLE IF NOT EXISTS fake_sms_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    notification_id UUID,
    to_phone TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_fake_sms_outbox_tenant
ON fake_sms_outbox (tenant_id);

CREATE INDEX IF NOT EXISTS idx_fake_sms_outbox_phone
ON fake_sms_outbox (to_phone);

CREATE INDEX IF NOT EXISTS idx_fake_sms_outbox_created_at
ON fake_sms_outbox (created_at DESC);
