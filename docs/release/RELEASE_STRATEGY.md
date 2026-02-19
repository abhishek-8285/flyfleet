# Release Strategy (Web + Backend + Mobile + Firmware)

## Goals
1. Predictable delivery with controlled risk.
2. Fast rollback and clear blast-radius boundaries.
3. Environment parity to reduce late surprises.

## Environment Model
1. `dev`: feature integration and rapid iteration.
2. `staging`: production-like validation and contract/performance checks.
3. `prod`: customer traffic.

Parity requirements:
1. Same core service topology and auth model.
2. Same observability pipelines and alert rules.
3. Versioned configuration with diff review.

## Release Trains
### Web
- Cadence: continuous deployment after green required gates.
- Guardrails: E2E smoke + API contract checks.

### Backend
- Cadence: daily/weekly train depending on domain criticality.
- Guardrails: integration tests + canary deployment + SLO burn check.
- Runtime matrix gate: `scripts/runtime_integration_gate.sh` with service-level health matrix and authenticated admin/report/governance route checks.

### Mobile
- Cadence: scheduled channels (internal -> beta -> production store).
- Guardrails: offline sync tests + crash-free threshold + startup performance target.

### Firmware
- Cadence: ring-based windows only.
- Guardrails: signed manifest verification + OTA simulator + canary soak approval.
- Automation: `scripts/ota_rollout_orchestrator.py` must return `ota_rollout_orchestration_simulation=PASS`.

## Promotion Criteria
1. Required tests and quality gates pass.
2. Error budget impact acceptable.
3. Rollback plan validated and owner on-call confirmed.

## Rollback Criteria
Immediate rollback if:
1. Critical auth failures above threshold.
2. Telemetry freshness SLO breach sustained above threshold.
3. OTA failure/rollback rate above threshold.
4. Command timeout surge above threshold.

## Rollback Playbooks
1. Web: redeploy previous artifact and invalidate CDN cache.
2. Backend: rollback deployment version and restore compatible config.
3. Mobile: disable via feature flags and/or halt phased rollout.
4. Firmware: pause rollout and issue forced rollback command for affected cohort.

## Change Windows
1. High-risk changes (auth, OTA, command plane) deployed in staffed windows.
2. Freeze windows for peak operations periods.

## Change Documentation
Every release must include:
1. Change summary by component.
2. Contract changes and migration notes.
3. Known risks and mitigation.
4. Rollback owner and execution steps.
