# Fleet-Core Integration Plan: Vehicle Tracking + Dispatch Flow

**Document ID:** ENG-FC-003  
**Version:** 1.0  
**Created:** April 7, 2026  
**Status:** Draft — Ready for Engineering Review  
**Related Specs:**  
- [Vehicle Tracking Product Spec](../product-spec/VEHICLE_TRACKING.md)  
- [Dispatch Flow Product Spec](../product-spec/DISPATCH_FLOW.md)  
- [Parent PRD](../docs/PRD.md)

---

## 1. Executive Summary

This document maps all Vehicle Tracking and Dispatch Flow product requirements to concrete implementation tasks within the `services/fleet-core` monolith. It identifies what already exists, what's partially implemented, and what needs to be built from scratch.

### Current State Assessment

| Domain | Status | Tables | Handlers | Service | Repository | Notes |
|--------|--------|--------|----------|---------|------------|-------|
| **trip** | ⚠️ Partial | ✅ 7 tables | ✅ Basic CRUD | ✅ State machine | ✅ Exists | Missing dispatch workflow, receipt gen, auto-detect |
| **tracking** | ⚠️ Partial | ✅ 8+ tables | ✅ Telemetry ingest | ✅ Basic service | ✅ Exists | Missing WebSocket live updates, history trail API |
| **geofencing** | ⚠️ Partial | ✅ 3 tables | ✅ Basic handlers | ✅ Service exists | ✅ Exists | Missing auto-alert integration with tracking |
| **billing** | ✅ Live | ✅ 15+ tables | ✅ Invoice CRUD | ✅ PDF gen | ✅ Exists | Missing auto-trigger from trip completion |
| **vehicle** | ✅ Live | ✅ 12+ tables | ✅ Vehicle CRUD | ✅ Service | ✅ Exists | Need `GET /vehicles/locations` endpoint |
| **driver** | ✅ Live | ✅ 10+ tables | ✅ Driver CRUD | ✅ Service | ✅ Exists | Need driver availability check |
| **alerts** | ✅ Live | ✅ 12+ tables | ✅ Alert CRUD | ✅ Rule engine | ✅ Exists | Missing geofence event integration |

---

## 2. Architecture Overview

### 2.1 Current Fleet-Core Structure

```
services/fleet-core/
├── cmd/main.go                           # Main entry point, router setup
├── internal/
│   ├── domains/
│   │   ├── tracking/                     # GPS telemetry domain
│   │   │   ├── handlers/                 # ingest_handler.go, telemetry_handler.go
│   │   │   ├── service/                  # trackingv2service
│   │   │   ├── repository/               # trackingv2repo
│   │   │   ├── models/                   # telemetry models
│   │   │   ├── worker/                   # background processors
│   │   │   └── analytics/                # analytics aggregation
│   │   ├── trip/                         # Trip management domain
│   │   │   ├── handlers/                 # trip_handler.go
│   │   │   ├── service/                  # trip_service.go, state_machine.go
│   │   │   ├── repository/               # trip repository
│   │   │   ├── models/                   # trip.go, trip_lifecycle.go
│   │   │   └── metrics.go
│   │   ├── geofencing/                   # Geofence domain
│   │   │   ├── handlers/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   └── engine/                   # spatial detection engine
│   │   ├── vehicle/                      # Vehicle management
│   │   ├── driver/                       # Driver management
│   │   ├── billing/                      # Billing & invoices
│   │   └── alerts/                       # Alert rules & notifications
│   ├── handlers/                         # Cross-domain handlers
│   ├── middleware/                        # Auth, CORS, rate limiting
│   ├── gateway/                          # API gateway/router
│   └── platform/                         # Shared platform code
└── migrations/                           # Database migrations (070-300+)
```

### 2.2 Target Architecture (After Integration)

```
services/fleet-core/
├── internal/domains/
│   ├── tracking/                         # Enhanced with:
│   │   ├── handlers/                     # + vehicle_locations_handler.go
│   │   │                               # + history_trail_handler.go
│   │   │                               # + websocket_handler.go (NEW)
│   │   ├── service/                      # + status_detector.go (NEW)
│   │   │                               # + history_service.go (NEW)
│   │   ├── repository/                   # + optimized queries
│   │   └── worker/                       # + geofence_checker.go (NEW)
│   │
│   ├── trip/                             # Enhanced with:
│   │   ├── handlers/                     # + dispatch_dashboard_handler.go (NEW)
│   │   │                               # + receipt_handler.go (NEW)
│   │   │                               # + exception_handler.go (NEW)
│   │   ├── service/                      # + auto_status_detector.go (NEW)
│   │   │                               # + receipt_generator.go (NEW)
│   │   │                               # + pricing_calculator.go (NEW)
│   │   │                               # + bulk_trip_service.go (NEW)
│   │   ├── repository/                   # + dashboard queries
│   │   └── models/                       # + dispatch_models.go (NEW)
│   │
│   ├── geofencing/                       # Enhanced with:
│   │   ├── service/                      # + alert_integration.go (NEW)
│   │   └── worker/                       # + continuous_checker.go (NEW)
│   │
│   ├── dispatch/                         # NEW DOMAIN (optional, or extend trip)
│   │   ├── handlers/dispatch_handler.go
│   │   ├── service/dispatch_service.go
│   │   ├── repository/dispatch_repo.go
│   │   └── models/dispatch.go
│   │
│   └── receipt/                          # NEW DOMAIN (or under trip/billing)
│       ├── handlers/receipt_handler.go
│       ├── service/receipt_generator.go
│       └── models/receipt.go
```

---

## 3. Vehicle Tracking Integration Tasks

### 3.1 What Already Exists ✅

| Component | File | Status | Notes |
|-----------|------|--------|-------|
| Telemetry ingest handler | `tracking/handlers/ingest_handler.go` | ✅ Live | Accepts POST `/api/v1/telemetry` |
| Telemetry handler | `tracking/handlers/telemetry_handler.go` | ✅ Live | Query telemetry data |
| Tracking query handler | `tracking/handlers/tracking_query_handler.go` | ✅ Live | Basic queries |
| Tracking service | `tracking/service/` | ✅ Exists | Basic service layer |
| Tracking repository | `tracking/repository/` | ✅ Exists | DB access |
| Telemetry tables | `telemetry_points` (partitioned) | ✅ Exists | TimescaleDB hypertable |
| GPS locations table | `gps_locations` | ✅ Exists | Latest positions |
| Analytics handler | `tracking/handlers/analytics_handler.go` | ✅ Exists | Fleet analytics |
| Tracking worker | `tracking/worker/` | ✅ Exists | Background processors |
| Tracking metrics | `tracking/metrics.go` | ✅ Exists | Prometheus metrics |

### 3.2 What Needs to Be Built 🆕

#### Epic VT-1: Vehicle Locations API

**Task VT-1.1**: Create `GET /api/v1/vehicles/locations` endpoint

**Location**: `internal/domains/tracking/handlers/vehicle_locations_handler.go` (NEW)

```go
package handlers

type VehicleLocationsHandler struct {
    service TrackingServiceInterface
    logger  *zap.Logger
}

func (h *VehicleLocationsHandler) HandleGetVehicleLocations(w http.ResponseWriter, r *http.Request) {
    // 1. Extract org_id from JWT context
    // 2. Query latest position for all org vehicles
    // 3. Join with vehicles table (registration, status)
    // 4. Join with drivers table (driver name)
    // 5. Return JSON response
    // Performance: <2s P95 for 500 vehicles
}
```

**Repository Query** (in `tracking/repository/`):
```sql
SELECT 
    v.id as vehicle_id,
    v.registration,
    v.status as vehicle_status,
    gl.latitude,
    gl.longitude,
    gl.speed_kmh,
    gl.heading,
    gl.ignition,
    gl.fuel_level,
    gl.recorded_at as last_update,
    d.name as driver_name,
    d.phone as driver_phone
FROM vehicles v
LEFT JOIN gps_locations gl ON gl.vehicle_id = v.id
    AND gl.recorded_at = (
        SELECT MAX(recorded_at) FROM gps_locations WHERE vehicle_id = v.id
    )
LEFT JOIN drivers d ON d.id = v.current_driver_id
WHERE v.org_id = $1 
  AND v.deleted_at IS NULL
ORDER BY v.registration;
```

**Deliverables**:
- [ ] `tracking/handlers/vehicle_locations_handler.go`
- [ ] Add query to `tracking/repository/tracking_repo.go`
- [ ] Add method to `tracking/service/tracking_service.go`
- [ ] Register route in `cmd/main.go`
- [ ] Unit tests (>90% coverage)
- [ ] Performance test with 500 vehicles

---

#### Epic VT-2: WebSocket Real-time Updates

**Task VT-2.1**: Create WebSocket handler for live position updates

**Location**: `internal/domains/tracking/handlers/websocket_handler.go` (NEW)

```go
package handlers

type WebSocketHandler struct {
    redis    *redis.Client
    nats     *nats.Conn
    logger   *zap.Logger
    clients  sync.Map // org_id → set of connections
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // 1. Upgrade HTTP to WebSocket
    // 2. Authenticate via JWT query param
    // 3. Subscribe to Redis channel for org_id
    // 4. Start read/write loop
    // 5. Push position updates as they arrive
}
```

**Redis Pub/Sub Integration**:
```go
// On telemetry ingest (in ingest_handler.go):
func (h *IngestHandler) PublishPositionUpdate(ctx context.Context, orgID, vehicleID string, data PositionUpdate) {
    message, _ := json.Marshal(map[string]interface{}{
        "type": "position_update",
        "vehicle_id": vehicleID,
        "data": data,
    })
    h.redis.Publish(ctx, fmt.Sprintf("org:%s:telemetry", orgID), message)
}
```

**Deliverables**:
- [ ] `tracking/handlers/websocket_handler.go`
- [ ] WebSocket upgrade logic in `cmd/main.go`
- [ ] Redis pub/sub integration in `ingest_handler.go`
- [ ] Client subscription protocol (JSON messages)
- [ ] Connection management (heartbeat, reconnect)
- [ ] Load test: 100 concurrent WS connections
- [ ] Fallback to polling if WS unavailable

---

#### Epic VT-3: History Trail API

**Task VT-3.1**: Create `GET /api/v1/vehicles/{id}/telemetry/history` endpoint

**Location**: `internal/domains/tracking/handlers/history_trail_handler.go` (NEW)

```go
func (h *HistoryTrailHandler) HandleGetHistoryTrail(w http.ResponseWriter, r *http.Request) {
    // 1. Parse vehicle_id, start, end, resolution params
    // 2. Query TimescaleDB with time_bucket for downsampling
    // 3. Calculate summary stats (distance, speed, time)
    // 4. Return JSON with points array + summary
    // Performance: <3s P95 for 24h data
}
```

**TimescaleDB Query**:
```sql
SELECT 
    time_bucket($1, time) as bucket_time,
    AVG(latitude) as latitude,
    AVG(longitude) as longitude,
    AVG(speed_kmh) as avg_speed,
    MAX(speed_kmh) as max_speed,
    COUNT(*) as point_count
FROM telemetry_points
WHERE vehicle_id = $2
  AND time >= $3
  AND time <= $4
GROUP BY bucket_time
ORDER BY bucket_time;
```

**Deliverables**:
- [ ] `tracking/handlers/history_trail_handler.go`
- [ ] Add query to `tracking/repository/`
- [ ] Add method to `tracking/service/`
- [ ] Auto-resolution logic (10s/30s/60s/5m/15m based on time range)
- [ ] Distance calculation (Haversine formula)
- [ ] Summary statistics computation
- [ ] Unit tests (>85% coverage)
- [ ] Performance test with 1,440 points (24h at 1-min)

---

#### Epic VT-4: Geofence Auto-Detection

**Task VT-4.1**: Create background worker for continuous geofence checking

**Location**: `internal/domains/geofencing/worker/geofence_checker.go` (NEW)

```go
package worker

type GeofenceChecker struct {
    db          *sql.DB
    geoService  GeofencingService
    alertService AlertService
    logger      *zap.Logger
    interval    time.Duration // 10 seconds
}

func (gc *GeofenceChecker) Run(ctx context.Context) {
    ticker := time.NewTicker(gc.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            gc.checkRecentPositions(ctx)
        }
    }
}

func (gc *GeofenceChecker) checkRecentPositions(ctx context.Context) {
    // 1. Fetch positions from last 10 seconds
    // 2. For each position, check against all active geofences (PostGIS ST_Contains)
    // 3. Compare with previous state (inside vs outside)
    // 4. If state changed → log geofence_event, trigger alert
    // 5. Send alert via NATS to notification service
}
```

**PostGIS Query**:
```sql
-- Check if point is inside polygon geofence
SELECT 
    g.id as geofence_id,
    g.name as geofence_name,
    ST_Contains(g.polygon, ST_SetSRID(ST_MakePoint($1, $2), 4326)) as is_inside
FROM geofences g
WHERE g.org_id = $3
  AND g.status = 'active'
  AND g.type = 'polygon';

-- Check if point is inside circle geofence
SELECT 
    g.id as geofence_id,
    g.name as geofence_name,
    ST_DWithin(
        ST_SetSRID(ST_MakePoint($1, $2), 4326),
        ST_SetSRID(ST_MakePoint(g.center_lng, g.center_lat), 4326),
        g.radius_m
    ) as is_inside
FROM geofences g
WHERE g.org_id = $3
  AND g.status = 'active'
  AND g.type = 'circle';
```

**Deliverables**:
- [ ] `geofencing/worker/geofence_checker.go`
- [ ] State tracking (Redis hash: `geofence_state:{vehicle_id}:{geofence_id}`)
- [ ] PostGIS spatial queries in `geofencing/repository/`
- [ ] Alert integration (send to `alerts` domain via NATS)
- [ ] Geofence event logging (`geofence_events` table)
- [ ] Throttle logic (require 2 consecutive detections)
- [ ] Performance test: 500 vehicles × 10 geofences

---

#### Epic VT-5: Performance Optimization

**Task VT-5.1**: Add database indexes for tracking queries

**Location**: `migrations/301_tracking_performance_indexes.sql` (NEW)

```sql
-- Latest position query
CREATE INDEX CONCURRENTLY idx_gps_locations_latest 
    ON gps_locations (vehicle_id, recorded_at DESC);

-- Telemetry history query (already partitioned, add composite index)
CREATE INDEX CONCURRENTLY idx_telemetry_vehicle_time 
    ON telemetry_points (vehicle_id, time DESC);

-- Geofence spatial indexes
CREATE INDEX CONCURRENTLY idx_geofences_polygon_spatial 
    ON geofences USING GIST (polygon) 
    WHERE status = 'active';

CREATE INDEX CONCURRENTLY idx_geofences_circle_lookup 
    ON geofences (org_id, status, type) 
    WHERE status = 'active';

-- Geofence event queries
CREATE INDEX CONCURRENTLY idx_geofence_events_vehicle_time 
    ON geofence_events (vehicle_id, occurred_at DESC);

-- Vehicle org lookup with status
CREATE INDEX CONCURRENTLY idx_vehicles_org_status 
    ON vehicles (org_id, status, deleted_at) 
    WHERE deleted_at IS NULL;
```

**Deliverables**:
- [ ] Migration file with all indexes
- [ ] Verify index usage with `EXPLAIN ANALYZE`
- [ ] Benchmark before/after query performance

---

### 3.3 Route Registration (Vehicle Tracking)

**File**: `cmd/main.go`

Add these routes to the tracking router section:

```go
// Vehicle Tracking routes
trackingRouter := chi.NewRouter()
trackingRouter.Use(authMiddleware.Authenticate)
trackingRouter.Use(governanceMiddleware.RequireOrgContext)

// Existing routes
trackingRouter.Post("/telemetry", trackingIngestHandler.HandleTelemetry)
trackingRouter.Get("/telemetry", trackingQueryHandler.HandleQueryTelemetry)
trackingRouter.Get("/analytics", trackingAnalyticsHandler.HandleAnalytics)

// NEW: Vehicle Tracking endpoints
trackingRouter.Get("/vehicles/locations", vehicleLocationsHandler.HandleGetVehicleLocations)
trackingRouter.Get("/vehicles/{id}/telemetry/history", historyTrailHandler.HandleGetHistoryTrail)
trackingRouter.Get("/ws/telemetry", websocketHandler.HandleWebSocket) // WebSocket upgrade

// Geofence routes
geoRouter := chi.NewRouter()
geoRouter.Use(authMiddleware.Authenticate)
geoRouter.Get("/geofences", geoHandlers.GetGeofences)
geoRouter.Post("/geofences", geoHandlers.CreateGeofence)
```

---

## 4. Dispatch Flow Integration Tasks

### 4.1 What Already Exists ✅

| Component | File | Status | Notes |
|-----------|------|--------|-------|
| Trip handler | `trip/handlers/trip_handler.go` | ✅ Live | Basic CRUD |
| Trip service | `trip/service/trip_service.go` | ✅ Live | State machine, CRUD |
| Trip state machine | `trip/service/state_machine.go` | ✅ Live | CREATED→STARTED→COMPLETED |
| Trip repository | `trip/repository/` | ✅ Exists | DB access |
| Trip models | `trip/models/trip.go` | ✅ Exists | Trip struct, events |
| Trip tables | `trips`, `trip_settings`, `trip_profitability` | ✅ Exists | Migration 070 |
| Billing service | `billing/service/` | ✅ Live | Invoice generation |
| Billing PDF | `billing/pdf/` | ✅ Live | PDF generation |
| Billing handlers | `billing/handlers/` | ✅ Live | Invoice CRUD |

### 4.2 What Needs to Be Built 🆕

#### Epic DF-1: Trip Creation Enhancement

**Task DF-1.1**: Enhance trip creation with dispatch workflow fields

**Location**: `internal/domains/trip/models/dispatch_models.go` (NEW)

```go
package models

// TripPriority represents delivery priority
type TripPriority string

const (
    PriorityLow      TripPriority = "low"
    PriorityMedium   TripPriority = "medium"
    PriorityHigh     TripPriority = "high"
    PriorityCritical TripPriority = "critical"
)

// TripType represents the type of trip
type TripType string

const (
    TripTypeDelivery TripType = "delivery"
    TripTypePickup   TripType = "pickup"
    TripTypeTransfer TripType = "transfer"
    TripTypeReturn   TripType = "return"
)

// CreateTripRequest enhanced for dispatch workflow
type CreateTripRequest struct {
    TripType   TripType     `json:"trip_type"`
    Priority   TripPriority `json:"priority"`
    VehicleID  string       `json:"vehicle_id"`
    DriverID   string       `json:"driver_id"`
    
    // Pickup location with contact info
    PickupName         string     `json:"pickup_name"`
    PickupAddress      string     `json:"pickup_address"`
    PickupLatitude     float64    `json:"pickup_latitude"`
    PickupLongitude    float64    `json:"pickup_longitude"`
    PickupContactName  string     `json:"pickup_contact_name"`
    PickupContactPhone string     `json:"pickup_contact_phone"`
    PickupScheduled    time.Time  `json:"pickup_scheduled"`
    
    // Drop location with contact info
    DropName           string     `json:"drop_name"`
    DropAddress        string     `json:"drop_address"`
    DropLatitude       float64    `json:"drop_latitude"`
    DropLongitude      float64    `json:"drop_longitude"`
    DropContactName    string     `json:"drop_contact_name"`
    DropContactPhone   string     `json:"drop_contact_phone"`
    DropExpected       time.Time  `json:"drop_expected"`
    
    // Goods details
    GoodsDescription   string  `json:"goods_description"`
    GoodsWeightKg      float64 `json:"goods_weight_kg"`
    GoodsVolumeM3      float64 `json:"goods_volume_m3"`
    GoodsPieces        int     `json:"goods_pieces"`
    GoodsValueINR      float64 `json:"goods_value_inr"`
    GoodsInvoiceNumber string  `json:"goods_invoice_number"`
    
    // Customer linkage
    CustomerID *string `json:"customer_id"`
    
    // Automation settings
    AutoStart    bool `json:"auto_start"`
    AutoComplete bool `json:"auto_complete"`
    
    // Notification settings
    NotifyCustomerOnStart    bool `json:"notify_customer_on_start"`
    NotifyCustomerOnComplete bool `json:"notify_customer_on_complete"`
    AlertDispatcherOnDelay   bool `json:"alert_dispatcher_on_delay"`
    
    // Metadata
    Notes string `json:"notes"`
}

// TripStatus for dispatch workflow
const (
    StatusPending    = "pending"
    StatusInTransit  = "in-transit"
    StatusDelivered  = "delivered"
    StatusFailed     = "failed"
    StatusCancelled  = "cancelled"
    StatusException  = "exception"
)
```

**Task DF-1.2**: Add vehicle/driver availability validation

**Location**: `internal/domains/trip/service/trip_service.go` (enhance existing `Create` method)

```go
func (s *TripService) Create(ctx context.Context, orgID string, req *CreateTripRequest) (*Trip, error) {
    // EXISTING: Basic validation
    
    // NEW: Check vehicle availability
    vehicleActiveTrips, err := s.repo.GetActiveTripsForVehicle(ctx, req.VehicleID)
    if err != nil {
        return nil, fmt.Errorf("check vehicle availability: %w", err)
    }
    if len(vehicleActiveTrips) > 0 {
        return nil, ErrVehicleAlreadyAssigned{TripID: vehicleActiveTrips[0].ID}
    }
    
    // NEW: Check driver availability
    driverActiveTrips, err := s.repo.GetActiveTripsForDriver(ctx, req.DriverID)
    if err != nil {
        return nil, fmt.Errorf("check driver availability: %w", err)
    }
    if len(driverActiveTrips) > 0 {
        return nil, ErrDriverAlreadyAssigned{TripID: driverActiveTrips[0].ID}
    }
    
    // NEW: Calculate estimated distance and duration
    distance, duration := CalculateRouteEstimates(
        req.PickupLatitude, req.PickupLongitude,
        req.DropLatitude, req.DropLongitude,
    )
    
    // EXISTING: Create trip record
    // ...
}
```

**Deliverables**:
- [ ] `trip/models/dispatch_models.go`
- [ ] Enhance `trip/service/trip_service.go` Create method
- [ ] Add availability check queries to `trip/repository/`
- [ ] Add distance/duration estimation (Haversine + avg speed)
- [ ] Custom error types for conflicts
- [ ] Unit tests for all validation scenarios

---

#### Epic DF-2: Auto Status Detection

**Task DF-2.1**: Create auto status detection service

**Location**: `internal/domains/trip/service/auto_status_detector.go` (NEW)

```go
package service

type AutoStatusDetector struct {
    db         *sql.DB
    redis      *redis.Client
    nats       *nats.Conn
    logger     *zap.Logger
    tripRepo   TripRepository
}

// CheckOnTelemetry called on every telemetry ingest
func (asd *AutoStatusDetector) CheckOnTelemetry(ctx context.Context, vehicleID string, lat, lng float64, ignition bool) error {
    // 1. Find pending trips for this vehicle
    pendingTrips, err := asd.tripRepo.FindPendingTrips(ctx, vehicleID)
    if err != nil {
        return fmt.Errorf("find pending trips: %w", err)
    }
    
    for _, trip := range pendingTrips {
        // 2. Check if vehicle is near pickup (<200m) and ignition ON
        distToPickup := CalculateDistance(lat, lng, trip.PickupLat, trip.PickupLng)
        
        if distToPickup < 200 && ignition {
            // Auto-start trip
            err = asd.tripRepo.UpdateTripStatus(ctx, trip.ID, "in-transit", "system_auto_detect")
            if err != nil {
                asd.logger.Error("Failed to auto-start trip", "trip_id", trip.ID, "error", err)
                continue
            }
            
            // Log event
            asd.tripRepo.InsertTripEvent(ctx, trip.ID, "trip_started", map[string]interface{}{
                "location": map[string]float64{"lat": lat, "lng": lng},
                "trigger": "vehicle_ignition_on",
                "distance_to_pickup_m": distToPickup,
            })
            
            // Notify dispatcher via NATS (WebSocket relay)
            asd.nats.Publish("dispatch.status_change", map[string]interface{}{
                "org_id": trip.OrgID,
                "trip_id": trip.ID,
                "status": "in-transit",
            })
        }
    }
    
    // 3. Check in-transit trips for completion
    inTransitTrips, err := asd.tripRepo.FindInTransitTrips(ctx, vehicleID)
    if err != nil {
        return fmt.Errorf("find in-transit trips: %w", err)
    }
    
    for _, trip := range inTransitTrips {
        // 4. Check if vehicle is near drop (<100m) and stopped for >5 min
        distToDrop := CalculateDistance(lat, lng, trip.DropLat, trip.DropLng)
        
        if distToDrop < 100 {
            stopped, err := asd.isVehicleStopped(ctx, vehicleID, 5*time.Minute)
            if err != nil {
                continue
            }
            
            if stopped {
                // Auto-complete trip
                err = asd.tripRepo.UpdateTripStatus(ctx, trip.ID, "delivered", "system_auto_detect")
                if err != nil {
                    asd.logger.Error("Failed to auto-complete trip", "trip_id", trip.ID, "error", err)
                    continue
                }
                
                // Log event
                asd.tripRepo.InsertTripEvent(ctx, trip.ID, "trip_completed", map[string]interface{}{
                    "location": map[string]float64{"lat": lat, "lng": lng},
                    "trigger": "destination_arrival",
                    "distance_to_drop_m": distToDrop,
                })
                
                // Trigger receipt generation (async)
                asd.triggerReceiptGeneration(ctx, trip.ID)
                
                // Trigger invoice creation (async, if enabled)
                asd.triggerInvoiceCreation(ctx, trip.ID)
                
                // Notify dispatcher
                asd.nats.Publish("dispatch.status_change", map[string]interface{}{
                    "org_id": trip.OrgID,
                    "trip_id": trip.ID,
                    "status": "delivered",
                })
            }
        }
    }
    
    return nil
}

func (asd *AutoStatusDetector) isVehicleStopped(ctx context.Context, vehicleID string, duration time.Duration) (bool, error) {
    query := `
        SELECT MAX(speed_kmh)
        FROM telemetry_points
        WHERE vehicle_id = $1
          AND time > NOW() - $2
    `
    var maxSpeed float64
    err := asd.db.QueryRowContext(ctx, query, vehicleID, duration).Scan(&maxSpeed)
    if err != nil {
        return false, err
    }
    
    return maxSpeed < 2.0, nil // Consider stopped if <2 km/h
}
```

**Task DF-2.2**: Integrate with telemetry ingest

**Location**: `internal/domains/tracking/handlers/ingest_handler.go` (modify existing)

```go
func (h *IngestHandler) HandleTelemetry(w http.ResponseWriter, r *http.Request) {
    // EXISTING: Parse telemetry, validate, save to DB
    
    // NEW: Trigger auto status detection
    if h.autoStatusDetector != nil {
        go func() {
            ctx := context.Background()
            err := h.autoStatusDetector.CheckOnTelemetry(ctx, 
                telemetry.VehicleID, 
                telemetry.Latitude, 
                telemetry.Longitude, 
                telemetry.Ignition,
            )
            if err != nil {
                h.logger.Error("Auto status detection failed", "error", err)
            }
        }()
    }
    
    // EXISTING: Return response
}
```

**Deliverables**:
- [ ] `trip/service/auto_status_detector.go`
- [ ] Distance calculation utility (Haversine)
- [ ] Integration with `tracking/handlers/ingest_handler.go`
- [ ] NATS event publishing for dispatch dashboard
- [ ] Threshold configuration (200m pickup, 100m drop, 5min stop)
- [ ] Unit tests for all auto-detect scenarios
- [ ] Integration test with mock telemetry

---

#### Epic DF-3: Dispatch Dashboard API

**Task DF-3.1**: Create dispatch dashboard summary endpoint

**Location**: `internal/domains/trip/handlers/dispatch_dashboard_handler.go` (NEW)

```go
package handlers

type DispatchDashboardHandler struct {
    service TripServiceInterface
    logger  *zap.Logger
}

func (h *DispatchDashboardHandler) HandleGetDashboard(w http.ResponseWriter, r *http.Request) {
    // 1. Extract org_id from JWT
    // 2. Get trip summary counts (pending, in-transit, delivered, failed, cancelled)
    // 3. Get active trips list with ETA
    // 4. Get alerts (delayed trips, offline vehicles)
    // 5. Get vehicle/driver utilization stats
    // 6. Return JSON response
    // Performance: <1s P95
}
```

**Repository Queries** (in `trip/repository/`):
```sql
-- Trip summary counts
SELECT 
    status,
    COUNT(*) as count
FROM trips
WHERE org_id = $1
  AND DATE(created_at) = CURRENT_DATE
  AND deleted_at IS NULL
GROUP BY status;

-- Active trips
SELECT 
    t.id,
    t.trip_number,
    t.status,
    v.registration as vehicle_registration,
    d.name as driver_name,
    t.drop_name,
    t.drop_expected,
    t.drop_latitude,
    t.drop_longitude
FROM trips t
JOIN vehicles v ON v.id = t.vehicle_id
JOIN drivers d ON d.id = t.driver_id
WHERE t.org_id = $1
  AND t.status IN ('pending', 'in-transit')
  AND t.deleted_at IS NULL
ORDER BY t.drop_expected ASC;

-- Vehicle utilization
SELECT 
    COUNT(*) as total_vehicles,
    COUNT(CASE WHEN status = 'active' THEN 1 END) as active_vehicles,
    COUNT(CASE WHEN current_driver_id IS NOT NULL THEN 1 END) as assigned_vehicles
FROM vehicles
WHERE org_id = $1
  AND deleted_at IS NULL;
```

**Deliverables**:
- [ ] `trip/handlers/dispatch_dashboard_handler.go`
- [ ] Dashboard queries in `trip/repository/`
- [ ] ETA calculation logic (distance / avg speed)
- [ ] Delay detection (>15 min behind schedule)
- [ ] Unit tests (>90% coverage)
- [ ] Performance test with 100 trips

---

#### Epic DF-4: Delivery Receipt Generation

**Task DF-4.1**: Create receipt generator service

**Location**: `internal/domains/trip/service/receipt_generator.go` (NEW)

```go
package service

type ReceiptGenerator struct {
    db       *sql.DB
    logger   *zap.Logger
    tpl      *template.Template // HTML template
}

func (rg *ReceiptGenerator) Generate(ctx context.Context, tripID string) error {
    // 1. Fetch trip data with all details
    trip, err := rg.fetchTripDetails(ctx, tripID)
    if err != nil {
        return fmt.Errorf("fetch trip: %w", err)
    }
    
    // 2. Generate receipt number
    receiptNumber := rg.generateReceiptNumber(trip.OrgID)
    
    // 3. Render HTML template
    html, err := rg.renderTemplate(trip, receiptNumber)
    if err != nil {
        return fmt.Errorf("render template: %w", err)
    }
    
    // 4. Convert to PDF (use existing billing/pdf engine)
    pdf, err := billingPDF.ConvertHTMLToPDF(html)
    if err != nil {
        return fmt.Errorf("convert to PDF: %w", err)
    }
    
    // 5. Save to disk
    filePath := fmt.Sprintf("receipts/%s/%s.pdf", 
        time.Now().Format("2006/01"), receiptNumber)
    err = os.MkdirAll(filepath.Dir(filePath), 0755)
    if err != nil {
        return fmt.Errorf("create directory: %w", err)
    }
    err = os.WriteFile(filePath, pdf, 0644)
    if err != nil {
        return fmt.Errorf("save PDF: %w", err)
    }
    
    // 6. Update trip record with receipt URL
    receiptURL := fmt.Sprintf("/api/v1/trips/%s/receipt.pdf", tripID)
    err = rg.updateTripReceiptURL(ctx, tripID, receiptURL)
    if err != nil {
        return fmt.Errorf("update trip: %w", err)
    }
    
    // 7. Log event
    err = rg.logTripEvent(ctx, tripID, "receipt_generated", map[string]interface{}{
        "receipt_number": receiptNumber,
        "file_path": filePath,
    })
    
    rg.logger.Info("Receipt generated", "trip_id", tripID, "receipt", receiptNumber)
    return nil
}
```

**Task DF-4.2**: Create receipt download handler

**Location**: `internal/domains/trip/handlers/receipt_handler.go` (NEW)

```go
func (h *ReceiptHandler) HandleDownloadReceipt(w http.ResponseWriter, r *http.Request) {
    tripID := chi.URLParam(r, "id")
    
    // 1. Fetch trip record
    trip, err := h.service.GetTripByID(r.Context(), tripID)
    if err != nil {
        http.Error(w, "Trip not found", http.StatusNotFound)
        return
    }
    
    // 2. Check if receipt exists
    if trip.ReceiptURL == "" {
        http.Error(w, "Receipt not yet generated", http.StatusAccepted)
        return
    }
    
    // 3. Read PDF file
    pdfData, err := os.ReadFile(trip.ReceiptURL)
    if err != nil {
        http.Error(w, "Failed to read receipt", http.StatusInternalServerError)
        return
    }
    
    // 4. Serve PDF with correct headers
    w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", 
        fmt.Sprintf("attachment; filename=\"%s-Receipt.pdf\"", trip.TripNumber))
    w.Write(pdfData)
}
```

**Task DF-4.3**: Trigger receipt on trip completion

**Location**: `internal/domains/trip/service/trip_service.go` (modify `Complete` method)

```go
func (s *TripService) Complete(ctx context.Context, id, driverID string, req *CompleteTripRequest) (*Trip, error) {
    // EXISTING: Validate, update status, log event
    
    // NEW: Trigger receipt generation (async)
    go func() {
        ctx := context.Background()
        err := s.receiptGenerator.Generate(ctx, id)
        if err != nil {
            s.logger.Error("Failed to generate receipt", "trip_id", id, "error", err)
        }
    }()
    
    // NEW: Trigger invoice creation (if auto-invoice enabled)
    if trip.AutoComplete && trip.InvoiceAutoCreate {
        go func() {
            ctx := context.Background()
            err := s.invoiceTriggerer.CreateInvoiceFromTrip(ctx, id)
            if err != nil {
                s.logger.Error("Failed to create invoice", "trip_id", id, "error", err)
            }
        }()
    }
    
    // EXISTING: Return updated trip
}
```

**Deliverables**:
- [ ] `trip/service/receipt_generator.go`
- [ ] `trip/handlers/receipt_handler.go`
- [ ] HTML template for receipt (in `internal/domains/trip/templates/`)
- [ ] Reuse `billing/pdf/` engine for PDF conversion
- [ ] Trip completion integration (auto-trigger)
- [ ] Download endpoint (`GET /api/v1/trips/{id}/receipt.pdf`)
- [ ] Unit tests (>85% coverage)

---

#### Epic DF-5: Exception Handling

**Task DF-5.1**: Create exception reporting endpoint

**Location**: `internal/domains/trip/handlers/exception_handler.go` (NEW)

```go
package handlers

func (h *ExceptionHandler) HandleReportException(w http.ResponseWriter, r *http.Request) {
    // 1. Parse exception report request
    // 2. Validate exception type and reason code
    // 3. Update trip status to 'exception' or add exception event
    // 4. Calculate new ETA if delay
    // 5. Trigger notifications (customer SMS, dispatcher alert)
    // 6. Return confirmation
}
```

**Exception Types** (in `trip/models/`):
```go
type ExceptionType string

const (
    ExceptionDelayTraffic      ExceptionType = "DELAY_TRAFFIC"
    ExceptionDelayBreakdown    ExceptionType = "DELAY_VEHICLE_BREAKDOWN"
    ExceptionDelayWeather      ExceptionType = "DELAY_WEATHER"
    ExceptionFailUnavailable   ExceptionType = "FAIL_CUSTOMER_UNAVAILABLE"
    ExceptionFailWrongAddress  ExceptionType = "FAIL_WRONG_ADDRESS"
    ExceptionFailDamaged       ExceptionType = "FAIL_GOODS_DAMAGED"
    ExceptionRejectRefused     ExceptionType = "REJECT_CUSTOMER_REFUSED"
    ExceptionRejectPartial     ExceptionType = "REJECT_PARTIAL_DELIVERY"
    ExceptionCancelCustomer    ExceptionType = "CANCEL_CUSTOMER_REQUEST"
    ExceptionCancelDuplicate   ExceptionType = "CANCEL_DUPLICATE"
)
```

**Deliverables**:
- [ ] `trip/handlers/exception_handler.go`
- [ ] Exception types and reason codes in `trip/models/`
- [ ] Notification trigger (NATS event to alert service)
- [ ] ETA recalculation logic
- [ ] Unit tests for all exception types

---

#### Epic DF-6: Billing Integration

**Task DF-6.1**: Create invoice trigger from trip completion

**Location**: `internal/domains/trip/service/invoice_triggerer.go` (NEW)

```go
package service

type InvoiceTriggerer struct {
    db            *sql.DB
    billingSvc    BillingServiceInterface
    logger        *zap.Logger
    pricingCalc   *PricingCalculator
}

func (it *InvoiceTriggerer) CreateInvoiceFromTrip(ctx context.Context, tripID string) error {
    // 1. Fetch trip details
    trip, err := it.fetchTripDetails(ctx, tripID)
    if err != nil {
        return fmt.Errorf("fetch trip: %w", err)
    }
    
    // 2. Calculate charges
    charges, err := it.pricingCalc.Calculate(ctx, trip)
    if err != nil {
        return fmt.Errorf("calculate charges: %w", err)
    }
    
    // 3. Create invoice via billing service
    invoiceID, err := it.billingSvc.CreateInvoiceFromTrip(ctx, trip, charges)
    if err != nil {
        return fmt.Errorf("create invoice: %w", err)
    }
    
    // 4. Update trip record with invoice_id
    err = it.updateTripInvoiceID(ctx, tripID, invoiceID)
    if err != nil {
        return fmt.Errorf("update trip: %w", err)
    }
    
    // 5. Log event
    err = it.logTripEvent(ctx, tripID, "invoice_created", map[string]interface{}{
        "invoice_id": invoiceID,
    })
    
    it.logger.Info("Invoice created from trip", "trip_id", tripID, "invoice_id", invoiceID)
    return nil
}
```

**Task DF-6.2**: Create pricing calculator

**Location**: `internal/domains/trip/service/pricing_calculator.go` (NEW)

```go
package service

type PricingCalculator struct {
    db     *sql.DB
    logger *zap.Logger
}

func (pc *PricingCalculator) Calculate(ctx context.Context, trip *Trip) (*TripCharges, error) {
    // 1. Fetch org pricing config from trip_settings
    config, err := pc.fetchOrgPricingConfig(ctx, trip.OrgID)
    if err != nil {
        return nil, fmt.Errorf("fetch pricing config: %w", err)
    }
    
    // 2. Calculate base rate
    baseRate := config.RatePerKm * trip.ActualDistanceKm
    
    // 3. Calculate fuel surcharge
    fuelSurcharge := config.FuelSurchargePerKm * trip.ActualDistanceKm
    
    // 4. Calculate waiting charges (if delivery took longer than expected)
    var waitingCharges float64
    if trip.ActualDurationMin > trip.EstimatedDurationMin {
        extraHours := float64(trip.ActualDurationMin-trip.EstimatedDurationMin) / 60.0
        waitingCharges = config.WaitingRatePerHour * extraHours
    }
    
    // 5. Calculate subtotal
    subtotal := baseRate + fuelSurcharge + waitingCharges
    
    // 6. Calculate GST
    gst := subtotal * config.GSTPercent
    
    // 7. Calculate total
    total := subtotal + gst
    
    return &TripCharges{
        BaseRate:          baseRate,
        FuelSurcharge:     fuelSurcharge,
        WaitingCharges:    waitingCharges,
        Subtotal:          subtotal,
        GST:               gst,
        GSTPercent:        config.GSTPercent,
        Total:             total,
    }, nil
}
```

**Deliverables**:
- [ ] `trip/service/invoice_triggerer.go`
- [ ] `trip/service/pricing_calculator.go`
- [ ] Org pricing config schema (in `trip_settings`)
- [ ] Integration with `billing/service/` CreateInvoiceFromTrip method
- [ ] Enhance billing service to accept trip data
- [ ] Unit tests for all pricing scenarios
- [ ] Integration test: trip complete → invoice created

---

#### Epic DF-7: Trip Event Timeline

**Task DF-7.1**: Create trip events table and repository

**Location**: `migrations/302_trip_events.sql` (NEW)

```sql
-- Trip events (audit trail + timeline)
CREATE TABLE IF NOT EXISTS trip_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id     UUID NOT NULL REFERENCES trips(id),
    event_type  VARCHAR(50) NOT NULL CHECK (event_type IN (
        'trip_created',
        'trip_started',
        'trip_completed',
        'trip_cancelled',
        'trip_failed',
        'status_change',
        'vehicle_assigned',
        'driver_assigned',
        'route_updated',
        'delay_reported',
        'exception_reported',
        'proof_uploaded',
        'receipt_generated',
        'invoice_created',
        'customer_notified'
    )),
    event_data  JSONB,
    triggered_by VARCHAR(50) CHECK (triggered_by IN ('user', 'system_auto_detect', 'scheduled_job')),
    user_id     UUID REFERENCES users(id),
    location    GEOMETRY(POINT, 4326),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_trip_events_trip ON trip_events (trip_id, created_at DESC);
```

**Deliverables**:
- [ ] Migration file for `trip_events` table
- [ ] Repository methods: `InsertTripEvent`, `GetTripEvents`
- [ ] Timeline API endpoint: `GET /api/v1/trips/{id}/events`
- [ ] Unit tests for event logging

---

### 4.3 Route Registration (Dispatch Flow)

**File**: `cmd/main.go`

Add these routes to the trip router section:

```go
// Trip/Dispatch routes
tripRouter := chi.NewRouter()
tripRouter.Use(authMiddleware.Authenticate)
tripRouter.Use(governanceMiddleware.RequireOrgContext)

// Existing routes
tripRouter.Post("/", tripHandler.CreateTrip)
tripRouter.Get("/", tripHandler.ListTrips)
tripRouter.Get("/{id}", tripHandler.GetTrip)
tripRouter.Patch("/{id}", tripHandler.UpdateTrip)
tripRouter.Delete("/{id}", tripHandler.DeleteTrip)
tripRouter.Patch("/{id}/start", tripHandler.StartTrip)
tripRouter.Patch("/{id}/complete", tripHandler.CompleteTrip)
tripRouter.Patch("/{id}/cancel", tripHandler.CancelTrip)

// NEW: Dispatch Flow endpoints
tripRouter.Get("/dispatch/dashboard", dispatchDashboardHandler.HandleGetDashboard)
tripRouter.Get("/{id}/receipt.pdf", receiptHandler.HandleDownloadReceipt)
tripRouter.Get("/{id}/events", tripEventsHandler.HandleGetTripEvents)
tripRouter.PATCH("/{id}/status", tripHandler.UpdateTripStatus)
tripRouter.POST("/{id}/exception", exceptionHandler.HandleReportException)
tripRouter.POST("/bulk", bulkTripHandler.HandleBulkCreateTrips)
```

---

## 5. Database Migrations Required

### 5.1 Migration Index

| Migration # | File | Purpose | Priority |
|-------------|------|---------|----------|
| 301 | `301_tracking_performance_indexes.sql` | Tracking query optimization | P0 |
| 302 | `302_trip_events.sql` | Trip event timeline table | P0 |
| 303 | `303_trip_status_history.sql` | Status change analytics | P1 |
| 304 | `303_delivery_receipts.sql` | Receipt metadata table | P1 |
| 305 | `305_trip_dispatch_columns.sql` | Add dispatch fields to trips table | P0 |
| 306 | `306_trip_pricing_config.sql` | Org pricing config for billing | P1 |

### 5.2 Migration 305: Trip Table Enhancements

**File**: `migrations/305_trip_dispatch_columns.sql`

```sql
-- Add dispatch workflow columns to trips table
ALTER TABLE trips ADD COLUMN IF NOT EXISTS trip_type VARCHAR(20) DEFAULT 'delivery' 
    CHECK (trip_type IN ('delivery', 'pickup', 'transfer', 'return'));

ALTER TABLE trips ADD COLUMN IF NOT EXISTS priority VARCHAR(10) DEFAULT 'medium' 
    CHECK (priority IN ('low', 'medium', 'high', 'critical'));

ALTER TABLE trips ADD COLUMN IF NOT EXISTS trip_number VARCHAR(50) UNIQUE;

-- Pickup contact fields
ALTER TABLE trips ADD COLUMN IF NOT EXISTS pickup_name VARCHAR(200);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS pickup_contact_name VARCHAR(100);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS pickup_contact_phone VARCHAR(20);

-- Drop contact fields
ALTER TABLE trips ADD COLUMN IF NOT EXISTS drop_name VARCHAR(200);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS drop_contact_name VARCHAR(100);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS drop_contact_phone VARCHAR(20);

-- Goods fields
ALTER TABLE trips ADD COLUMN IF NOT EXISTS goods_description TEXT;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS goods_weight_kg DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS goods_volume_m3 DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS goods_pieces INTEGER;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS goods_value_inr DECIMAL(12, 2);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS goods_invoice_number VARCHAR(50);

-- Customer linkage
ALTER TABLE trips ADD COLUMN IF NOT EXISTS customer_id UUID REFERENCES customers(id);

-- Automation settings
ALTER TABLE trips ADD COLUMN IF NOT EXISTS auto_start BOOLEAN DEFAULT true;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS auto_complete BOOLEAN DEFAULT true;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS notify_customer_on_start BOOLEAN DEFAULT false;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS notify_customer_on_complete BOOLEAN DEFAULT false;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS alert_dispatcher_on_delay BOOLEAN DEFAULT true;

-- Completion metadata
ALTER TABLE trips ADD COLUMN IF NOT EXISTS completion_reason VARCHAR(50);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS completion_notes TEXT;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS proof_photos JSONB;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS proof_signature TEXT;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS customer_acknowledged BOOLEAN DEFAULT FALSE;

-- Receipt & Invoice
ALTER TABLE trips ADD COLUMN IF NOT EXISTS receipt_url TEXT;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS invoice_id UUID REFERENCES billing_invoices(id);

-- Exception handling
ALTER TABLE trips ADD COLUMN IF NOT EXISTS exception_reason VARCHAR(100);
ALTER TABLE trips ADD COLUMN IF NOT EXISTS exception_notes TEXT;

-- Route estimates
ALTER TABLE trips ADD COLUMN IF NOT EXISTS estimated_distance_km DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS actual_distance_km DOUBLE PRECISION;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS estimated_duration_min INTEGER;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS actual_duration_min INTEGER;

-- Create trip number sequence
CREATE SEQUENCE IF NOT EXISTS trip_number_2026 START 1;

-- Generate trip numbers for existing rows
DO $$
DECLARE
    r RECORD;
    seq INTEGER;
BEGIN
    FOR r IN SELECT id FROM trips WHERE trip_number IS NULL ORDER BY created_at LOOP
        SELECT nextval('trip_number_2026') INTO seq;
        UPDATE trips SET trip_number = format('TRIP-2026-%05d', seq) WHERE id = r.id;
    END LOOP;
END $$;

-- Add indexes for dispatch queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_org_status_date 
    ON trips (org_id, status, DATE(created_at));

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_vehicle_active 
    ON trips (vehicle_id, status) 
    WHERE status IN ('pending', 'in-transit');

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_driver_active 
    ON trips (driver_id, status) 
    WHERE status IN ('pending', 'in-transit');

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_trip_number 
    ON trips (trip_number);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_trips_customer 
    ON trips (customer_id, created_at DESC);
```

---

## 6. Configuration & Feature Flags

### 6.1 Config Structure

**File**: `internal/config/config.go` (add these fields)

```go
type Config struct {
    // EXISTING fields...
    
    // Dispatch Flow Configuration
    AutoStatusDetectionEnabled    bool          `toml:"auto_status_detection_enabled"`
    AutoReceiptGenerationEnabled bool          `toml:"auto_receipt_generation_enabled"`
    AutoInvoiceCreationEnabled   bool          `toml:"auto_invoice_creation_enabled"`
    GeofenceCheckInterval        time.Duration `toml:"geofence_check_interval"`
    PickupProximityThreshold     float64       `toml:"pickup_proximity_threshold_m"`
    DropProximityThreshold       float64       `toml:"drop_proximity_threshold_m"`
    StopDetectionDuration        time.Duration `toml:"stop_detection_duration"`
    MaxTripsPerDayPerOrg         int           `toml:"max_trips_per_day_per_org"`
    
    // WebSocket Configuration
    WebSocketEnabled             bool          `toml:"websocket_enabled"`
    WebSocketMaxConnections      int           `toml:"websocket_max_connections"`
    WebSocketHeartbeatInterval   time.Duration `toml:"websocket_heartbeat_interval"`
    
    // Receipt Configuration
    ReceiptStoragePath           string        `toml:"receipt_storage_path"`
    ReceiptTemplatePath          string        `toml:"receipt_template_path"`
    
    // Notification Configuration
    CustomerNotificationsEnabled bool          `toml:"customer_notifications_enabled"`
    SMSProvider                  string        `toml:"sms_provider"` // msg91, twilio
}
```

### 6.2 Feature Flags (Database-Driven)

**File**: `migrations/307_dispatch_feature_flags.sql`

```sql
-- Insert dispatch feature flags into settings table
INSERT INTO settings (key, value, description) VALUES
    ('dispatch.enabled', 'true', 'Master kill switch for dispatch feature'),
    ('dispatch.auto_status_detection', 'true', 'Automatic trip status detection from GPS'),
    ('dispatch.receipt_generation', 'true', 'Auto-generate delivery receipts'),
    ('dispatch.invoice_creation', 'false', 'Auto-create invoices from deliveries'),
    ('dispatch.customer_notifications', 'false', 'Send SMS to customers'),
    ('dispatch.bulk_operations', 'false', 'Enable bulk trip creation'),
    ('dispatch.max_trips_per_day', '200', 'Maximum trips per org per day')
ON CONFLICT (key) DO NOTHING;
```

---

## 7. Testing Strategy

### 7.1 Unit Tests Required

| Component | File | Coverage Target | Key Scenarios |
|-----------|------|----------------|---------------|
| Vehicle Locations Handler | `tracking/handlers/vehicle_locations_handler_test.go` | >90% | Valid request, empty org, 500 vehicles |
| WebSocket Handler | `tracking/handlers/websocket_handler_test.go` | >80% | Connect, subscribe, disconnect, message format |
| History Trail Handler | `tracking/handlers/history_trail_handler_test.go` | >85% | 24h query, empty result, downsampling |
| Auto Status Detector | `trip/service/auto_status_detector_test.go` | >90% | Auto-start, auto-complete, edge cases |
| Receipt Generator | `trip/service/receipt_generator_test.go` | >85% | Valid trip, missing data, PDF generation |
| Pricing Calculator | `trip/service/pricing_calculator_test.go` | >95% | All charge types, GST, edge cases |
| Dispatch Dashboard | `trip/handlers/dispatch_dashboard_handler_test.go` | >90% | Summary counts, active trips, alerts |
| Exception Handler | `trip/handlers/exception_handler_test.go` | >85% | All exception types, notifications |
| Invoice Triggerer | `trip/service/invoice_triggerer_test.go` | >90% | Valid trip, billing error, retry |
| Geofence Checker | `geofencing/worker/geofence_checker_test.go` | >85% | Entry, exit, boundary, throttle |

### 7.2 Integration Tests Required

| Test | Description | Validation |
|------|-------------|------------|
| End-to-End Trip Lifecycle | Create → auto-start → auto-complete → receipt → invoice | All steps succeed, events logged |
| Vehicle Conflict Detection | Assign vehicle to trip while on another trip | Rejected with clear error |
| Driver Conflict Detection | Assign driver to trip while on another trip | Rejected with clear error |
| Multi-Tenant Isolation | Two orgs create trips | Each org sees only their trips |
| WebSocket Fan-Out | 10 position updates via telemetry | All WS clients receive within 5s |
| Geofence Alert Delivery | Vehicle crosses geofence | Alert logged, notification sent |
| Receipt Generation Trigger | Complete trip manually | PDF created within 3s |
| Invoice Auto-Creation | Complete trip with auto-invoice enabled | Invoice linked to trip |

### 7.3 Performance Tests Required

| Test | Load | Target |
|------|------|--------|
| Vehicle Locations API | 500 vehicles | <2s P95 |
| Trip Creation | 10 concurrent users | <1s P95 |
| Trip List Query | 100 trips today | <500ms P95 |
| Dispatch Dashboard | 20 concurrent dispatchers | <1s P95 |
| History Trail Query | 1,440 points (24h) | <3s P95 |
| WebSocket Fan-Out | 100 connections, 500 vehicles | <5s update latency |
| Receipt Generation | 50 concurrent completions | <3s P95 each |
| Geofence Detection | 500 vehicles × 10 geofences | <10s lag |

---

## 8. Implementation Timeline

### Phase 1: Foundation (Weeks 1-2)

**Vehicle Tracking**:
- [ ] Migration 301 (tracking indexes)
- [ ] Vehicle Locations API (VT-1)
- [ ] History Trail API (VT-3)
- [ ] Unit tests for both

**Dispatch Flow**:
- [ ] Migration 305 (trip table enhancements)
- [ ] Trip creation enhancement (DF-1)
- [ ] Vehicle/driver availability validation
- [ ] Unit tests

### Phase 2: Real-time & Auto-Detect (Weeks 3-4)

**Vehicle Tracking**:
- [ ] WebSocket Handler (VT-2)
- [ ] Redis pub/sub integration
- [ ] Load test: 100 WS connections

**Dispatch Flow**:
- [ ] Auto Status Detector (DF-2)
- [ ] Integration with telemetry ingest
- [ ] Trip Event Timeline (DF-7)
- [ ] Integration tests

### Phase 3: Dashboard & Receipts (Weeks 5-6)

**Dispatch Flow**:
- [ ] Dispatch Dashboard API (DF-3)
- [ ] Receipt Generator (DF-4)
- [ ] Receipt download handler
- [ ] Trip completion integration
- [ ] Unit + integration tests

### Phase 4: Advanced Features (Weeks 7-8)

**Vehicle Tracking**:
- [ ] Geofence Auto-Detection (VT-4)
- [ ] PostGIS spatial queries
- [ ] Alert integration

**Dispatch Flow**:
- [ ] Exception Handling (DF-5)
- [ ] Billing Integration (DF-6)
- [ ] Pricing Calculator
- [ ] Invoice auto-creation
- [ ] End-to-end integration tests

### Phase 5: Polish & Performance (Weeks 9-10)

- [ ] All performance tests
- [ ] Optimize slow queries
- [ ] Feature flags
- [ ] Documentation
- [ ] Beta deployment

---

## 9. File Inventory

### 9.1 New Files to Create (48 files)

#### Vehicle Tracking (18 files)

| File Path | Purpose | Epic |
|-----------|---------|------|
| `tracking/handlers/vehicle_locations_handler.go` | GET /vehicles/locations | VT-1 |
| `tracking/handlers/history_trail_handler.go` | GET history trail | VT-3 |
| `tracking/handlers/websocket_handler.go` | WebSocket live updates | VT-2 |
| `tracking/handlers/vehicle_locations_handler_test.go` | Unit tests | VT-1 |
| `tracking/handlers/history_trail_handler_test.go` | Unit tests | VT-3 |
| `tracking/handlers/websocket_handler_test.go` | Unit tests | VT-2 |
| `tracking/service/status_detector.go` | Geofence status logic | VT-4 |
| `tracking/service/history_service.go` | History query service | VT-3 |
| `tracking/repository/vehicle_locations.go` | Vehicle location queries | VT-1 |
| `tracking/repository/history_trail.go` | History trail queries | VT-3 |
| `tracking/worker/geofence_checker.go` | Geofence detection worker | VT-4 |
| `tracking/worker/geofence_checker_test.go` | Unit tests | VT-4 |
| `migrations/301_tracking_performance_indexes.sql` | DB indexes | VT-5 |
| `migrations/302_trip_events.sql` | Trip events table | DF-7 |
| `migrations/303_trip_status_history.sql` | Status history table | DF-7 |
| `migrations/304_delivery_receipts.sql` | Receipt metadata | DF-4 |
| `migrations/305_trip_dispatch_columns.sql` | Trip table enhancements | DF-1 |
| `migrations/306_trip_pricing_config.sql` | Pricing config | DF-6 |

#### Dispatch Flow (30 files)

| File Path | Purpose | Epic |
|-----------|---------|------|
| `trip/handlers/dispatch_dashboard_handler.go` | Dashboard API | DF-3 |
| `trip/handlers/receipt_handler.go` | Receipt download | DF-4 |
| `trip/handlers/exception_handler.go` | Exception reporting | DF-5 |
| `trip/handlers/dispatch_dashboard_handler_test.go` | Unit tests | DF-3 |
| `trip/handlers/receipt_handler_test.go` | Unit tests | DF-4 |
| `trip/handlers/exception_handler_test.go` | Unit tests | DF-5 |
| `trip/service/auto_status_detector.go` | Auto status detection | DF-2 |
| `trip/service/auto_status_detector_test.go` | Unit tests | DF-2 |
| `trip/service/receipt_generator.go` | Receipt PDF generation | DF-4 |
| `trip/service/receipt_generator_test.go` | Unit tests | DF-4 |
| `trip/service/pricing_calculator.go` | Charge calculation | DF-6 |
| `trip/service/pricing_calculator_test.go` | Unit tests | DF-6 |
| `trip/service/invoice_triggerer.go` | Invoice auto-creation | DF-6 |
| `trip/service/invoice_triggerer_test.go` | Unit tests | DF-6 |
| `trip/service/bulk_trip_service.go` | Bulk trip creation | DF-9 |
| `trip/models/dispatch_models.go` | Dispatch data models | DF-1 |
| `trip/models/dispatch_models_test.go` | Unit tests | DF-1 |
| `trip/repository/dispatch_queries.go` | Dashboard queries | DF-3 |
| `trip/repository/trip_events.go` | Event logging | DF-7 |
| `trip/repository/availability_check.go` | Vehicle/driver availability | DF-1 |
| `trip/templates/receipt.html` | Receipt HTML template | DF-4 |
| `internal/pkg/distance/haversine.go` | Distance calculation utility | DF-2 |
| `internal/pkg/distance/haversine_test.go` | Unit tests | DF-2 |
| `migrations/307_dispatch_feature_flags.sql` | Feature flags | Config |
| `docs/api/vehicle_tracking.yaml` | OpenAPI spec - tracking | Docs |
| `docs/api/dispatch_flow.yaml` | OpenAPI spec - dispatch | Docs |
| `tests/integration/trip_lifecycle_test.go` | E2E trip lifecycle test | Tests |
| `tests/integration/websocket_test.go` | WebSocket integration test | Tests |
| `tests/integration/geofence_test.go` | Geofence integration test | Tests |
| `tests/performance/dispatch_benchmarks_test.go` | Performance benchmarks | Tests |

### 9.2 Existing Files to Modify (15 files)

| File Path | Modification | Purpose |
|-----------|--------------|---------|
| `cmd/main.go` | Add routes | Register new endpoints |
| `internal/config/config.go` | Add config fields | Dispatch/Tracking config |
| `tracking/handlers/ingest_handler.go` | Add auto-detect trigger | Call status detector on telemetry |
| `tracking/service/tracking_service.go` | Add methods | Support vehicle locations |
| `trip/service/trip_service.go` | Enhance Create/Complete | Add dispatch workflow logic |
| `trip/models/trip.go` | Add fields | Dispatch workflow fields |
| `trip/handlers/trip_handler.go` | Add validation | Vehicle/driver availability |
| `trip/repository/trip_repo.go` | Add queries | Availability, dashboard |
| `billing/service/billing_service.go` | Add method | CreateInvoiceFromTrip |
| `geofencing/service/geofence_service.go` | Add methods | Alert integration |
| `geofencing/worker/` | Add worker | Geofence checker |
| `internal/gateway/routes.go` | Add routes | Route registration |
| `internal/middleware/auth.go` | Add WS auth | JWT for WebSocket |
| `go.mod` | Add dependency | WebSocket library |
| `config.toml` | Add config section | Dispatch/Tracking settings |

---

## 10. Dependencies & Integration Points

### 10.1 Internal Dependencies

| Dependency | Source | Target | Integration Method |
|-----------|--------|--------|-------------------|
| Telemetry Ingest | `tracking/ingest_handler` | `trip/auto_status_detector` | Function call |
| Trip Completion | `trip/trip_service` | `billing/billing_service` | NATS event or function call |
| Geofence Detection | `geofencing/worker` | `alerts/alert_service` | NATS event |
| WebSocket Updates | `tracking/websocket` | Redis pub/sub | Subscribe to `org:{id}:telemetry` |
| Receipt Generation | `trip/receipt_generator` | `billing/pdf` | Function call (reuse PDF engine) |
| Invoice Creation | `trip/invoice_triggerer` | `billing/billing_service` | Function call |

### 10.2 External Dependencies

| Dependency | Purpose | Status | Alternative |
|-----------|---------|--------|-------------|
| PostgreSQL + PostGIS | Spatial queries, TimescaleDB | ✅ Available | N/A |
| Redis | Pub/Sub for WebSocket, caching | ✅ Available | N/A |
| NATS JetStream | Event bus for notifications | ✅ Available | N/A |
| Go WebSocket library | WebSocket support | ❌ Need to add | `github.com/gorilla/websocket` |
| Go HTML templating | Receipt generation | ✅ stdlib | N/A |
| wkhtmltopdf or similar | HTML to PDF conversion | ⚠️ Check billing/pdf | Use billing engine |

---

## 11. Risk Register

| Risk | Probability | Impact | Mitigation | Owner |
|------|------------|--------|-----------|-------|
| WebSocket connection instability | Medium | High | Fallback to polling (10s interval) | Engineering |
| Auto status detection false positives | Medium | High | Require 2 consecutive detections, allow manual override | Engineering |
| Geofence detection performance degradation | Low | Medium | Spatial indexes, limit active geofences per org to 50 | Engineering |
| Receipt generation timeout | Low | Medium | Async background job, 5s timeout, retry on failure | Engineering |
| Trip table migration locks production | Low | Critical | Use `CONCURRENTLY` for indexes, test on staging first | Engineering |
| Billing integration creates duplicate invoices | Low | High | Idempotency check (trip.invoice_id IS NULL) | Engineering |
| Scope creep (multi-stop, routing) | High | Medium | Strict P0/P1/P2 enforcement, defer to Phase 2 | Product |

---

## 12. Success Criteria

### 12.1 Engineering Success Criteria

| Criteria | Target | Measurement |
|----------|--------|-------------|
| Unit test coverage | >85% average | `go test -cover` |
| Integration test pass rate | 100% | CI pipeline |
| Performance test pass rate | 100% | Benchmark suite |
| Zero breaking changes | 100% backward compat | Existing tests pass |
| Code review approval | 2+ approvals | PR process |
| Documentation complete | 100% of endpoints | OpenAPI spec generated |

### 12.2 Product Success Criteria (from Product Specs)

| Criteria | Target | Measurement |
|----------|--------|-------------|
| Map load time | <2s P95 | RUM monitoring |
| Position update latency | <5s P95 | WebSocket timestamps |
| Trip creation time | <3 min P95 | Form submission analytics |
| Auto status detection accuracy | >95% | Manual audit of 100 trips |
| Receipt generation success | 100% | Receipt logs |
| Beta customer adoption | 3/3 using daily | Usage analytics |

---

## 13. Next Steps

1. **Engineering Review**: Share this document with backend team for feedback
2. **Prioritization**: Confirm P0/P1/P2 epics with product team
3. **Sprint Planning**: Break epics into sprint-sized tasks
4. **Dependency Check**: Verify all external libraries are available
5. **Staging Environment**: Set up test database with PostGIS + TimescaleDB
6. **Start Implementation**: Begin with Phase 1 (Foundation) — Week 1

---

**Reviewers:**  
- Backend Engineering Lead  
- Product Manager  
- DevOps Engineer  

**Approved By:** (TBD)  
**Last Updated:** April 7, 2026  
**Next Review:** April 14, 2026
