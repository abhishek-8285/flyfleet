# Fleet-Core Domain Architecture

## Overview

`fleet-core` is a **Go modular monolith** — a single binary containing all
business logic as separate domain packages under `internal/domains/`.
Each domain follows a hexagonal/clean architecture with:
`models/`, `repository/`, `service/`, `handlers/` subdirectories.

The only binary is in `cmd/main.go`, which wires every active domain together
via Chi router, Prometheus metrics, Redis, PostgreSQL, and NATS.

---

## Active Domains (wired into main.go)

These domains are imported by `main.go` or by other active domains and are
part of the production system:

| Domain | Business Capability | Key Sub-packages |
|--------|-------------------|------------------|
| **dispatch** | Trip orchestration — assign shipments to vehicles/drivers, score suggestions, auto-start/complete trips from telemetry | `engine`, `service`, `repository`, `handlers`, `worker` |
| **trip** | Full trip lifecycle state machine (Create → Start → Complete/Cancel), trip offers, events, analytics | `models`, `repository`, `service`, `handlers` |
| **vehicle** | Fleet vehicle management — registration, status, device crypto auth | `models`, `repository`, `service`, `handlers` |
| **driver** | Driver profiles, availability, assignments, mobile app support | `models`, `repository`, `service`, `handlers` |
| **customer** | Customer accounts, bookings, quotes, shipments, pricing, credit | `models`, `repository`, `service`, `handlers` |
| **iam** | Identity & access — tenants, users, roles, RBAC, API keys | `models`, `repository`, `service`, `handlers` |
| **billing** | Quote generation, billing engine, payment processing, PDF generation | `repository`, `service`, `handlers`, `payments`, `pdf` |
| **invoicing** | Invoice creation and management | `repository`, `service`, `handlers` |
| **route** | Route optimization, planned routes, waypoints, OSRM integration | `models`, `repository`, `service`, `handlers` |
| **india** | India-specific compliance — e-way bills, GST, regional regulations | `models`, `repository`, `service`, `handlers` |
| **assignment** | Assignment management — vehicle/driver to trip assignments | `models`, `service`, `handlers` |
| **finance** | Financial operations — cost tracking, expense reconciliation | `models`, `repository`, `service`, `handlers` |
| **emergency** | Emergency response — panic alerts, nearest vehicle dispatch | `models`, `repository`, `service`, `handlers` |
| **fuel** | Fuel management — transaction tracking, efficiency analytics | `models`, `repository`, `service`, `handlers` |
| **integration** | Third-party connectors — ERP systems, API sync | `models`, `repository`, `service`, `handlers` |
| **admin_support** | Admin tooling — device replacement, org management | `domains/`, `handlers` |

## Tracking Infrastructure (top-level `internal/`)

| Package | Purpose |
|--------|---------|
| `internal/tracking/` | Telemetry ingestion, live cache (Redis), WebSocket broadcasting, geofencing, history queries |
| `internal/tracking/worker/` | Geofence checker worker |
| `internal/tracking/repository/` | Telemetry repository, geofence repository, live cache repository |
| `internal/tracking/service/` | TelemetryQueryService, IngestionService, BatchLocationService |

## Dead/Removal Domains

See `docs/dead-code-audit.md` for the full audit. The following domains were
**never imported** by any active code and were removed/cleanup candidates:

| Domain | Lines | Status |
|--------|-------|--------|
| `auth` | 5,405 | Removed — superseded by `iam` module |
| `auth_deprecated` | 0 | Removed — empty skeleton |
| `internal/domains/tracking` | 8,140 | Removed — active tracking is in `internal/tracking/` |
| `document` | 0 | **Empty** — needs decision (see below) |
| `alert` / `alerts` | 12,663 | Never imported — candidates for removal |
| `analytics` | 8,702 | Never imported — candidate for removal |
| `compliance` | 3,375 | Never imported — candidate for removal |
| `diagnostics` | 2,802 | Never imported — candidate for removal |
| `geofencing` | 2,774 | Never imported — active geofencing is in `internal/tracking/` |
| `governance` | 3,134 | Never imported — candidate for removal |
| `maintenance` | 7,426 | Never imported — candidate for removal |
| `platform_compat` | 1,372 | Never imported — candidate for removal |
| `reports` | 1,473 | Never imported — candidate for removal |
| `settings` | 4,264 | Never imported — candidate for removal |
| `video` | 6,011 | Never imported — candidate for removal |

## Why We Need `document` Domain (or Not)

The `document` domain is currently an **empty directory** — no `.go` files exist.
It was planned to house:

1. **Proof of Delivery (POD) documents** — signatures, photos captured at delivery
2. **Vehicle documents** — registration, insurance, permits with expiry tracking
3. **Driver documents** — licenses, certifications, training records
4. **Shipment documents** — e-way bills, invoices, delivery notes

**Why implement it:**
- Fleet operations require document lifecycle management (upload, versioning,
  expiry alerts, audit trail)
- Currently POD documents are stored as JSON blobs in trip records and
  customer documents live in the customer domain — a dedicated domain would
  centralize this

**Why leave it empty for now:**
- No active feature requests require it
- Existing document storage (trip POD JSON, customer e-way bills) is sufficient
- Can be added later if/when document management features are requested

**Decision needed:** Keep empty as a placeholder for future work, or remove
it alongside the other dead domains.

## Why We Need the Active Domains

Each active domain maps to a concrete business need:

- **Dispatch** is the core value proposition — matching shipments to vehicles/drivers
  using a scoring engine that considers proximity, driver ratings, and vehicle capacity.
- **Trip** is the execution unit — once dispatched, a trip follows a state machine
  (scheduled → started → completed/cancelled) with telemetry tracking.
- **Vehicle/Driver** are the operational assets — without them, dispatch has no
  pool to select from.
- **Customer** is the demand side — bookings, shipments, pricing, credit;
  dispatch only works if there are confirmed shipments to dispatch.
- **IAM** secures multi-tenant access — auth, RBAC, and tenant isolation are
  non-negotiable for a SaaS fleet platform.
- **Billing/Invoicing** monetizes completed trips — without this, the platform
  can't charge customers.
- **Route** optimizes vehicle paths — reduces fuel costs and improves delivery times.
- **Tracking** provides real-time visibility — telemetry ingestion, live cache,
  WebSocket streaming, geofencing for status automation.
