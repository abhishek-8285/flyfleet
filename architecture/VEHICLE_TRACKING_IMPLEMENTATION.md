# Vehicle Tracking — Implementation Guide

**Document ID:** ARCH-VT-IMP-001
**Version:** 1.0
**Created:** April 7, 2026
**Parent Spec:** [Vehicle Tracking PSPEC](../product-spec/VEHICLE_TRACKING.md)
**Parent Architecture:** [Vehicle Tracking Architecture](../architecture/VEHICLE_TRACKING.md)

---

## 1. What This Document Is

This is the **build plan** for the Vehicle Tracking feature. It translates the product spec and architecture doc into:
- Concrete Go package structure for `services/fleet-core/`
- File-by-file implementation with key code snippets
- Database migration sequence
- Integration points with existing infrastructure

---

## 2. Service Directory Structure

```
services/fleet-core/
├── cmd/
│   └── main.go                          # Entry point, router setup, service wiring
├── internal/
│   ├── config/
│   │   └── config.go                    # Viper-based configuration
│   ├── database/
│   │   └── postgres.go                  # PostgreSQL + TimescaleDB connection pool
│   ├── middleware/
│   │   ├── jwks_auth.go                 # JWT validation via JWKS
│   │   ├── tenant_context.go            # Extract org_id from JWT → context
│   │   └── role_check.go                # RBAC role verification
│   ├── pkg/
│   │   └── distance/
│   │       └── haversine.go             # Haversine distance calculation
│   └── tracking/
│       ├── models/
│       │   ├── telemetry.go             # TelemetryEvent, LiveLocation, status constants
│       │   ├── geofence.go              # Geofence, GeofenceEvent models
│       │   └── websocket.go             # WS message types
│       ├── repository/
│       │   ├── telemetry.go             # TelemetryPoints writes + queries
│       │   ├── livecache.go             # Redis-backed LiveLocation cache
│       │   ├── geofence.go              # Geofence CRUD + spatial queries
│       │   └── vehicle.go               # Vehicle + driver enrichment queries
│       ├── service/
│       │   ├── ingestion.go             # Telemetry ingest → DB + Redis + Redis pub/sub
│       │   ├── query.go                 # Live location, history with auto-resolution
│       │   └── batch.go                 # Batch location fetch for bulk endpoint
│       ├── handlers/
│       │   ├── ingest.go                # POST /api/v1/telemetry
│       │   ├── locations.go             # GET /api/v1/vehicles/locations
│       │   ├── history.go               # GET /api/v1/vehicles/{id}/telemetry/history
│       │   ├── websocket.go             # WS /ws/telemetry
│       │   └── geofence.go              # GET/POST /api/v1/geofences
│       └── worker/
│           └── geofence_checker.go      # Background geofence detection worker
├── migrations/
│   └── 301_tracking_performance_indexes.sql
├── Dockerfile
├── go.mod
└── config.toml
```

---

## 3. File-by-File Implementation

### 3.1 Models Layer

| File | Key Types | Purpose |
|------|-----------|---------|
| `models/telemetry.go` | `TelemetryEvent`, `LiveLocation`, `VehicleStatus` constants, `VehicleLocationResponse` | Core data structures for GPS telemetry |
| `models/geofence.go` | `Geofence`, `GeofenceEvent`, `GeofenceType` constants | Geofence geometry and event models |
| `models/websocket.go` | `WSMessage`, `PositionUpdate`, `SubscribeRequest` | WebSocket message schemas |

### 3.2 Repository Layer

| File | Key Methods | Purpose |
|------|-------------|---------|
| `repository/telemetry.go` | `Write()`, `GetHistoryBucket()`, `GetLatestByVehicle()` | TimescaleDB telemetry operations |
| `repository/livecache.go` | `SetLiveLocation()`, `GetLiveLocation()`, `GetBatchLiveLocations()` | Redis cache for live positions |
| `repository/geofence.go` | `GetOrgGeofences()`, `CheckPolygon()`, `CheckCircle()`, `LogEvent()` | PostGIS spatial queries |
| `repository/vehicle.go` | `GetOrgVehicles()`, `GetDriverInfo()` | Vehicle + driver enrichment |

### 3.3 Service Layer

| File | Key Methods | Purpose |
|------|-------------|---------|
| `service/ingestion.go` | `Ingest()`, `PublishPositionUpdate()` | Validate, write, cache, broadcast |
| `service/query.go` | `GetLiveLocation()`, `GetHistoryWithSummary()`, `resolveResolution()` | Query logic with auto-resolution |
| `service/batch.go` | `GetAllOrgLocations()` | Bulk fetch + enrich for map load |

### 3.4 Handler Layer

| File | HTTP Method/Path | Purpose |
|------|-----------------|---------|
| `handlers/ingest.go` | POST `/api/v1/telemetry` | GPS position ingest |
| `handlers/locations.go` | GET `/api/v1/vehicles/locations` | Bulk org vehicle positions |
| `handlers/history.go` | GET `/api/v1/vehicles/{id}/telemetry/history` | 24h history with summary |
| `handlers/websocket.go` | WS `/ws/telemetry` | Real-time position push |
| `handlers/geofence.go` | GET/POST `/api/v1/geofences` | Geofence CRUD |

### 3.5 Worker

| File | Purpose |
|------|---------|
| `worker/geofence_checker.go` | Polls recent telemetry, checks against active geofences, logs events, publishes alerts |

---

## 4. Data Flow Sequence

### 4.1 Telemetry Ingest Flow

```
1. GPS device → POST /api/v1/telemetry
2. handlers/ingest.go validates request body
3. middleware.jwks_auth validates JWT
4. middleware.tenant_context extracts org_id → context
5. service/ingestion.go:
   a. Validates GPS coordinates (lat ∈ [-90,90], lng ∈ [-180,180])
   b. Deduplicates by sequence_no (if present)
   c. Writes to telemetry_points (TimescaleDB)
   d. Updates Redis live cache (TTL: 5 min)
   e. Publishes to Redis pub/sub channel "org:{org_id}:telemetry"
6. Returns 202 Accepted
```

### 4.2 Bulk Map Load Flow

```
1. Dashboard → GET /api/v1/vehicles/locations
2. middleware.jwks_auth validates JWT
3. middleware.tenant_context extracts org_id → context
4. handlers/locations.go calls service/batch.go.GetAllOrgLocations()
5. service/batch.go:
   a. Queries vehicle repo for org's vehicle list
   b. Batch-fetches live locations from Redis (pipeline)
   c. Falls back to DB for cache misses
   d. Enriches with driver names
   e. Calculates status (moving/idle/stopped/offline)
6. Returns JSON {data: [...], meta: {total, online, offline}}
```

### 4.3 WebSocket Push Flow

```
1. Dashboard → WS /ws/telemetry?token=xxx&org_id=yyy
2. handlers/websocket.go validates JWT from query param
3. Upgrades HTTP → WebSocket
4. Registers client in hub (org_id → clients map)
5. Subscribes to Redis channel "org:{org_id}:telemetry"
6. On Redis message → broadcast to all org's WS clients
7. Client disconnect → unregister, close channel
```

### 4.4 Geofence Detection Flow

```
1. worker/geofence_checker.go runs every 10s
2. Fetches recent telemetry points (last 10s) per vehicle
3. For each point:
   a. Check against all active org geofences (PostGIS)
   b. Compare with previous state in Redis (inside/outside)
   c. If state changed:
      - Log geofence_event to DB
      - Update Redis state
      - Publish alert to NATS
      - Push to WebSocket
```

---

## 5. Database Migration Sequence

```sql
-- Migration 301: Tracking Performance Indexes
-- Run online, CONCURRENTLY to avoid locking

-- 1. Latest position query optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_vehicle_time
    ON telemetry_points (vehicle_id, time DESC);

-- 2. Geofence spatial indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_polygon_gist
    ON geofences USING GIST (polygon);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_org_status
    ON geofences (org_id, status);

-- 3. Geofence event query optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_events_vehicle_time
    ON geofence_events (vehicle_id, occurred_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_events_geofence_time
    ON geofence_events (geofence_id, occurred_at DESC);

-- 4. Vehicle org lookup (should exist, ensure it does)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_org_id
    ON vehicles (org_id, status, deleted_at);

-- 5. Geofence event alert tracking
ALTER TABLE geofence_events
    ADD COLUMN IF NOT EXISTS alert_sent BOOLEAN DEFAULT false,
    ADD COLUMN IF NOT EXISTS alert_channels JSONB;
```

---

## 6. Performance Targets

| Endpoint | P95 Target | Strategy |
|----------|-----------|----------|
| GET /api/v1/vehicles/locations (500 vehicles) | <2s | Redis pipeline batch + DB fallback |
| GET /api/v1/vehicles/{id}/telemetry/history (24h) | <3s | TimescaleDB time_bucket + chunk pruning |
| WS /ws/telemetry (ingest → delivery) | <5s | Redis pub/sub, no polling |
| Geofence detection lag | <10s | 10s polling worker + PostGIS indexes |

---

## 7. Feature Flags

```go
type FeatureFlags struct {
    VehicleMapEnabled  bool `mapstructure:"vehicle_map_enabled"`   // Kill switch
    RealTimeUpdates    bool `mapstructure:"real_time_updates"`     // WS toggle
    GeofenceAlerts     bool `mapstructure:"geofence_alerts"`       // Geofence worker toggle
    HistoryTrail       bool `mapstructure:"history_trail"`         // History endpoint toggle
    MaxVehiclesOnMap   int  `mapstructure:"max_vehicles_on_map"`   // Performance guardrail (default 500)
}
```

---

## 8. Testing Strategy

### 8.1 Unit Tests

| Test File | Coverage Target | Scenarios |
|-----------|----------------|-----------|
| `handlers/locations_test.go` | >90% | Empty org, 1 vehicle, 500 vehicles, unauthorized |
| `handlers/history_test.go` | >85% | Auto-resolution, custom resolution, empty range, invalid params |
| `handlers/websocket_test.go` | >80% | Auth fail, connect, subscribe, disconnect, reconnect |
| `service/batch_test.go` | >90% | Redis hit, Redis miss, DB fallback, driver enrichment |
| `service/query_test.go` | >85% | Resolution calculation, summary stats (distance, speed) |
| `worker/geofence_checker_test.go` | >90% | Entry, exit, no change, boundary edge case |

### 8.2 Integration Tests

| Test | Setup | Validation |
|------|-------|------------|
| End-to-end ingest → WS | Mock GPS POST → WS client listening | Position received within 5s |
| Geofence alert delivery | Vehicle enters polygon geofence | geofence_event logged, NATS message published |
| History accuracy | Known 10-point GPS track | Summary distance matches Haversine within 1% |
| Multi-tenant isolation | 2 orgs, each with vehicles | Each org sees only their vehicles |

---

## 9. Open Decisions

| # | Question | Decision | Rationale |
|---|----------|----------|-----------|
| 1 | WebSocket library? | `gorilla/websocket` | Industry standard, well-maintained |
| 2 | Redis client? | `github.com/redis/go-redis/v9` | Official Go driver, pipeline support |
| 3 | Router? | `github.com/go-chi/chi/v5` | Lightweight, middleware-friendly |
| 4 | History resolution: client or server? | **Server-side auto** | Client shouldn't know DB performance details |
