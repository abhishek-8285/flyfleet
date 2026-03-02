# Monolith Consolidation Status

## Objective
Expose one backend entrypoint and one API inventory surface so API management is centralized while service internals are progressively merged.

## Implemented
- Single control-plane entrypoint in `fleet-core-service-legacy`.
- Swagger endpoints:
  - `/swagger`
  - `/swagger/openapi.json`
  - `/swagger/openapi.combined.json` (local + upstream mounted paths)
- Single-entry upstream proxy:
  - `/api/edge/{service}/*`
- Central control endpoints:
  - `/api/control/upstreams`
  - `/api/control/upstreams/health`
  - `/api/control/apis/routes`
  - `/api/control/apis/counts`
  - `/api/control/overview`
- Registry automation:
  - `make monolith-api-registry`
  - writes timestamped counts + combined OpenAPI snapshots into `tmp/`
- Governance checks:
  - `make monolith-overview`
  - `make monolith-governance-check`
  - `make backend-api-deep-scan` (full backend code scan with exclusions)
  - `make monolith-api-inventory` (monolith-only unique + duplicate METHOD/PATH inventory)
  - `make web-bff-rewrite-parity` (ensure web-bff routes.yaml rewrite prefixes are covered)
  - `make bff-public-route-parity` (ensure monolith auth-gate public allowlist parity)
  - `make monolith-compat-matrix` (full web-bff routes.yaml coverage matrix against monolith mappings)
  - `make monolith-compat-gate` (one-shot gate: rewrite parity + public-route parity + matrix + inventory)
  - `make monolith-status-bundle` (one-shot JSON snapshot: compat status + monolith inventory + backend totals)

## Migrated Domains (Local in Monolith)
- `settings-service` routes are now served directly by monolith at `/api/v1/settings/*`.
- `governance-service` routes are now served directly by monolith at `/api/v1/governance/*`.
- `integrations-service` routes are now served directly by monolith at:
  - `/api/v1/integrations/*`
  - `/api/v1/api-clients/*`
  - `/api/v1/webhooks/*`
- `reports-service` routes are now served directly by monolith at `/api/v1/reports/*`.
- `maintenance-service` routes are now served directly by monolith at:
  - `/api/v1/maintenance/*`
  - compatibility alias: `/api/v1/vehicles/{vehicle_id}/maintenance-history`
- `analytics-service` routes are now served directly by monolith at:
  - existing local paths from analytics domain
  - compatibility alias: `/api/v1/analytics/*`
- `trip-service` routes are now served directly by monolith with compatibility aliases:
  - `/api/v1/trips/settings`
  - `/api/v1/trips/{id}/telemetry`
  - `/api/v1/stats/{drivers,trips,summary}`
- `vehicle-service` routes are now served directly by monolith with compatibility aliases:
  - `/api/v1/bulk-jobs/*`
- `billing-service` routes are now served directly by monolith at `/api/v1/billing/*`.
- `admin-support-service` routes are now served directly by monolith for support, DR, organizations, and emergency APIs.
- `alert-service` routes are now served directly by monolith at:
  - `/api/v1/alerts/*`
  - `/api/v1/notifications/*`
- `tracking-service` routes are now served directly by monolith at:
  - `/api/v1/device/register`
  - `/api/v1/device/heartbeat`
  - `/api/v1/device/config`
  - `/api/v1/device/commands`
  - `/api/v1/telemetry/batch`
  - `/api/v1/device/telemetry/*`
  - `/api/v1/vehicles/{id}/history`
  - `/api/v1/telemetry/gps`
  - `/api/v1/telemetry/obd`
  - `/api/v1/analytics/vehicle/{id}`
  - `/api/v1/analytics/anomalies/vehicle/{id}`
  - `/api/v1/media/upload-url`
  - `/api/v1/media/confirm`
  - `/api/v1/forensics/attest`
  - `/api/v1/driver/login`
  - `/api/v1/driver/logout`
  - `/api/v1/driver/hos`
- `auth-service` APIs are now served directly by monolith at `/api/v1/auth/*`.
- `compliance-service` APIs are now served directly by monolith at `/api/v1/compliance/duty-status/*`.
- `india-integration-service` APIs are now served directly by monolith at:
  - `/api/v1/india/*`
  - compatibility aliases: `/api/fastag/*`, `/api/ewaybill/*`, `/api/vahan/*`, `/api/sarathi/*`, `/api/permits/*`, `/api/gst/*`, `/api/config/*`
- `web-bff` compatibility prefixes are now handled directly by monolith rewrite layer, including:
  - `/api/v1/admin/*` aliases for governance/DR/reports/notifications/billing/work-orders/inventory/vendors/inspections/fuel/finance/video/emergency/organizations/devices
  - legacy diagnostics aliases: `/api/devices/*`, `/api/firmware/*`, `/api/v1/devices/*`, `/api/v1/firmware/*`
  - additional aliases: `/api/search`, `/api/timeline`, `/api/v1/geofences`, `/api/v1/driver`, `/api/v1/tracking/telemetry/*`, `/api/v1/tracking/vehicles/*`
  - identity propagation parity: rewrite proxy derives `X-Organization-ID` and `X-User-ID` from bearer token claims (`org_id`, `user_id`/`sub`) when missing.
- `web-bff` auth edge behavior parity in monolith:
  - `/api/v1/auth/login` is fixed-window rate limited to `5/minute` per client IP
  - `/api/v1/auth/register` is fixed-window rate limited to `3/minute` per client IP
- Trace propagation parity in monolith:
  - global middleware sets and returns `X-Trace-ID` and `X-Request-ID` on every request
  - if inbound headers are missing, IDs are derived from chi request-id or generated locally
- Remaining utility/runtime services are exposed in monolith compatibility layer:
  - `fleet-ml-service` compatibility endpoints: `/api/v1/ml/health`, `/api/v1/ml/demand/predict`, `/api/v1/ml/pricing/suggest`, `/api/v1/ml/jobs/train`
  - `optimization-service` compatibility endpoints: `/api/v1/routes/{route_id}`, `/api/v1/optimize`
  - `ingestion-service` and `worker-service` health compatibility endpoint: `/api/v1/services/{name}/health`
  - `web-bff` compatibility endpoints: `/api/v1/ws`, `/api/v1/dashboard/stats`, `/api/v1/admin/dashboard/data`, `/api/v1/compliance/hos`, `/api/v1/telemetry`, `/api/v1/telemetry/dashcam`
- `settings`, `governance`, `india`, `integrations`, `reports`, `maintenance`, `analytics`, `trip`, `vehicle`, `billing`, `admin`, `alert`, `auth`, `compliance`, and `tracking` are excluded from upstream proxy registration.

## Upstream Environment Variables
- `AUTH_SERVICE_URL`
- `ALERT_SERVICE_URL`
- `ADMIN_SUPPORT_SERVICE_URL`
- `ANALYTICS_SERVICE_URL`
- `BILLING_SERVICE_URL`
- `COMPLIANCE_SERVICE_URL`
- `GOVERNANCE_SERVICE_URL`
- `INTEGRATIONS_SERVICE_URL`
- `MAINTENANCE_SERVICE_URL`
- `REPORTS_SERVICE_URL`
- `SETTINGS_SERVICE_URL`
- `TRACKING_SERVICE_URL`
- `TRIP_SERVICE_URL`
- `VEHICLE_SERVICE_URL`

## Operating Model
- Client traffic can be pointed to one service base URL.
- Endpoints already migrated are served locally.
- Endpoints not yet migrated are forwarded through `/api/edge/*`.
- API visibility and counts come from one control service.

## Next Migration Steps
1. Move one external service domain into `fleet-core-service-legacy` code.
2. Keep backward-compatible route paths in monolith.
3. Remove that domain from `/api/edge/*` proxy map.
4. Rebuild OpenAPI and verify `/api/control/apis/counts`.
5. Repeat until no upstream proxies remain.
