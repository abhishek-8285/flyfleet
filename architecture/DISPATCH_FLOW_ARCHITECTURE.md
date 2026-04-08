# Dispatch Flow — System Architecture

**Document ID:** ARCH-DF-001  
**Version:** 1.0  
**Status:** Draft  
**Created:** April 8, 2026  
**Author:** Senior Architect  
**Parent Spec:** [Dispatch Flow PSPEC](../product-spec/DISPATCH_FLOW.md)

---

## 1. Executive Summary

### 1.1 Purpose

This document defines the **canonical architecture** for the Dispatch Flow feature within the FlyFleet modular monolith. The product spec identifies this as a **P0 Q3 delivery** that wires existing but unwired domains together and adds a dispatch suggestion engine, dispatch dashboard, and booking→trip auto-linking.

### 1.2 User Story

> As a dispatcher, I want to assign a vehicle + driver to a delivery route so that the delivery is tracked from start to completion.

### 1.3 Key Finding: Most Code Already Exists

A codebase audit reveals that **the trip domain, vehicle domain, driver domain, customer portal, and route optimization are all fully built** — but **NOT wired into `cmd/main.go`**. The dispatch flow is primarily **integration work**, not greenfield development.

| Domain | Code Status | Wired? | Effort |
|--------|------------|--------|--------|
| **trip/** | ✅ 22 endpoints, state machine, NATS events, financials, compliance, POD | ❌ No | Mount + wire |
| **vehicle/** | ✅ 8 endpoints, CRUD, search, stats, driver assignment | ❌ No | Mount + wire |
| **driver/** | ✅ 30+ endpoints, CRUD, behavior scoring, HOS, duty status, shifts, payroll | ❌ No | Mount + wire |
| **customer/** | ✅ 15 endpoints, bookings, shipments, stops, delivery proof | ❌ No | Mount + wire + **fix stub** |
| **route/** | ✅ VRP solver (Clarke-Wright, Heuristic, NN), time windows, capacity | ❌ No | Mount + wire |
| **tracking/** | ✅ GPS telemetry, live cache, WebSocket | ✅ Yes | None — already live |
| **geofencing/** | ✅ PostGIS detection, NATS subscriber | ✅ Yes | None — already live |
| **alerts/** | ✅ Multi-channel notifications | ❌ No | Mount + wire |

---

## 2. System Context

### 2.1 C4 Level 1 — System Context

```
┌──────────────────────────────────────────────────────────────────────┐
│                        External Actors                                │
│                                                                       │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────────┐   │
│  │ GPS Devices  │    │ Dispatcher   │    │ Customers (Portal)   │   │
│  │ (AIS-140)    │    │ (Dashboard)  │    │ (Bookings, Payments) │   │
│  └──────┬───────┘    └──────┬───────┘    └──────────────────────┘   │
│         │                   │                                        │
└─────────┼───────────────────┼────────────────────────────────────────┘
          │                   │
          ▼                   ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    FlyFleet Platform Boundary                         │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Fleet-Core Monolith (Go / Chi Router / Port 8084)              │ │
│  │                                                                  │ │
│  │  Domains: dispatch (NEW), trip, vehicle, driver, customer,      │ │
│  │           route, tracking, geofencing, alerts, iam, finance     │ │
│  │  Shared: middleware, platform, messaging, database              │ │
│  └────────────────────────┬────────────────────────────────────────┘ │
│                           │                                          │
│                           ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Data Layer: PostgreSQL 15 + TimescaleDB + PostGIS              │ │
│  │  Cache: Redis 7+                                                │ │
│  │  Message Bus: NATS JetStream                                    │ │
│  └─────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

### 2.2 Container-Level (C4 Level 2)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     dashboard-web (React SPA)                            │
│  Components: DispatchDashboard, TripDetail, VehicleSelector, MapView    │
└───────────────────────┬─────────────────────────────────────────────────┘
                        │ HTTPS
                        ▼
┌─────────────────────────────────────────────────────────────────────────┐
│              Fleet-Core Monolith (Go / Chi / Port 8084)                  │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │  Middleware Stack                                                  │  │
│  │  CORS → SecurityHeaders → RequestLogging → TraceID →             │  │
│  │  JWKSAuth → TenantContext → RBAC → RateLimit → Idempotency       │  │
│  └────────────────────────────┬──────────────────────────────────────┘  │
│                               │                                          │
│  ┌──────────────┐ ┌──────────┴──────────┐ ┌────────────────────────┐   │
│  │ iam/         │ │ dispatch/ (NEW)     │ │ trip/                  │   │
│  │ - JWT auth   │ │ - Dashboard agg     │ │ - 22 endpoints         │   │
│  │ - RBAC       │ │ - Suggest engine    │ │ - State machine        │   │
│  │ - Sessions   │ │ - Assign handler    │ │ - NATS events (5)      │   │
│  └──────────────┘ │ - Redispatch        │ │ - Financials, POD      │   │
│                   └──────────┬──────────┘ │ - Waypoints, compliance│   │
│                              │            └───────────┬────────────┘   │
│  ┌──────────────┐ ┌─────────┴──────────┐ ┌────────────┴────────────┐  │
│  │ vehicle/     │ │ driver/            │ │ customer/               │  │
│  │ - Vehicle    │ │ - Driver CRUD      │ │ - Bookings (REQUESTED→  │  │
│  │   CRUD       │ │ - Behavior scoring │ │   QUOTED→CONFIRMED)     │  │
│  │ - Search     │ │ - HOS/duty status  │ │ - Shipments (BOOKED→    │  │
│  │ - Stats      │ │ - Shifts, payroll  │ │   DELIVERED)            │  │
│  │ - Assign drv │ │ - Availability     │ │ - Delivery proof (POD)  │  │
│  └──────────────┘ └────────────────────┘ │ - ⚠️ Stub assignment   │  │
│                                           └───────────┬────────────┘  │
│  ┌──────────────┐ ┌────────────────────┐ ┌────────────┴────────────┐  │
│  │ route/       │ │ tracking/          │ │ finance/                │  │
│  │ - VRP solver │ │ - GPS ingest       │ │ - Trip profitability    │  │
│  │ - Clarke-    │ │ - Live location    │ │ - Toll, fuel, GST       │  │
│  │   Wright     │ │ - WebSocket push   │ │ - E-invoice, e-way bill │  │
│  │ - Time wnds  │ │ - Anomaly detect   │ │                         │  │
│  └──────────────┘ └────────────────────┘ └─────────────────────────┘  │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │  Shared Infrastructure (internal/platform/)                        │  │
│  │  - EventBus (NATS)    - Health Checks      - Metrics (Prometheus) │  │
│  │  - Observability      - Response Helpers    - IAM Platform Svc    │  │
│  └───────────────────────────────────────────────────────────────────┘  │
└───────────────────────────────┬─────────────────────────────────────────┘
                                │
          ┌─────────────────────┼─────────────────────┐
          ▼                     ▼                     ▼
┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐
│ PostgreSQL 15     │ │    Redis 7+       │ │  NATS JetStream   │
│ + TimescaleDB     │ │  - Live cache     │ │  - trip.completed  │
│ + PostGIS         │ │  - Session store  │ │  - trip.assigned   │
│                   │ │  - Pub/Sub        │ │  - trip.started    │
│ Tables:           │ │                   │ │  - shipment.assigned│
│ - trips           │ │                   │ │  - vehicle.available│
│ - trip_events     │ │                   │ │  - driver.available │
│ - trip_offers     │ │                   │ │  - fleet.telemetry  │
│ - vehicles        │ │                   │ │  - fleet.geofence   │
│ - drivers         │ │                   │ │                     │
│ - customer_v2_*   │ │                   │ │                     │
│ - route_opt_*     │ │                   │ │                     │
│ - geofences       │ │                   │ │                     │
│ - telemetry_points│ │                   │ │                     │
│ - dispatch_actions│ │                   │ │                     │
│ - dispatch_sugg.  │ │                   │ │                     │
└───────────────────┘ └───────────────────┘ └───────────────────┘
```

---

## 3. Domain Architecture

### 3.1 Existing Domain Details

#### 3.1.1 Trip Domain (`internal/domains/trip/`)

**Files:** 16 Go files, ~201KB

**Endpoints (22 total, RegisterRoutes in `trip_handler.go`):**

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| GET | `/trips/` | `ListTrips` | ✅ Built, ❌ Not wired |
| POST | `/trips/` | `CreateTrip` | ✅ Built, ❌ Not wired |
| GET | `/trips/stats` | `GetStats` | ✅ Built, ❌ Not wired |
| GET | `/trips/stats/drivers` | `GetDriverStats` | ✅ Built, ❌ Not wired |
| GET | `/trips/stats/trips` | `GetTripStats` | ✅ Built, ❌ Not wired |
| GET | `/trips/offers` | `GetDriverOffers` | ✅ Built, ❌ Not wired |
| POST | `/trips/offers/{id}/respond` | `RespondToOffer` | ✅ Built, ❌ Not wired |
| GET | `/trips/{id}/` | `GetTrip` | ✅ Built, ❌ Not wired |
| PUT | `/trips/{id}/` | `UpdateTrip` | ✅ Built, ❌ Not wired |
| DELETE | `/trips/{id}/` | `DeleteTrip` | ✅ Built, ❌ Not wired |
| POST | `/trips/{id}/approve` | `ApproveTrip` | ✅ Built, ❌ Not wired |
| POST | `/trips/{id}/start` | `StartTrip` | ✅ Built, ❌ Not wired |
| POST | `/trips/{id}/pause` | `PauseTrip` | ✅ Built, ❌ Not wired |
| POST | `/trips/{id}/resume` | `ResumeTrip` | ✅ Built, ❌ Not wired |
| POST | `/trips/{id}/complete` | `CompleteTrip` | ✅ Built, ❌ Not wired |
| POST | `/trips/{id}/cancel` | `CancelTrip` | ✅ Built, ❌ Not wired |
| POST | `/trips/{id}/fail` | `FailTrip` | ✅ Built, ❌ Not wired |
| POST | `/trips/{id}/events` | `AddEvent` | ✅ Built, ❌ Not wired |
| GET | `/trips/{id}/telemetry` | `GetTelemetry` | ✅ Built, ❌ Not wired |
| POST | `/trips/{id}/pod` | `SubmitPOD` | ✅ Built, ❌ Not wired |
| POST | `/trips/{id}/offer` | `OfferTrip` | ✅ Built, ❌ Not wired |
| GET | `/trips/{id}/optimize-route` | `OptimizeRoute` | ✅ Built, ❌ Not wired |

**State Machine (8 states, hash-chained immutable events):**

```
CREATED ──approve──→ SCHEDULED ──start──→ STARTED ──pause──→ PAUSED
   │                      │                  │                  │
   │ cancel               │ cancel           │ fail             │ resume
   ▼                      ▼                  ▼                  ▼
CANCELLED             CANCELLED          FAILED         ┌──→ RESUMED ──complete──→ COMPLETED
                                                        │         │
                                                        │         │ fail
                                                        │         ▼
                                                        └────── FAILED
```

**Business Rules (enforced in state machine):**
- **STARTED**: requires `permit_verified`, `puc_valid`, `insurance_valid`
- **COMPLETED**: requires `pod_count >= 1`, `end_odometer >= start_odometer`
- **CANCELLED**: requires reason (except system-initiated)

**NATS Events Published (5 types):**

| Subject | Trigger | Payload |
|---------|---------|---------|
| `trip.completed` | `Complete()` | `{trip_id, org_id, vehicle_id, driver_id, billable_distance, duration_minutes}` |
| `trip.pod_submitted` | `SubmitPOD()` | `ProofOfDelivery` struct |
| `trip.offered` | `OfferTrip()` | `TripOffer` struct |
| `trip.accepted` | `RespondToOffer()` | `{trip_id, driver_id}` |
| `trip.rejected` | `RespondToOffer()` | `{trip_id, driver_id, reason}` |

**Financial Engine (`financial_service.go`):**
- Toll cost (10% distance assumption), fuel cost, driver allowance, SLA penalties
- GST (CGST/SGST intrastate, IGST interstate), TDS, reverse charge
- Invoice generation with full GST breakdowns
- E-way bill validation (>₹50,000 interstate/intrastate threshold)

**Compliance (`TripCompliance` struct):**
- Permit type/validity, PUC, insurance, fitness certificates
- E-way bill, FASTag, AIS-140, HOS compliance
- State permit validation

**Multi-stop Support:** `PlannedRoute []Waypoint`, `Waypoints []Waypoint`, `ActualRoute []Waypoint`
- Waypoint types: pickup, dropoff, checkpoint
- Waypoint states: pending, arrived, departed, skipped
- Minimum 2 waypoints required

**Key Repository Queries:**
- `GetActiveTripsForVehicle` — non-terminal trips for a vehicle
- `GetActiveTripsForDriver` — non-terminal trips for a driver
- `CheckOverlappingTrips` — prevents double-booking vehicle/driver
- `GetLastCompletedTripForVehicle` — previous trip data (odometer baseline)

---

#### 3.1.2 Vehicle Domain (`internal/domains/vehicle/`)

**Files:** 32 Go files across handlers/, service/, repository/, models/, crypto/, processor/, middleware/

**Endpoints (8 total, RegisterRoutes in `vehicle_handlers.go`):**

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| GET | `/vehicles/` | `ListVehicles` | ✅ Built, ❌ Not wired |
| POST | `/vehicles/` | `CreateVehicle` | ✅ Built, ❌ Not wired |
| GET | `/vehicles/stats` | `GetStats` | ✅ Built, ❌ Not wired |
| GET | `/vehicles/search` | `SearchVehicles` | ✅ Built, ❌ Not wired |
| GET | `/vehicles/{id}/` | `GetVehicle` | ✅ Built, ❌ Not wired |
| PUT | `/vehicles/{id}/` | `UpdateVehicle` | ✅ Built, ❌ Not wired |
| DELETE | `/vehicles/{id}/` | `DeleteVehicle` | ✅ Built, ❌ Not wired |
| POST | `/vehicles/{id}/assign-driver` | `AssignDriver` | ✅ Built, ❌ Not wired |

**Vehicle Model (key fields for dispatch):**
- `ID`, `TenantID`, `RegistrationNumber`, `VehicleType` (TRUCK/BUS/VAN/CAR/TRAILER/etc.)
- `CapacityTonnage *float64`, `FuelType`, `FuelTankCapacity`
- `Status` (ACTIVE/MAINTENANCE/SUSPENDED/DECOMMISSIONED)
- `GPSDeviceID *string`, `ImmobilizerEnabled bool`, `AIS140Compliant bool`
- `Make`, `Model`, `Year` (convenience fields)
- `ChassisNumber`, `EngineNumber` (AES-256 encrypted)

**Additional Handlers:**
- `bulk_handlers.go` — Bulk vehicle import/export
- `compliance_handlers.go` — Permit, PUC, insurance, fitness tracking
- `documents_handlers.go` — Vehicle document management
- `lifecycle_handlers.go` — Registration, RC transfer, NOC
- `trip_handlers.go` — Vehicle-scoped trip views

**Service Layer:**
- `vehicle_service.go` — CRUD, validation, tenant isolation
- `trip_service.go` — Vehicle-trip association queries
- `bulk_service.go` — Bulk import processing
- `compliance_service.go` — Certificate expiry tracking
- `documents_service.go` — Document upload/verification

**Repository:**
- `vehicle_repository.go` — CRUD, search, stats
- `trip_repository.go` — Vehicle-scoped trip queries
- `bulk_repository.go` — Bulk import operations
- `compliance_repository.go` — Certificate queries
- `documents_repository.go` — Document storage
- `events_repository.go` — Vehicle event log
- `financial_repository.go` — Vehicle financials
- `maintenance_repository.go` — Maintenance records

---

#### 3.1.3 Driver Domain (`internal/domains/driver/`)

**Files:** 33 Go files across handlers/, service/, repository/, models/, middleware/, workers/

**Endpoints (30+ total, RegisterRoutes in `driver_handler.go`):**

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| GET/POST | `/drivers/` | List/Create | ✅ Built, ❌ Not wired |
| GET/PUT/DELETE | `/drivers/{id}/` | CRUD | ✅ Built, ❌ Not wired |
| POST | `/drivers/{id}/assign-vehicle` | Assign vehicle | ✅ Built, ❌ Not wired |
| POST | `/drivers/{id}/unassign-vehicle` | Unassign vehicle | ✅ Built, ❌ Not wired |
| POST | `/drivers/{id}/update-status` | Update status | ✅ Built, ❌ Not wired |
| GET | `/drivers/{id}/behavior-score` | Get behavior score | ✅ Built, ❌ Not wired |
| POST | `/drivers/{id}/update-behavior` | Update behavior score | ✅ Built, ❌ Not wired |
| GET | `/drivers/{id}/analytics` | Driver analytics | ✅ Built, ❌ Not wired |
| GET/POST | `/drivers/{id}/shifts/` | Shift management | ✅ Built, ❌ Not wired |
| GET | `/drivers/{id}/payroll/earnings` | Earnings | ✅ Built, ❌ Not wired |
| GET/POST | `/drivers/{id}/ratings/` | Rating management | ✅ Built, ❌ Not wired |
| GET/POST/DELETE | `/drivers/{id}/documents/` | Document management | ✅ Built, ❌ Not wired |
| GET/POST | `/drivers/{id}/attendance/` | Attendance tracking | ✅ Built, ❌ Not wired |
| GET/POST | `/drivers/{id}/schedule/` | Work schedule | ✅ Built, ❌ Not wired |
| POST | `/drivers/{id}/duty-status` | Update duty status | ✅ Built, ❌ Not wired |
| POST | `/drivers/{id}/sos` | Trigger SOS | ✅ Built, ❌ Not wired |
| POST | `/drivers/{id}/sync` | Mobile sync | ✅ Built, ❌ Not wired |
| GET/POST | `/drivers/{id}/messages/` | Driver messaging | ✅ Built, ❌ Not wired |
| GET/POST | `/drivers/{id}/expenses/` | Expense management | ✅ Built, ❌ Not wired |
| GET | `/drivers/{id}/performance` | Performance report | ✅ Built, ❌ Not wired |

**Driver Model (key fields for dispatch):**
- `ID`, `OrgID`, `FirstName`, `LastName`, `Email`, `Phone`
- `LicenseNumber`, `LicenseType` (LMV/HMV/Transport), `LicenseExpiry`
- `Status` (active/inactive/suspended/resigned)
- `CurrentVehicleID *string`, `CurrentTripID *string`
- `Rating float64` (0-5), `TotalTrips int`, `TotalDistanceKm`
- `BehaviorScore int` (0-100)
- `EmploymentType` (FULL_TIME/CONTRACT), `ExperienceYears`

**HOS (Hours of Service) — `DutyStatusLog`:**
- Status: `OFF_DUTY`, `SLEEPER_BERTH`, `DRIVING`, `ON_DUTY`, `REST`, `BREAK`
- Location tracking per status change
- Duration calculation

**Behavior Scoring (`behavior_score_service.go`):**
- Penalty constants: speeding (-5), harsh brake (-5), harsh accel (-3), phone usage (-15), seatbelt (-10)
- Score range: 0-100, adjustments -20 to +20 per event
- Safety, efficiency, compliance sub-scores

**Background Workers:**
- `document_expiry_worker.go` — Certificate expiry monitoring
- `fatigue_monitor_worker.go` — Driver fatigue detection

**Service Layer:**
- `driver_service.go` — CRUD, validation, tenant isolation
- `behavior_score_service.go` — Scoring calculations
- `duty_status_service.go` — HOS tracking, shift limits
- `shift_service.go` — Work schedule management
- `rating_service.go` — Performance ratings
- `payroll_service.go` — Earnings, payouts
- `training_service.go` — Training/certification tracking
- `safety_event_service.go` — Incident logging
- `onboarding_service.go` — Driver onboarding workflow
- `mobile_sync_service.go` — Mobile app sync

---

#### 3.1.4 Customer Domain (`internal/domains/customer/`)

**Files:** 26 Go files, ~79KB

**Endpoints (15 total, RegisterRoutes in `customer_handler.go`):**

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| POST/GET | `/customers/` | Create/Get | ✅ Built, ❌ Not wired |
| POST | `/bookings` | `CreateBooking` | ✅ Built, ❌ Not wired |
| GET | `/bookings/{id}` | `GetBooking` | ✅ Built, ❌ Not wired |
| POST | `/shipments` | `CreateShipment` | ✅ Built, ❌ Not wired |
| GET | `/shipments/{id}` | `GetShipment` | ✅ Built, ❌ Not wired |
| GET | `/shipments/{id}/tracking` | `GetShipmentTracking` | ✅ Built, ❌ Not wired |
| GET | `/shipments/search` | `SearchShipments` | ✅ Built, ❌ Not wired |
| GET | `/invoices` | `GetInvoices` | ⚠️ Stub |
| GET | `/analytics` | `GetAnalytics` | ⚠️ Stub (hardcoded) |
| GET | `/dashboard` | `GetDashboard` | ✅ Built |
| POST/GET | `/api/bookings` | API booking | ✅ Built, ❌ Not wired |
| GET | `/api/shipments` | API shipments | ✅ Built, ❌ Not wired |
| GET | `/api/tracking` | API tracking | ✅ Built, ❌ Not wired |

**Booking Model:**
- `Status`: `REQUESTED` → `QUOTED` → `CONFIRMED` → `REJECTED`
- `PickupLocation`, `DropLocation` (JSON blobs)
- `CargoType`, `Weight`, `Volume`, `VehicleType`
- `RequestedPickupTime`

**Shipment Model:**
- `Status`: `BOOKED` → `CONFIRMED` → `ASSIGNED` → `IN_TRANSIT` → `DELIVERED` → `CANCELLED`
- `TripID *uuid.UUID` — **the key link to dispatch**
- `BookingID *uuid.UUID`
- `Origin`, `Destination` (JSON blobs)
- `DistanceKm *float64`
- `EwayBillNumber`, `EwayBillExpiry` (India compliance)

**⚠️ CRITICAL STUB — `shipment_assignment_service.go`:**
```go
// Current implementation (STUB):
func (s *ShipmentAssignmentService) AssignShipment(...) {
    tripID := uuid.New() // MOCK — creates fake UUID
    // Vehicle/driver assignment is NOT implemented
    // eventBus.Publish is COMMENTED OUT
    // booking update is COMMENTED OUT
}
```

**This must be replaced with real trip creation via `tripService.CreateTrip()`.**

**Zero NATS Publishing:** All `eventBus.Publish` calls are commented out. The domain has no EventBus dependency injection.

**38 Model Types** covering: Customer accounts, users, sessions, API keys, bookings, shipments, stops, delivery proof, disputes, pricing, credit, SLA, audit logs.

---

#### 3.1.5 Route Domain (`internal/domains/route/`)

**Files:** 10 Go files, ~109KB

**Endpoints (14 total, RegisterRoutes in `route_handlers.go`):**

| Method | Path | Handler | Status |
|--------|------|---------|--------|
| POST/GET/PUT/DELETE | `/requests/` | Optimization request CRUD | ✅ Built, ❌ Not wired |
| GET | `/results` | List results | ✅ Built, ❌ Not wired |
| GET | `/results/{id}` | Get result | ✅ Built, ❌ Not wired |
| POST/GET | `/jobs/` | Async optimization jobs | ✅ Built, ❌ Not wired |
| POST | `/recalculate` | Real-time rerouting | ✅ Built, ❌ Not wired |

**Three Solvers:**

1. **ClarkeWrightSolver** (default) — Savings algorithm:
   - `S(i,j) = d(i,depot) + d(j,depot) - d(i,j)`
   - Multi-vehicle routes with capacity constraints
   - Merges stops greedily based on savings

2. **HeuristicSolver** — Grearest-neighbor with full constraint validation:
   - Multi-depot optimization
   - Time window respect
   - Vehicle class, capacity, road restrictions, state permits
   - Waiting time calculation

3. **NearestNeighbor** (legacy) — Greedy NN with road restrictions

**Request Model:**
- `Stops []Stop` with `Lat`, `Lon`, `TimeWindowStart/End`, `Demand`, `Priority`
- `Vehicles []VehicleProfile` with `VehicleClass`, `Capacity`, `CostPerKM`, `PermitTypes`
- `Depots []Depot` — start/end points
- `Constraints` — `MaxTravelTime`, `MaxDistance`, `MaxStopsPerRoute`, `MaxVehicles`
- `OptimizationGoals` — `MinimizeTravelTime`, `MinimizeDistance`, `MinimizeCost`, etc.

**Result Model:**
- `Routes []Route` — one per assigned vehicle
- Each `Route` has `VehicleID`, `Stops []RouteStop` with `Sequence`, `ArrivalTime`, `DepartureTime`
- `TotalDistance`, `TotalTime`, `TotalCost`, `TotalVehiclesUsed`

**⚠️ No Trip Converter:** Route results produce vehicle assignments and stop sequences, but **there is no code that converts results to Trip records.** This is a gap that the dispatch flow must fill.

**⚠️ No Database Migrations:** The route domain's tables (`route_optimization_requests`, `route_optimization_results`, `route_optimization_jobs`, `road_restrictions`, `vehicle_permits`) have **no migration files**. The domain is dead code until migrations exist.

---

### 3.2 New Dispatch Domain (`internal/domains/dispatch/`)

**What needs to be built:**

#### 3.2.1 Dispatch Suggestion Engine

```
internal/domains/dispatch/
├── engine/
│   └── suggestion_engine.go       # Proximity + availability + type matching
├── handlers/
│   ├── dispatch_handler.go        # Dashboard, suggest, assign, redispatch
│   └── bulk_dispatch_handler.go   # Batch assignment (P2)
├── service/
│   ├── dispatch_service.go        # Orchestrate assign/redispatch
│   ├── availability_service.go    # Vehicle + driver availability checks
│   └── scoring_service.go         # Score calculation (40/30/20/10)
├── repository/
│   └── dispatch_repository.go     # dispatch_actions, dispatch_suggestions tables
├── models/
│   ├── dispatch.go                # DispatchSuggestion, DispatchAction, DashboardResponse
│   └── scoring.go                 # Scoring weights, thresholds
├── worker/
│   └── auto_status_worker.go      # GPS-triggered status changes
└── metrics.go                     # Prometheus metrics
```

**Scoring Algorithm (40/30/20/10):**
```
score = proximityScore * 0.4 + fitScore * 0.3 + driverScore * 0.2 + 10.0

proximityScore = max(0, 100 - distKm * 5)    // Closer = better
fitScore       = 100.0                        // Default if no cargo info
driverScore    = driver.Rating * 20           // 0-5 → 0-100
```

#### 3.2.2 Dispatch Dashboard Endpoint

**`GET /api/v1/dispatch/dashboard?date=2026-04-07`**

Aggregates from multiple domains:
- **Summary counts** → Trip repository (COUNT by status, on-time rate)
- **Pending shipments** → Customer repository (CONFIRMED shipments without trip_id)
- **Active trips** → Trip repository (STARTED/PAUSED/RESUMED) + Tracking Redis (last GPS position)
- **Available vehicles** → Vehicle repository (ACTIVE, not on active trip)
- **Available drivers** → Driver repository (active, no current_trip_id)
- **Alerts** → Alerts repository + Trip repository (delayed trips)

**Implementation:** Single handler that queries 5+ domains, assembles response. Redis cache (5s TTL) for summary counts.

#### 3.2.3 Dispatch Assign Endpoint

**`POST /api/v1/dispatch/assign`**

Transaction flow:
1. Fetch shipment from customer domain
2. Validate vehicle availability (not on active trip)
3. Validate driver availability (no current trip, HOS within limits)
4. Call `tripService.CreateTrip()` — real trip creation
5. Link shipment to trip (`shipment.trip_id = trip.id`)
6. Update shipment status to `ASSIGNED`
7. Insert `dispatch_actions` audit record
8. Publish NATS events: `trip.assigned`, `shipment.assigned`
9. Update vehicle `current_driver_id`
10. Update driver `current_trip_id`

#### 3.2.4 Redispatch Endpoint

**`POST /api/v1/dispatch/redispatch`**

Flow:
1. Fetch failed trip
2. Cancel old trip (`tripService.CancelTrip()` or `FailTrip()`)
3. Find new vehicle+driver (reuse suggestion engine)
4. Create new trip
5. Re-link shipment to new trip
6. Publish `trip.reassigned`, `shipment.assigned` events
7. Audit log in `dispatch_actions`

---

## 4. Data Flow — End-to-End

### 4.1 Booking → Auto-Dispatch → Trip → Delivery

```
1. Customer creates BookingRequest
   POST /api/v1/customers/bookings
   → Status: REQUESTED

2. Dispatcher reviews and quotes
   PATCH /api/v1/customers/bookings/{id}/quote
   → Status: QUOTED

3. Customer confirms booking
   PATCH /api/v1/customers/bookings/{id}/confirm
   → Status: CONFIRMED
   → Shipment created (BOOKED → CONFIRMED)

4. AUTO-DISPATCH TRIGGERS
   ┌─────────────────────────────────────────────────┐
   │ Dispatch Suggestion Engine                       │
   │ 1. Get vehicle last GPS from Redis live cache    │
   │ 2. Find available vehicles (not on active trip)  │
   │ 3. Filter by vehicle type + capacity match       │
   │ 4. Find available drivers (no current trip)      │
   │ 5. Check driver HOS (not over shift limit)       │
   │ 6. Score each match (40% proximity, 30% fit,    │
   │    20% driver rating, 10% base availability)     │
   │ 7. Return top 3 suggestions                      │
   └─────────────────────────────────────────────────┘

5. Dispatcher reviews suggestions
   GET /api/v1/dispatch/suggest?shipment_id={id}
   → Returns: [{vehicle, driver, score, reason}, ...]

6. Dispatcher accepts suggestion (or overrides)
   POST /api/v1/dispatch/assign
   → Transaction:
      a. tripService.CreateTrip()
      b. shipment.trip_id = trip.id
      c. shipment.status = ASSIGNED
      d. dispatch_actions INSERT (audit)
      e. NATS: trip.assigned, shipment.assigned
   → Returns: {trip_id, trip_number, status}

7. Trip starts (auto-detect via GPS telemetry)
   Vehicle ignition ON near pickup → tripService.Start()
   → Trip status: STARTED
   → Shipment status: IN_TRANSIT
   → NATS: trip.started
   → WebSocket push to dispatch dashboard

8. Trip completes (auto-detect via GPS + geofence)
   Vehicle stops at drop location → tripService.Complete()
   → Trip status: COMPLETED
   → Shipment status: DELIVERED
   → POD captured (GPS timestamp + location)
   → NATS: trip.completed
   → Delivery receipt PDF generated
   → Invoice created (finance domain)
```

### 4.2 Dispatch Dashboard Aggregate Flow

```
1. Dispatcher opens dashboard
   GET /api/v1/dispatch/dashboard?date=2026-04-07

2. Dispatch handler queries in parallel:
   a. TripRepository.GetTripSummary(orgID, date)     → counts by status
   b. CustomerRepository.GetPendingShipments(orgID)   → CONFIRMED, no trip_id
   c. TripRepository.List(orgID, status=active)       → in-transit trips
   d. VehicleRepository.GetAvailableVehicles(orgID)   → ACTIVE, no active trip
   e. DriverRepository.List(orgID, status=active)     → active, no current_trip
   f. TripRepository.GetDelayedTrips(orgID)           → behind schedule

3. Redis cache check for summary counts (5s TTL)
   → If hit, skip DB queries for summary

4. Assemble response:
   {
     summary: {total, pending, in_transit, delivered, failed, on_time_rate},
     pending_shipments: [...],
     active_trips: [{id, vehicle, driver, eta, is_delayed, current_location}],
     available_vehicles: [...],
     available_drivers: [...],
     alerts: [{type, trip_id, message, severity}]
   }

5. Return JSON (target: <1s P95)
```

### 4.3 Auto Status Detection Flow

```
1. GPS device sends position → POST /api/v1/telemetry
2. IngestionService writes to DB + Redis live cache
3. Auto Status Worker (background goroutine) runs every 5s:
   a. Fetch all STARTED trips from trip repository
   b. For each trip, check vehicle's last GPS position
   c. If vehicle near pickup geofence AND ignition ON:
      → tripService.StartTrip() (if not already STARTED)
   d. If vehicle near drop geofence AND speed = 0 for >2min:
      → tripService.CompleteTrip() (with POD from GPS)
4. Status change → NATS event → WebSocket push to dashboard
```

---

## 5. API Specifications

### 5.1 GET /api/v1/dispatch/dashboard

**Purpose:** Aggregate dispatch dashboard view

**Request:**
```http
GET /api/v1/dispatch/dashboard?date=2026-04-07
Authorization: Bearer <JWT via JWKS>
```

**Query Parameters:**
| Param | Type | Description | Default |
|-------|------|-------------|---------|
| `date` | YYYY-MM-DD | Filter date | Today |
| `include_available` | bool | Include available vehicles/drivers | `true` |

**Response (200 OK):**
```json
{
  "summary": {
    "total_today": 42,
    "pending": 8,
    "in_transit": 18,
    "delivered": 14,
    "failed": 1,
    "cancelled": 1,
    "on_time_rate": 0.92,
    "avg_delay_min": 8
  },
  "pending_shipments": [
    {
      "shipment_id": "shp-001",
      "booking_id": "bk-001",
      "customer_name": "ABC Trading",
      "pickup": "Delhi Warehouse",
      "drop": "Noida Sector 62",
      "cargo_type": "Electronics",
      "weight_kg": 250,
      "vehicle_type": "LCV",
      "suggested_vehicle": "DL01AB1234",
      "suggested_driver": "Rajesh Kumar",
      "match_score": 92,
      "distance_to_pickup_km": 3.2
    }
  ],
  "active_trips": [
    {
      "id": "trip-001",
      "trip_number": "TRIP-2026-00123",
      "status": "in-transit",
      "vehicle_registration": "DL01AB1234",
      "driver_name": "Rajesh Kumar",
      "drop_name": "Customer: ABC Trading",
      "expected_time": "2026-04-07T10:30:00Z",
      "eta_minutes": 25,
      "is_delayed": false,
      "current_location": {"lat": 28.5706, "lng": 77.2620}
    }
  ],
  "available_vehicles": [
    {"id": "v-789", "registration": "DL01XY9876", "type": "Tata Ace Gold",
     "capacity_tonnage": 1.5, "status": "available"}
  ],
  "available_drivers": [
    {"id": "d-456", "name": "Amit Singh", "rating": 4.5, "total_trips": 342, "status": "available"}
  ],
  "alerts": [
    {"type": "delay_warning", "trip_id": "trip-002",
     "message": "TRIP-2026-00124 is 15 min behind schedule",
     "severity": "warning", "timestamp": "2026-04-07T10:45:00Z"}
  ]
}
```

### 5.2 POST /api/v1/dispatch/suggest

**Purpose:** Get ranked vehicle+driver suggestions for a shipment

**Request:**
```json
{
  "shipment_id": "shp-001",
  "pickup_latitude": 28.5355,
  "pickup_longitude": 77.2910,
  "vehicle_type": "LCV",
  "cargo_weight_kg": 250.0,
  "top_n": 3
}
```

**Response (200 OK):**
```json
{
  "suggestions": [
    {
      "rank": 1,
      "vehicle": {
        "id": "v-123", "registration": "DL01AB1234",
        "type": "Tata Ace Gold", "capacity_tonnage": 1.5
      },
      "driver": {"id": "d-456", "name": "Rajesh Kumar", "rating": 4.5, "total_trips": 342},
      "score": 92.5,
      "distance_to_pickup_km": 3.2,
      "eta_to_pickup_min": 12,
      "reasons": ["Closest available vehicle", "Perfect capacity match", "High driver rating"]
    }
  ]
}
```

### 5.3 POST /api/v1/dispatch/assign

**Purpose:** Create trip from shipment (the core dispatch action)

**Request:**
```json
{
  "shipment_id": "shp-001",
  "vehicle_id": "v-123",
  "driver_id": "d-456",
  "scheduled_start": "2026-04-07T09:00:00Z",
  "auto_start": true,
  "auto_complete": true
}
```

**Response (201 Created):**
```json
{
  "trip_id": "trip-001",
  "trip_number": "TRIP-2026-00123",
  "status": "scheduled",
  "shipment_status": "assigned",
  "message": "Trip created and linked to shipment"
}
```

### 5.4 POST /api/v1/dispatch/redispatch

**Purpose:** Reassign a failed trip to a new vehicle+driver

**Request:**
```json
{
  "trip_id": "trip-002",
  "new_vehicle_id": "v-789",
  "new_driver_id": "d-789",
  "reason": "Vehicle breakdown"
}
```

**Response (200 OK):**
```json
{
  "old_trip_id": "trip-002",
  "old_trip_status": "cancelled",
  "new_trip_id": "trip-003",
  "new_trip_number": "TRIP-2026-00124",
  "shipment_id": "shp-001",
  "message": "Trip reassigned successfully"
}
```

### 5.5 POST /api/v1/dispatch/bulk-assign (P2)

**Purpose:** Batch dispatch multiple shipments

**Request:**
```json
{
  "assignments": [
    {"shipment_id": "shp-001", "vehicle_id": "v-123", "driver_id": "d-456"},
    {"shipment_id": "shp-002", "vehicle_id": "v-456", "driver_id": "d-789"}
  ]
}
```

**Response (200 OK):**
```json
{
  "total": 2,
  "succeeded": 2,
  "failed": 0,
  "results": [
    {"shipment_id": "shp-001", "trip_id": "trip-001", "status": "created"},
    {"shipment_id": "shp-002", "trip_id": "trip-002", "status": "created"}
  ]
}
```

---

## 6. Data Model

### 6.1 New Tables

**`dispatch_actions`** — Audit trail for dispatch decisions
```sql
CREATE TABLE dispatch_actions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    action_type     VARCHAR(30) NOT NULL CHECK (action_type IN (
        'auto_suggest', 'manual_assign', 'auto_assign',
        'bulk_assign', 'redispatch', 'override'
    )),
    shipment_id     UUID REFERENCES customer_v2_shipments(id),
    trip_id         UUID REFERENCES trips(id),
    vehicle_id      UUID REFERENCES vehicles(id),
    driver_id       UUID REFERENCES drivers(id),
    suggestion_data JSONB,          -- Top 3 suggestions with scores
    decision_reason TEXT,
    performed_by    UUID REFERENCES users(id),
    performed_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_dispatch_actions_org ON dispatch_actions (org_id, performed_at DESC);
CREATE INDEX idx_dispatch_actions_shipment ON dispatch_actions (shipment_id);
CREATE INDEX idx_dispatch_actions_trip ON dispatch_actions (trip_id);
```

**`dispatch_suggestions`** — Cache for dashboard (short TTL)
```sql
CREATE TABLE dispatch_suggestions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    shipment_id     UUID NOT NULL REFERENCES customer_v2_shipments(id),
    suggestion_rank INT NOT NULL,     -- 1, 2, 3
    vehicle_id      UUID REFERENCES vehicles(id),
    driver_id       UUID REFERENCES drivers(id),
    score           DOUBLE PRECISION, -- 0-100
    distance_km     DOUBLE PRECISION,
    eta_min         INT,
    reasons         TEXT[],
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    expires_at      TIMESTAMPTZ DEFAULT NOW() + INTERVAL '5 minutes'
);

CREATE INDEX idx_dispatch_suggestions_shipment ON dispatch_suggestions (shipment_id, expires_at DESC);
```

### 6.2 Existing Tables Used (No Schema Changes)

| Table | Domain | Purpose |
|-------|--------|---------|
| `trips` | trip | Trip records, status, waypoints, financials |
| `trip_events` | trip | Hash-chained immutable event log |
| `trip_offers` | trip | Driver offer workflow |
| `proof_of_delivery` | trip | POD documents |
| `vehicles` | vehicle | Fleet assets |
| `drivers` | driver | Personnel |
| `customer_v2_shipments` | customer | Shipment records (with `trip_id` FK) |
| `customer_v2_booking_requests` | customer | Booking records |
| `customer_v2_shipment_stops` | customer | Multi-stop shipments |
| `customer_v2_delivery_proofs` | customer | Delivery proof (photo, signature) |
| `telemetry_points` | tracking | GPS telemetry (TimescaleDB) |
| `gps_locations` | tracking | Latest GPS positions |
| `geofences` | geofencing | Spatial zones |
| `geofence_events` | geofencing | Entry/exit/dwell events |
| `route_optimization_requests` | route | Optimization inputs |
| `route_optimization_results` | route | Optimization outputs |

### 6.3 NATS Events — New Dispatch Events

| Subject | Publisher | Consumer | Payload |
|---------|-----------|----------|---------|
| `trip.assigned` | DispatchService | Alerts, Analytics | `{trip_id, vehicle_id, driver_id, shipment_id, org_id}` |
| `trip.dispatched` | DispatchService | Tracking, Analytics | `{trip_id, scheduled_start, org_id}` |
| `trip.started` | TripService (on auto-detect) | Customer, Alerts | `{trip_id, vehicle_id, started_at, org_id}` |
| `trip.reassigned` | DispatchService | Alerts, Analytics | `{old_trip_id, new_trip_id, reason, org_id}` |
| `shipment.assigned` | DispatchService | Customer | `{shipment_id, trip_id, vehicle_id, driver_id}` |
| `shipment.in_transit` | TripService (on start) | Customer | `{shipment_id, trip_id, started_at}` |
| `shipment.delivered` | TripService (on complete) | Customer, Billing | `{shipment_id, trip_id, delivered_at, pod}` |
| `vehicle.available` | TripService (on complete) | Dispatch | `{vehicle_id, org_id, last_location}` |
| `driver.available` | TripService (on complete) | Dispatch | `{driver_id, org_id, total_trips_today}` |

### 6.4 Redis Key Additions

| Key Pattern | Type | TTL | Purpose |
|-------------|------|-----|---------|
| `dispatch:dashboard_summary:{org_id}` | Hash | 5s | Cached dashboard aggregate |
| `dispatch:suggestion:{shipment_id}` | JSON | 5min | Pre-computed suggestions |
| `vehicle:{id}:on_active_trip` | Boolean | Trip duration | Quick availability check |
| `driver:{id}:on_active_trip` | Boolean | Trip duration | Quick availability check |

---

## 7. Route Registration in `cmd/main.go`

### 7.1 Current State

```go
// Current main.go — ONLY these domains are wired:
// - IAM (74 endpoints under /api/v1/auth/*)
// - Tracking (telemetry, geofences, WebSocket under /api/*)
// - Health check
```

### 7.2 Target State

```go
// ALL domains wired into main.go:

// IAM (already wired)
authRouter := chi.NewRouter()
iamModule.RegisterRoutes(authRouter)
api.Route("/auth", authRouter)

// Tracking (already wired)
if cfg.Flags.VehicleMapEnabled {
    api.Post("/v1/telemetry", ingestHandler)
    api.Get("/v1/vehicles/locations", vehicleLocationsHandler)
    api.Get("/v1/geofences", geofenceHandler)
    api.Get("/ws/telemetry", wsHandler)
}

// NEW: Dispatch
dispatchRouter := chi.NewRouter()
dispatchHandler := dispatch.NewHandler(deps)
dispatchHandler.RegisterRoutes(dispatchRouter)
api.Route("/v1/dispatch", dispatchRouter)

// Vehicle
vehicleRouter := chi.NewRouter()
vehicleHandler := vehicle.NewHandler(deps)
vehicleHandler.RegisterRoutes(vehicleRouter)
api.Route("/v1/vehicles", vehicleRouter)

// Driver
driverRouter := chi.NewRouter()
driverHandler := driver.NewHandler(deps)
driverHandler.RegisterRoutes(driverRouter)
api.Route("/v1/drivers", driverRouter)

// Trip
tripRouter := chi.NewRouter()
tripHandler := trip.NewHandler(deps)
tripHandler.RegisterRoutes(tripRouter)
api.Route("/v1/trips", tripRouter)

// Customer
customerRouter := chi.NewRouter()
customerHandler := customer.NewHandler(deps)
customerHandler.RegisterRoutes(customerRouter)
api.Route("/v1/customers", customerRouter)

// Route (optimization)
routeRouter := chi.NewRouter()
routeHandler := route.NewHandler(deps)
routeHandler.RegisterRoutes(routeRouter)
api.Route("/v1/routes", routeRouter)

// Alerts
alertsRouter := chi.NewRouter()
alertsHandler := alerts.NewHandler(deps)
alertsHandler.RegisterRoutes(alertsRouter)
api.Route("/v1/alerts", alertsRouter)
```

**Middleware applied to all domain groups:**
- JWKSAuth gate (for authenticated routes)
- TenantContext (org_id extraction)
- RBAC (role-based access)
- RateLimit (per-user throttling)

---

## 8. Fix the Stub — Shipment Assignment

### 8.1 Current Stub

```go
// customer/service/shipment_assignment_service.go
func (s *ShipmentAssignmentService) AssignShipment(ctx, shipmentID, vehicleID, driverID) {
    tripID := uuid.New() // ← MOCK TRIP ID
    // eventBus.Publish commented out
    // booking update commented out
}
```

### 8.2 Fixed Implementation

```go
func (s *ShipmentAssignmentService) AssignShipment(
    ctx context.Context,
    shipmentID uuid.UUID,
    vehicleID uuid.UUID,
    driverID uuid.UUID,
) (*AssignmentResult, error) {
    // 1. Fetch shipment
    shipment, err := s.shipmentRepo.GetByID(ctx, shipmentID)

    // 2. Fetch booking details
    booking, err := s.bookingRepo.GetByID(ctx, shipment.BookingID)

    // 3. Create trip via trip service (REAL, NOT MOCK)
    trip, err := s.tripService.CreateTrip(ctx, &trip.CreateTripRequest{
        OrgID:       shipment.OrgID,
        VehicleID:   vehicleID,
        DriverID:    driverID,
        PickupGeo:   booking.PickupLocation,
        DropGeo:     booking.DropLocation,
        CargoDetails: trip.CargoDetails{
            Type:   booking.CargoType,
            Weight: booking.Weight,
        },
        ScheduledStart: booking.RequestedPickupTime,
    })

    // 4. Link shipment to trip (single transaction)
    tx, _ := s.db.Begin(ctx)
    defer tx.Rollback(ctx)

    err = s.shipmentRepo.UpdateTripID(ctx, tx, shipmentID, trip.ID)
    err = s.shipmentRepo.UpdateStatus(ctx, tx, shipmentID, "ASSIGNED")

    tx.Commit(ctx)

    // 5. Publish NATS event (side effect, best-effort)
    s.eventBus.Publish(ctx, "shipment.assigned", map[string]interface{}{
        "shipment_id": shipmentID,
        "trip_id":     trip.ID,
        "vehicle_id":  vehicleID,
        "driver_id":   driverID,
    })

    return &AssignmentResult{
        ShipmentID: shipmentID,
        TripID:     trip.ID,
        TripNumber: trip.TripNumber,
    }, nil
}
```

**Key changes:**
- Inject `TripService` as dependency (was missing)
- Inject `EventBus` as dependency (was missing)
- Real `CreateTrip()` call instead of `uuid.New()`
- Single DB transaction for shipment update
- NATS event publishing uncommented

---

## 9. Non-Functional Requirements

### 9.1 Performance

| Operation | Target | Strategy |
|-----------|--------|----------|
| Suggestion engine | <500ms P95 | Redis cached vehicle positions, indexed availability queries |
| Dispatch assign | <1s P95 | Single transaction: create trip + link shipment + publish event |
| Dispatch dashboard | <1s P95 | Aggregate queries, cached summary counts (5s TTL in Redis) |
| Auto status detection | <5s from telemetry | Async worker, batch every 5s |
| Bulk dispatch (50 trips) | <10s P95 | Parallel goroutines, single DB transaction per trip |

### 9.2 Scalability

| Metric | Current | Target | Strategy |
|--------|---------|--------|----------|
| Trips per org/day | — | 200 | `max_trips_per_day` feature flag |
| Concurrent dispatchers | — | 50 | Stateless handlers, shared DB pool |
| Suggestion queries/min | — | 200 | Redis cache, indexed queries |
| Dashboard concurrent users | — | 100 | Redis summary cache (5s TTL) |

### 9.3 Security & RBAC

| Action | Permission | Roles |
|--------|-----------|-------|
| View dispatch dashboard | `dispatch:read` | dispatcher, fleet_manager, org_admin |
| Create dispatch suggestion | `dispatch:read` | dispatcher, fleet_manager, org_admin |
| Assign vehicle+driver | `dispatch:write` | dispatcher, fleet_manager, org_admin |
| Bulk dispatch | `dispatch:write` | dispatcher, fleet_manager, org_admin |
| Re-dispatch (reassign) | `dispatch:write` | fleet_manager, org_admin |
| Override auto-suggestion | `dispatch:write` | dispatcher, fleet_manager, org_admin |
| View dispatch audit log | `audit:read` | fleet_manager, org_admin |

All queries scoped by `org_id` via `TenantContextMiddleware`.

---

## 10. Feature Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `dispatch_enabled` | `true` | Kill switch for entire dispatch feature |
| `auto_suggest_enabled` | `true` | Show vehicle+driver suggestions |
| `auto_status_detection` | `true` | GPS-triggered status changes |
| `auto_receipt_generation` | `true` | Delivery receipt PDF on completion |
| `auto_invoice_creation` | `false` | Auto-create invoice on trip completion |
| `customer_notifications` | `false` | SMS to customer on status changes |
| `bulk_dispatch_enabled` | `false` | Batch assignment |
| `route_optimization_dispatch` | `false` | Generate trips from VRP results |
| `max_trips_per_day` | `200` | Performance guardrail per org |

---

## 11. Testing Strategy

### 11.1 Unit Tests

| Component | Coverage | Scenarios |
|-----------|----------|-----------|
| Dispatch suggestion engine | >90% | Proximity scoring, type matching, capacity filtering, HOS check, no-match |
| Shipment assignment (fixed) | >95% | Real trip creation, trip_id linking, NATS event publishing, rollback on failure |
| Dispatch dashboard handler | >85% | Aggregation, available vehicle/driver queries, alert generation, empty org |
| Redispatch handler | >90% | Cancel old trip, create new trip, re-link shipment, audit log |
| Scoring service | >90% | Weighted scoring (40/30/20/10), edge cases (no driver, zero distance) |

### 11.2 Integration Tests

| Test | Validation |
|------|-----------|
| End-to-end dispatch flow | Booking confirmed → suggestion → assign → trip created → shipment linked → NATS event published |
| Auto status detection | Vehicle ignition → trip STARTED → shipment IN_TRANSIT (within 5s) |
| Auto completion | Vehicle at destination → trip COMPLETED → shipment DELIVERED → receipt generated |
| Redispatch flow | Failed trip → redispatch → new trip created → old trip CANCELLED |
| Multi-tenant isolation | Org A dispatch cannot see/access Org B shipments or trips |
| Double-booking prevention | Same vehicle/driver cannot be assigned to two overlapping trips |

### 11.3 Performance Tests

| Test | Load | Target |
|------|------|--------|
| Suggestion engine | 50 concurrent requests | <500ms P95 |
| Dispatch assign | 20 concurrent assignments | <1s P95 |
| Dispatch dashboard | 20 concurrent dispatchers | <1s P95 |
| Bulk dispatch (50 trips) | 1 request | <10s |

---

## 12. Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Domain wiring complexity | Medium | Medium | Each domain has `RegisterRoutes` — mount under prefix groups |
| Suggestion engine performance | Low | Medium | Cache vehicle positions in Redis (already done by tracking) |
| GPS accuracy in urban areas | High | Medium | Increase radius threshold (100m → 500m) for dense areas |
| Shipment-trip data consistency | Medium | High | Single DB transaction for assign, NATS event as side effect |
| Scope creep (multi-stop, routing) | High | Medium | Strict P0/P1/P2, defer to Phase 2/3 |
| Dispatcher resistance to digital workflow | Medium | Medium | Beta onboarding, show time savings metrics |
| Route domain has no migrations | High | High | Write migrations before wiring route handler |
| Customer domain has zero NATS publishing | Medium | Medium | Add EventBus dependency, uncomment all publish calls |
| Trip domain not wired but has 22 endpoints | Medium | High | Wire first — it's the core of dispatch |

---

## 13. Open Decisions

| # | Question | Decision Needed | Owner |
|---|----------|----------------|-------|
| 1 | Auto-dispatch trigger: on booking confirm or manual? | Default to manual suggest, auto on flag | Product |
| 2 | Max suggestions per shipment? | 3 (configurable) | Engineering |
| 3 | Re-dispatch: cancel old trip or mark failed? | Mark failed, create new trip | Engineering |
| 4 | Dispatch audit retention? | 1 year (compliance) | Compliance |
| 5 | VRP→Trip converter: direct or via dispatch? | Via dispatch assign (reuses existing flow) | Engineering |

---

## 14. Glossary

| Term | Definition | Code Location |
|------|-----------|---------------|
| **Trip** | Delivery assignment: vehicle + driver + pickup + drop | `internal/domains/trip/models/trip.go` |
| **Shipment** | Customer-facing delivery record, linked to trip | `internal/domains/customer/models/models.go` |
| **Booking** | Customer request for delivery service | `internal/domains/customer/models/models.go` |
| **Dispatch** | Process of assigning vehicle+driver to shipment | NEW — this spec |
| **Suggestion** | Ranked vehicle+driver match for a shipment | NEW — dispatch/engine |
| **Redispatch** | Re-assigning a failed trip to new vehicle+driver | NEW — this spec |
| **POD** | Proof of Delivery (signature, photo, OTP) | `internal/domains/trip/models/trip.go` |
| **HOS** | Hours of Service — driver shift limits | `internal/domains/driver/models/driver.go` |
| **VRP** | Vehicle Routing Problem — optimization algorithm | `internal/domains/route/service/solver.go` |

---

## 15. Related Documents

| Document | Path |
|----------|------|
| Product Spec | `product-spec/DISPATCH_FLOW.md` |
| Vehicle Tracking Architecture | `architecture/VEHICLE_TRACKING_ARCHITECTURE.md` |
| Fleet-Core Integration Plan | `services/fleet-core/FLEET_CORE_INTEGRATION_PLAN.md` |
| Database Migrations | `services/fleet-core/migrations/` |
| OpenAPI Spec | `openapi.combined.json` |

---

## 16. Sign-Off

| Role | Name | Date |
|------|------|------|
| Product Manager | _Pending_ | — |
| Engineering Lead | _Pending_ | — |
| Architecture Review | _Pending_ | — |

---

**Document History:**
| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | April 8, 2026 | Senior Architect | Initial architecture based on codebase audit |
