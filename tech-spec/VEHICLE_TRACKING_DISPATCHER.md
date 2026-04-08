# Technical Specification: Vehicle Tracking & Dispatcher

**Document ID:** TECH-VT-001
**Version:** 1.0
**Status:** Draft
**Created:** April 8, 2026
**Author:** Senior Backend Engineer
**Parent Spec:** [Vehicle Tracking Dispatcher PSPEC](../product-spec/VEHICLE_TRACKING_DISPATCHER.md)
**Parent Architecture:** [Vehicle Tracking Architecture](../architecture/VEHICLE_TRACKING_ARCHITECTURE.md)
**Priority:** P0 — Q3 Delivery

---

## 1. Problem Statement & Goals

### 1.1 Problem

Fleet dispatchers currently lack real-time visibility into vehicle locations. They rely on:
- Phone calls to drivers (2–3 hours/day wasted)
- WhatsApp check-ins (unreliable, no audit trail)
- Manual paper logs (error-prone, delayed)

This creates **30–45 minute response times** to emergencies, customer dissatisfaction from inability to provide accurate ETAs, and revenue loss from inefficient route utilization.

### 1.2 User Story

> As a dispatcher, I want to see all my vehicles on a map with live positions so that I know where every vehicle is right now.

### 1.3 Goals & Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Map load time (bulk API) | **<2s P95** for 500 vehicles | Time from page load to all markers visible |
| Position update latency | **<5s P95** | Time from GPS ingest → map marker update |
| Bulk locations API latency | **<500ms P95** | GET /api/v1/vehicles/locations response time |
| WebSocket delivery latency | **<100ms P95** | Redis pub/sub → WS client delivery |
| History query (24h) | **<3s P95** | GET /api/v1/vehicles/{id}/telemetry/history |
| Daily active usage | **>80% of dispatchers** | Users who view map ≥1x per session |
| Max vehicles per org | **500** | Hard guardrail via feature flag |
| Concurrent WS connections | **1,000** | Per fleet-core instance |

### 1.4 Current State Summary

The fleet-core monolith at `services/fleet-core/` already has extensive tracking infrastructure:

- **Telemetry ingest** (`POST /api/v1/telemetry`) — HMAC validation, dedup, DB write (`telemetry_points`), Redis cache (`live:loc:{id}`), Redis pub/sub ✅
- **Single vehicle live location** (`GET /api/v1/vehicles/{id}/location`) — Redis → DB fallback ✅
- **Telemetry history** (`GET /api/v1/vehicles/{id}/history`) — Paginated query ✅
- **Map replay** (`GET /api/v1/vehicles/{id}/replay`) — Ordered telemetry, 24h/1000pt cap ✅
- **Nearby vehicles** (`GET /api/v1/vehicles/nearby`) — Spatial query, 50km cap ✅
- **WebSocket** (`WS /ws/telemetry`) — Real-time position push via Redis pub/sub (`org:{org}:telemetry`) ✅
- **Geofence engine** — PostGIS detection, NATS publish (`fleet.geofence.events`), Redis state persistence ✅
- **Batch locations service** — `BatchLocationService` with Redis pipeline + DB fallback ✅
- **Vehicle repository** — `GetOrgVehicles()`, `GetDriversBatch()` ✅

**Gaps to fill:**
1. WebSocket lacks heartbeat (ping/pong), guaranteed 5s updates, differential pushes
2. Geofence alerts not delivered via WebSocket to dashboard
3. History endpoint needs auto-resolution downsampling + summary stats integration
4. Frontend dashboard components not yet built

---

## 2. API Specifications

### 2.1 GET /api/v1/vehicles/locations — Bulk Org Vehicle Positions (EXISTING — ENHANCE)

**Status:** Already live via `internal/tracking/handlers/locations.go` and `internal/tracking/service/batch.go`.

**Current capabilities (live and verified):**
- Fetches all org vehicles via `VehicleRepository.GetOrgVehicles()` ✅
- Redis pipeline batch fetch via `LiveCacheRepository.GetBatchLiveLocations()` (key: `live:loc:{id}`) ✅
- DB fallback via `TelemetryRepository.GetBatchLatest()` ✅
- Driver enrichment via `VehicleRepository.GetDriversBatch()` ✅
- Status calculation (moving/idle/stopped/offline) via `models.CalculateStatus()` ✅
- Performance guardrail (`maxVehicles` cap) ✅

> ⚠️ **Pre-existing bug:** `locations.go:31` calls `middleware.GetOrgID(r)` passing `*http.Request`, but the actual `GetOrgID` in `bff_auth_gate.go` takes `context.Context`. Fix: `middleware.GetOrgID(r.Context())`.

**Current request:**
```http
GET /api/v1/vehicles/locations
Authorization: Bearer <JWT>
```
(No query parameters currently supported)

**Current response:**
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
      "heading": 135.0,
      "ignition": true,
      "fuel_level": 72.5,
      "odometer": 45230.5,
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
    "timestamp": "2026-04-08T10:30:05Z"
  }
}
```

**Enhancements to add:**

#### New Query Parameters

| Param | Type | Description | Default |
|-------|------|-------------|---------|
| `status` | string | Filter by status: `moving`, `idle`, `stopped`, `offline` | all |
| `driver_id` | UUID | Filter by assigned driver | all |
| `search` | string | Search by registration number or driver name | — |

#### Enhanced Response Meta

Add breakdown counts to meta:
```json
{
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

#### Auth & Authorization

| Layer | Middleware | Detail |
|-------|-----------|--------|
| Auth | `JWKSAuthGate(cfg.JWTSecret, logger)` | JWT validation via HMAC-SHA256 |
| Tenant | `TenantMiddleware(logger)` | Extracts `org_id` from JWT claims → context *(note: `cmd/main.go` currently references `TenantContextMiddleware` which doesn't compile — actual function is `TenantMiddleware`)* |
| RBAC | `RequireRole("dispatcher", "fleet_manager", "admin")` | Read-only role check |
| Rate Limit | `RateLimit` (shared) | 60 req/min per user |

#### Performance Budget

| Stage | Target | Strategy |
|-------|--------|----------|
| Vehicle query | <100ms | `GetOrgVehicles()` — indexed on `org_id, status` |
| Redis batch get | <50ms | Pipeline 500 keys in single round-trip |
| DB fallback | <300ms | `GetBatchLatest()` — `DISTINCT ON (vehicle_id)` |
| Driver enrichment | <200ms | `GetDriversBatch()` — `ANY($1)` |
| JSON encode | <50ms | Pre-allocated slice, no reflection |
| **Total P95** | **<500ms** | Well within 2s budget for network + frontend |

#### Error Responses

| Status | Body | Condition |
|--------|------|-----------|
| 401 | `{"error":"missing organization context"}` | No org in JWT or context |
| 403 | `{"error":"insufficient permissions"}` | RBAC role check failed |
| 429 | `{"error":"rate limit exceeded"}` | >60 req/min |
| 500 | `{"error":"internal error"}` | DB/Redis unavailable |

---

### 2.2 GET /api/v1/vehicles/{id}/telemetry/history — Auto-Resolution + Summary (EXISTING — ENHANCE)

**Status:** Handler and service already exist at:
- `internal/tracking/handlers/history.go`
- `internal/tracking/service/query.go` (`GetHistoryWithSummary`)
- `internal/tracking/repository/telemetry.go` (`GetHistoryBucket`)

**Current capabilities:**
- `GetHistoryWithSummary(ctx, vehicleID, start, end, resolution)` — full implementation
- `resolveResolution(start, end, requested)` — auto-resolution logic ✅
- `calculateSummary(points)` — Haversine distance, time classification, speed aggregates ✅
- `GetHistoryBucket(ctx, vehicleID, start, end, bucketInterval)` — TimescaleDB `time_bucket` ✅

**Enhancements to add:**

#### Query Parameters

| Param | Type | Required | Description | Default |
|-------|------|----------|-------------|---------|
| `start` | ISO 8601 | Yes | Start time | — |
| `end` | ISO 8601 | Yes | End time | — |
| `resolution` | string | No | `auto`, `10s`, `30s`, `60s`, `5m`, `15m` | `auto` |
| `include_summary` | bool | No | Include distance/speed stats | `true` |

#### Auto-Resolution Table

| Time Range | Effective Resolution | Max Points |
|------------|---------------------|------------|
| < 1 hour | 10 seconds | 360 |
| 1–6 hours | 30 seconds | 720 |
| 6–24 hours | 60 seconds | 1,440 |
| 1–7 days | 5 minutes | 2,016 |
| 7–30 days | 15 minutes | 2,880 |

*(Already implemented in `query.go:resolveResolution`)*

#### Response Schema

```json
{
  "vehicle_id": "v-123-abc",
  "points": [
    {
      "event_id": "00000000-0000-0000-0000-000000000000",
      "device_id": "00000000-0000-0000-0000-000000000000",
      "vehicle_id": "v-123-abc",
      "tenant_id": "org-uuid",
      "latitude": 28.6139,
      "longitude": 77.2090,
      "speed_kmh": 32.1,
      "heading": 140.0,
      "ignition": true,
      "fuel_level": 72.5,
      "odometer": 45230.5,
      "sequence_no": 0,
      "recorded_at": "2026-04-07T10:01:00Z",
      "received_at": "2026-04-07T10:01:00Z",
      "raw_data": ""
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
  },
  "meta": {
    "point_count": 1440,
    "resolution_used": "60 seconds",
    "time_range_hours": 24.0
  }
}
```

#### Auth Requirements

| Layer | Middleware |
|-------|-----------|
| Auth | `JWKSAuthGate(cfg.JWTSecret, logger)` |
| Tenant | `TenantMiddleware(logger)` |
| RBAC | `RequireRole("dispatcher", "fleet_manager", "admin")` |
| Tenant isolation | Vehicle must belong to requesting org (enforced in handler) |

#### Performance Budget

| Stage | Target |
|-------|--------|
| Resolution calc | <1ms |
| TimescaleDB `time_bucket` query | <2s (chunk pruning + index) |
| Summary calculation | <50ms (in-memory) |
| JSON encode (1,440 pts) | <50ms |
| **Total P95** | **<3s** |

---

### 2.3 WS /ws/telemetry — Real-Time Position Push (EXISTING — ENHANCE)

**Status:** Already live at `internal/tracking/handlers/websocket.go`.

**Current capabilities:**
- JWT auth via query param (`?token=xxx`)
- Org-scoped client registry
- Redis pub/sub subscription to `org:{org_id}:telemetry`
- Vehicle subscription filtering (client → server `subscribe_vehicles`)
- Connection ack on connect

**Gaps identified in current implementation:**
1. **No heartbeat/ping-pong** — dead connections not detected
2. **No guaranteed 5s position updates** — relies solely on Redis pub/sub (which only fires on ingest)
3. **No differential pushes** — sends full position even if unchanged
4. **No geofence alert delivery** — only receives telemetry, not geofence events
5. **No exponential backoff reconnect guidance** — client must implement independently

#### Enhanced Protocol

**Connection URL:**
```
wss://api.flyfleet.io/ws/telemetry?token=<JWT>&org_id=<org-uuid>
```

**Server → Client Message Types:**

##### 1. Connection Ack (existing)
```json
{
  "type": "connection_ack",
  "data": {
    "status": "connected",
    "heartbeat_interval_ms": 30000,
    "position_update_interval_ms": 5000
  }
}
```

##### 2. Position Update (enhanced — differential push)
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

##### 3. Geofence Alert (NEW)
```json
{
  "type": "geofence_alert",
  "data": {
    "vehicle_id": "v-123-abc",
    "registration": "DL01AB1234",
    "geofence_id": "g-789-xyz",
    "geofence_name": "Delhi Warehouse",
    "event_type": "enter",
    "latitude": 28.6139,
    "longitude": 77.2090,
    "timestamp": "2026-04-08T10:35:00Z"
  }
}
```

##### 4. Pong (NEW — heartbeat response)
```json
{
  "type": "pong",
  "data": {
    "timestamp": "2026-04-08T10:30:00Z"
  }
}
```

**Client → Server Message Types:**

##### 1. Subscribe to Vehicles (existing)
```json
{
  "type": "subscribe_vehicles",
  "data": {
    "vehicle_ids": ["v-123-abc", "v-456-def"]
  }
}
```

##### 2. Ping (NEW — heartbeat)
```json
{
  "type": "ping"
}
```

#### Implementation Enhancements

**File:** `internal/tracking/handlers/websocket.go`

**Changes:**

1. **Add heartbeat mechanism** — server sends ping every 30s, expects pong within 10s
2. **Add guaranteed 5s position ticker** — independent of Redis pub/sub, fetches from Redis cache and pushes only changed positions
3. **Subscribe to geofence alerts** — add second Redis pub/sub subscription to `org:{org_id}:geofence_alerts`
4. **Differential push** — track `lastSent` map per client, only send if position changed beyond threshold (lat/lng delta > 0.0001 ≈ 11m, speed delta > 2 km/h)

```go
type WSClient struct {
    conn       *websocket.Conn
    orgID      string
    send       chan []byte
    subscribed map[string]bool
    lastSent   map[string]*models.LiveLocation // differential tracking
    lastPong   time.Time                       // heartbeat health
}

// StartHeartbeat sends WebSocket ping frames every 30s.
func (h *WebSocketHandler) startHeartbeat(client *WSClient) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            if time.Since(client.lastPong) > 90*time.Second {
                // No pong received in 90s — connection dead
                client.conn.Close()
                return
            }
            client.conn.WriteMessage(websocket.PingMessage, nil)
        case <-client.conn.CloseChan():
            return
        }
    }
}

// StartPositionTicker pushes positions every 5s regardless of ingest frequency.
func (h *WebSocketHandler) startPositionTicker(ctx context.Context, client *WSClient) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            h.pushOrgPositions(ctx, client)
        case <-ctx.Done():
            return
        }
    }
}

func (h *WebSocketHandler) pushOrgPositions(ctx context.Context, client *WSClient) {
    // 1. Fetch all org vehicle IDs
    vehicles, err := h.vehicleRepo.GetOrgVehicles(ctx, orgUUID)
    // 2. Batch fetch from Redis
    locations, _ := h.liveCache.GetBatchLiveLocations(ctx, vehicleIDs)
    // 3. Differential push
    for vid, loc := range locations {
        if last, ok := client.lastSent[vid]; ok {
            if !hasSignificantChange(last, loc) {
                continue
            }
        }
        msg := models.WSMessage{
            Type: models.WSMsgPositionUpdate,
            Data: models.PositionUpdate{...},
        }
        data, _ := json.Marshal(msg)
        select {
        case client.send <- data:
            client.lastSent[vid] = &loc
        default:
            // Buffer full — skip, next tick will catch up
        }
    }
}

func hasSignificantChange(old, new models.LiveLocation) bool {
    latDelta := math.Abs(new.Lat - old.Lat)
    lngDelta := math.Abs(new.Lng - old.Lng)
    speedDelta := math.Abs(new.Speed - old.Speed)
    return latDelta > 0.0001 || lngDelta > 0.0001 || speedDelta > 2.0 || old.Status != new.Status
}
```

**Geofence Alert Subscription (in `subscribeToOrgUpdates`):**

```go
// Add second subscription for geofence alerts
geoPubsub := h.redis.Subscribe(ctx, fmt.Sprintf("org:%s:geofence_alerts", orgID))
defer geoPubsub.Close()
geoCh := geoPubsub.Channel()

// Merge both channels with select
for {
    select {
    case msg := <-ch:        // telemetry updates
        h.forwardToClient(client, msg.Payload)
    case msg := <-geoCh:     // geofence alerts
        h.forwardGeofenceAlert(client, msg.Payload)
    case <-ctx.Done():
        return
    }
}
```

---

### 2.4 Geofence Alerts → WebSocket Fan-Out (NEW INTEGRATION)

**Current state:** The `GeofenceEngine` (`internal/domains/geofencing/engine/geofence_engine.go`) already:
- Detects entry/exit/dwell events via PostGIS
- Inserts to `geofence_events` table
- Publishes to NATS `fleet.geofence.events`

**Gap:** Does NOT publish to Redis pub/sub for WebSocket delivery.

**Enhancement to `geofence_engine.go`:**

In `triggerGeofenceEvent()`, after NATS publish, also publish to Redis:

```go
func (ge *GeofenceEngine) triggerGeofenceEvent(ctx context.Context, vehicleID, geofenceID uuid.UUID, geofenceName, eventType string, event GPSEvent) {
    // ... existing DB insert and NATS publish ...

    // NEW: Publish to Redis for WebSocket fan-out
    orgID := ge.getOrgIDForVehicle(ctx, vehicleID)
    if orgID != uuid.Nil && ge.redis != nil {
        geoMsg := map[string]interface{}{
            "type":            "geofence_alert",
            "vehicle_id":      vehicleID.String(),
            "geofence_id":     geofenceID.String(),
            "geofence_name":   geofenceName,
            "event_type":      eventType,
            "latitude":        event.Latitude,
            "longitude":       event.Longitude,
            "timestamp":       time.Unix(event.Timestamp, 0).UTC().Format(time.RFC3339),
        }
        data, _ := json.Marshal(geoMsg)
        channel := fmt.Sprintf("org:%s:geofence_alerts", orgID.String())
        ge.redis.Publish(ctx, channel, string(data)).Err()
    }
}
```

**Required change to `GeofenceEngine` constructor:** Add `getOrgIDForVehicle` method:

```go
func (ge *GeofenceEngine) getOrgIDForVehicle(ctx context.Context, vehicleID uuid.UUID) uuid.UUID {
    var orgID uuid.UUID
    err := ge.db.QueryRow(ctx,
        `SELECT tenant_id FROM vehicles WHERE id = $1 AND deleted_at IS NULL`,
        vehicleID).Scan(&orgID)
    if err != nil {
        ge.logger.Warn("failed to resolve org_id for vehicle", zap.String("vehicle_id", vehicleID.String()), zap.Error(err))
        return uuid.Nil
    }
    return orgID
}
```

---

## 3. Domain Impact Analysis

### 3.1 Tracking Domain (`internal/tracking/`)

This is the primary domain for all changes. It already contains the canonical implementations used by `cmd/main.go`.

#### New Files to Create

| File Path | Purpose |
|-----------|---------|
| *(None needed)* | All required files already exist |

#### Existing Files to Modify

| File | Change | Detail |
|------|--------|--------|
| `handlers/locations.go` | **ENHANCE** — Add query parameter support | Add `status`, `driver_id`, `search` filters. Pass to `BatchLocationService`. Add moving/idle/stopped counts to meta. |
| `handlers/websocket.go` | **ENHANCE** — Add heartbeat, position ticker, geofence subscription | Add `startHeartbeat()`, `startPositionTicker()`, differential push via `lastSent` map, subscribe to `org:{org}:geofence_alerts` channel. |
| `handlers/history.go` | **ENHANCE** — Wire up `GetHistoryWithSummary` | Already exists. Ensure query params `resolution` and `include_summary` are parsed and passed. Add tenant isolation check. |
| `service/batch.go` | **ENHANCE** — Add filtering support | Add `GetFilteredOrgLocations(ctx, orgID, statusFilter, driverID, search)` method that filters after enrichment. |
| `service/query.go` | **VERIFY** — `GetHistoryWithSummary` already implemented | Confirm tenant isolation is enforced (vehicle must belong to org). |
| `repository/livecache.go` | **NO CHANGE** — `GetBatchLiveLocations()` already exists and works | |
| `repository/telemetry.go` | **NO CHANGE** — `GetHistoryBucket()` and `GetBatchLatest()` already work | |
| `repository/vehicle.go` | **NO CHANGE** — `GetOrgVehicles()` and `GetDriversBatch()` already work | |
| `models/telemetry.go` | **NO CHANGE** — All types defined | |
| `models/websocket.go` | **ENHANCE** — Add `WSMsgPing`, `WSMsgPong` constants | |
| `models/geofence.go` | **VERIFY** — Ensure `GeofenceWSMessage` type matches WS protocol | |

#### Interface Changes

**`BatchLocationService`** — Add filtered variant:

```go
// GetFilteredOrgLocations returns enriched positions with optional filters.
func (s *BatchLocationService) GetFilteredOrgLocations(
    ctx context.Context,
    orgID uuid.UUID,
    statusFilter string,   // "" = all, or "moving"/"idle"/"stopped"/"offline"
    driverID uuid.UUID,    // uuid.Nil = all
    search string,         // "" = no search
) ([]models.VehicleLocationResponse, error)
```

**`WebSocketHandler`** — New methods on `WSClient`:

```go
type WSClient struct {
    // ... existing fields ...
    lastSent map[string]*models.LiveLocation // differential push
    lastPong time.Time                       // heartbeat tracking
}
```

---

### 3.2 Geofencing Domain (`internal/domains/geofencing/`)

#### Existing Files to Modify

| File | Change | Detail |
|------|--------|--------|
| `engine/geofence_engine.go` | **ENHANCE** — Add Redis pub/sub for geofence alerts | Add `redis *redis.Client` field (already available via `NewGeofenceEngineWithRedis`). Add `getOrgIDForVehicle()` method. In `triggerGeofenceEvent()`, publish to `org:{org_id}:geofence_alerts` Redis channel after NATS publish. |

**No new files needed.**

---

### 3.3 Vehicle Domain (`internal/domains/vehicle/`)

**No changes needed.** The tracking domain's `VehicleRepository` (`internal/tracking/repository/vehicle.go`) already provides:
- `GetOrgVehicles()` — returns all active vehicles for an org
- `GetDriversBatch()` — batch driver enrichment

The vehicle domain's own repository (`internal/domains/vehicle/repository/vehicle_repository.go`) has `GetByOrgID()` which is the source of truth, but the tracking layer already has its own lightweight `GetOrgVehicles()` query that's sufficient for the bulk locations endpoint.

---

### 3.4 Driver Domain (`internal/domains/driver/`)

**No changes needed.** The tracking layer's `VehicleRepository.GetDriversBatch()` already handles batch driver queries via `ANY($1)`.

---

### 3.5 Alerts Domain (`internal/domains/alerts/`)

**No changes needed for WebSocket fan-out.** Geofence alerts are delivered via:
1. NATS `fleet.geofence.events` → *(⚠️ no subscriber found in codebase)*
2. Redis `org:{org_id}:geofence_alerts` → WebSocket (for in-dashboard notification) **(NEW — must be added to GeofenceEngine)**

The alerts domain handles multi-channel dispatch independently. The WebSocket path bypasses the alerts domain for **in-dashboard real-time display only**.

---

## 4. Data Flow Sequences

### 4.1 Telemetry Ingest → Redis Cache → Redis Pub/Sub → WebSocket Delivery

```
1. GPS device sends POST /api/v1/telemetry with HMAC-signed payload
2. IngestHandler validates request body, extracts tenant from context
3. IngestionService:
   a. Validates HMAC signature
   b. Checks timestamp drift (<300s)
   c. Validates coordinates and speed
   d. Checks device registration
   e. Deduplicates by event_id
   f. Writes to telemetry_points (TimescaleDB)
      NOTE: table defined in migrations/072_tracking_enterprise.sql
   g. Updates Redis live cache (live:loc:{id}, TTL 5min)
   h. Buffers to NATS "telemetry.raw" (JetStream) when BufferEnabled=true
   i. Publishes position update JSON to Redis pub/sub "org:{org_id}:telemetry"
4. WebSocketHandler (subscribed to Redis "org:{org_id}:telemetry") receives message
5. WS handler broadcasts to all org's connected WS clients
6. Dashboard Leaflet marker animates to new position
```

> ⚠️ **NATS correction:** IngestionService publishes to `telemetry.raw` (JetStream), NOT `fleet.telemetry.gps`. The `fleet.telemetry.gps` topic is only **subscribed to** (by GeofenceEngine, TelemetryWorker) — no code in fleet-core publishes to it. Geofence detection currently runs via the `GeofenceChecker` worker polling DB, not via NATS stream.

**Latency breakdown (corrected — no NATS hop in real-time path):**
- Ingest → DB write: ~10ms
- DB → Redis cache: ~5ms
- Redis pub/sub → WS handler: ~5ms
- WS handler → dashboard: ~10ms
- **Total: ~30ms** (well within 5s budget)

### 4.2 Bulk Map Load → Redis Pipeline → DB Fallback → Driver Enrichment → JSON

```
1. Dashboard loads → GET /api/v1/vehicles/locations
2. LocationsHandler extracts org_id from context (TenantMiddleware)
   ⚠️ Actual call should be: middleware.GetOrgID(r.Context())
3. BatchLocationService.GetAllOrgLocations():
   a. VehicleRepository.GetOrgVehicles(orgID) → all active vehicles
      (Query: SELECT id, registration, make, model, current_driver_id
              FROM vehicles WHERE org_id = $1 AND status = 'active')
   b. Collect vehicle IDs + driver IDs
   c. LiveCacheRepository.GetBatchLiveLocations(vehicleIDs)
      → Redis HGETALL pipeline (500 keys in single round-trip ~50ms)
      → Keys: live:loc:{vehicleID}
   d. For cache misses: TelemetryRepository.GetBatchLatest(missIDs)
      → SELECT DISTINCT ON (vehicle_id) ... ORDER BY vehicle_id, time DESC
   e. VehicleRepository.GetDriversBatch(driverIDs)
      → SELECT id, name, phone FROM drivers WHERE id = ANY($1)
   f. Calculate status per vehicle (moving/idle/stopped/offline)
   g. Assemble VehicleLocationResponse slice
4. Return JSON {data: [...], meta: {total, online, offline, moving, idle, stopped}}
```

**Latency breakdown (500 vehicles):**
- Vehicle query: ~100ms
- Redis pipeline: ~50ms
- DB fallback (cache misses): ~300ms
- Driver enrichment: ~200ms
- Status calculation + assembly: ~50ms
- JSON encode: ~50ms
- **Total P95: ~750ms** (within 2s budget)

### 4.3 Geofence Detection → NATS → AlertService → Multi-Channel Dispatch

> ⚠️ **Current reality:** The `fleet.telemetry.gps` NATS topic has **no publisher in fleet-core**. GeofenceEngine subscribes to it but nothing publishes within this codebase. The `GeofenceChecker` worker in `cmd/main.go` implements geofence detection via **DB polling** (every 10s), not via NATS stream. This section describes the **intended** architecture; the current implementation differs.

**Intended flow (when NATS publisher is wired):**
```
1. GPS device → POST /api/v1/telemetry
2. IngestionService publishes GPS event to NATS (topic TBD — currently "telemetry.raw")
3. GeofenceEngine (NATS subscriber) receives GPS event
4. Engine queries active geofences assigned to vehicle:
   SELECT g.id, g.name, g.type, g.center, g.radius_meters, g.coordinates
   FROM geofences g
   JOIN geofence_assignments ga ON ga.geofence_id = g.id
   WHERE ga.vehicle_id = $1 AND ga.is_active = true AND g.enabled = true
5. PostGIS containment check per geofence:
   - Circle: ST_DWithin(center::geography, point::geography, radius_meters)
   - Polygon: ST_Within(point::geography, polygon::geography)
6. Compare with previous state (Redis: serialized VehicleState, 30-day TTL)
7. If state changed:
   - INSERT INTO geofence_events (time, vehicle_id, geofence_id, event_type, location)
   - Publish to NATS "fleet.geofence.events"
   - **NEW** Publish to Redis "org:{org_id}:geofence_alerts" (for WebSocket delivery)
8. AlertService (NATS subscriber) consumes event from "fleet.geofence.events"
   ⚠️ NOTE: No subscriber to "fleet.geofence.events" was found in the codebase.
   This is a gap — alerts may not be delivered via the multi-channel path.
9. PreferenceEngine looks up user's notification preferences
10. TemplateEngine renders alert message
11. NotificationWorker dispatches via email/SMS/push/in-app
```

### 4.4 Geofence Alert → Redis Pub/Sub → WebSocket → Dashboard Notification **(NEW)**

> ⚠️ **Not yet implemented.** The `org:{org_id}:geofence_alerts` channel does not exist. Both the GeofenceEngine (publisher) and WebSocketHandler (subscriber) must be enhanced.

```
1. GeofenceEngine detects entry/exit → publishes to Redis "org:{org_id}:geofence_alerts" (NEW in geofence_engine.go)
2. WebSocketHandler (subscribed to same channel) receives message (NEW in websocket.go)
3. WS handler wraps message as WSMessage{type: "geofence_alert", data: {...}}
4. WS handler sends to all org's connected WS clients
5. Dashboard Redux middleware receives message
6. Redux dispatches addGeofenceEvent action
7. UI shows toast notification + updates geofence event log
```

### 4.5 History Query → Auto-Resolution → TimescaleDB time_bucket → Summary Stats

```
1. Dashboard → GET /api/v1/vehicles/{id}/telemetry/history?start=...&end=...&resolution=auto
2. HistoryHandler extracts vehicle_id from URL path
   ⚠️ PRE-EXISTING BUG: history.go calls extractPathParam(r, "id") which
      does NOT exist. Fix: use chi.URLParam(r, "id") + import "github.com/go-chi/chi/v5"
3. org_id from context via middleware.GetOrgID(r.Context())
4. Validates vehicle belongs to org (tenant isolation — NOT YET implemented)
5. TelemetryQueryService.GetHistoryWithSummary(vehicleID, start, end, "auto"):
   a. resolveResolution(start, end, "auto") → determines bucket interval
      (e.g., 24h range → "60 seconds")
   b. TelemetryRepository.GetHistoryBucket(vehicleID, start, end, "60 seconds")
      → TimescaleDB query:
        SELECT time_bucket('60 seconds', time) AS bucket_time,
               AVG(latitude), AVG(longitude), AVG(speed_kmh),
               AVG(heading), BOOL_OR(ignition), MAX(speed_kmh)
        FROM telemetry_points
        WHERE vehicle_id = $1 AND time >= $2 AND time <= $3
        GROUP BY bucket_time
        ORDER BY bucket_time
   c. calculateSummary(points):
      - Haversine distance between consecutive points
      - Classify segments as moving (speed>5), idle (0<speed≤5), stopped (speed=0)
      - Compute max/avg speed
   d. Assemble HistoryResult{points, summary, meta}
6. Return JSON response
```

---

## 5. Data Model Changes

### 5.1 Database Migration

> ⚠️ **Table creation status from audit:**
> - `telemetry_points` — Created in `fleet-core/migrations/072_tracking_enterprise.sql` (NOT in legacy infra migrations)
> - `gps_locations` — Created in legacy `infrastructure/migrations/032_telemetry_optimization.sql`
> - `vehicles`, `drivers`, `geofences` — Created in legacy infra migrations ✅
> - `geofence_events` — **NOT found** in any migration file. Referenced in Go code and docs only. **Must be created.**
> - `geofence_assignments` — **NOT found** in any migration file. Referenced in Go code and docs only. **Must be created.**
> - `telemetry_outbox` — **NOT found** in any migration file. Referenced in Go code and docs only. **Must be created.**

**File:** `services/fleet-core/migrations/302_vehicle_tracking_enrichment.sql`

```sql
-- Migration 302: Vehicle Tracking Enrichment
-- Run online, CONCURRENTLY to avoid locking

-- 1. Ensure gps_locations table exists (upsert table for latest positions)
CREATE TABLE IF NOT EXISTS gps_locations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id      UUID NOT NULL UNIQUE REFERENCES vehicles(id),
    device_id       UUID,
    latitude        DOUBLE PRECISION NOT NULL,
    longitude       DOUBLE PRECISION NOT NULL,
    speed_kmh       DOUBLE PRECISION,
    heading         DOUBLE PRECISION,
    ignition        BOOLEAN,
    fuel_level      DOUBLE PRECISION,
    odometer_km     DOUBLE PRECISION,
    last_seen       TIMESTAMPTZ NOT NULL,
    status          VARCHAR(20) DEFAULT 'online',
    tenant_id       UUID NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Performance index on vehicles for bulk org query
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vehicles_org_status_active
    ON vehicles (tenant_id, status)
    WHERE deleted_at IS NULL AND status = 'active';

-- 3. Index for driver batch lookup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_drivers_id_lookup
    ON drivers (id) INCLUDE (name, phone);

-- 4. Geofence event query optimization (org-scoped)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_events_org_time
    ON geofence_events (tenant_id, time DESC);

-- 5. Ensure geofence_assignments has proper indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_assignments_vehicle_active
    ON geofence_assignments (vehicle_id, is_active);

-- 6. Telemetry batch latest optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_vehicle_time_desc
    ON telemetry_points (vehicle_id, time DESC);

-- 7. Upsert function for gps_locations (called by ingestion service)
CREATE OR REPLACE FUNCTION upsert_gps_location(
    p_vehicle_id UUID,
    p_device_id UUID,
    p_lat DOUBLE PRECISION,
    p_lng DOUBLE PRECISION,
    p_speed DOUBLE PRECISION,
    p_heading DOUBLE PRECISION,
    p_ignition BOOLEAN,
    p_fuel DOUBLE PRECISION,
    p_odometer DOUBLE PRECISION,
    p_last_seen TIMESTAMPTZ,
    p_tenant_id UUID
) RETURNS VOID AS $$
BEGIN
    INSERT INTO gps_locations (
        vehicle_id, device_id, latitude, longitude, speed_kmh,
        heading, ignition, fuel_level, odometer_km, last_seen,
        status, tenant_id
    ) VALUES (
        p_vehicle_id, p_device_id, p_lat, p_lng, p_speed,
        p_heading, p_ignition, p_fuel, p_odometer, p_last_seen,
        'online', p_tenant_id
    )
    ON CONFLICT (vehicle_id) DO UPDATE SET
        device_id     = EXCLUDED.device_id,
        latitude      = EXCLUDED.latitude,
        longitude     = EXCLUDED.longitude,
        speed_kmh     = EXCLUDED.speed_kmh,
        heading       = EXCLUDED.heading,
        ignition      = EXCLUDED.ignition,
        fuel_level    = EXCLUDED.fuel_level,
        odometer_km   = EXCLUDED.odometer_km,
        last_seen     = EXCLUDED.last_seen,
        status        = 'online',
        updated_at    = NOW();
END;
$$ LANGUAGE plpgsql;
```

### 5.2 Redis Key Schema

| Key | Type | TTL | Content |
|-----|------|-----|---------|
| `live:loc:{vehicle_id}` | Hash | 5 min | Live location (lat, lng, speed, heading, ignition, fuel, odometer, status, last_seen) |
| `vehicle_state:{vehicle_id}` | JSON string | 30 days | Geofence containment state (serialized VehicleState) |
| `org:{org_id}:telemetry` | Pub/Sub channel | N/A | Real-time position fan-out |
| `org:{org_id}:geofence_alerts` | Pub/Sub channel | N/A | Geofence alert fan-out (NEW) |

**No new Redis key patterns needed.** The existing schema supports all requirements.

### 5.3 NATS Topics

| Topic | Publisher | Consumer | Status |
|-------|-----------|----------|--------|
| `telemetry.raw` | IngestionService (PublishJS) | TelemetrySink worker | ✅ Existing |
| `fleet.telemetry.gps` | ⚠️ **No publisher in fleet-core** | GeofenceEngine, TelemetryWorker | ⚠️ Subscribed only |
| `fleet.geofence.events` | GeofenceEngine | ⚠️ **No subscriber found** | ⚠️ Published only |
| `fleet.alerts.dispatch` | ❌ Does not exist | ❌ | ❌ **Phantom — remove all references** |

**Corrections from codebase audit:**
- `fleet.telemetry.gps` — GeofenceEngine subscribes but **nothing in fleet-core publishes** to it. The GeofenceChecker worker in `cmd/main.go` uses DB polling instead.
- `fleet.geofence.events` — GeofenceEngine publishes but **no subscriber found** in fleet-core. AlertService may not be consuming this topic.
- `telemetry.raw` — This is the **actual** NATS topic used by IngestionService for buffered ingest.
- `fleet.alerts.dispatch` — **Does not exist** anywhere. Remove all references.

**NEW topic needed:** None — the geofence alert path uses **Redis pub/sub**, not NATS.

### 5.4 Performance Indexes Summary

| Index | Table | Columns | Type | Purpose |
|-------|-------|---------|------|---------|
| `idx_vehicles_org_status_active` | vehicles | (tenant_id, status) | BTree (partial) | Bulk org vehicle query |
| `idx_drivers_id_lookup` | drivers | (id) INCLUDE (name, phone) | Covering | Batch driver enrichment |
| `idx_geofence_events_org_time` | geofence_events | (tenant_id, time DESC) | BTree | Org-scoped event queries |
| `idx_geofence_assignments_vehicle_active` | geofence_assignments | (vehicle_id, is_active) | BTree | Geofence engine vehicle lookup |
| `idx_telemetry_vehicle_time_desc` | telemetry_points | (vehicle_id, time DESC) | BTree | Batch latest telemetry |

---

## 6. Implementation Plan (File-by-File)

All paths relative to `services/fleet-core/`.

### 6.1 Tracking Handlers

#### `internal/tracking/handlers/locations.go` — ENHANCE

**What:** Add query parameter support (`status`, `driver_id`, `search`) and enhanced meta counts.

**Changes:**
```go
func (h *LocationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // ... existing auth/org extraction ...

    // NEW: Parse query parameters
    statusFilter := r.URL.Query().Get("status")
    driverIDStr := r.URL.Query().Get("driver_id")
    search := r.URL.Query().Get("search")

    var driverID uuid.UUID
    if driverIDStr != "" {
        driverID, _ = uuid.FromString(driverIDStr)
    }

    // Use filtered method if any filter is set
    var locations []models.VehicleLocationResponse
    var err error
    if statusFilter != "" || driverID != uuid.Nil || search != "" {
        locations, err = h.batchService.GetFilteredOrgLocations(
            r.Context(), orgID, statusFilter, driverID, search)
    } else {
        locations, err = h.batchService.GetAllOrgLocations(r.Context(), orgID)
    }

    // ... existing error handling ...

    // ENHANCED: Count by status category
    online, offline, moving, idle, stopped := 0, 0, 0, 0, 0
    for _, loc := range locations {
        switch loc.Status {
        case models.StatusMoving: moving++; online++
        case models.StatusIdle: idle++; online++
        case models.StatusStopped: stopped++; online++
        default: offline++
        }
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "data": locations,
        "meta": map[string]interface{}{
            "total":     len(locations),
            "online":    online,
            "offline":   offline,
            "moving":    moving,
            "idle":      idle,
            "stopped":   stopped,
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        },
    })
}
```

#### `internal/tracking/handlers/websocket.go` — ENHANCE

**What:** Add heartbeat, guaranteed 5s position ticker, differential push, geofence alert subscription.

**Changes to `WSClient` struct:**
```go
type WSClient struct {
    conn       *websocket.Conn
    orgID      string
    send       chan []byte
    subscribed map[string]bool
    lastSent   map[string]*models.LiveLocation // differential push tracking
    lastPong   time.Time                        // heartbeat health
}
```

**Changes to `ServeHTTP`:**
```go
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // ... existing auth and upgrade ...

    client := &WSClient{
        conn:       conn,
        orgID:      orgID,
        send:       make(chan []byte, 256),
        subscribed: make(map[string]bool),
        lastSent:   make(map[string]*models.LiveLocation),
        lastPong:   time.Now(),
    }

    h.registerClient(orgID, client)

    // Enhanced ack with heartbeat info
    ack, _ := json.Marshal(models.WSMessage{
        Type: models.WSMsgConnectionAck,
        Data: map[string]interface{}{
            "status":                     "connected",
            "heartbeat_interval_ms":      30000,
            "position_update_interval_ms": 5000,
        },
    })
    client.send <- ack

    go client.writePump()
    go h.readPump(client)
    go h.subscribeToOrgUpdates(ctx, client)
    go h.startHeartbeat(client)      // NEW
    go h.startPositionTicker(ctx, client) // NEW
}
```

**New methods on `WebSocketHandler`:**
```go
func (h *WebSocketHandler) startHeartbeat(client *WSClient) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            if time.Since(client.lastPong) > 90*time.Second {
                client.conn.WriteControl(websocket.CloseMessage,
                    websocket.FormatCloseMessage(websocket.CloseNormalClosure, "heartbeat timeout"),
                    time.Now().Add(5*time.Second))
                client.conn.Close()
                return
            }
            client.conn.WriteMessage(websocket.PingMessage, nil)
        case <-client.conn.CloseChan():
            return
        }
    }
}

func (h *WebSocketHandler) startPositionTicker(ctx context.Context, client *WSClient) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            orgUUID, _ := uuid.FromString(client.orgID)
            vehicles, err := h.vehicleRepo.GetOrgVehicles(ctx, orgUUID)
            if err != nil {
                continue
            }

            vehicleIDs := make([]uuid.UUID, len(vehicles))
            for i, v := range vehicles {
                vehicleIDs[i] = v.ID
            }

            locations, err := h.liveCache.GetBatchLiveLocations(ctx, vehicleIDs)
            if err != nil {
                continue
            }

            for _, v := range vehicles {
                loc, ok := locations[v.ID]
                if !ok {
                    continue
                }

                // Differential push
                if last, exists := client.lastSent[v.ID.String()]; exists {
                    if !hasSignificantChange(*last, *loc) {
                        continue
                    }
                }

                status := models.CalculateStatus(loc.Speed, loc.Ignition, loc.LastSeen)
                msg := models.WSMessage{
                    Type: models.WSMsgPositionUpdate,
                    Data: models.PositionUpdate{
                        VehicleID: v.ID.String(),
                        Latitude:  loc.Lat,
                        Longitude: loc.Lng,
                        SpeedKmh:  loc.Speed,
                        Heading:   loc.Heading,
                        Status:    status,
                        Timestamp: loc.LastSeen.Format(time.RFC3339),
                    },
                }
                data, _ := json.Marshal(msg)
                select {
                case client.send <- data:
                    client.lastSent[v.ID.String()] = loc
                default:
                    // Buffer full — skip
                }
            }
        case <-ctx.Done():
            return
        }
    }
}

func hasSignificantChange(old, new models.LiveLocation) bool {
    latDelta := new.Lat - old.Lat
    if latDelta < 0 { latDelta = -latDelta }
    lngDelta := new.Lng - old.Lng
    if lngDelta < 0 { lngDelta = -lngDelta }
    speedDelta := new.Speed - old.Speed
    if speedDelta < 0 { speedDelta = -speedDelta }
    return latDelta > 0.0001 || lngDelta > 0.0001 || speedDelta > 2.0
}
```

**Enhanced `subscribeToOrgUpdates`** to also subscribe to geofence alerts:

```go
func (h *WebSocketHandler) subscribeToOrgUpdates(ctx context.Context, client *WSClient) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
    defer cancel()

    // Subscribe to telemetry channel
    telPubsub := h.redis.Subscribe(ctx, database.RedisChannelForOrg(client.orgID))
    defer telPubsub.Close()
    telCh := telPubsub.Channel()

    // Subscribe to geofence alerts channel
    geoPubsub := h.redis.Subscribe(ctx, fmt.Sprintf("org:%s:geofence_alerts", client.orgID))
    defer geoPubsub.Close()
    geoCh := geoPubsub.Channel()

    for {
        select {
        case msg, ok := <-telCh:
            if !ok { return }
            select {
            case client.send <- []byte(msg.Payload):
            default:
                h.logger.Warn("ws telemetry buffer full", zap.String("org_id", client.orgID))
            }
        case msg, ok := <-geoCh:
            if !ok { return }
            // Wrap geofence payload as WSMessage if not already
            select {
            case client.send <- []byte(msg.Payload):
            default:
                h.logger.Warn("ws geofence buffer full", zap.String("org_id", client.orgID))
            }
        case <-ctx.Done():
            return
        }
    }
}
```

**Enhanced `readPump`** to handle ping messages:

```go
func (h *WebSocketHandler) readPump(client *WSClient) {
    defer func() {
        h.unregisterClient(client.orgID, client)
        client.conn.Close()
    }()

    for {
        _, message, err := client.conn.ReadMessage()
        if err != nil {
            break
        }

        var wsMsg models.WSMessage
        if err := json.Unmarshal(message, &wsMsg); err != nil {
            continue
        }

        switch wsMsg.Type {
        case models.WSMsgSubscribe:
            var subReq models.SubscribeRequest
            data, _ := json.Marshal(wsMsg.Data)
            json.Unmarshal(data, &subReq)
            client.subscribed = make(map[string]bool)
            for _, vid := range subReq.VehicleIDs {
                client.subscribed[vid] = true
            }

        case "ping": // NEW
            client.lastPong = time.Now()
            pong, _ := json.Marshal(models.WSMessage{
                Type: "pong",
                Data: map[string]string{"timestamp": time.Now().UTC().Format(time.RFC3339)},
            })
            client.send <- pong
        }
    }
}
```

**WebSocketHandler needs new dependencies:**

```go
type WebSocketHandler struct {
    redis       *redis.Client
    logger      *zap.Logger
    jwtSecret   string
    upgrader    websocket.Upgrader
    vehicleRepo *repository.VehicleRepository    // NEW
    liveCache   *repository.LiveCacheRepository  // NEW

    mu      sync.RWMutex
    clients map[string][]*WSClient
}

func NewWebSocketHandler(
    rdb *redis.Client,
    jwtSecret string,
    vehicleRepo *repository.VehicleRepository,
    liveCache *repository.LiveCacheRepository,
    logger *zap.Logger,
) *WebSocketHandler {
    return &WebSocketHandler{
        redis: rdb, logger: logger, jwtSecret: jwtSecret,
        vehicleRepo: vehicleRepo, liveCache: liveCache,
        upgrader:  websocket.Upgrader{...},
        clients:   make(map[string][]*WSClient),
    }
}
```

#### `internal/tracking/handlers/history.go` — ENHANCE

**What:** Ensure query params are parsed and tenant isolation enforced.

```go
func (h *HistoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // ... existing vehicleID extraction (FIX: use chi.URLParam, not extractPathParam) ...

    orgID := middleware.GetOrgID(r.Context())  // FIXED: was GetOrgID(r) — takes context, not *http.Request
    // NEW: Validate vehicle belongs to org
    belongs, err := h.querySvc.VehicleBelongsToOrg(r.Context(), orgID, vehicleID)
    if err != nil || !belongs {
        http.Error(w, `{"error":"vehicle not found or not accessible"}`, http.StatusNotFound)
        return
    }

    start, _ := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
    end, _ := time.Parse(time.RFC3339, r.URL.Query().Get("end"))
    resolution := r.URL.Query().Get("resolution")

    result, err := h.querySvc.GetHistoryWithSummary(r.Context(), vehicleID, start, end, resolution)
    // ...
}
```

**Add to `TelemetryQueryService`:**
```go
func (s *TelemetryQueryService) VehicleBelongsToOrg(ctx context.Context, orgID, vehicleID uuid.UUID) (bool, error) {
    var exists bool
    err := s.db.QueryRow(ctx,
        `SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL)`,
        vehicleID, orgID).Scan(&exists)
    return exists, err
}
```

### 6.2 Tracking Service

#### `internal/tracking/service/batch.go` — ENHANCE

**What:** Add filtered location method.

```go
func (s *BatchLocationService) GetFilteredOrgLocations(
    ctx context.Context,
    orgID uuid.UUID,
    statusFilter string,
    driverID uuid.UUID,
    search string,
) ([]models.VehicleLocationResponse, error) {
    // Get all locations (reuse existing logic)
    all, err := s.GetAllOrgLocations(ctx, orgID)
    if err != nil {
        return nil, err
    }

    // Apply filters in-memory (small dataset, <500 items)
    filtered := make([]models.VehicleLocationResponse, 0, len(all))
    for _, loc := range all {
        if statusFilter != "" && loc.Status != statusFilter {
            continue
        }
        if driverID != uuid.Nil {
            if loc.Driver == nil || loc.Driver.ID != driverID.String() {
                continue
            }
        }
        if search != "" {
            searchLower := strings.ToLower(search)
            matchReg := strings.Contains(strings.ToLower(loc.Registration), searchLower)
            matchDriver := loc.Driver != nil && strings.Contains(strings.ToLower(loc.Driver.Name), searchLower)
            if !matchReg && !matchDriver {
                continue
            }
        }
        filtered = append(filtered, loc)
    }

    return filtered, nil
}
```

#### `internal/tracking/service/query.go` — ADD tenant isolation method

**What:** Add `VehicleBelongsToOrg` method (shown above in handler section).

Also add `db *pgxpool.Pool` field to the service struct:

```go
type TelemetryQueryService struct {
    telemRepo *repository.TelemetryRepository
    liveCache *repository.LiveCacheRepository
    db        *pgxpool.Pool  // NEW — for tenant isolation checks
    logger    *zap.Logger
}
```

### 6.3 Tracking Models

#### `internal/tracking/models/websocket.go` — ADD constants

```go
const (
    WSMsgPositionUpdate = "position_update"
    WSMsgGeofenceEvent  = "geofence_event"
    WSMsgConnectionAck  = "connection_ack"
    WSMsgSubscribe      = "subscribe_vehicles"
    WSMsgPing           = "ping"  // NEW
    WSMsgPong           = "pong"  // NEW
)
```

### 6.4 Geofencing Engine

#### `internal/domains/geofencing/engine/geofence_engine.go` — ENHANCE

**What:** Add Redis pub/sub for geofence alerts in `triggerGeofenceEvent()`.

The engine already has Redis via `NewGeofenceEngineWithRedis`. Add:

```go
func (ge *GeofenceEngine) triggerGeofenceEvent(ctx context.Context, vehicleID, geofenceID uuid.UUID, geofenceName, eventType string, event GPSEvent) {
    // ... existing DB insert ...
    // ... existing NATS publish ...

    // NEW: Publish to Redis for WebSocket fan-out
    if ge.redis != nil {
        orgID := ge.getOrgIDForVehicle(ctx, vehicleID)
        if orgID != uuid.Nil {
            geoMsg := map[string]interface{}{
                "type":            "geofence_alert",
                "vehicle_id":      vehicleID.String(),
                "geofence_id":     geofenceID.String(),
                "geofence_name":   geofenceName,
                "event_type":      eventType,
                "latitude":        event.Latitude,
                "longitude":       event.Longitude,
                "timestamp":       time.Unix(event.Timestamp, 0).UTC().Format(time.RFC3339),
            }
            data, _ := json.Marshal(geoMsg)
            channel := fmt.Sprintf("org:%s:geofence_alerts", orgID.String())
            if err := ge.redis.Publish(ctx, channel, string(data)).Err(); err != nil {
                ge.logger.Warn("failed to publish geofence to Redis", zap.Error(err))
            }
        }
    }
}

func (ge *GeofenceEngine) getOrgIDForVehicle(ctx context.Context, vehicleID uuid.UUID) uuid.UUID {
    var orgID uuid.UUID
    _ = ge.db.QueryRow(ctx,
        `SELECT tenant_id FROM vehicles WHERE id = $1 AND deleted_at IS NULL`,
        vehicleID).Scan(&orgID)
    return orgID
}
```

### 6.5 Route Registration (`cmd/main.go`)

**Current route registration** already handles all endpoints:

```go
// POST /api/v1/telemetry — GPS position ingest
api.Post("/v1/telemetry", ingestHandler.ServeHTTP)

// GET /api/v1/vehicles/locations — bulk org vehicle positions (enhanced with filters)
api.Get("/v1/vehicles/locations", locationsHandler.ServeHTTP)

// GET /api/v1/vehicles/{id}/telemetry/history — 24h history (enhanced with summary)
if cfg.Flags.HistoryTrail {
    api.Get("/v1/vehicles/{id}/telemetry/history", historyHandler.ServeHTTP)
}

// WebSocket /ws/telemetry — real-time position push (enhanced with heartbeat)
if cfg.Flags.RealTimeUpdates {
    api.Get("/ws/telemetry", wsHandler.ServeHTTP)
}

// GET/POST /api/v1/geofences — geofence management
api.Get("/v1/geofences", geofenceHandler.ServeHTTP)
api.Post("/v1/geofences", geofenceHandler.ServeHTTP)
```

**Changes to `cmd/main.go`:** Wire up new dependencies for `WebSocketHandler` and fix middleware:

```go
// CURRENT (as of codebase audit):
wsHandler := handlers.NewWebSocketHandler(rdb, cfg.JWTSecret, logger)
// ⚠️ Also: r.Use(middleware.TenantContextMiddleware) — doesn't compile
// Fix:    r.Use(middleware.TenantMiddleware(logger))

// AFTER (for position ticker + differential push):
wsHandler := handlers.NewWebSocketHandler(rdb, cfg.JWTSecret, vehicleRepo, liveCache, logger)
```

> ⚠️ **Note:** The `WebSocketHandler` struct needs two new fields: `vehicleRepo *repository.VehicleRepository` and `liveCache *repository.LiveCacheRepository`. The constructor signature will change. All callers must be updated.

### 6.6 Feature Flag Status

**No new flags needed.** The existing `FeatureFlags` struct already covers all requirements:

```go
type FeatureFlags struct {
    VehicleMapEnabled  bool `mapstructure:"vehicle_map_enabled"`   // Master kill switch
    RealTimeUpdates    bool `mapstructure:"real_time_updates"`     // WS toggle
    GeofenceAlerts     bool `mapstructure:"geofence_alerts"`       // Geofence worker toggle
    HistoryTrail       bool `mapstructure:"history_trail"`         // History endpoint toggle
    MaxVehiclesOnMap   int  `mapstructure:"max_vehicles_on_map"`   // Performance guardrail
}
```

> ⚠️ **Do NOT add `geofence_ws_push`** — it was proposed in an earlier draft but doesn't exist. Geofence WebSocket delivery is enabled when: (a) the GeofenceEngine has Redis configured, and (b) the WebSocketHandler subscribes to `org:{org}:geofence_alerts`.

Default values are set in `config.Load()`. Runtime changes require a service restart.

---

## 7. Frontend Specification

### 7.1 Component Hierarchy

**Location:** `apps/dashboard-web/src/features/tracking/`

```
apps/dashboard-web/
└── src/
    └── features/
        └── tracking/
            ├── components/
            │   ├── VehicleMap.tsx              # Main Leaflet map component
            │   ├── VehicleMarker.tsx           # Individual colored vehicle marker
            │   ├── SearchBar.tsx               # Vehicle search + filter bar
            │   ├── VehicleDetailPanel.tsx      # Slide-over detail panel
            │   └── MapStatusBar.tsx            # Bottom summary (total, online, moving, etc.)
            ├── hooks/
            │   ├── useVehiclePositions.ts      # WS subscription + HTTP polling fallback
            │   └── useVehicleSearch.ts         # Debounced search
            ├── store/
            │   └── trackingSlice.ts            # Redux state + reducers
            ├── services/
            │   ├── trackingApi.ts              # REST API calls (RTK Query)
            │   └── websocketService.ts         # WS connection manager
            ├── types/
            │   └── tracking.ts                 # TypeScript interfaces
            └── utils/
                ├── statusColor.ts              # Status → Leaflet marker color
                └── sampling.ts                 # History resolution display
```

### 7.2 Redux State Shape

```typescript
interface TrackingState {
    vehicles: Record<string, VehicleLocation>;   // vehicle_id → location
    selectedVehicleId: string | null;
    searchQuery: string;
    statusFilter: VehicleStatus[];               // moving | idle | stopped | offline
    driverFilter: string | null;

    ws: {
        connected: boolean;
        lastMessage: number | null;              // timestamp
        reconnectAttempts: number;
        fallbackToPolling: boolean;
    };

    loading: boolean;
    error: string | null;
    lastUpdated: number | null;                  // timestamp
}

interface VehicleLocation {
    vehicle_id: string;
    registration: string;
    make: string;
    model: string;
    latitude: number;
    longitude: number;
    speed_kmh: number;
    heading: number;
    ignition: boolean;
    fuel_level: number;
    odometer: number;
    status: 'moving' | 'idle' | 'stopped' | 'offline';
    driver?: { id: string; name: string; phone: string };
    last_update: string;
    last_seen_seconds_ago: number;
}
```

### 7.3 WebSocket Connection Lifecycle

```typescript
class TelemetryWebSocket {
    private ws: WebSocket | null = null;
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
    private positionTimer: ReturnType<typeof setInterval> | null = null;

    connect(token: string, orgId: string) {
        const url = `wss://${API_HOST}/ws/telemetry?token=${token}&org_id=${orgId}`;
        this.ws = new WebSocket(url);

        this.ws.onopen = () => {
            this.reconnectAttempts = 0;
            store.dispatch(wsConnected(true));
            this.startHeartbeat();
        };

        this.ws.onmessage = (event) => {
            const msg: WSMessage = JSON.parse(event.data);
            switch (msg.type) {
                case 'position_update':
                    store.dispatch(updateVehiclePosition(msg.data));
                    break;
                case 'geofence_alert':
                    store.dispatch(addGeofenceAlert(msg.data));
                    showGeofenceNotification(msg.data);
                    break;
                case 'connection_ack':
                    store.dispatch(setWsConfig(msg.data));
                    break;
                case 'pong':
                    // Heartbeat acknowledged
                    break;
            }
        };

        this.ws.onclose = () => {
            this.stopHeartbeat();
            store.dispatch(wsConnected(false));
            this.scheduleReconnect(token, orgId);
        };

        this.ws.onerror = () => {
            this.stopHeartbeat();
            if (this.reconnectAttempts >= this.maxReconnectAttempts) {
                this.startPolling(token, orgId);  // Fallback to HTTP
            }
        };
    }

    private startHeartbeat() {
        // Client-side: send ping every 25s (server pings every 30s)
        this.heartbeatTimer = setInterval(() => {
            if (this.ws?.readyState === WebSocket.OPEN) {
                this.ws.send(JSON.stringify({ type: 'ping' }));
            }
        }, 25000);
    }

    private scheduleReconnect(token: string, orgId: string) {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            this.startPolling(token, orgId);
            return;
        }
        const delay = 1000 * Math.pow(2, this.reconnectAttempts);
        setTimeout(() => {
            this.reconnectAttempts++;
            this.connect(token, orgId);
        }, delay);
    }

    // HTTP polling fallback
    private startPolling(token: string, orgId: string) {
        store.dispatch(setFallbackToPolling(true));
        this.positionTimer = setInterval(async () => {
            const response = await fetch('/api/v1/vehicles/locations', {
                headers: { Authorization: `Bearer ${token}` },
            });
            const data = await response.json();
            store.dispatch(setAllVehicles(data.data));
        }, 10000); // Poll every 10s
    }
}
```

### 7.4 Marker Color Coding

```typescript
// utils/statusColor.ts
const STATUS_COLORS: Record<VehicleStatus, string> = {
    moving:  '#22c55e',  // Green
    idle:    '#f59e0b',  // Amber
    stopped: '#6b7280',  // Gray
    offline: '#ef4444',  // Red
};

function getStatusColor(status: VehicleStatus): string {
    return STATUS_COLORS[status] ?? STATUS_COLORS.offline;
}
```

Markers use Leaflet's `DivIcon` with CSS custom properties for colors. Moving vehicles show a pulsing green dot; offline vehicles show a faded red dot.

### 7.5 Search & Filter Bar

**Behavior:**
- Debounced search (300ms) by registration number or driver name
- Filter chips for status (moving, idle, stopped, offline) — multi-select
- Filter dropdown for driver selection
- Filters are applied client-side on the Redux store (server-side filters available for large orgs)
- Real-time count updates in status bar

---

## 8. Non-Functional Requirements

| Metric | Target | Notes |
|--------|--------|-------|
| Map load time | **<2s P95** | GET /vehicles/locations for 500 vehicles |
| Position update latency | **<5s P95** | GPS ingest → dashboard marker update |
| Bulk locations API | **<500ms P95** | Server-side response time |
| WebSocket delivery | **<100ms P95** | Redis pub/sub → WS client |
| History query (24h) | **<3s P95** | With auto-resolution |
| Max vehicles per org | **500** | Hard limit via `max_vehicles_on_map` flag |
| WS connections per instance | **1,000** | Goroutine-per-connection model |
| Redis memory usage | **<200MB** | 500 vehicles × 5min TTL |
| DB connection pool | **40 max / 5 min** | Shared pool (Critical tier) |
| Geofence detection lag | **<10s** | NATS → PostGIS check → publish |

---

## 9. Feature Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `vehicle_map_enabled` | bool | `true` | Master kill switch for all tracking endpoints |
| `real_time_updates` | bool | `true` | WebSocket endpoint toggle |
| `geofence_alerts` | bool | `false` | GeofenceChecker worker toggle (DB polling mode) |
| `history_trail` | bool | `true` | History endpoint toggle |
| `max_vehicles_on_map` | int | `500` | Performance guardrail — truncates response |

> ⚠️ **`geofence_ws_push` does NOT exist** in `config.FeatureFlags`. The original spec proposed it, but it was not added. Geofence WebSocket delivery is controlled by whether the `geofence_alerts` flag enables the worker AND whether the GeofenceEngine has Redis configured.

All flags are configured via environment variables (`FEATURE_FLAGS_*`) or `config.toml`. Runtime changes require a service restart.

---

## 10. Testing Strategy

### 10.1 Unit Tests

| Test File | Coverage Target | Scenarios |
|-----------|----------------|-----------|
| `handlers/locations_test.go` | >90% | Empty org, 1 vehicle, 500 vehicles, status filter, driver filter, search filter, unauthorized |
| `handlers/websocket_test.go` | >85% | Auth fail, connect, subscribe, position update, geofence alert, ping/pong, disconnect, reconnect, differential push |
| `handlers/history_test.go` | >85% | Auto-resolution, custom resolution, empty range, invalid params, tenant isolation violation |
| `service/batch_test.go` | >90% | Redis all-hit, Redis all-miss, partial miss (DB fallback), driver enrichment, status filter, search filter |
| `service/query_test.go` | >85% | Resolution calculation (all ranges), Haversine distance accuracy, time classification, empty points |
| `repository/livecache_test.go` | >90% | Set/get single, batch pipeline, parse hash errors, cache miss (nil return) |
| `repository/vehicle_test.go` | >85% | GetOrgVehicles, GetDriversBatch, empty org, tenant isolation |

### 10.2 Integration Tests

| Test | Setup | Validation |
|------|-------|------------|
| **End-to-end ingest → WS** | Mock GPS POST → WS client listening | Position received within 5s |
| **Bulk map load accuracy** | Seed 50 vehicles with known positions | Response matches seed data, status correct |
| **Geofence alert delivery** | Vehicle enters polygon geofence | geofence_event logged, WS receives `geofence_alert` message |
| **History accuracy** | Known 10-point GPS track | Summary distance matches Haversine within 1% |
| **Multi-tenant isolation** | 2 orgs, 5 vehicles each | Each org sees only their vehicles |
| **WebSocket heartbeat** | Connect WS client, wait 90s without pong | Server closes connection |
| **Differential push** | Send same position twice | Second message not sent (no change) |

### 10.3 Load Tests (k6)

**Script:** `load-tests/vehicle-tracking.js`

| Test | Target | Threshold |
|------|--------|-----------|
| **Bulk API** — 50 concurrent GET /vehicles/locations (500 vehicles) | P95 <500ms | 99% pass |
| **WebSocket** — 1,000 concurrent connections, 5s position updates | P95 delivery <100ms | 99% pass |
| **Telemetry ingest** — 1,000 events/sec (simulated 500 vehicles @ 2Hz) | P95 <100ms | 99% pass |
| **History** — 50 concurrent GET /vehicles/{id}/history (24h) | P95 <3s | 99% pass |

```javascript
// k6 bulk API test
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 50,
    duration: '2m',
    thresholds: {
        'http_req_duration': ['p(95)<500'],
        'http_req_failed': ['rate<0.01'],
    },
};

export default function () {
    const res = http.get('http://localhost:8084/api/v1/vehicles/locations', {
        headers: { Authorization: `Bearer ${__ENV.JWT_TOKEN}` },
    });
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response < 500ms': (r) => r.timings.duration < 500,
        'has data array': (r) => JSON.parse(r.body).data.length > 0,
    });
    sleep(1);
}
```

### 10.4 Multi-Tenant Isolation Tests

| Test | Validation |
|------|-----------|
| Org A queries /vehicles/locations | No Org B vehicles in response |
| Org A queries /vehicles/{org-b-vehicle}/history | 404 or empty |
| Org A WS connection | Only receives Org A telemetry |
| Cross-org geofence | Vehicle in Org A cannot trigger Org B geofence |

---

## 11. Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **WebSocket scalability** — 1,000 concurrent connections × 5s ticker = heavy | High | Medium | Differential push reduces bandwidth. Position ticker uses batch Redis fetch. If needed, scale horizontally with sticky sessions. |
| **Redis eviction** — Memory pressure evicts live cache keys | High | Low | 5-min TTL + DB fallback. Monitor `maxmemory-policy` (set to `allkeys-lru`). Alert on eviction count. |
| **Geofence engine performance** — PostGIS checks per GPS event for many geofences | Medium | Medium | PostGIS GIST indexes. Geofence assignment scoping (only check assigned geofences). Redis state caching. |
| **Marker clustering** — 500 markers at country zoom level | Low (frontend) | High | Leaflet.markercluster plugin. Client-side concern, not backend. |
| **GPS downtime** — Vehicles go offline, stale positions shown | Medium | Medium | 5-min threshold marks vehicles as "offline". Dashboard shows faded markers. Alert on fleet-wide offline count. |
| **Large orgs** — >500 vehicles, query slow | High | Low | `max_vehicles_on_map` guardrail truncates. Paginated loading if needed. Feature flag controlled. |
| **NATS outage** — Geofence engine loses GPS stream | High | Low | IngestionService writes to DB regardless. GeofenceEngine rebuilds state from DB on restart (`rebuildVehicleState`). |
| **JWT expiry in WS** — Long-lived WS connection with expired JWT | Medium | Medium | WS auth validates at connect time only. For production: add periodic re-auth or short JWT TTL with refresh endpoint. |

---

## 12. Open Questions

| # | Question | Recommendation | Owner |
|---|----------|---------------|-------|
| 1 | **Max geofences per org?** | Recommend 100 hard limit. Most orgs use 10–20. 100 × 500 vehicles × 2Hz = 100K checks/sec worst case. PostGIS handles this with GIST indexes. | Product |
| 2 | **Multi-vehicle bulk actions?** (e.g., send command to all selected vehicles) | Out of scope for P0. Phase 2 feature. Would require device command queue + batch dispatch. | Product |
| 3 | **Custom map provider support?** | P0 uses Leaflet with OpenStreetMap tiles. Support for Mapbox, Google Maps via provider abstraction in Phase 2. Not a backend concern. | Frontend |
| 4 | **Telemetry retention period?** | Current policy: 90-day raw data (TimescaleDB `drop_chunks` policy). Aggregated summaries retained indefinitely in separate analytics tables. | Platform |
| 5 | **Manual geofence trigger by dispatcher?** (e.g., "check if vehicle X is near location Y") | Out of scope for P0. Could be implemented as ad-hoc `POST /api/v1/geofences/check` endpoint in Phase 2. | Product |
| 6 | **WebSocket horizontal scaling?** | For single-instance fleet-core, not needed. For multi-instance: use Redis pub/sub (already the transport) with each instance subscribing. No sticky sessions needed. | Platform |
| 7 | **Driver multi-assignment?** (multiple drivers per vehicle over a shift) | Current data model: `current_driver_id` (single). Phase 2: trip-based driver assignment with history table. | Product |

---

## 13. Implementation Sequence

### Phase 1: Backend Enhancements (Week 1–2)

1. ✅ Read existing codebase (completed)
2. Enhance `WebSocketHandler` — heartbeat, position ticker, differential push, geofence subscription
3. Enhance `LocationsHandler` — query parameter filters, enhanced meta counts
4. Enhance `HistoryHandler` — wire up `GetHistoryWithSummary`, tenant isolation
5. Enhance `GeofenceEngine` — Redis pub/sub for geofence alerts
6. Add `VehicleBelongsToOrg` tenant isolation method
7. Write unit tests for all changes
8. Run integration tests

### Phase 2: Database Migration (Week 2)

9. Write migration `302_vehicle_tracking_enrichment.sql`
10. Run migration on staging
11. Verify index creation (CONCURRENTLY)
12. Validate `gps_locations` upsert function

### Phase 3: Load Testing (Week 3)

13. Write k6 load test scripts
14. Run bulk API load test (500 vehicles, 50 VUs)
15. Run WebSocket load test (1,000 connections)
16. Run history query load test (24h range, 50 VUs)
17. Optimize any bottlenecks

### Phase 4: Frontend Integration (Week 3–4)

18. Build React components (`VehicleMap`, `VehicleMarker`, `SearchBar`, `VehicleDetailPanel`)
19. Implement Redux state + WebSocket middleware
20. Connect to existing REST + WS endpoints
21. Implement marker color coding + clustering
22. Add heartbeat + reconnection + polling fallback
23. Manual QA on staging

### Phase 5: Release (Week 4)

24. Feature flag rollout (internal → beta customers → GA)
25. Monitor metrics (map load time, WS connections, error rates)
26. Document API for external consumers

---

## 14. Appendix

### A. Existing Endpoint Summary

| Endpoint | Method | Handler File | Service | Status |
|----------|--------|-------------|---------|--------|
| `/api/v1/telemetry` | POST | `handlers/ingest.go` | `IngestionService` | ✅ Live |
| `/api/v1/vehicles/locations` | GET | `handlers/locations.go` | `BatchLocationService` | ✅ Live (enhance) |
| `/api/v1/vehicles/{id}/location` | GET | `tracking_query_handler.go` (domains) | `TelemetryQueryService` | ✅ Live |
| `/api/v1/vehicles/{id}/history` | GET | `tracking_query_handler.go` (domains) | `TelemetryQueryService` | ✅ Live |
| `/api/v1/vehicles/{id}/replay` | GET | `tracking_query_handler.go` (domains) | `TelemetryQueryService` | ✅ Live |
| `/api/v1/vehicles/{id}/telemetry/history` | GET | `handlers/history.go` | `TelemetryQueryService` | ✅ Live (⚠️ pre-existing bug: `extractPathParam` undefined) |
| `/api/v1/vehicles/nearby` | GET | `tracking_query_handler.go` (domains) | `TelemetryQueryService` | ✅ Live |
| `/ws/telemetry` | WS | `handlers/websocket.go` | — | ✅ Live (enhance) |
| `/api/v1/geofences` | GET/POST | `handlers/geofence.go` | `GeofenceService` | ✅ Live |
| `/api/v1/devices/{id}` | Various | `handlers/device_handler.go` | `DeviceService` | ✅ Live |
| `/api/v1/devices/{id}/commands` | POST | `handlers/device_command_handler.go` | `DeviceCommandService` | ✅ Live |

### B. Codebase Dual-Tracking Note

The fleet-core codebase has **two parallel tracking implementations**:

| Path | Purpose | Used by |
|------|---------|---------|
| `internal/tracking/` | Canonical implementations wired in `cmd/main.go` | **Production** |
| `internal/domains/tracking/` | Domain-driven restructure (hexagonal) | Future migration target |

**This spec targets `internal/tracking/`** as it is the production-wired code. The `domains/tracking/` directory contains a parallel implementation not yet wired into the application entry point.

### C. Validation Report & Corrections

This spec was validated against the actual codebase by two independent audits. Here are all discrepancies found and their resolutions:

| # | Severity | Original Spec Claim | Actual Code | Fix Applied |
|---|----------|-------------------|-------------|-------------|
| 1 | 🔴 CRITICAL | `middleware.GetOrgID(r)` — passes `*http.Request` | `GetOrgID(ctx context.Context)` — takes context | All code snippets corrected to `GetOrgID(r.Context())` |
| 2 | 🔴 CRITICAL | `extractPathParam(r, "id")` in history.go | Function does not exist | Flagged as pre-existing bug; fix: `chi.URLParam(r, "id")` |
| 3 | 🔴 CRITICAL | `middleware.TenantContextMiddleware` in cmd/main.go | Function is `TenantMiddleware(logger)` | All references corrected |
| 4 | 🟡 MODERATE | Redis key: `vehicle:{id}:live` | Actual: `live:loc:{vehicleID}` (via `database.LiveLocationKey()`) | All references corrected |
| 5 | 🟡 MODERATE | IngestionService publishes to NATS `fleet.telemetry.gps` | Publishes to `telemetry.raw`; `fleet.telemetry.gps` has no publisher | Data flow section 4.1 corrected |
| 6 | 🟡 MODERATE | `fleet.alerts.dispatch` NATS topic | Does not exist | Removed all references; NATS table flagged as phantom |
| 7 | 🟡 MODERATE | AlertService subscribes to `fleet.geofence.events` | No subscriber found in codebase | Flagged as gap in section 4.3 |
| 8 | 🟡 MODERATE | `telemetry_points` table created by infra migrations | Created in `fleet-core/migrations/072_tracking_enterprise.sql` | Clarified in section 4.1 |
| 9 | 🟡 MODERATE | `geofence_events`, `geofence_assignments`, `telemetry_outbox` tables in migrations | Not found in any migration file — only in Go code and docs | Flagged in section 5.1 |
| 10 | 🟢 MINOR | `GetBatchLiveLocations` returns `map[uuid.UUID]models.LiveLocation` | Returns `map[uuid.UUID]*models.LiveLocation` (pointers) | Flagged in section 5.4 |
| 11 | 🟢 MINOR | `geofence_ws_push` feature flag | Does not exist in `config.FeatureFlags` | Removed from section 9; added warning |
| 12 | ℹ️ INFO | WebSocket has heartbeat, position ticker, differential push, geofence alerts | None of these are implemented — all are NEW features to add | Section 2.3 clearly marks each as **ENHANCE** with code to add |

### D. Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/go-chi/chi/v5` | — | HTTP router |
| `github.com/gofrs/uuid/v5` | v5 | UUID types |
| `github.com/gorilla/websocket` | — | WebSocket handling |
| `github.com/redis/go-redis/v9` | v9 | Redis client |
| `github.com/jackc/pgx/v5` | v5 | PostgreSQL driver |
| `go.uber.org/zap` | — | Structured logging |
| `github.com/nats-io/nats.go` | — | NATS messaging |
| `github.com/prometheus/client_golang` | — | Metrics |
| `go.opentelemetry.io/otel` | — | Distributed tracing |
