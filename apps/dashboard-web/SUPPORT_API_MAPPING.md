# Support Page API Mapping

This tracks support pages against backend endpoints and marks API gaps so we can implement them later.

## Implemented / Wired in Frontend
- `src/pages/support/SupportDashboard.tsx`
  - `GET /bff/support/dashboard/metrics`
  - `GET /bff/support/dashboard/trends`
  - `GET /bff/support/dashboard/top-errors`
  - `GET /bff/support/tickets/recent`
  - `GET /bff/support/dashboard/live`
- `src/pages/support/IncidentManagement.tsx`
  - `GET /bff/support/incidents/metrics`
  - `GET /bff/support/incidents/trends`
  - `GET /bff/support/incidents/feed`
  - `GET /bff/support/incidents`
- `src/pages/support/IncidentDetail.tsx`
  - `GET /bff/support/incidents/detail?id={id}`
  - `POST /api/support/incidents/comment?id={id}`
  - `PATCH /api/support/incidents/update?id={id}`
- `src/pages/support/UserConsole.tsx`
  - `GET /bff/support/users/{id}/profile`
  - `GET /bff/support/users/{id}/security`
  - `GET /bff/support/users/{id}/sessions`
  - `GET /bff/support/users/{id}/logins`
  - `GET /bff/support/users/{id}/roles`
  - `GET /bff/support/users/{id}/permissions`
  - `GET/POST /bff/support/users/{id}/notes`
  - `POST /bff/support/users/{id}/reset-password|unlock|mfa-reset|force-logout`
  - `POST /bff/support/users/{id}/sessions/{sessionId}/revoke`
- `src/pages/support/TroubleshootingLogs.tsx`
  - `GET /bff/support/logs`
  - `GET /bff/support/logs/filters`
  - `GET /bff/support/logs/{traceId}`
- `src/pages/support/OrganizationHub.tsx`
  - `GET /bff/support/orgs/{id}/summary`
  - `GET /bff/support/orgs/{id}/devices/stats`
  - `GET /bff/support/orgs/{id}/integration/health`
  - `GET /bff/support/orgs/{id}/fleet/locations`
  - `GET /bff/support/orgs/{id}/diagnostics/recent`
  - `GET /bff/support/orgs/{id}/alerts`
  - `GET /bff/support/orgs/{id}/contacts`
  - `POST /bff/support/orgs/{id}/impersonate`
- `src/pages/support/SLAReports.tsx`
  - `GET /bff/monitoring/sla`
- `src/pages/support/OrgConsole.tsx`
  - `GET /bff/support/orgs/{id}/summary` (primary)
  - `GET /api/support/orgs/overview` (fallback)

## API Gaps / Needs Backend Contract
- Support route set under `/bff/support/*` and `/api/support/*`
  - Not currently represented in root OpenAPI artifacts.
  - Needs official schema coverage to avoid frontend/backend drift.

## Next Build-After-Backend
- Add OpenAPI definitions for all `/api/v1/support/*` endpoints consumed by support pages.
- Confirm service ownership for support endpoints (currently routed by web-bff prefix proxy).
- Add typed frontend client (`supportService`) once endpoint contracts are finalized.
