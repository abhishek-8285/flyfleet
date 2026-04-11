-- ============================================================================
-- Seed: Essential data for fresh database setup
-- Run after migrations to bootstrap the system with default data
-- ============================================================================

-- Default super-admin user (password: change_me_123)
-- This is the bootstrap admin account - change password on first login
INSERT INTO users (id, email, password_hash, name, status, email_verified, permission_mask)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@flyfleet.local',
    '$2a$10$X7yMjZkzPzGqKX4fGqKX4eYjHZqKX4fGqKX4eYjHZqKX4eYjHZqKX',
    'System Administrator',
    'active',
    true,
    9223372036854775807  -- All permissions (bitmask all 1s)
)
ON CONFLICT (id) DO NOTHING;

-- Default organization
INSERT INTO organizations (id, name, slug, status, plan_type)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Default Fleet',
    'default-fleet',
    'active',
    'enterprise'
)
ON CONFLICT (id) DO NOTHING;

-- Link admin to default organization
INSERT INTO user_organizations (user_id, organization_id, role)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'admin'
)
ON CONFLICT (user_id, organization_id) DO NOTHING;

-- Default RBAC roles
INSERT INTO roles (id, name, description, permissions, is_system)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'Super Admin', 'Full system access', 9223372036854775807, true),
    ('00000000-0000-0000-0000-000000000002', 'Fleet Manager', 'Manage vehicles, drivers, trips', 281474976710655, true),
    ('00000000-0000-0000-0000-000000000003', 'Driver', 'View assigned trips and vehicles', 1023, true),
    ('00000000-0000-0000-0000-000000000004', 'Viewer', 'Read-only access', 1, true),
    ('00000000-0000-0000-0000-000000000005', 'Dispatcher', 'Manage trips and routing', 65535, true)
ON CONFLICT (id) DO NOTHING;

-- System configuration defaults
INSERT INTO system_config (key, value, description)
VALUES
    ('auth.mfa_required', 'false', 'Require MFA for all users'),
    ('auth.session_timeout', '3600', 'Session timeout in seconds'),
    ('auth.max_concurrent_sessions', '3', 'Max concurrent sessions per user'),
    ('tracking.gps_interval', '30', 'GPS telemetry interval in seconds'),
    ('tracking.websocket_port', '8084', 'WebSocket server port'),
    ('features.geofencing', 'true', 'Enable geofencing module'),
    ('features.maintenance', 'true', 'Enable maintenance module'),
    ('features.billing', 'true', 'Enable billing module'),
    ('features.india_integration', 'true', 'Enable India-specific integrations'),
    ('ui.default_language', 'en', 'Default UI language'),
    ('ui.default_map_center', '{"lat":28.6139,"lng":77.2090}', 'Default map center (New Delhi)')
ON CONFLICT (key) DO NOTHING;

-- Audit log table bootstrap (ensure it exists and has retention policy)
DO $$
BEGIN
    -- Ensure audit log retention policy is set
    INSERT INTO audit_policies (event_pattern, retention_days, created_at)
    VALUES
        ('%', 90, NOW()),
        ('auth.%', 365, NOW()),
        ('user.%', 365, NOW())
    ON CONFLICT DO NOTHING;
END $$;

-- Sample vehicle types
INSERT INTO vehicle_types (name, description, capacity_unit)
VALUES
    ('truck_heavy', 'Heavy duty truck', 'tons'),
    ('truck_medium', 'Medium duty truck', 'tons'),
    ('truck_light', 'Light commercial vehicle', 'kg'),
    ('van', 'Delivery van', 'cubic_meters'),
    ('car', 'Passenger car', 'passengers'),
    ('bike', 'Motorcycle/scooter', 'passengers'),
    ('bus', 'Passenger bus', 'passengers'),
    ('trailer', 'Semi-trailer', 'tons')
ON CONFLICT (name) DO NOTHING;

-- Sample fuel types
INSERT INTO fuel_types (name, unit)
VALUES
    ('diesel', 'liters'),
    ('petrol', 'liters'),
    ('cng', 'kg'),
    ('electric', 'kWh'),
    ('lpg', 'liters')
ON CONFLICT (name) DO NOTHING;

-- Notification templates
INSERT INTO notification_templates (name, channel, subject, body, is_system)
VALUES
    ('geofence_entry', 'push', 'Vehicle entered geofence', 'Vehicle {{vehicle_name}} entered zone {{zone_name}}', true),
    ('geofence_exit', 'push', 'Vehicle exited geofence', 'Vehicle {{vehicle_name}} exited zone {{zone_name}}', true),
    ('maintenance_due', 'email', 'Maintenance Due', 'Vehicle {{vehicle_name}} is due for {{maintenance_type}}', true),
    ('sos_alert', 'push', 'SOS ALERT', 'Emergency SOS from vehicle {{vehicle_name}} at {{location}}', true),
    ('trip_completed', 'email', 'Trip Completed', 'Trip {{trip_id}} has been completed successfully', true),
    ('mfa_code', 'email', 'Your MFA Code', 'Your verification code is: {{code}}', true)
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- Done! Default bootstrap data is ready.
-- Admin login: admin@flyfleet.local / change_me_123
-- Change password immediately after first login!
-- ============================================================================
