-- Migration: 052_webhook_integrations
-- Description: Add webhook support for external system integrations

-- Create webhook subscriptions table
CREATE TABLE IF NOT EXISTS webhook_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    
    -- Webhook configuration
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    secret VARCHAR(255) NOT NULL, -- HMAC secret for signature verification
    is_active BOOLEAN DEFAULT true,
    
    -- Event filters
    events TEXT[] NOT NULL, -- Array of event types to subscribe to
    
    -- Headers (optional custom headers)
    custom_headers JSONB DEFAULT '{}',
    
    -- Retry configuration
    max_retries INTEGER DEFAULT 3,
    retry_delay_seconds INTEGER DEFAULT 60,
    timeout_seconds INTEGER DEFAULT 30,
    
    -- Rate limiting
    rate_limit_per_minute INTEGER DEFAULT 60,
    
    -- Statistics
    total_deliveries INTEGER DEFAULT 0,
    successful_deliveries INTEGER DEFAULT 0,
    failed_deliveries INTEGER DEFAULT 0,
    last_delivery_at TIMESTAMPTZ,
    last_success_at TIMESTAMPTZ,
    last_failure_at TIMESTAMPTZ,
    
    -- Metadata
    description TEXT,
    metadata JSONB DEFAULT '{}',
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT chk_webhook_max_retries CHECK (max_retries >= 0 AND max_retries <= 10),
    CONSTRAINT chk_webhook_retry_delay CHECK (retry_delay_seconds >= 10 AND retry_delay_seconds <= 3600),
    CONSTRAINT chk_webhook_timeout CHECK (timeout_seconds >= 5 AND timeout_seconds <= 300),
    CONSTRAINT chk_webhook_rate_limit CHECK (rate_limit_per_minute >= 1 AND rate_limit_per_minute <= 1000)
);

-- Create webhook delivery log table
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES webhook_subscriptions(id) ON DELETE CASCADE,
    
    -- Event details
    event_type VARCHAR(100) NOT NULL,
    event_id UUID NOT NULL,
    payload JSONB NOT NULL,
    
    -- Delivery attempt
    attempt INTEGER DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, success, failed, retrying
    http_status INTEGER,
    response_body TEXT,
    error_message TEXT,
    
    -- Timing
    sent_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ,
    
    -- Request details
    request_headers JSONB,
    response_headers JSONB,
    duration_ms INTEGER,
    
    -- Signature
    signature VARCHAR(255),
    
    CONSTRAINT chk_webhook_attempt CHECK (attempt >= 1 AND attempt <= 10)
);

-- Create webhook event types reference table
CREATE TABLE IF NOT EXISTS webhook_event_types (
    event_type VARCHAR(100) PRIMARY KEY,
    description TEXT,
    category VARCHAR(50), -- invoice, payment, credit_note, etc.
    is_active BOOLEAN DEFAULT true
);

-- Insert standard event types
INSERT INTO webhook_event_types (event_type, description, category) VALUES
    ('invoice.created', 'Triggered when a new invoice is created', 'invoice'),
    ('invoice.updated', 'Triggered when an invoice is updated', 'invoice'),
    ('invoice.issued', 'Triggered when an invoice is issued', 'invoice'),
    ('invoice.cancelled', 'Triggered when an invoice is cancelled', 'invoice'),
    ('invoice.paid', 'Triggered when an invoice is fully paid', 'invoice'),
    ('invoice.overdue', 'Triggered when an invoice becomes overdue', 'invoice'),
    ('payment.recorded', 'Triggered when a payment is recorded', 'payment'),
    ('payment.refunded', 'Triggered when a payment is refunded', 'payment'),
    ('credit_note.created', 'Triggered when a credit note is created', 'credit_note'),
    ('credit_note.issued', 'Triggered when a credit note is issued', 'credit_note'),
    ('credit_note.applied', 'Triggered when a credit note is applied', 'credit_note'),
    ('approval.requested', 'Triggered when an invoice approval is requested', 'approval'),
    ('approval.approved', 'Triggered when an invoice is approved', 'approval'),
    ('approval.rejected', 'Triggered when an invoice is rejected', 'approval')
ON CONFLICT (event_type) DO NOTHING;

-- Indexes
CREATE INDEX idx_webhook_subscriptions_org ON webhook_subscriptions(org_id);
CREATE INDEX idx_webhook_subscriptions_active ON webhook_subscriptions(is_active);
CREATE INDEX idx_webhook_subscriptions_events ON webhook_subscriptions USING GIN(events);
CREATE INDEX idx_webhook_deliveries_subscription ON webhook_deliveries(subscription_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_event ON webhook_deliveries(event_type);
CREATE INDEX idx_webhook_deliveries_sent ON webhook_deliveries(sent_at);
CREATE INDEX idx_webhook_deliveries_retry ON webhook_deliveries(next_retry_at) WHERE status = 'retrying';

-- Comments
COMMENT ON TABLE webhook_subscriptions IS 'Webhook subscription configuration for external integrations';
COMMENT ON TABLE webhook_deliveries IS 'Webhook delivery attempt log';
COMMENT ON TABLE webhook_event_types IS 'Available webhook event types';
COMMENT ON COLUMN webhook_subscriptions.secret IS 'HMAC secret for request signature verification';
COMMENT ON COLUMN webhook_subscriptions.events IS 'Array of event types to subscribe to';
COMMENT ON COLUMN webhook_deliveries.signature IS 'HMAC signature of payload for verification';

-- Function to queue webhook delivery
CREATE OR REPLACE FUNCTION queue_webhook_delivery(
    p_subscription_id UUID,
    p_event_type VARCHAR,
    p_event_id UUID,
    p_payload JSONB
) RETURNS UUID AS $$
DECLARE
    v_delivery_id UUID;
    v_signature TEXT;
    v_secret TEXT;
BEGIN
    -- Get subscription secret
    SELECT secret INTO v_secret FROM webhook_subscriptions WHERE id = p_subscription_id;
    
    -- Generate HMAC signature
    v_signature := encode(hmac(p_payload::text, v_secret, 'sha256'), 'hex');
    
    -- Create delivery record
    INSERT INTO webhook_deliveries (
        subscription_id, event_type, event_id, payload, signature, status
    ) VALUES (
        p_subscription_id, p_event_type, p_event_id, p_payload, v_signature, 'pending'
    ) RETURNING id INTO v_delivery_id;
    
    -- Update subscription statistics
    UPDATE webhook_subscriptions 
    SET total_deliveries = total_deliveries + 1,
        last_delivery_at = CURRENT_TIMESTAMP
    WHERE id = p_subscription_id;
    
    RETURN v_delivery_id;
END;
$$ LANGUAGE plpgsql;

-- Function to mark webhook delivery as successful
CREATE OR REPLACE FUNCTION mark_webhook_success(
    p_delivery_id UUID,
    p_http_status INTEGER,
    p_response_body TEXT,
    p_duration_ms INTEGER
) RETURNS void AS $$
BEGIN
    UPDATE webhook_deliveries SET
        status = 'success',
        http_status = p_http_status,
        response_body = p_response_body,
        delivered_at = CURRENT_TIMESTAMP,
        duration_ms = p_duration_ms
    WHERE id = p_delivery_id;
    
    -- Update subscription statistics
    UPDATE webhook_subscriptions SET
        successful_deliveries = successful_deliveries + 1,
        last_success_at = CURRENT_TIMESTAMP
    WHERE id = (SELECT subscription_id FROM webhook_deliveries WHERE id = p_delivery_id);
END;
$$ LANGUAGE plpgsql;

-- Function to mark webhook delivery as failed
CREATE OR REPLACE FUNCTION mark_webhook_failed(
    p_delivery_id UUID,
    p_http_status INTEGER,
    p_error_message TEXT,
    p_response_body TEXT
) RETURNS void AS $$
DECLARE
    v_subscription_id UUID;
    v_attempt INTEGER;
    v_max_retries INTEGER;
    v_retry_delay INTEGER;
BEGIN
    -- Get subscription details
    SELECT ws.id, wd.attempt, ws.max_retries, ws.retry_delay_seconds
    INTO v_subscription_id, v_attempt, v_max_retries, v_retry_delay
    FROM webhook_deliveries wd
    JOIN webhook_subscriptions ws ON wd.subscription_id = ws.id
    WHERE wd.id = p_delivery_id;
    
    -- Update delivery record
    UPDATE webhook_deliveries SET
        status = CASE 
            WHEN v_attempt >= v_max_retries THEN 'failed'
            ELSE 'retrying'
        END,
        http_status = p_http_status,
        error_message = p_error_message,
        response_body = p_response_body,
        next_retry_at = CASE 
            WHEN v_attempt >= v_max_retries THEN NULL
            ELSE CURRENT_TIMESTAMP + (v_retry_delay * INTERVAL '1 second' * POWER(2, v_attempt))
        END
    WHERE id = p_delivery_id;
    
    -- Update subscription statistics
    UPDATE webhook_subscriptions SET
        failed_deliveries = failed_deliveries + 1,
        last_failure_at = CURRENT_TIMESTAMP
    WHERE id = v_subscription_id;
END;
$$ LANGUAGE plpgsql;

-- View for webhook delivery statistics
CREATE OR REPLACE VIEW webhook_subscription_stats AS
SELECT 
    ws.id,
    ws.org_id,
    ws.name,
    ws.url,
    ws.is_active,
    ws.total_deliveries,
    ws.successful_deliveries,
    ws.failed_deliveries,
    CASE 
        WHEN ws.total_deliveries > 0 THEN 
            ROUND((ws.successful_deliveries::DECIMAL / ws.total_deliveries) * 100, 2)
        ELSE 0 
    END as success_rate,
    ws.last_delivery_at,
    ws.last_success_at,
    ws.last_failure_at
FROM webhook_subscriptions ws
WHERE ws.deleted_at IS NULL;
