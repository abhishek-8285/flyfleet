# fleet-core-service

## Purpose
fleet-core-service runtime component in FlyFleet platform.

## Run
Use the root docker compose stack or run the service binary from cmd/main.go.

## Health
- Endpoint: /health.

## Monolith Control Endpoints
- Swagger UI: `/swagger`
- OpenAPI JSON: `/swagger/openapi.json`
- Combined OpenAPI JSON (local + upstream proxy mounts): `/swagger/openapi.combined.json`
- Upstream map: `/api/control/upstreams`
- Upstream health: `/api/control/upstreams/health`
- Migration status: `/api/control/migrations`
- Local routes: `/api/control/apis/routes`
- API counts (local + upstream OpenAPI): `/api/control/apis/counts`
- Combined API duplicate/uniqueness report: `/api/control/apis/combined-report`
- One-call control summary: `/api/control/overview`
- Offline monolith inventory report: `make monolith-api-inventory` (writes summary + CSV files in `tmp/`)

## Single Entry Proxy Mode
This service can front non-migrated services so clients use one base URL.

- Proxy pattern: `/api/edge/{service-name}/*`
- Example: `/api/edge/billing/...` forwards to `BILLING_SERVICE_URL`
- Only HTTP/HTTPS upstreams are enabled.
- Migrated domains are served locally and removed from proxy map (currently `settings`, `governance`, `integrations`, `reports`, `maintenance`, `analytics`, `trip`, `vehicle`, `billing`, `admin`, `alert`, `auth`, `compliance`, `tracking`).
- India integration APIs are served locally under `/api/v1/india/*` with legacy compatibility aliases under `/api/*` for Fastag/EWayBill/Vahan/Sarathi/GST/Permits/Config flows.
- Web-BFF compatibility rewrites are served in monolith for major `/api/v1/admin/*` and diagnostics legacy prefixes (`/api/devices/*`, `/api/firmware/*`, `/api/v1/devices/*`, `/api/v1/firmware/*`).
- Platform compatibility endpoints are served in monolith for remaining runtime services:
  - ML: `/api/v1/ml/*`
  - Optimization: `/api/v1/optimize`, `/api/v1/routes/{route_id}`
  - Utility service health: `/api/v1/services/{name}/health`
- Optional BFF auth-gate parity:
  - set `ENFORCE_BFF_AUTH_GATE=true` to require bearer tokens on `/api/*` except public allowlist (`/health`, `/swagger*`, `/api/control/*`, `/api/v1/auth/{login,register,validate-token}`, `/api/v1/ws`).

Supported upstream env vars:
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

## Ownership
See docs/SERVICE_OWNERSHIP.md for domain responsibility.
