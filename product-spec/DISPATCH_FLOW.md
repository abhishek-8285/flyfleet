# Dispatch Flow — Product Specification

**Document ID:** PSPEC-DF-002
**Version:** 2.0
**Status:** Approved
**Created:** April 7, 2026
**Last Updated:** April 7, 2026
**Author:** Product Team (Codex-audited against codebase)
**Review Cycle:** Bi-weekly
**Parent PRD:** FlyFleet Enterprise PRD
**Related Specs:** VEHICLE_TRACKING.md, VEHICLE_TRACKING_DISPATCHER.md, IAM-AUTH.md
**Priority:** P0 — Q3 Dispatch Flow

---

## 0. Codebase Reality Check

> This spec has been cross-referenced against the actual fleet-core monolith codebase (`services/fleet-core/`). Every table, model, endpoint, and domain referenced here has been verified against existing code.

### 0.1 What Already Exists (No New Code Needed)

| Component | Location | Status |
|-----------|----------|--------|
| **Trip entity** with 50+ fields | `internal/domains/trip/models/trip.go` | ✅ Complete |
| **Trip state machine** (CREATED→SCHEDULED→STARTED→PAUSED→COMPLETED) | `internal/domains/trip/models/trip_lifecycle.go` | ✅ Complete |
| **Trip events** (hash-chained immutable) | `internal/domains/trip/models/trip.go` | ✅ Complete |
| **Trip offers** (assign to driver, accept/reject) | `internal/domains/trip/models/trip.go` | ✅ Complete |
| **POD documents** (signature, photo, OTP) | `internal/domains/trip/models/trip.go` | ✅ Complete |
| **Trip financials** (billing, toll, fuel, GST) | `internal/domains/trip/models/trip.go` | ✅ Complete |
| **Trip compliance** (permit, PUC, e-way bill) | `internal/domains/trip/models/trip.go` | ✅ Complete |
| **Telemetry points** (GPS, speed, heading, fuel) | `internal/domains/trip/models/trip.go` | ✅ Complete |
| **Waypoints** (multi-stop support) | `internal/domains/trip/models/trip.go` | ✅ Complete |
| **`trips` DB table** (PostGIS, JSONB, partitioned telemetry) | `migrations/070_trip_lifecycle.sql` | ✅ Complete |
| **`trip_events` DB table** (hash-chained) | `migrations/070_trip_lifecycle.sql` | ✅ Complete |
| **`trip_offers` DB table** | `migrations/070_trip_lifecycle.sql` | ✅ Complete |
| **`telemetry_points` table** (RANGE partitioned) | `migrations/070_trip_lifecycle.sql` | ✅ Complete |
| **Trip CRUD handlers** | `internal/domains/trip/handlers/` | ✅ Complete |
| **Trip service layer** (create, approve, start, pause, resume, complete, cancel, fail) | `internal/domains/trip/service/` | ✅ Complete |
| **Trip offer endpoints** (POST /trips/{id}/offer, POST /trips/offers/{id}/respond) | `internal/domains/trip/handlers/` | ✅ Complete |
| **Vehicle entity** (9 types, 4 fuel types, GPS device, immobilizer) | `internal/domains/vehicle/models/vehicle.go` | ✅ Complete |
| **Vehicle-driver assignment** | `internal/domains/vehicle/` | ✅ Complete |
| **Driver entity** (55+ fields, behavior scoring, duty status/HOS) | `internal/domains/driver/models/driver.go` | ✅ Complete |
| **Driver behavior scoring** (safety, efficiency, compliance) | `internal/domains/driver/` | ✅ Complete |
| **Customer portal** (bookings, shipments, pricing, SLA) | `internal/domains/customer/` | ✅ Complete |
| **BookingRequest** (REQUESTED→QUOTED→CONFIRMED→REJECTED) | `internal/domains/customer/models/models.go` | ✅ Complete |
| **Shipment** (BOOKED→CONFIRMED→ASSIGNED→IN_TRANSIT→DELIVERED) | `internal/domains/customer/models/models.go` | ✅ Complete |
| **ShipmentStop** (PICKUP/DROPOFF/CHECKPOINT, sequence, ETA) | `internal/domains/customer/models/models.go` | ✅ Complete |
| **DeliveryProof** (receiver_name, signature, photo) | `internal/domains/customer/models/models.go` | ✅ Complete |
| **Shipment assignment service** | `internal/domains/customer/service/shipment_assignment_service.go` | ⚠️ Stub — mock trip ID |
| **Route optimization** (VRP solver, time windows, capacity) | `internal/domains/route/` | ✅ Complete |
| **Geofencing** (zones, entry/exit events, alerts) | `internal/domains/geofencing/` | ✅ Complete |
| **GPS telemetry ingestion** (Redis live cache, history, WebSocket) | `internal/domains/tracking/` | ✅ Complete |
| **Emergency SOS** (8 incident types, vehicle halt, escalation) | `internal/domains/emergency/` | ✅ Complete |
| **NATS event bus** (trip.completed, geofence events) | `internal/platform/events/` | ✅ Complete |
| **PostGIS** (GEOGRAPHY points, GIST indexes) | `migrations/` | ✅ Complete |
| **Trip profitability** analytics | `internal/domains/finance/` | ✅ Complete |
| **Finance** (toll, fuel cards, vendor payments, GST e-invoice, trip profitability) | `internal/domains/finance/` | ✅ Complete |

### 0.2 What Is MISSING (This Spec's Scope)

| Gap | Description | Priority |
|-----|-------------|----------|
| **Auto-dispatch** | No algorithm to match available vehicles+drivers to shipments/bookings | P0 |
| **Dispatch queue** | No UI showing pending assignments, available capacity | P0 |
| **Booking→Trip auto-linking** | Shipment assignment is a stub with mock trip ID | P0 |
| **Route→Trip generation** | VRP results are standalone; no trip creation from optimized routes | P1 |
| **Dispatch rules engine** | No proximity matching, vehicle-cargo matching, driver shift rules | P1 |
| **Real-time dispatch dashboard** | No endpoint showing available vehicles/drivers/pending work | P0 |
| **Bulk dispatch** | No batch assignment of multiple shipments | P2 |
| **Dispatch-specific NATS events** | No `trip.assigned`, `shipment.assigned`, `vehicle.available` events | P1 |
| **Dispatch audit table** | No dedicated dispatch action log (trip_events covers trip-level only) | P2 |
| **Delivery scheduling** | No time slot management, capacity calendar | P2 |
| **Re-dispatch/reassignment** | No flow for failed trips → new vehicle/driver | P1 |
| **Feature flags for dispatch** | Only tracking flags exist (`VehicleMapEnabled`, etc.) | P1 |

---

## 1. Executive Summary

### 1.1 Problem Statement

Fleet dispatchers in Indian logistics companies currently manage deliveries through **manual, error-prone processes**:

- **WhatsApp groups** for driver communication (no audit trail)
- **Phone calls** to assign routes (2-3 hours/day wasted)
- **Paper logs** for delivery tracking (lost records, disputes)
- **Excel sheets** for delivery records (no real-time visibility)
- **No status updates** until driver calls back (30-90 min delays)

This creates:
- **15-20% revenue leakage** from untracked or misreported deliveries
- **No customer visibility** into delivery status or ETA
- **Dispute resolution impossible** (no proof of delivery, no timestamps)
- **Inefficient resource utilization** (vehicles idle while others overloaded)

### 1.2 Solution

A **digital dispatch workflow** that enables dispatchers to create delivery assignments (vehicle + driver + route), track progress in real-time from pending → in-transit → delivered, and auto-generate delivery receipts with GPS-verified timestamps and locations.

**Key insight:** The trip domain, vehicle domain, driver domain, customer portal, and route optimization are **already built**. The missing piece is **wiring them together** into a cohesive dispatch flow — auto-assigning vehicles/drivers to shipments, creating trips from bookings, and providing a real-time dispatch dashboard.

### 1.3 Value Proposition

| Metric | Current State | With Dispatch Flow | Improvement |
|--------|--------------|-------------------|-------------|
| Dispatch time per delivery | 10-15 min (phone + WhatsApp) | 2-3 min (form fill) | **80% faster** |
| Status update latency | 30-90 min (driver callback) | Real-time (GPS-triggered) | **95% faster** |
| Delivery record accuracy | ~70% (manual entry errors) | ~99% (auto-captured) | **29% better** |
| Dispute resolution time | 2-4 hours (investigation) | <5 min (digital receipt) | **95% faster** |
| Daily coordination time | 2-3 hours | 30-45 min | **75% reduction** |

### 1.4 Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Dispatch creation time | <3 min P95 | Time from form open to submission |
| Delivery completion rate | >95% of assigned | Completed / Total assigned |
| Status update accuracy | >98% | GPS-verified vs manual reports |
| Dispatcher adoption | >80% daily usage | % of logged-in dispatchers using feature |
| Delivery receipt generation | 100% of completed | Auto-generated / Completed deliveries |
| Billing integration rate | >90% | Invoices created from deliveries / Completed deliveries |

---

## 2. User Stories

### 2.1 Primary Story (P0)

```
As a dispatcher
I want to assign a vehicle + driver to a delivery route
So that the delivery is tracked from start to completion
```

**Acceptance Criteria:**
- [ ] Create delivery: pickup location → drop location → assigned vehicle/driver
- [ ] Dispatcher sees delivery status: pending → in-transit → delivered
- [ ] Driver can mark delivery complete (via web or future mobile app)
- [ ] Auto-generate delivery receipt with timestamp + location

### 2.2 Extended User Stories

#### Story 2: Auto-Dispatch from Booking (P0)
```
As a dispatcher
I want the system to suggest the best vehicle+driver for a confirmed booking
So that I don't have to manually search for available resources

Acceptance Criteria:
- [ ] When booking is CONFIRMED, system shows top 3 vehicle+driver matches
- [ ] Matches ranked by: proximity to pickup, vehicle type fit, driver availability
- [ ] Dispatcher can accept suggestion or override with manual selection
- [ ] One-click dispatch creates trip and links to shipment
```

#### Story 3: Real-time Dispatch Dashboard (P0)
```
As a dispatcher
I want to see all active deliveries on a dashboard with live status
So that I know what's happening right now

Acceptance Criteria:
- [ ] Dashboard shows: pending, in-transit, delivered counts
- [ ] Each delivery shows: vehicle, driver, route, ETA, status
- [ ] Clicking delivery opens detail view with map
- [ ] Overdue deliveries highlighted in red
- [ ] Available vehicles and drivers shown in sidebar
```

#### Story 4: Delivery Completion with Proof (P1)
```
As a driver or dispatcher
I want to mark a delivery as complete with proof
So that the customer gets confirmation and we can bill accurately

Acceptance Criteria:
- [ ] Mark complete with GPS location capture
- [ ] Optional: photo of delivered goods
- [ ] Optional: customer signature (web pad)
- [ ] Auto-generate delivery receipt (PDF)
- [ ] Trigger invoice creation (if configured)
```

#### Story 5: Exception Handling (P1)
```
As a dispatcher
I want to handle delivery exceptions (delays, failed, rejected)
So that I can take corrective action and document issues

Acceptance Criteria:
- [ ] Mark delivery as delayed with reason code
- [ ] Mark delivery as failed with notes + photo
- [ ] Mark delivery as rejected by customer with reason
- [ ] Reassign to different vehicle/driver if needed
- [ ] Alert fleet manager for critical failures
```

#### Story 6: Bulk Dispatch (P2)
```
As a dispatcher during morning rush
I want to assign multiple deliveries at once
So that I can plan the entire day's schedule in one batch

Acceptance Criteria:
- [ ] Select multiple confirmed bookings from queue
- [ ] Auto-suggest vehicle+driver for each
- [ ] See conflicts (double-booked vehicles/drivers)
- [ ] Confirm all assignments with one click
```

#### Story 7: Route-Optimized Dispatch (P2)
```
As a dispatcher
I want to generate trips from optimized route results
So that I can dispatch multiple deliveries efficiently

Acceptance Criteria:
- [ ] Run route optimization on pending bookings
- [ ] Review suggested vehicle assignments and stop sequences
- [ ] Approve → auto-create trips linked to shipments
- [ ] Drivers receive trip assignments
```

---

## 3. Scope & Prioritization

### 3.1 RICE Score

| Factor | Score | Justification |
|--------|-------|---------------|
| **Reach** | 500 | All beta customers (500 vehicles, ~2,000 deliveries/month) |
| **Impact** | 3 | Massive — core revenue-generating workflow |
| **Confidence** | 85% | Trip tables, vehicle, driver, customer, route domains ALL exist. Only wiring needed. |
| **Effort** | 8 | 8 person-sprints (less than original estimate because domains are pre-built) |
| **RICE** | **1,594** | (500 × 3 × 0.85) / 8 |

**Note:** HIGHER RICE than original estimate (1,231) because codebase audit revealed most domains are already built. Effort is lower — this is primarily integration work, not greenfield development.

### 3.2 Epics Breakdown

| Epic | Priority | Effort | Dependencies | Sprint |
|------|----------|--------|--------------|--------|
| **E1: Wire domains into main.go** | P0 | 1 sprint | None — code exists | Q3 W1 |
| **E2: Booking→Trip auto-linking** | P0 | 1 sprint | E1, existing shipment_assignment_service.go | Q3 W2 |
| **E3: Auto-dispatch suggestion engine** | P0 | 2 sprints | E2, existing route optimization + vehicle/driver availability | Q3 W3-4 |
| **E4: Dispatch dashboard endpoint** | P0 | 1 sprint | E1, existing trip queries | Q3 W5 |
| **E5: Dispatch NATS events** | P1 | 1 sprint | E1, existing event bus | Q4 W1 |
| **E6: Re-dispatch/reassignment flow** | P1 | 1 sprint | E2, existing trip cancel + create | Q4 W2 |
| **E7: Dispatch feature flags** | P1 | 0.5 sprint | E1 | Q4 W3 |
| **E8: Bulk dispatch** | P2 | 1 sprint | E3 | Q4 W4 |
| **E9: Route→Trip generation** | P2 | 1 sprint | E3, existing route results | Q4 W5 |

### 3.3 In Scope (MVP — Q3-Q4)

- ✅ Wire all existing domain handlers into `cmd/main.go` router
- ✅ Auto-link confirmed bookings to trips (replace mock trip ID in shipment_assignment_service.go)
- ✅ Auto-dispatch suggestion engine (proximity + availability + vehicle type matching)
- ✅ Dispatch dashboard endpoint (aggregate active trips, available vehicles/drivers, alerts)
- ✅ Dispatch-specific NATS events (`trip.assigned`, `trip.dispatched`, `shipment.assigned`)
- ✅ Re-dispatch flow (cancel failed trip → create new trip → reassign)
- ✅ Feature flags for dispatch (`dispatch_enabled`, `auto_suggest`, `auto_dispatch`)

### 3.4 Out of Scope (Future Phases)

- ❌ AI route optimization (Phase 3)
- ❌ Driver mobile app (Phase 2 — Q9)
- ❌ Customer tracking portal (Phase 2)
- ❌ Automated ETA calculation (Phase 3)
- ❌ Dynamic re-routing (Phase 3)
- ❌ Load optimization (weight/volume constraints) (Phase 3)
- ❌ Driver shift scheduling (Phase 2)
- ❌ Customer signature capture (Phase 2 — mobile app dependent)

---

## 4. Technical Architecture

### 4.1 Current State vs Target State

#### Current: Domains Exist But Not Wired

```
services/fleet-core/cmd/main.go
  └── Chi router
       ├── /health
       ├── /api/v1/telemetry*          (tracking domain — wired)
       ├── /api/v1/vehicles/locations* (tracking domain — wired)
       ├── /ws/telemetry*              (tracking domain — wired)
       ├── /api/v1/geofences*          (geofencing domain — wired)
       └── /api/v1/auth/*              (IAM domain — wired)

  ❌ trip/       handlers exist, NOT mounted
  ❌ vehicle/    handlers exist, NOT mounted
  ❌ driver/     handlers exist, NOT mounted
  ❌ customer/   handlers exist, NOT mounted
  ❌ route/      handlers exist, NOT mounted
  ❌ maintenance/handlers exist, NOT mounted
  ❌ fuel/       handlers exist, NOT mounted
  ❌ emergency/  handlers exist, NOT mounted
  ❌ video/      handlers exist, NOT mounted
  ❌ compliance/ handlers exist, NOT mounted
  ❌ finance/    handlers exist, NOT mounted
  ❌ analytics/  handlers exist, NOT mounted
  ❌ reports/    handlers exist, NOT mounted
  ❌ billing/    handlers exist, NOT mounted
  ❌ governance/ handlers exist, NOT mounted
  ❌ diagnostics/handlers exist, NOT mounted
  ❌ admin_support/handlers exist, NOT mounted
```

#### Target: All Domains Wired + Dispatch Flow

```
services/fleet-core/cmd/main.go
  └── Chi router
       ├── /health
       │
       ├── /api/v1/auth/*              (IAM — 74 endpoints)
       │
       ├── /api/v1/tracking/*          (GPS telemetry — already wired)
       │
       ├── /api/v1/vehicles/*          (Vehicle CRUD, driver assignment)
       ├── /api/v1/drivers/*           (Driver CRUD, behavior, HOS)
       ├── /api/v1/trips/*             (Trip CRUD, state machine, offers, POD)
       ├── /api/v1/customers/*         (Customer portal, bookings, shipments)
       ├── /api/v1/routes/*            (Route optimization, VRP solver)
       ├── /api/v1/geofences/*         (Geofence CRUD, alerts)
       │
       ├── /api/v1/dispatch/*          (NEW — Dispatch flow)
       │    ├── GET  /dashboard                    (aggregate view)
       │    ├── POST /suggest                      (auto-suggest vehicle+driver)
       │    ├── POST /assign                       (create trip from shipment)
       │    ├── POST /bulk-assign                   (batch dispatch)
       │    └── POST /redispatch                    (reassign failed trip)
       │
       └── /api/v1/...             (maintenance, fuel, emergency, video, etc.)
```

### 4.2 Dispatch Flow — End-to-End

#### Flow 1: Booking → Auto-Dispatch → Trip → Delivery

```
1. Customer creates BookingRequest via API
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
   ┌─────────────────────────────────────────────┐
   │ Dispatch Suggestion Engine                   │
   │ 1. Find available vehicles near pickup       │
   │ 2. Filter by vehicle type matching cargo     │
   │ 3. Find available drivers for those vehicles │
   │ 4. Check driver HOS (not over shift limit)   │
   │ 5. Score each match (proximity, fit, rating) │
   │ 6. Return top 3 suggestions                  │
   └─────────────────────────────────────────────┘

5. Dispatcher reviews suggestions
   GET /api/v1/dispatch/suggest?shipment_id={id}
   → Returns: [{vehicle, driver, score, reason}, ...]

6. Dispatcher accepts suggestion (or overrides)
   POST /api/v1/dispatch/assign
   → Creates Trip (CREATED → SCHEDULED)
   → Links Shipment to Trip (shipment.trip_id)
   → Shipment status: ASSIGNED
   → NATS event: trip.assigned

7. Trip starts (auto-detect via GPS telemetry)
   Vehicle ignition ON near pickup → Trip status: STARTED
   → Shipment status: IN_TRANSIT
   → NATS event: trip.started
   → WebSocket push to dispatch dashboard

8. Trip completes (auto-detect via GPS)
   Vehicle stops at drop location → Trip status: COMPLETED
   → Shipment status: DELIVERED
   → POD captured (GPS timestamp + location)
   → NATS event: trip.completed
   → Delivery receipt PDF generated
   → Invoice created (if auto-invoice enabled)
```

### 4.3 Dispatch Suggestion Engine — Algorithm

```go
// DispatchSuggestionEngine suggests best vehicle+driver for a shipment
type DispatchSuggestionEngine struct {
    vehicleRepo  *vehicle.Repository
    driverRepo   *driver.Repository
    tripRepo     *trip.Repository
    geofenceRepo *geofencing.Repository
    log          *zap.Logger
}

type DispatchSuggestion struct {
    Vehicle       *models.Vehicle
    Driver        *models.Driver
    Score         float64  // 0-100, higher is better
    Reasons       []string // Why this match scored high
    DistanceToPickupKm float64
    ETAToPickupMin     int
}

// Suggest returns top N vehicle+driver matches for a shipment
func (e *DispatchSuggestionEngine) Suggest(
    ctx context.Context,
    orgID uuid.UUID,
    pickupLat, pickupLng float64,
    vehicleType models.VehicleType,
    cargoWeightKg float64,
    topN int,
) ([]DispatchSuggestion, error) {
    // 1. Find vehicles in org that are available (not on active trip, not in maintenance)
    vehicles, err := e.vehicleRepo.GetAvailableVehicles(ctx, orgID)

    // 2. Filter by vehicle type match (TRUCK for heavy, VAN for light, etc.)
    vehicles = e.filterByType(vehicles, vehicleType)

    // 3. Filter by capacity (cargoWeight <= vehicle capacity)
    vehicles = e.filterByCapacity(vehicles, cargoWeightKg)

    // 4. Score each vehicle by distance to pickup
    var suggestions []DispatchSuggestion
    for _, v := range vehicles {
        // Get vehicle's last known GPS position from Redis live cache
        lastPos, err := e.vehicleRepo.GetLastKnownPosition(ctx, v.ID)

        dist := haversineDistance(lastPos.Lat, lastPos.Lng, pickupLat, pickupLng)
        eta := estimateTravelTime(dist, v.VehicleType)

        // Find assigned driver for this vehicle
        driver, err := e.driverRepo.GetByVehicleID(ctx, v.ID)

        // Check driver availability (not on active trip, not over HOS limit)
        if driver != nil && !e.isDriverAvailable(ctx, driver.ID) {
            continue // Skip — driver is busy
        }

        // Check vehicle not halted due to emergency
        if e.isVehicleHalted(ctx, v.ID) {
            continue
        }

        // Calculate score
        score := e.calculateScore(dist, v, driver)

        suggestions = append(suggestions, DispatchSuggestion{
            Vehicle:            v,
            Driver:             driver,
            Score:              score,
            DistanceToPickupKm: dist,
            ETAToPickupMin:     eta,
            Reasons:            e.explainScore(dist, v, driver),
        })
    }

    // 5. Sort by score descending, return top N
    sort.Slice(suggestions, func(i, j int) bool {
        return suggestions[i].Score > suggestions[j].Score
    })

    if len(suggestions) > topN {
        suggestions = suggestions[:topN]
    }

    return suggestions, nil
}

// Score: 40% proximity, 30% vehicle fit, 20% driver rating, 10% availability
func (e *DispatchSuggestionEngine) calculateScore(
    distKm float64, v *models.Vehicle, d *models.Driver,
) float64 {
    // Proximity score (closer = better): 100 - (distKm * 5), min 0
    proximityScore := max(0, 100-distKm*5)

    // Vehicle fit score: capacity utilization (70-90% is ideal)
    fitScore := 100.0 // Default if no cargo info

    // Driver rating score (0-5 → 0-100)
    driverScore := d.Rating * 20

    // Weighted combination
    return proximityScore*0.4 + fitScore*0.3 + driverScore*0.2 + 10.0 // 10% base availability
}
```

### 4.4 Dispatch Dashboard Endpoint

```
GET /api/v1/dispatch/dashboard?date=2026-04-07
```

**Response:**
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
      "current_location": {
        "latitude": 28.5706,
        "longitude": 77.2620
      }
    }
  ],
  "available_vehicles": [
    {
      "id": "v-789",
      "registration": "DL01XY9876",
      "type": "Tata Ace Gold",
      "capacity_tonnage": 1.5,
      "last_known_location": {"lat": 28.5355, "lng": 77.2910},
      "status": "available"
    }
  ],
  "available_drivers": [
    {
      "id": "d-456",
      "name": "Amit Singh",
      "rating": 4.5,
      "total_trips": 342,
      "current_status": "available"
    }
  ],
  "alerts": [
    {
      "type": "delay_warning",
      "trip_id": "trip-002",
      "message": "TRIP-2026-00124 is 15 min behind schedule",
      "severity": "warning",
      "timestamp": "2026-04-07T10:45:00Z"
    }
  ]
}
```

### 4.5 Database — What Exists vs What's New

#### Already Exists (No Schema Changes Needed)

```sql
-- trips table (50+ columns, PostGIS, JSONB)
-- trip_events table (hash-chained immutable events)
-- trip_offers table (OFFERED/ACCEPTED/REJECTED/EXPIRED)
-- telemetry_points table (RANGE partitioned by month)
-- customer_v2_shipments table (with trip_id FK)
-- customer_v2_booking_requests table
-- customer_v2_shipment_stops table
-- customer_v2_delivery_proofs table
-- vehicles table (encrypted chassis/engine numbers)
-- drivers table (55+ fields)
-- route_optimization_requests/results/jobs tables
```

#### New Tables Required

```sql
-- Dispatch action log (audit trail for dispatch decisions)
CREATE TABLE dispatch_actions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    action_type     VARCHAR(30) NOT NULL CHECK (action_type IN (
        'auto_suggest',
        'manual_assign',
        'auto_assign',
        'bulk_assign',
        'redispatch',
        'override'
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

-- Dispatch suggestions cache (short TTL, for dashboard)
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

### 4.6 NATS Events — New Dispatch Events

```go
// Existing events
"trip.completed"          // ✅ Already published
"fleet.geofence.events"   // ✅ Already published

// NEW dispatch events
"trip.assigned"            // Trip created + vehicle+driver assigned
"trip.dispatched"          // Trip moved to SCHEDULED, ready to start
"trip.started"             // Auto-detected: vehicle ignition at pickup
"trip.reassigned"          // Failed trip → new vehicle+driver
"shipment.assigned"        // Shipment linked to trip
"shipment.in_transit"      // Shipment status changed
"shipment.delivered"       // Shipment completed
"vehicle.available"        // Vehicle freed up (trip completed)
"driver.available"         // Driver freed up (trip completed)
```

### 4.7 API Contracts — New Dispatch Endpoints

#### POST /api/v1/dispatch/suggest

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

**Response:**
```json
{
  "suggestions": [
    {
      "rank": 1,
      "vehicle": {
        "id": "v-123",
        "registration": "DL01AB1234",
        "type": "Tata Ace Gold",
        "capacity_tonnage": 1.5
      },
      "driver": {
        "id": "d-456",
        "name": "Rajesh Kumar",
        "rating": 4.5,
        "total_trips": 342
      },
      "score": 92.5,
      "distance_to_pickup_km": 3.2,
      "eta_to_pickup_min": 12,
      "reasons": ["Closest available vehicle", "Perfect capacity match", "High driver rating"]
    }
  ]
}
```

#### POST /api/v1/dispatch/assign

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

**Response:**
```json
{
  "trip_id": "trip-001",
  "trip_number": "TRIP-2026-00123",
  "status": "scheduled",
  "shipment_status": "assigned",
  "message": "Trip created and linked to shipment"
}
```

---

## 5. Integration Points

### 5.1 Existing Domain Integrations

| Domain | Integration | Direction |
|--------|------------|-----------|
| **trip/** | Dispatch creates trips via existing `CreateTrip` handler | Dispatch → Trip |
| **vehicle/** | Dispatch queries vehicle availability, last GPS position | Dispatch ← Vehicle |
| **driver/** | Dispatch queries driver availability, HOS status | Dispatch ← Driver |
| **customer/** | Dispatch reads confirmed bookings, updates shipment status | Dispatch ↔ Customer |
| **tracking/** | Dispatch reads live vehicle positions (Redis cache) | Dispatch ← Tracking |
| **route/** | Dispatch can trigger route optimization for batch dispatch | Dispatch → Route |
| **finance/** | Trip completion triggers invoice via finance domain | Trip → Finance |
| **geofencing/** | Auto-complete uses geofence proximity detection | Dispatch ← Geofencing |
| **emergency/** | Vehicle halt state blocks trip starts | Emergency → Dispatch |

### 5.2 Shipment Assignment — Fix the Stub

**Current code** (`customer/service/shipment_assignment_service.go`):
```go
// Stub: In production, we'd interact with trip_v2 to assign a vehicle,
// driver and create a trip.
tripID := uuid.New() // Mock trip ID
```

**Fixed implementation:**
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

    // 3. Create trip via trip service (NOT mock)
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

    // 4. Link shipment to trip
    err = s.shipmentRepo.UpdateTripID(ctx, shipmentID, trip.ID)

    // 5. Update shipment status
    err = s.shipmentRepo.UpdateStatus(ctx, shipmentID, "ASSIGNED")

    // 6. Publish NATS event
    s.eventBus.Publish(ctx, "shipment.assigned", &events.Event{
        Data: map[string]interface{}{
            "shipment_id": shipmentID,
            "trip_id":     trip.ID,
            "vehicle_id":  vehicleID,
            "driver_id":   driverID,
        },
    })

    return &AssignmentResult{
        ShipmentID: shipmentID,
        TripID:     trip.ID,
        TripNumber: trip.TripNumber,
    }, nil
}
```

---

## 6. State Machines

### 6.1 Trip State Machine (Already Exists)

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

### 6.2 Shipment State Machine (Already Exists)

```
BOOKED ──confirm──→ CONFIRMED ──assign──→ ASSIGNED ──start──→ IN_TRANSIT
                                                              │
                           ┌─cancel──→ CANCELLED              │
                           │                                  │ complete
                           │                                  ▼
                           └──────────────────────────── DELIVERED
```

### 6.3 Dispatch State Machine (New)

```
PENDING_BOOKING ──booking_confirmed──→ PENDING_DISPATCH
     │                                        │
     │ (booking rejected)                     │ auto_suggest
     ▼                                        ▼
  REJECTED                             SUGGESTION_READY
                                              │
                                              │ dispatch_assign
                                              ▼
                                        DISPATCHED ──trip_started──→ IN_TRANSIT
                                              │                          │
                                   redispatch   │                          │ trip_completed
                                   (if failed)  │ fail                     ▼
                                                ▼                   DELIVERED
                                          REASSIGNED
```

---

## 7. Performance Requirements

| Operation | Target | Strategy |
|-----------|--------|----------|
| Suggestion engine | <500ms P95 | Redis cached vehicle positions, indexed availability queries |
| Dispatch assign | <1s P95 | Single transaction: create trip + link shipment + publish event |
| Dispatch dashboard | <1s P95 | Aggregate queries, cached summary counts (5s TTL in Redis) |
| Auto status detection | <5s from telemetry | Async via NATS, batch every 5s |
| Bulk dispatch (50 trips) | <10s P95 | Parallel goroutines, single DB transaction per trip |

---

## 8. Security & Access Control

### 8.1 RBAC Requirements

| Action | Required Permission | Roles |
|--------|-------------------|-------|
| View dispatch dashboard | `dispatch:read` | dispatcher, fleet_manager, org_admin |
| Create dispatch suggestion | `dispatch:read` | dispatcher, fleet_manager, org_admin |
| Assign vehicle+driver | `dispatch:write` | dispatcher, fleet_manager, org_admin |
| Bulk dispatch | `dispatch:write` | dispatcher, fleet_manager, org_admin |
| Re-dispatch (reassign) | `dispatch:write` | fleet_manager, org_admin |
| Override auto-suggestion | `dispatch:write` | dispatcher, fleet_manager, org_admin |
| View dispatch audit log | `audit:read` | fleet_manager, org_admin |

### 8.2 Multi-Tenant Isolation

All dispatch queries scoped by `org_id` via:
- RLS policies on PostgreSQL tables
- Tenant context middleware extracts org_id from JWT claims
- NATS events include `org_id` for consumer filtering

---

## 9. Feature Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `dispatch_enabled` | `true` | Kill switch for entire dispatch feature |
| `auto_suggest_enabled` | `true` | Show vehicle+driver suggestions for confirmed bookings |
| `auto_status_detection` | `true` | GPS-triggered status changes (existing trip feature) |
| `auto_receipt_generation` | `true` | Delivery receipt PDF creation on completion |
| `auto_invoice_creation` | `false` | Auto-create invoice on trip completion |
| `customer_notifications` | `false` | SMS to customer on status changes |
| `bulk_dispatch_enabled` | `false` | Batch assignment of multiple shipments |
| `route_optimization_dispatch` | `false` | Generate trips from VRP results |
| `max_trips_per_day` | `200` | Performance guardrail per org |

---

## 10. Testing Strategy

### 10.1 Unit Tests

| Component | Coverage Target | Key Scenarios |
|-----------|----------------|---------------|
| Dispatch suggestion engine | >90% | Proximity scoring, type matching, capacity filtering, HOS check |
| Shipment assignment (fixed) | >95% | Real trip creation, trip_id linking, NATS event publishing |
| Dispatch dashboard handler | >85% | Aggregation, available vehicle/driver queries, alert generation |
| Redispatch handler | >90% | Cancel old trip, create new trip, re-link shipment |

### 10.2 Integration Tests

| Test | Validation |
|------|-----------|
| End-to-end dispatch flow | Booking confirmed → suggestion → assign → trip created → shipment linked → NATS event published |
| Auto status detection | Vehicle ignition → trip STARTED → shipment IN_TRANSIT (within 5s) |
| Auto completion | Vehicle at destination → trip COMPLETED → shipment DELIVERED → receipt generated |
| Redispatch flow | Failed trip → redispatch → new trip created → old trip CANCELLED |
| Multi-tenant isolation | Org A dispatch cannot see/access Org B shipments or trips |

### 10.3 Performance Tests

| Test | Load | Target |
|------|------|--------|
| Suggestion engine | 50 concurrent requests | <500ms P95 |
| Dispatch assign | 20 concurrent assignments | <1s P95 |
| Dispatch dashboard | 20 concurrent dispatchers | <1s P95 |
| Bulk dispatch (50 trips) | 1 request | <10s |

---

## 11. Rollout Plan

### 11.1 Phased Deployment

| Phase | Timeline | Scope | Success Criteria |
|-------|----------|-------|-----------------|
| **E1: Wire domains** | Q3 W1 | Mount all handlers in main.go | All domain endpoints respond |
| **E2: Fix shipment assignment** | Q3 W2 | Replace mock trip ID with real trip creation | Booking → Trip linkage works end-to-end |
| **E3: Suggestion engine** | Q3 W3-4 | Proximity + availability + type matching | Top 3 suggestions returned in <500ms |
| **E4: Dispatch dashboard** | Q3 W5 | Aggregate endpoint | Dashboard loads in <1s |
| **E5: NATS events** | Q4 W1 | Dispatch-specific events | Events published and consumed |
| **E6: Redispatch** | Q4 W2 | Failed trip → new assignment | Redispatch creates new trip, cancels old |
| **Beta** | Q4 W3-4 | 3 beta customers | <3 min dispatch time, >95% completion |

### 11.2 Monitoring & Alerts

| Metric | Threshold | Alert Channel |
|--------|-----------|---------------|
| Suggestion latency P95 | >1s | Engineering Slack |
| Dispatch assign failure rate | >2% | PagerDuty |
| Auto status detection lag | >10s | PagerDuty |
| Dashboard load time P95 | >2s | Engineering Slack |
| Shipment-trip linkage failures | >0 | Engineering Slack (data integrity) |

---

## 12. Dependencies & Risks

### 12.1 Dependencies

| Dependency | Status | Impact |
|-----------|--------|--------|
| Trip domain | ✅ Complete, NOT wired | P0 — wire into main.go first |
| Vehicle domain | ✅ Complete, NOT wired | P0 — needed for availability checks |
| Driver domain | ✅ Complete, NOT wired | P0 — needed for availability checks |
| Customer domain | ✅ Complete, NOT wired | P0 — booking/shipment data source |
| Tracking domain | ✅ Wired | None — live GPS positions available |
| Route optimization | ✅ Complete, NOT wired | P1 — for route-optimized dispatch |
| NATS event bus | ✅ Available | P1 — dispatch events |
| IAM (auth) | ✅ Wired, 74 endpoints | None — RBAC in place |

### 12.2 Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Domain wiring complexity | Medium | Medium | Each domain has `RegisterRoutes` — mount under prefix groups |
| Suggestion engine performance | Low | Medium | Cache vehicle positions in Redis (already done by tracking domain) |
| GPS accuracy in urban areas | High | Medium | Increase radius threshold (100m → 500m) for dense areas |
| Shipment-trip data consistency | Medium | High | Single DB transaction for assign operation, NATS event as side effect |
| Scope creep (multi-stop, routing) | High | Medium | Strict P0/P1/P2, defer to Phase 2/3 |
| Dispatcher resistance to digital workflow | Medium | Medium | Beta onboarding, show time savings metrics |

---

## 13. Glossary

| Term | Definition | Code Location |
|------|-----------|---------------|
| **Trip** | Delivery assignment: vehicle + driver + pickup + drop | `internal/domains/trip/models/trip.go` |
| **Shipment** | Customer-facing delivery record, linked to trip | `internal/domains/customer/models/models.go` |
| **Booking** | Customer request for delivery service | `internal/domains/customer/models/models.go` |
| **Dispatch** | Process of assigning vehicle+driver to shipment | NEW — this spec |
| **Suggestion** | Ranked vehicle+driver match for a shipment | NEW — this spec |
| **Redispatch** | Re-assigning a failed trip to new vehicle+driver | NEW — this spec |
| **POD** | Proof of Delivery (signature, photo, OTP) | `internal/domains/trip/models/trip.go` |
| **HOS** | Hours of Service — driver shift limits | `internal/domains/driver/models/driver.go` |

---

## 14. Appendix

### 14.1 API Endpoints — Complete Dispatch Set

| Endpoint | Method | Purpose | Status |
|----------|--------|---------|--------|
| `/api/v1/dispatch/dashboard` | GET | Dispatch dashboard aggregate | ❌ New |
| `/api/v1/dispatch/suggest` | POST | Get vehicle+driver suggestions | ❌ New |
| `/api/v1/dispatch/assign` | POST | Create trip from shipment | ❌ New |
| `/api/v1/dispatch/bulk-assign` | POST | Batch dispatch multiple shipments | ❌ New |
| `/api/v1/dispatch/redispatch` | POST | Reassign failed trip | ❌ New |
| `/api/v1/trips` | POST | Create trip (existing) | ✅ Exists, ❌ Not wired |
| `/api/v1/trips/{id}` | GET | Get trip details (existing) | ✅ Exists, ❌ Not wired |
| `/api/v1/trips/{id}/status` | PATCH | Update trip status (existing) | ✅ Exists, ❌ Not wired |
| `/api/v1/trips/{id}/offer` | POST | Offer trip to driver (existing) | ✅ Exists, ❌ Not wired |
| `/api/v1/customers/bookings` | POST | Create booking (existing) | ✅ Exists, ❌ Not wired |
| `/api/v1/customers/shipments/{id}` | PATCH | Update shipment (existing) | ✅ Exists, ❌ Not wired |

### 14.2 Codebase Map

```
services/fleet-core/
├── cmd/main.go                          ← Mount all domain routers here
├── internal/domains/
│   ├── trip/                            ← Trip CRUD, state machine, offers
│   │   ├── handlers/trip_handler.go     ← RegisterRoutes()
│   │   ├── service/trip_service.go      ← CreateTrip, ApproveTrip, StartTrip...
│   │   ├── models/trip.go               ← Trip, TripEvent, TripOffer, POD...
│   │   └── models/trip_lifecycle.go     ← State machine transitions
│   ├── vehicle/                         ← Vehicle management
│   │   ├── handlers/vehicle_handler.go  ← RegisterRoutes()
│   │   ├── models/vehicle.go            ← Vehicle entity
│   │   └── models/trip.go               ← Vehicle-scoped trip view
│   ├── driver/                          ← Driver management
│   │   ├── handlers/driver_handler.go   ← RegisterRoutes()
│   │   └── models/driver.go             ← Driver entity, HOS, behavior
│   ├── customer/                        ← Customer portal
│   │   ├── handlers/booking_handler.go  ← RegisterRoutes()
│   │   ├── handlers/shipment_handler.go ← RegisterRoutes()
│   │   ├── models/models.go             ← Booking, Shipment, DeliveryProof...
│   │   └── service/shipment_assignment_service.go ← FIX: Replace mock trip ID
│   ├── route/                           ← Route optimization
│   │   └── handlers/route_handler.go    ← RegisterRoutes(), VRP solver
│   ├── tracking/                        ← GPS telemetry (already wired)
│   │   └── service/ingestion_service.go ← Live vehicle positions (Redis)
│   └── geofencing/                      ← Geofence detection
│       └── engine/geofence_engine.go    ← Proximity checks
└── internal/platform/
    └── events/events.go                 ← NATS event bus
```

### 14.3 Change Log

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | April 7, 2026 | Product Team | Initial draft |
| 2.0 | April 7, 2026 | PM Toolkit + Codex Audit | Cross-referenced against actual codebase. Updated scope, effort, and dependencies to match reality. Identified 12 gaps. Lowered effort from 13 to 8 sprints (domains pre-built). |

---

**Next Review:** April 21, 2026
**Approvers:** Product Lead, Engineering Lead, Design Lead
**Stakeholders:** Dispatchers (beta), Fleet Managers, Engineering, Customer Success
**Audit Trail:** Codex security audit applied to auth-service (9 vulnerabilities found + fixed). This spec verified against live codebase.
