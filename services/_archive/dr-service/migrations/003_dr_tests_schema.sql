-- DR Testing Module Schema
-- Tables for managing DR test scenarios

-- DR Tests: Main test scenarios table
CREATE TABLE IF NOT EXISTS dr_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    test_type VARCHAR(20) NOT NULL CHECK (test_type IN ('FAILOVER', 'BACKUP', 'RESTORE')),
    target_env VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'SCHEDULED' CHECK (status IN ('SCHEDULED', 'RUNNING', 'PASSED', 'FAILED', 'ABORTED')),
    progress_percent INT DEFAULT 0,
    initiated_by VARCHAR(64),
    success_criteria TEXT,
    scheduled_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- DR Test Steps: Track individual steps in a test
CREATE TABLE IF NOT EXISTS dr_test_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id UUID NOT NULL REFERENCES dr_tests(id) ON DELETE CASCADE,
    step_order INT NOT NULL,
    step_name VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'RUNNING', 'PASSED', 'FAILED', 'SKIPPED')),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- DR Test Logs: Logs for each test execution
CREATE TABLE IF NOT EXISTS dr_test_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id UUID NOT NULL REFERENCES dr_tests(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    level VARCHAR(10) NOT NULL DEFAULT 'INFO',
    module VARCHAR(64),
    message TEXT NOT NULL
);

-- DR Test Settings: Per-organization settings
CREATE TABLE IF NOT EXISTS dr_test_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID,
    schedule_enabled BOOLEAN DEFAULT false,
    frequency VARCHAR(20) DEFAULT 'WEEKLY',
    schedule_day VARCHAR(10) DEFAULT 'SATURDAY',
    schedule_time TIME DEFAULT '02:00:00',
    intensity VARCHAR(20) DEFAULT 'NORMAL',
    notification_emails JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_dr_tests_status ON dr_tests(status);
CREATE INDEX IF NOT EXISTS idx_dr_tests_scheduled_at ON dr_tests(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_dr_test_steps_test_id ON dr_test_steps(test_id);
CREATE INDEX IF NOT EXISTS idx_dr_test_logs_test_id ON dr_test_logs(test_id);
CREATE INDEX IF NOT EXISTS idx_dr_test_logs_timestamp ON dr_test_logs(timestamp);
