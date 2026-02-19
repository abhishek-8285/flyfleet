# Cross-Platform Blueprint (Backend + Frontend + Mobile + Firmware)

## 1) Platform Scope and Boundaries
- Surfaces:
  - Web admin (`apps/dashboard-web`)
  - Mobile apps (driver/ops/customer)
  - Backend services (BFF + domain microservices)
  - Device gateway + firmware agents
- Critical user journeys:
  - Login and session continuity
  - Live fleet visibility and command dispatch
  - Work orders and maintenance completion
  - Firmware rollout and rollback
  - Alerts and incident response

## 2) Target Architecture (Reference)
- Client layer:
  - Web + Mobile use generated API clients from one OpenAPI source.
  - Shared auth/session SDK for token lifecycle, refresh, logout, revoke reaction.
- Edge layer:
  - Web BFF as policy and aggregation boundary.
  - Mobile API gateway profile (same contracts, mobile-specific pagination/compression).
- Domain services:
  - Auth, Fleet Core, Tracking, Alerts, Billing, Support, Reports, Settings, Integrations.
- Device plane:
  - Device gateway for MQTT/NATS command+telemetry ingress.
  - Device digital twin service for desired/reported state convergence.
- Data plane:
  - OLTP store (org data), time-series store (telemetry), object store (artifacts/log bundles/OTA binaries).

## 3) Canonical Contracts
### 3.1 API Contract Program
- Single source: `docs/api/openapi.yaml` (or versioned folder by domain).
- Rules:
  - No unversioned public endpoint additions.
  - Request/response schema changes are additive unless major version bump.
  - Error envelope standardized: `{ code, message, details, trace_id }`.
- Generation:
  - Web client package and mobile client package generated in CI.
  - PRs fail on stale generated clients.
- Deprecation:
  - `deprecated: true` + replacement endpoint + sunset date.

### 3.2 Event Contract Program
- Topics versioned: `fleet.telemetry.v1`, `fleet.command.ack.v1`, `fleet.ota.status.v1`.
- Schema registry strategy (JSON schema/protobuf) with backward compatibility checks.

## 4) Identity and Authorization End-to-End
### 4.1 Token Model
- Access token (short TTL), refresh token (rotating), device token (scoped + hardware-bound).
- Claims:
  - `sub`, `org_id`, `tenant_id`, `roles`, `permissions`, `session_id`, `device_id(optional)`.
- Session behaviors:
  - Immediate revoke support via cache-backed denylist and short access token TTL.

### 4.2 RBAC/ABAC
- Server is source of truth.
- UI gates mirror server permissions but never replace server checks.
- Policy matrix maintained in one versioned file (`docs/security/rbac-matrix.yaml`).

### 4.3 Security Controls
- Mandatory:
  - MFA for privileged operations.
  - CSRF for cookie session flows.
  - Rate limiting by org + IP + route sensitivity.
  - Audit events for auth changes, role edits, command dispatch, OTA actions.

## 5) Frontend/Web Standards
- Data layer:
  - Use generated clients, no raw hardcoded URLs.
  - Query/caching strategy standardized (stale time, retries, cancellation).
- UX consistency:
  - Shared error/loading/empty states.
  - Operation state feedback for asynchronous jobs (queued/running/completed/failed).
- Accessibility:
  - WCAG AA targets for contrast and keyboard navigation on core flows.

## 6) Mobile Architecture (Operational Grade)
### 6.1 Offline-first Sync
- Local store with op-log queue:
  - Each mutation has `op_id`, timestamp, retry_count, dedupe key.
- Retry policy:
  - exponential backoff + jitter, max retry, dead-letter bucket.
- Conflict resolution:
  - Server-wins by default; per-domain override where needed.
  - Conflict UI for manual resolution in non-trivial entities.

### 6.2 Mobile Runtime Hardening
- Secure storage for tokens/secrets.
- Background sync workers with power/network-aware policies.
- Crash monitoring + release channel gating.
- Feature flags and remote kill-switch for risky modules.

## 7) Device Firmware Lifecycle
### 7.1 OTA Security
- Firmware artifacts signed (vendor key), signature verified on-device before install.
- Manifest includes:
  - firmware version, compatible hardware, minimum bootloader, hash, signature.

### 7.2 Rollout Strategy
- Stages:
  - internal lab -> canary cohort -> 5% -> 25% -> 100%.
- Auto-stop conditions:
  - failure rate threshold, boot-loop detection, telemetry silence threshold.
- Rollback:
  - previous known-good slot retained; forced rollback command available.

### 7.3 Observability and Audit
- Per-device OTA timeline with statuses and failure reason taxonomy.
- Immutable audit entries for who approved/triggered rollout and policy used.

## 8) Device Digital Twin + Command Reliability
### 8.1 Twin Model
- `desired_state`, `reported_state`, `last_seen`, `drift_status`.
- Reconciliation loop marks convergence/drift with reason codes.

### 8.2 Command Protocol
- Command IDs are idempotent and globally unique.
- States: `created -> queued -> delivered -> acked -> applied | failed | timed_out`.
- SLA metrics per command type and device class.

### 8.3 Operator Experience
- Web/mobile show real-time command state and remediation action.
- Failed commands link to device diagnostics snapshot.

## 9) Full-Stack Observability
- Trace propagation:
  - `x-trace-id` generated at client, propagated through BFF/services/gateway.
- Logging:
  - structured logs with org/device/session correlation keys.
- Metrics:
  - RED + USE metrics for services and gateway.
- Alerts:
  - auth failure spikes, telemetry lag, command timeout surge, OTA failure surge.

## 10) Testing and Quality Gates
### 10.1 Required Pipelines
- Contract tests:
  - OpenAPI drift, event schema compatibility.
- Integration tests:
  - BFF rewrites, auth enforcement, service fanout behavior.
- UI tests:
  - web E2E critical paths.
- Mobile tests:
  - offline/online transition, retry/conflict behavior.
- Device tests:
  - simulator-based command and OTA lifecycle tests.

### 10.2 Merge Gates
- PR blocked if:
  - contract drift, failing critical path tests, security policy regression, or coverage drop below thresholds.

## 11) Environment and Release Management
- Environments:
  - dev, staging, prod with parity checklist and drift detector.
- Release trains:
  - web (continuous), mobile (scheduled channels), firmware (ring-based rollout windows).
- Rollback playbooks:
  - one-click per surface with precomputed blast radius checks.

## 12) SLOs and Operations
- SLOs (examples):
  - Login success >= 99.9%
  - Fleet telemetry freshness P95 <= 10s
  - Command ack P95 <= 5s
  - OTA success >= 98% per batch
- Incident ops:
  - runbooks by alert type, pager ownership matrix, postmortem template.

## 13) Governance and Ownership
- Domain ownership file:
  - service owner, code owner, on-call rotation, escalation path.
- Change governance:
  - architectural decision records (ADR) required for cross-domain changes.

## 14) 90-Day Execution Plan
### Days 1-30
- Lock API/event contract governance and CI drift gates.
- Introduce trace-id propagation in web + BFF + key services.
- Draft mobile sync queue RFC and implement base queue.

### Days 31-60
- Implement firmware signing + manifest verification end-to-end in staging.
- Implement command lifecycle state machine and UI exposure.
- Add mobile retry/conflict instrumentation.

### Days 61-90
- Run canary OTA with automatic stop/rollback policy.
- Enforce critical-path full matrix in CI nightly + pre-release.
- Finalize SLO dashboards and incident response runbooks.

## 15) Repo-Level Immediate Actions
1. Add `docs/SERVICE_OWNERSHIP.md` with source-of-truth route ownership.
2. Add OpenAPI/event schema CI workflows with required status checks.
3. Add mobile sync design doc under `docs/architecture/mobile-sync.md`.
4. Add OTA manifest schema and verification tooling under firmware/devops paths.
5. Add trace-id middleware to frontend network layer and BFF proxy path.
