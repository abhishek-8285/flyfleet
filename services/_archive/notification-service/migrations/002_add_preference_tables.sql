-- User Preference Table
CREATE TABLE user_notification_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    email_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sms_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    push_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    in_app_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_modified_by UUID NOT NULL,
    last_modified_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_user_notification_preferences_user ON user_notification_preferences(user_id);

-- Audit Table
CREATE TABLE notification_preference_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    changed_by UUID NOT NULL,
    old_values JSONB,
    new_values JSONB,
    change_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notification_preference_audit_user ON notification_preference_audit(user_id);

-- Global Defaults Table
CREATE TABLE global_notification_defaults (
    id SERIAL PRIMARY KEY,
    email_enabled BOOLEAN DEFAULT TRUE,
    sms_enabled BOOLEAN DEFAULT FALSE,
    push_enabled BOOLEAN DEFAULT TRUE,
    in_app_enabled BOOLEAN DEFAULT TRUE,
    updated_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed defaults
INSERT INTO global_notification_defaults (email_enabled, sms_enabled, push_enabled, in_app_enabled)
VALUES (TRUE, FALSE, TRUE, TRUE);
