# Vehicle Tracking (Q3 — P0) — Product Specification

**Document ID:** PSPEC-VT-002
**Version:** 1.0
**Status:** Draft
**Created:** April 8, 2026
**Author:** Product Team (AI-Assisted)
**Review Cycle:** Bi-weekly
**Parent Epic:** FlyFleet Enterprise Fleet Management Platform
**Priority:** P0 (Q3 Delivery)
**User Story Owner:** Dispatcher Persona

---

## 1. Executive Summary

### 1.1 User Story

```
As a dispatcher
I want to see all my vehicles on a map with live positions
So that I know where every vehicle is right now
```

### 1.2 Problem Statement

Fleet dispatchers currently lack real-time visibility into vehicle locations. They rely on:
- Phone calls to drivers (2-3 hours/day wasted)
- WhatsApp check-ins (unreliable, no audit trail)
- Manual paper logs (error-prone, delayed)

This creates:
- **30-45 minute response time** to emergencies (accidents, breakdowns)
- **Customer dissatisfaction** from inability to provide accurate ETAs
- **Revenue loss** from inefficient route utilization
- **Dispatcher frustration** from juggling multiple communication channels

### 1.3 Solution

A **real-time vehicle tracking map** integrated into the FlyFleet dashboard that shows all organization vehicles with live GPS positions, status indicators, and historical movement trails. Dispatchers get instant fleet visibility without calling drivers.

### 1.4 Success Metrics

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| Map load time | <2s P95 | Time from page load to all markers visible |
| Position update latency | <5s | Time from GPS ingest to map marker update |
| Daily active usage | >80% of dispatchers | Users who view map ≥1x per session |
| Dispatcher coordination time reduction | 50% reduction | Pre/post deployment survey |
| Emergency response time | <15 minutes | Time from incident to dispatcher awareness |

---

## 2. Acceptance Criteria

### 2.1 Core Acceptance Criteria (from User Story)

- [x] Map loads with all organization vehicles in <2s
- [x] Vehicle markers update every 5 seconds
- [x] Clicking a vehicle shows: driver name, speed, last update, status
- [x] Geofence alerts when vehicle enters/exits zone
- [x] Vehicle history trail for last 24 hours

### 2.2 Extended Acceptance Criteria

- [x] Map auto-fits to show all vehicles on initial load
- [x] Vehicle markers are color-coded by status (moving/idle/stopped/offline)
- [x] Search/filter vehicles by registration number, driver name, or status
- [x] Clicking vehicle opens detail panel without leaving map view
- [x] History trail toggle shows/hides 24h movement path
- [x] Map supports zoom from country view to street-level detail
- [x] Works on desktop (1920x1080) and tablet (1024x768) viewports
- [x] WebSocket connection for real-time updates with automatic fallback to polling

---

## 3. Current State Analysis

### 3.1 Existing Infrastructure (services/fleet-core)

The fleet-core monolith already has **extensive tracking infrastructure** built:

#### ✅ Existing Tracking Endpoints

| Endpoint | Method | Handler | Status | Notes |
|----------|--------|---------|--------|-------|
| `/internal/v1/telemetry` | POST | `tracking/handlers/ingest_handler.go` | ✅ Live | AIS-140 aligned ingest with HMAC validation, per-device rate limiting (5Hz) |
| `/v1/vehicles/{id}/location` | GET | `tracking/handlers/tracking_query_handler.go` | ✅ Live | Single vehicle live location (Redis cache → DB fallback) |
| `/v1/vehicles/{id}/history` | GET | `tracking/handlers/tracking_query_handler.go` | ✅ Live | Paginated telemetry history with tenant isolation |
| `/v1/vehicles/{id}/replay` | GET | `tracking/handlers/tracking_query_handler.go` | ✅ Live | Ordered telemetry for map replay (max 24h, 1000 points) |
| `/v1/vehicles/nearby` | GET | `tracking/handlers/tracking_query_handler.go` | ✅ Live | Nearby vehicles by lat/lng/radius (cap 50km) |
| `/v1/devices/{id}` | Various | `tracking/handlers/device_handler.go` | ✅ Live | Device management |
| `/v1/devices/{id}/commands` | POST | `tracking/handlers/device_command_handler.go` | ✅ Live | Device commands (ping, reboot, immobilize) |

#### ✅ Existing Tracking Services

| Service | File | Capability |
|---------|------|------------|
| **IngestionService** | `tracking/service/ingestion_service.go` | Validates, deduplicates, writes to DB + Redis live cache, publishes to NATS |
| **TelemetryQueryService** | `tracking/service/telemetry_query_service.go` | GetLiveLocation (Redis→DB fallback, 5-min offline threshold), GetHistory (paginated), GetReplay (ASC order, 24h cap), GetNearby (spatial query, 50km cap) |
| **AnomalyEngine** | `tracking/service/anomaly_engine.go` | Detects overspeed, route deviation, tamper, teleport, fuel drop |
| **AlertService** | `tracking/service/alert_service.go` | Emits alert events, publishes to MQTT for real-time notifications |

#### ✅ Existing Geofencing Domain

| Component | File | Capability |
|-----------|------|------------|
| **GeofenceEngine** | `geofencing/engine/geofence_engine.go` | PostGIS-based spatial detection, subscribes to `fleet.telemetry.gps` NATS topic |
| **GeofenceService** | `geofencing/service/geofence_service.go` | CRUD operations, event publishing to NATS |
| **GeofenceHandler** | `geofencing/handlers/geofence_handler.go` | HTTP handlers for geofence management |
| **Geofence Models** | `geofencing/models/geofence.go` | Circle, polygon, route corridor geofences with entry/exit/dwell events |

#### ✅ Existing WebSocket Infrastructure

| Component | File | Capability |
|-----------|------|------------|
| **WebSocketHandler** | `tracking/handlers/websocket.go` | Real-time position push to dashboard clients via Redis pub/sub |
| **WSClient** | `tracking/handlers/websocket.go` | Per-connection client with read/write pumps, vehicle subscription support |
| **Redis Pub/Sub** | `tracking/service/ingestion_service.go` | Position update publishing to org-specific channels |

#### ✅ Existing Data Models

| Model | File | Fields |
|-------|------|--------|
| **TelemetryEvent** | `tracking/models/telemetry.go` | event_id, device_id, vehicle_id, lat, lng, speed, heading, ignition, fuel_level, odometer, event_type, sequence_no, recorded_at, received_at |
| **LiveLocation** | `tracking/models/telemetry.go` | vehicle_id, lat, lng, speed, heading, ignition, last_seen, status (online/offline with 5-min threshold) |
| **GeofenceEvent** | `tracking/models/telemetry.go` | id, vehicle_id, org_id, geofence_id, event_type (ENTRY/EXIT/DWELL), lat, lng, timestamp |
| **AnomalyEvent** | `tracking/models/telemetry.go` | type, vehicle_id, device_id, severity, message, payload, timestamp |

#### ✅ Existing Database Tables

| Table | Type | Purpose |
|-------|------|---------|
| `telemetry_points` | TimescaleDB hypertable | GPS telemetry (partitioned by month) |
| `gps_locations` | Regular table | Latest GPS positions per vehicle |
| `telemetry_events` | Partitioned table | Event log |
| `telemetry_analytics` | Regular table | Analytics cache |
| `telemetry_anomalies` | Regular table | Anomaly detection results |
| `telemetry_outbox` | Regular table | Outbox pattern for reliable publishing |
| `geofences` | Regular table | Geofence definitions |
| `geofence_assignments` | Regular table | Vehicle-geofence mappings |
| `geofence_events` | Regular table | Entry/exit/dwell event log |

### 3.2 Gap Analysis

| Requirement | Existing Foundation | Gap to Fill |
|-------------|-------------------|-------------|
| Map loads ALL org vehicles in <2s | `GetLiveLocation` is single-vehicle only | **NEW**: Bulk vehicle locations endpoint (`GET /v1/vehicles/locations`) |
| Markers update every 5 seconds | Redis live cache + WebSocket exists | **ENHANCE**: Ensure WebSocket pushes every 5s, add connection health monitoring |
| Click vehicle shows driver name, speed, status | `LiveLocation` has speed+status, but no driver join | **NEW**: Enriched response joining vehicles+drivers tables |
| Geofence alerts on enter/exit | `GeofenceEngine` exists, processes GPS events | **INTEGRATE**: Connect geofence events to alert service + WebSocket notifications |
| 24h history trail | `GetReplay` exists (24h, 1000 point cap) | **ENHANCE**: Auto-resolution downsampling, distance/speed summary stats |

---

## 4. Technical Architecture

### 4.1 System Context Diagram

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
│  │ (GET locations)│ │ (GET vehicles)│ │ (Push updates)  │ │
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
          │ GPS Ingest (device → NATS → DB)
┌─────────┴──────────┐
│   GPS Devices /    │
│   Mobile App       │
└────────────────────┘
```

### 4.2 Data Flow

```
1. Vehicle GPS device sends position → POST /api/v1/telemetry
2. Fleet-core validates HMAC, writes to telemetry_points (TimescaleDB)
3. IngestionService updates Redis live cache (vehicle:{id}:live)
4. IngestionService publishes to NATS topic "fleet.telemetry.gps"
5. GeofenceEngine subscribes to NATS, checks against active geofences
6. If geofence state changed → publish geofence event to NATS
7. AlertService consumes geofence event → publish to MQTT + WebSocket
8. WebSocket handler polls Redis pub/sub for org channel
9. WebSocket pushes position update → dashboard-web
10. Leaflet marker animates to new position on map
```

---

## 5. API Specifications

### 5.1 GET /api/v1/vehicles/locations (NEW)

**Purpose**: Fetch current positions of ALL vehicles in organization (bulk live location)

**Why needed?** The existing `/v1/vehicles/{id}/location` is single-vehicle only. The dispatcher map needs ALL vehicles in one request for <2s load time.

**Request**:
```http
GET /api/v1/vehicles/locations
Authorization: Bearer <JWT via JWKS>
X-Org-ID: <org-uuid> (extracted from JWT by TenantContextMiddleware)
```

**Query Parameters** (optional):
| Param | Type | Description | Default |
|-------|------|-------------|---------|
| status | string | Filter by status: moving, idle, stopped, offline | all |
| driver_id | UUID | Filter by assigned driver | all |
| search | string | Search by registration or driver name | — |

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
      "odometer_km": 45230.5,
      "status": "moving",
      "driver": {
        "id": "d-456-def",
        "name": "Rajesh Kumar",
        "phone": "+91 98765 43210"
      },
      "last_update": "2026-04-08T10:30:00Z",
      "last_seen_seconds_ago": 5
    }
  ],
  "meta": {
    "total": 42,
    "online": 38,
    "offline": 4,
    "moving": 15,
    "idle": 12,
    "stopped": 11,
    "timestamp": "2026-04-08T10:30:05Z"
  }
}
```

**Implementation** (new `tracking/handlers/vehicle_locations_handler.go`):
```go
package handlers

type VehicleLocationsHandler struct {
    queryService *service.TelemetryQueryService
    liveCache    repository.LiveCacheRepository
    vehicleRepo  vehicleRepo.VehicleRepository
    driverRepo   driverRepo.DriverRepository
    logger       *zap.Logger
}

func (h *VehicleLocationsHandler) HandleGetVehicleLocations(w http.ResponseWriter, r *http.Request) {
    orgID := middleware.GetOrgID(r.Context())

    // 1. Get all org vehicles from vehicle repo
    vehicles, err := h.vehicleRepo.GetOrgVehicles(r.Context(), orgID)

    // 2. Batch-fetch latest locations from Redis (pipeline for performance)
    locations := h.liveCache.GetBatchLiveLocations(r.Context(), vehicleIDs)

    // 3. Enrich with driver info (batch query)
    drivers := h.driverRepo.GetByIDs(r.Context(), driverIDs)

    // 4. Calculate status (moving/idle/stopped/offline) from speed + ignition + age
    // 5. Return enriched response
}
```

**Repository Query** (new method in `vehicle/repository/vehicle_repository.go`):
```sql
SELECT
    v.id,
    v.registration_number,
    v.make,
    v.model,
    v.current_driver_id,
    v.status as vehicle_status
FROM vehicles v
WHERE v.tenant_id = $1
  AND v.deleted_at IS NULL
ORDER BY v.registration_number;
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

**Performance Target**: <2s P95 for 500 vehicles
- Redis batch get: ~50ms (500 keys via pipeline)
- Vehicle query: ~100ms (indexed on tenant_id)
- Driver enrichment: ~200ms (batch by IDs)
- JSON encode: ~50ms

---

### 5.2 WebSocket /ws/telemetry (ENHANCE EXISTING)

**Purpose**: Real-time vehicle position updates to dashboard (push model instead of polling)

**Current state**: WebSocket handler exists at `tracking/handlers/websocket.go` with Redis pub/sub integration. Need to enhance to support:
- Guaranteed 5-second update interval
- Connection health monitoring (ping/pong)
- Auto-reconnection with exponential backoff
- Vehicle subscription filtering (optional)

**Connection**:
```
wss://api.flyfleet.io/ws/telemetry?token=<JWT>&org_id=<org-uuid>
```

**Server → Client Messages**:

#### Position Update (every 5s)
```json
{
  "type": "position_update",
  "data": {
    "vehicle_id": "v-123-abc",
    "registration": "DL01AB1234",
    "latitude": 28.6139,
    "longitude": 77.2090,
    "speed_kmh": 45.2,
    "heading": 135,
    "ignition": true,
    "status": "moving",
    "timestamp": "2026-04-08T10:30:00Z",
    "last_seen_seconds_ago": 5
  }
}
```

#### Geofence Alert
```json
{
  "type": "geofence_alert",
  "data": {
    "vehicle_id": "v-123-abc",
    "registration": "DL01AB1234",
    "geofence_id": "g-789-xyz",
    "geofence_name": "Delhi Warehouse",
    "event_type": "ENTER",
    "latitude": 28.6139,
    "longitude": 77.2090,
    "timestamp": "2026-04-08T10:35:00Z"
  }
}
```

#### Connection Ack
```json
{
  "type": "connection_ack",
  "data": {
    "status": "connected",
    "client_id": "ws-client-uuid",
    "heartbeat_interval_ms": 30000
  }
}
```

**Client → Server Messages**:

#### Subscribe to Specific Vehicles (optional)
```json
{
  "type": "subscribe_vehicles",
  "data": {
    "vehicle_ids": ["v-123-abc", "v-456-def"]
  }
}
```

#### Unsubscribe
```json
{
  "type": "unsubscribe_vehicles",
  "data": {
    "vehicle_ids": ["v-123-abc"]
  }
}
```

#### Ping (health check)
```json
{
  "type": "ping"
}
```

**Server Response to Ping**:
```json
{
  "type": "pong",
  "data": {
    "timestamp": "2026-04-08T10:30:00Z"
  }
}
```

**Implementation Enhancements** (modify existing `tracking/handlers/websocket.go`):
```go
// Add heartbeat mechanism
func (h *WebSocketHandler) StartHeartbeat(client *WSClient) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        case <-client.ctx.Done():
            return
        }
    }
}

// Add guaranteed 5s position updates
func (h *WebSocketHandler) StartPositionUpdates(orgID string, client *WSClient) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Batch fetch latest positions for org
            locations := h.liveCache.GetBatchLiveLocations(ctx, orgVehicleIDs)
            
            // Send only changed positions to reduce bandwidth
            for _, loc := range locations {
                if loc.HasChanged(client.lastSent[loc.VehicleID]) {
                    msg := PositionUpdateMessage{
                        Type: "position_update",
                        Data: loc,
                    }
                    client.send <- msg
                    client.lastSent[loc.VehicleID] = loc
                }
            }
        case <-client.ctx.Done():
            return
        }
    }
}
```

**Redis Pub/Sub Integration** (modify existing `tracking/service/ingestion_service.go`):
```go
// In IngestionService.Ingest(), after writing to DB + Redis cache:
func (s *IngestionService) PublishPositionUpdate(ctx context.Context, event *models.TelemetryEvent) {
    message, _ := json.Marshal(map[string]interface{}{
        "type": "position_update",
        "vehicle_id": event.VehicleID.String(),
        "latitude": event.Lat,
        "longitude": event.Lng,
        "speed_kmh": event.Speed,
        "heading": event.Heading,
        "timestamp": event.RecordedAt.Format(time.RFC3339),
    })

    // Publish to org-specific channel
    channel := fmt.Sprintf("org:%s:telemetry", event.TenantID.String())
    s.redis.Publish(ctx, channel, message).Err()
}
```

**Performance Target**: <5s P95 from telemetry ingest → WebSocket delivery to dashboard
- Telemetry ingest → Redis publish: ~10ms
- Redis → WebSocket delivery: ~50ms
- Total latency: <100ms (well within 5s target)

**Fallback**: If WebSocket connection fails, dashboard falls back to polling `GET /api/v1/vehicles/locations` every 10 seconds.

---

### 5.3 GET /api/v1/vehicles/{id}/telemetry/history (ENHANCE EXISTING)

> **Note:** The existing `/v1/vehicles/{id}/replay` endpoint already provides ordered telemetry for map replay with a 24h limit and 1000-point cap. This spec **enhances** it with auto-resolution downsampling and summary statistics.

**Purpose**: Fetch 24-hour movement history for a vehicle with intelligent downsampling

**Request**:
```http
GET /api/v1/vehicles/{id}/telemetry/history?start=2026-04-07T10:00:00Z&end=2026-04-08T10:00:00Z&resolution=auto
Authorization: Bearer <JWT via JWKS>
```

**Query Parameters**:
| Param | Type | Required | Description | Default |
|-------|------|----------|-------------|---------|
| start | ISO 8601 | Yes | Start time | — |
| end | ISO 8601 | Yes | End time | — |
| resolution | string | No | `auto`, `10s`, `30s`, `60s`, `5m`, `15m` | `auto` |
| include_summary | bool | No | Include distance/speed stats | `false` |

**Auto-Resolution Logic** (when `resolution=auto`):
| Time Range | Effective Resolution | Max Points |
|------------|---------------------|------------|
| < 1 hour | 10 seconds | 360 |
| 1-6 hours | 30 seconds | 720 |
| 6-24 hours | 60 seconds | 1,440 |
| 1-7 days | 5 minutes | 2,016 |
| 7-30 days | 15 minutes | 2,880 |

**Response** (200 OK):
```json
{
  "data": {
    "vehicle_id": "v-123-abc",
    "points": [
      {
        "timestamp": "2026-04-07T10:00:00Z",
        "latitude": 28.6139,
        "longitude": 77.2090,
        "speed_kmh": 0,
        "heading": 135,
        "ignition": false
      },
      {
        "timestamp": "2026-04-07T10:01:00Z",
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
-- Use time_bucket for downsampling
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

**Performance Target**: <3s P95 for 24h data (1,440 points at 1-min resolution)

---

### 5.4 Geofence Alert Integration (ENHANCE EXISTING)

> **Note:** The `geofencing/` domain already exists with `GeofenceEngine` that subscribes to NATS topic `fleet.telemetry.gps` and detects entry/exit/dwell events. The `GeofenceEvent` model (ENTRY/EXIT/DWELL) is already defined. This spec **integrates** geofence detection with live telemetry and adds dispatcher alert delivery.

**Existing Geofencing Endpoints** (already live):
| Endpoint | Method | Status | Notes |
|----------|--------|--------|-------|
| `/api/v1/geofences` | GET | ✅ Exists | List org geofences |
| `/api/v1/geofences` | POST | ✅ Exists | Create geofence |
| `/api/v1/geofences/{id}` | GET/PUT/DELETE | ✅ Exists | CRUD operations |
| `/api/v1/geofences/events` | GET | ✅ Exists | Query geofence events |

**Geofence Detection Flow** (already implemented in `geofencing/engine/geofence_engine.go`):

1. GPS device sends position → `POST /api/v1/telemetry`
2. `IngestionService` validates, writes to DB + Redis, publishes to NATS `fleet.telemetry.gps`
3. `GeofenceEngine` (subscribed to NATS) receives GPS event
4. Engine queries all active geofences for vehicle from DB
5. Uses PostGIS `ST_DWithin` (circle) or `ST_Within` (polygon) to check containment
6. Compares with previous state (stored in Redis: `vehicle_state:{vehicle_id}`)
7. If state changed (ENTER/EXIT) → insert `geofence_events` record
8. Publish geofence event to NATS `fleet.geofence.events`
9. `AlertService` consumes event → save to DB + publish to MQTT + WebSocket

**Enhancement**: Connect geofence events to WebSocket for real-time dashboard notifications

```go
// In geofencing/engine/geofence_engine.go, after triggering geofence event:
func (ge *GeofenceEngine) triggerGeofenceEvent(ctx context.Context, vehicleID, geofenceID uuid.UUID, geofenceName, eventType string, event GPSEvent) {
    // ... existing DB insert code ...

    // Publish to NATS for alert service
    eventData, _ := json.Marshal(map[string]interface{}{
        "vehicle_id": vehicleID.String(),
        "geofence_id": geofenceID.String(),
        "geofence_name": geofenceName,
        "event_type": eventType,
        "latitude": event.Latitude,
        "longitude": event.Longitude,
        "timestamp": event.Timestamp,
    })

    ge.nc.Publish("fleet.geofence.events", eventData)

    // NEW: Also publish to Redis for WebSocket fan-out
    orgID := ge.getOrgIDForVehicle(vehicleID)
    channel := fmt.Sprintf("org:%s:geofence_alerts", orgID)
    ge.redis.Publish(ctx, channel, eventData).Err()
}
```

**WebSocket Handler Enhancement** (modify `tracking/handlers/websocket.go`):
```go
// Subscribe to both telemetry and geofence alert channels
func (h *WebSocketHandler) subscribeToOrgUpdates(orgID string, client *WSClient) {
    ctx := context.Background()
    
    // Subscribe to telemetry updates
    telemetryPubsub := h.redis.Subscribe(ctx, database.RedisChannelForOrg(orgID))
    defer telemetryPubsub.Close()
    telemetryCh := telemetryPubsub.Channel()
    
    // Subscribe to geofence alerts
    geofencePubsub := h.redis.Subscribe(ctx, fmt.Sprintf("org:%s:geofence_alerts", orgID))
    defer geofencePubsub.Close()
    geofenceCh := geofencePubsub.Channel()
    
    for {
        select {
        case msg := <-telemetryCh:
            h.forwardToClient(client, []byte(msg.Payload))
        case msg := <-geofenceCh:
            // Wrap in geofence_alert message type
            alert := map[string]interface{}{
                "type": "geofence_alert",
                "data": msg.Payload,
            }
            encoded, _ := json.Marshal(alert)
            h.forwardToClient(client, encoded)
        case <-ctx.Done():
            return
        }
    }
}
```

**Geofence Alert WebSocket Message**:
```json
{
  "type": "geofence_alert",
  "data": {
    "vehicle_id": "v-123-abc",
    "registration": "DL01AB1234",
    "geofence_id": "g-789-xyz",
    "geofence_name": "Delhi Warehouse",
    "event_type": "ENTER",
    "latitude": 28.6139,
    "longitude": 77.2090,
    "timestamp": "2026-04-08T10:35:00Z",
    "severity": "HIGH"
  }
}
```

**Alert Severity Levels**:
| Event Type | Severity | Dispatcher Action |
|------------|----------|-------------------|
| ENTER (customer site) | LOW | Log for customer ETA update |
| EXIT (customer site) | MEDIUM | Confirm delivery completion |
| ENTER (restricted zone) | HIGH | Immediate dispatcher notification |
| EXIT (warehouse after hours) | CRITICAL | Investigate potential theft |

---

## 6. Frontend Specifications

### 6.1 Map Component (Leaflet/MapLibre)

**Library Choice**: Leaflet.js (lightweight, well-documented, good mobile support)

**Key Features**:
- Interactive map with zoom (country view → street level)
- Vehicle markers with custom icons (color-coded by status)
- Marker clustering for dense areas (>50 vehicles)
- Auto-fit bounds to show all vehicles on initial load
- Smooth marker animations on position updates
- History trail polyline overlay (toggle on/off)
- Geofence polygon/circle overlays (semi-transparent)
- Full-screen mode support

**Marker Color Coding**:
| Status | Color | Icon |
|--------|-------|------|
| Moving | 🟢 Green | Truck icon with motion lines |
| Idle | 🟡 Yellow | Truck icon (stationary) |
| Stopped | 🟠 Orange | Truck icon with X |
| Offline | 🔴 Red | Truck icon with offline badge |

**Map Layers** (toggleable):
1. Base map (OpenStreetMap / Mapbox Streets)
2. Vehicle markers (default: on)
3. History trails (default: off, toggled per vehicle)
4. Geofence zones (default: off, toggled globally)
5. Traffic layer (future phase)

### 6.2 Vehicle Detail Panel (Slide-over)

**Trigger**: Click on vehicle marker

**Content**:
```
┌─────────────────────────────────────┐
│ [X] DL01AB1234                      │
├─────────────────────────────────────┤
│ Driver: Rajesh Kumar                │
│ Phone: +91 98765 43210 [📞 Call]   │
│                                     │
│ Status: 🟢 Moving                   │
│ Speed: 45.2 km/h                    │
│ Heading: Southeast (135°)           │
│ Last Update: 5 seconds ago          │
│                                     │
│ Fuel: 72.5%                         │
│ Odometer: 45,230.5 km               │
│ Ignition: ON                        │
│                                     │
│ [📍 History Trail] [🔔 Alerts]     │
│ [📋 Trip Details] [📞 Contact]     │
└─────────────────────────────────────┘
```

**Interactions**:
- Click marker → panel slides in from right
- Click X → panel slides out
- Click [History Trail] → show 24h trail on map, highlight in panel
- Click [Alerts] → show recent alerts for vehicle
- Click [Trip Details] → navigate to trip view (future)
- Click [Contact] → trigger driver call/SMS (future)

### 6.3 Search & Filter Bar

**Position**: Top of map view

**Features**:
```
┌─────────────────────────────────────────────────────────┐
│ 🔍 Search vehicles...    [Status ▼] [Driver ▼] [Filter]│
└─────────────────────────────────────────────────────────┘
```

**Search**:
- By registration number (exact match)
- By driver name (fuzzy match)
- By vehicle make/model (fuzzy match)
- Real-time results as user types (debounced 300ms)

**Filters**:
- Status: All, Moving, Idle, Stopped, Offline
- Driver: Dropdown of active drivers
- Custom: Multi-select vehicle tags (future)

**Results**: Show count badge (e.g., "42 vehicles") next to search bar

### 6.4 WebSocket Integration

**Connection Lifecycle**:
```javascript
class TrackingWebSocket {
  constructor(token, orgId) {
    this.url = `wss://api.flyfleet.io/ws/telemetry?token=${token}&org_id=${orgId}`;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 10;
    this.reconnectDelay = 1000; // 1s, with exponential backoff
    this.connect();
  }

  connect() {
    this.ws = new WebSocket(this.url);
    
    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
    };
    
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      
      switch (message.type) {
        case 'position_update':
          this.updateVehicleOnMap(message.data);
          break;
        case 'geofence_alert':
          this.showGeofenceAlert(message.data);
          break;
        case 'connection_ack':
          console.log('Connection acknowledged');
          break;
        case 'pong':
          this.lastPong = Date.now();
          break;
      }
    };
    
    this.ws.onclose = () => {
      console.log('WebSocket closed, reconnecting...');
      this.reconnect();
    };
    
    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    // Start heartbeat
    this.startHeartbeat();
  }

  reconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
      setTimeout(() => this.connect(), delay);
    } else {
      console.error('Max reconnect attempts reached, falling back to polling');
      this.startPolling();
    }
  }

  startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      if (this.lastPong && Date.now() - this.lastPong > 60000) {
        console.warn('No pong received in 60s, connection may be dead');
        this.ws.close();
      } else {
        this.ws.send(JSON.stringify({ type: 'ping' }));
      }
    }, 30000); // Every 30s
  }

  startPolling() {
    // Fallback to polling every 10s
    this.pollingInterval = setInterval(async () => {
      const response = await fetch('/api/v1/vehicles/locations');
      const data = await response.json();
      this.updateAllVehiclesOnMap(data.data);
    }, 10000);
  }
}
```

**Performance Optimizations**:
- Only update markers that have changed position (diff with previous state)
- Use `requestAnimationFrame` for smooth marker animations
- Debounce rapid position updates (max 1 update/second per vehicle)
- Use Web Workers for history trail calculations (Haversine distance)

---

## 7. Non-Functional Requirements

### 7.1 Performance

| Requirement | Target | Measurement |
|-------------|--------|-------------|
| Map load time (initial) | <2s P95 | Time from page load to all markers visible |
| Position update latency | <5s P95 | Time from GPS ingest to map marker update |
| API response time (bulk locations) | <500ms P95 | `GET /api/v1/vehicles/locations` |
| WebSocket message delivery | <100ms P95 | Redis publish → WebSocket delivery |
| History query time (24h) | <3s P95 | `GET /api/v1/vehicles/{id}/telemetry/history` |
| Max vehicles on map | 500 | Before performance degradation |
| Marker update frequency | Every 5s | Configurable per org |

### 7.2 Scalability

| Metric | Current | Target | Scaling Strategy |
|--------|---------|--------|-----------------|
| Vehicles per org | 50 | 500 | Redis pipeline batch queries |
| Concurrent WebSocket connections | 100 | 1,000 | Horizontal scale with NATS fan-out |
| Telemetry ingest rate | 100 events/sec | 10,000 events/sec | NATS buffering + TimescaleDB partitions |
| Geofence checks | 10 geofences/vehicle | 100 geofences/vehicle | PostGIS spatial indexing |

### 7.3 Reliability

| Requirement | Target | Implementation |
|-------------|--------|----------------|
| Uptime | 99.9% | Health checks, auto-restart on crash |
| Data loss | 0% | Outbox pattern for telemetry, TimescaleDB WAL |
| WebSocket reconnection | <30s | Exponential backoff (1s → 2s → 4s → 8s → 16s → 32s) |
| Fallback mechanism | Polling every 10s | If WebSocket fails after 10 attempts |
| Graceful degradation | Show last-known position | If Redis cache expired, query DB |

### 7.4 Security

| Requirement | Implementation |
|-------------|----------------|
| Authentication | JWT via JWKS middleware (same as all fleet-core APIs) |
| Authorization | Tenant isolation via `TenantContextMiddleware` (org-scoped queries) |
| WebSocket auth | JWT in query param (validated on handshake) |
| Data encryption | TLS 1.3 for in-transit, AES-256 for DB at rest |
| Rate limiting | 5Hz per device (ingest), 100 req/min per user (queries) |
| Audit trail | All geofence alerts logged to `geofence_events` table |

### 7.5 Accessibility

| Requirement | Implementation |
|-------------|----------------|
| WCAG 2.1 AA | Map alternative: list view with vehicle statuses |
| Keyboard navigation | Tab through vehicles, Enter to open detail panel |
| Screen reader | ARIA labels on markers, live region for alerts |
| Color contrast | Markers use color + shape (accessible to colorblind users) |
| Reduced motion | Respect `prefers-reduced-motion` (disable marker animations) |

---

## 8. Testing Strategy

### 8.1 Unit Tests

| Component | Test Cases | Framework |
|-----------|-----------|-----------|
| `VehicleLocationsHandler` | Valid org, empty org, filtered by status, driver enrichment | Go `testing` |
| `TelemetryQueryService.GetHistoryWithSummary` | Resolution calculation, summary stats, empty range | Go `testing` |
| `CalculateVehicleStatus` | Moving, idle, stopped, offline edge cases | Go `testing` |
| WebSocket message encoding | Position update, geofence alert, connection ack | Go `testing` |

### 8.2 Integration Tests

| Test Scenario | Setup | Assertion |
|---------------|-------|-----------|
| Bulk vehicle locations | 500 vehicles in DB, Redis cache populated | Response contains all 500, <2s P95 |
| WebSocket real-time update | Send telemetry event via API | WebSocket receives position update <5s |
| Geofence alert flow | Create geofence, send telemetry inside | Geofence event in DB, WebSocket receives alert |
| History trail with downsampling | 10,000 telemetry points over 24h | Query returns ~1,440 points at 60s resolution |

### 8.3 Load Tests

| Test | Target | Tool |
|------|--------|------|
| Bulk locations endpoint | 1000 concurrent users, <500ms P95 | k6 or Locust |
| WebSocket connections | 1000 concurrent connections, <100ms message delivery | k6 WebSocket extension |
| Telemetry ingest | 10,000 events/sec sustained | NATS benchmark + custom script |
| History query (24h) | 100 concurrent queries, <3s P95 | k6 |

### 8.4 End-to-End Tests

| User Journey | Steps | Expected Result |
|--------------|-------|-----------------|
| Dispatcher opens map | 1. Login as dispatcher<br>2. Navigate to map view | Map loads with all org vehicles in <2s |
| Vehicle marker updates | 1. Open map<br>2. Send telemetry via API<br>3. Observe marker | Marker moves to new position within 5s |
| View vehicle details | 1. Click vehicle marker | Detail panel opens with driver name, speed, status, last update |
| Geofence alert received | 1. Create geofence<br>2. Send telemetry inside<br>3. Observe WebSocket | Alert received in <5s, shows geofence name + event type |
| View history trail | 1. Click vehicle<br>2. Toggle history trail | 24h trail displayed on map, summary stats shown |

---

## 9. Deployment Plan

### 9.1 Feature Flags

| Flag | Default | Description |
|------|---------|-------------|
| `vehicle_map_enabled` | `true` | Master flag for all tracking features |
| `real_time_updates` | `true` | Enable WebSocket real-time updates |
| `history_trail` | `true` | Enable 24h history trail queries |
| `geofence_alerts` | `true` | Enable geofence detection engine |
| `max_vehicles_on_map` | `500` | Configurable limit per org |

### 9.2 Phased Rollout

| Phase | Duration | Scope | Success Criteria |
|-------|----------|-------|-----------------|
| **Alpha** | 1 week | Internal team (5 orgs, 50 vehicles) | No critical bugs, WebSocket stable |
| **Beta** | 2 weeks | Beta customers (20 orgs, 200 vehicles) | <2s map load, <5s update latency |
| **GA** | Ongoing | All customers | 99.9% uptime, <1% error rate |

### 9.3 Monitoring & Alerts

| Metric | Alert Threshold | Action |
|--------|----------------|--------|
| Map load time P95 | >3s | Investigate DB query performance, Redis cache hit rate |
| WebSocket connection failures | >5% | Check Redis pub/sub, NATS connectivity |
| Telemetry ingest errors | >1% | Review HMAC validation, dedup failures |
| Geofence detection latency | >10s | Optimize PostGIS queries, increase engine workers |
| DB query time (history) | >5s P95 | Add indexes, tune TimescaleDB partitions |

---

## 10. Dependencies & Risks

### 10.1 Dependencies

| Dependency | Status | Impact if Delayed |
|------------|--------|-------------------|
| IAM authentication (JWKS) | ✅ Complete | Cannot authenticate WebSocket connections |
| Vehicle management domain | ✅ Complete | Cannot fetch vehicle metadata (make, model, registration) |
| Driver management domain | ✅ Complete | Cannot enrich with driver name/phone |
| Telemetry ingest pipeline | ✅ Complete | No GPS data to display |
| Redis infrastructure | ✅ Complete | No live cache, no WebSocket pub/sub |
| NATS JetStream | ✅ Complete | No async telemetry buffering |
| TimescaleDB | ✅ Complete | No efficient time-series queries |

### 10.2 Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| WebSocket scalability | Medium | High | Fallback to polling, horizontal scale with NATS fan-out |
| Redis cache eviction | Low | Medium | DB fallback for live location, increased TTL |
| Geofence detection performance | Medium | Medium | PostGIS spatial indexing, batch geofence queries |
| Marker clustering at scale | High | Low | Use Leaflet.markercluster, limit to 500 vehicles |
| GPS device downtime | Medium | Medium | Show "offline" status after 5-min threshold, alert dispatcher |
| Large orgs (>500 vehicles) | Low | High | Paginate map view, virtual scrolling for list view |

---

## 11. Future Enhancements (Out of Scope for MVP)

### 11.1 Phase 2 (Q4)

- Multi-stop trip visualization on map
- Driver mobile app GPS reporting (iOS/Android)
- Offline map support (downloaded tiles for low-connectivity areas)
- Custom marker icons per vehicle type
- Export vehicle positions to CSV/GPX

### 11.2 Phase 3 (Q5-Q6)

- AI-powered route optimization
- Predictive ETA calculation with traffic data
- Real-time traffic layer on map
- Geofence templates (warehouses, customer sites, depots)
- Automated dispatcher notifications (SMS/Push on critical alerts)

### 11.3 Nice-to-Have

- 3D terrain view
- Street view integration
- Weather overlay
- Fuel station markers
- Toll plaza tracking (FASTag integration)

---

## 12. Open Questions

| Question | Owner | Target Resolution Date |
|----------|-------|----------------------|
| Should we support multi-vehicle selection for bulk actions? | Product | April 15, 2026 |
| What is the maximum number of geofences per org? | Engineering | April 15, 2026 |
| Do we need to support custom map providers (Mapbox, Google Maps)? | Product | April 22, 2026 |
| Should dispatcher be able to manually trigger geofence checks? | Engineering | April 22, 2026 |
| What is the retention period for telemetry data? | Product/Compliance | April 29, 2026 |

---

## 13. Appendix

### 13.1 Glossary

| Term | Definition |
|------|-----------|
| **Dispatcher** | Fleet operations user who assigns trips and monitors vehicles |
| **Telemetry** | GPS position + metadata (speed, heading, ignition) from vehicle device |
| **Live Location** | Most recent vehicle position (cached in Redis with 5-min TTL) |
| **Geofence** | Virtual geographic boundary (circle, polygon, or route corridor) |
| **History Trail** | Time-ordered sequence of vehicle positions over a time window |
| **WebSocket** | Persistent bidirectional connection for real-time push updates |
| **TimescaleDB** | PostgreSQL extension for time-series data (partitioned by time) |

### 13.2 Related Documents

- [FlyFleet Enterprise PRD](../docs/PRD.md)
- [Vehicle Tracking Spec v1.0](./VEHICLE_TRACKING.md) (existing spec)
- [Dispatch Flow Spec](./DISPATCH_FLOW.md)
- [IAM/Auth Spec](./IAM-AUTH.md)
- [Fleet-Core Integration Plan](../services/fleet-core/FLEET_CORE_INTEGRATION_PLAN.md)

### 13.3 Implementation Checklist

**Backend** (services/fleet-core):
- [ ] Create `VehicleLocationsHandler` with batch Redis + DB queries
- [ ] Add `GetOrgVehicles` method to `vehicle/repository/`
- [ ] Add `GetBatchLiveLocations` method to `tracking/repository/live_cache_repository.go`
- [ ] Enhance WebSocket handler with heartbeat + guaranteed 5s updates
- [ ] Integrate geofence alerts with WebSocket (Redis pub/sub)
- [ ] Add `GetHistoryWithSummary` to `TelemetryQueryService`
- [ ] Add `GetHistoryBucket` to `TelemetryRepository` (TimescaleDB time_bucket)
- [ ] Write unit tests for all new handlers/services
- [ ] Run load tests (1000 concurrent WebSocket connections)

**Frontend** (apps/dashboard-web):
- [ ] Add Leaflet map component to dashboard
- [ ] Implement vehicle markers with color-coding by status
- [ ] Add marker clustering for dense areas
- [ ] Build vehicle detail panel (slide-over)
- [ ] Implement search & filter bar
- [ ] Integrate WebSocket client with auto-reconnect + fallback
- [ ] Add history trail toggle with polyline overlay
- [ ] Add geofence zone overlays (toggleable)
- [ ] Implement auto-fit bounds on initial load
- [ ] Add loading states & error boundaries
- [ ] Write E2E tests (Playwright)

**Infrastructure**:
- [ ] Verify Redis pub/sub configured for org-scoped channels
- [ ] Verify NATS topic subscriptions (`fleet.telemetry.gps`, `fleet.geofence.events`)
- [ ] Configure TimescaleDB partitions for telemetry_points
- [ ] Set up monitoring dashboards (Grafana)
- [ ] Configure alerts for WebSocket connection failures, high latency

### 13.4 Route Registration (cmd/main.go)

Add to existing tracking route group in `services/fleet-core/cmd/main.go`:

```go
// Tracking endpoints (already enabled if feature flag is on)
if cfg.Flags.VehicleMapEnabled {
    // POST /api/v1/telemetry — GPS position ingest
    api.Post("/v1/telemetry", ingestHandler.ServeHTTP)

    // NEW: GET /api/v1/vehicles/locations — bulk org vehicle positions
    api.Get("/v1/vehicles/locations", vehicleLocationsHandler.ServeHTTP)

    // GET /api/v1/vehicles/{id}/telemetry/history — 24h history with summary
    if cfg.Flags.HistoryTrail {
        api.Get("/v1/vehicles/{id}/telemetry/history", historyHandler.ServeHTTP)
    }

    // WebSocket /ws/telemetry — real-time position push (enhanced)
    if cfg.Flags.RealTimeUpdates {
        api.Get("/ws/telemetry", wsHandler.ServeHTTP)
    }

    // GET/POST /api/v1/geofences — geofence management (already exists)
    api.Get("/v1/geofences", geofenceHandler.ServeHTTP)
    api.Post("/v1/geofences", geofenceHandler.ServeHTTP)
}
```

---

## 14. Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Product Manager | _Pending_ | | |
| Engineering Lead | _Pending_ | | |
| UX Designer | _Pending_ | | |
| QA Lead | _Pending_ | | |

---

**Document History**:
| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | April 8, 2026 | Product Team (AI-Assisted) | Initial spec creation based on user story |

---

**Next Steps**:
1. Review with engineering for feasibility assessment
2. Validate effort estimates (target: 8 person-sprints)
3. Create Jira/Linear tickets from implementation checklist
4. Schedule kickoff meeting with frontend + backend teams
5. Set up feature flags in fleet-core config
