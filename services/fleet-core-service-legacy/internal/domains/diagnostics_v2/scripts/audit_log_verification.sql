-- Step 53: Compliance Logging Audit SQL
-- Verify that critical system actions are correctly recorded.

-- 1. Firmware update audits
SELECT 'Firmware updates missing audit' as check_type, COUNT(*)
FROM firmware_rollout_events fre
LEFT JOIN diagnostics_audit_logs dal ON fre.rollout_id::text = dal.entity_id
WHERE dal.id IS NULL;

-- 2. Device status change audits
SELECT 'Device status changes logged' as check_type, COUNT(*)
FROM diagnostics_audit_logs
WHERE entity_type = 'devices' AND (action = 'INSERT' OR action = 'UPDATE');

-- 3. Critical alert triggers
SELECT 'Alert audits' as check_type, COUNT(*)
FROM diagnostics_audit_logs
WHERE entity_type = 'device_alerts';
