CREATE TABLE settings_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    action VARCHAR(50) NOT NULL,
    actor_id UUID NOT NULL,
    actor_ip VARCHAR(50),
    changes JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_entity_type ON settings_audit_logs(entity_type);
CREATE INDEX idx_audit_actor_id ON settings_audit_logs(actor_id);
CREATE INDEX idx_audit_created_at ON settings_audit_logs(created_at DESC);
