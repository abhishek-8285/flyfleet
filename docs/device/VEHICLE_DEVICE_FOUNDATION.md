# Vehicle Device Foundation

## Runtime State Contract

The vehicle device runtime follows a deterministic state machine with explicit transitions:

- `BOOTING`: service/process startup, bootstrap checks, and secure registration.
- `IDLE`: default low-intensity state when ignition is off and not in depot sync mode.
- `ACTIVE_DRIVE`: high-frequency capture and inference mode while vehicle is driving.
- `DEPOT_MODE`: controlled stationary mode for maintenance actions and unmetered sync.
- `DEGRADED`: entered when critical sensor or module failures occur.
- `RECOVERY`: bounded attempt window after module recovery signals.
- `SAFE_SHUTDOWN`: terminal state to preserve data and stop unsafe operations.

All states must transition to `SAFE_SHUTDOWN` via `SHUTDOWN_REQUESTED`.

## Configuration Contract

Runtime configuration contract is versioned and validated with:

- `apps/mobile/vehicle-device/spec/device_runtime_config.schema.json`
- `apps/mobile/vehicle-device/spec/device_runtime_config.sample.json`

Required control envelope:

- heartbeat: `30-60` seconds.
- ring buffer: `5-15` minutes.
- gps active interval: `1-2` seconds.
- gps idle interval: `15-60` seconds.
- AI severity thresholds must preserve `critical > warn` ordering.

## Event Priority Contract

Event taxonomy is defined in `apps/mobile/vehicle-device/spec/event_taxonomy.json`.

Tiering minimum:

- Tier1: includes `panic_triggered`, `harsh_brake_detected`, `speed_spike_detected`.
- Tier2/Tier3: non-life-critical and operational events.

Events must not be assigned to multiple tiers.
All declared event types must be assigned to exactly one tier.

## Offline Queue Contract

Offline upload queue behavior is deterministic and tested:

- queue acquisition order: tier priority first, then FIFO within same tier.
- failures return items to pending with exponential backoff.
- max retry exhaustion transitions item to `DEAD_LETTER`.

## Event Store Contract

Durable event store is journal-backed and restart-safe:

- append persists events with monotonic sequence IDs.
- pending replay order is deterministic: tier priority, then FIFO.
- uploaded marking updates state without losing chronological replay history.
- replay can resume from a sequence checkpoint after process restart.

## Sync Pipeline Contract

Sync orchestration links durable store, queue, transport, and forensics:

- ingest writes to durable store first, then enqueues upload work.
- successful upload marks event uploaded, removes queue item, appends evidence hash-chain record.
- failed upload preserves retriable queue state with error context.
- queue item without backing durable event is treated as failure.

## Forensics Hash Chain Contract

Evidence records are hash-chained with SHA-256:

- each record hash includes previous chain hash.
- chain verification fails on payload tamper.
- chain verification fails on broken previous-hash linkage.

## Heartbeat Contract

Heartbeat snapshots include runtime state and module health matrix:

- modules: camera, gps, canBus, network, ai, storage.
- overall health derives deterministically:
  - `OFFLINE` if any module offline
  - `DEGRADED` if none offline and any degraded
  - `HEALTHY` otherwise
- retention metrics expose `retentionAllowed`/`retentionBlocked` counts per snapshot.

## Command Safety Contract

Remote commands are guarded by lifecycle validation and apply-time preconditions:

- lifecycle transitions must follow canonical states:
  `created -> queued -> delivered -> acked -> applied|failed` and `delivered -> timed_out`.
- command apply decisions use runtime context (`state`, `speed`).
- risky commands (reboot, AI toggle, force sync, camera control) are blocked while moving or outside safe states.
- panic alert test command remains available as an emergency-safe path.

## Runtime Coordinator Contract

A runtime coordinator composes core modules into deterministic flows:

- command decision evaluation uses live runtime state from monitoring service.
- upload processing delegates through sync orchestrator and exposes step outcomes.
- heartbeat generation includes runtime state plus queue/dead-letter counts.

## Vehicle Telemetry Contract

Vehicle telemetry adapters support direct CAN, OBD, and fallback modes:

- direct CAN path reports high-confidence telemetry.
- OBD path can operate with reduced fields and confidence tagging.
- when CAN is unavailable, fallback uses GPS/IMU consistency scoring.
- fallback confidence is explicit (`MEDIUM` or `LOW`) with machine-readable reason.

## Telemetry/Drift Propagation Contract

Runtime coordinator propagates telemetry and drift into heartbeat and event stream:

- heartbeat includes integration mode, telemetry confidence/reason, drift status/reason, mismatch count.
- low telemetry confidence emits `can_bus_degraded` operational sync event.
- drifted twin state emits `twin_drift_detected` operational sync event.

## Audit Envelope Contract

Command decisions and operational signals are audit-enveloped with signatures:

- every command decision emits a signed command audit envelope.
- operational signals emitted by coordinator also produce signed audit envelopes.
- audit verification fails if envelope payload is tampered.

## Command Envelope Contract

Signed remote command envelopes are validated before command safety evaluation:

- envelope signature must verify against canonical payload.
- command timestamp must be within accepted skew window.
- nonce replay is rejected deterministically.
- duplicate command IDs are rejected for idempotency protection.

## Incident Bundle Contract

Forensic export bundles are signed and reproducible:

- bundle manifest includes counts and digests for events, chain, and audit envelopes.
- manifest signature verifies bundle integrity.
- verification fails on tampered event payloads.
- verification fails when evidence chain linkage is broken.

## Retention Policy Contract

Evidence retention follows a deterministic policy:

- critical clips kept for `criticalRetentionDays`; default 30.
- `legalHold` prevents any overwrite until cleared.
- evidence reserve (`evidenceReserveMb`) maintains 5+ GB of capacity.
- overwrites only happen when age > retention + reserve satisfied.

## Retention Config Contract

Retention configuration is loaded from runtime config (e.g., `retention` block):

- `criticalRetentionDays` defines the minimum lifetime for evidence clips.
- `evidenceReserveMb` maintains free space for incident artifacts.
- `legalHold` overrides deletes until cleared.
- `RetentionSupervisor` reloads retention config each run so legal-hold/reserve changes are applied immediately.
- `RetentionCleanupWorker` tracks metrics so you can monitor how often deletes are allowed vs. blocked.
- `legal_hold` (JSON field) mirrors `legalHold` flag used in retention policy.

## Twin Drift Contract

Digital twin drift detection compares desired and reported runtime controls:

- missing desired/reported data yields `UNKNOWN` drift status.
- exact field match yields `IN_SYNC`.
- mismatches yield `DRIFTED` with explicit mismatched field list.

## Feature Flag Contract

Feature flags must support deterministic block/unblock behavior with safety fallbacks.

Resolution precedence is strict:

- `global` -> `fleet` -> `vehicle` -> `driver` -> `device` -> `localEmergencyOverride`.

Safety constraints:

- remote config must be signature-verified before use.
- signature verification failure must fall back to last-known-good config.
- signed config rollback is blocked (`remote_version < last_known_good_version` is rejected).
- local emergency overrides are allowed to disable risky features immediately.

Enforced runtime policy checks:

- camera/AI execution depends on both feature flags and runtime state.
- UI interaction lockout depends on `driver_lockout_enabled` and speed threshold.
- audio event checks are restricted to active-drive state.

## Anti-Breakage Gates

The following check must pass before merge:

- `python3 scripts/check_vehicle_device_foundation.py`

This gate verifies:

- runtime state/signal coverage expected by current architecture.
- service API handlers for depot and recovery lifecycle hooks.
- schema and sample config semantic bounds.
- event taxonomy tiering consistency.
- typed event priority resolver coverage (`tier1/tier2/tier3`) for queue ordering.

## Acceptance Criteria

A change is acceptable only when all of the following hold:

- foundation checker reports `vehicle_device_foundation=PASS`.
- no runtime state contract regressions in unit tests.
- feature flag precedence and signature-fallback tests remain green.
- spec and implementation remain aligned for depot/recovery behavior.
