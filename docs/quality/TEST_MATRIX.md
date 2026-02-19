# End-to-End Test Matrix

## Objective
Ensure platform reliability across backend, frontend, mobile, device gateway, and firmware release lifecycle.

## Test Layers
1. Contract tests (OpenAPI + event schemas)
2. Service integration tests
3. Web E2E tests
4. Mobile E2E and offline-sync tests
5. Device simulator tests
6. OTA simulation tests
7. Performance and load tests

## Matrix
| Layer | Scope | Trigger | Gate Type | Owner |
|---|---|---|---|---|
| API contract | `docs/api/openapi.yaml` + service contracts | PR + push | Required | API platform |
| BFF rewrite integration | `services/web-bff` route mappings | PR + push | Required | Backend |
| Auth security integration | auth + BFF + RBAC checks | PR + nightly | Required | Security/backend |
| Runtime lifecycle gate | cross-service boot + authenticated smoke + firmware rollout transitions | PR + nightly | Required | Platform/backend |
| Service integration matrix | per-service health + authenticated admin/report/governance API coverage on full stack | PR + nightly | Required | Platform/backend |
| Web critical E2E | login, dashboard, work-order core | PR + nightly | Required | Frontend |
| Mobile offline E2E | queue, retry, conflict | nightly + pre-release | Required (release) | Mobile |
| Device command simulator | command state lifecycle | nightly | Required (release) | Device platform |
| OTA simulation | ring rollout + rollback | PR + nightly + pre-release | Required (release) | Firmware/platform |
| Performance smoke | p95 latency + UI budget + queue latency | PR + nightly | Required | Platform |

## Minimum Required Journeys
1. Login -> org selection -> dashboard load.
2. Dispatch command -> device ack -> applied state reflected in UI.
3. Maintenance work order create/update/close.
4. Mobile offline mutation -> reconnect sync success.
5. OTA canary rollout -> success and rollback scenario.
6. Firmware rollout lifecycle transitions: `scheduled -> in_progress -> paused -> in_progress -> completed`, plus invalid transition rejection.
7. OTA orchestration decision simulation validates auto-pause/rollback thresholds from manifest rollout policy.
8. Settings snapshot lifecycle: `/api/v1/admin/settings/export` and `/api/v1/admin/settings/import` runtime smoke checks.

## Performance Gates (Initial)
1. API p95 latency for critical endpoints <= 300ms in staging profile.
2. Command ack p95 <= 5s (simulated fleet baseline).
3. Web initial JS payload gate <= 550KB gzipped for main route bundle (ratchet target: 300KB).
4. Mobile cold start <= 2.5s on reference device profile.

## Exit Criteria Per Release
1. No failing required gate.
2. No untriaged critical severity defects from test runs.
3. Contract and migration notes updated for externally visible changes.
4. Performance gates met or approved exception documented.

## Runtime Gate Toggle
1. `scripts/runtime_integration_gate.sh` runs by default.
2. Use `RUNTIME_GATE_DISABLED=1` only for controlled local bypass, never for CI required checks.

## Defect Handling
1. Failures classified by severity and ownership.
2. Critical regressions block release.
3. Non-critical regressions require explicit waiver with expiry date.
