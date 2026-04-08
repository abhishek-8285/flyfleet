# Vehicle Tracking — Architecture Design

**Document ID:** ARCH-VT-001
**Version:** 1.0
**Status:** Draft
**Created:** April 7, 2026
**Parent Spec:** [Vehicle Tracking Product Spec](../product-spec/VEHICLE_TRACKING.md)
**Architecture Review Date:** April 7, 2026

---

## 1. Architecture Overview

### 1.1 Design Goals

1. **Real-time Performance**: Sub-5-second position update latency from GPS ingest to UI display
2. **Scalability**: Support 500+ vehicles per organization with <2s map load time (P95)
3. **Multi-tenant Isolation**: Strict Row-Level Security (RLS) ensuring org-scoped data access
4. **Resilience**: Graceful degradation — WebSocket fallback to HTTP polling if connection fails
5. **Cost Efficiency**: Open-source map provider (MapLibre), no per-request API costs

### 1.2 Architectural Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Map Provider | **MapLibre GL** | Open-source, vector tiles, no API costs, active community |
| Telemetry Ingest | **REST API** | Aligns with current fleet-core architecture, simpler than MQTT |
| Real-time Transport | **WebSocket** | Bidirectional, low-latency, native browser support |
| Time-series Storage | **TimescaleDB** | PostgreSQL native extension, automatic partitioning, efficient time-range queries |
| Geofence Detection | **PostGIS + Background Job** | Leverages existing spatial indexes, deterministic detection |
| State Management | **Redux + WebSocket Store** | Predictable state updates, time-travel debugging |
| Marker Rendering | **Canvas-based (not DOM)** | Performance for 500+ markers, smooth animations |

---

## 2. System Context Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           External Systems                                   │
│                                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────────────────┐   │
│  │ GPS Devices  │    │ Mobile App   │    │ Third-party Integrations     │   │
│  │ (OBD-II)     │    │ (Phase 2)    │    │ (FASTag, VAHAN)              │   │
│  └──────┬───────┘    └──────┬───────┘    └──────────────────────────────┘   │
│         │                   │                                               │
└─────────┼───────────────────┼───────────────────────────────────────────────┘
          │ HTTP POST         │ (future)
          ▼                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        FlyFleet Platform Boundary                            │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                     dashboard-web (React SPA)                        │    │
│  │                                                                      │    │
│  │  ┌──────────────┐  ┌────────────────┐  ┌─────────────────────────┐ │    │
│  │  │ Map Component│  │ Vehicle Panel  │  │ WebSocket Client        │ │    │
│  │  │ (MapLibre)   │◄─┤ (Slide-over)   │  │ (Position Updates)      │ │    │
│  │  └──────┬───────┘  └───────▲────────┘  └──────────┬──────────────┘ │    │
│  │         │                  │                       │                │    │
│  └─────────┼──────────────────┼───────────────────────┼────────────────┘    │
│            │ HTTP REST        │ HTTP REST             │ WebSocket           │
│            ▼                  ▼                       ▼                     │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                  web-bff (Backend-for-Frontend)                      │    │
│  │  Port: 8081                                                          │    │
│  │                                                                      │    │
│  │  Routes:                                                             │    │
│  │    GET  /api/v1/vehicles/locations        ──► fleet-core-service     │    │
│  │    GET  /api/v1/vehicles/{id}             ──► vehicle-service        │    │
│  │    GET  /api/v1/vehicles/{id}/telemetry/history ──► tracking-service │    │
│  │    GET  /api/v1/geofences                 ──► fleet-core-service     │    │
│  │    WS   /ws/telemetry                     ──► tracking-service       │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│            │                  │                       │                     │
│            ▼                  ▼                       ▼                     │
│  ┌──────────────────┐ ┌──────────────────┐ ┌──────────────────────┐        │
│  │ fleet-core-svc   │ │ vehicle-service  │ │ tracking-service     │        │
│  │ (Port 8084)      │ │ (Port 8093)      │ │ (Port 8082)          │        │
│  │                  │ │                  │ │                      │        │
│  │ Telemetry Ingest │ │ Vehicle Details  │ │ WebSocket Handler    │        │
│  │ Geofence Detect  │ │ Location Query   │ │ Real-time Push       │        │
│  └────────┬─────────┘ └────────┬─────────┘ └──────────┬───────────┘        │
│           │                    │                       │                    │
└───────────┼────────────────────┼───────────────────────┼────────────────────┘
            │                    │                       │
            ▼                    ▼                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Data Layer                                         │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │              PostgreSQL 15 + TimescaleDB + PostGIS                   │    │
│  │                                                                      │    │
│  │  Tables:                                                             │    │
│  │    • telemetry_points    (TimescaleDB hypertable, monthly partitions) │    │
│  │    • vehicles            (RLS tenant-scoped)                         │    │
│  │    • drivers             (RLS tenant-scoped)                         │    │
│  │    • geofences           (PostGIS geometry column)                   │    │
│  │    • geofence_events     (audit trail)                               │    │
│  │                                                                      │    │
│  │  Indexes:                                                            │    │
│  │    • idx_telemetry_latest        (vehicle_id, time DESC)             │    │
│  │    • idx_geofences_spatial       (GIST on polygon)                   │    │
│  │    • idx_geofences_org           (org_id, status)                    │    │
│  │    • idx_geofence_events_vehicle (vehicle_id, occurred_at DESC)      │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ┌──────────────┐                                                           │
│  │    Redis     │ ◄── WebSocket pub/sub, session cache, rate limiting       │
│  └──────────────┘                                                           │
│                                                                              │
│  ┌──────────────┐                                                           │
│  │    NATS      │ ◄── Geofence alert events → alert-service (future)        │
│  └──────────────┘                                                           │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Service Architecture

### 3.1 tracking-service (Port 8082)

**Responsibility**: Real-time telemetry WebSocket handler, history queries, geofence detection

```
tracking-service/
├── cmd/
│   └── main.go                    # Entry point, service initialization
├── internal/
│   ├── config/
│   │   └── config.go              # Service configuration
│   ├── database/
│   │   ├── connection.go          # PostgreSQL + TimescaleDB connection
│   │   └── queries/
│   │       ├── telemetry.sql      # Telemetry point queries
│   │       └── geofence.sql       # Geofence spatial queries
│   ├── websocket/
│   │   ├── hub.go                 # WebSocket connection manager (fan-out)
│   │   ├── handler.go             # WS message handler (subscribe, unsubscribe)
│   │   └── client.go              # Individual client connection
│   ├── handler/
│   │   ├── telemetry.go           # HTTP handlers for telemetry queries
│   │   └── geofence.go            # HTTP handlers for geofence queries
│   ├── service/
│   │   ├── telemetry.go           # Telemetry business logic
│   │   ├── geofence.go            # Geofence detection logic
│   │   └── history.go             # History trail query logic
│   ├── model/
│   │   ├── telemetry.go           # Telemetry point structs
│   │   ├── geofence.go            # Geofence structs
│   │   └── vehicle.go             # Vehicle position structs
│   └── middleware/
│       ├── auth.go                # JWT validation
│       └── cors.go                # CORS headers
├── migrations/
│   └── 001_create_telemetry_indexes.sql
├── Dockerfile
├── go.mod
└── config.toml
```

**Key Interfaces**:

```go
// TelemetryService handles position data operations
type TelemetryService interface {
    // GetLatestPositions returns the most recent position for each vehicle in an org
    GetLatestPositions(ctx context.Context, orgID uuid.UUID) ([]VehiclePosition, error)
    
    // GetHistory returns time-sampled telemetry for a vehicle within a time range
    GetHistory(ctx context.Context, vehicleID uuid.UUID, start, end time.Time, resolution time.Duration) ([]TelemetryPoint, error)
    
    // Ingest writes a new telemetry point to the database
    Ingest(ctx context.Context, point TelemetryPoint) error
}

// GeofenceService handles geofence detection
type GeofenceService interface {
    // CheckGeofences checks if a position triggers any geofence entry/exit events
    CheckGeofences(ctx context.Context, vehicleID uuid.UUID, lat, lng float64) ([]GeofenceEvent, error)
    
    // GetGeofences returns all active geofences for an organization
    GetGeofences(ctx context.Context, orgID uuid.UUID) ([]Geofence, error)
}

// WebSocketHub manages client connections and message fan-out
type WebSocketHub interface {
    // Run starts the hub's main event loop
    Run()
    
    // Subscribe registers a client to receive updates for specific vehicles
    Subscribe(client *Client, vehicleIDs []uuid.UUID) error
    
    // Unsubscribe removes a client's subscription
    Unsubscribe(client *Client) error
    
    // BroadcastPosition sends a position update to all subscribed clients
    BroadcastPosition(vehicleID uuid.UUID, position VehiclePosition)
}
```

### 3.2 fleet-core-service (Port 8084)

**Responsibility**: Vehicle location aggregation endpoint, geofence management API

**Endpoints Added**:

```
GET /api/v1/vehicles/locations
  - Joins vehicles table with latest telemetry position
  - Returns array of VehicleLocation objects
  - RLS enforced: WHERE org_id = current_user_org_id()
  - Performance target: <2s P95 for 500 vehicles

GET /api/v1/geofences
  - Returns all active geofences for organization
  - Includes polygon/circle geometry

POST /api/v1/geofences
  - Creates new geofence (circle or polygon)
  - Validates geometry
  - Returns created geofence object
```

### 3.3 web-bff (Port 8081)

**Responsibility**: API aggregation, WebSocket proxy, session management

**Changes Required**:

1. **Route Addition**: Proxy `/api/v1/vehicles/locations` → fleet-core-service
2. **Route Addition**: Proxy `/api/v1/vehicles/{id}/telemetry/history` → tracking-service
3. **WebSocket Proxy**: Upgrade client WS connection → tracking-service WS handler
4. **Auth Middleware**: JWT validation + org_id extraction for all tracking endpoints

```go
// web-bff route configuration
var trackingRoutes = []Route{
    {
        Method:  "GET",
        Path:    "/api/v1/vehicles/locations",
        Target:  "http://fleet-core-service:8084",
        Auth:    true,
        Cache:   5 * time.Second, // Cache for 5s to reduce DB load
    },
    {
        Method:  "GET",
        Path:    "/api/v1/vehicles/{id}/telemetry/history",
        Target:  "http://tracking-service:8082",
        Auth:    true,
        Cache:   0, // No cache — user-specific queries
    },
    {
        Method:  "GET",
        Path:    "/api/v1/geofences",
        Target:  "http://fleet-core-service:8084",
        Auth:    true,
        Cache:   30 * time.Second, // Geofences change infrequently
    },
    {
        Method:  "WS",
        Path:    "/ws/telemetry",
        Target:  "http://tracking-service:8082",
        Auth:    true,
        Upgrade: true, // WebSocket upgrade
    },
}
```

---

## 4. Data Architecture

### 4.1 Database Schema (Enhanced)

#### telemetry_points (Existing — TimescaleDB Hypertable)

```sql
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

-- Monthly partitioning for performance
SELECT create_hypertable('telemetry_points', 'time', 
    chunk_time_interval => INTERVAL '1 month',
    if_not_exists => true
);

-- Compression policy: compress chunks older than 7 days
SELECT add_compression_policy('telemetry_points', 
    INTERVAL '7 days', if_not_exists => true);

-- Retention policy: drop raw data older than 90 days
SELECT add_drop_chunks_policy('telemetry_points', 
    INTERVAL '90 days', if_not_exists => true);
```

#### New: latest_vehicle_positions (Materialized View)

```sql
-- Optimized query for "where are all my vehicles right now?"
CREATE MATERIALIZED VIEW latest_vehicle_positions AS
SELECT DISTINCT ON (t.vehicle_id)
    t.vehicle_id,
    v.org_id,
    v.registration,
    v.make,
    v.model,
    v.year,
    v.status AS vehicle_status,
    v.current_driver_id,
    d.name AS driver_name,
    d.phone AS driver_phone,
    t.time AS last_update,
    t.latitude,
    t.longitude,
    t.speed_kmh,
    t.heading,
    t.ignition,
    t.fuel_level,
    t.altitude_m,
    -- Derive status from speed + ignition
    CASE
        WHEN t.speed_kmh > 5 THEN 'moving'
        WHEN t.speed_kmh > 0 AND t.speed_kmh <= 5 THEN 'idle'
        WHEN t.ignition = false THEN 'stopped'
        WHEN t.time < NOW() - INTERVAL '5 minutes' THEN 'offline'
        ELSE 'idle'
    END AS derived_status
FROM telemetry_points t
JOIN vehicles v ON t.vehicle_id = v.id
LEFT JOIN drivers d ON v.current_driver_id = d.id
ORDER BY t.vehicle_id, t.time DESC;

-- Index for org-scoped queries
CREATE INDEX idx_latest_positions_org ON latest_vehicle_positions (org_id);

-- Refresh every 5 seconds (application-level, not DB schedule)
-- REFRESH MATERIALIZED VIEW CONCURRENTLY latest_vehicle_positions;
```

**Rationale**: A materialized view avoids scanning the entire telemetry hypertable on every map load. `REFRESH CONCURRENTLY` allows reads during refresh without locking.

#### geofences (Existing — PostGIS)

```sql
CREATE TABLE geofences (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    type        VARCHAR(20) CHECK (type IN ('circle', 'polygon', 'corridor')),
    center_lat  DOUBLE PRECISION,
    center_lng  DOUBLE PRECISION,
    radius_m    DOUBLE PRECISION,
    polygon     GEOMETRY(POLYGON, 4326),
    corridor    GEOMETRY(LINESTRING, 4326),
    corridor_buffer_m DOUBLE PRECISION,
    status      VARCHAR(20) DEFAULT 'active',
    alert_channels JSONB DEFAULT '["in_app"]', -- ["in_app", "email", "sms", "webhook"]
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Spatial indexes
CREATE INDEX idx_geofences_polygon ON geofences USING GIST (polygon);
CREATE INDEX idx_geofences_org_status ON geofences (org_id, status);
```

#### geofence_events (Existing — Audit Trail)

```sql
CREATE TABLE geofence_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    geofence_id UUID REFERENCES geofences(id),
    vehicle_id  UUID REFERENCES vehicles(id),
    event_type  VARCHAR(20) CHECK (event_type IN ('entry', 'exit')),
    occurred_at TIMESTAMPTZ NOT NULL,
    latitude    DOUBLE PRECISION,
    longitude   DOUBLE PRECISION,
    alert_sent  BOOLEAN DEFAULT false,
    alert_channels JSONB,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_geofence_events_vehicle ON geofence_events (vehicle_id, occurred_at DESC);
CREATE INDEX idx_geofence_events_geofence ON geofence_events (geofence_id, occurred_at DESC);
CREATE INDEX idx_geofence_events_org ON geofence_events (geofence_id, occurred_at DESC);
```

### 4.2 Data Flow — Position Ingest

```
┌──────────────┐
│ GPS Device   │
│ (OBD-II)     │
└──────┬───────┘
       │ POST /api/v1/telemetry
       │ {vehicle_id, lat, lng, speed, heading, ignition, fuel}
       ▼
┌──────────────────────────────┐
│ fleet-core-service           │
│ 1. Validate JWT              │
│ 2. Validate GPS coordinates  │
│ 3. Check rate limit (1/min)  │
│ 4. Insert into telemetry_points
│ 5. Publish to Redis channel  │
│    "telemetry:{org_id}"      │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ Redis Pub/Sub                │
│ Channel: telemetry:{org_id}  │
│ Payload: {vehicle_id, lat,   │
│           lng, speed, time}  │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ tracking-service             │
│ 1. Subscribe to Redis channel│
│ 2. Check geofence triggers   │
│    (PostGIS ST_Contains)     │
│ 3. If geofence event:        │
│    a. Log to geofence_events │
│    b. Publish to NATS        │
│ 4. Broadcast via WebSocket   │
│    to subscribed clients     │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ alert-service (via NATS)     │
│ 1. Receive geofence event    │
│ 2. Look up user preferences  │
│ 3. Deliver via channels      │
│    (email, SMS, in-app)      │
└──────────────────────────────┘
```

### 4.3 Data Flow — History Query

```
┌──────────────┐
│ dashboard-web│
│ User clicks  │
│ "View History"│
└──────┬───────┘
       │ GET /api/v1/vehicles/{id}/telemetry/history
       │ ?start=...&end=...&resolution=60s
       ▼
┌──────────────────────────────┐
│ web-bff                      │
│ 1. Validate JWT              │
│ 2. Extract org_id            │
│ 3. Forward to tracking-svc   │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ tracking-service             │
│ 1. Validate vehicle belongs  │
│    to user's org (RLS)       │
│ 2. Calculate time_bucket     │
│    based on resolution       │
│ 3. Query TimescaleDB:        │
│                              │
│    SELECT time_bucket(       │
│      '60 seconds', time      │
│    ) AS bucket,              │
│    FIRST(latitude, time),    │
│    FIRST(longitude, time),   │
│    AVG(speed_kmh),           │
│    FIRST(heading, time),     │
│    FIRST(ignition, time),    │
│    FIRST(fuel_level, time)   │
│    FROM telemetry_points     │
│    WHERE vehicle_id = $1     │
│    AND time BETWEEN $2 AND $3│
│    GROUP BY bucket           │
│    ORDER BY bucket           │
│                              │
│ 4. Calculate summary stats   │
│ 5. Return JSON response      │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────┐
│ dashboard-web│
│ 1. Receive   │
│    points    │
│ 2. Draw      │
│    polyline  │
│ 3. Add       │
│    markers   │
└──────────────┘
```

---

## 5. Frontend Architecture

### 5.1 Component Hierarchy

```
dashboard-web/
└── src/
    └── features/
        └── tracking/
            ├── components/
            │   ├── VehicleMap.tsx              # Main map component (MapLibre wrapper)
            │   ├── VehicleMarker.tsx           # Individual vehicle marker
            │   ├── MarkerCluster.tsx           # Clustered marker group
            │   ├── HistoryTrail.tsx            # Polyline overlay for history
            │   ├── GeofenceOverlay.tsx         # Geofence boundary overlays
            │   ├── VehicleDetailPanel.tsx      # Slide-over detail panel
            │   ├── SearchBar.tsx               # Vehicle search input
            │   ├── FilterBar.tsx               # Status/type filter chips
            │   ├── PlaybackControls.tsx        # History playback UI
            │   ├── MapStatusBar.tsx            # Bottom status summary bar
            │   └── MapSettingsModal.tsx        # Map configuration modal
            ├── hooks/
            │   ├── useVehiclePositions.ts      # WebSocket subscription hook
            │   ├── useVehicleHistory.ts        # History query hook
            │   ├── useGeofences.ts             # Geofence data fetch hook
            │   ├── useVehicleSearch.ts         # Debounced search hook
            │   └── useMapBounds.ts             # Auto-fit map bounds hook
            ├── store/
            │   ├── trackingSlice.ts            # Redux slice for tracking state
            │   └── websocketMiddleware.ts      # Redux middleware for WS
            ├── services/
            │   ├── trackingApi.ts              # RTK Query endpoints
            │   └── websocketService.ts         # WebSocket connection manager
            ├── types/
            │   └── tracking.ts                 # TypeScript interfaces
            └── utils/
                ├── statusColor.ts              # Status → color mapping
                ├── geometry.ts                 # Geofence containment checks
                └── sampling.ts                 # History resolution logic
```

### 5.2 State Management

```typescript
// trackingSlice.ts
interface TrackingState {
    // Map state
    vehicles: Record<string, VehiclePosition>;    // vehicle_id → position
    geofences: Geofence[];
    selectedVehicleId: string | null;
    historyTrail: HistoryTrail | null;
    
    // Filter state
    searchQuery: string;
    statusFilters: VehicleStatus[];
    typeFilters: VehicleType[];
    
    // WebSocket state
    wsConnected: boolean;
    wsLastMessage: number | null;                 // timestamp
    wsSubscribedVehicles: string[];
    
    // UI state
    detailPanelOpen: boolean;
    historyPanelOpen: boolean;
    playbackSpeed: PlaybackSpeed;
    isPlaying: boolean;
    currentTime: number | null;                   // playback position
    
    // Loading state
    loading: {
        vehicles: boolean;
        history: boolean;
        geofences: boolean;
    };
    error: string | null;
}
```

### 5.3 WebSocket Connection Lifecycle

```typescript
// websocketService.ts
class TelemetryWebSocket {
    private ws: WebSocket | null = null;
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectDelay = 1000; // ms, doubles each attempt
    
    connect(token: string, orgId: string) {
        const url = `wss://api.flyfleet.io/ws/telemetry?token=${token}&org_id=${orgId}`;
        this.ws = new WebSocket(url);
        
        this.ws.onopen = () => {
            this.reconnectAttempts = 0;
            this.subscribeAllVehicles();
        };
        
        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };
        
        this.ws.onclose = () => {
            this.scheduleReconnect();
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }
    
    private handleMessage(message: WebSocketMessage) {
        switch (message.type) {
            case 'position_update':
                store.dispatch(updateVehiclePosition(message.data));
                break;
            case 'geofence_event':
                store.dispatch(addGeofenceEvent(message.data));
                break;
            case 'connection_ack':
                store.dispatch(setWsConnected(true));
                break;
        }
    }
    
    private scheduleReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            store.dispatch(setWsConnected(false));
            store.dispatch(setError('Connection lost. Refresh page.'));
            return;
        }
        
        const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
        setTimeout(() => {
            this.reconnectAttempts++;
            this.connect(this.token, this.orgId);
        }, delay);
    }
    
    // Fallback to HTTP polling if WebSocket fails permanently
    startPolling(token: string, orgId: string) {
        const poll = async () => {
            const response = await fetch('/api/v1/vehicles/locations', {
                headers: { Authorization: `Bearer ${token}`, 'X-Org-ID': orgId }
            });
            const data = await response.json();
            store.dispatch(setAllVehiclePositions(data.data));
            setTimeout(poll, 10000); // 10-second poll interval
        };
        poll();
    }
}
```

### 5.4 Performance Optimizations

| Optimization | Implementation | Impact |
|-------------|----------------|--------|
| **Marker Clustering** | Supercluster library, cluster within 30px radius | 500 markers → 50-100 clusters at zoom <10 |
| **Canvas Rendering** | MapLibre custom layer with canvas, not DOM markers | 60fps with 500 markers vs 15fps DOM |
| **Debounced Search** | 300ms debounce on search input | Reduced API calls during typing |
| **Memoized Selectors** | Reselect memoized selectors for filtered vehicles | Avoids recalculation on unrelated state changes |
| **Virtual Scrolling** | Virtual list for vehicle list in filter panel | Smooth scrolling with 500 items |
| **HTTP Caching** | 5s cache on GET /vehicles/locations in web-bff | Reduced DB load during page load spike |
| **Time-series Downsampling** | time_bucket() in TimescaleDB for history | 1,440 points → 1,440 (60s resolution), manageable for rendering |

---

## 6. Security Architecture

### 6.1 Authentication Flow

```
1. User logs in via auth-service → receives JWT
2. JWT contains: {sub, org_id, roles, exp}
3. All tracking API requests include JWT in Authorization header
4. WebSocket connection includes JWT as query parameter (initial handshake)
5. web-bff validates JWT signature + expiration on every request
6. tracking-service validates JWT on WebSocket upgrade
```

### 6.2 Authorization (RBAC + RLS)

| Role | Permissions |
|------|-------------|
| `super_admin` | View all orgs' vehicles (cross-org oversight) |
| `org_admin` | View all org vehicles, manage geofences |
| `dispatcher` | View all org vehicles, view history |
| `driver` | View only assigned vehicle's position |

**Row-Level Security Policies**:

```sql
-- Vehicles: users see only their org's vehicles
CREATE POLICY vehicle_org_isolation ON vehicles
    USING (org_id = current_setting('app.current_org_id')::uuid);

-- Telemetry: scoped through vehicle JOIN
-- (telemetry_points has no org_id — access controlled via vehicle JOIN)
CREATE POLICY telemetry_org_isolation ON telemetry_points
    USING (vehicle_id IN (
        SELECT id FROM vehicles WHERE org_id = current_setting('app.current_org_id')::uuid
    ));

-- Geofences: org-scoped
CREATE POLICY geofence_org_isolation ON geofences
    USING (org_id = current_setting('app.current_org_id')::uuid);
```

### 6.3 Data Protection

| Aspect | Implementation |
|--------|---------------|
| In Transit | TLS 1.3 enforced (HTTPS + WSS) |
| At Rest | PostgreSQL TDE + encrypted tablespaces |
| JWT Storage | HttpOnly, Secure, SameSite=Strict cookies |
| Audit Logging | All location access logged to `audit_logs` table |
| Rate Limiting | Telemetry ingest: 1 request/minute/vehicle (prevent spoofing) |

---

## 7. Performance Architecture

### 7.1 Query Performance

| Query | Strategy | Expected Latency |
|-------|----------|-----------------|
| Latest positions (500 vehicles) | Materialized view refresh every 5s | <200ms (indexed org lookup) |
| 24h history (1,440 points) | time_bucket + FIRST aggregate | <500ms (TimescaleDB chunk pruning) |
| Geofence check (10 geofences) | PostGIS ST_Contains + spatial index | <50ms (GIST index scan) |
| Vehicle search (by registration) | B-tree index on registration | <10ms |

### 7.2 WebSocket Fan-out Performance

```
100 concurrent users × 500 vehicles = 50,000 potential updates/second

Optimization strategy:
1. Batch position updates every 5 seconds (not every GPS ping)
2. Only send updates for vehicles the user is subscribed to
3. Compress WebSocket messages (permessage-deflate)
4. Use Redis pub/sub for horizontal scaling (multiple WS instances)

With batching:
500 vehicles / 5s batch = 100 updates/second per instance
100 users × 100 updates = 10,000 messages/second (manageable)
```

### 7.3 Memory Budget

| Component | Memory Usage | Notes |
|-----------|-------------|-------|
| MapLibre GL (React) | ~80 MB | Vector tiles + canvas layer |
| Redux Store (500 vehicles) | ~2 MB | 500 × 4KB position objects |
| WebSocket buffers | ~5 MB | 100 connections × 50KB buffer |
| Browser total | ~120 MB | Within 2GB tab limit |

---

## 8. Resilience & Failure Modes

### 8.1 Failure Scenarios

| Scenario | Detection | Recovery |
|----------|-----------|----------|
| WebSocket disconnect | `onclose` event | Exponential backoff reconnect (max 5 attempts) |
| Telemetry ingest failure | HTTP 5xx rate >1% | Circuit breaker, fallback to cached positions |
| Database connection loss | Connection pool timeout | Retry with backoff, serve stale data from cache |
| TimescaleDB query timeout (>3s) | Query cancellation | Reduce resolution, retry with coarser bucket |
| MapLibre rendering freeze | RAF frame rate <30fps | Disable animations, reduce marker count |

### 8.2 Graceful Degradation Path

```
Normal Operation (WebSocket + live data)
    ↓ WebSocket disconnect
HTTP Polling (10s interval)
    ↓ Polling failures >3
Cached Positions (stale data, last known position)
    ↓ Cache stale >5 min
Error State ("Tracking unavailable, please refresh")
```

---

## 9. Observability

### 9.1 Metrics (Prometheus)

```go
// tracking-service metrics
var (
    telemetryIngestTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "telemetry_ingest_total"},
        []string{"status"}, // success, error, duplicate
    )
    
    telemetryIngestLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "telemetry_ingest_latency_ms",
            Buckets: []float64{10, 50, 100, 200, 500, 1000},
        },
        []string{"endpoint"},
    )
    
    wsConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{Name: "websocket_connections"},
        []string{"status"}, // active, reconnecting, failed
    )
    
    wsUpdateLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "websocket_update_latency_ms",
            Buckets: []float64{100, 500, 1000, 2000, 5000},
        },
        nil,
    )
    
    geofenceDetectionLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "geofence_detection_latency_ms",
            Buckets: []float64{10, 50, 100, 200, 500},
        },
        nil,
    )
    
    historyQueryLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "history_query_latency_ms",
            Buckets: []float64{100, 500, 1000, 2000, 3000, 5000},
        },
        []string{"time_range"},
    )
    
    mapLoadLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "map_load_latency_ms",
            Buckets: []float64{200, 500, 1000, 1500, 2000, 3000},
        },
        []string{"vehicle_count"},
    )
)
```

### 9.2 Distributed Tracing

```
Trace: Vehicle Position Update
├── Span 1: GPS Device → fleet-core-service (HTTP POST /telemetry)
│   ├── Span 2: JWT validation
│   ├── Span 3: DB insert (telemetry_points)
│   └── Span 4: Redis publish (telemetry:{org_id})
├── Span 5: tracking-service receives Redis message
│   ├── Span 6: Geofence check (PostGIS ST_Contains)
│   │   └── Span 7: Geofence event log (if triggered)
│   └── Span 8: WebSocket broadcast to clients
└── Span 9: dashboard-web receives WS message
    └── Span 10: Marker render/update (browser performance API)
```

### 9.3 Alerting Rules

| Alert | Condition | Severity | Channel |
|-------|-----------|----------|---------|
| Map load P95 >3s | `histogram_quantile(0.95, rate(map_load_latency_ms_bucket[5m])) > 3000` | Warning | Slack #eng-alerts |
| WS disconnect rate >10% | `rate(websocket_connections{status="failed"}[5m]) / rate(websocket_connections[5m]) > 0.1` | Critical | PagerDuty |
| Telemetry ingest error >1% | `rate(telemetry_ingest_total{status="error"}[5m]) / rate(telemetry_ingest_total[5m]) > 0.01` | Warning | Slack #eng-alerts |
| Geofence detection lag >30s | `geofence_detection_latency_ms > 30000` | Warning | Slack #eng-alerts |
| DB connection pool exhausted | `pg_stat_activity_count / max_connections > 0.9` | Critical | PagerDuty |

---

## 10. Migration Plan

### 10.1 Database Migrations

```sql
-- Migration 001: Add indexes for tracking queries
CREATE INDEX CONCURRENTLY idx_telemetry_latest 
    ON telemetry_points (vehicle_id, time DESC);

CREATE INDEX CONCURRENTLY idx_geofences_spatial 
    ON geofences USING GIST (polygon);

CREATE INDEX CONCURRENTLY idx_geofences_org 
    ON geofences (org_id, status);

CREATE INDEX CONCURRENTLY idx_geofence_events_vehicle 
    ON geofence_events (vehicle_id, occurred_at DESC);

-- Migration 002: Create materialized view for latest positions
CREATE MATERIALIZED VIEW IF NOT EXISTS latest_vehicle_positions AS
SELECT DISTINCT ON (t.vehicle_id)
    t.vehicle_id, v.org_id, v.registration, v.make, v.model,
    v.status AS vehicle_status, v.current_driver_id,
    d.name AS driver_name, d.phone AS driver_phone,
    t.time AS last_update, t.latitude, t.longitude,
    t.speed_kmh, t.heading, t.ignition, t.fuel_level,
    CASE
        WHEN t.speed_kmh > 5 THEN 'moving'
        WHEN t.speed_kmh > 0 THEN 'idle'
        WHEN t.ignition = false THEN 'stopped'
        WHEN t.time < NOW() - INTERVAL '5 minutes' THEN 'offline'
        ELSE 'idle'
    END AS derived_status
FROM telemetry_points t
JOIN vehicles v ON t.vehicle_id = v.id
LEFT JOIN drivers d ON v.current_driver_id = d.id
ORDER BY t.vehicle_id, t.time DESC;

CREATE INDEX CONCURRENTLY idx_latest_positions_org 
    ON latest_vehicle_positions (org_id);

-- Migration 003: Add geofence alert channels column
ALTER TABLE geofences 
    ADD COLUMN IF NOT EXISTS alert_channels JSONB DEFAULT '["in_app"]';
```

### 10.2 Feature Flag Rollout

```yaml
# config.toml
[feature_flags]
vehicle_map_enabled = true        # Kill switch for entire feature
real_time_updates = true          # Toggle WebSocket live updates
geofence_alerts = false           # Enable after geofence setup complete
history_trail = true              # Toggle history visualization
max_vehicles_on_map = 500         # Performance guardrail
```

### 10.3 Deployment Sequence

```
Phase 1 (Week 1): Database migrations + tracking-service deploy
  → Run migrations (online, no downtime)
  → Deploy tracking-service alongside existing services
  → Enable feature flag for internal users only

Phase 2 (Week 2-3): Frontend deployment
  → Deploy dashboard-web with map component
  → Enable for 3 beta orgs (feature flag per org_id)
  → Monitor performance metrics, gather feedback

Phase 3 (Week 4-5): General availability
  → Enable feature flag for all orgs
  → Monitor error rates, latency metrics
  → Iterate on performance optimizations
```

---

## 11. Future Extensions

### 11.1 Phase 2 Additions

| Feature | Architecture Change |
|---------|-------------------|
| Mobile App GPS reporting | MQTT ingest endpoint → NATS → telemetry_points |
| Multi-stop trip visualization | New trips table + route waypoints |
| Offline map tiles | Service worker + MapLibre offline tile cache |
| Driver behavior scoring | Telemetry analysis pipeline (acceleration, braking, cornering) |

### 11.2 Phase 3 Additions

| Feature | Architecture Change |
|---------|-------------------|
| AI route optimization | ML model service → Python/FastAPI → route recommendations |
| Predictive ETA | Traffic data ingestion → ETA calculation service |
| Traffic-aware routing | Third-party traffic API integration → route scoring |
| Real-time traffic overlay | MapLibre traffic layer + live traffic feed |

---

## 12. Decision Log

| # | Decision | Date | Rationale | Alternatives Considered |
|---|----------|------|-----------|----------------------|
| 1 | MapLibre GL as map provider | Apr 7, 2026 | Open-source, no API costs, vector tiles | Leaflet (raster, less performant), Google Maps (cost at scale) |
| 2 | REST for telemetry ingest | Apr 7, 2026 | Aligns with current architecture, simpler ops | MQTT (better for IoT, but adds infrastructure complexity) |
| 3 | Materialized view for latest positions | Apr 7, 2026 | Avoids full hypertable scan on every map load | DISTINCT ON query (slower for 500+ vehicles), Redis cache |
| 4 | Canvas-based marker rendering | Apr 7, 2026 | 60fps with 500 markers | DOM markers (15fps at 500 markers), SVG (similar performance) |
| 5 | PostGIS for geofence detection | Apr 7, 2026 | Leverages existing spatial indexes, deterministic | In-memory point-in-polygon (scales poorly, duplicates data) |
| 6 | WebSocket with HTTP polling fallback | Apr 7, 2026 | Best UX with graceful degradation | Server-Sent Events (one-way only, no client → server messages) |
| 7 | 500 vehicle hard limit with clustering | Apr 7, 2026 | Performance guardrail, prevents browser overload | Virtual scrolling on map (not practical), pagination (poor UX) |

---

## 13. Appendix

### 13.1 OpenAPI Specification Additions

See [`VEHICLE_TRACKING.md`](../product-spec/VEHICLE_TRACKING.md) §4.3 for complete API contracts.

### 13.2 Glossary

| Term | Definition |
|------|-----------|
| **Hypertable** | TimescaleDB abstraction that auto-partitions time-series data |
| **time_bucket** | TimescaleDB function for grouping time-series data into intervals |
| **Materialized View** | Pre-computed query result stored as a table, refreshed periodically |
| **PostGIS** | PostgreSQL extension for spatial/geographic data operations |
| **ST_Contains** | PostGIS function checking if geometry contains another geometry |
| **Fan-out** | Distributing a single message to multiple subscribers |
| **Supercluster** | Library for fast geospatial point clustering |
| **RAF** | requestAnimationFrame — browser API for smooth animations |

### 13.3 Related Documents

- [Vehicle Tracking Product Spec](../product-spec/VEHICLE_TRACKING.md)
- [FlyFleet Architecture Overview](../architecture/)
- [OpenAPI Specification](../openapi.yaml)
- [Database Schema Documentation](../docs/database/)

---

**Next Review:** April 21, 2026
**Approvers:** Engineering Lead, Architecture Review Board
**Stakeholders:** Engineering Team, Product Team, DevOps
