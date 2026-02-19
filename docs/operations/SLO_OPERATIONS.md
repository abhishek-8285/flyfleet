# SLO and Operations Specification

## Purpose
Define measurable reliability and performance targets across web, backend, mobile, gateway, and firmware operations.

## Measurement Principles
1. SLOs are user-journey oriented.
2. Metrics are measured in production and staging canary.
3. Error budgets are explicit and guide release risk decisions.

## SLO Catalog

### 1. Authentication Availability
- Indicator: successful login requests / total login attempts.
- Target: >= 99.9% over 30 days.
- Alert:
  - warning: < 99.95% (1h rolling)
  - critical: < 99.9% (30m rolling)

### 2. Dashboard Initial Load Performance
- Indicator: P95 time to usable dashboard view (web).
- Target: <= 2.0s on reference network profile.
- Alert:
  - warning: > 2.2s (1h)
  - critical: > 2.5s (30m)

### 3. API Latency (Critical Paths)
- Indicator: P95 latency for login, trips, work-orders, command endpoints.
- Target: <= 300ms (excluding network transit).
- Alert:
  - warning: > 350ms (30m)
  - critical: > 450ms (15m)

### 4. Telemetry Freshness
- Indicator: P95 lag between device event timestamp and UI availability.
- Target: <= 10s.
- Alert:
  - warning: > 12s (30m)
  - critical: > 20s (10m)

### 5. Command Reliability
- Indicator A: command delivery success rate.
- Target A: >= 99.5%.
- Indicator B: P95 command ack latency.
- Target B: <= 5s.
- Alert:
  - warning: success < 99.7% or ack > 6s
  - critical: success < 99.5% or ack > 8s

### 6. Mobile Sync Reliability
- Indicator: sync op apply success within retry budget.
- Target: >= 99.0% within 10 minutes.
- Alert:
  - warning: < 99.3% (1h)
  - critical: < 99.0% (30m)

### 7. OTA Success Rate
- Indicator: successful updates / attempted updates per batch.
- Target: >= 98.0%.
- Alert:
  - warning: < 98.5% (batch)
  - critical: < 98.0% (batch, auto-pause rollout)

## Error Budget Policy
1. If service consumes >50% monthly budget, require release risk review.
2. If service consumes >80% monthly budget, freeze non-critical releases.
3. Exceptions require engineering manager and on-call approver.

## Alert Routing
1. Auth and RBAC alerts -> security/backend primary on-call.
2. Telemetry and command alerts -> platform/device gateway on-call.
3. Web/mobile UX performance alerts -> frontend/mobile on-call.
4. OTA alerts -> firmware/platform release on-call.

## Incident Operating Procedure
1. Classify severity and assign incident commander.
2. Stabilize system (rollback/feature flag/rollout pause).
3. Communicate status cadence every 15 minutes for Sev1.
4. Produce post-incident review with tracked remediation items.

## Required Dashboards
1. User journeys dashboard (login, dispatch, maintenance, OTA).
2. API latency and error-rate dashboard by endpoint.
3. Mobile sync queue health dashboard.
4. OTA rollout ring and failure reasons dashboard.
5. Command lifecycle dashboard (state transition timing).

## Implemented Observability Artifacts
1. Prometheus SLO alert rules: `infrastructure/monitoring/slo-alert-rules.yml`
2. Grafana SLO journeys dashboard: `infrastructure/monitoring/slo-journeys-dashboard.json`
3. Foundation validator: `scripts/check_observability_foundation.py`
4. CI workflow gate: `.github/workflows/observability-foundation.yml`

## Review Cadence
1. Weekly SLO review in engineering operations sync.
2. Monthly error budget review for release governance.
3. Quarterly target recalibration based on growth and architecture changes.
