# Vehicle Tracking — Product Specification

**Document ID:** PSPEC-VT-001
**Version:** 2.0 (Fleet-Core Aligned)
**Status:** Draft
**Created:** April 7, 2026
**Updated:** April 7, 2026
**Author:** Product Team
**Review Cycle:** Bi-weekly
**Parent PRD:** [FlyFleet Enterprise PRD](../docs/PRD.md)
**Priority:** P0 (Q3 Delivery)

---

## 0. Fleet-Core Architecture Alignment

### 0.1 What Already Exists in `services/fleet-core`

This product spec is written with full knowledge of the existing fleet-core monolith implementation.
Below is the accurate current state of the tracking domain:

#### Existing Tracking Endpoints (Already Live)

| Endpoint | Method | Handler | Status | Notes |
|----------|--------|---------|--------|-------|
| `/internal/v1/telemetry` | POST | `tracking/handlers/ingest_handler.go` | ✅ Live | AIS-140 aligned ingest with HMAC validation, per-device rate limiting (5Hz) |
| `/v1/vehicles/{id}/location` | GET | `tracking/handlers/tracking_query_handler.go` | ✅ Live | Single vehicle live location (Redis cache → DB fallback) |
| `/v1/vehicles/{id}/history` | GET | `tracking/handlers/tracking_query_handler.go` | ✅ Live | Paginated telemetry history with tenant isolation |
| `/v1/vehicles/{id}/replay` | GET | `tracking/handlers/tracking_query_handler.go` | ✅ Live | Ordered telemetry for map replay (max 24h, 1000 points) |
| `/v1/vehicles/nearby` | GET | `tracking/handlers/tracking_query_handler.go` | ✅ Live | Nearby vehicles by lat/lng/radius (cap 50km) |
| `/tracking/*` | GET | `tracking/handlers/analytics_handler.go` | ✅ Live | Fleet analytics |
| `/v1/devices/{id}` | Various | `tracking/handlers/device_handler.go` | ✅ Live | Device management |
| `/v1/devices/{id}/commands` | POST | `tracking/handlers/device_command_handler.go` | ✅ Live | Device commands (ping, reboot, immobilize) |

#### Existing Tracking Data Models (`tracking/models/`)

- **`TelemetryEvent`** — AIS-140 aligned GPS event (event_id, device_id, vehicle_id, lat, lng, speed, heading, ignition, fuel_level, odometer, sequence_no, recorded_at, received_at)
- **`LiveLocation`** — Redis-cached live state (vehicle_id, lat, lng, speed, heading, ignition, last_seen, status: online/offline with 5-min threshold)
- **`GPSTelemetry`** / **`OBDTelemetry`** / **`FuelTelemetry`** — Legacy batch formats (backward compat)
- **`GeofenceEvent`** — Entry/Exit/Dwell events (already modeled)
- **`AnomalyEvent`** — Overspeed, route deviation, tamper, panic, harsh event, fuel drop, teleport
- **`DeviceCommand`** — ping, reboot, immobilize, config_update

#### Existing Tracking Service Layer (`tracking/service/`)

- **`IngestionService`** — Validates, deduplicates, writes to DB + Redis live cache, publishes to NATS
- **`TelemetryQueryService`** — GetLiveLocation (Redis→DB fallback, 5-min offline threshold), GetHistory (paginated), GetReplay (ASC order, 24h cap), GetNearby (spatial query, 50km cap)
- **`AnomalyEngine`** — Detects overspeed, route deviation, tamper, teleport, fuel drop
- **`AlertService`** — Emits alert events for anomaly detection
- **`DeviceService`** — Device registry and management

#### Existing Tracking Repository Layer (`tracking/repository/`)

- **`TelemetryRepository`** — Write telemetry, GetLatestByVehicle, GetHistory (paginated DESC), GetNearby (spatial)
- **`LiveCacheRepository`** — Redis-backed live location cache (set/get with TTL)
- **`DeviceRepository`** / **`DeviceRegistryRepository`** — Device persistence
- **`OutboxRepository`** — Outbox pattern for reliable event publishing

#### Existing Tracking Workers (`tracking/worker/`)

- **`TelemetrySink`** — Background writer for telemetry batches
- **`OutboxRelay`** — Relays outbox events to NATS
- **`PartitionManager`** — TimescaleDB partition lifecycle management

#### Existing Geofencing Domain (`geofencing/`)

- **`engine/`** — Spatial detection engine (PostGIS-based)
- **`handlers/`** — Geofence CRUD handlers
- **`service/`** — Geofence management service
- **`repository/`** — Geofence storage and spatial queries
- **`models/`** — Geofence models (circle, polygon)

#### Existing Trip Domain (`trip/`)

- **`handlers/`** — Trip CRUD + offer management
- **`service/`** — Trip service with state machine (CREATED→SCHEDULED→STARTED→PAUSED→COMPLETED/FAILED/CANCELLED)
- **`repository/`** — Trip data access
- **`models/`** — Trip model with pickup/drop geo, waypoints, cargo details, India compliance (e-way bill, GST, AIS-140)

#### Existing Telemetry Tables

- **`telemetry_points`** — TimescaleDB hypertable (partitioned by month)
- **`gps_locations`** — Latest GPS positions per vehicle
- **`telemetry_events`** — Partitioned event log
- **`telemetry_analytics`** / **`telemetry_anomalies`** / **`telemetry_outbox`** — Analytics and anomaly storage
- **`geofences`** / **`geofence_assignments`** / **`geofence_events`** — Geofence data

### 0.2 What This Spec Adds (Gaps to Fill)

Based on the user story and acceptance criteria, here is what needs to be **built on top of** the existing infrastructure:

| Requirement | Existing Foundation | Gap to Fill |
|-------------|-------------------|-------------|
| Map loads ALL org vehicles in <2s | `GetLiveLocation` is single-vehicle only | **NEW**: Bulk vehicle locations endpoint (`GET /v1/vehicles/locations`) |
| Markers update every 5 seconds | Redis live cache exists, but no push mechanism | **NEW**: WebSocket handler for real-time push to dashboard |
| Click vehicle shows driver name, speed, status | `LiveLocation` has speed+status, but no driver join | **NEW**: Enriched response joining vehicles+drivers tables |
| Geofence alerts on enter/exit | `GeofenceEvent` model exists, engine exists | **NEW**: Continuous background worker, alert service integration, dispatcher notifications |
| 24h history trail | `GetReplay` exists (24h, 1000 point cap) | **ENHANCE**: Auto-resolution downsampling, distance/speed summary stats |

---

## 1. Executive Summary

### 1.1 Problem Statement

Fleet dispatchers in Indian logistics companies currently have **no real-time visibility** into vehicle locations. They rely on phone calls to drivers, WhatsApp check-ins, and manual paper logs to track where vehicles are. This creates:

- **2-3 hours/day wasted** on coordination calls/messages
- **30-45 minute response time** to emergencies (accidents, breakdowns)
- **Customer dissatisfaction** from inability to provide accurate ETAs
- **Revenue loss** from inefficient route utilization

### 1.2 Solution

A **real-time vehicle tracking dashboard** that shows all organization vehicles on an interactive map with live GPS positions, status indicators, and historical movement trails. Dispatchers get instant visibility into fleet operations without calling drivers.

### 1.3 Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Map load time | <2s P95 | Time from page load to all markers visible |
| Position update latency | <5s | Time from GPS ingest to map marker update |
| Daily active usage | >80% of logged-in dispatchers | Users who view map at least once per session |
| Dispatcher coordination time reduction | 50% reduction | Pre/post deployment survey |
| Emergency response time | <15 minutes | Time from incident to dispatcher awareness |

---

## 2. User Story

### Primary Story

```
As a dispatcher
I want to see all my vehicles on a map with live positions
So that I know where every vehicle is right now
```

### Acceptance Criteria

- [ ] Map loads with all organization vehicles in <2s
- [ ] Vehicle markers update every 5 seconds
- [ ] Clicking a vehicle shows: driver name, speed, last update, status
- [ ] Geofence alerts when vehicle enters/exits zone
- [ ] Vehicle history trail for last 24 hours

### Extended Acceptance Criteria

- [ ] Map auto-fits to show all vehicles on initial load
- [ ] Vehicle markers are color-coded by status (moving=idle=stopped=offline)
- [ ] Search/filter vehicles by registration number, driver name, or status
- [ ] Clicking vehicle opens detail panel without leaving map view
- [ ] History trail toggle shows/hides 24h movement path
- [ ] Map supports zoom from country view to street-level detail
- [ ] Works on desktop (1920x1080) and tablet (1024x768) viewports

---

## 3. Scope & Prioritization

### 3.1 RICE Score

| Factor | Score | Justification |
|--------|-------|---------------|
| **Reach** | 500 | All beta customers (500 vehicles across 20 orgs) |
| **Impact** | 3 | Massive — core differentiator vs competitors |
| **Confidence** | 80% | Proven pattern (Leaflet + WebSocket), existing telemetry tables |
| **Effort** | 8 | 8 person-sprints (backend + frontend + integration) |
| **RICE** | **2,308** | (500 × 3 × 0.8) / 8 |

### 3.2 Epics Breakdown

| Epic | Priority | Effort | Dependencies |
|------|----------|--------|--------------|
| **E1: Map Foundation** | P0 | 3 sprints | IAM complete (Q2) |
| **E2: Real-time Updates** | P0 | 2 sprints | E1, telemetry ingest API |
| **E3: Vehicle Detail Panel** | P0 | 1 sprint | E1, driver-service integration |
| **E4: Geofence Alerts** | P1 | 3 sprints | E2, alert-service |
| **E5: History Trail** | P1 | 2 sprints | E2, telemetry query optimization |
| **E6: Performance Optimization** | P0 | 1 sprint | E1-E5 complete |

### 3.3 In Scope (MVP — Q3)

- ✅ Interactive map with Leaflet/MapLibre
- ✅ Vehicle markers with real-time position updates (5s interval)
- ✅ Vehicle detail panel (driver, speed, status, last update)
- ✅ Color-coded markers by vehicle status
- ✅ Search and filter vehicles
- ✅ Auto-fit map to show all vehicles
- ✅ Geofence entry/exit detection and alerts
- ✅ 24-hour history trail visualization
- ✅ WebSocket connection for live updates

### 3.4 Out of Scope (Future Phases)

- ❌ AI-powered route optimization (Phase 3)
- ❌ Predictive ETA calculation (Phase 3)
- ❌ Traffic-aware routing (Phase 3)
- ❌ Multi-stop trip visualization (Phase 2)
- ❌ Driver mobile app GPS reporting (Phase 2)
- ❌ Offline map support (Phase 2)
- ❌ 3D terrain view (Nice-to-have)
- ❌ Street view integration (Nice-to-have)

---

## 4. Technical Architecture

### 4.1 System Context

```
┌─────────────────────────────────────────────────────────────┐
│                    dashboard-web (React)                     │
│  ┌──────────────┐  ┌────────────────┐  ┌─────────────────┐ │
│  │ Map Component│  │ Vehicle Panel  │  │ WebSocket Client│ │
│  │ (Leaflet)    │◄─┤ (Slide-over)   │  │ (Live Updates)  │ │
│  └──────┬───────┘  └───────▲────────┘  └────────┬────────┘ │
│         │                  │                     │          │
└─────────┼──────────────────┼─────────────────────┼──────────┘
          │ HTTP REST        │ HTTP REST           │ WebSocket
          ▼                  ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│              Fleet-Core Monolith (Go / Chi)                  │
│  ┌──────────────┐  ┌────────────────┐  ┌─────────────────┐ │
│  │ Tracking API │  │ Vehicle API    │  │ WS Handler      │ │
│  │ (GET /telemetry)│ │ (GET /vehicles)│ │ (Push updates)  │ │
│  └──────┬───────┘  └───────▲────────┘  └────────┬────────┘ │
│         │                  │                     │          │
└─────────┼──────────────────┼─────────────────────┼──────────┘
          │                  │                     │
          ▼                  ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│              PostgreSQL 15 + TimescaleDB                     │
│  ┌──────────────────┐  ┌──────────────────────────────────┐ │
│  │ telemetry_points │  │ vehicles, drivers, geofences     │ │
│  │ (partitioned)    │  │ (RLS tenant-scoped)              │ │
│  └──────────────────┘  └──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
          ▲
          │ GPS Ingest (future: device → NATS → DB)
┌─────────┴──────────┐
│   GPS Devices /    │
│   Mobile App       │
└────────────────────┘
```

### 4.2 Data Flow

```
1. Vehicle GPS device sends position → POST /api/v1/telemetry
2. Fleet-core validates JWT, writes to telemetry_points (TimescaleDB)
3. WebSocket handler polls recent telemetry (last 5s) via Redis pub/sub
4. WebSocket pushes position update → dashboard-web
5. Leaflet marker animates to new position on map
```

### 4.3 API Contracts

> **Design Note:** All endpoints follow the existing fleet-core patterns:
> - JWT auth via JWKS middleware (`coremiddleware.NewJWKSAuthGate`)
> - Org context via `coremiddleware.TenantContextMiddleware`
> - Role checks via `coremiddleware.RequireRole`
> - Response format: JSON with `{data, meta}` structure
> - Registered in `cmd/main.go` under appropriate route groups

#### GET /api/v1/vehicles/locations (NEW)

**Purpose**: Fetch current positions of ALL vehicles in organization (bulk live location)

**Why not use existing `/v1/vehicles/{id}/location`?**
The existing endpoint is single-vehicle only. The dispatcher map needs ALL vehicles in one request for <2s load time.

**Request**:
```http
GET /api/v1/vehicles/locations
Authorization: Bearer <JWT via JWKS>
X-Org-ID: <org-uuid> (extracted from JWT by TenantContextMiddleware)
```

**Response** (200 OK):
```json
{
  "data": [
    {
      "vehicle_id": "v-123-abc",
      "registration": "DL01AB1234",
      "make": "Tata",
      "model": "Ace Gold",
      "latitude": 28.6139,
      "longitude": 77.2090,
      "speed_kmh": 45.2,
      "heading": 135,
      "ignition": true,
      "fuel_level": 72.5,
      "odometer": 45230.5,
      "status": "moving",
      "driver": {
        "id": "d-456-def",
        "name": "Rajesh Kumar",
        "phone": "+91 98765 43210"
      },
      "last_update": "2026-04-07T10:30:00Z",
      "last_seen_seconds_ago": 5
    }
  ],
  "meta": {
    "total": 42,
    "online": 38,
    "offline": 4,
    "timestamp": "2026-04-07T10:30:05Z"
  }
}
```

**Implementation** (in `tracking/handlers/vehicle_locations_handler.go`):
```go
package handlers

type VehicleLocationsHandler struct {
    queryService *service.TelemetryQueryService
    liveCache    repository.LiveCacheRepository
    vehicleRepo  vehicleRepo.VehicleRepository    // existing vehicle domain
    driverRepo   driverRepo.DriverRepository      // existing driver domain
    logger       *zap.Logger
}

func (h *VehicleLocationsHandler) HandleGetVehicleLocations(w http.ResponseWriter, r *http.Request) {
    orgID := middleware.GetOrgID(r.Context())
    
    // 1. Get all org vehicles from vehicle repo
    vehicles, err := h.vehicleRepo.GetOrgVehicles(r.Context(), orgID)
    
    // 2. Batch-fetch latest locations from Redis (or DB fallback)
    locations := h.liveCache.GetBatchLiveLocations(r.Context(), vehicleIDs)
    
    // 3. Enrich with driver info
    // 4. Calculate status (moving/idle/stopped/offline) from speed + ignition + age
    // 5. Return response
}
```

**Repository Query** (new method in `vehicle/repository/`):
```sql
SELECT 
    v.id,
    v.registration,
    v.make,
    v.model,
    v.current_driver_id,
    v.status as vehicle_status
FROM vehicles v
WHERE v.org_id = $1 
  AND v.deleted_at IS NULL
ORDER BY v.registration;
```

**Status Calculation Logic** (in service layer):
```go
func CalculateVehicleStatus(loc LiveLocation) string {
    age := time.Since(loc.LastSeen)
    
    if age > 5*time.Minute {
        return "offline"
    }
    if !loc.Ignition {
        return "stopped"
    }
    if loc.Speed < 2.0 {
        return "idle"
    }
    return "moving"
}
```

**Performance**: <2s P95 for 500 vehicles
- Redis batch get: ~50ms (500 keys via pipeline)
- Vehicle query: ~100ms (indexed on org_id)
- Driver enrichment: ~200ms (batch by IDs)
- JSON encode: ~50ms

---

#### GET /api/v1/vehicles/{id}/telemetry/history (ENHANCE existing `/v1/vehicles/{id}/replay`)

> **Note:** The existing `/v1/vehicles/{id}/replay` endpoint already provides ordered telemetry for map replay with a 24h limit and 1000-point cap. This spec **enhances** it with auto-resolution downsampling and summary statistics.

**Purpose**: Fetch 24-hour movement history for a vehicle with intelligent downsampling

**Enhancement to existing `GetReplay` in `tracking/handlers/tracking_query_handler.go`**:

**Request** (enhanced query params):
```http
GET /api/v1/vehicles/{id}/telemetry/history?start=2026-04-06T10:00:00Z&end=2026-04-07T10:00:00Z&resolution=auto
Authorization: Bearer <JWT via JWKS>
```

**New Query Parameters**:
| Param | Required | Description | Default |
|-------|----------|-------------|---------|
| start | Yes | ISO 8601 start time | — |
| end | Yes | ISO 8601 end time | — |
| resolution | No | `auto`, `10s`, `30s`, `60s`, `5m`, `15m` | `auto` |
| include_summary | No | Include distance/speed stats | `false` |

**Auto-Resolution Logic** (when `resolution=auto`):
| Time Range | Effective Resolution | Max Points |
|------------|---------------------|------------|
| < 1 hour | 10 seconds | 360 |
| 1-6 hours | 30 seconds | 720 |
| 6-24 hours | 60 seconds | 1,440 |
| 1-7 days | 5 minutes | 2,016 |
| 7-30 days | 15 minutes | 2,880 |

**Enhanced Response** (200 OK):
```json
{
  "data": {
    "vehicle_id": "v-123-abc",
    "points": [
      {
        "timestamp": "2026-04-06T10:00:00Z",
        "latitude": 28.6139,
        "longitude": 77.2090,
        "speed_kmh": 0,
        "heading": 135,
        "ignition": false
      },
      {
        "timestamp": "2026-04-06T10:01:00Z",
        "latitude": 28.6145,
        "longitude": 77.2098,
        "speed_kmh": 32.1,
        "heading": 140,
        "ignition": true
      }
    ],
    "summary": {
      "total_distance_km": 142.5,
      "moving_time_hours": 6.2,
      "idle_time_hours": 1.8,
      "stopped_time_hours": 16.0,
      "max_speed_kmh": 85.3,
      "avg_speed_kmh": 38.7,
      "harsh_events": 2,
      "overspeed_events": 1
    }
  },
  "meta": {
    "point_count": 1440,
    "resolution_used": "60s",
    "time_range_hours": 24
  }
}
```

**Implementation** (enhance existing `TelemetryQueryService.GetReplay`):
```go
// Add to tracking/service/telemetry_query_service.go
func (s *TelemetryQueryService) GetHistoryWithSummary(ctx context.Context, tenantID, vehicleID uuid.UUID, from, to time.Time, resolution string) (*HistoryTrailResponse, error) {
    // 1. Determine resolution
    effectiveRes := s.resolveResolution(from, to, resolution)
    
    // 2. Query TimescaleDB with time_bucket for downsampling
    points, err := s.telemRepo.GetHistoryBucket(ctx, tenantID, vehicleID, from, to, effectiveRes)
    
    // 3. Calculate summary stats (distance via Haversine, speed aggregates, time classification)
    summary := s.calculateSummary(points)
    
    return &HistoryTrailResponse{
        VehicleID: vehicleID,
        Points: points,
        Summary: summary,
        Meta: HistoryMeta{
            PointCount:    len(points),
            ResolutionUsed: effectiveRes,
            TimeRangeHours: to.Sub(from).Hours(),
        },
    }, nil
}
```

**TimescaleDB Query** (new method in `tracking/repository/telemetry_repository.go`):
```sql
-- Use existing GetHistory but add time_bucket downsampling
SELECT 
    time_bucket($1, time) as bucket_time,
    AVG(latitude) as latitude,
    AVG(longitude) as longitude,
    AVG(speed) as speed_kmh,
    AVG(heading) as heading,
    BOOL_OR(ignition) as ignition,
    MAX(speed) as max_speed,
    COUNT(*) as point_count
FROM telemetry_points
WHERE vehicle_id = $2
  AND time >= $3
  AND time <= $4
GROUP BY bucket_time
ORDER BY bucket_time;
```

**Performance**: <3s P95 for 24h data (1,440 points at 1-min resolution)

---

#### WebSocket /ws/telemetry (NEW)

**Purpose**: Real-time vehicle position updates to dashboard (push model instead of polling)

**Why needed?** The existing `LiveLocation` model is cached in Redis with a 5-minute offline threshold, but there's no **push mechanism** to notify the dashboard when positions update. The dispatcher would need to poll every 5 seconds. WebSocket eliminates polling overhead and provides sub-second update delivery.

**Connection**:
```
wss://api.flyfleet.io/ws/telemetry?token=<JWT>&org_id=<org-uuid>
```

**Implementation** (new `tracking/handlers/websocket_handler.go`):
```go
package handlers

import (
    "github.com/gorilla/websocket"
    "github.com/redis/go-redis/v9"
    "go.uber.org/zap"
)

type WebSocketHandler struct {
    redis   *redis.Client
    logger  *zap.Logger
    upgrader websocket.Upgrader
    clients sync.Map // orgID → []*WebSocketClient
}

type WebSocketClient struct {
    conn  *websocket.Conn
    orgID uuid.UUID
    send  chan []byte
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    // 1. Authenticate via JWT query param (reuse JWKS logic)
    token := r.URL.Query().Get("token")
    claims, err := verifyJWT(token)
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
    
    // 2. Upgrade HTTP to WebSocket
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    
    // 3. Register client for org
    orgID := claims.OrgID
    client := &WebSocketClient{conn: conn, orgID: orgID, send: make(chan []byte, 256)}
    h.registerClient(orgID, client)
    
    // 4. Start read/write loops
    go client.writePump()
    go client.readPump()
    
    // 5. Subscribe to Redis channel for org
    go h.subscribeToOrgUpdates(orgID, client)
}

// Server → Client Message
type PositionUpdateMessage struct {
    Type       string  `json:"type"` // "position_update"
    VehicleID  string  `json:"vehicle_id"`
    Latitude   float64 `json:"latitude"`
    Longitude  float64 `json:"longitude"`
    SpeedKmh   float64 `json:"speed_kmh"`
    Heading    int     `json:"heading"`
    Status     string  `json:"status"` // moving/idle/stopped/offline
    Timestamp  string  `json:"timestamp"`
}

// Client → Server Message (optional: subscribe to specific vehicles)
type SubscribeMessage struct {
    Type        string   `json:"type"` // "subscribe_vehicles"
    VehicleIDs  []string `json:"vehicle_ids"`
}
```

**Redis Pub/Sub Integration** (modify existing `tracking/service/ingestion_service.go`):
```go
// In IngestionService.Ingest(), after writing to DB + Redis cache:
func (s *IngestionService) PublishPositionUpdate(ctx context.Context, event *models.TelemetryEvent) {
    message, _ := json.Marshal(PositionUpdateMessage{
        Type:      "position_update",
        VehicleID: event.VehicleID.String(),
        Latitude:  event.Lat,
        Longitude: event.Lng,
        SpeedKmh:  event.Speed,
        Heading:   event.Heading,
        Timestamp: event.RecordedAt.Format(time.RFC3339),
    })
    
    // Publish to org-specific channel
    channel := fmt.Sprintf("org:%s:telemetry", event.TenantID.String())
    s.redis.Publish(ctx, channel, message).Err()
}
```

**Client Subscription Flow**:
```go
func (h *WebSocketHandler) subscribeToOrgUpdates(orgID uuid.UUID, client *WebSocketClient) {
    pubsub := h.redis.Subscribe(context.Background(), fmt.Sprintf("org:%s:telemetry", orgID))
    defer pubsub.Close()
    
    ch := pubsub.Channel()
    for msg := range ch {
        select {
        case client.send <- []byte(msg.Payload):
        default:
            // Client buffer full, drop message
            h.logger.Warn("Dropped WebSocket message", "org_id", orgID)
        }
    }
}
```

**Route Registration** (in `cmd/main.go`):
```go
// Add to tracking route group
r.Get("/ws/telemetry", websocketHandler.HandleWebSocket)
```

**Performance**: <5s P95 from telemetry ingest → WebSocket delivery to dashboard
- Telemetry ingest → Redis publish: ~10ms
- Redis → WebSocket delivery: ~50ms
- Total latency: <100ms (well within 5s target)

**Fallback**: If WebSocket connection fails, dashboard falls back to polling `GET /api/v1/vehicles/locations` every 10 seconds.

---

#### Geofence Endpoints (ENHANCE existing `geofencing/` domain)

> **Note:** The `geofencing/` domain already exists with handlers, service, repository, engine, and models for circle/polygon geofences. The `GeofenceEvent` model (ENTRY/EXIT/DWELL) is already defined in `tracking/models/`. This spec **integrates** geofence detection with live telemetry and adds dispatcher alert delivery.

**Existing Geofencing Endpoints** (already live):
| Endpoint | Method | Status | Notes |
|----------|--------|--------|-------|
| `/api/v1/geofences` | GET | ✅ Exists | List org geofences |
| `/api/v1/geofences` | POST | ✅ Exists | Create geofence |
| `/api/v1/geofences/{id}` | GET/PUT/DELETE | ✅ Exists | CRUD operations |

**New Integration**: Geofence detection triggered on telemetry ingest

When `tracking/service/ingestion_service.go` processes a new GPS position:
1. Call `geofencing/engine/` to check if point is inside any active geofence
2. Compare with previous state (stored in Redis: `geofence_state:{vehicle_id}:{geofence_id}` = inside/outside)
3. If state changed → create `GeofenceEvent` (ENTRY or EXIT)
4. Publish alert via NATS to notification service
5. Push to WebSocket for real-time dashboard notification

**Geofence Check Query** (existing `geofencing/engine/` uses PostGIS):
```sql
-- Polygon geofence: ST_Contains
SELECT g.id, g.name
FROM geofences g
WHERE g.org_id = $1
  AND g.status = 'active'
  AND g.type = 'polygon'
  AND ST_Contains(g.polygon, ST_SetSRID(ST_MakePoint($2, $3), 4326));

-- Circle geofence: ST_DWithin
SELECT g.id, g.name
FROM geofences g
WHERE g.org_id = $1
  AND g.status = 'active'
  AND g.type = 'circle'
  AND ST_DWithin(
      ST_SetSRID(ST_MakePoint($2, $3), 4326),
      ST_SetSRID(ST_MakePoint(g.center_lng, g.center_lat), 4326),
      g.radius_m
  );
```

**Alert Message Format** (published to NATS):
```json
{
  "alert_type": "geofence_entry",
  "vehicle": {
    "id": "v-123-abc",
    "registration": "DL01AB1234"
  },
  "geofence": {
    "id": "gf-001",
    "name": "Delhi Warehouse"
  },
  "occurred_at": "2026-04-07T10:30:00Z",
  "location": {
    "latitude": 28.6139,
    "longitude": 77.2090
  },
  "message": "Vehicle DL01AB1234 entered Delhi Warehouse at 10:30 AM"
}
```

---

### 4.4 Database Schema

#### Existing Tables (Leverage)

```sql
-- Telemetry points (TimescaleDB partitioned hypertable)
CREATE TABLE telemetry_points (
    time        TIMESTAMPTZ NOT NULL,
    vehicle_id  UUID NOT NULL REFERENCES vehicles(id),
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    speed_kmh   DOUBLE PRECISION,
    heading     DOUBLE PRECISION,
    altitude_m  DOUBLE PRECISION,
    ignition    BOOLEAN,
    fuel_level  DOUBLE PRECISION,
    raw_data    JSONB,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Partition by month for performance
SELECT create_hypertable('telemetry_points', 'time', chunk_time_interval => INTERVAL '1 month');

-- Vehicles (existing)
CREATE TABLE vehicles (
    id              UUID PRIMARY KEY,
    org_id          UUID NOT NULL REFERENCES organizations(id),
    registration    VARCHAR(20) NOT NULL,
    make            VARCHAR(50),
    model           VARCHAR(50),
    year            INTEGER,
    status          VARCHAR(20) DEFAULT 'active',
    current_driver_id UUID REFERENCES drivers(id),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Geofences (existing)
CREATE TABLE geofences (
    id          UUID PRIMARY KEY,
    org_id      UUID NOT NULL REFERENCES organizations(id),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    type        VARCHAR(20) CHECK (type IN ('circle', 'polygon')),
    center_lat  DOUBLE PRECISION,
    center_lng  DOUBLE PRECISION,
    radius_m    DOUBLE PRECISION,
    polygon     GEOMETRY(POLYGON, 4326),
    status      VARCHAR(20) DEFAULT 'active',
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Geofence events (existing)
CREATE TABLE geofence_events (
    id          UUID PRIMARY KEY,
    geofence_id UUID REFERENCES geofences(id),
    vehicle_id  UUID REFERENCES vehicles(id),
    event_type  VARCHAR(20) CHECK (event_type IN ('entry', 'exit')),
    occurred_at TIMESTAMPTZ NOT NULL,
    latitude    DOUBLE PRECISION,
    longitude   DOUBLE PRECISION,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
```

#### New Indexes Required

```sql
-- Optimize latest position query
CREATE INDEX idx_telemetry_latest ON telemetry_points (vehicle_id, time DESC);

-- Optimize geofence spatial queries
CREATE INDEX idx_geofences_spatial ON geofences USING GIST (polygon);
CREATE INDEX idx_geofences_org ON geofences (org_id, status);

-- Optimize geofence event queries
CREATE INDEX idx_geofence_events_vehicle ON geofence_events (vehicle_id, occurred_at DESC);
CREATE INDEX idx_geofence_events_org ON geofence_events (geofence_id, occurred_at DESC);
```

### 4.5 Performance Considerations

| Operation | Target | Strategy |
|-----------|--------|----------|
| Initial map load (<500 vehicles) | <2s P95 | Single query with latest telemetry JOIN vehicles |
| WebSocket update propagation | <5s P95 | Redis pub/sub, batch updates every 5s |
| 24h history trail fetch | <3s P95 | TimescaleDB time_bucket, downsample to 60s resolution |
| Geofence detection | <10s lag | Background job checking every 10s against recent telemetry |
| Map rendering (500 markers) | <100ms | Canvas-based marker rendering, not DOM |

---

## 5. UX Design Specification

### 5.1 Screen Layout

```
┌─────────────────────────────────────────────────────────────┐
│  🔍 Search vehicles...              [🔲 Filters] [⚙️ Settings]│
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                                                      │   │
│  │                    MAP VIEW                          │   │
│  │                                                      │   │
│  │         🚛 (moving)                                  │   │
│  │                  🚛 (idle)                           │   │
│  │                        🔴 (stopped)                  │   │
│  │   🚛 (moving)                   🚛 (idle)            │   │
│  │                                                      │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│  Vehicles: 42 total  │  🟢 Moving: 18  │  🟡 Idle: 12  │   │
│  🔴 Stopped: 8  │  ⚫ Offline: 4                            │
└─────────────────────────────────────────────────────────────┘
```

### 5.2 Vehicle Detail Panel (Slide-over)

```
┌─────────────────────────────┐
│  Vehicle Detail          ✕  │
├─────────────────────────────┤
│                             │
│  🚛 DL01AB1234              │
│  Tata Ace Gold              │
│                             │
│  Status:  🟢 Moving         │
│  Speed:   45 km/h           │
│  Heading: Southeast         │
│                             │
│  Driver:  Rajesh Kumar      │
│  Phone:   +91 98765 43210  │
│                             │
│  Last Update: 5s ago        │
│  Ignition:  ON              │
│  Fuel:      72%             │
│                             │
│  ────────────────────────   │
│  📍 Current Location        │
│  Near Connaught Place       │
│  New Delhi, 110001          │
│                             │
│  ────────────────────────   │
│  📊 Today's Summary         │
│  Distance: 142.5 km         │
│  Moving Time: 6h 12m        │
│  Idle Time: 1h 48m          │
│  Stops: 8                   │
│                             │
│  [📜 View History] [📞 Call Driver] │
│                             │
└─────────────────────────────┘
```

### 5.3 Interaction Flows

#### Flow 1: View Vehicle Details
```
User clicks vehicle marker on map
  → Slide-over panel opens from right
  → Fetches vehicle details via GET /api/v1/vehicles/{id}
  → Shows driver, speed, status, location, today's stats
  → Panel can be closed with ✕ or Escape key
```

#### Flow 2: View History Trail
```
User clicks "View History" in vehicle panel
  → Date/time picker appears
  → User selects time range (default: last 24h)
  → Map displays trail as polyline overlay
  → Play button animates vehicle movement through time
  → Speed control: 1x, 5x, 10x, 30x
```

#### Flow 3: Search & Filter
```
User types in search bar
  → Debounced search (300ms)
  → Matches: registration, driver name, vehicle type
  → Results highlight on map (pulse animation)
  → Filters: Status (Moving/Idle/Stopped/Offline), Vehicle Type
  → Multi-select filters combine with AND logic
```

### 5.4 Visual Design

#### Marker Colors by Status
| Status | Color | Hex | Icon |
|--------|-------|-----|------|
| Moving | Green | `#10B981` | 🚛 |
| Idle | Amber | `#F59E0B` | 🚛 |
| Stopped | Red | `#EF4444` | 🔴 |
| Offline | Gray | `#6B7280` | ⚫ |

#### Map Style
- **Provider**: MapLibre GL (open-source, no API costs)
- **Style**: Light base with muted roads, emphasis on vehicle markers
- **Zoom Levels**: 4 (India view) → 18 (street-level)
- **Default**: Fit bounds to show all vehicles

#### Animations
- Marker position updates: Smooth interpolation (500ms ease-in-out)
- Panel open/close: Slide from right (200ms)
- Search highlight: Pulse effect (3 cycles, 1s each)
- History trail playback: Configurable speed

---

## 6. Geofence Alerts

### 6.1 Geofence Types

| Type | Shape | Use Case |
|------|-------|----------|
| Circle | Center point + radius | Depot, customer location |
| Polygon | Multiple coordinates | Warehouse complex, city boundary |
| Route corridor | Polyline + buffer | Expected route adherence |

### 6.2 Alert Logic

```
For each vehicle position update:
  1. Check if point is inside any active geofence (PostGIS ST_Contains)
  2. Compare with previous state (inside vs outside)
  3. If state changed:
     - Entry: Log geofence_event (type=entry), trigger alert
     - Exit: Log geofence_event (type=exit), trigger alert
  4. Alert sent to alert-service via NATS
  5. Alert-service delivers via email/SMS/in-app per user preferences
```

### 6.3 Alert Message Format

```json
{
  "alert_type": "geofence_entry",
  "vehicle": {
    "id": "v-123-abc",
    "registration": "DL01AB1234"
  },
  "geofence": {
    "id": "gf-001",
    "name": "Delhi Warehouse"
  },
  "occurred_at": "2026-04-07T10:30:00Z",
  "location": {
    "latitude": 28.6139,
    "longitude": 77.2090
  },
  "message": "Vehicle DL01AB1234 entered Delhi Warehouse at 10:30 AM"
}
```

### 6.4 Alert Delivery Channels

| Channel | Priority | Use Case |
|---------|----------|----------|
| In-app notification | Real-time | All geofence events |
| Email | Within 5 min | Summary of events |
| SMS | Immediate | Critical geofences (configurable) |
| Webhook | Immediate | Third-party integrations |

---

## 7. History Trail

### 7.1 Trail Visualization

- **Rendering**: Polyline overlay on map (MapLibre Layer)
- **Color Gradient**: Green (fast) → Amber (slow) → Red (stopped) based on speed
- **Width**: 4px, semi-transparent
- **Markers**: Start (green flag), End (red flag), Stops (circle with duration)

### 7.2 Playback Controls

```
[⏮️] [▶️ Play] [⏭️]  ────●────────────  [1x ▼]
                     Time Slider

Speed: 1x | 5x | 10x | 30x
```

### 7.3 Data Sampling

| Time Range | Resolution | Points (approx) |
|------------|------------|-----------------|
| Last 1 hour | 10 seconds | 360 |
| Last 6 hours | 30 seconds | 720 |
| Last 24 hours | 60 seconds | 1,440 |
| Last 7 days | 5 minutes | 2,016 |
| Last 30 days | 15 minutes | 2,880 |

**Auto-resolution**: System selects resolution based on time range to keep point count <3,000 for rendering performance.

---

## 8. Security & Privacy

### 8.1 Data Access Control

| Rule | Implementation |
|------|---------------|
| Users see only their org's vehicles | RLS policy: `WHERE org_id = current_user_org_id()` |
| Dispatchers can view all org vehicles | Role check: `dispatcher` or higher |
| Drivers see only assigned vehicle | Role check: `driver` → filter by `current_driver_id` |
| GPS data encrypted at rest | PostgreSQL TDE + encrypted tablespaces |
| GPS data encrypted in transit | TLS 1.3 enforced |

### 8.2 Privacy Considerations

- **Data Retention**: Telemetry points retained for 90 days, then downsampled to hourly aggregates
- **Driver Consent**: Drivers informed of GPS tracking via employment agreement
- **Off-Duty Tracking**: Vehicles can be marked "off-duty" to pause location logging
- **Audit Trail**: All location data access logged to `audit_logs` table

---

## 9. Testing Strategy

### 9.1 Unit Tests

| Component | Coverage Target | Key Scenarios |
|-----------|----------------|---------------|
| Telemetry ingest handler | >90% | Valid/invalid GPS, missing fields, duplicate positions |
| Geofence detection logic | >90% | Entry, exit, boundary edge cases, polygon vs circle |
| History query builder | >85% | Time ranges, resolution downsampling, empty results |
| WebSocket handler | >80% | Connect, subscribe, disconnect, reconnect |

### 9.2 Integration Tests

| Test | Setup | Validation |
|------|-------|------------|
| End-to-end position update | Mock GPS device sends position | Position appears on map within 5s |
| Geofence alert delivery | Vehicle crosses geofence boundary | Alert logged, notification sent |
| History trail accuracy | Known GPS track | Trail matches expected path within 10m |
| Multi-tenant isolation | Two orgs, 500 vehicles each | Each org sees only their vehicles |

### 9.3 Performance Tests

| Test | Load | Target |
|------|------|--------|
| Map load (50 vehicles) | 1 org, 50 vehicles | <1s P95 |
| Map load (500 vehicles) | 1 org, 500 vehicles | <2s P95 |
| WebSocket fan-out | 100 concurrent users, 500 vehicles | <5s update latency |
| History query (24h) | 1,440 points | <3s P95 |
| Geofence detection | 500 vehicles × 10 geofences | <10s lag |

### 9.4 User Acceptance Tests

| Scenario | Expected Outcome |
|----------|-----------------|
| Dispatcher opens map | All vehicles visible within 2s |
| Vehicle moves on road | Marker updates within 5s |
| Click vehicle marker | Detail panel shows correct info |
| Search for vehicle by registration | Vehicle highlighted on map |
| View 24h history | Trail displayed with correct path |
| Vehicle enters geofence | Alert appears in notification center |

---

## 10. Rollout Plan

### 10.1 Phased Deployment

| Phase | Timeline | Scope | Success Criteria |
|-------|----------|-------|-----------------|
| **Alpha (Internal)** | Week 1-2 | Dev team, mock data | Map loads, markers update |
| **Beta (3 customers)** | Week 3-5 | 3 beta orgs, real GPS data | <2s load, <5s updates, positive feedback |
| **General Availability** | Week 6-7 | All customers | <1% error rate, >80% daily usage |

### 10.2 Feature Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `vehicle_map_enabled` | `true` | Kill switch for entire feature |
| `real_time_updates` | `true` | Toggle WebSocket live updates |
| `geofence_alerts` | `false` | Enable after geofence setup flow complete |
| `history_trail` | `true` | Toggle 24h history visualization |
| `max_vehicles_on_map` | `500` | Performance guardrail |

### 10.3 Monitoring & Alerts

| Metric | Threshold | Alert Channel |
|--------|-----------|---------------|
| Map load time P95 | >3s | Engineering Slack |
| WebSocket disconnect rate | >10% of connections | PagerDuty |
| Telemetry ingest error rate | >1% of requests | Engineering Slack |
| Geofence detection lag | >30s | Engineering Slack |
| Map component error rate | >0.5% of sessions | Sentry → Slack |

---

## 11. Open Questions & Decisions

| # | Question | Options | Recommendation | Owner | Due |
|---|----------|---------|---------------|-------|-----|
| 1 | Map provider? | Leaflet vs MapLibre vs Google Maps | **MapLibre** (open-source, no API costs, vector tiles) | Engineering | Apr 10 |
| 2 | Telemetry ingest protocol? | REST vs MQTT vs WebSocket | **REST** (simpler, aligns with current architecture) | Engineering | Apr 10 |
| 3 | Geofence UI creation? | In-app polygon editor vs admin-only API | **In-app editor** (dispatchers need self-service) | Product | Apr 15 |
| 4 | Offline vehicle handling? | Gray out markers vs hide after N minutes | **Gray out + tooltip** ("Last seen 15m ago") | Product | Apr 10 |
| 5 | History trail storage? | Raw telemetry vs materialized views | **Materialized hourly aggregates** (faster queries) | Engineering | Apr 15 |
| 6 | Max vehicles on map? | Hard limit vs virtualization | **Hard limit 500 + cluster markers** (performance) | Engineering | Apr 10 |

---

## 12. Dependencies & Risks

### 12.1 Dependencies

| Dependency | Status | Impact if Delayed |
|-----------|--------|-------------------|
| IAM/User Management (Q2) | In progress | Blocking — need auth for API access |
| Telemetry table schema | ✅ Exists | None — already in place |
| Vehicle API endpoints | ⚠️ Partial | Medium — need `GET /vehicles/locations` |
| Alert service integration | ⚠️ Partial | Medium — geofence alerts depend on this |
| Frontend map component | ❌ Not started | High — core UI element |

### 12.2 Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| GPS device integration delayed | High | High | Use mock data for Q3 alpha, integrate real devices in Q5 |
| Map performance degrades with 500+ vehicles | Medium | High | Implement marker clustering, canvas rendering |
| WebSocket connection instability | Medium | Medium | Fallback to polling (10s interval) if WS fails |
| TimescaleDB query performance | Low | Medium | Pre-test with production-scale data (500 vehicles × 24h) |
| Geofence detection accuracy (edge cases) | Medium | Low | Throttle detection (require 2 consecutive points inside/outside) |

---

## 13. Success Criteria & KPIs

### 13.1 MVP Success Criteria (Q3)

| Criteria | Target | Measurement |
|----------|--------|-------------|
| Map load time | <2s P95 | RUM (Real User Monitoring) |
| Position update latency | <5s P95 | WebSocket message timestamps |
| Beta customer adoption | 3/3 using daily | Usage analytics |
| Dispatcher satisfaction | >4/5 rating | Post-deployment survey |
| Zero data loss incidents | 0 | Telemetry gap analysis |

### 13.2 Long-term KPIs (Q4+)

| KPI | Target | Timeline |
|-----|--------|----------|
| Daily active dispatchers | >80% of logged-in users | Q4 2026 |
| Emergency response time improvement | 50% reduction (45min → 22min) | Q4 2026 |
| Geofence alert accuracy | >95% true positive rate | Q4 2026 |
| Map feature NPS | >50 | Q4 2026 |
| Support tickets related to tracking | <2% of total | Ongoing |

---

## 14. Glossary

| Term | Definition |
|------|-----------|
| **Telemetry** | GPS position + metadata (speed, heading, ignition, fuel) from vehicle |
| **Geofence** | Virtual geographic boundary (circle or polygon) |
| **Geofence Event** | Vehicle entering or exiting a geofence |
| **History Trail** | Visualization of vehicle movement over a time period |
| **Marker** | Visual representation of a vehicle on the map |
| **RLS** | Row-Level Security — PostgreSQL feature for multi-tenant isolation |
| **TimescaleDB** | PostgreSQL extension for time-series data optimization |
| **WebSocket** | Persistent bidirectional connection for real-time updates |
| **P95** | 95th percentile — 95% of requests faster than this |

---

## 15. Appendix

### 15.1 Related Documents

- [FlyFleet Enterprise PRD](../PRD.md) — Parent product requirements document
- [Driver Mobile App PRD](../product/DRIVER_MOBILE_APP_PRD.md) — Driver app specs (Phase 2)
- [OpenAPI Specification](../openapi.yaml) — Complete API documentation
- [Architecture Overview](../architecture/) — System architecture docs

### 15.2 API Endpoints Summary

| Endpoint | Method | Purpose | Status |
|----------|--------|---------|--------|
| `/api/v1/vehicles/locations` | GET | Get ALL org vehicle positions (bulk) | 🆕 NEW |
| `/api/v1/vehicles/{id}/location` | GET | Get single vehicle position | ✅ EXISTS |
| `/api/v1/vehicles/{id}/telemetry/history` | GET | Get 24h history with summary | 🔧 ENHANCE |
| `/api/v1/vehicles/{id}/replay` | GET | Get ordered telemetry for map replay | ✅ EXISTS |
| `/api/v1/vehicles/nearby` | GET | Find vehicles near a point | ✅ EXISTS |
| `/api/v1/geofences` | GET/POST | Manage geofences | ✅ EXISTS |
| `/ws/telemetry` | WebSocket | Real-time position push | 🆕 NEW |
| `/internal/v1/telemetry` | POST | Ingest GPS position (device → server) | ✅ EXISTS |

### 15.3 Implementation Tasks (Fleet-Core)

Based on the existing fleet-core architecture, here's what needs to be built:

#### New Files to Create (8 files)

| File | Purpose | Epic |
|------|---------|------|
| `tracking/handlers/vehicle_locations_handler.go` | Bulk vehicle locations endpoint | VT-1 |
| `tracking/handlers/history_trail_handler.go` | History with summary stats | VT-3 |
| `tracking/handlers/websocket_handler.go` | WebSocket real-time push | VT-2 |
| `tracking/handlers/websocket_handler_test.go` | WebSocket unit tests | VT-2 |
| `tracking/service/batch_location_service.go` | Batch Redis live-cache fetch | VT-1 |
| `tracking/repository/telemetry_batch.go` | Batch location queries | VT-1 |
| `geofencing/worker/geofence_checker.go` | Continuous geofence detection | VT-4 |
| `internal/pkg/distance/haversine.go` | Distance calculation utility | VT-3 |

#### Existing Files to Modify (6 files)

| File | Modification | Purpose |
|------|-------------|---------|
| `cmd/main.go` | Add 3 route registrations | Register new endpoints |
| `tracking/handlers/tracking_query_handler.go` | Add `HandleGetHistoryTrail` method | Enhanced history endpoint |
| `tracking/service/telemetry_query_service.go` | Add `GetHistoryWithSummary` method | Auto-resolution + summary |
| `tracking/service/ingestion_service.go` | Add Redis publish for position updates | WebSocket integration |
| `tracking/repository/telemetry_repository.go` | Add `GetHistoryBucket` method | TimescaleDB time_bucket query |
| `tracking/models/telemetry.go` | Add `Status` constants (moving/idle/stopped) | Status calculation |

#### Database Migrations (1 file)

| Migration | Purpose |
|-----------|---------|
| `301_tracking_performance_indexes.sql` | Add indexes for bulk location query, geofence spatial queries, history trail queries |

### 15.4 Change Log

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | April 7, 2026 | Product Team | Initial draft |
| 2.0 | April 7, 2026 | Product Team | **Fleet-Core Alignment**: Mapped all requirements to existing `tracking/`, `geofencing/`, `trip/` domains. Identified 8 new files, 6 existing files to modify, 1 migration. All API contracts follow fleet-core patterns (JWKS auth, tenant middleware, role checks). |

---

**Next Review:** April 21, 2026
**Approvers:** Product Lead, Engineering Lead, Design Lead
**Stakeholders:** Dispatchers (beta customers), Fleet Managers, Engineering Team
