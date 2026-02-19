# Firmware OTA Specification

## Scope
Defines secure firmware release, verification, rollout, rollback, and observability across all supported device classes.

## Goals
1. Prevent unauthorized firmware installation.
2. Minimize blast radius through staged rollout.
3. Guarantee rollback path for failed updates.
4. Provide full auditability for compliance and incident response.

## Artifact Model
Each release includes:
1. Firmware binary artifact.
2. Signed manifest (`ota-manifest.json`) following `docs/firmware/ota-manifest.schema.json`.
3. Key registry reference in `firmware/keys/trusted_keys.json` (maps `signature.key_id` to trusted public key material).
4. Release notes with known issues and compatibility constraints.

## Security Requirements
1. Firmware must be signed with vendor private key (offline/HSM managed).
2. Device verifies signature before install.
3. Hash algorithm: SHA-256 minimum.
4. Manifest signature verification failure must hard-stop update.
5. Downgrade protection: device rejects firmware below minimum permitted version unless forced rollback flag is present.

## Manifest Signature Contract
1. `signature.value` must be Base64-encoded raw signature bytes.
2. `signature.key_id` must exist in `firmware/keys/trusted_keys.json`.
3. Current CI/runtime verifier supports:
   - `ed25519` via public key verification (`openssl pkeyutl -verify`).
4. Canonical signed payload is JSON with sorted keys and no whitespace over these fields only:
   - `release_id`
   - `firmware_version`
   - `min_bootloader_version`
   - `hardware_compatibility`
   - `artifact.uri`
   - `artifact.size_bytes`
   - `artifact.sha256`
5. Any payload field mismatch (including artifact hash/size drift) invalidates signature verification.

## Device Preconditions
1. Battery threshold met (or external power available).
2. Storage threshold met for A/B image staging.
3. Connectivity stable enough for download windows.
4. Hardware and bootloader compatibility validated against manifest.

## Rollout Rings
1. Ring 0: lab/internal devices.
2. Ring 1: canary cohort (1-2% representative fleet).
3. Ring 2: progressive rollout (5%, then 25%).
4. Ring 3: full rollout (100%).

Promotion rules:
1. Ring success rate >= target threshold.
2. No active stop conditions for at least soak duration.
3. Approval recorded with actor, timestamp, and evidence.

Automation hook:
1. `scripts/ota_rollout_orchestrator.py` evaluates ring telemetry against `rollout_policy.auto_pause_thresholds`.
2. Decision outputs:
   - `promote` -> proceed to next ring.
   - `pause_and_rollback` -> hard stop and rollback required for affected cohort.
3. Baseline scenarios are maintained in `firmware/releases/sample/ota-rollout-scenarios.json` and must pass in CI.

## Stop Conditions (Auto-Pause)
1. Install failure rate exceeds threshold.
2. Boot failure/watchdog reboot rate exceeds threshold.
3. Post-update telemetry silence exceeds threshold.
4. Critical regression alert triggered (safety-related diagnostics).

## Rollback Strategy
1. A/B partition required for supported devices.
2. On boot failure or health-check timeout, auto-revert to previous slot.
3. Control plane can force rollback by cohort/device.
4. Rollback events must be audited and visible in operator console.

## Device OTA State Machine
1. `scheduled`
2. `download_started`
3. `download_completed`
4. `signature_verified`
5. `install_started`
6. `install_completed`
7. `rebooting`
8. `post_boot_healthcheck`
9. `success` or `failed` or `rolled_back`

Each transition emits a status event with reason code.

## Reason Code Taxonomy
- `download_timeout`
- `manifest_invalid`
- `signature_invalid`
- `insufficient_battery`
- `insufficient_storage`
- `incompatible_hardware`
- `install_failure`
- `boot_failure`
- `post_boot_healthcheck_failed`

## Control Plane APIs (Conceptual)
1. Create rollout batch.
2. Approve/pause/resume rollout.
3. Force rollback.
4. Query per-device OTA timeline.
5. Query aggregated OTA metrics per cohort/version.

## Current Platform Implementation
The following APIs and UI are now wired and usable end-to-end through the Web BFF:

1. Create firmware package
   - BFF route: `POST /api/firmware`
   - Core route: `POST /api/v1/diagnostics/firmware`
2. List firmware packages
   - BFF route: `GET /api/firmware`
   - Core route: `GET /api/v1/diagnostics/firmware`
3. Create rollout campaign
   - BFF route: `POST /api/firmware/rollouts`
   - Core route: `POST /api/v1/diagnostics/rollouts`
4. List rollout campaigns
   - BFF route: `GET /api/firmware/rollouts`
   - Core route: `GET /api/v1/diagnostics/rollouts`
5. Get rollout campaign by ID
   - BFF route: `GET /api/firmware/rollouts/{rollout_id}`
   - Core route: `GET /api/v1/diagnostics/rollouts/{rollout_id}`
6. Transition rollout lifecycle
   - BFF route: `PATCH /api/firmware/rollouts/{rollout_id}/status`
   - Core route: `PATCH /api/v1/diagnostics/rollouts/{rollout_id}/status`
   - Supported transitions:
     - `scheduled -> in_progress|cancelled`
     - `in_progress -> paused|completed|cancelled`
     - `paused -> in_progress|cancelled`
7. Rollout event history (audit trail)
   - BFF route: `GET /api/firmware/rollouts/{rollout_id}/events`
   - Core route: `GET /api/v1/diagnostics/rollouts/{rollout_id}/events`
   - Query parameters:
     - `limit` (default `50`, minimum `1`, maximum `500`)
     - `offset`
     - `event_type` filter: `created|transition|note` (aliases) or `rollout.created|rollout.transition|rollout.note`
   - Includes create and status transition events with timestamps and metadata.
   - Transition events capture optional operator reason and actor user ID when provided.
8. Manual rollout notes
   - BFF route: `POST /api/firmware/rollouts/{rollout_id}/events`
   - Core route: `POST /api/v1/diagnostics/rollouts/{rollout_id}/events`
   - Appends operator-authored notes into the same rollout event timeline for auditable handoffs.

Validation currently enforced server-side:
1. Firmware create requires `version`, `hardware_target`, `file_url`, `file_size > 0`.
2. Rollout create requires `firmware_id`, `name`, `device_group`, `total_devices > 0`.
3. Rollout create verifies firmware package existence before insert.

Operator UI:
1. `apps/dashboard-web/src/pages/devices/FirmwareManager.tsx`
2. Supports upload firmware package + create rollout directly from package card.
3. Displays rollout progress and failed-device counts.
4. Rollout detail panel supports timeline filtering (`All`, `Transitions`, `Created`, `Notes`) and manual note entry.

## Observability Requirements
1. Metrics:
   - rollout success rate
   - install latency percentiles
   - rollback rate
   - failure reason distribution
2. Logs:
   - signed, immutable rollout action logs
3. Tracing:
   - link rollout command to device result path via trace ID

## SLO Targets
1. OTA success >= 98% per batch.
2. Unplanned rollback <= 1% per batch.
3. P95 install completion <= 10 minutes (network permitting).

## Acceptance Tests
1. Unsigned artifact is rejected.
2. Invalid signature is rejected.
3. Incompatible hardware is rejected.
4. Canary failure triggers auto-pause.
5. Failed boot triggers automatic rollback.
6. Operator forced rollback completes within defined SLA.

## Compliance and Audit
1. Every rollout action includes actor identity and justification.
2. Manifest, binary hash, and signature retained for compliance retention period.
3. Incident postmortem must link to rollout audit IDs.
