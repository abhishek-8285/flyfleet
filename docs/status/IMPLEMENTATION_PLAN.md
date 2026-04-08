# FlyFleet Implementation Plan (Sequential)

## Phase 1: Gateway Contract Safety (Done)
1. Add BFF contract checker from frontend to BFF routes.
2. Add BFF smoke script for critical admin modules.
3. Add root Makefile + npm scripts to run both checks.

## Phase 2: API Contract Alignment (In Progress)
1. Normalize frontend endpoint prefixes to one scheme (`/api/v1/...`). (In progress)
2. Remove mixed absolute calls (`http://localhost:8081/...`) from UI. (Done)
3. Add explicit rewrites in BFF for remaining naming mismatches. (Done for security/mfa/audit/integrations legacy paths)
4. Add CI/API hygiene checks to block hardcoded local URLs. (Done)

## Phase 3: Service Ownership Clarity
1. Define source-of-truth service per domain:
   - vehicles
   - trips
   - maintenance
   - diagnostics
   - governance
2. Remove duplicate route exposure paths where possible.

## Phase 4: CI Enforcement
1. Run `bff:contract-check` in CI on every PR.
2. Run `bff:smoke` in integration job with running stack.
3. Fail PR if route drift appears.

## Phase 5: Repo Hygiene
1. Clean tracked binaries/log artifacts.
2. Separate active docs from legacy docs.
3. Keep archive code excluded from active workflows.

## Phase 6: Cross-Platform Program
1. Execute `docs/CROSS_PLATFORM_BACKLOG.md` in order:
   - API contracts + generated clients
   - mobile offline/retry foundation
   - firmware OTA safety pipeline
   - E2E observability + test matrix
2. Promote each epic only after acceptance criteria are met.
3. Use `docs/CROSS_PLATFORM_BLUEPRINT.md` as detailed execution reference.
4. Treat `docs/README.md` as mandatory index for non-code onboarding.

## Phase 7: Performance Program (All Components)
1. Enforce SLO baselines from `docs/operations/SLO_OPERATIONS.md`.
2. Enforce release and rollback standards from `docs/release/RELEASE_STRATEGY.md`.
3. Enforce test gates from `docs/quality/TEST_MATRIX.md`.
4. Enforce docs completeness via `docs-governance` CI workflow.
5. Enforce performance CI gates via `.github/workflows/performance-gates.yml`:
   - web bundle budget
   - API latency smoke

## Phase 8: Traceability Program
1. Propagate `X-Trace-ID` from frontend (axios + fetch) to BFF and upstream proxies. (Done)
2. Enforce trace wiring checks in CI via `.github/workflows/trace-propagation.yml`. (Done)

## Phase 9: Mobile Sync Foundation
1. Define machine-readable mobile sync operation schema and policy. (Done)
2. Add deterministic retry/conflict simulator as baseline validation. (Done)
3. Enforce mobile sync foundation checks in CI via `.github/workflows/mobile-sync-foundation.yml`. (Done)

## Phase 10: Firmware OTA Foundation
1. Add OTA manifest validator and sample release artifact for pipeline validation. (Done)
2. Enforce OTA foundation checks in CI via `.github/workflows/firmware-ota-foundation.yml`. (Done)

## Phase 11: Device Reliability Foundation
1. Add digital twin and command lifecycle schemas. (Done)
2. Add command lifecycle simulator baseline validation. (Done)
3. Enforce reliability foundation checks in CI via `.github/workflows/device-reliability-foundation.yml`. (Done)

## Phase 12: Enterprise Foundation Gate
1. Aggregate non-runtime quality checks into one command (`enterprise-foundation-check`). (Done)
2. Enforce aggregate checks in CI via `.github/workflows/enterprise-foundation.yml`. (Done)

## Phase 13: Runtime Integration Gate
1. Add full stack runtime gate script (`scripts/runtime_integration_gate.sh`). (Done)
2. Add runtime workflow (`.github/workflows/runtime-integration.yml`) for nightly/manual execution. (Done)
3. Validate stack bring-up + smoke + latency checks with automatic cleanup. (Done)
