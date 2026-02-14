-- Seed Telemetry for Active Trip (DLT-001)

-- 1. Start point
INSERT INTO telemetry (trip_id, latitude, longitude, speed, heading, event_type, event_time)
SELECT 
    id, 
    12.9716, 77.5946, 
    0.0, 0.0, 
    'start', 
    NOW() - INTERVAL '2 hours'
FROM trips 
WHERE vehicle_id = (SELECT id FROM vehicles WHERE license_plate = 'DLT-001') 
AND status = 'in_progress';

-- 2. Moving
INSERT INTO telemetry (trip_id, latitude, longitude, speed, heading, event_type, event_time)
SELECT 
    id, 
    12.9650, 77.6000, 
    45.0, 135.0, 
    'normal', 
    NOW() - INTERVAL '1 hour 50 minutes'
FROM trips 
WHERE vehicle_id = (SELECT id FROM vehicles WHERE license_plate = 'DLT-001') 
AND status = 'in_progress';

-- 3. Overspeed Event (Matches the Alert)
INSERT INTO telemetry (trip_id, latitude, longitude, speed, heading, event_type, event_time)
SELECT 
    id, 
    12.9716, 77.5946, -- Using same location as alert roughly
    85.0, 140.0, 
    'overspeed', 
    NOW() - INTERVAL '1 hour'
FROM trips 
WHERE vehicle_id = (SELECT id FROM vehicles WHERE license_plate = 'DLT-001') 
AND status = 'in_progress';

-- 4. Current Position (Latest)
INSERT INTO telemetry (trip_id, latitude, longitude, speed, heading, event_type, event_time)
SELECT 
    id, 
    12.9300, 77.6300, 
    60.0, 145.0, 
    'normal', 
    NOW()
FROM trips 
WHERE vehicle_id = (SELECT id FROM vehicles WHERE license_plate = 'DLT-001') 
AND status = 'in_progress';
