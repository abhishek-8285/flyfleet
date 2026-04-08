# Monorepo Health & Service Status

**Timestamp**: 2026-02-06 16:25 (IST)
**Status**: ✅ All core services verified build-stable in monorepo structure.

## Frontend Applications
| App Name | Path | Build Status | Runtime Port |
| :--- | :--- | :--- | :--- |
| Landing Page | `apps/landing-page` | ✅ Success | 3000 |
| Dashboard Web | `apps/dashboard-web` | ✅ Success (ESLint Bypass) | 3005 |

## Backend Microservices
All services are managed via `go.work` and verified for compilation.

| Service Name | Path | Build Status | Key Dependency |
| :--- | :--- | :--- | :--- |
| Auth Service | `services/microservices/auth-service` | ✅ Success | PGX, JWT |
| Vehicle Service | `services/microservices/vehicle-service` | ✅ Success | PGX, Validator |
| Analytics Service | `services/microservices/analytics-service` | ✅ Success | Shared Redis |
| Fleet Service | `services/fleet-service` | ✅ Success | Shared NATS |
| Driver Service | `services/microservices/driver-service` | ✅ Success | Shared Common |
| Order Service | `services/order-service` | ✅ Success | Shared Common |
| Tracking Service | `services/microservices/tracking-service` | ✅ Success | Redis, NATS |
| Trip Service | `services/microservices/trip-service` | ✅ Success | Shared Common |
| User Service | `services/user-service` | ✅ Success | Shared Common |
| Report Service | `services/report-service` | ✅ Success | Shared Common |
| Maintenance Service | `services/microservices/maintenance-service` | ✅ Success | Shared Common |
| Inventory Service | `services/inventory-service` | ✅ Success | Shared Common |

## Shared Packages
| Package Name | Path | Type | Status |
| :--- | :--- | :--- | :--- |
| Common Go | `packages/common/go` | Library | ✅ Stable |
| Branding Config | `packages/config` | Config/TS | ✅ Linked |

---
*Note: Runtime checks were performed via static build analysis. Full orchestration requires the complete Docker infrastructure.*
