# Runbook: Firmware Rollout Failure Recovery

## Scenario

A firmware rollout has exceeded the failure threshold (5%) and has been automatically paused by the safety engine.

## Immediate Actions

1. **Identify Failure Cause**:
   - Check Prometheus alert: `FirmwareRolloutFailure`.
   - Query logs: `diagnostics_audit_logs` where `action='ROLLOUT_PAUSED'`.
   - View device error logs: `device_diagnostics_logs`.
2. **Perform Root Cause Analysis (RCA)**:
   - Is it a specific device model?
   - Is it a network-wide failure?
   - Is the firmware signature invalid?

## Recovery Steps

1. **Rollback if necessary**:
   - Use the `UpdateRolloutStatus` API to set status to `CANCELLED`.
   - Trigger a new rollout with the previous stable version if critical fix is needed.
2. **Fix and Resume**:
   - Once the issue is identified and resolved in a new patch, create a new rollout.
   - Do NOT resume a failed rollout without code changes.

## Verification

- Monitor `diagnostics_rollouts_active` metrics.
- Ensure `diagnostics_commands_failed_total` stabilizes.
