-- Consent logs track explicit opt-in/opt-out actions for marketing and digest
-- categories. A preference toggle alone is not sufficient for compliance;
-- this table provides auditable proof of when and how the user consented.
CREATE TABLE IF NOT EXISTS consent_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    category TEXT NOT NULL,
    channel TEXT NOT NULL,
    action TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'api',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_consent_logs_user
    ON consent_logs (tenant_id, user_id);

CREATE INDEX IF NOT EXISTS idx_consent_logs_user_category_channel
    ON consent_logs (tenant_id, user_id, category, channel);

CREATE INDEX IF NOT EXISTS idx_consent_logs_created_at
    ON consent_logs (created_at DESC);
