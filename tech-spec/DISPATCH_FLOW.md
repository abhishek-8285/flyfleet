# Technical Specification: Dispatch Flow

**Document ID:** TECH-DF-001
**Version:** 1.0
**Status:** Draft
**Created:** April 8, 2026
**Author:** Senior Backend Engineer
**Parent Spec:** [Dispatch Flow PSPEC](../product-spec/DISPATCH_FLOW.md)
**Parent Architecture:** [Dispatch Flow Architecture](../architecture/DISPATCH_FLOW_ARCHITECTURE.md)
**Priority:** P0 — Q3-Q4 Delivery

---

## 1. Problem Statement & Goals

### 1.1 Problem

Fleet dispatchers manage deliveries through manual, error-prone processes:
- WhatsApp groups for driver communication (no audit trail)
- Phone calls to assign routes (2–3 hours/day wasted)
- Paper logs for delivery tracking (lost records, disputes)
- No status updates until driver calls back (30–90 min delays)

This creates **15–20% revenue leakage** from untracked deliveries and **no customer visibility** into delivery status or ETA.

### 1.2 User Story

> As a dispatcher, I want to assign a vehicle + driver to a delivery route so that the delivery is tracked from start to completion.

### 1.3 Goals & Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Dispatch creation time | **<3 min P95** | Time from form open to submission |
| Delivery completion rate | **>95%** | Completed / Total assigned |
| Status update accuracy | **>98%** | GPS-verified vs manual reports |
| Dispatcher adoption | **>80% daily usage** | % of logged-in dispatchers using feature |
| Dispatch suggestion latency | **<500ms P95** | POST /dispatch/suggest response time |
| Dispatch assign latency | **<1s P95** | POST /dispatch/assign response time |
| Dispatch dashboard latency | **<1s P95** | GET /dispatch/dashboard response time |
| Auto status detection | **<5s from telemetry** | GPS → trip status transition |
| Bulk dispatch (50 trips) | **<10s P95** | POST /dispatch/bulk-assign |
| Max trips per org/day | **200** | Guardrail via feature flag |
| Concurrent dispatchers | **50** | Per fleet-core instance |

### 1.4 Current State Summary — Verified Against Codebase

**Only 2 domains are wired into `cmd/main.go`:** IAM (auth) and Tracking (telemetry, geofences, WebSocket). All other domains — including trip, vehicle, driver, customer, route, alerts — are **built but NOT mounted**.

| Domain | Code Status | Wired? | Key Reality |
|--------|------------|--------|-------------|
| **trip/** | ✅ 22 endpoints, state machine (8 states), NATS events, financials, POD, multi-stop | ❌ No | `TripService.Create()` (not `CreateTrip`); methods named `Create`, `Start`, `Complete`, `Cancel`, `Fail`, `Pause`, `Resume` |
| **vehicle/** | ✅ 8 endpoints, CRUD, search, stats, driver assignment | ❌ No | No `GetAvailableVehicles()` method — must be added |
| **driver/** | ✅ 30+ endpoints, CRUD, behavior scoring, HOS, duty status, shifts, payroll | ❌ No | No `GetByVehicleID()` method — must be added |
| **customer/** | ✅ 15 endpoints, bookings, shipments, stops, delivery proof | ❌ No | `ShipmentAssignmentService` is a **stub** — uses `uuid.New()` for fake trip ID, EventBus not injected, Publish calls commented out |
| **route/** | ✅ 3 solvers (Clarke-Wright, Heuristic, NN), time windows, capacity | ❌ No | Tables exist in migrations 068/028. **No `road_restrictions` or `vehicle_permits` tables** — not found in any migration. No Trip converter. |
| **tracking/** | ✅ GPS telemetry, live cache, WebSocket | ✅ Yes | Redis `live:loc:{id}`, pub/sub `org:{org}:telemetry` |
| **geofencing/** | ✅ PostGIS detection, NATS subscriber | ✅ Yes | `fleet.telemetry.gps` subscriber, `fleet.geofence.events` publisher |
| **alerts/** | ✅ Multi-channel, preference engine, template engine | ❌ No | Built but not mounted |
| **dispatch/** | ❌ Does NOT exist | N/A | Must be created from scratch |

> ⚠️ **Critical finding:** This feature is primarily **integration work** — wiring existing domains into `cmd/main.go` and fixing the shipment assignment stub — plus **one new domain** (dispatch/) for suggestion engine and orchestration.

---

## 2. API Specifications

### 2.1 GET /api/v1/dispatch/dashboard — Aggregate Dashboard View (NEW)

**Purpose:** Single-call aggregate view of dispatch operations for a given date.

**Request:**
```http
GET /api/v1/dispatch/dashboard?date=2026-04-07&include_available=true
Authorization: Bearer <JWT via JWKSAuthGate>
```

**Query Parameters:**

| Param | Type | Description | Default |
|-------|------|-------------|---------|
| `date` | YYYY-MM-DD | Filter date | Today |
| `include_available` | bool | Include available vehicles/drivers | `true` |
| `include_pending` | bool | Include pending shipments | `true` |

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
      "shipment_id": "shp-uuid-001",
      "booking_id": "bk-uuid-001",
      "customer_name": "ABC Trading",
      "pickup": {"address": "Delhi Warehouse", "lat": 28.5355, "lng": 77.2910},
      "drop": {"address": "Noida Sector 62", "lat": 28.5706, "lng": 77.2620},
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
      "id": "trip-uuid-001",
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
    {
      "id": "v-uuid-789",
      "registration": "DL01XY9876",
      "type": "Tata Ace Gold",
      "capacity_tonnage": 1.5,
      "status": "available"
    }
  ],
  "available_drivers": [
    {
      "id": "d-uuid-456",
      "name": "Amit Singh",
      "rating": 4.5,
      "total_trips": 342,
      "status": "available"
    }
  ],
  "alerts": [
    {
      "type": "delay_warning",
      "trip_id": "trip-uuid-002",
      "message": "TRIP-2026-00124 is 15 min behind schedule",
      "severity": "warning",
      "timestamp": "2026-04-07T10:45:00Z"
    }
  ]
}
```

**Auth & Authorization:**

| Layer | Middleware | Detail |
|-------|-----------|--------|
| Auth | `JWKSAuthGate(cfg.JWTSecret, logger)` | JWT validation |
| Tenant | `TenantMiddleware(logger)` | Extracts org_id from context |
| RBAC | `RequireRole("dispatcher", "fleet_manager", "org_admin")` | Read access |
| Rate Limit | `RateLimit` | 60 req/min per user |

**Performance Budget:**

| Stage | Target | Strategy |
|-------|--------|----------|
| Redis cache check | <5ms | Key: `dispatch:dashboard_summary:{org_id}`, TTL 5s |
| Trip summary query | <200ms | COUNT by status, indexed on org_id + date |
| Pending shipments | <300ms | Query customer_v2_shipments WHERE trip_id IS NULL |
| Active trips | <200ms | Query trips WHERE status IN (STARTED, PAUSED, RESUMED) |
| Available vehicles | <100ms | Query vehicles WHERE status='ACTIVE' AND not on active trip |
| Available drivers | <100ms | Query drivers WHERE status='active' AND current_trip_id IS NULL |
| Alert generation | <100ms | Check delayed trips |
| JSON assembly | <50ms | Pre-allocated structs |
| **Total P95** | **<1s** | Well within budget |

**Error Responses:**

| Status | Body | Condition |
|--------|------|-----------|
| 401 | `{"error":"unauthorized"}` | Invalid/missing JWT |
| 403 | `{"error":"insufficient permissions"}` | RBAC role check failed |
| 429 | `{"error":"rate limit exceeded"}` | >60 req/min |
| 500 | `{"error":"internal error"}` | DB/Redis unavailable |

---

### 2.2 POST /api/v1/dispatch/suggest — Ranked Vehicle+Driver Suggestions (NEW)

**Purpose:** Get top-N ranked vehicle+driver matches for a pending shipment.

**Request:**
```http
POST /api/v1/dispatch/suggest
Authorization: Bearer <JWT>
Content-Type: application/json
```

```json
{
  "shipment_id": "shp-uuid-001",
  "pickup_latitude": 28.5355,
  "pickup_longitude": 77.2910,
  "vehicle_type": "LCV",
  "cargo_weight_kg": 250.0,
  "top_n": 3
}
```

**Request Body Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `shipment_id` | UUID | Yes | Shipment to find vehicle+driver for |
| `pickup_latitude` | float64 | No | Pickup location lat (fallback from shipment) |
| `pickup_longitude` | float64 | No | Pickup location lng (fallback from shipment) |
| `vehicle_type` | string | No | Required vehicle type filter |
| `cargo_weight_kg` | float64 | No | Cargo weight for capacity filtering |
| `top_n` | int | No | Number of suggestions to return | 3 |

**Response (200 OK):**
```json
{
  "suggestions": [
    {
      "rank": 1,
      "vehicle": {
        "id": "v-uuid-123",
        "registration": "DL01AB1234",
        "type": "Tata Ace Gold",
        "capacity_tonnage": 1.5
      },
      "driver": {
        "id": "d-uuid-456",
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

**Scoring Algorithm:**

```
score = proximityScore × 0.4 + fitScore × 0.3 + driverScore × 0.2 + 10.0

proximityScore = max(0, 100 - distKm × 5)       // Closer = better, 0 at 20km
fitScore       = 100.0                           // Default if no cargo info
driverScore    = driver.Rating × 20              // 0-5 rating → 0-100
```

**Performance Budget:**

| Stage | Target | Strategy |
|-------|--------|----------|
| Fetch available vehicles | <100ms | Indexed query on status + org_id |
| Redis position lookups | <50ms | Batch pipeline on `live:loc:{id}` keys |
| Driver availability checks | <100ms | Check `current_trip_id IS NULL` |
| HOS validation | <50ms | Check latest duty status log |
| Scoring + sorting | <10ms | In-memory sort |
| **Total P95** | **<500ms** | Well within budget |

---

### 2.3 POST /api/v1/dispatch/assign — Create Trip from Shipment (NEW)

**Purpose:** The core dispatch action — create a trip, link it to a shipment, and publish events.

**Request:**
```http
POST /api/v1/dispatch/assign
Authorization: Bearer <JWT>
Content-Type: application/json
```

```json
{
  "shipment_id": "shp-uuid-001",
  "vehicle_id": "v-uuid-123",
  "driver_id": "d-uuid-456",
  "scheduled_start": "2026-04-07T09:00:00Z",
  "auto_start": true,
  "auto_complete": true,
  "notes": "Morning delivery, customer prefers contactless"
}
```

**Request Body Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `shipment_id` | UUID | Yes | Shipment to dispatch |
| `vehicle_id` | UUID | Yes | Vehicle to assign |
| `driver_id` | UUID | Yes | Driver to assign |
| `scheduled_start` | ISO 8601 | No | Planned start time | Now + 15min |
| `auto_start` | bool | No | Enable GPS-triggered auto-start | true |
| `auto_complete` | bool | No | Enable GPS-triggered auto-complete | true |
| `notes` | string | No | Dispatcher notes | — |

**Response (201 Created):**
```json
{
  "trip_id": "trip-uuid-001",
  "trip_number": "TRIP-2026-00123",
  "status": "SCHEDULED",
  "shipment_status": "ASSIGNED",
  "message": "Trip created and linked to shipment"
}
```

**Transaction Flow:**

```
1. Fetch shipment from customer repo
2. Validate shipment status is CONFIRMED (not already ASSIGNED)
3. Validate vehicle availability (not on active trip, status=ACTIVE)
4. Validate driver availability (current_trip_id IS NULL, HOS within limits)
5. Check for overlapping trips (tripRepo.CheckOverlappingTrips)
6. Build CreateTripRequest from shipment data
7. Call tripService.Create(ctx, orgID, req) — REAL trip creation
8. In single DB transaction:
   a. Update shipment.trip_id = trip.id
   b. Update shipment.status = 'ASSIGNED'
   c. Insert dispatch_actions audit record
   d. Update driver.current_trip_id = trip.id
   e. Update vehicle.current_driver_id = driver.id
9. Publish NATS events (best-effort, outside transaction):
   a. trip.assigned
   b. shipment.assigned
10. Return trip_id + trip_number
```

**Performance Budget:**

| Stage | Target | Strategy |
|-------|--------|----------|
| Shipment fetch | <50ms | Single SELECT by ID |
| Availability checks | <100ms | 3 queries: vehicle, driver, overlap |
| Trip creation | <300ms | tripService.Create with full validation |
| DB transaction (5 writes) | <200ms | Single tx, all indexed |
| NATS publish (2 events) | <20ms | Fire-and-forget |
| **Total P95** | **<1s** | Within budget |

---

### 2.4 POST /api/v1/dispatch/redispatch — Reassign Failed Trip (NEW)

**Purpose:** Cancel a failed trip and create a new one, re-linking the shipment.

**Request:**
```json
{
  "trip_id": "trip-uuid-002",
  "new_vehicle_id": "v-uuid-789",
  "new_driver_id": "d-uuid-789",
  "reason": "Vehicle breakdown"
}
```

**Response (200 OK):**
```json
{
  "old_trip_id": "trip-uuid-002",
  "old_trip_status": "CANCELLED",
  "new_trip_id": "trip-uuid-003",
  "new_trip_number": "TRIP-2026-00124",
  "shipment_id": "shp-uuid-001",
  "message": "Trip reassigned successfully"
}
```

**Flow:**
```
1. Fetch old trip — verify it's in a terminal state (FAILED) or can be cancelled
2. Cancel old trip via tripService.Cancel(ctx, tripID, reason)
3. Find new vehicle+driver (reuse suggestion engine if none provided)
4. Validate new vehicle+driver availability
5. Create new trip via tripService.Create()
6. Re-link shipment to new trip (shipment.trip_id = newTrip.id)
7. Insert dispatch_actions audit record (action_type=redispatch)
8. Publish NATS events: trip.reassigned, shipment.assigned
```

---

### 2.5 POST /api/v1/dispatch/bulk-assign — Batch Dispatch (P2)

**Purpose:** Dispatch multiple shipments in one request.

**Request:**
```json
{
  "assignments": [
    {"shipment_id": "shp-uuid-001", "vehicle_id": "v-uuid-123", "driver_id": "d-uuid-456"},
    {"shipment_id": "shp-uuid-002", "vehicle_id": "v-uuid-456", "driver_id": "d-uuid-789"}
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
    {"shipment_id": "shp-uuid-001", "trip_id": "trip-uuid-001", "status": "created", "error": ""},
    {"shipment_id": "shp-uuid-002", "trip_id": "trip-uuid-002", "status": "created", "error": ""}
  ]
}
```

**Implementation:** Parallel goroutines, one DB transaction per trip. Aggregate results. **No single transaction across all trips** — partial failures are expected and reported.

**Performance Budget:** 50 trips in <10s P95. Each trip <1s, parallelized across worker pool (max 10 concurrent).

---

## 3. Domain Impact Analysis

### 3.1 NEW: dispatch/ Domain

The dispatch domain does not exist and must be created. It orchestrates cross-domain interactions — it does NOT own data, it coordinates trip, vehicle, driver, and customer domains.

**Directory structure:**

```
internal/domains/dispatch/
├── engine/
│   └── suggestion_engine.go       # Proximity + availability + type matching scorer
├── handlers/
│   ├── dispatch_handler.go        # Dashboard, suggest, assign, redispatch
│   └── bulk_dispatch_handler.go   # Batch assignment (P2)
├── service/
│   ├── dispatch_service.go        # Orchestrate assign/redispatch transactions
│   ├── availability_service.go    # Vehicle + driver availability checks
│   └── scoring_service.go         # Weighted scoring (40/30/20/10)
├── repository/
│   └── dispatch_repository.go     # dispatch_actions, dispatch_suggestions tables
├── models/
│   ├── dispatch.go                # DispatchSuggestion, DispatchAction, DashboardResponse
│   └── scoring.go                 # Scoring weights, thresholds
├── worker/
│   └── auto_status_worker.go      # GPS-triggered status transitions
└── metrics.go                     # Prometheus metrics
```

#### 3.1.1 `engine/suggestion_engine.go`

**NEW file.** Core suggestion algorithm.

```go
type SuggestionEngine struct {
    vehicleRepo   *vehicle.Repository
    driverRepo    *driver.Repository
    tripRepo      *trip.Repository
    liveCache     *tracking.LiveCacheRepository  // Redis positions
    scoringSvc    *service.ScoringService
    logger        *zap.Logger
}

type SuggestionRequest struct {
    OrgID        uuid.UUID
    ShipmentID   uuid.UUID
    PickupLat    float64
    PickupLng    float64
    VehicleType  string
    CargoWeight  float64
    TopN         int
}

type Suggestion struct {
    Rank               int                `json:"rank"`
    Vehicle            VehicleSummary     `json:"vehicle"`
    Driver             *DriverSummary     `json:"driver"`
    Score              float64            `json:"score"`
    DistanceToPickupKm float64            `json:"distance_to_pickup_km"`
    ETAToPickupMin     int                `json:"eta_to_pickup_min"`
    Reasons            []string           `json:"reasons"`
}

func (e *SuggestionEngine) Suggest(ctx context.Context, req SuggestionRequest) ([]Suggestion, error) {
    // 1. Get all ACTIVE vehicles for org (actual method: GetByOrgID, not ListByOrg)
    vehicles, _, err := e.vehicleRepo.GetByOrgID(ctx, req.OrgID, 500, 0)

    // 2. Filter out vehicles on active trips
    vehicles = e.filterBusyVehicles(ctx, vehicles)

    // 3. Filter by vehicle type match (if specified)
    if req.VehicleType != "" {
        vehicles = e.filterByType(vehicles, req.VehicleType)
    }

    // 4. Filter by capacity (if cargo weight specified)
    if req.CargoWeight > 0 {
        vehicles = e.filterByCapacity(vehicles, req.CargoWeight)
    }

    // 5. For each vehicle, find assigned driver and check availability
    var suggestions []Suggestion
    for _, v := range vehicles {
        driver, err := e.driverRepo.GetByVehicleAssignment(ctx, v.ID)
        if err != nil {
            continue // No assigned driver — skip or include vehicle-only
        }

        // Check driver not on active trip
        if driver.CurrentTripID != nil && *driver.CurrentTripID != "" {
            continue
        }

        // Get vehicle position from Redis
        // NOTE: Vehicle.ID is google/uuid.UUID, liveCache expects gofrs/uuid.UUID
        // Parse via string conversion to bridge the two UUID packages
        pos, _ := e.liveCache.GetLiveLocation(ctx, uuidParse(v.ID.String()))
        dist := 0.0
        if pos != nil {
            dist = haversineKm(pos.Lat, pos.Lng, req.PickupLat, req.PickupLng)
        }

        score := e.scoringSvc.Calculate(dist, v, driver)
        eta := estimateETA(dist)

        suggestions = append(suggestions, Suggestion{
            Vehicle:            toVehicleSummary(v),
            Driver:             toDriverSummary(driver),
            Score:              score,
            DistanceToPickupKm: dist,
            ETAToPickupMin:     eta,
            Reasons:            e.scoringSvc.Explain(dist, v, driver),
        })
    }

    // 6. Sort by score desc, return top N
    sort.Slice(suggestions, func(i, j int) bool {
        return suggestions[i].Score > suggestions[j].Score
    })
    if req.TopN > 0 && len(suggestions) > req.TopN {
        suggestions = suggestions[:req.TopN]
    }
    for i := range suggestions {
        suggestions[i].Rank = i + 1
    }

    return suggestions, nil
}
```

#### 3.1.2 `handlers/dispatch_handler.go`

**NEW file.** HTTP handler for all dispatch endpoints.

```go
type DispatchHandler struct {
    dispatchSvc *service.DispatchService
    suggestEng  *engine.SuggestionEngine
    logger      *zap.Logger
}

func (h *DispatchHandler) RegisterRoutes(r chi.Router) {
    r.Get("/dashboard", h.GetDashboard)
    r.Post("/suggest", h.GetSuggestions)
    r.Post("/assign", h.Assign)
    r.Post("/redispatch", h.Redispatch)
}

func (h *DispatchHandler) GetDashboard(w http.ResponseWriter, r *http.Request) { ... }
func (h *DispatchHandler) GetSuggestions(w http.ResponseWriter, r *http.Request) { ... }
func (h *DispatchHandler) Assign(w http.ResponseWriter, r *http.Request) { ... }
func (h *DispatchHandler) Redispatch(w http.ResponseWriter, r *http.Request) { ... }
```

#### 3.1.3 `service/dispatch_service.go`

**NEW file.** Orchestrates the assign/redispatch transaction flow.

```go
type DispatchService struct {
    tripService      *trip.TripService
    vehicleRepo      *vehicle.Repository
    driverRepo       *driver.Repository
    customerRepo     *customer.CustomerV2Repository
    dispatchRepo     *repository.DispatchRepository
    eventBus         *messaging.EventBus
    logger           *zap.Logger
}

type AssignRequest struct {
    ShipmentID     uuid.UUID `json:"shipment_id"`
    VehicleID      uuid.UUID `json:"vehicle_id"`
    DriverID       uuid.UUID `json:"driver_id"`
    ScheduledStart time.Time `json:"scheduled_start"`
    AutoStart      bool      `json:"auto_start"`
    AutoComplete   bool      `json:"auto_complete"`
    Notes          string    `json:"notes"`
}

type AssignResult struct {
    TripID         string `json:"trip_id"`
    TripNumber     string `json:"trip_number"`
    Status         string `json:"status"`
    ShipmentStatus string `json:"shipment_status"`
}

func (s *DispatchService) Assign(ctx context.Context, orgID uuid.UUID, req AssignRequest) (*AssignResult, error) {
    // 1. Fetch shipment
    shipment, err := s.customerRepo.GetShipmentByID(ctx, req.ShipmentID)
    if shipment.OrgID != orgID {
        return nil, ErrTenantMismatch
    }
    if shipment.Status != "CONFIRMED" {
        return nil, ErrShipmentNotConfirmable
    }

    // 2. Validate vehicle availability
    vehicle, err := s.vehicleRepo.GetByID(ctx, req.VehicleID.String())
    if vehicle.Status != "ACTIVE" {
        return nil, ErrVehicleNotAvailable
    }
    // Check not on active trip
    activeTrips, _ := s.tripRepo.GetActiveTripsForVehicle(ctx, req.VehicleID.String())
    if len(activeTrips) > 0 {
        return nil, ErrVehicleOnActiveTrip
    }

    // 3. Validate driver availability
    driver, err := s.driverRepo.GetByID(ctx, req.DriverID.String())
    if driver.CurrentTripID != nil && *driver.CurrentTripID != "" {
        return nil, ErrDriverOnActiveTrip
    }

    // 4. Check overlapping trips (actual sig: ctx, vehicleID, driverID string, start, end time.Time)
    overlaps, _ := s.tripRepo.CheckOverlappingTrips(ctx, req.VehicleID.String(), req.DriverID.String(),
        req.ScheduledStart, req.ScheduledStart.Add(time.Duration(req.EstimatedDurationMin)*time.Minute))
    if overlaps {
        return nil, ErrOverlappingTrips
    }

    // 5. Build trip request from shipment data
    tripReq := s.buildTripRequest(shipment, req)

    // 6. Create trip
    trip, err := s.tripService.Create(ctx, orgID.String(), tripReq)

    // 7. Transaction: link shipment + audit + driver/vehicle updates
    result, err := s.dispatchRepo.ExecuteAssignment(ctx, &repository.AssignmentTx{
        ShipmentID: req.ShipmentID,
        TripID:     trip.ID,
        VehicleID:  req.VehicleID,
        DriverID:   req.DriverID,
        OrgID:      orgID,
    })

    // 8. Publish NATS events (best-effort, outside tx)
    // NOTE: EventBus.Publish requires 4 params: (ctx, subject, data, orgID)
    if s.eventBus != nil {
        _ = s.eventBus.Publish(ctx, "trip.assigned", map[string]interface{}{
            "trip_id":     trip.ID,
            "vehicle_id":  req.VehicleID.String(),
            "driver_id":   req.DriverID.String(),
            "shipment_id": req.ShipmentID.String(),
            "org_id":      orgID.String(),
        }, orgID.String())
        _ = s.eventBus.Publish(ctx, "shipment.assigned", map[string]interface{}{
            "shipment_id": req.ShipmentID.String(),
            "trip_id":     trip.ID,
            "vehicle_id":  req.VehicleID.String(),
            "driver_id":   req.DriverID.String(),
            "org_id":      orgID.String(),
        }, orgID.String())
    }

    return result, nil
}
```

#### 3.1.4 `service/availability_service.go`

**NEW file.** Vehicle and driver availability checks.

```go
type AvailabilityService struct {
    vehicleRepo *vehicle.Repository
    driverRepo  *driver.Repository
    tripRepo    *trip.Repository
    liveCache   *tracking.LiveCacheRepository
}

func (s *AvailabilityService) IsVehicleAvailable(ctx context.Context, vehicleID uuid.UUID) (bool, error) {
    v, err := s.vehicleRepo.GetByID(ctx, vehicleID.String())
    if err != nil { return false, err }
    if v.Status != "ACTIVE" { return false, nil }

    activeTrips, err := s.tripRepo.GetActiveTripsForVehicle(ctx, vehicleID.String())
    if err != nil { return false, err }
    return len(activeTrips) == 0, nil
}

func (s *AvailabilityService) IsDriverAvailable(ctx context.Context, driverID uuid.UUID) (bool, error) {
    d, err := s.driverRepo.GetByID(ctx, driverID.String())
    if err != nil { return false, err }
    if d.Status != "active" { return false, nil }
    if d.CurrentTripID != nil && *d.CurrentTripID != "" { return false, nil }

    // Check HOS — if in DRIVING status for >4 hours, not available
    return s.isWithinHOSLimits(ctx, driverID), nil
}

func (s *AvailabilityService) GetAvailableVehicles(ctx context.Context, orgID uuid.UUID) ([]vehicle.Vehicle, error) {
    // Actual method: GetByOrgID, not ListByOrg
    vehicles, _, err := s.vehicleRepo.GetByOrgID(ctx, orgID, 500, 0)
    if err != nil { return nil, err }

    var available []vehicle.Vehicle
    for _, v := range vehicles {
        isAvail, _ := s.IsVehicleAvailable(ctx, v.ID)
        if isAvail {
            available = append(available, *v)
        }
    }
    return available, nil
}
```

#### 3.1.5 `service/scoring_service.go`

**NEW file.** Weighted scoring implementation.

```go
type ScoringService struct{}

func (s *ScoringService) Calculate(distKm float64, v *vehicle.Vehicle, d *driver.Driver) float64 {
    // Proximity: 100 - (distKm × 5), min 0
    proximity := math.Max(0, 100-distKm*5)

    // Vehicle fit: 100 by default (could be enhanced with cargo type matching)
    fit := 100.0

    // Driver rating: 0-5 → 0-100
    driverRating := d.Rating * 20.0

    // Weighted: 40% proximity + 30% fit + 20% driver + 10% base
    return proximity*0.4 + fit*0.3 + driverRating*0.2 + 10.0
}

func (s *ScoringService) Explain(distKm float64, v *vehicle.Vehicle, d *driver.Driver) []string {
    var reasons []string
    if distKm < 5 {
        reasons = append(reasons, "Closest available vehicle")
    } else if distKm < 15 {
        reasons = append(reasons, "Within reasonable distance")
    }
    if d.Rating >= 4.0 {
        reasons = append(reasons, fmt.Sprintf("High driver rating (%.1f)", d.Rating))
    }
    return reasons
}
```

#### 3.1.6 `repository/dispatch_repository.go`

**NEW file.** Data access for dispatch_actions and dispatch_suggestions tables.

```go
type DispatchRepository struct {
    pool *pgxpool.Pool
}

type AssignmentTx struct {
    ShipmentID uuid.UUID
    TripID     string
    VehicleID  uuid.UUID
    DriverID   uuid.UUID
    OrgID      uuid.UUID
}

func (r *DispatchRepository) ExecuteAssignment(ctx context.Context, tx *AssignmentTx) (*AssignResult, error) {
    pgxTx, err := r.pool.Begin(ctx)
    if err != nil { return nil, err }
    defer pgxTx.Rollback(ctx)

    // a. Update shipment trip_id and status
    _, err = pgxTx.Exec(ctx,
        `UPDATE customer_v2_shipments SET trip_id = $1, status = 'ASSIGNED', updated_at = NOW()
         WHERE id = $2 AND org_id = $3`,
        tx.TripID, tx.ShipmentID, tx.OrgID)
    if err != nil { return nil, fmt.Errorf("update shipment: %w", err) }

    // b. Insert dispatch_actions audit record
    _, err = pgxTx.Exec(ctx,
        `INSERT INTO dispatch_actions (org_id, action_type, shipment_id, trip_id, vehicle_id, driver_id, performed_by)
         VALUES ($1, 'manual_assign', $2, $3, $4, $5, $6)`,
        tx.OrgID, tx.ShipmentID, tx.TripID, tx.VehicleID, tx.DriverID, "system")
    if err != nil { return nil, fmt.Errorf("insert dispatch_actions: %w", err) }

    // c. Update driver current_trip_id
    _, err = pgxTx.Exec(ctx,
        `UPDATE drivers SET current_trip_id = $1::text WHERE id = $2`,
        tx.TripID, tx.DriverID)

    // d. Update vehicle current_driver_id
    _, err = pgxTx.Exec(ctx,
        `UPDATE vehicles SET current_driver_id = $1 WHERE id = $2`,
        tx.DriverID, tx.VehicleID)

    if err := pgxTx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("commit assignment: %w", err)
    }

    return &AssignResult{
        TripID:         tx.TripID,
        Status:         "SCHEDULED",
        ShipmentStatus: "ASSIGNED",
    }, nil
}
```

#### 3.1.7 `models/dispatch.go`

**NEW file.** Domain models.

```go
type DashboardResponse struct {
    Summary          SummaryCounts       `json:"summary"`
    PendingShipments []PendingShipment    `json:"pending_shipments"`
    ActiveTrips      []ActiveTripSummary  `json:"active_trips"`
    AvailableVehicles []VehicleSummary    `json:"available_vehicles"`
    AvailableDrivers []DriverSummary      `json:"available_drivers"`
    Alerts           []DispatchAlert      `json:"alerts"`
}

type SummaryCounts struct {
    TotalToday   int     `json:"total_today"`
    Pending      int     `json:"pending"`
    InTransit    int     `json:"in_transit"`
    Delivered    int     `json:"delivered"`
    Failed       int     `json:"failed"`
    Cancelled    int     `json:"cancelled"`
    OnTimeRate   float64 `json:"on_time_rate"`
    AvgDelayMin  int     `json:"avg_delay_min"`
}

type DispatchActionType string

const (
    ActionAutoSuggest  DispatchActionType = "auto_suggest"
    ActionManualAssign DispatchActionType = "manual_assign"
    ActionAutoAssign   DispatchActionType = "auto_assign"
    ActionBulkAssign   DispatchActionType = "bulk_assign"
    ActionRedispatch   DispatchActionType = "redispatch"
    ActionOverride     DispatchActionType = "override"
)
```

#### 3.1.8 `worker/auto_status_worker.go`

**NEW file.** Background worker for GPS-triggered status changes.

```go
type AutoStatusWorker struct {
    tripRepo    *trip.Repository
    tripSvc     *trip.TripService
    liveCache   *tracking.LiveCacheRepository
    geofenceRepo *geofencing.Repository
    eventBus    *messaging.EventBus
    logger      *zap.Logger
    interval    time.Duration
    stopCh      chan struct{}
}

func NewAutoStatusWorker(tripRepo, tripSvc, liveCache, geofenceRepo, eventBus, logger, interval) *AutoStatusWorker {
    return &AutoStatusWorker{...}
}

func (w *AutoStatusWorker) Start() {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            w.processTrips()
        case <-w.stopCh:
            return
        }
    }
}

func (w *AutoStatusWorker) processTrips() {
    ctx := context.Background()
    // 1. Fetch all STARTED trips (using List with status filter — GetTripsByStatus doesn't exist)
    trips, _, err := w.tripRepo.List(ctx, tripRepository.TripListFilter{Status: "STARTED", Limit: 200})
    if err != nil { return }

    for _, trip := range trips {
        // 2. Get vehicle GPS position
        // NOTE: trip.VehicleID is string, liveCache.GetLiveLocation expects uuid.UUID
        vehicleUUID, err := uuid.Parse(trip.VehicleID)
        if err != nil { continue }
        pos, err := w.liveCache.GetLiveLocation(ctx, vehicleUUID)
        if err != nil || pos == nil { continue }

        // 3. Check if vehicle is at drop location (within 500m geofence)
        if pos.Speed == 0 && w.isAtDropLocation(ctx, trip, pos) {
            // Check if stopped for >2 minutes
            if time.Since(pos.LastSeen) > 2*time.Minute {
                // 4. Auto-complete trip
                // NOTE: CompleteTripRequest uses OdometerReading (not EndOdometer)
                // and PODDocuments []PODDocument (not PODData)
                driverID := ""
                if trip.DriverID != nil {
                    driverID = *trip.DriverID
                }
                err := w.tripSvc.Complete(ctx, trip.ID, driverID, &tripModels.CompleteTripRequest{
                    OdometerReading: pos.Odometer,
                    PODDocuments: []tripModels.PODDocument{
                        {
                            Type:      "gps_auto_capture",
                            Timestamp: time.Now(),
                        },
                    },
                })
                if err != nil {
                    w.logger.Warn("auto-complete failed", zap.String("trip_id", trip.ID), zap.Error(err))
                    continue
                }
                w.logger.Info("auto-completed trip", zap.String("trip_id", trip.ID))
                // NOTE: Trip model has NO ShipmentID field — shipment linked via customer_v2_shipments.trip_id.
                // To publish shipment.delivered, we'd need to query the shipment by trip_id first.
                // For P0, publish trip.completed instead and let customer domain resolve shipment separately.
                if w.eventBus != nil {
                    _ = w.eventBus.Publish(ctx, "trip.completed", map[string]interface{}{
                        "trip_id":    trip.ID,
                        "vehicle_id": trip.VehicleID,
                        "org_id":     trip.OrgID,
                        "end_time":   time.Now(),
                    }, trip.OrgID)
                }
            }
        }
    }
}

func (w *AutoStatusWorker) isAtDropLocation(ctx context.Context, trip *tripModels.Trip, pos *tracking.LiveLocation) bool {
    if len(trip.PlannedRoute) == 0 { return false }
    lastWaypoint := trip.PlannedRoute[len(trip.PlannedRoute)-1]
    // Waypoint.Location has Latitude/Longitude, not Waypoint.Lat/Lng
    dist := haversineKm(pos.Lat, pos.Lng, lastWaypoint.Location.Latitude, lastWaypoint.Location.Longitude)
    return dist < 0.5 // 500m threshold
}
```

---

### 3.2 EXISTING Domains to Modify

#### 3.2.1 `customer/service/shipment_assignment_service.go` — FIX STUB

**Current state (verified):**
```go
// Line 44: mockTripID := uuid.New() — creates fake UUID
// Line 59: // s.eventBus.Publish("shipment.assigned", shipment) — commented out
// No TripService dependency
// No EventBus dependency
```

**Fixed implementation:**

```go
package service

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "go.uber.org/zap"

    "github.com/flyfleet/fleet-core/internal/domains/customer/models"
    "github.com/flyfleet/fleet-core/internal/domains/customer/repository"
    "github.com/flyfleet/fleet-core/internal/domains/trip/models" as tripModels
    "github.com/flyfleet/fleet-core/internal/domains/trip/service" as tripService
    "github.com/flyfleet/fleet-core/internal/messaging"
)

type ShipmentAssignmentService struct {
    repo       repository.CustomerV2Repository
    tripSvc    *tripService.TripService    // NEW: injected
    eventBus   *messaging.EventBus          // NEW: injected
    logger     *zap.Logger
}

func NewShipmentAssignmentService(
    repo repository.CustomerV2Repository,
    tripSvc *tripService.TripService,
    eventBus *messaging.EventBus,
    logger *zap.Logger,
) *ShipmentAssignmentService {
    return &ShipmentAssignmentService{
        repo: repo, tripSvc: tripSvc, eventBus: eventBus, logger: logger,
    }
}

type AssignmentResult struct {
    ShipmentID uuid.UUID `json:"shipment_id"`
    TripID     string    `json:"trip_id"`
    TripNumber string    `json:"trip_number"`
}

func (s *ShipmentAssignmentService) AssignShipment(
    ctx context.Context,
    shipmentID uuid.UUID,
    vehicleID uuid.UUID,
    driverID uuid.UUID,
    orgID uuid.UUID,
) (*AssignmentResult, error) {
    // 1. Fetch shipment
    shipment, err := s.repo.GetShipmentByID(ctx, shipmentID)
    if err != nil {
        return nil, fmt.Errorf("get shipment: %w", err)
    }
    // NOTE: Shipment struct has NO OrgID field. Org is inferred from CustomerID
    // or validated at the handler level via TenantContextMiddleware.
    if shipment.Status != "CONFIRMED" {
        return nil, fmt.Errorf("shipment status %s is not CONFIRMED", shipment.Status)
    }

    // 2. Fetch booking for cargo/pickup details
    if shipment.BookingID == nil {
        return nil, fmt.Errorf("shipment has no booking ID")
    }
    booking, err := s.repo.GetBookingByID(ctx, *shipment.BookingID)
    if err != nil {
        return nil, fmt.Errorf("get booking: %w", err)
    }

    // 3. Parse Origin/Destination from JSON blobs ([]byte → Location struct)
    var originLoc, destLoc tripModels.Location
    if shipment.Origin != nil {
        json.Unmarshal(shipment.Origin, &originLoc)
    }
    if shipment.Destination != nil {
        json.Unmarshal(shipment.Destination, &destLoc)
    }

    // 4. Build CreateTripRequest from shipment/booking data
    distKm := 0.0
    if shipment.DistanceKm != nil {
        distKm = *shipment.DistanceKm
    }
    title := "Delivery"
    if originLoc.Address != "" && destLoc.Address != "" {
        title = fmt.Sprintf("Delivery %s → %s", originLoc.Address, destLoc.Address)
    }

    cargoType := ""
    if booking.CargoType != nil {
        cargoType = *booking.CargoType
    }
    cargoWeight := 0.0
    if booking.Weight != nil {
        cargoWeight = *booking.Weight
    }

    tripReq := &tripModels.CreateTripRequest{
        VehicleID: vehicleID.String(),
        DriverID:  driverID.String(),
        Title:     title,
        PlannedDistance:      distKm,
        EstimatedDurationMin: estimateDuration(distKm),
        CargoDetails: tripModels.CargoDetails{
            Type:     cargoType,
            WeightKg: cargoWeight,
        },
        ScheduledStart: time.Now().Add(15 * time.Minute),
        CustomerName:   booking.CustomerName,
        CustomerPhone:  booking.CustomerPhone,
    }
    if originLoc.Latitude != 0 || originLoc.Longitude != 0 {
        tripReq.PickupGeo = &originLoc
    }
    if destLoc.Latitude != 0 || destLoc.Longitude != 0 {
        tripReq.DropGeo = &destLoc
    }

    // Add waypoints from shipment stops
    stops, _ := s.repo.GetShipmentStops(ctx, shipmentID)
    for _, stop := range stops {
        var loc tripModels.Location
        if stop.Location != nil {
            json.Unmarshal(stop.Location, &loc)
        }
        tripReq.PlannedRoute = append(tripReq.PlannedRoute, tripModels.Waypoint{
            Type:     stop.StopType,   // "PICKUP", "DROPOFF", "CHECKPOINT"
            Sequence: stop.Sequence,
            Location: loc,
            Status:   "pending",
        })
    }
    if len(tripReq.PlannedRoute) < 2 {
        // Minimum 2 waypoints required — create from origin/destination
        if originLoc.Latitude != 0 {
            tripReq.PlannedRoute = append([]tripModels.Waypoint{{
                Type:     "pickup",
                Sequence: 0,
                Location: originLoc,
                Status:   "pending",
            }}, tripReq.PlannedRoute...)
        }
        if destLoc.Latitude != 0 {
            tripReq.PlannedRoute = append(tripReq.PlannedRoute, tripModels.Waypoint{
                Type:     "dropoff",
                Sequence: len(tripReq.PlannedRoute) + 1,
                Location: destLoc,
                Status:   "pending",
            })
        }
    }

    // 4. Create REAL trip via trip service
    trip, err := s.tripSvc.Create(ctx, orgID.String(), tripReq)
    if err != nil {
        return nil, fmt.Errorf("create trip: %w", err)
    }

    // 5. Single DB transaction: link shipment to trip
    tx, err := s.repo.BeginTx(ctx)
    if err != nil {
        return nil, fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback(ctx)

    if err := s.repo.UpdateShipmentTripID(ctx, tx, shipmentID, trip.ID); err != nil {
        return nil, fmt.Errorf("update shipment trip_id: %w", err)
    }
    if err := s.repo.UpdateShipmentStatus(ctx, tx, shipmentID, "ASSIGNED"); err != nil {
        return nil, fmt.Errorf("update shipment status: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("commit: %w", err)
    }

    // 6. Publish NATS event (best-effort, outside tx)
    // NOTE: EventBus.Publish requires 4 params: (ctx, subject, data, orgID)
    if s.eventBus != nil {
        _ = s.eventBus.Publish(ctx, "shipment.assigned", map[string]interface{}{
            "shipment_id": shipmentID.String(),
            "trip_id":     trip.ID,
            "vehicle_id":  vehicleID.String(),
            "driver_id":   driverID.String(),
            "org_id":      orgID.String(),
        }, orgID.String())
    }

    s.logger.Info("shipment assigned successfully",
        zap.String("shipment_id", shipmentID.String()),
        zap.String("trip_id", trip.ID))
    // NOTE: Trip model has NO TripNumber field. Trip.ID is the unique identifier.

    return &AssignmentResult{
        ShipmentID: shipmentID,
        TripID:     trip.ID,
    }, nil
}
```

#### 3.2.2 `customer/` — Add EventBus Dependency Injection

**All customer service constructors that need event publishing must accept `EventBus`:**

- `NewShipmentAssignmentService(repo, tripSvc, eventBus, logger)` — shown above
- `NewDeliveryVerificationService(repo, eventBus, logger)` — uncomment the Publish call

#### 3.2.3 `vehicle/` — Add Missing Methods

**NEW method on `vehicle.Repository`:**

```go
// GetByOrgID already exists — just needs a status filter variant
// GetAvailableVehicles returns active vehicles not on active trips.
// Uses existing GetByOrgID — no new method needed on vehicle repo.
func (r *DispatchRepository) GetAvailableVehicles(ctx context.Context, orgID uuid.UUID, vehicleRepo *vehicle.Repository, tripRepo *trip.Repository) ([]*vehicle.Vehicle, error) {
    vehicles, _, err := vehicleRepo.GetByOrgID(ctx, orgID, 500, 0)
    if err != nil { return nil, err }

    var available []*vehicle.Vehicle
    for _, v := range vehicles {
        if v.Status != "ACTIVE" { continue }
        activeTrips, _ := tripRepo.GetActiveTripsForVehicle(ctx, v.ID.String())
        if len(activeTrips) == 0 {
            available = append(available, v)
        }
    }
    return available, nil
}
```

> **Reality check:** `GetByOrgID(ctx, orgID, limit, offset)` already exists and accepts status filtering through the caller. No new method needed — just use the existing one with `status="ACTIVE"`.

#### 3.2.4 `driver/` — Add Missing Methods

**NEW method on `driver.Repository`:**

```go
// GetByVehicleAssignment finds the driver currently assigned to a vehicle.
// Uses the driver-vehicle assignment relationship.
func (r *DriverRepository) GetByVehicleAssignment(ctx context.Context, vehicleID uuid.UUID) (*Driver, error) {
    q := `SELECT d.id, d.name, d.phone, d.license_number, d.status, d.rating,
                 d.total_trips, d.current_trip_id
          FROM drivers d
          WHERE d.current_vehicle_id = $1
            AND d.status = 'active'
            AND d.deleted_at IS NULL`

    var d Driver
    err := r.pool.QueryRow(ctx, q, vehicleID).Scan(
        &d.ID, &d.Name, &d.Phone, &d.LicenseNumber, &d.Status,
        &d.Rating, &d.TotalTrips, &d.CurrentTripID,
    )
    if err != nil {
        return nil, err
    }
    return &d, nil
}
```

> **Reality check:** The driver repository has `List(ctx, filter)` with a vehicle filter. `GetByVehicleAssignment` is a convenience wrapper that can be added, or the existing `List` method with `VehicleID` filter can be used directly.

---

### 3.3 cmd/main.go — Wire ALL Unwired Domains

**This is the single most impactful change.** All domains are built but not mounted. The wire-up code:

```go
// ── NEW: Shared dependencies ──
eventBus, err := messaging.NewNatsEventBus(cfg.NatsURL, logger)
if err != nil {
    logger.Fatal("event bus connection failed", zap.Error(err))
}

// ── Trip Domain ──
tripRepo := tripRepository.NewTripRepository(pool.High, logger)  // Use High-priority pool
tripSvc := tripService.NewTripService(tripRepo, eventBus, logger)
tripHandler := tripHandlers.NewTripHandler(tripSvc, logger)

// ── Vehicle Domain ──
vehicleRepo := vehicleRepository.NewVehicleRepository(pool.High, logger)
vehicleSvc := vehicleService.NewVehicleService(vehicleRepo, logger)
vehicleHandler := vehicleHandlers.NewVehicleHandler(vehicleSvc, logger)

// ── Driver Domain ──
driverRepo := driverRepository.NewDriverRepository(pool.High, logger)
driverSvc := driverService.NewDriverService(driverRepo, logger)
driverHandler := driverHandlers.NewDriverHandler(driverSvc, logger)

// ── Customer Domain ──
customerRepo := customerRepository.NewCustomerV2Repository(pool.High, logger)
tripSvcForAssignment := tripSvc  // Reuse trip service
assignmentSvc := customerService.NewShipmentAssignmentService(customerRepo, tripSvcForAssignment, eventBus, logger)
customerHandler := customerHandlers.NewCustomerHandler(customerRepo, assignmentSvc, logger)

// ── Route Domain ──
routeRepo := routeRepository.NewRouteRepository(pool.Medium, logger)
routeSvc := routeService.NewRouteService(routeRepo, logger)
routeHandler := routeHandlers.NewRouteHandler(routeSvc, logger)

// ── Dispatch Domain (NEW) ──
dispatchRepo := dispatchRepository.NewDispatchRepository(pool.High)
liveCache := tracking_repository.NewLiveCacheRepository(rdb)
suggestionEngine := dispatch_engine.NewSuggestionEngine(vehicleRepo, driverRepo, tripRepo, liveCache, dispatch_scoring, logger)
availabilitySvc := dispatch_service.NewAvailabilityService(vehicleRepo, driverRepo, tripRepo, liveCache)
dispatchSvc := dispatch_service.NewDispatchService(tripSvc, vehicleRepo, driverRepo, customerRepo, dispatchRepo, eventBus, logger)
dispatchHandler := dispatch_handlers.NewDispatchHandler(dispatchSvc, suggestionEngine, logger)

// ── Auto Status Worker ──
autoStatusWorker := dispatch_worker.NewAutoStatusWorker(tripRepo, tripSvc, liveCache, routeRepo, eventBus, logger, 5*time.Second)
go autoStatusWorker.Start()

// ── Router Mounting ──
r.Group(func(r chi.Router) {
    r.Use(middleware.JWKSAuthGate(cfg.JWTSecret, logger))
    r.Use(middleware.TenantMiddleware(logger))

    api := chi.NewRouter()

    // IAM routes (already wired)
    if iamModule != nil {
        r.Route("/api/v1", func(r chi.Router) {
            iamModule.RegisterPublicRoutes(r, logger)
            iamModule.RegisterAuthRoutes(r, logger)
        })
    }

    // ── Tracking (already wired) ──
    if cfg.Flags.VehicleMapEnabled {
        api.Post("/v1/telemetry", ingestHandler.ServeHTTP)
        api.Get("/v1/vehicles/locations", locationsHandler.ServeHTTP)
        if cfg.Flags.HistoryTrail {
            api.Get("/v1/vehicles/{id}/telemetry/history", historyHandler.ServeHTTP)
        }
        if cfg.Flags.RealTimeUpdates {
            api.Get("/ws/telemetry", wsHandler.ServeHTTP)
        }
        api.Get("/v1/geofences", geofenceHandler.ServeHTTP)
        api.Post("/v1/geofences", geofenceHandler.ServeHTTP)
    }

    // ── NEW: Dispatch ──
    if cfg.Flags.DispatchEnabled {
        dispatchRouter := chi.NewRouter()
        dispatchHandler.RegisterRoutes(dispatchRouter)
        api.Route("/v1/dispatch", dispatchRouter)
    }

    // ── NEW: Trips ──
    tripRouter := chi.NewRouter()
    tripHandler.RegisterRoutes(tripRouter)
    api.Route("/v1/trips", tripRouter)

    // ── NEW: Vehicles ──
    vehicleRouter := chi.NewRouter()
    vehicleHandler.RegisterRoutes(vehicleRouter)
    api.Route("/v1/vehicles", vehicleRouter)

    // ── NEW: Drivers ──
    driverRouter := chi.NewRouter()
    driverHandler.RegisterRoutes(driverRouter)
    api.Route("/v1/drivers", driverRouter)

    // ── NEW: Customers ──
    customerRouter := chi.NewRouter()
    customerHandler.RegisterRoutes(customerRouter)
    api.Route("/v1/customers", customerRouter)

    // ── NEW: Routes (optimization) ──
    routeRouter := chi.NewRouter()
    routeHandler.RegisterRoutes(routeRouter)
    api.Route("/v1/routes", routeRouter)

    r.Mount("/api", api)
})
```

---

### 3.4 Route Domain — Missing Migrations

**Verified:** `route_optimization_requests`, `route_optimization_results`, and `route_optimization_jobs` tables exist in `migrations/068_route_optimization.sql` and `infrastructure/migrations/028_route_optimization_schema.sql`.

**NOT found in any migration:** `road_restrictions`, `vehicle_permits`. These tables are referenced in Go code but have no migration.

**Action:** Create migration `303_route_missing_tables.sql` if these tables are needed for the dispatch flow. For P0, they are NOT required — the route solvers work without them. Defer to P1.

---

## 4. Data Flow Sequences

### 4.1 Booking → Dispatch → Trip → Delivery (End-to-End)

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
   → Shipment created (BOOKED → CONFIRMED, trip_id = NULL)

4. Dispatcher opens dispatch dashboard
   GET /api/v1/dispatch/dashboard?date=2026-04-07
   → Shows pending shipments, available vehicles/drivers

5. Dispatcher requests suggestions
   POST /api/v1/dispatch/suggest {shipment_id, ...}
   → SuggestionEngine:
     a. Fetches ACTIVE vehicles for org
     b. Filters out vehicles on active trips (tripRepo.GetActiveTripsForVehicle)
     c. Filters by type + capacity match
     d. Gets GPS positions from Redis live cache (live:loc:{id})
     e. Finds assigned drivers, checks availability (current_trip_id IS NULL)
     f. Scores each match (40% proximity, 30% fit, 20% rating, 10% base)
     g. Returns top 3

6. Dispatcher accepts top suggestion
   POST /api/v1/dispatch/assign {shipment_id, vehicle_id, driver_id, ...}
   → DispatchService.Assign():
     a. Validates shipment status = CONFIRMED
     b. Validates vehicle available (not on active trip)
     c. Validates driver available (current_trip_id IS NULL, HOS OK)
     d. Checks overlapping trips
     e. Builds CreateTripRequest from shipment data
     f. tripService.Create() → creates trip in CREATED state
     g. tripService.Approve() → transitions to SCHEDULED
     h. DB transaction:
        - shipment.trip_id = trip.id, status = 'ASSIGNED'
        - dispatch_actions INSERT (audit)
        - driver.current_trip_id = trip.id
        - vehicle.current_driver_id = driver.id
     i. Publish NATS: trip.assigned, shipment.assigned

7. Trip starts (auto-detect via GPS)
   AutoStatusWorker (runs every 5s):
   a. Fetches STARTED trips
   b. Checks vehicle GPS from Redis
   c. If vehicle near pickup AND ignition ON → tripService.Start()
   → Trip: SCHEDULED → STARTED
   → Shipment: ASSIGNED → IN_TRANSIT
   → NATS: trip.started

8. Trip completes (auto-detect via GPS + geofence)
   AutoStatusWorker:
   a. Fetches STARTED trips
   b. Checks vehicle at drop location (within 500m)
   c. If speed=0 for >2min → tripService.Complete() with POD
   → Trip: STARTED → COMPLETED
   → Shipment: IN_TRANSIT → DELIVERED
   → NATS: trip.completed, shipment.delivered
   → Delivery receipt PDF generated
   → Invoice created (if configured)
```

### 4.2 Dispatch Dashboard Aggregate

```
1. Dashboard → GET /api/v1/dispatch/dashboard?date=2026-04-07
2. Check Redis cache: dispatch:dashboard_summary:{org_id} (TTL 5s)
   → If hit, return cached summary
3. Parallel queries (if cache miss):
   a. TripRepo.GetTripSummary(orgID, date) → counts by status
   b. CustomerRepo.GetPendingShipments(orgID) → CONFIRMED with trip_id=NULL
   c. TripRepo.GetTripsByStatus(orgID, ["STARTED","PAUSED","RESUMED"]) → active trips
   d. VehicleRepo.ListByOrg(orgID, status="ACTIVE") → available vehicles
      - Filter out vehicles on active trips (in-memory check)
   e. DriverRepo.List(orgID, status="active") → available drivers
      - Filter out drivers with current_trip_id (in-memory)
   f. Check delayed trips (ETA passed, still not COMPLETED)
4. Assemble DashboardResponse
5. Cache summary in Redis (5s TTL)
6. Return JSON
```

### 4.3 Auto Status Detection (Background Worker, Every 5s)

```
1. AutoStatusWorker tick (every 5s)
2. Fetch all trips with status STARTED from trip repo
3. For each trip:
   a. Get vehicle GPS from Redis live:loc:{vehicle_id}
   b. If no GPS data (offline) → skip, next tick
   c. Get drop location from trip.PlannedRoute last waypoint
   d. Calculate distance (Haversine)
   e. If distance < 500m AND speed == 0:
      - Check how long vehicle has been stationary
      - If > 2 minutes → auto-complete
4. Auto-complete flow:
   a. tripService.Complete(tripID, driverID, CompleteTripRequest{
        EndOdometer: gps.Odometer,
        PODData: {GPSLat, GPSLng, CapturedAt: now()}
   })
   b. Update shipment status → DELIVERED
   c. Publish NATS: trip.completed, shipment.delivered
   d. Push to dashboard WebSocket (if connected)
5. Log result, continue to next trip
```

### 4.4 Redispatch (Failed Trip → New Trip)

```
1. Dispatcher identifies failed trip
2. POST /api/v1/dispatch/redispatch {trip_id, new_vehicle_id, new_driver_id, reason}
3. DispatchService.Redispatch():
   a. Fetch old trip → verify status (must be FAILED or CANCELLED-able)
   b. tripService.Cancel(oldTripID, reason) → status CANCELLED
   c. Get shipment from old trip → shipment.trip_id
   d. Validate new vehicle+driver availability
   e. Create new trip via tripService.Create()
   f. tripService.Approve() → SCHEDULED
   g. DB transaction:
      - shipment.trip_id = newTrip.id
      - dispatch_actions INSERT (action_type=redispatch)
      - Update driver.current_trip_id
   h. Publish NATS: trip.reassigned, shipment.assigned
4. Return {old_trip_id, old_status, new_trip_id, new_trip_number}
```

### 4.5 VRP Results → Trip Generation

> **P2 feature.** Route optimization produces `RouteOptimizationResult` with `Routes []Route`, each having `VehicleID` and `Stops []RouteStop`. No Trip converter exists.

```
1. Dispatcher runs route optimization
   POST /api/v1/routes/requests → async job
   → RouteOptimizationResult with Routes[]

2. VRP-to-Trip Converter (NEW service):
   For each Route in result:
   a. Create CreateTripRequest:
      - VehicleID = route.VehicleID
      - PlannedRoute = map RouteStop[] to Waypoint[]
      - Waypoints = map RouteStop[] with sequence, ETA, address
   b. Find available driver for vehicle
   c. tripService.Create() → new trip
   d. Link shipments to trip (shipment.trip_id = trip.id)
   e. Publish NATS: trip.assigned, shipment.assigned

3. All trips created, dispatcher reviews and confirms
```

---

## 5. Data Model Changes

### 5.1 New Tables

**Migration:** `services/fleet-core/migrations/303_dispatch_tables.sql`

```sql
-- Dispatch action log (audit trail for all dispatch decisions)
CREATE TABLE IF NOT EXISTS dispatch_actions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    action_type     VARCHAR(30) NOT NULL CHECK (action_type IN (
        'auto_suggest', 'manual_assign', 'auto_assign',
        'bulk_assign', 'redispatch', 'override'
    )),
    shipment_id     UUID REFERENCES customer_v2_shipments(id),
    trip_id         VARCHAR(255),             -- trips.id is VARCHAR in trip model
    vehicle_id      UUID REFERENCES vehicles(id),
    driver_id       UUID REFERENCES drivers(id),
    suggestion_data JSONB,                    -- Top 3 suggestions with scores
    decision_reason TEXT,
    performed_by    UUID,                     -- user who performed action
    performed_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dispatch_actions_org ON dispatch_actions (org_id, performed_at DESC);
CREATE INDEX IF NOT EXISTS idx_dispatch_actions_shipment ON dispatch_actions (shipment_id);
CREATE INDEX IF NOT EXISTS idx_dispatch_actions_trip ON dispatch_actions (trip_id);

-- Dispatch suggestions cache (short-lived, for dashboard)
CREATE TABLE IF NOT EXISTS dispatch_suggestions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    shipment_id     UUID NOT NULL,
    suggestion_rank INT NOT NULL,
    vehicle_id      UUID REFERENCES vehicles(id),
    driver_id       UUID REFERENCES drivers(id),
    score           DOUBLE PRECISION,
    distance_km     DOUBLE PRECISION,
    eta_min         INT,
    reasons         TEXT[],
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    expires_at      TIMESTAMPTZ DEFAULT NOW() + INTERVAL '5 minutes'
);

CREATE INDEX IF NOT EXISTS idx_dispatch_suggestions_shipment ON dispatch_suggestions (shipment_id, expires_at DESC);
```

### 5.2 Redis Key Additions

| Key Pattern | Type | TTL | Purpose |
|-------------|------|-----|---------|
| `dispatch:dashboard_summary:{org_id}` | JSON string | 5s | Cached dashboard aggregate summary counts |
| `dispatch:suggestion:{shipment_id}` | JSON string | 5min | Pre-computed suggestions (cache) |
| `vehicle:{id}:on_active_trip` | Boolean (string "true"/"false") | Trip duration | Quick availability check (set on assign, clear on complete) |
| `driver:{id}:on_active_trip` | Boolean (string "true"/"false") | Trip duration | Quick availability check (set on assign, clear on complete) |

### 5.3 New NATS Event Subjects

| Subject | Publisher | Consumer | Payload |
|---------|-----------|----------|---------|
| `trip.assigned` | DispatchService | Alerts, Analytics | `{trip_id, vehicle_id, driver_id, shipment_id, org_id}` |
| `trip.reassigned` | DispatchService | Alerts, Analytics | `{old_trip_id, new_trip_id, reason, org_id}` |
| `shipment.assigned` | DispatchService | Customer portal | `{shipment_id, trip_id, vehicle_id, driver_id}` |
| `shipment.delivered` | AutoStatusWorker / TripService | Customer, Billing | `{shipment_id, trip_id, delivered_at, pod}` |

**Existing NATS events (reused, not new):**
- `trip.completed` — Already published by TripService.Complete()
- `trip.started` — Already published by TripService.Start() (when wired)
- `trip.pod_submitted` — Already published by TripService.SubmitPOD()

### 5.4 Existing Tables Used (No Schema Changes)

| Table | Domain | Purpose |
|-------|--------|---------|
| `trips` | trip | Trip records (VARCHAR ID, 50+ fields) |
| `trip_events` | trip | Hash-chained immutable event log |
| `trip_offers` | trip | Driver offer workflow |
| `customer_v2_shipments` | customer | Shipments (has `trip_id` column) |
| `customer_v2_booking_requests` | customer | Bookings |
| `customer_v2_shipment_stops` | customer | Multi-stop shipments |
| `vehicles` | vehicle | Fleet assets |
| `drivers` | driver | Personnel |
| `telemetry_points` | tracking | GPS telemetry (TimescaleDB) |
| `gps_locations` | tracking | Latest GPS positions |
| `geofences` | geofencing | Spatial zones |
| `route_optimization_requests` | route | Optimization inputs |
| `route_optimization_results` | route | Optimization outputs |

---

## 6. Implementation Plan (File-by-File)

### Phase 1: Wire Unwired Domains (Week 1)

| # | File Path | Action | Detail |
|---|-----------|--------|--------|
| 1 | `cmd/main.go` | **MODIFY** | Add EventBus initialization, wire trip/vehicle/driver/customer/route handlers, mount dispatch routes |
| 2 | `internal/config/config.go` | **MODIFY** | Add dispatch feature flags to `FeatureFlags` struct |
| 3 | `internal/middleware/` | **VERIFY** | `TenantMiddleware(logger)` and `JWKSAuthGate(secret, logger)` work correctly |

### Phase 2: Fix Shipment Assignment Stub (Week 1-2)

| # | File Path | Action | Detail |
|---|-----------|--------|--------|
| 4 | `customer/service/shipment_assignment_service.go` | **REWRITE** | Inject TripService + EventBus, replace `uuid.New()` with `tripService.Create()`, uncomment NATS publishing, add single-tx shipment update |
| 5 | `customer/` (handler wiring) | **MODIFY** | Update constructor calls in `cmd/main.go` to pass new dependencies |

### Phase 3: Create Dispatch Domain (Week 2-3)

| # | File Path | Action | Detail |
|---|-----------|--------|--------|
| 6 | `dispatch/models/dispatch.go` | **CREATE** | DashboardResponse, DispatchAction, SummaryCounts, etc. |
| 7 | `dispatch/models/scoring.go` | **CREATE** | Scoring weights, thresholds, constants |
| 8 | `dispatch/service/scoring_service.go` | **CREATE** | Weighted scoring (40/30/20/10), explanation generator |
| 9 | `dispatch/service/availability_service.go` | **CREATE** | Vehicle/driver availability checks, HOS validation |
| 10 | `dispatch/service/dispatch_service.go` | **CREATE** | Assign/redispatch orchestration, transaction management |
| 11 | `dispatch/engine/suggestion_engine.go` | **CREATE** | Proximity + availability + type matching scorer |
| 12 | `dispatch/repository/dispatch_repository.go` | **CREATE** | dispatch_actions writes, batch queries |
| 13 | `dispatch/handlers/dispatch_handler.go` | **CREATE** | HTTP handlers for dashboard, suggest, assign, redispatch |
| 14 | `dispatch/handlers/bulk_dispatch_handler.go` | **CREATE** | Batch dispatch handler (P2) |
| 15 | `dispatch/worker/auto_status_worker.go` | **CREATE** | GPS-triggered status transitions |
| 16 | `dispatch/metrics.go` | **CREATE** | Prometheus: dispatch_assignments_total, dispatch_suggestion_latency, auto_status_transitions |

### Phase 4: Database Migrations (Week 2)

| # | File Path | Action | Detail |
|---|-----------|--------|--------|
| 17 | `migrations/303_dispatch_tables.sql` | **CREATE** | dispatch_actions, dispatch_suggestions tables |
| 18 | Run migration | Execute on staging → production | CONCURRENTLY index creation |

### Phase 5: Add Missing Repository Methods (Week 2)

| # | File Path | Action | Detail |
|---|-----------|--------|--------|
| 19 | `driver/repository/driver_repository.go` | **ADD** | `GetByVehicleAssignment(ctx, vehicleID)` convenience method |
| 20 | `trip/repository/trip_repository.go` | **VERIFY** | `GetActiveTripsForVehicle`, `CheckOverlappingTrips`, `GetTripsByStatus` exist |
| 21 | `vehicle/repository/vehicle_repository.go` | **VERIFY** | `GetByOrgID` with status filter works |

### Phase 6: Testing & QA (Week 3-4)

| # | Activity | Detail |
|---|----------|--------|
| 22 | Unit tests | All dispatch domain files, >85% coverage |
| 23 | Integration tests | End-to-end dispatch flow, multi-tenant isolation |
| 24 | Stub fix tests | Prove real trip creation, NATS event publishing |
| 25 | Load tests | k6 for dashboard, suggest, assign endpoints |

---

## 7. Feature Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `dispatch_enabled` | bool | `true` | Master kill switch for entire dispatch feature |
| `auto_suggest_enabled` | bool | `true` | Toggle vehicle+driver suggestions on dashboard |
| `auto_status_detection` | bool | `true` | GPS-triggered status transitions (AutoStatusWorker) |
| `auto_receipt_generation` | bool | `true` | Delivery receipt PDF on trip completion |
| `auto_invoice_creation` | bool | `false` | Auto-create invoice on trip completion |
| `bulk_dispatch_enabled` | bool | `false` | Batch assignment endpoint |
| `route_optimization_dispatch` | bool | `false` | VRP→Trip converter |
| `max_trips_per_day` | int | `200` | Performance guardrail per org |

**Config struct addition:**
```go
type FeatureFlags struct {
    // ... existing fields ...
    DispatchEnabled          bool `mapstructure:"dispatch_enabled"`
    AutoSuggestEnabled       bool `mapstructure:"auto_suggest_enabled"`
    AutoStatusDetection      bool `mapstructure:"auto_status_detection"`
    AutoReceiptGeneration    bool `mapstructure:"auto_receipt_generation"`
    AutoInvoiceCreation      bool `mapstructure:"auto_invoice_creation"`
    BulkDispatchEnabled      bool `mapstructure:"bulk_dispatch_enabled"`
    RouteOptimizationDispatch bool `mapstructure:"route_optimization_dispatch"`
    MaxTripsPerDay           int  `mapstructure:"max_trips_per_day"`
}
```

---

## 8. Frontend Specification

### 8.1 Component Hierarchy

```
apps/dashboard-web/src/features/dispatch/
├── components/
│   ├── DispatchDashboard.tsx        # Main page — aggregate view
│   ├── SummaryCards.tsx             # Top row: total, pending, in-transit, delivered
│   ├── PendingShipmentsList.tsx     # Left panel: unassigned shipments
│   ├── ActiveTripsTable.tsx         # Center: in-transit trips with live status
│   ├── AvailableVehiclesPanel.tsx   # Right sidebar: available fleet
│   ├── AvailableDriversPanel.tsx    # Right sidebar: available drivers
│   ├── SuggestionModal.tsx          # Top 3 ranked matches with scores
│   ├── DispatchForm.tsx             # Manual assign form (vehicle+driver selector)
│   ├── RedispatchDialog.tsx         # Failed trip reassignment
│   ├── BulkDispatchDialog.tsx       # Multi-select batch assign (P2)
│   └── DispatchAlertsBanner.tsx     # Top banner: delayed trip alerts
├── hooks/
│   ├── useDispatchDashboard.ts      # Dashboard data fetching + polling
│   ├── useDispatchSuggestions.ts    # Suggestion fetching
│   └── useDispatchAssign.ts         # Assign mutation
├── services/
│   └── dispatchApi.ts               # RTK Query endpoints
├── store/
│   └── dispatchSlice.ts             # Redux state
├── types/
│   └── dispatch.ts                  # TypeScript interfaces
└── utils/
    └── statusColor.ts               # Trip status → color mapping
```

### 8.2 Redux State Shape

```typescript
interface DispatchState {
    dashboard: DashboardResponse | null;
    suggestions: Suggestion[];
    selectedShipment: string | null;
    assignForm: {
        vehicleId: string | null;
        driverId: string | null;
        scheduledStart: string | null;
        autoStart: boolean;
        autoComplete: boolean;
    };
    loading: {
        dashboard: boolean;
        suggestions: boolean;
        assigning: boolean;
    };
    error: string | null;
    lastUpdated: number | null;
}
```

### 8.3 Dispatch Dashboard UX

1. **Load:** GET /dispatch/dashboard → show summary cards → populate tables
2. **Poll:** Refresh every 15s (or WebSocket push if connected)
3. **Pending shipment click:** Open suggestion modal → POST /dispatch/suggest → show top 3 → "Assign" button
4. **One-click assign:** POST /dispatch/assign → success toast → refresh dashboard
5. **Active trip click:** Open trip detail with map (reuse VehicleTracking map component)
6. **Delayed trip alert:** Red banner at top → click to view trip → redispatch option

### 8.4 Real-Time Updates

**WebSocket integration:** If connected to `/ws/telemetry`, listen for `trip.completed`, `trip.started` events and update dashboard in real-time without polling.

---

## 9. Non-Functional Requirements

| Metric | Target | Strategy |
|--------|--------|----------|
| Suggestion engine | **<500ms P95** | Redis cached positions, indexed queries |
| Dispatch assign | **<1s P95** | Single transaction, pre-validated inputs |
| Dispatch dashboard | **<1s P95** | Redis summary cache (5s TTL), parallel queries |
| Auto status detection | **<5s from telemetry** | Async worker, 5s interval |
| Bulk dispatch (50 trips) | **<10s P95** | Worker pool (10 concurrent), per-trip tx |
| Max trips per org/day | **200** | Feature flag guardrail |
| Concurrent dispatchers | **50** | Stateless handlers, shared DB pool |

---

## 10. Testing Strategy

### 10.1 Unit Tests

| File | Coverage Target | Scenarios |
|------|----------------|-----------|
| `engine/suggestion_engine_test.go` | >90% | Proximity scoring, type filtering, capacity filtering, HOS check, no-match, empty org |
| `service/dispatch_service_test.go` | >90% | Happy path, vehicle busy, driver on trip, overlapping trips, shipment not CONFIRMED |
| `service/scoring_service_test.go` | >90% | Weighted scoring, edge cases (no driver, zero distance, max rating) |
| `service/availability_service_test.go` | >85% | Vehicle available/busy, driver available/on-trip, HOS exceeded |
| `repository/dispatch_repository_test.go` | >85% | Assignment tx success, shipment update, audit log insert, rollback on failure |
| `handlers/dispatch_handler_test.go` | >85% | Auth fail, valid request, invalid body, tenant mismatch |
| `worker/auto_status_worker_test.go` | >85% | Auto-complete triggered, vehicle not at drop, vehicle still moving, GPS missing |

### 10.2 Stub Fix Test Suite

The shipment assignment fix needs its own dedicated test suite:

| Test | Validation |
|------|-----------|
| `TestAssignShipment_CreatesRealTrip` | `uuid.New()` replaced — trip ID comes from tripService.Create |
| `TestAssignShipment_PublishesNATS` | EventBus.Publish is called with "shipment.assigned" |
| `TestAssignShipment_UpdatesShipment` | shipment.trip_id is set, status = ASSIGNED |
| `TestAssignShipment_RollbackOnFailure` | If trip creation succeeds but shipment update fails, transaction rolls back |
| `TestAssignShipment_TenantIsolation` | Org A cannot assign Org B's shipment |

### 10.3 Integration Tests

| Test | Validation |
|------|-----------|
| End-to-end dispatch | Booking CONFIRMED → suggestion → assign → trip CREATED → shipment ASSIGNED → NATS published |
| Auto status detection | Vehicle GPS near drop location → trip COMPLETED → shipment DELIVERED (within 10s) |
| Redispatch flow | Failed trip → redispatch → old trip CANCELLED → new trip CREATED → shipment re-linked |
| Multi-tenant isolation | Org A dispatch cannot see/access Org B shipments, vehicles, or trips |
| Double-booking prevention | Same vehicle cannot be assigned to overlapping trips |

### 10.4 Performance Tests (k6)

| Test | Load | Threshold |
|------|------|-----------|
| Suggestion engine | 50 concurrent, 2min | P95 <500ms |
| Dispatch assign | 20 concurrent, 2min | P95 <1s |
| Dispatch dashboard | 20 concurrent, 2min | P95 <1s |
| Bulk dispatch (50 trips) | 1 request | Total <10s |

---

## 11. Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Domain wiring complexity | **High** | Medium | Each domain has `RegisterRoutes` — mount under `/api/v1/{domain}` groups. Test each domain independently before integration. |
| GPS accuracy in urban areas | **High** | Medium | Increase geofence threshold to 500m for dense areas. Add time-based confirmation (stopped >2min). |
| Data consistency (shipment-trip link) | **Medium** | **High** | Single DB transaction for assignment. NATS events are best-effort (idempotent consumers). |
| `cmd/main.go` becomes too large | **High** | Low | Extract domain wiring into `internal/bootstrap/domains.go` function. Keep main.go thin. |
| Route optimization tables missing migrations | **Medium** | Medium | `road_restrictions` and `vehicle_permits` not in any migration. P0 doesn't need them — defer to P1. |
| Customer domain has zero NATS publishing | **Medium** | Medium | Fix in Phase 2 — inject EventBus, uncomment Publish calls. Each publish is wrapped in `if s.eventBus != nil`. |
| TripService.Create() parameter mismatch | **Medium** | Medium | The `CreateTripRequest` has 20+ fields. Map carefully from shipment data. Write comprehensive unit test. |
| Scope creep (multi-stop, routing) | **High** | Medium | Strict P0/P1/P2 boundaries. Defer VRP→Trip converter to P2. |
| Auto-status worker performance with many trips | **Medium** | Medium | Batch fetch all STARTED trips in one query. Use Redis pipeline for position lookups. Cap at 200 trips per cycle. |

---

## 12. Open Questions

| # | Question | Recommendation | Owner |
|---|----------|---------------|-------|
| 1 | Auto-dispatch trigger: on booking confirm or manual? | **Manual by default** — dispatcher reviews suggestions, clicks assign. Auto-dispatch via flag for Phase 2. | Product |
| 2 | Max suggestions per shipment? | **3** (configurable via `top_n` param). Most dispatchers decide within top 2. | Engineering |
| 3 | Redispatch: cancel old trip or mark failed? | **Mark failed first**, then create new trip. Cancel only if trip was never started. | Engineering |
| 4 | Dispatch audit retention? | **1 year** (compliance). `dispatch_actions` table with date-partitioned index. | Compliance |
| 5 | VRP→Trip converter: direct or via dispatch? | **Via dispatch** — reuse the assign flow. Converter generates `AssignRequest` objects that go through the same pipeline. | Engineering |
| 6 | TripService.Create() returns VARCHAR ID, but trip model uses string. Consistency? | The trip model uses `string` for IDs (not UUID). Dispatch repository must use `VARCHAR(255)` for trip_id FK in dispatch_actions. | Engineering |
| 7 | What happens if vehicle has no assigned driver? | Allow vehicle-only dispatch. Dispatcher assigns driver separately via vehicle.AssignDriver endpoint. Suggestion engine should still include driverless vehicles with score reduced by 20% (missing driver rating). | Product |

---

## 13. Appendix

### A. Cross-Reference: Architecture Doc Claims vs Codebase Reality

| Architecture Claim | Verified Reality | Status |
|-------------------|-----------------|--------|
| Trip domain has 22 endpoints | ✅ Verified — RegisterRoutes mounts 22 endpoints | ✅ Correct |
| Trip domain NOT wired into main.go | ✅ Confirmed — zero trip references in main.go | ✅ Correct |
| ShipmentAssignmentService uses `uuid.New()` mock | ✅ Confirmed — line 44: `mockTripID := uuid.New()` | ✅ Correct |
| EventBus.Publish commented out in customer domain | ✅ Confirmed — both calls are commented | ✅ Correct |
| Customer domain has zero EventBus injection | ✅ Confirmed — no EventBus field or constructor param | ✅ Correct |
| Route domain has 3 solvers | ✅ Confirmed — ClarkeWrightSolver, HeuristicSolver, NearestNeighbor | ✅ Correct |
| Route tables have no migrations for road_restrictions/vehicle_permits | ✅ Confirmed — only route_optimization_* tables exist | ✅ Correct |
| No VRP→Trip converter exists | ✅ Confirmed — no code converts RouteOptimizationResult to trips | ✅ Correct |
| Driver domain has 30+ endpoints | ✅ Confirmed — handlers cover CRUD, HOS, behavior, shifts, payroll, etc. | ✅ Correct |
| Trip state machine: 8 states, hash-chained events | ✅ Confirmed — lifecycle.go defines transitions, events.go has hash chain | ✅ Correct |

### B. Trip State Machine — Valid Transitions

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

Terminal states: `COMPLETED`, `CANCELLED`, `FAILED` — no further transitions allowed.

### B1. Independent Validation Report — All Discrepancies Found & Fixed

This spec was validated by TWO independent codebase audits. Here are ALL discrepancies found and their resolutions:

| # | Severity | Original Spec Claim | Actual Code | Fix Applied |
|---|----------|-------------------|-------------|-------------|
| 1 | 🔴 HIGH | Trip struct has `ShipmentID` field | **Trip struct has NO `ShipmentID` field** — linked only via `customer_v2_shipments.trip_id` | All code uses shipment-side link. AutoStatusWorker publishes `trip.completed` instead of `shipment.delivered`. |
| 2 | 🔴 HIGH | `EventBus.Publish(ctx, subject, data)` — 3 params | **Actual: 4 params** `(ctx, subject, data, orgID)` | All 5 Publish calls fixed with 4th `orgID` param, nil-checked. |
| 3 | 🔴 HIGH | `trip.TripNumber` field exists | **Trip model has NO `TripNumber` field** — Trip.ID is the unique identifier | Removed all `trip.TripNumber` references. Uses `trip.ID` throughout. |
| 4 | 🟡 MODERATE | `CheckOverlappingTrips(ctx, vehicleID, driverID)` — 3 params | **Actual: 5 params** `(ctx, vehicleID, driverID string, start, end time.Time)` | Fixed: passes `ScheduledStart` and `ScheduledStart + duration` as time bounds. |
| 5 | 🟡 MODERATE | `liveCache.GetLiveLocation(ctx, trip.VehicleID)` — VehicleID is string | **GetLiveLocation expects `uuid.UUID`** (gofrs package) | Added `uuid.Parse(trip.VehicleID)` conversion. For suggestion engine: `uuidParse(v.ID.String())` to bridge google→gofrs UUID packages. |
| 6 | 🟡 MODERATE | `CompleteTripRequest{EndOdometer, PODData: ...}` | **Actual: `OdometerReading` field, `PODDocuments []PODDocument`** (not PODData) | Fixed: uses `OdometerReading`, `PODDocuments: []PODDocument{{Type: "gps_auto_capture", Timestamp: now()}}`. |
| 7 | 🟡 MODERATE | `Waypoint{Lat, Lng, Address}` flat fields | **Actual: `Waypoint.Location Location{Latitude, Longitude, Address}`** nested struct | Fixed: `Location: loc` where loc is unmarshaled from stop.Location []byte. |
| 8 | 🟡 MODERATE | `Shipment.Origin.Address` — Origin is a struct | **Actual: `Origin []byte`** (JSON blob) | Fixed: `json.Unmarshal(shipment.Origin, &originLoc)` before use. Same for Destination. |
| 9 | 🟡 MODERATE | `shipment.OrgID` field exists | **Shipment has NO `OrgID` field** — has `CustomerID uuid.UUID` | Fixed: org validation done at handler level via TenantContextMiddleware. Shipment org inferred from context. |
| 10 | 🟡 MODERATE | `CargoDetails{Type, Weight}` with pointer fields | **Actual: `Type string, WeightKg float64`** — both non-pointer | Fixed: nil-checked `booking.CargoType` → `cargoType`, `booking.Weight` → `cargoWeight`, used as `Type: cargoType, WeightKg: cargoWeight`. |
| 11 | 🟡 MODERATE | `GetTripsByStatus(ctx, "STARTED")` exists | **Does NOT exist** — only `List(ctx, filter)` with `TripListFilter{Status: ...}` | Fixed: uses `List(ctx, TripListFilter{Status: "STARTED", Limit: 200})`. |
| 12 | 🟡 MODERATE | `vehicleRepo.ListByOrg()` exists | **Does NOT exist** — `GetByOrgID(ctx, orgID, limit, offset)` is the actual method | All 3 occurrences fixed to use `GetByOrgID`. |
| 13 | 🟢 MINOR | `trip.assigned` / `shipment.assigned` NATS subjects exist | **Never published** — truly new subjects | Documented as NEW in NATS events table. |
| 14 | ℹ️ INFO | `TenantContextMiddleware` in cmd/main.go | Defined in `tenant_context.go.bak` | Flagged. Working middleware is `TenantMiddleware(logger)`. |

### C. Key Dependencies Graph

```
dispatch/
├── → trip/          (Create, Approve, Start, Complete, Cancel, GetActiveTripsForVehicle, CheckOverlappingTrips)
├── → vehicle/       (GetByID, ListByOrg)
├── → driver/        (GetByID, GetByVehicleAssignment, List)
├── → customer/      (GetShipmentByID, UpdateShipmentTripID, UpdateShipmentStatus, GetBookingByID)
├── → tracking/      (LiveCacheRepository for GPS positions)
├── → geofencing/    (Geofence proximity for auto-status — optional)
├── → messaging/     (EventBus for NATS publishing)
├── → database/      (pgxpool for dispatch_actions writes)
└── → platform/      (Response helpers, metrics)
```

### D. Glossary

| Term | Definition | Code Location |
|------|-----------|---------------|
| **Trip** | Delivery assignment: vehicle + driver + route + waypoints | `internal/domains/trip/models/trip.go` |
| **Shipment** | Customer-facing delivery record, linked to trip via `trip_id` | `internal/domains/customer/models/models.go` |
| **Booking** | Customer request for delivery service | `internal/domains/customer/models/models.go` |
| **Dispatch** | Process of assigning vehicle+driver to shipment | NEW — this spec |
| **Suggestion** | Ranked vehicle+driver match for a shipment | NEW — dispatch/engine |
| **Redispatch** | Re-assigning a failed trip to new vehicle+driver | NEW — this spec |
| **POD** | Proof of Delivery (signature, photo, GPS) | `internal/domains/trip/models/trip.go` |
| **HOS** | Hours of Service — driver shift limits | `internal/domains/driver/models/driver.go` |
| **VRP** | Vehicle Routing Problem — optimization | `internal/domains/route/service/solver.go` |
