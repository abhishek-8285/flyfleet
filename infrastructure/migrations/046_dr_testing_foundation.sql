-- Migration: 046_dr_testing_foundation.sql
-- Description: DR test execution and settings foundation

CREATE TABLE IF NOT EXISTS dr_tests (
    id UUID PRIMARY KEY,
    test_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    test_type TEXT NOT NULL,
    target_env TEXT NOT NULL,
    status TEXT NOT NULL,
    progress_percent INTEGER NOT NULL DEFAULT 0,
    initiated_by TEXT NOT NULL DEFAULT '',
    success_criteria TEXT DEFAULT '',
    scheduled_at TIMESTAMPTZ NULL,
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dr_tests_status ON dr_tests(status);
CREATE INDEX IF NOT EXISTS idx_dr_tests_created_at ON dr_tests(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_dr_tests_scheduled_at ON dr_tests(scheduled_at);

CREATE TABLE IF NOT EXISTS dr_test_steps (
    id UUID PRIMARY KEY,
    test_id UUID NOT NULL REFERENCES dr_tests(id) ON DELETE CASCADE,
    step_order INTEGER NOT NULL,
    step_name TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ NULL,
    completed_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_dr_test_steps_test_id ON dr_test_steps(test_id, step_order);

CREATE TABLE IF NOT EXISTS dr_test_logs (
    id UUID PRIMARY KEY,
    test_id UUID NOT NULL REFERENCES dr_tests(id) ON DELETE CASCADE,
    ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    level TEXT NOT NULL,
    message TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dr_test_logs_test_id_ts ON dr_test_logs(test_id, ts DESC);

CREATE TABLE IF NOT EXISTS dr_test_settings (
    id BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id = TRUE),
    schedule_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    frequency TEXT NOT NULL DEFAULT 'WEEKLY',
    schedule_day TEXT NOT NULL DEFAULT 'SATURDAY',
    schedule_time TEXT NOT NULL DEFAULT '02:00',
    intensity TEXT NOT NULL DEFAULT 'NORMAL',
    notification_emails JSONB NOT NULL DEFAULT '[]'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO dr_test_settings (id, schedule_enabled, frequency, schedule_day, schedule_time, intensity, notification_emails)
VALUES (TRUE, FALSE, 'WEEKLY', 'SATURDAY', '02:00', 'NORMAL', '[]'::jsonb)
ON CONFLICT (id) DO NOTHING;
