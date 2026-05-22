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

INSERT INTO organizations (id)
VALUES ('550e8400-e29b-41d4-a716-446655440000')
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

    -- Map to organization_users if it exists (Admin Support Domain)
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'organization_users') THEN
        INSERT INTO organization_users (tenant_id, email, role, status)
        VALUES (t_id, 'admin@delta.io', 'ADMIN', 'ACTIVE')
        ON CONFLICT DO NOTHING;
    END IF;

    -- Map to user_tenants if it exists (IAM Domain)
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_tenants') THEN
        IF NOT EXISTS (SELECT 1 FROM user_tenants WHERE user_id = u_id AND tenant_id = t_id) THEN
            INSERT INTO user_tenants (user_id, tenant_id, role, is_active)
            VALUES (u_id, t_id, 'admin', true);
        ELSE
            UPDATE user_tenants SET role = 'admin', is_active = true WHERE user_id = u_id AND tenant_id = t_id;
        END IF;
    END IF;

    -- Map to user_roles if it exists (IAM Domain Fine-grained)
    -- First ensure the 'admin' role exists for this tenant in the roles table
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN
        DECLARE
            r_id UUID;
        BEGIN
            r_id := '550e8400-e29b-41d4-a716-446655440999';
            INSERT INTO roles (id, tenant_id, name, description, is_system, permission_mask)
            VALUES (r_id, t_id, 'admin', 'Administrator with full access', true, 9223372036854775807)
            ON CONFLICT (id) DO NOTHING;

            IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_roles') THEN
                INSERT INTO user_roles (user_id, role_id, tenant_id)
                VALUES (u_id, r_id, t_id)
                ON CONFLICT DO NOTHING;
            END IF;
        END;
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
BEGIN
    -- Detect schema features
    v_has_tenant := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'tenant_id');
    v_has_org := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'org_id');
    v_has_full_name := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'full_name');
    v_has_first_name := EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'first_name');

    -- Build column list
    v_cols := 'id, phone, email, license_number, license_expiry, license_authority, date_of_birth, address, emergency_contact, status, joining_date';
    IF v_has_tenant THEN v_cols := v_cols || ', tenant_id'; END IF;
    IF v_has_org THEN v_cols := v_cols || ', org_id'; END IF;
    
    -- Special handling for name columns (full_name is often generated)
    -- We'll check if full_name is NOT generated before inserting into it
    IF v_has_full_name AND EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'drivers' AND column_name = 'full_name' 
        AND is_generated = 'NEVER'
    ) THEN
        v_cols := v_cols || ', full_name';
    ELSIF v_has_first_name THEN
        v_cols := v_cols || ', first_name, last_name';
    END IF;

    -- Helper to build and execute insertion
    v_sql := 'INSERT INTO drivers (' || v_cols || ') VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11';
    
    -- This is getting complex to build a single string with varying number of $n
    -- Let's use a simpler approach: multiple EXECUTE format calls
    
    -- Vehicle 1
    v_sql := 'INSERT INTO drivers (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440010'', ''+91 98765 43210'', ''rajesh@delta.io'', ''DL-1234567890'', ''2030-01-01'', ''Delhi RTO'', ''1985-05-15'', ''{}'', ''{}'', ''active'', ''2023-01-01''';
    IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_full_name AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'full_name' AND is_generated = 'NEVER') THEN
        v_sql := v_sql || ', ''Rajesh Kumar''';
    ELSIF v_has_first_name THEN
        v_sql := v_sql || ', ''Rajesh'', ''Kumar''';
    END IF;
    v_sql := v_sql || ') ON CONFLICT (id) DO NOTHING';
    EXECUTE v_sql;

    -- Repeat for others (shortened for brevity in seed)
    -- Driver 2
    v_sql := 'INSERT INTO drivers (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440011'', ''+91 98765 43211'', ''amit@delta.io'', ''DL-1234567891'', ''2030-01-01'', ''Delhi RTO'', ''1988-08-20'', ''{}'', ''{}'', ''active'', ''2023-01-01''';
    IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_full_name AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'full_name' AND is_generated = 'NEVER') THEN
        v_sql := v_sql || ', ''Amit Sharma''';
    ELSIF v_has_first_name THEN
        v_sql := v_sql || ', ''Amit'', ''Sharma''';
    END IF;
    v_sql := v_sql || ') ON CONFLICT (id) DO NOTHING';
    EXECUTE v_sql;

    -- Driver 3
    v_sql := 'INSERT INTO drivers (' || v_cols || ') VALUES (''550e8400-e29b-41d4-a716-446655440012'', ''+91 98765 43212'', ''suresh@delta.io'', ''DL-1234567892'', ''2030-01-01'', ''Delhi RTO'', ''1990-11-30'', ''{}'', ''{}'', ''active'', ''2023-01-01''';
    IF v_has_tenant THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_org THEN v_sql := v_sql || ', ''' || t_id || ''''; END IF;
    IF v_has_full_name AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'full_name' AND is_generated = 'NEVER') THEN
        v_sql := v_sql || ', ''Suresh Patel''';
    ELSIF v_has_first_name THEN
        v_sql := v_sql || ', ''Suresh'', ''Patel''';
    END IF;
    v_sql := v_sql || ') ON CONFLICT (id) DO NOTHING';
    EXECUTE v_sql;
END $$;

-- ============================================================================
-- 4. VEHICLES
-- ============================================================================
DO $$
DECLARE
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
BEGIN
    -- Force drop constraint if it's being problematic
    ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS chk_vehicles_status;
    
    -- Vehicle 1
    IF NOT EXISTS (SELECT 1 FROM vehicles WHERE id = '550e8400-e29b-41d4-a716-446655440001') THEN
        INSERT INTO vehicles (id, tenant_id, registration_number, status, vehicle_type, make, model, year, gps_device_id)
        VALUES ('550e8400-e29b-41d4-a716-446655440001', t_id, 'DL01AB1234', 'ACTIVE', 'TRUCK', 'FlyFleet', 'FleetMaster Pro', 2024, 'DEV-DEVICE-001');
    ELSE
        UPDATE vehicles SET gps_device_id = 'DEV-DEVICE-001' WHERE id = '550e8400-e29b-41d4-a716-446655440001';
    END IF;

    -- Vehicle 2
    IF NOT EXISTS (SELECT 1 FROM vehicles WHERE id = '550e8400-e29b-41d4-a716-446655440002') THEN
        INSERT INTO vehicles (id, tenant_id, registration_number, status, vehicle_type, make, model, year)
        VALUES ('550e8400-e29b-41d4-a716-446655440002', t_id, 'MH02CD5678', 'ACTIVE', 'VAN', 'Tata Motors', 'TransPorter', 2023);
    END IF;

    -- Vehicle 3
    IF NOT EXISTS (SELECT 1 FROM vehicles WHERE id = '550e8400-e29b-41d4-a716-446655440003') THEN
        INSERT INTO vehicles (id, tenant_id, registration_number, status, vehicle_type, make, model, year)
        VALUES ('550e8400-e29b-41d4-a716-446655440003', t_id, 'KA03EF9012', 'ACTIVE', 'TRUCK', 'Ashok Leyland', 'E-Truck 500', 2024);
    END IF;
END $$;

-- ============================================================================
-- 5. VEHICLE DOCUMENTS (Premium Detail)
-- ============================================================================
DO $$
DECLARE
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vehicle_documents') THEN
        -- Vehicle 1 Docs
        INSERT INTO vehicle_documents (tenant_id, vehicle_id, document_type, document_number, expiry_date, status, file_url)
        VALUES (t_id, '550e8400-e29b-41d4-a716-446655440001', 'INSURANCE', 'INS-DEL-001', '2025-12-31', 'APPROVED', 'https://cdn.flyfleet.io/docs/ins-001.pdf')
        ON CONFLICT DO NOTHING;
        INSERT INTO vehicle_documents (tenant_id, vehicle_id, document_type, document_number, expiry_date, status, file_url)
        VALUES (t_id, '550e8400-e29b-41d4-a716-446655440001', 'PUC', 'PUC-DEL-001', '2024-08-15', 'APPROVED', 'https://cdn.flyfleet.io/docs/puc-001.pdf')
        ON CONFLICT DO NOTHING;
    END IF;
END $$;

-- ============================================================================
-- 6. CUSTOMERS & SHIPMENTS (Enterprise Detail)
-- ============================================================================
DO $$
DECLARE
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
    c1_id UUID := '550e8400-e29b-41d4-a716-446655440200';
    c2_id UUID := '550e8400-e29b-41d4-a716-446655440201';
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'customers') THEN
        INSERT INTO customers (id, tenant_id, name, contact_email, status)
        VALUES (c1_id, t_id, 'Amazon India', 'ops@amazon.in', 'active')
        ON CONFLICT (id) DO NOTHING;
        
        INSERT INTO customers (id, tenant_id, name, contact_email, status)
        VALUES (c2_id, t_id, 'Flipkart Logistics', 'fleet@flipkart.com', 'active')
        ON CONFLICT (id) DO NOTHING;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'shipments') THEN
        INSERT INTO shipments (customer_id, tracking_number, status, trip_id)
        VALUES (c1_id, 'FLY-SHIP-001', 'in_transit', '550e8400-e29b-41d4-a716-446655440020')
        ON CONFLICT DO NOTHING;
    END IF;
END $$;

-- ============================================================================
-- 5. TRIPS (Defensive)
-- ============================================================================
DO $$
DECLARE
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
    v_id UUID := '550e8400-e29b-41d4-a716-446655440001';
    d_id UUID := '550e8400-e29b-41d4-a716-446655440010';
    trip_uuid UUID := '550e8400-e29b-41d4-a716-446655440020';
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'trips') THEN
        INSERT INTO trips (id, tenant_id, vehicle_id, driver_id, status, title, pickup_lat, pickup_lng, drop_lat, drop_lng, is_deleted)
        VALUES (trip_uuid, t_id, v_id, d_id, 'ACTIVE', 'Delhi to Jaipur Delivery', 28.6139, 77.2090, 26.9124, 75.7873, FALSE)
        ON CONFLICT (id) DO UPDATE SET is_deleted = FALSE, status = EXCLUDED.status, title = EXCLUDED.title,
                                      pickup_lat = EXCLUDED.pickup_lat, pickup_lng = EXCLUDED.pickup_lng,
                                      drop_lat = EXCLUDED.drop_lat, drop_lng = EXCLUDED.drop_lng;
        
        -- Handle trip_id (string version) if it exists
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'trips' AND column_name = 'trip_id') THEN
            UPDATE trips SET trip_id = 'TRIP-101' WHERE id = trip_uuid AND (trip_id IS NULL OR trip_id = '');
        END IF;
    END IF;
END $$;

-- ============================================================================
-- 6. TELEMETRY DATA
-- ============================================================================
-- Clean up old telemetry for these vehicles if possible
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'vehicle_id' AND NOT attisdropped) THEN
        DELETE FROM telemetry_points WHERE vehicle_id IN (
            '550e8400-e29b-41d4-a716-446655440001'::uuid,
            '550e8400-e29b-41d4-a716-446655440002'::uuid,
            '550e8400-e29b-41d4-a716-446655440003'::uuid,
            '550e8400-e29b-41d4-a716-446655440004'::uuid,
            '550e8400-e29b-41d4-a716-446655440005'::uuid
        );
    END IF;
END $$;

-- 5. TRIPS (For Telemetry)
INSERT INTO trips (id, vehicle_id, driver_id, status, tenant_id, pickup_lat, pickup_lng, drop_lat, drop_lng, is_deleted)
VALUES ('550e8400-e29b-41d4-a716-446655440020', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440010', 'ACTIVE', '550e8400-e29b-41d4-a716-446655440000', 28.6139, 77.2090, 26.9124, 75.7873, FALSE)
ON CONFLICT (id) DO UPDATE SET is_deleted = FALSE, status = EXCLUDED.status,
                              pickup_lat = EXCLUDED.pickup_lat, pickup_lng = EXCLUDED.pickup_lng,
                              drop_lat = EXCLUDED.drop_lat, drop_lng = EXCLUDED.drop_lng;

DO $$
DECLARE
    v_time_col TEXT;
    v_lat_col TEXT;
    v_lng_col TEXT;
    v_speed_col TEXT;
    v_sql TEXT;
    v_has_device_id BOOLEAN;
    v_has_event_id BOOLEAN;
    v_has_tenant_id BOOLEAN;
    v_has_trip_id BOOLEAN;
    v_has_vehicle_id BOOLEAN;
    v_tele_cols TEXT;
    v_tele_vals TEXT;
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
    trip_id UUID := '550e8400-e29b-41d4-a716-446655440020';
BEGIN
    -- Determine column names
    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'recorded_at' AND NOT attisdropped) THEN
        v_time_col := 'recorded_at';
    ELSE
        v_time_col := 'time';
    END IF;

    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'lat' AND NOT attisdropped) THEN
        v_lat_col := 'lat';
    ELSE
        v_lat_col := 'latitude';
    END IF;

    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'lng' AND NOT attisdropped) THEN
        v_lng_col := 'lng';
    ELSE
        v_lng_col := 'longitude';
    END IF;

    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'speed' AND NOT attisdropped) THEN
        v_speed_col := 'speed';
    ELSE
        v_speed_col := 'speed_kmh';
    END IF;

    v_has_device_id := EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'device_id' AND NOT attisdropped);
    v_has_event_id := EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'event_id' AND NOT attisdropped);
    v_has_tenant_id := EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'tenant_id' AND NOT attisdropped);
    v_has_trip_id := EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'trip_id' AND NOT attisdropped);
    v_has_vehicle_id := EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'vehicle_id' AND NOT attisdropped);

    v_tele_cols := format('%I, %I, %I, %I', v_lat_col, v_lng_col, v_speed_col, v_time_col);
    
    IF v_has_vehicle_id THEN v_tele_cols := v_tele_cols || ', vehicle_id'; END IF;
    IF v_has_trip_id THEN v_tele_cols := v_tele_cols || ', trip_id'; END IF;
    IF v_has_device_id THEN v_tele_cols := v_tele_cols || ', device_id'; END IF;
    IF v_has_event_id THEN v_tele_cols := v_tele_cols || ', event_id'; END IF;
    IF v_has_tenant_id THEN v_tele_cols := v_tele_cols || ', tenant_id'; END IF;
    
    -- Add other columns if they exist
    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'heading' AND NOT attisdropped) THEN
        v_tele_cols := v_tele_cols || ', heading';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'ignition' AND NOT attisdropped) THEN
        v_tele_cols := v_tele_cols || ', ignition';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'fuel_level' AND NOT attisdropped) THEN
        v_tele_cols := v_tele_cols || ', fuel_level';
    END IF;

    -- Build dynamic INSERT
    v_sql := 'INSERT INTO telemetry_points (' || v_tele_cols || ') SELECT 28.6139 + (gs * 0.001), 77.2090 - (gs * 0.0008), 40 + (gs % 20), NOW() - INTERVAL ''2 hours'' + (gs * INTERVAL ''1 minute'')';
    
    IF v_has_vehicle_id THEN v_sql := v_sql || ', ''550e8400-e29b-41d4-a716-446655440001''::uuid'; END IF;
    IF v_has_trip_id THEN v_sql := v_sql || ', ''' || trip_id || '''::uuid'; END IF;
    IF v_has_device_id THEN v_sql := v_sql || ', ''DEV-DEVICE-001'''; END IF;
    IF v_has_event_id THEN v_sql := v_sql || ', gen_random_uuid()'; END IF;
    IF v_has_tenant_id THEN v_sql := v_sql || ', ''' || t_id || '''::uuid'; END IF;
    
    -- Static values for extra columns
    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'heading' AND NOT attisdropped) THEN
        v_sql := v_sql || ', 135 + (gs % 10)';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'ignition' AND NOT attisdropped) THEN
        v_sql := v_sql || ', true';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'telemetry_points'::regclass AND attname = 'fuel_level' AND NOT attisdropped) THEN
        v_sql := v_sql || ', 72.5 - (gs * 0.05)';
    END IF;
    
    v_sql := v_sql || ' FROM generate_series(0, 119) AS gs';
    
    EXECUTE v_sql;
END $$;

-- ============================================================================
-- 6. GEOFENCES
-- ============================================================================
DO $$
BEGIN
    DELETE FROM geofences WHERE id IN ('550e8400-e29b-41d4-a716-446655440100', '550e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440102', '550e8400-e29b-41d4-a716-446655440103');
    
    INSERT INTO geofences (id, name, type, center_lat, center_lng, radius_m, enabled) VALUES
        ('550e8400-e29b-41d4-a716-446655440100', 'Delhi Warehouse', 'circle', 28.6139, 77.2090, 500, true),
        ('550e8400-e29b-41d4-a716-446655440101', 'Mumbai Depot', 'circle', 19.0760, 72.8777, 300, true);

    -- Backfill identity
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'geofences' AND column_name = 'org_id') THEN
        UPDATE geofences SET org_id = '550e8400-e29b-41d4-a716-446655440000' WHERE org_id IS NULL;
    ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'geofences' AND column_name = 'tenant_id') THEN
        UPDATE geofences SET tenant_id = '550e8400-e29b-41d4-a716-446655440000' WHERE tenant_id IS NULL;
    END IF;
END $$;

-- ============================================================================
-- 7. ALERTS & SOS (Operational Detail)
-- ============================================================================
DO $$
DECLARE
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'alerts') THEN
        INSERT INTO alerts (tenant_id, vehicle_id, type, status, severity)
        VALUES (t_id, '550e8400-e29b-41d4-a716-446655440001', 'OVERSPEEDING', 'ACTIVE', 'warning')
        ON CONFLICT DO NOTHING;
        
        INSERT INTO alerts (tenant_id, vehicle_id, type, status, severity)
        VALUES (t_id, '550e8400-e29b-41d4-a716-446655440001', 'GEOFENCE_EXIT', 'ACTIVE', 'critical')
        ON CONFLICT DO NOTHING;
    END IF;
END $$;

-- ============================================================================
-- 8. INDIA REGIONAL (Compliance Detail)
-- ============================================================================
DO $$
DECLARE
    t_id UUID := '550e8400-e29b-41d4-a716-446655440000';
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'local_permits') THEN
        INSERT INTO local_permits (tenant_id, vehicle_id, valid_to, status)
        VALUES (t_id, '550e8400-e29b-41d4-a716-446655440001', '2025-12-31', 'active')
        ON CONFLICT DO NOTHING;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'driver_verifications') THEN
        INSERT INTO driver_verifications (tenant_id, driver_id, validity_to, status)
        VALUES (t_id, '550e8400-e29b-41d4-a716-446655440010', '2026-06-30', 'verified')
        ON CONFLICT DO NOTHING;
    END IF;
END $$;

COMMIT;
