# FlyFleet Reality Map (2026-02-18)

This is a code-verified map of what is wired now vs what frontend expects.

## 1) Frontend -> API -> Backend Matrix

Assumption: `apps/dashboard-web` runs with CRA proxy to `http://127.0.0.1:8081`, so browser calls go to `web-bff` first.

| Frontend area | Endpoint pattern used by UI | Backend endpoint exists | Exposed by current `web-bff` | Status |
|---|---|---|---|---|
| Login | `/api/v1/auth/login` | Yes (`auth-service`) | Yes | Works |
| Auth profile/register/reset/verify | `/api/v1/auth/*` | Mostly yes (`auth-service`) | No (except login) | Blocked by BFF routing |
| User dashboard stats | `/api/v1/dashboard/stats` | Yes (`web-bff` handler) | Yes | Works |
| Monitoring basic | `/api/v1/monitoring/sla`, `/api/v1/monitoring/health` | Yes (`web-bff`) | Yes | Works |
| Admin org/users/vehicles/trips | `/api/v1/admin/*` (via `apiClient`) | Yes in `auth-service` for many admin routes | No | Blocked by BFF routing |
| Security pages | `/api/security/*` and `/api/v1/security/*` mixed | Yes in `auth-service` under `/api/v1/security/*` | No | Path mismatch + missing BFF routing |
| Governance UI | `/api/v1/admin/governance/*` | Yes in `admin-support-service` as `/api/v1/governance/*` | No | Prefix mismatch + missing BFF |
| DR UI | `/api/v1/admin/dr/*` incl `/restores`, `/tests` | Partial in `admin-support-service` as `/api/v1/dr/*` + `restore-jobs` | No | Contract mismatch |
| Reports UI | `/api/v1/admin/reports/*` | `reports-service` exposes `/api/v1/reports/*` | No | Prefix mismatch + missing BFF |
| Settings UI | `/api/v1/settings/*` (or `/settings/*` via client) | `settings-service` exposes `/settings/*` | No | Prefix mismatch + missing BFF |
| Notifications UI | `/api/v1/admin/notifications*` | `alert-service` exposes `/api/v1/notifications/*` | No | Prefix mismatch + missing BFF |
| Maintenance UI | `/api/v1/admin/work-orders|inventory|vendors|inspections*` | Similar routes exist in `fleet-core` but without `/admin` prefix | No | Prefix mismatch + missing BFF |
| Device diagnostics UI | `/api/devices/*`, `/api/firmware*` | `fleet-core` diagnostics under `/api/v1/diagnostics/*` | No | Prefix mismatch + missing BFF |
| Timeline/Search UI | `http://localhost:8081/api/timeline`, `/api/search*` | Handler files exist in BFF repo | Not mounted in BFF main router | Declared but not wired |
| Customer portal UI | `/api/v1/customer/*` | `fleet-core` has `/api/v1/customer/*` | No | Missing BFF proxy |
| Support console UI | `/bff/support/*` + `/api/support/*` | `admin-support-service` has `/api/v1/support/*` | No | Prefix mismatch + missing BFF |

## 2) Minimum Runnable Stack (today)

### What actually works end-to-end through BFF today

1. `POST /api/v1/auth/login`
2. `GET /api/v1/dashboard/stats` (with auth token)
3. `GET /api/v1/monitoring/sla` (with auth token)
4. `GET /api/v1/monitoring/health` (with auth token)

### Practical minimum containers

1. Required infra: `postgres`, `redis`, `nats`
2. Required services: `auth-service`, `web-bff`
3. Recommended for realistic dashboard stats: `fleet-core-service`, `alert-service`, `admin-support-service`, `billing-service`, `tracking-service`

Command:

```bash
docker compose up -d postgres redis nats auth-service web-bff fleet-core-service alert-service admin-support-service billing-service tracking-service
```

Quick checks:

```bash
curl -s http://localhost:8081/health
curl -s -X POST http://localhost:8081/api/v1/auth/login -H 'Content-Type: application/json' -d '{"email":"admin@example.com","password":"admin123"}'
```

## 3) Refactor Map (recommended order)

### Phase 1: Freeze API contract

Pick one external contract and enforce it:

1. Keep frontend contract as-is (`/api/v1/admin/...`, `/api/v1/...`) and adapt BFF.
2. Or align frontend to backend-native routes (more churn in UI).

Recommended: keep frontend contract stable, fix BFF/proxy layer.

### Phase 2: Make BFF the real gateway

In `services/web-bff/cmd/main.go`:

1. Mount all existing handler modules (security, support, settings, integrations, tracking, devices, timeline, search, websocket).
2. Add proxy routes for missing domains with path translation:
   - `/api/v1/admin/governance/*` -> `admin-support /api/v1/governance/*`
   - `/api/v1/admin/dr/*` -> `admin-support /api/v1/dr/*` (+ `restores <-> restore-jobs` mapping)
   - `/api/v1/admin/reports/*` -> `reports /api/v1/reports/*`
   - `/api/v1/admin/notifications/*` -> `alert /api/v1/notifications/*`
   - `/api/v1/admin/settings*` or `/api/v1/settings*` -> `settings-service /settings/*`
   - `/api/devices/*` -> `fleet-core /api/v1/diagnostics/*`

### Phase 3: Normalize prefixes + versioning

Standardize to one public scheme:

1. `/api/v1/auth/*`
2. `/api/v1/admin/*`
3. `/api/v1/customer/*`
4. `/api/v1/support/*`

And avoid mixed `/api/*`, `/bff/*`, hardcoded absolute `http://localhost:8081/*`.

### Phase 4: De-duplicate domain ownership

Current overlap exists between `fleet-core` and standalone services (`vehicle`, `trip`, `driver`, `maintenance`, `analytics`, `governance`).

Decide one:

1. Keep `fleet-core` as composition monolith and retire duplicates, or
2. Complete extraction into standalone services and slim `fleet-core`.

Until this is decided, endpoint ownership will keep drifting.

## 4) Key Files Used

1. `services/web-bff/cmd/main.go`
2. `apps/dashboard-web/src/services/apiClient.ts`
3. `apps/dashboard-web/src/services/api.ts`
4. `apps/dashboard-web/src/App.tsx`
5. `services/auth-service/cmd/main.go`
6. `services/fleet-core-service/cmd/main.go`
7. `services/admin-support-service/cmd/main.go`
8. `services/alert-service/cmd/main.go`
9. `services/reports-service/cmd/main.go`
10. `services/settings-service/cmd/main.go`
11. `services/integrations-service/cmd/main.go`
12. `services/trip-service/cmd/main.go`
13. `services/vehicle-service/cmd/main.go`
14. `services/driver-service/cmd/main.go`
15. `services/maintenance-service/cmd/main.go`
