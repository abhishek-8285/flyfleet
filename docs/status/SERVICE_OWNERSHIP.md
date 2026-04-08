# Service Ownership and Source-of-Truth Map

## Purpose
Defines authoritative service ownership for business domains and API prefixes so product and platform changes do not duplicate logic across services.

## Domain Ownership
| Domain | Source-of-Truth Service | Primary API Prefix | Notes |
|---|---|---|---|
| Authentication, MFA, sessions, RBAC | `auth-service` | `/api/v1/auth`, `/api/v1/security`, `/api/v1/admin/rbac` | BFF may expose legacy aliases during migration windows. |
| Vehicles, trips, maintenance, diagnostics | `fleet-core-service` | `/api/v1/admin/*`, `/api/v1/trips`, `/api/v1/diagnostics/*` | Device legacy routes should converge to versioned prefixes. |
| Alerts and notifications | `alert-service` | `/api/v1/notifications`, `/api/v1/alerts` | Notification preferences owned here. |
| Billing and invoices | `billing-service` | `/api/v1/billing*` | Metrics and settings remain in billing domain. |
| Support and DR workflows | `admin-support-service` | `/api/v1/support`, `/api/v1/dr` | BFF may expose `/bff/support` compatibility routes. |
| Reports and schedules | `reports-service` | `/api/v1/reports` | Includes report execution and schedules. |
| Settings and platform config | `settings-service` | `/settings`, `/api/v1/settings` | Versioning migration to `/api/v1/settings` recommended. |
| Integrations, API clients, webhooks | `integrations-service` | `/api/v1/integrations`, `/api/v1/api-clients`, `/api/v1/webhooks` | Legacy aliases must map and be documented. |
| Telemetry ingestion and live tracking | `tracking-service` + `ingestion-service` | internal ingest + `/api/v1/tracking*` | Ownership split: ingest pipeline vs query APIs. |

## BFF Responsibilities
1. Authentication enforcement and route-level policy enforcement.
2. Backward-compatible route mapping and response normalization.
3. Aggregation endpoints only when cross-service composition is needed.
4. No permanent business logic duplication from domain services.

## Ownership Rules
1. New domain logic is implemented only in source-of-truth service.
2. BFF aliases require deprecation plan and sunset date.
3. API docs must reflect both canonical and temporary compatibility paths.
4. On-call owner and code owner must be defined per service in team operations docs.

## Change Control
1. Cross-domain route changes require ADR.
2. API prefix reassignment requires migration plan and backward compatibility strategy.
3. Contract tests must be updated in same PR.
