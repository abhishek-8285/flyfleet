-- ============================================================================
-- FlyFleet - Hyper-Defensive Dev/Test Seed Data
-- ============================================================================
-- Purpose: Seed a complete test organization with user, vehicles, drivers,
--          telemetry, geofences, and events for development & testing.
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. ORGANIZATION
-- ============================================================================
INSERT INTO tenants (id, name)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'Delta Logistics Pvt Ltd')
ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- 2. USERS (Defensive)
-- ============================================================================
DO $$
DECLARE
    u_id UUID;
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
    p_hash TEXT := '$2a$04$WOojoGDL.4DzIGNd3opDV.phzq0f8tFT8pEQXk2StJ7YvyGPDBahq';
BEGIN
    -- Admin
    u_id := '550e8400-e29b-41d4-a716-446655440001';
    IF NOT EXISTS (SELECT 1 FROM users WHERE id = u_id) THEN
        INSERT INTO users (id, email, password_hash, name)
        VALUES (u_id, 'admin@delta.io', p_hash, 'Admin User');
        
        -- Backfill identity column
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'tenant_id') THEN
            UPDATE users SET tenant_id = t_id WHERE id = u_id;
        ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'organization_id') THEN
            UPDATE users SET organization_id = t_id WHERE id = u_id;
        END IF;

        -- Handle Role mapping
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role') THEN
            UPDATE users SET role = 'admin' WHERE id = u_id;
        END IF;
    END IF;

    -- Map to organization_users if it exists
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'organization_users') THEN
        INSERT INTO organization_users (tenant_id, email, role, status)
        VALUES (t_id, 'admin@delta.io', 'ADMIN', 'ACTIVE')
        ON CONFLICT DO NOTHING;
    END IF;
END $$;

-- ============================================================================
-- 3. DRIVERS (Hyper-Defensive)
-- ============================================================================
DO $$
DECLARE
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
    v_sql TEXT;
    v_cols TEXT;
    v_has_tenant BOOLEAN;
    v_has_org BOOLEAN;
    v_has_full_name BOOLEAN;
    v_has_first_name BOOLEAN;
    v_has_exp BOOLEAN;
    v_has_rating BOOLEAN;
    v_has_trips BOOLEAN;
    v_has_dist BOOLEAN;
    v_has_hours BOOLEAN;
    v_has_score BOOLEAN;
BEGIN
    -- Detect schema features
    v_has_tenant := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'tenant_id');
    v_has_org := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'org_id');
    v_has_full_name := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'full_name');
    v_has_first_name := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'first_name');
    v_has_exp := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'experience_years');
    v_has_rating := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'rating');
    v_has_trips := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'total_trips');
    v_has_dist := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'total_distance_km');
    v_has_hours := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'total_hours_driven');
    v_has_score := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'behavior_score');

    -- Build column list
    v_cols := 'id, phone, email, license_number, license_expiry, license_authority, date_of_birth, address, emergency_contact, status, joining_date';
    IF v_has_tenant THEN v_cols := v_cols || ', tenant_id'; END IF;
    IF v_has_org THEN v_cols := v_cols || ', org_id'; END IF;
    IF v_has_exp THEN v_cols := v_cols || ', experience_years'; END IF;
    IF v_has_rating THEN v_cols := v_cols || ', rating'; END IF;
    IF v_has_trips THEN v_cols := v_cols || ', total_trips'; END IF;
    IF v_has_dist THEN v_cols := v_cols || ', total_distance_km'; END IF;
    IF v_has_hours THEN v_cols := v_cols || ', total_hours_driven'; END IF;
    IF v_has_score THEN v_cols := v_cols || ', behavior_score'; END IF;

    -- Cleanup: Fix any existing records with NULL mandatory fields
    IF v_has_exp THEN EXECUTE 'UPDATE drivers SET experience_years = 0 WHERE experience_years IS NULL'; END IF;
    IF v_has_rating THEN EXECUTE 'UPDATE drivers SET rating = 5.0 WHERE rating IS NULL'; END IF;
    IF v_has_trips THEN EXECUTE 'UPDATE drivers SET total_trips = 0 WHERE total_trips IS NULL'; END IF;
    IF v_has_dist THEN EXECUTE 'UPDATE drivers SET total_distance_km = 0 WHERE total_distance_km IS NULL'; END IF;
    IF v_has_hours THEN EXECUTE 'UPDATE drivers SET total_hours_driven = 0 WHERE total_hours_driven IS NULL'; END IF;
    IF v_has_score THEN EXECUTE 'UPDATE drivers SET behavior_score = 100 WHERE behavior_score IS NULL'; END IF;
    EXECUTE 'UPDATE drivers SET status = ''active'' WHERE status IS NULL OR status = ''''';

    -- Driver 1
    v_sql := 'INSERT INTO drivers (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440010'', ''+91 98765 43210'', ''rajesh@delta.io'', ''DL-1234567890'', ''2030-01-01'', ''Delhi RTO'', ''1985-05-15'', ''{}'', ''{}'', ''active'', ''2023-01-01''';
    IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_exp THEN v_sql := v_sql || ', 5'; END IF;
    IF v_has_rating THEN v_sql := v_sql || ', 4.8'; END IF;
    IF v_has_trips THEN v_sql := v_sql || ', 150'; END IF;
    IF v_has_dist THEN v_sql := v_sql || ', 12500.5'; END IF;
    IF v_has_hours THEN v_sql := v_sql || ', 450.2'; END IF;
    IF v_has_score THEN v_sql := v_sql || ', 98'; END IF;
    IF v_has_full_name AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'full_name' AND is_generated = 'NEVER') THEN
        v_sql := v_sql || ', ''Rajesh Kumar''';
    ELSIF v_has_first_name THEN
        v_sql := v_sql || ', ''Rajesh'', ''Kumar''';
    END IF;
    v_sql := v_sql || ') ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status';
    IF v_has_rating THEN v_sql := v_sql || ', rating = EXCLUDED.rating'; END IF;
    IF v_has_score THEN v_sql := v_sql || ', behavior_score = EXCLUDED.behavior_score'; END IF;
    EXECUTE v_sql;

    -- Driver 2
    v_sql := 'INSERT INTO drivers (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440011'', ''+91 98765 43211'', ''amit@delta.io'', ''DL-1234567891'', ''2030-01-01'', ''Delhi RTO'', ''1988-08-20'', ''{}'', ''{}'', ''active'', ''2023-01-01''';
    IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_exp THEN v_sql := v_sql || ', 3'; END IF;
    IF v_has_rating THEN v_sql := v_sql || ', 4.5'; END IF;
    IF v_has_trips THEN v_sql := v_sql || ', 85'; END IF;
    IF v_has_dist THEN v_sql := v_sql || ', 7200.0'; END IF;
    IF v_has_hours THEN v_sql := v_sql || ', 310.5'; END IF;
    IF v_has_score THEN v_sql := v_sql || ', 92'; END IF;
    IF v_has_full_name AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'full_name' AND is_generated = 'NEVER') THEN
        v_sql := v_sql || ', ''Amit Sharma''';
    ELSIF v_has_first_name THEN
        v_sql := v_sql || ', ''Amit'', ''Sharma''';
    END IF;
    v_sql := v_sql || ') ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status';
    IF v_has_rating THEN v_sql := v_sql || ', rating = EXCLUDED.rating'; END IF;
    IF v_has_score THEN v_sql := v_sql || ', behavior_score = EXCLUDED.behavior_score'; END IF;
    EXECUTE v_sql;

    -- Driver 3
    v_sql := 'INSERT INTO drivers (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440012'', ''+91 98765 43212'', ''suresh@delta.io'', ''DL-1234567892'', ''2030-01-01'', ''Delhi RTO'', ''1990-11-30'', ''{}'', ''{}'', ''active'', ''2023-01-01''';
    IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_exp THEN v_sql := v_sql || ', 7'; END IF;
    IF v_has_rating THEN v_sql := v_sql || ', 4.9'; END IF;
    IF v_has_trips THEN v_sql := v_sql || ', 320'; END IF;
    IF v_has_dist THEN v_sql := v_sql || ', 28000.0'; END IF;
    IF v_has_hours THEN v_sql := v_sql || ', 1150.0'; END IF;
    IF v_has_score THEN v_sql := v_sql || ', 96'; END IF;
    IF v_has_full_name AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'full_name' AND is_generated = 'NEVER') THEN
        v_sql := v_sql || ', ''Suresh Patel''';
    ELSIF v_has_first_name THEN
        v_sql := v_sql || ', ''Suresh'', ''Patel''';
    END IF;
    v_sql := v_sql || ') ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status';
    IF v_has_rating THEN v_sql := v_sql || ', rating = EXCLUDED.rating'; END IF;
    IF v_has_score THEN v_sql := v_sql || ', behavior_score = EXCLUDED.behavior_score'; END IF;
    EXECUTE v_sql;
END $$;

-- ============================================================================
-- 4. VEHICLES
-- ============================================================================
DO $$
DECLARE
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
    v_sql TEXT;
    v_cols TEXT;
    v_has_model BOOLEAN;
    v_has_make BOOLEAN;
    v_has_year BOOLEAN;
    v_has_reg_legacy BOOLEAN;
    v_has_reg BOOLEAN;
    v_has_org BOOLEAN;
    v_has_tenant BOOLEAN;
    v_has_type BOOLEAN;
    v_has_fuel BOOLEAN;
    v_has_ownership BOOLEAN;
BEGIN
    -- Detect columns
    v_has_model := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'model');
    v_has_make := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'make');
    v_has_year := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'year');
    v_has_reg_legacy := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'registration_number_legacy');
    v_has_reg := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'registration_number');
    v_has_org := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'org_id');
    v_has_tenant := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'tenant_id');
    v_has_type := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'vehicle_type');
    v_has_fuel := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'fuel_type');
    v_has_ownership := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'ownership_type');

    -- Build column list
    v_cols := 'id, status, current_driver_id';
    IF v_has_model THEN v_cols := v_cols || ', model'; END IF;
    IF v_has_make THEN v_cols := v_cols || ', make'; END IF;
    IF v_has_year THEN v_cols := v_cols || ', year'; END IF;
    IF v_has_reg_legacy THEN v_cols := v_cols || ', registration_number_legacy'; END IF;
    IF v_has_reg THEN v_cols := v_cols || ', registration_number'; END IF;
    IF v_has_org THEN v_cols := v_cols || ', org_id'; END IF;
    IF v_has_tenant THEN v_cols := v_cols || ', tenant_id'; END IF;
    IF v_has_type THEN v_cols := v_cols || ', vehicle_type'; END IF;
    IF v_has_fuel THEN v_cols := v_cols || ', fuel_type'; END IF;
    IF v_has_ownership THEN v_cols := v_cols || ', ownership_type'; END IF;

    -- Cleanup: Fix any existing records with NULL mandatory fields
    IF v_has_type THEN EXECUTE 'UPDATE vehicles SET vehicle_type = ''TRUCK'' WHERE vehicle_type IS NULL'; END IF;
    IF v_has_fuel THEN EXECUTE 'UPDATE vehicles SET fuel_type = ''DIESEL'' WHERE fuel_type IS NULL'; END IF;
    IF v_has_ownership THEN EXECUTE 'UPDATE vehicles SET ownership_type = ''OWNED'' WHERE ownership_type IS NULL'; END IF;
    EXECUTE 'UPDATE vehicles SET status = ''ACTIVE'' WHERE status IS NULL OR status = ''''';

    -- Vehicle 1
    BEGIN
        v_sql := 'INSERT INTO vehicles (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440001'', ''ACTIVE'', ''550e8400-e29b-41d4-a716-446655440010''';
        IF v_has_model THEN v_sql := v_sql || ', ''FleetMaster Pro'''; END IF;
        IF v_has_make THEN v_sql := v_sql || ', ''FlyFleet'''; END IF;
        IF v_has_year THEN v_sql := v_sql || ', 2024'; END IF;
        IF v_has_reg_legacy THEN v_sql := v_sql || ', ''DL01AB1234'''; END IF;
        IF v_has_reg THEN v_sql := v_sql || ', ''DL01AB1234'''; END IF;
        IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
        IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
        IF v_has_type THEN v_sql := v_sql || ', ''TRUCK'''; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', ''DIESEL'''; END IF;
        IF v_has_ownership THEN v_sql := v_sql || ', ''OWNED'''; END IF;
        v_sql := v_sql || ') ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status';
        IF v_has_type THEN v_sql := v_sql || ', vehicle_type = EXCLUDED.vehicle_type'; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', fuel_type = EXCLUDED.fuel_type'; END IF;
        IF v_has_ownership THEN v_sql := v_sql || ', ownership_type = EXCLUDED.ownership_type'; END IF;
        EXECUTE v_sql;
    END;

    -- Vehicle 2
    BEGIN
        v_sql := 'INSERT INTO vehicles (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440002'', ''ACTIVE'', ''550e8400-e29b-41d4-a716-446655440011''';
        IF v_has_model THEN v_sql := v_sql || ', ''TransPorter'''; END IF;
        IF v_has_make THEN v_sql := v_sql || ', ''Tata Motors'''; END IF;
        IF v_has_year THEN v_sql := v_sql || ', 2023'; END IF;
        IF v_has_reg_legacy THEN v_sql := v_sql || ', ''MH02CD5678'''; END IF;
        IF v_has_reg THEN v_sql := v_sql || ', ''MH02CD5678'''; END IF;
        IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
        IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
        IF v_has_type THEN v_sql := v_sql || ', ''VAN'''; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', ''PETROL'''; END IF;
        IF v_has_ownership THEN v_sql := v_sql || ', ''LEASED'''; END IF;
        v_sql := v_sql || ') ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status';
        IF v_has_type THEN v_sql := v_sql || ', vehicle_type = EXCLUDED.vehicle_type'; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', fuel_type = EXCLUDED.fuel_type'; END IF;
        IF v_has_ownership THEN v_sql := v_sql || ', ownership_type = EXCLUDED.ownership_type'; END IF;
        EXECUTE v_sql;
    END;

    -- Vehicle 3
    BEGIN
        v_sql := 'INSERT INTO vehicles (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440003'', ''ACTIVE'', ''550e8400-e29b-41d4-a716-446655440012''';
        IF v_has_model THEN v_sql := v_sql || ', ''E-Truck 500'''; END IF;
        IF v_has_make THEN v_sql := v_sql || ', ''Ashok Leyland'''; END IF;
        IF v_has_year THEN v_sql := v_sql || ', 2024'; END IF;
        IF v_has_reg_legacy THEN v_sql := v_sql || ', ''KA03EF9012'''; END IF;
        IF v_has_reg THEN v_sql := v_sql || ', ''KA03EF9012'''; END IF;
        IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
        IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
        IF v_has_type THEN v_sql := v_sql || ', ''BUS'''; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', ''ELECTRIC'''; END IF;
        IF v_has_ownership THEN v_sql := v_sql || ', ''OWNED'''; END IF;
        v_sql := v_sql || ') ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status';
        IF v_has_type THEN v_sql := v_sql || ', vehicle_type = EXCLUDED.vehicle_type'; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', fuel_type = EXCLUDED.fuel_type'; END IF;
        IF v_has_ownership THEN v_sql := v_sql || ', ownership_type = EXCLUDED.ownership_type'; END IF;
        EXECUTE v_sql;
    END;

    -- Vehicle 4
    BEGIN
        v_sql := 'INSERT INTO vehicles (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440004'', ''ACTIVE'', NULL';
        IF v_has_model THEN v_sql := v_sql || ', ''GigaTruck'''; END IF;
        IF v_has_make THEN v_sql := v_sql || ', ''Mahindra'''; END IF;
        IF v_has_year THEN v_sql := v_sql || ', 2022'; END IF;
        IF v_has_reg_legacy THEN v_sql := v_sql || ', ''TN04GH3456'''; END IF;
        IF v_has_reg THEN v_sql := v_sql || ', ''TN04GH3456'''; END IF;
        IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
        IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
        IF v_has_type THEN v_sql := v_sql || ', ''CAR'''; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', ''CNG'''; END IF;
        IF v_has_ownership THEN v_sql := v_sql || ', ''FINANCED'''; END IF;
        v_sql := v_sql || ') ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status';
        IF v_has_type THEN v_sql := v_sql || ', vehicle_type = EXCLUDED.vehicle_type'; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', fuel_type = EXCLUDED.fuel_type'; END IF;
        IF v_has_ownership THEN v_sql := v_sql || ', ownership_type = EXCLUDED.ownership_type'; END IF;
        EXECUTE v_sql;
    END;

    -- Vehicle 5
    BEGIN
        v_sql := 'INSERT INTO vehicles (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440005'', ''ACTIVE'', ''550e8400-e29b-41d4-a716-446655440014''';
        IF v_has_model THEN v_sql := v_sql || ', ''Hauler X'''; END IF;
        IF v_has_make THEN v_sql := v_sql || ', ''Eicher'''; END IF;
        IF v_has_year THEN v_sql := v_sql || ', 2024'; END IF;
        IF v_has_reg_legacy THEN v_sql := v_sql || ', ''GJ05JK7890'''; END IF;
        IF v_has_reg THEN v_sql := v_sql || ', ''GJ05JK7890'''; END IF;
        IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
        IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
        IF v_has_type THEN v_sql := v_sql || ', ''TANKER'''; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', ''DIESEL'''; END IF;
        IF v_has_ownership THEN v_sql := v_sql || ', ''VENDOR'''; END IF;
        v_sql := v_sql || ') ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status';
        IF v_has_type THEN v_sql := v_sql || ', vehicle_type = EXCLUDED.vehicle_type'; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', fuel_type = EXCLUDED.fuel_type'; END IF;
        IF v_has_ownership THEN v_sql := v_sql || ', ownership_type = EXCLUDED.ownership_type'; END IF;
        EXECUTE v_sql;
    END;
END $$;

-- ============================================================================
-- 5. TELEMETRY DATA
-- ============================================================================
DELETE FROM telemetry_points WHERE vehicle_id IN (
    '550e8400-e29b-41d4-a716-446655440001'::uuid,
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    '550e8400-e29b-41d4-a716-446655440003'::uuid,
    '550e8400-e29b-41d4-a716-446655440004'::uuid,
    '550e8400-e29b-41d4-a716-446655440005'::uuid
);

DO $$
DECLARE
    v_time_col TEXT;
    v_lat_col TEXT;
    v_lng_col TEXT;
    v_speed_col TEXT;
    v_sql TEXT;
BEGIN
    -- Determine column names
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'recorded_at') THEN
        v_time_col := 'recorded_at';
    ELSE
        v_time_col := 'time';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'lat') THEN
        v_lat_col := 'lat';
    ELSE
        v_lat_col := 'latitude';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'lng') THEN
        v_lng_col := 'lng';
    ELSE
        v_lng_col := 'longitude';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'speed') THEN
        v_speed_col := 'speed';
    ELSE
        v_speed_col := 'speed_kmh';
    END IF;

    DECLARE
        v_has_device_id BOOLEAN;
        v_has_event_id BOOLEAN;
        v_has_tenant_id BOOLEAN;
        v_has_heading BOOLEAN;
        v_has_ignition BOOLEAN;
        v_has_fuel BOOLEAN;
        v_has_trip BOOLEAN;
        v_tele_cols TEXT;
        v_tele_vals TEXT;
    BEGIN
        v_has_device_id := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'device_id');
        v_has_event_id := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'event_id');
        v_has_tenant_id := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'tenant_id');
        v_has_heading := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'heading');
        v_has_ignition := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'ignition');
        v_has_fuel := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'fuel_level');
        v_has_trip := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'telemetry_points' AND column_name = 'trip_id');

        v_tele_cols := format('vehicle_id, %I, %I, %I, %I', v_lat_col, v_lng_col, v_speed_col, v_time_col);
        IF v_has_device_id THEN v_tele_cols := v_tele_cols || ', device_id'; END IF;
        IF v_has_event_id THEN v_tele_cols := v_tele_cols || ', event_id'; END IF;
        IF v_has_tenant_id THEN v_tele_cols := v_tele_cols || ', tenant_id'; END IF;
        IF v_has_heading THEN v_tele_cols := v_tele_cols || ', heading'; END IF;
        IF v_has_ignition THEN v_tele_cols := v_tele_cols || ', ignition'; END IF;
        IF v_has_fuel THEN v_tele_cols := v_tele_cols || ', fuel_level'; END IF;
        IF v_has_trip THEN v_tele_cols := v_tele_cols || ', trip_id'; END IF;

        -- Build dynamic INSERT
        v_sql := 'INSERT INTO telemetry_points (' || v_tele_cols || ') SELECT ''550e8400-e29b-41d4-a716-446655440001''::uuid, 28.6139 + (gs * 0.001), 77.2090 - (gs * 0.0008), 40 + (gs % 20), NOW() - INTERVAL ''2 hours'' + (gs * INTERVAL ''1 minute'')';
        
        IF v_has_device_id THEN v_sql := v_sql || ', ''DEV-DEVICE-001'''; END IF;
        IF v_has_event_id THEN v_sql := v_sql || ', gen_random_uuid()'; END IF;
        IF v_has_tenant_id THEN v_sql := v_sql || ', ''550e8400-e29b-41d4-a716-446655440000''::uuid'; END IF;
        IF v_has_heading THEN v_sql := v_sql || ', 135 + (gs % 10)'; END IF;
        IF v_has_ignition THEN v_sql := v_sql || ', true'; END IF;
        IF v_has_fuel THEN v_sql := v_sql || ', 72.5 - (gs * 0.05)'; END IF;
        IF v_has_trip THEN v_sql := v_sql || ', ''550e8400-e29b-41d4-a716-446655440200''::uuid'; END IF;
        
        v_sql := v_sql || ' FROM generate_series(0, 119) AS gs';
        
        EXECUTE v_sql;
    END;
END $$;

-- ============================================================================
-- 6. GEOFENCES
-- ============================================================================
DO $$
DECLARE
    v_has_center_lat BOOLEAN;
    v_has_lat BOOLEAN;
    v_has_org BOOLEAN;
    v_has_tenant BOOLEAN;
    v_has_radius BOOLEAN;
    v_has_enabled BOOLEAN;
    v_geo_cols TEXT;
    v_geo_sql TEXT;
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
BEGIN
    DELETE FROM geofences WHERE id IN ('550e8400-e29b-41d4-a716-446655440100', '550e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440102', '550e8400-e29b-41d4-a716-446655440103');
    
    v_has_center_lat := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'geofences' AND column_name = 'center_lat');
    v_has_lat := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'geofences' AND column_name = 'latitude');
    v_has_org := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'geofences' AND column_name = 'org_id');
    v_has_tenant := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'geofences' AND column_name = 'tenant_id');
    v_has_radius := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'geofences' AND column_name = 'radius_m');
    v_has_enabled := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'geofences' AND column_name = 'enabled');

    v_geo_cols := 'id, name, type';
    IF v_has_center_lat THEN v_geo_cols := v_geo_cols || ', center_lat, center_lng'; END IF;
    IF v_has_lat THEN v_geo_cols := v_geo_cols || ', latitude, longitude'; END IF;
    IF v_has_org THEN v_geo_cols := v_geo_cols || ', org_id'; END IF;
    IF v_has_tenant THEN v_geo_cols := v_geo_cols || ', tenant_id'; END IF;
    IF v_has_radius THEN v_geo_cols := v_geo_cols || ', radius_m'; END IF;
    IF v_has_enabled THEN v_geo_cols := v_geo_cols || ', enabled'; END IF;

    -- Delhi Warehouse
    v_geo_sql := 'INSERT INTO geofences (' || v_geo_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440100'', ''Delhi Warehouse'', ''circle''';
    IF v_has_center_lat THEN v_geo_sql := v_geo_sql || ', 28.6139, 77.2090'; END IF;
    IF v_has_lat THEN v_geo_sql := v_geo_sql || ', 28.6139, 77.2090'; END IF;
    IF v_has_org THEN v_geo_sql := v_geo_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_tenant THEN v_geo_sql := v_geo_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_radius THEN v_geo_sql := v_geo_sql || ', 500'; END IF;
    IF v_has_enabled THEN v_geo_sql := v_geo_sql || ', true'; END IF;
    v_geo_sql := v_geo_sql || ') ON CONFLICT (id) DO NOTHING';
    EXECUTE v_geo_sql;

    -- Mumbai Depot
    v_geo_sql := 'INSERT INTO geofences (' || v_geo_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440101'', ''Mumbai Depot'', ''circle''';
    IF v_has_center_lat THEN v_geo_sql := v_geo_sql || ', 19.0760, 72.8777'; END IF;
    IF v_has_lat THEN v_geo_sql := v_geo_sql || ', 19.0760, 72.8777'; END IF;
    IF v_has_org THEN v_geo_sql := v_geo_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_tenant THEN v_geo_sql := v_geo_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_radius THEN v_geo_sql := v_geo_sql || ', 300'; END IF;
    IF v_has_enabled THEN v_geo_sql := v_geo_sql || ', true'; END IF;
    v_geo_sql := v_geo_sql || ') ON CONFLICT (id) DO NOTHING';
    EXECUTE v_geo_sql;
END $$;

COMMIT;
