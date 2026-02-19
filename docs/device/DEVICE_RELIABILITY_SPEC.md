# Device Reliability Specification

## Scope
Defines digital twin and command lifecycle reliability requirements across gateway, backend, and operator surfaces.

## Digital Twin Model
Each device twin tracks:
- `device_id`
- `desired_state`
- `reported_state`
- `last_seen`
- `drift_status` (`in_sync`, `drifted`, `unknown`)
- `drift_reason`

## Command Lifecycle
Canonical states:
1. `created`
2. `queued`
3. `delivered`
4. `acked`
5. `applied` or `failed` or `timed_out`

Allowed transitions:
- `created -> queued`
- `queued -> delivered`
- `delivered -> acked`
- `acked -> applied`
- `acked -> failed`
- `delivered -> timed_out`
- `queued -> failed`

## Idempotency Rules
1. Command IDs must be globally unique.
2. Duplicate command submissions with same idempotency key must not create duplicate side effects.
3. Retries must preserve command ID and trace context.

## SLA Targets
1. Command delivery success >= 99.5%.
2. P95 ack latency <= 5s.
3. Twin drift unresolved >15 minutes must trigger alert.

## Observability
1. Every command state transition emits event with trace ID.
2. Twin reconciliation loop emits drift/convergence events.
3. Operator UI must show per-device latest command status.

## Implemented APIs (Current)
Through Web BFF:
1. `GET /api/devices/twins` -> list twin records.
2. `GET /api/devices/twins/{device_id}` -> get a single twin.
3. `PUT /api/devices/twins/{device_id}/desired` -> update desired state.
4. `PUT /api/devices/twins/{device_id}/reported` -> update reported state.
5. `GET /api/devices/{device_id}/commands` -> command history with lifecycle timestamps.
6. `PATCH /api/devices/{device_id}/commands/{command_id}/status` -> apply validated lifecycle transitions.

Core service routes:
1. `/api/v1/diagnostics/twins`
2. `/api/v1/diagnostics/twins/{device_id}`
3. `/api/v1/diagnostics/twins/{device_id}/desired`
4. `/api/v1/diagnostics/twins/{device_id}/reported`
5. `/api/v1/diagnostics/devices/{device_id}/commands/{command_id}/status`

Transition enforcement:
1. `created -> queued`
2. `queued -> delivered|failed`
3. `delivered -> acked|timed_out`
4. `acked -> applied|failed`

Audit behavior:
1. Every transition writes a diagnostics log with source `command-lifecycle`.
2. Transition metadata includes previous state, next state, command ID, device ID, and transition result/error.

Operator UI:
1. `apps/dashboard-web/src/pages/devices/DeviceTwins.tsx`
2. `apps/dashboard-web/src/pages/devices/RemoteCommands.tsx`

## Firmware Rollout Lifecycle (Current)
Supported rollout states:
1. `scheduled`
2. `in_progress`
3. `paused`
4. `completed`
5. `cancelled`

Allowed rollout transitions:
1. `scheduled -> in_progress|cancelled`
2. `in_progress -> paused|completed|cancelled`
3. `paused -> in_progress|cancelled`

Implemented APIs:
1. `GET /api/firmware/rollouts`
2. `POST /api/firmware/rollouts`
3. `PATCH /api/firmware/rollouts/{id}/status`

Validation and runtime coverage:
1. Unit tests: `services/fleet-core-service/internal/domains/diagnostics/service/rollout_transition_test.go`.
2. Handler tests: `services/fleet-core-service/internal/domains/diagnostics/handlers/rollout_handler_test.go`.
3. Runtime integration gate executes lifecycle path and invalid transition assertion in `scripts/runtime_integration_gate.sh`.
