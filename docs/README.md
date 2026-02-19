# FlyFleet Documentation Index

This index is the non-code source of truth for platform behavior and delivery standards.

## Start Here
1. `docs/CROSS_PLATFORM_BLUEPRINT.md`: Target architecture and operating model.
2. `docs/CROSS_PLATFORM_BACKLOG.md`: Ordered implementation program and acceptance criteria.
3. `docs/IMPLEMENTATION_PLAN.md`: Current phased execution status.
4. `docs/EXECUTION_STATUS.md`: Current implementation and enforcement status.

## Detailed Specs (No Code Dependency)
- `docs/api/API_GOVERNANCE.md`: OpenAPI governance rules, versioning, deprecation, and CI gates.
- `docs/api/ADMIN_SUPPORT_DR_API.md`: Disaster recovery admin API contract for backups, restore jobs, and DR tests.
- `docs/architecture/mobile-sync.md`: Mobile offline sync protocol, retries, and conflict handling.
- `docs/architecture/identity-auth-contract.md`: Token/session claims and RBAC scoping contract across web/mobile/backend.
- `docs/architecture/design-system-contract.md`: Shared token/state contract for web and mobile UI consistency.
- `docs/firmware/OTA_SPEC.md`: OTA signing, rollout stages, and rollback behavior.
- `docs/firmware/ota-manifest.schema.json`: Canonical OTA manifest schema.
- `firmware/keys/trusted_keys.json`: Trusted OTA signing key registry (`key_id` -> algorithm/public key path).
- `firmware/releases/sample/ota-rollout-scenarios.json`: Baseline canary/rollback orchestration scenarios.
- `docs/device/DEVICE_RELIABILITY_SPEC.md`: Digital twin and command lifecycle reliability spec.
- `docs/device/device_twin.schema.json`: Digital twin schema.
- `docs/device/command_lifecycle.schema.json`: Command lifecycle event schema.
- `docs/SERVICE_OWNERSHIP.md`: Domain ownership and API source-of-truth mapping.
- `docs/quality/TEST_MATRIX.md`: End-to-end test matrix across web/mobile/backend/device/firmware.
- `docs/operations/SLO_OPERATIONS.md`: SLO catalog, alert thresholds, and incident process.
- `infrastructure/monitoring/slo-alert-rules.yml`: Prometheus SLO alert definitions aligned with operations policy.
- `infrastructure/monitoring/slo-journeys-dashboard.json`: Grafana dashboard for journey-level SLO visibility.
- `scripts/runtime_integration_gate.sh`: Full-stack runtime gate including service integration matrix checks (`RUNTIME_GATE_DISABLED=1` for explicit local-only bypass).
- `docs/release/RELEASE_STRATEGY.md`: Environment parity, release trains, and rollback policy.
- `docs/EXECUTION_STATUS.md`: Current implementation and enforcement status.

## Governance Rule
If behavior in code and docs diverges, treat docs as stale and update docs in the same PR as code changes.
