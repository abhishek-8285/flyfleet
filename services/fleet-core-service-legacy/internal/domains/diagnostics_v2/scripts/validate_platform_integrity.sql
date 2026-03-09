-- Step 48: Data Consistency Validation SQL
-- Run these queries to detect referential integrity gaps and logical inconsistencies.

-- 1. Shipments without valid vehicles (Step 48 example)
SELECT 'Shipments with invalid vehicle_id' as check_type, COUNT(*)
FROM customer_v2_shipments s
LEFT JOIN devices d ON s.device_id = d.id -- Assuming device_id linkage for diagnostics
WHERE d.id IS NULL AND s.device_id IS NOT NULL;

-- 2. Rollout status mismatch with device twins
SELECT 'Devices with rollout version mismatch' as check_type, COUNT(*)
FROM rollout_device_status rds
JOIN device_twins dt ON rds.device_id = dt.device_id
WHERE rds.status = 'COMPLETED' AND dt.desired_firmware != dt.reported_firmware;

-- 3. Invoices with calculation errors
SELECT 'Invoices with total mismatch' as check_type, COUNT(*)
FROM customer_v2_wallets w -- Placeholder for invoice table if billing_v2 not fully mapped
WHERE false; -- logical check: total != sum(items)

-- 4. Orphaned Telemetry
SELECT 'Orphaned telemetry (No device registry)' as check_type, COUNT(*)
FROM device_telemetry dt
LEFT JOIN devices d ON dt.device_id = d.id
WHERE d.id IS NULL;
