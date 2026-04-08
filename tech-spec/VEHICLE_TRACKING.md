# Tech Spec: Vehicle Tracking System

**Document ID:** TECH-VT-001  
**Version:** 1.0  
**Status:** Draft  
**Created:** April 7, 2026  
**Author:** Senior Backend Engineer  
**Based On:** [Vehicle Tracking Architecture](../architecture/VEHICLE_TRACKING.md)  
**Parent Spec:** [Vehicle Tracking Product Spec](../product-spec/VEHICLE_TRACKING.md)

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [System Requirements](#2-system-requirements)
3. [Architecture Decisions](#3-architecture-decisions)
4. [Backend Implementation Plan](#4-backend-implementation-plan)
5. [Frontend Implementation Plan](#5-frontend-implementation-plan)
6. [Database Design](#6-database-design)
7. [API Contracts](#7-api-contracts)
8. [WebSocket Protocol](#8-websocket-protocol)
9. [Security Implementation](#9-security-implementation)
10. [Performance Engineering](#10-performance-engineering)
11. [Testing Strategy](#11-testing-strategy)
12. [Observability & Monitoring](#12-observability--monitoring)
13. [Migration & Deployment Plan](#13-migration--deployment-plan)
14. [Risk Assessment](#14-risk-assessment)
15. [Implementation Timeline](#15-implementation-timeline)
16. [Appendix](#16-appendix)

---

## 1. Executive Summary

### 1.1 Objective

Build a real-time vehicle tracking system for the FlyFleet platform enabling fleet managers to monitor 500+ vehicles simultaneously with sub-5-second update latency, geofence detection, and historical trip replay — all powered by open-source technology with zero per-request API costs.

### 1.2 Scope

**In Scope:**
- Real-time GPS telemetry ingestion via REST API
- WebSocket-based live position updates to dashboard clients
- Map visualization using MapLibre GL with canvas-based rendering
- Marker clustering for 500+ vehicles
- Historical trip replay with configurable time resolution
- Geofence creation, detection, and event logging
- Row-level security for multi-tenant isolation
- HTTP polling fallback for WebSocket failures
- Performance optimizations for large fleets

**Out of Scope (Future Phases):**
- Mobile app GPS reporting (Phase 2)
- Driver behavior scoring (Phase 2)
- Offline map tiles (Phase 2)
- AI route optimization (Phase 3)
- Traffic-aware routing (Phase 3)

### 1.3 Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Map load time (P95) | <2s for 500 vehicles | `map_load_latency_ms` histogram |
| Position update latency | <5s (GPS → UI) | End-to-end trace correlation ID |
| WebSocket connection success rate | >99% | `websocket_connections` gauge |
| Geofence detection latency | <50ms (P95) | `geofence_detection_latency_ms` |
| History query time (24h) | <3s (P95) | `history_query_latency_ms` |
| Browser memory usage | <150MB | Chrome DevTools Memory tab |
| Telemetry ingest throughput | 1000 req/s | `telemetry_ingest_total` counter |

---

## 2. System Requirements

### 2.1 Functional Requirements

| ID | Requirement | Priority | Notes |
|----|-------------|----------|-------|
| FR-1 | System shall ingest GPS positions from vehicles via HTTP POST | P0 | 1 req/min/vehicle rate limit |
| FR-2 | System shall broadcast position updates to subscribed WebSocket clients | P0 | 5-second batch interval |
| FR-3 | System shall display all org vehicles on interactive map | P0 | MapLibre GL, canvas rendering |
| FR-4 | System shall cluster markers when zoom level < 10 | P0 | Supercluster library, 30px radius |
| FR-5 | System shall derive vehicle status from telemetry (moving/idle/stopped/offline) | P0 | Speed + ignition + recency logic |
| FR-6 | System shall support vehicle search by registration | P0 | Debounced 300ms |
| FR-7 | System shall filter vehicles by status and type | P0 | Multi-select filter chips |
| FR-8 | System shall display vehicle detail panel on marker click | P0 | Slide-over panel |
| FR-9 | System shall query historical telemetry for a vehicle | P0 | Time-range + resolution params |
| FR-10 | System shall render history trail as polyline overlay | P1 | Color-coded by speed |
| FR-11 | System shall support history playback with speed controls | P1 | 1x, 2x, 4x, 8x speeds |
| FR-12 | System shall create/manage geofences (circle/polygon) | P0 | PostGIS geometry storage |
| FR-13 | System shall detect geofence entry/exit events | P0 | Background job on ingest |
| FR-14 | System shall log geofence events to audit trail | P0 | `geofence_events` table |
| FR-15 | System shall enforce org-scoped data access (RLS) | P0 | Row-Level Security on all tables |
| FR-16 | System shall fallback to HTTP polling if WS fails | P1 | 10-second poll interval |
| FR-17 | System shall display connection status indicator | P1 | Green/yellow/red dot |

### 2.2 Non-Functional Requirements

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-1 | Position update latency (GPS → UI) | <5 seconds (P95) |
| NFR-2 | Map load time (500 vehicles) | <2 seconds (P95) |
| NFR-3 | History query (24h, 1-min resolution) | <3 seconds (P95) |
| NFR-4 | Geofence detection | <50ms per position (P95) |
| NFR-5 | WebSocket reconnect attempts | Max 5, exponential backoff |
| NFR-6 | Browser memory (500 vehicles) | <150 MB |
| NFR-7 | Telemetry ingest throughput | 1000 req/s per instance |
| NFR-8 | Database retention | 90 days raw, indefinite aggregated |
| NFR-9 | Multi-tenant isolation | 100% RLS coverage |
| NFR-10 | API availability | 99.9% uptime (monthly) |

---

## 3. Architecture Decisions

### 3.1 Technology Stack

| Component | Choice | Rationale | Alternatives Rejected |
|-----------|--------|-----------|----------------------|
| Map Rendering | **MapLibre GL JS** | Open-source, vector tiles, no API costs, active community | Google Maps ($7/1000 loads), Leaflet (raster-only) |
| Telemetry Ingest | **REST API (HTTP POST)** | Aligns with existing fleet-core architecture, simpler ops | MQTT (adds broker complexity), gRPC (browser support) |
| Real-time Transport | **WebSocket** | Bidirectional, low-latency, native browser support | SSE (one-way only), long-polling (higher latency) |
| Time-series Storage | **TimescaleDB** | PostgreSQL native, auto partitioning, efficient time queries | InfluxDB (separate infra), Cassandra (overkill) |
| Spatial Operations | **PostGIS** | Native to PostgreSQL, mature, comprehensive spatial functions | Custom in-memory (scales poorly) |
| State Management | **Redux Toolkit + RTK Query** | Predictable, time-travel debugging, ecosystem | Zustand (less mature), Context (perf issues) |
| Marker Clustering | **Supercluster** | Fast, geospatial, well-maintained | MarkerClusterer (slower), custom (reinventing wheel) |
| WebSocket Library (Go) | **gorilla/websocket** | Battle-tested, standard in Go ecosystem | nhooyr.io/websocket (newer, less proven) |

### 3.2 Key Architectural Patterns

#### Pattern 1: Materialized View for Latest Positions

**Problem:** Querying the telemetry hypertable for the latest position of 500 vehicles on every map load causes full table scans.

**Solution:** Maintain a materialized view `latest_vehicle_positions` refreshed every 5 seconds.

```sql
-- Refresh strategy (application-level control)
REFRESH MATERIALIZED VIEW CONCURRENTLY latest_vehicle_positions;
```

**Trade-off:** Up to 5-second staleness vs 10x query performance improvement.

#### Pattern 2: Redis Pub/Sub for Horizontal Scaling

**Problem:** Multiple tracking-service instances need to coordinate WebSocket broadcasts.

**Solution:** Use Redis pub/sub channels scoped by org_id.

```
Channel: telemetry:{org_id}
Payload: {"vehicle_id": "...", "lat": 12.34, "lng": 56.78, ...}
```

**Scaling Model:** Each tracking-service instance subscribes to channels for connected clients' orgs. Broadcasts fan out via Redis.

#### Pattern 3: Graceful Degradation Chain

```
WebSocket (live) → HTTP Polling (10s) → Cached Data (stale) → Error State
```

Each level triggers automatically on failure of the previous level.

---

## 4. Backend Implementation Plan

### 4.1 tracking-service (Port 8082)

#### 4.1.1 Service Structure

```
services/tracking-service/
├── cmd/
│   └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   ├── connection.go
│   │   └── migrations/
│   │       ├── 001_create_telemetry_indexes.up.sql
│   │       ├── 001_create_telemetry_indexes.down.sql
│   │       ├── 002_create_materialized_view.up.sql
│   │       └── 002_create_materialized_view.down.sql
│   ├── websocket/
│   │   ├── hub.go
│   │   ├── client.go
│   │   └── handler.go
│   ├── handler/
│   │   ├── telemetry.go
│   │   ├── geofence.go
│   │   └── health.go
│   ├── service/
│   │   ├── telemetry.go
│   │   ├── geofence.go
│   │   └── history.go
│   ├── repository/
│   │   ├── telemetry.go
│   │   └── geofence.go
│   ├── model/
│   │   ├── telemetry.go
│   │   ├── geofence.go
│   │   └── vehicle.go
│   └── middleware/
│       ├── auth.go
│       ├── cors.go
│       └── metrics.go
├── Dockerfile
├── go.mod
├── go.sum
└── config.toml
```

#### 4.1.2 Core Go Interfaces

```go
// service/telemetry.go

type TelemetryService struct {
    db      *sql.DB
    redis   *redis.Client
    hub     WebSocketHub
    logger  *zap.Logger
    metrics *Metrics
}

// Ingest writes a new telemetry point and triggers downstream processing
func (s *TelemetryService) Ingest(ctx context.Context, point TelemetryPoint) error {
    // 1. Validate coordinates
    // 2. Insert into telemetry_points
    // 3. Publish to Redis channel
    // 4. Trigger async geofence check
    // 5. Return
}

// GetLatestPositions returns current positions for all vehicles in an org
func (s *TelemetryService) GetLatestPositions(ctx context.Context, orgID uuid.UUID) ([]VehiclePosition, error) {
    // Query latest_vehicle_positions materialized view
    // WHERE org_id = $1
}

// GetHistory returns time-sampled telemetry for history visualization
func (s *TelemetryService) GetHistory(ctx context.Context, req HistoryRequest) ([]HistoryPoint, error) {
    // time_bucket query with FIRST/AVG aggregates
    // Dynamic resolution based on time range
}
```

```go
// websocket/hub.go

type WebSocketHub struct {
    clients    map[uuid.UUID]map[*Client]bool  // vehicle_id → clients
    register   chan *Client
    unregister chan *Client
    broadcast  chan BroadcastMessage
    logger     *zap.Logger
    metrics    *Metrics
}

// Run starts the hub's event loop
func (h *WebSocketHub) Run() {
    for {
        select {
        case client := <-h.register:
            h.addClient(client)
        case client := <-h.unregister:
            h.removeClient(client)
        case msg := <-h.broadcast:
            h.fanOut(msg)
        }
    }
}

// Subscribe registers a client for specific vehicle updates
func (h *WebSocketHub) Subscribe(client *Client, vehicleIDs []uuid.UUID) error
```

```go
// service/geofence.go

type GeofenceService struct {
    db     *sql.DB
    logger *zap.Logger
}

// CheckGeofences tests if a position triggers any geofence events
func (s *GeofenceService) CheckGeofences(ctx context.Context, vehicleID uuid.UUID, lat, lng float64) ([]GeofenceEvent, error) {
    // 1. Get all active geofences for vehicle's org
    // 2. For each geofence:
    //    - Check ST_Contains(polygon, point) or ST_DWithin(center, point, radius)
    //    - Compare with last known state (entry vs exit)
    //    - If state changed, log to geofence_events
    // 3. Return events
}
```

#### 4.1.3 Telemetry Ingest Endpoint

```go
// handler/telemetry.go

func (h *TelemetryHandler) IngestPosition(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body
    var req IngestRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
        return
    }

    // 2. Validate coordinates
    if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
        h.respondError(w, http.StatusBadRequest, "INVALID_COORDINATES", "Coordinates out of range")
        return
    }

    // 3. Check rate limit (1 req/min/vehicle)
    if !h.rateLimiter.Allow(req.VehicleID.String()) {
        h.respondError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Max 1 update per minute")
        return
    }

    // 4. Insert telemetry
    point := model.TelemetryPoint{
        VehicleID: req.VehicleID,
        Time:      time.Now().UTC(),
        Latitude:  req.Latitude,
        Longitude: req.Longitude,
        SpeedKmh:  req.SpeedKmh,
        Heading:   req.Heading,
        Ignition:  req.Ignition,
        FuelLevel: req.FuelLevel,
        RawData:   req.RawData,
    }

    if err := h.telemetryService.Ingest(r.Context(), point); err != nil {
        h.respondError(w, http.StatusInternalServerError, "INGEST_FAILED", err.Error())
        return
    }

    h.metrics.IngestTotal.WithLabelValues("success").Inc()
    w.WriteHeader(http.StatusNoContent)
}
```

#### 4.1.4 WebSocket Handler

```go
// websocket/handler.go

func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
    // 1. Upgrade HTTP to WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        h.logger.Error("WS upgrade failed", zap.Error(err))
        return
    }

    // 2. Authenticate from query params
    token := r.URL.Query().Get("token")
    claims, err := h.auth.ValidateToken(token)
    if err != nil {
        conn.WriteJSON(WSMessage{Type: "error", Data: "UNAUTHORIZED"})
        conn.Close()
        return
    }

    // 3. Register client with hub
    client := &Client{
        Conn:  conn,
        OrgID: claims.OrgID,
        Send:  make(chan []byte, 256),
    }

    h.hub.Register(client)

    // 4. Start read/write pumps
    go client.WritePump()
    go client.ReadPump(h.hub)
}
```

### 4.2 fleet-core-service (Port 8084)

#### 4.2.1 New Endpoints

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/v1/vehicles/locations` | `GetVehicleLocations` | Latest positions for all org vehicles |
| GET | `/api/v1/geofences` | `ListGeofences` | Active geofences for org |
| POST | `/api/v1/geofences` | `CreateGeofence` | Create new geofence |
| PUT | `/api/v1/geofences/{id}` | `UpdateGeofence` | Update geofence |
| DELETE | `/api/v1/geofences/{id}` | `DeleteGeofence` | Soft-delete geofence |

#### 4.2.2 Geofence Management

```go
// handler/geofence.go

func (h *GeofenceHandler) CreateGeofence(w http.ResponseWriter, r *http.Request) {
    var req CreateGeofenceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
        return
    }

    // Validate geometry based on type
    switch req.Type {
    case "circle":
        if req.CenterLat == nil || req.CenterLng == nil || req.RadiusM == nil {
            h.respondError(w, http.StatusBadRequest, "INVALID_GEOFENCE", 
                "Circle geofence requires center_lat, center_lng, radius_m")
            return
        }
    case "polygon":
        if len(req.Polygon) < 3 {
            h.respondError(w, http.StatusBadRequest, "INVALID_GEOFENCE",
                "Polygon must have at least 3 points")
            return
        }
    }

    orgID := h.extractOrgID(r)
    geofence := model.Geofence{
        OrgID:       orgID,
        Name:        req.Name,
        Description: req.Description,
        Type:        req.Type,
        // ... map fields
    }

    created, err := h.geofenceService.Create(r.Context(), geofence)
    if err != nil {
        h.respondError(w, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
        return
    }

    h.respondJSON(w, http.StatusCreated, created)
}
```

### 4.3 web-bff (Port 8081)

#### 4.3.1 Route Configuration Changes

```go
// routes/tracking.go

func RegisterTrackingRoutes(router *mux.Router) {
    // Vehicle locations (cached 5s)
    router.HandleFunc("/api/v1/vehicles/locations", 
        proxyHandler("http://fleet-core-service:8084")).
        Methods("GET").
        Middleware(authMiddleware, cacheMiddleware(5*time.Second))

    // Telemetry history (no cache)
    router.HandleFunc("/api/v1/vehicles/{id}/telemetry/history",
        proxyHandler("http://tracking-service:8082")).
        Methods("GET").
        Middleware(authMiddleware)

    // Geofence CRUD
    router.HandleFunc("/api/v1/geofences",
        proxyHandler("http://fleet-core-service:8084")).
        Methods("GET", "POST").
        Middleware(authMiddleware, cacheMiddleware(30*time.Second))

    router.HandleFunc("/api/v1/geofences/{id}",
        proxyHandler("http://fleet-core-service:8084")).
        Methods("PUT", "DELETE").
        Middleware(authMiddleware)

    // WebSocket endpoint
    router.HandleFunc("/ws/telemetry",
        wsProxyHandler("http://tracking-service:8082")).
        Methods("GET").
        Middleware(authMiddleware)
}
```

#### 4.3.2 WebSocket Proxy

```go
// proxy/websocket.go

func WSProxyHandler(targetURL string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Upgrade client connection
        clientConn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            return
        }
        defer clientConn.Close()

        // 2. Connect to backend tracking-service
        backendURL := targetURL + r.URL.Path + "?" + r.URL.RawQuery
        backendConn, _, err := websocket.DefaultDialer.Dial(backendURL, nil)
        if err != nil {
            clientConn.WriteJSON(WSMessage{Type: "error", Data: "BACKEND_UNAVAILABLE"})
            return
        }
        defer backendConn.Close()

        // 3. Bidirectional relay
        errCh := make(chan error, 2)
        go relayMessages(clientConn, backendConn, errCh)
        go relayMessages(backendConn, clientConn, errCh)
        
        <-errCh // Wait for first error, then close both
    }
}
```

---

## 5. Frontend Implementation Plan

### 5.1 Component Architecture

```
src/features/tracking/
├── components/
│   ├── VehicleMap.tsx               # MapLibre wrapper, main entry point
│   ├── VehicleMarker.tsx            # Single vehicle marker (canvas layer)
│   ├── MarkerClusterGroup.tsx       # Clustered marker layer
│   ├── HistoryTrailLayer.tsx        # Polyline overlay for trip history
│   ├── GeofenceLayer.tsx           # Geofence boundary overlays
│   ├── VehicleDetailPanel.tsx       # Slide-over detail panel
│   ├── SearchBar.tsx               # Debounced vehicle search
│   ├── FilterBar.tsx               # Status/type filter chips
│   ├── PlaybackControls.tsx        # History playback UI
│   ├── MapStatusBar.tsx            # Bottom status bar
│   ├── MapSettingsModal.tsx        # Map config (refresh interval, etc.)
│   └── ConnectionIndicator.tsx     # WS connection status dot
├── hooks/
│   ├── useVehiclePositions.ts      # WS subscription + position state
│   ├── useVehicleHistory.ts        # History query + trail rendering
│   ├── useGeofences.ts             # Geofence data management
│   ├── useVehicleSearch.ts         # Debounced search logic
│   └── useMapBounds.ts             # Auto-fit map to visible vehicles
├── store/
│   ├── trackingSlice.ts            # Redux state + reducers
│   └── websocketMiddleware.ts      # Redux middleware for WS events
├── services/
│   ├── trackingApi.ts              # RTK Query endpoints
│   └── websocketService.ts         # WebSocket connection manager
├── types/
│   └── tracking.ts                 # TypeScript interfaces
└── utils/
    ├── statusColor.ts              # Status → color mapping
    ├── geometry.ts                 # Point-in-polygon checks
    └── sampling.ts                 # History resolution calculator
```

### 5.2 TypeScript Interfaces

```typescript
// types/tracking.ts

export interface VehiclePosition {
    vehicle_id: string;
    org_id: string;
    registration: string;
    make: string;
    model: string;
    year: number;
    vehicle_status: 'moving' | 'idle' | 'stopped' | 'offline';
    current_driver_id: string | null;
    driver_name: string | null;
    driver_phone: string | null;
    last_update: string;           // ISO 8601
    latitude: number;
    longitude: number;
    speed_kmh: number | null;
    heading: number | null;
    ignition: boolean | null;
    fuel_level: number | null;
    derived_status: VehicleStatus;
}

export interface TelemetryPoint {
    time: string;                  // ISO 8601
    vehicle_id: string;
    latitude: number;
    longitude: number;
    speed_kmh: number | null;
    heading: number | null;
    altitude_m: number | null;
    ignition: boolean | null;
    fuel_level: number | null;
}

export interface Geofence {
    id: string;
    org_id: string;
    name: string;
    description: string;
    type: 'circle' | 'polygon' | 'corridor';
    center_lat: number | null;
    center_lng: number | null;
    radius_m: number | null;
    polygon: [number, number][] | null;  // GeoJSON-style coordinates
    status: 'active' | 'inactive';
    alert_channels: string[];
    created_at: string;
    updated_at: string;
}

export interface GeofenceEvent {
    id: string;
    geofence_id: string;
    vehicle_id: string;
    event_type: 'entry' | 'exit';
    occurred_at: string;
    latitude: number;
    longitude: number;
    alert_sent: boolean;
}

export type VehicleStatus = 'moving' | 'idle' | 'stopped' | 'offline';
export type PlaybackSpeed = 1 | 2 | 4 | 8;

export interface HistoryRequest {
    vehicle_id: string;
    start: string;                 // ISO 8601
    end: string;                   // ISO 8601
    resolution?: number;           // seconds (default: 60)
}
```

### 5.3 Redux State Shape

```typescript
// store/trackingSlice.ts

interface TrackingState {
    // Vehicle data
    vehicles: Record<string, VehiclePosition>;
    
    // Geofence data
    geofences: Geofence[];
    geofenceEvents: Record<string, GeofenceEvent[]>; // vehicle_id → events
    
    // Selection state
    selectedVehicleId: string | null;
    historyTrail: {
        vehicle_id: string;
        points: TelemetryPoint[];
        summary: {
            total_distance_km: number;
            avg_speed_kmh: number;
            max_speed_kmh: number;
            duration_minutes: number;
        };
    } | null;
    
    // Filters
    searchQuery: string;
    statusFilters: VehicleStatus[];
    typeFilters: string[];
    
    // WebSocket state
    wsConnected: boolean;
    wsLastMessage: number | null;
    wsSubscribedVehicles: string[];
    wsReconnectAttempts: number;
    
    // Playback state
    isPlaying: boolean;
    playbackSpeed: PlaybackSpeed;
    playbackPosition: number | null;  // index in history trail
    
    // UI state
    detailPanelOpen: boolean;
    historyPanelOpen: boolean;
    mapSettingsOpen: boolean;
    mapBounds: [number, number][] | null;  // [lat, lng] bounds
    
    // Loading state
    loading: {
        vehicles: boolean;
        history: boolean;
        geofences: boolean;
    };
    
    // Error state
    error: string | null;
    fallbackToPolling: boolean;
}
```

### 5.4 WebSocket Service

```typescript
// services/websocketService.ts

class TelemetryWebSocket {
    private ws: WebSocket | null = null;
    private reconnectAttempts = 0;
    private readonly MAX_RECONNECT = 5;
    private readonly BASE_DELAY = 1000;
    private token: string = '';
    private orgId: string = '';

    connect(token: string, orgId: string) {
        this.token = token;
        this.orgId = orgId;
        
        const wsUrl = `${this.getWsBaseUrl()}/ws/telemetry?token=${token}&org_id=${orgId}`;
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = this.handleOpen.bind(this);
        this.ws.onmessage = this.handleMessage.bind(this);
        this.ws.onclose = this.handleClose.bind(this);
        this.ws.onerror = this.handleError.bind(this);
    }

    private handleOpen() {
        this.reconnectAttempts = 0;
        store.dispatch(setWsConnected(true));
        store.dispatch(setWsReconnectAttempts(0));
    }

    private handleMessage(event: MessageEvent) {
        const message = JSON.parse(event.data) as WSMessage;
        
        switch (message.type) {
            case 'position_update':
                store.dispatch(updateVehiclePosition(message.data as VehiclePosition));
                break;
            case 'geofence_event':
                store.dispatch(addGeofenceEvent(message.data as GeofenceEvent));
                break;
            case 'connection_ack':
                break;
            case 'error':
                store.dispatch(setError(message.data as string));
                break;
        }
    }

    private handleClose() {
        store.dispatch(setWsConnected(false));
        this.scheduleReconnect();
    }

    private handleError(error: Event) {
        console.error('WebSocket error:', error);
    }

    private scheduleReconnect() {
        if (this.reconnectAttempts >= this.MAX_RECONNECT) {
            store.dispatch(setError('Connection lost. Refresh page.'));
            store.dispatch(setFallbackToPolling(true));
            this.startPolling();
            return;
        }

        const delay = this.BASE_DELAY * Math.pow(2, this.reconnectAttempts);
        this.reconnectAttempts++;
        store.dispatch(setWsReconnectAttempts(this.reconnectAttempts));

        setTimeout(() => {
            this.connect(this.token, this.orgId);
        }, delay);
    }

    private startPolling() {
        const poll = async () => {
            try {
                const response = await fetch('/api/v1/vehicles/locations', {
                    headers: {
                        'Authorization': `Bearer ${this.token}`,
                        'X-Org-ID': this.orgId
                    }
                });
                const data = await response.json();
                store.dispatch(setAllVehiclePositions(data.data));
            } catch (error) {
                console.error('Polling failed:', error);
            }
            setTimeout(poll, 10000); // 10-second interval
        };
        poll();
    }

    private getWsBaseUrl(): string {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        return `${protocol}//${window.location.host}`;
    }
}

export const telemetryWS = new TelemetryWebSocket();
```

### 5.5 Map Component

```typescript
// components/VehicleMap.tsx

import maplibregl from 'maplibre-gl';
import Supercluster from 'supercluster';
import { useVehiclePositions } from '../hooks/useVehiclePositions';
import { statusColor } from '../utils/statusColor';

export const VehicleMap: React.FC = () => {
    const mapRef = useRef<maplibregl.Map | null>(null);
    const superclusterRef = useRef<Supercluster | null>(null);
    const vehicles = useVehiclePositions();
    
    useEffect(() => {
        // Initialize MapLibre
        mapRef.current = new maplibregl.Map({
            container: 'map-container',
            style: 'https://demotiles.maplibre.org/style.json', // Replace with production tiles
            center: [78.9629, 20.5937], // India center
            zoom: 4,
        });

        // Add custom canvas layer for markers
        const customLayer: maplibregl.CustomLayerInterface = {
            id: 'vehicle-markers',
            type: 'custom',
            renderingMode: '2d',
            render: (gl, matrix) => {
                renderMarkersOnCanvas(gl, matrix, vehicles, superclusterRef.current);
            },
        };

        mapRef.current.addLayer(customLayer);

        return () => {
            mapRef.current?.remove();
        };
    }, []);

    useEffect(() => {
        // Update supercluster with new vehicle positions
        const points = vehicles.map(v => ({
            type: 'Feature' as const,
            properties: {
                vehicle_id: v.vehicle_id,
                status: v.derived_status,
                registration: v.registration,
            },
            geometry: {
                type: 'Point' as const,
                coordinates: [v.longitude, v.latitude],
            },
        }));

        superclusterRef.current?.load(points);
        mapRef.current?.triggerRepaint();
    }, [vehicles]);

    return <div id="map-container" className="w-full h-full" />;
};
```

---

## 6. Database Design

### 6.1 Schema Overview

```
┌─────────────────────────┐
│     telemetry_points     │  (TimescaleDB hypertable)
│  time PK, vehicle_id PK │
│  lat, lng, speed, ...   │
│  Partitioned: monthly    │
│  Retention: 90 days      │
└──────────┬──────────────┘
           │
           │ JOIN vehicle_id
           ▼
┌─────────────────────────┐
│       vehicles           │
│  id PK, org_id FK       │
│  registration, make, ...│
│  RLS: org_isolation     │
└──────────┬──────────────┘
           │
           │ JOIN current_driver_id
           ▼
┌─────────────────────────┐
│        drivers           │
│  id PK, org_id FK       │
│  name, phone, ...       │
│  RLS: org_isolation     │
└─────────────────────────┘

┌─────────────────────────┐     ┌─────────────────────────┐
│       geofences          │     │     geofence_events      │
│  id PK, org_id FK       │◄────│  id PK                   │
│  name, type, geometry   │     │  geofence_id FK          │
│  status, alert_channels │     │  vehicle_id FK           │
│  RLS: org_isolation     │     │  event_type, occurred_at │
└─────────────────────────┘     └─────────────────────────┘

┌──────────────────────────────────────┐
│  latest_vehicle_positions            │
│  (MATERIALIZED VIEW)                 │
│  vehicle_id, org_id, registration   │
│  lat, lng, speed, status, ...       │
│  Refreshed: every 5 seconds          │
│  Index: (org_id)                     │
└──────────────────────────────────────┘
```

### 6.1 telemetry_points

```sql
CREATE TABLE IF NOT EXISTS telemetry_points (
    time        TIMESTAMPTZ NOT NULL,
    vehicle_id  UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    latitude    DOUBLE PRECISION NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude   DOUBLE PRECISION NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    speed_kmh   DOUBLE PRECISION CHECK (speed_kmh >= 0),
    heading     DOUBLE PRECISION CHECK (heading BETWEEN 0 AND 360),
    altitude_m  DOUBLE PRECISION,
    ignition    BOOLEAN,
    fuel_level  DOUBLE PRECISION CHECK (fuel_level BETWEEN 0 AND 100),
    raw_data    JSONB,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- TimescaleDB hypertable (monthly partitions)
SELECT create_hypertable('telemetry_points', 'time',
    chunk_time_interval => INTERVAL '1 month',
    if_not_exists => true,
    migrate_data => false
);

-- Compression: chunks older than 7 days
SELECT add_compression_policy('telemetry_points',
    INTERVAL '7 days', if_not_exists => true);

-- Retention: drop raw data older than 90 days
SELECT add_drop_chunks_policy('telemetry_points',
    INTERVAL '90 days', if_not_exists => true);

-- Indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_vehicle_time
    ON telemetry_points (vehicle_id, time DESC);
```

### 6.2 latest_vehicle_positions (Materialized View)

```sql
CREATE MATERIALIZED VIEW IF NOT EXISTS latest_vehicle_positions AS
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

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_latest_positions_org
    ON latest_vehicle_positions (org_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_latest_positions_status
    ON latest_vehicle_positions (derived_status);
```

**Refresh Strategy:** Application-level refresh every 5 seconds using `REFRESH MATERIALIZED VIEW CONCURRENTLY`. This allows reads during refresh without locking.

```go
// database/refresher.go

type MaterializedViewRefresher struct {
    db       *sql.DB
    logger   *zap.Logger
    interval time.Duration
    ticker   *time.Ticker
}

func (r *MaterializedViewRefresher) Start(ctx context.Context) {
    r.ticker = time.NewTicker(r.interval)
    go func() {
        for {
            select {
            case <-r.ticker.C:
                start := time.Now()
                _, err := r.db.ExecContext(ctx, 
                    "REFRESH MATERIALIZED VIEW CONCURRENTLY latest_vehicle_positions")
                if err != nil {
                    r.logger.Error("Failed to refresh materialized view", zap.Error(err))
                } else {
                    r.logger.Debug("Materialized view refreshed",
                        zap.Duration("duration", time.Since(start)))
                }
            case <-ctx.Done():
                r.ticker.Stop()
                return
            }
        }
    }()
}
```

### 6.3 geofences

```sql
CREATE TABLE IF NOT EXISTS geofences (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    type        VARCHAR(20) NOT NULL CHECK (type IN ('circle', 'polygon', 'corridor')),
    center_lat  DOUBLE PRECISION,
    center_lng  DOUBLE PRECISION,
    radius_m    DOUBLE PRECISION,
    polygon     GEOMETRY(POLYGON, 4326),
    corridor    GEOMETRY(LINESTRING, 4326),
    corridor_buffer_m DOUBLE PRECISION,
    status      VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    alert_channels JSONB DEFAULT '["in_app"]' CHECK (
        alert_channels ?| ARRAY['in_app', 'email', 'sms', 'webhook']
    ),
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Spatial indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_polygon
    ON geofences USING GIST (polygon);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_corridor
    ON geofences USING GIST (corridor);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_org_status
    ON geofences (org_id, status);
```

### 6.4 geofence_events

```sql
CREATE TABLE IF NOT EXISTS geofence_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    geofence_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
    vehicle_id  UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    event_type  VARCHAR(20) NOT NULL CHECK (event_type IN ('entry', 'exit')),
    occurred_at TIMESTAMPTZ NOT NULL,
    latitude    DOUBLE PRECISION,
    longitude   DOUBLE PRECISION,
    alert_sent  BOOLEAN DEFAULT false,
    alert_channels JSONB,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_events_vehicle
    ON geofence_events (vehicle_id, occurred_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_events_geofence
    ON geofence_events (geofence_id, occurred_at DESC);

-- RLS
ALTER TABLE geofence_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY geofence_events_org_isolation ON geofence_events
    USING (vehicle_id IN (
        SELECT id FROM vehicles 
        WHERE org_id = current_setting('app.current_org_id')::uuid
    ));
```

### 6.5 Row-Level Security Policies

```sql
-- Enable RLS on all tracking-related tables
ALTER TABLE telemetry_points ENABLE ROW LEVEL SECURITY;
ALTER TABLE geofences ENABLE ROW LEVEL SECURITY;
ALTER TABLE geofence_events ENABLE ROW LEVEL SECURITY;

-- Telemetry: access controlled via vehicle JOIN
CREATE POLICY telemetry_org_isolation ON telemetry_points
    USING (vehicle_id IN (
        SELECT id FROM vehicles 
        WHERE org_id = current_setting('app.current_org_id')::uuid
    ));

-- Geofences: direct org_id check
CREATE POLICY geofence_org_isolation ON geofences
    USING (org_id = current_setting('app.current_org_id')::uuid);

-- Geofence events: via vehicle JOIN
CREATE POLICY geofence_events_org_isolation ON geofence_events
    USING (vehicle_id IN (
        SELECT id FROM vehicles 
        WHERE org_id = current_setting('app.current_org_id')::uuid
    ));
```

---

## 7. API Contracts

### 7.1 Telemetry Ingest

```
POST /api/v1/telemetry
Content-Type: application/json
Authorization: Bearer <token>

Request Body:
{
    "vehicle_id": "550e8400-e29b-41d4-a716-446655440000",
    "latitude": 28.6139,
    "longitude": 77.2090,
    "speed_kmh": 45.5,
    "heading": 180.0,
    "altitude_m": 216.0,
    "ignition": true,
    "fuel_level": 72.5,
    "raw_data": {"obd_protocol": "CAN", "dtc_codes": []}
}

Response 204 No Content:
(empty body)

Response 400 Bad Request:
{
    "error": {
        "code": "INVALID_COORDINATES",
        "message": "Coordinates out of range",
        "details": [{"field": "latitude", "message": "must be between -90 and 90"}]
    }
}

Response 429 Too Many Requests:
{
    "error": {
        "code": "RATE_LIMITED",
        "message": "Max 1 update per minute per vehicle"
    }
}
```

### 7.2 Vehicle Locations

```
GET /api/v1/vehicles/locations
Authorization: Bearer <token>
X-Org-ID: <org_id>

Query Parameters:
  status_filter: moving,idle,stopped,offline (optional, comma-separated)
  type_filter: truck,bus,van,bike (optional, comma-separated)
  search: registration substring (optional)
  limit: max results (optional, default: 500, max: 500)
  offset: pagination (optional, default: 0)

Response 200 OK:
{
    "data": [
        {
            "vehicle_id": "550e8400-e29b-41d4-a716-446655440000",
            "org_id": "org-uuid",
            "registration": "DL-01-AB-1234",
            "make": "Tata",
            "model": "Ace Gold",
            "year": 2024,
            "vehicle_status": "active",
            "current_driver_id": "driver-uuid",
            "driver_name": "Rajesh Kumar",
            "driver_phone": "+91-9876543210",
            "last_update": "2026-04-07T10:30:00Z",
            "latitude": 28.6139,
            "longitude": 77.2090,
            "speed_kmh": 45.5,
            "heading": 180.0,
            "ignition": true,
            "fuel_level": 72.5,
            "derived_status": "moving"
        }
    ],
    "meta": {
        "total": 500,
        "limit": 500,
        "offset": 0,
        "requestId": "abc-123",
        "timestamp": "2026-04-07T10:30:05Z"
    }
}
```

### 7.3 Telemetry History

```
GET /api/v1/vehicles/{vehicle_id}/telemetry/history
Authorization: Bearer <token>
X-Org-ID: <org_id>

Query Parameters:
  start: 2026-04-06T10:00:00Z (required)
  end: 2026-04-07T10:00:00Z (required)
  resolution: 60 (optional, seconds, default: 60)

Resolution auto-calculates based on time range:
  < 1 hour:    10 seconds
  1-6 hours:   30 seconds
  6-24 hours:  60 seconds
  1-7 days:    300 seconds
  7-30 days:   900 seconds
  30-90 days:  3600 seconds

Response 200 OK:
{
    "data": {
        "vehicle_id": "550e8400-e29b-41d4-a716-446655440000",
        "points": [
            {
                "time": "2026-04-06T10:00:00Z",
                "latitude": 28.6139,
                "longitude": 77.2090,
                "speed_kmh": 45.5,
                "heading": 180.0,
                "ignition": true,
                "fuel_level": 72.5
            }
        ],
        "summary": {
            "total_distance_km": 125.3,
            "avg_speed_kmh": 42.1,
            "max_speed_kmh": 85.0,
            "duration_minutes": 180,
            "stops_count": 3
        }
    },
    "meta": {
        "resolution_seconds": 60,
        "original_points": 1440,
        "downsampled_points": 1440,
        "requestId": "abc-123"
    }
}
```

### 7.4 Geofence List

```
GET /api/v1/geofences
Authorization: Bearer <token>
X-Org-ID: <org_id>

Query Parameters:
  status: active,inactive (optional, default: active)
  type: circle,polygon,corridor (optional)

Response 200 OK:
{
    "data": [
        {
            "id": "geofence-uuid",
            "org_id": "org-uuid",
            "name": "Delhi Warehouse",
            "description": "Main distribution center",
            "type": "circle",
            "center_lat": 28.6139,
            "center_lng": 77.2090,
            "radius_m": 500,
            "polygon": null,
            "status": "active",
            "alert_channels": ["in_app", "email"],
            "created_at": "2026-04-01T08:00:00Z",
            "updated_at": "2026-04-01T08:00:00Z"
        }
    ],
    "meta": {
        "total": 15,
        "requestId": "abc-123"
    }
}
```

### 7.5 Create Geofence

```
POST /api/v1/geofences
Content-Type: application/json
Authorization: Bearer <token>
X-Org-ID: <org_id>

Request Body (circle):
{
    "name": "Mumbai Depot",
    "description": "Fleet parking area",
    "type": "circle",
    "center_lat": 19.0760,
    "center_lng": 72.8777,
    "radius_m": 300,
    "alert_channels": ["in_app", "email", "sms"]
}

Request Body (polygon):
{
    "name": "Industrial Zone",
    "description": "Factory perimeter",
    "type": "polygon",
    "polygon": [
        [77.2090, 28.6139],
        [77.2100, 28.6139],
        [77.2100, 28.6150],
        [77.2090, 28.6150],
        [77.2090, 28.6139]
    ],
    "alert_channels": ["in_app"]
}

Response 201 Created:
{
    "data": {
        "id": "geofence-uuid",
        "org_id": "org-uuid",
        "name": "Mumbai Depot",
        "type": "circle",
        ...
    },
    "meta": {"requestId": "abc-123"}
}
```

---

## 8. WebSocket Protocol

### 8.1 Connection Setup

```
Client → Server:
  wss://api.flyfleet.io/ws/telemetry?token=<jwt>&org_id=<org_uuid>

Server → Client (on connect):
  {"type": "connection_ack", "data": {"client_id": "uuid"}}
```

### 8.2 Message Format

All messages follow this envelope:

```json
{
    "type": "position_update|geofence_event|error|subscription_ack",
    "data": { ... },
    "timestamp": "2026-04-07T10:30:00Z"
}
```

### 8.3 Message Types

#### Position Update (Server → Client)

```json
{
    "type": "position_update",
    "data": {
        "vehicle_id": "550e8400-e29b-41d4-a716-446655440000",
        "registration": "DL-01-AB-1234",
        "latitude": 28.6139,
        "longitude": 77.2090,
        "speed_kmh": 45.5,
        "heading": 180.0,
        "derived_status": "moving",
        "last_update": "2026-04-07T10:30:00Z"
    },
    "timestamp": "2026-04-07T10:30:00Z"
}
```

#### Geofence Event (Server → Client)

```json
{
    "type": "geofence_event",
    "data": {
        "event_id": "event-uuid",
        "geofence_id": "geofence-uuid",
        "geofence_name": "Delhi Warehouse",
        "vehicle_id": "550e8400-e29b-41d4-a716-446655440000",
        "registration": "DL-01-AB-1234",
        "event_type": "entry",
        "occurred_at": "2026-04-07T10:30:00Z",
        "latitude": 28.6139,
        "longitude": 77.2090
    },
    "timestamp": "2026-04-07T10:30:00Z"
}
```

#### Subscription Management (Client → Server)

```json
{
    "type": "subscribe",
    "data": {
        "vehicle_ids": ["uuid1", "uuid2", "..."]
    }
}

Server Response:
{
    "type": "subscription_ack",
    "data": {
        "subscribed_count": 3
    }
}
```

#### Error (Server → Client)

```json
{
    "type": "error",
    "data": "UNAUTHORIZED",
    "timestamp": "2026-04-07T10:30:00Z"
}
```

### 8.4 Connection Lifecycle

```
Client                              Server
  │                                   │
  │────── WebSocket Upgrade ─────────>│
  │   GET /ws/telemetry?token=...    │
  │                                   │
  │<───── 101 Switching Protocols ───│
  │                                   │
  │<───── connection_ack ────────────│
  │                                   │
  │────── subscribe ────────────────>│
  │   {vehicle_ids: [...]}           │
  │                                   │
  │<───── subscription_ack ──────────│
  │                                   │
  │<───── position_update ───────────│  (every 5s batch)
  │<───── position_update ───────────│
  │<───── geofence_event ────────────│  (on trigger)
  │                                   │
  │────── ping (application) ───────>│  (every 30s)
  │<───── pong ──────────────────────│
  │                                   │
  │────── Close (1000) ─────────────>│
  │                                   │
```

### 8.5 Heartbeat / Ping-Pong

```
Client sends application-level ping every 30 seconds:
  {"type": "ping"}

Server responds:
  {"type": "pong", "timestamp": "2026-04-07T10:30:00Z"}

If no pong received within 60 seconds → client considers connection dead → reconnect
```

---

## 9. Security Implementation

### 9.1 Authentication

| Layer | Mechanism | Details |
|-------|-----------|---------|
| REST API | JWT Bearer Token | `Authorization: Bearer <token>` |
| WebSocket | JWT Query Parameter | `?token=<jwt>` (initial handshake only) |
| Service-to-Service | mTLS | Between microservices in Docker network |

**JWT Claims Required:**
```json
{
    "sub": "user-uuid",
    "org_id": "org-uuid",
    "roles": ["dispatcher"],
    "exp": 1712500000,
    "iat": 1712496400
}
```

### 9.2 Authorization Matrix

| Endpoint | super_admin | org_admin | dispatcher | driver |
|----------|:-----------:|:---------:|:----------:|:------:|
| POST /telemetry | ✓ (all orgs) | ✓ (own org) | ✓ (own org) | ✓ (own org) |
| GET /vehicles/locations | ✓ (all orgs) | ✓ (own org) | ✓ (own org) | ✗ (assigned only) |
| GET /vehicles/{id}/history | ✓ | ✓ | ✓ | ✓ (own vehicle) |
| GET /geofences | ✓ | ✓ | ✓ | ✗ |
| POST /geofences | ✓ | ✓ | ✗ | ✗ |
| PUT /geofences/{id} | ✓ | ✓ | ✗ | ✗ |
| DELETE /geofences/{id} | ✓ | ✓ | ✗ | ✗ |
| WS /ws/telemetry | ✓ | ✓ | ✓ | ✓ |

### 9.3 Row-Level Security

```sql
-- Set org context at session start (application-level)
SET app.current_org_id = 'org-uuid';

-- super_admin bypass
-- For cross-org queries, super_admin can SET app.current_org_id = NULL
-- and bypass RLS via a separate policy
CREATE POLICY super_admin_bypass ON vehicles
    USING (current_setting('app.user_role') = 'super_admin');
```

### 9.4 Rate Limiting

| Endpoint | Limit | Strategy |
|----------|-------|----------|
| POST /telemetry | 1 req/min/vehicle | Token bucket (Redis) |
| GET /vehicles/locations | 60 req/min/user | Sliding window (Redis) |
| GET /vehicles/{id}/history | 30 req/min/user | Sliding window (Redis) |
| WS /ws/telemetry | 100 concurrent connections/org | Connection counter |

```go
// middleware/rate_limit.go

type RateLimiter struct {
    redis *redis.Client
}

func (rl *RateLimiter) Check(key string, limit int, window time.Duration) bool {
    ctx := context.Background()
    current, _ := rl.redis.Incr(ctx, key).Result()
    if current == 1 {
        rl.redis.Expire(ctx, key, window)
    }
    return int(current) <= limit
}
```

### 9.5 Input Validation

```go
// All external inputs validated before processing

type IngestRequest struct {
    VehicleID uuid.UUID `json:"vehicle_id" validate:"required,uuid"`
    Latitude  float64   `json:"latitude" validate:"required,min=-90,max=90"`
    Longitude float64   `json:"longitude" validate:"required,min=-180,max=180"`
    SpeedKmh  *float64  `json:"speed_kmh" validate:"omitempty,min=0"`
    Heading   *float64  `json:"heading" validate:"omitempty,min=0,max=360"`
    Ignition  *bool     `json:"ignition" validate:"omitempty"`
    FuelLevel *float64  `json:"fuel_level" validate:"omitempty,min=0,max=100"`
    RawData   JSON      `json:"raw_data" validate:"omitempty"`
}
```

### 9.6 Security Headers

```go
// middleware/security.go

func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self' ws: wss:")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Permissions-Policy", "geolocation=()")
        next.ServeHTTP(w, r)
    })
}
```

---

## 10. Performance Engineering

### 10.1 Query Performance Budget

| Query | Target (P95) | Strategy | Fallback |
|-------|-------------|----------|----------|
| Latest positions (500 vehicles) | <200ms | Materialized view + org_id index | Direct DISTINCT ON query |
| 24h history (1-min resolution) | <500ms | time_bucket + chunk pruning | Coarser resolution (5-min) |
| Geofence check (10 geofences) | <50ms | PostGIS spatial index | Reduce to bounding box check |
| Vehicle search (substring) | <10ms | B-tree on registration | LIKE '%query%' with limit 20 |

### 10.2 TimescaleDB Optimization

```sql
-- Chunk pruning: ensure queries include time predicate
EXPLAIN ANALYZE
SELECT time_bucket('60 seconds', time) AS bucket,
       FIRST(latitude, time) AS lat,
       FIRST(longitude, time) AS lng,
       AVG(speed_kmh) AS avg_speed
FROM telemetry_points
WHERE vehicle_id = 'uuid'
  AND time >= '2026-04-06'      -- Chunk pruning trigger
  AND time < '2026-04-07'
GROUP BY bucket
ORDER BY bucket;

-- Verify: only relevant chunks are scanned
-- Look for "Chunks excluded: N" in EXPLAIN output
```

### 10.3 WebSocket Fan-out Optimization

```
Problem: 100 users × 500 vehicles = 50,000 potential updates/second

Solution:
1. Batch position updates: aggregate every 5 seconds (not every GPS ping)
2. Delta updates: only send position if changed >10m from last sent position
3. Compression: enable permessage-deflate on WebSocket
4. Redis pub/sub: horizontal scaling across multiple tracking-service instances

Resulting load:
500 vehicles / 5s batch = 100 updates/second
100 users × 100 updates = 10,000 messages/second (per instance)
With 3 instances: 30,000 messages/second total capacity
```

### 10.4 Frontend Performance

| Optimization | Implementation | Impact |
|-------------|----------------|--------|
| Canvas rendering | Custom MapLibre layer with WebGL canvas | 60fps with 500 markers (vs 15fps DOM) |
| Marker clustering | Supercluster, 30px cluster radius | 500 → 50-100 visual elements at low zoom |
| Memoized selectors | Reselect for filtered vehicle lists | Avoids recalculation on unrelated state changes |
| Debounced search | 300ms debounce on search input | Reduces API calls during typing |
| Virtual scrolling | react-window for vehicle list panel | Smooth scrolling with 500 items |
| Request deduplication | RTK Query prevents duplicate requests | Eliminates redundant API calls |

### 10.5 Memory Budget

| Component | Allocation | Notes |
|-----------|-----------|-------|
| MapLibre GL | ~80 MB | Vector tiles + canvas + style |
| Redux store (500 vehicles) | ~2 MB | 500 × 4KB position objects |
| WebSocket buffers (100 connections) | ~5 MB | 50KB buffer per connection |
| Browser chrome/extensions | ~30 MB | Baseline overhead |
| **Total** | **~120 MB** | Well within 2GB tab limit |

---

## 11. Testing Strategy

### 11.1 Unit Tests

**Go Services:**
```
services/tracking-service/
├── internal/service/telemetry_test.go    # Ingest logic, validation
├── internal/service/geofence_test.go     # Point-in-polygon, entry/exit
├── internal/websocket/hub_test.go        # Subscribe, unsubscribe, fan-out
└── internal/repository/telemetry_test.go # SQL query correctness
```

**Test Coverage Target:** 80%+ for all new code

```go
func TestTelemetryService_Ingest_ValidPoint(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    mock.ExpectBegin()
    mock.ExpectExec("INSERT INTO telemetry_points").
        WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 28.6139, 77.2090).
        WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    svc := NewTelemetryService(db, nil, nil, zap.NewNop())
    err := svc.Ingest(context.Background(), TelemetryPoint{
        VehicleID: uuid.New(),
        Latitude:  28.6139,
        Longitude: 77.2090,
    })
    assert.NoError(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

**Frontend:**
```
src/features/tracking/
├── components/VehicleMap.test.tsx         # Map initialization, rendering
├── hooks/useVehiclePositions.test.tsx     # WebSocket state management
├── store/trackingSlice.test.ts            # Redux reducers
├── services/websocketService.test.ts      # Connection lifecycle
└── utils/statusColor.test.ts              # Status → color mapping
```

### 11.2 Integration Tests

```go
// tests/integration/tracking_test.go

func TestTelemetryIngestToEndToEnd(t *testing.T) {
    // 1. Start test environment (Docker Compose)
    // 2. POST telemetry point
    // 3. Query materialized view — verify position appears
    // 4. Connect WebSocket client
    // 5. POST another telemetry point
    // 6. Verify WebSocket client receives position_update
    // 7. Verify geofence check executed
}

func TestGeofenceDetection(t *testing.T) {
    // 1. Create circle geofence at known location
    // 2. POST position outside geofence — no event
    // 3. POST position inside geofence — entry event logged
    // 4. POST position outside geofence — exit event logged
    // 5. Verify event sequence: entry → exit
}
```

### 11.3 Load Tests

```bash
# Telemetry ingest load test
python scripts/api_load_tester.py \
    http://localhost:8084/api/v1/telemetry \
    --method POST \
    --body '{"vehicle_id": "test-uuid", "latitude": 28.6, "longitude": 77.2}' \
    --concurrency 50 \
    --duration 60

# Expected: 1000 req/s throughput, <100ms P95 latency

# WebSocket connection stress test
# Simulate 100 concurrent users with 500 vehicles each
# Expected: <5% reconnect rate, <5s position update latency
```

### 11.4 Contract Tests

```bash
# Verify API contracts between services
npm run bff:contract-check

# Verify frontend API integration
npm run frontend-api-check
```

### 11.5 E2E Tests (Playwright)

```typescript
// tests/e2e/vehicle-tracking.spec.ts

test('map displays vehicles with real-time updates', async ({ page }) => {
    // 1. Login as org_admin
    // 2. Navigate to /tracking
    // 3. Verify map loads with vehicle markers
    // 4. Inject WebSocket position_update message
    // 5. Verify marker moves to new position on map
    // 6. Click marker → detail panel opens
    // 7. Verify detail panel shows correct data
});

test('history trail renders correctly', async ({ page }) => {
    // 1. Navigate to /tracking
    // 2. Search for vehicle by registration
    // 3. Open history panel
    // 4. Select 24h time range
    // 5. Verify polyline overlay renders on map
    // 6. Verify summary stats display
    // 7. Click play → markers animate along trail
});

test('geofence event triggers alert', async ({ page }) => {
    // 1. Create circle geofence
    // 2. Ingest position inside geofence
    // 3. Verify geofence event appears in UI
    // 4. Verify notification badge increments
});
```

### 11.6 Performance Tests

```bash
# Map load performance: 500 vehicles
curl -w "@curl-format.txt" -o /dev/null -s \
    "http://localhost:8081/api/v1/vehicles/locations" \
    -H "Authorization: Bearer <token>"

# History query performance: 24h range
curl -w "@curl-format.txt" -o /dev/null -s \
    "http://localhost:8081/api/v1/vehicles/uuid/telemetry/history?start=...&end=..." \
    -H "Authorization: Bearer <token>"

# Browser performance: Chrome DevTools Performance tab
# - First Contentful Paint < 1.5s
# - Time to Interactive < 3s
# - Max memory < 150MB
```

---

## 12. Observability & Monitoring

### 12.1 Prometheus Metrics

```go
// All metrics registered in tracking-service

var (
    telemetryIngestTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "telemetry_ingest_total",
            Help: "Total number of telemetry ingest requests",
        },
        []string{"status"}, // success, error, duplicate, rate_limited
    )

    telemetryIngestLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "telemetry_ingest_latency_ms",
            Help:    "Latency of telemetry ingest operations",
            Buckets: []float64{10, 50, 100, 200, 500, 1000},
        },
        []string{"endpoint"},
    )

    websocketConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "websocket_connections",
            Help: "Current number of active WebSocket connections",
        },
        []string{"status"}, // active, reconnecting, failed
    )

    websocketUpdateLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "websocket_update_latency_ms",
            Help:    "Time from ingest to WS broadcast",
            Buckets: []float64{100, 500, 1000, 2000, 5000},
        },
        nil,
    )

    geofenceDetectionLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "geofence_detection_latency_ms",
            Help:    "Time to check geofences for a position",
            Buckets: []float64{10, 50, 100, 200, 500},
        },
        nil,
    )

    historyQueryLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "history_query_latency_ms",
            Help:    "Latency of history queries",
            Buckets: []float64{100, 500, 1000, 2000, 3000, 5000},
        },
        []string{"time_range"}, // <1h, 1-6h, 6-24h, 1-7d, 7-30d, 30-90d
    )

    mapLoadLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "map_load_latency_ms",
            Help:    "End-to-end map load latency",
            Buckets: []float64{200, 500, 1000, 1500, 2000, 3000},
        },
        []string{"vehicle_count_bucket"}, // <50, 50-200, 200-500, >500
    )

    materializedViewRefreshLatency = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "materialized_view_refresh_latency_ms",
            Help:    "Time to refresh latest_vehicle_positions view",
            Buckets: []float64{100, 200, 500, 1000, 2000},
        },
    )
)
```

### 12.2 Grafana Dashboards

**Dashboard 1: Vehicle Tracking Overview**
- Total vehicles on map (gauge)
- Active WebSocket connections (gauge)
- Telemetry ingest rate (req/s)
- Position update latency P50/P95/P99 (histogram)
- Map load latency P50/P95/P99 (histogram)

**Dashboard 2: Geofence Monitoring**
- Geofence events per hour (counter)
- Geofence detection latency P95 (histogram)
- Geofences by type/status (pie chart)
- Alert delivery success rate (gauge)

**Dashboard 3: Database Performance**
- TimescaleDB chunk scan time (histogram)
- Materialized view refresh time (histogram)
- DB connection pool utilization (gauge)
- Slow queries (>1s) count (counter)

### 12.3 Alerting Rules

```yaml
# alerting/tracking-rules.yaml
groups:
  - name: vehicle_tracking
    rules:
      - alert: MapLoadP95High
        expr: histogram_quantile(0.95, rate(map_load_latency_ms_bucket[5m])) > 3000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Map load P95 latency >3s"
          runbook_url: "https://wiki.flyfleet.io/runbooks/map-load-latency"

      - alert: WebSocketDisconnectRateHigh
        expr: |
          rate(websocket_connections{status="failed"}[5m]) / 
          rate(websocket_connections[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "WebSocket disconnect rate >10%"
          runbook_url: "https://wiki.flyfleet.io/runbooks/ws-disconnect-rate"

      - alert: TelemetryIngestErrorRateHigh
        expr: |
          rate(telemetry_ingest_total{status="error"}[5m]) / 
          rate(telemetry_ingest_total[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Telemetry ingest error rate >1%"

      - alert: GeofenceDetectionLag
        expr: geofence_detection_latency_ms > 30000
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Geofence detection taking >30s"

      - alert: DBConnectionPoolExhausted
        expr: |
          pg_stat_activity_count{datname="fleet_db"} / 
          pg_settings{name="max_connections"} > 0.9
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool >90% utilized"

      - alert: MaterializedViewStale
        expr: materialized_view_refresh_latency_ms > 10000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Materialized view refresh taking >10s"
```

### 12.4 Distributed Tracing

**Trace Example: Position Update End-to-End**

```
Trace ID: abc123def456
│
├─ [0ms] GPS Device → fleet-core-service (HTTP POST /telemetry)
│   ├─ [0-2ms] JWT validation
│   ├─ [2-5ms] Input validation
│   ├─ [5-30ms] DB insert (telemetry_points)
│   └─ [30-35ms] Redis publish (telemetry:{org_id})
│
├─ [35ms] tracking-service receives Redis message
│   ├─ [35-50ms] Geofence check (PostGIS ST_Contains × 10 geofences)
│   │   └─ [40-45ms] No geofence triggered → no event logged
│   └─ [50-55ms] WebSocket broadcast to 10 subscribed clients
│
└─ [55-60ms] dashboard-web receives WS message
    └─ [60-65ms] Redux state update + marker re-render
```

**Total E2E Latency: ~65ms (well within 5s target)**

### 12.5 Structured Logging

```go
// All logs in JSON format for aggregation

logger.Info("telemetry_ingested",
    zap.String("vehicle_id", vehicleID.String()),
    zap.Float64("latitude", lat),
    zap.Float64("longitude", lng),
    zap.String("org_id", orgID.String()),
    zap.Duration("duration", time.Since(start)),
    zap.String("trace_id", traceID),
)

logger.Warn("geofence_event_detected",
    zap.String("event_type", "entry"),
    zap.String("geofence_id", geofenceID),
    zap.String("vehicle_id", vehicleID.String()),
    zap.String("org_id", orgID.String()),
)
```

---

## 13. Migration & Deployment Plan

### 13.1 Database Migrations

```sql
-- File: migrations/001_create_telemetry_indexes.up.sql

-- Index for latest position queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_telemetry_vehicle_time
    ON telemetry_points (vehicle_id, time DESC);

-- Spatial indexes for geofences
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_polygon
    ON geofences USING GIST (polygon);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofences_org_status
    ON geofences (org_id, status);

-- Geofence event indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_geofence_events_vehicle
    ON geofence_events (vehicle_id, occurred_at DESC);

-- File: migrations/002_create_materialized_view.up.sql

CREATE MATERIALIZED VIEW IF NOT EXISTS latest_vehicle_positions AS
SELECT DISTINCT ON (t.vehicle_id)
    t.vehicle_id, v.org_id, v.registration, v.make, v.model,
    v.status AS vehicle_status, v.current_driver_id,
    d.name AS driver_name, d.phone AS driver_phone,
    t.time AS last_update, t.latitude, t.longitude,
    t.speed_kmh, t.heading, t.ignition, t.fuel_level,
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

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_latest_positions_org
    ON latest_vehicle_positions (org_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_latest_positions_status
    ON latest_vehicle_positions (derived_status);

-- File: migrations/003_add_geofence_alert_channels.up.sql

ALTER TABLE geofences
    ADD COLUMN IF NOT EXISTS alert_channels JSONB DEFAULT '["in_app"]';

ALTER TABLE geofences
    ADD CONSTRAINT chk_alert_channels CHECK (
        alert_channels ?| ARRAY['in_app', 'email', 'sms', 'webhook']
    );

-- File: migrations/004_enable_rls.up.sql

ALTER TABLE telemetry_points ENABLE ROW LEVEL SECURITY;
ALTER TABLE geofences ENABLE ROW LEVEL SECURITY;
ALTER TABLE geofence_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY telemetry_org_isolation ON telemetry_points
    USING (vehicle_id IN (
        SELECT id FROM vehicles 
        WHERE org_id = current_setting('app.current_org_id')::uuid
    ));

CREATE POLICY geofence_org_isolation ON geofences
    USING (org_id = current_setting('app.current_org_id')::uuid);

CREATE POLICY geofence_events_org_isolation ON geofence_events
    USING (vehicle_id IN (
        SELECT id FROM vehicles 
        WHERE org_id = current_setting('app.current_org_id')::uuid
    ));
```

### 13.2 Feature Flags

```yaml
# config.toml

[feature_flags]
# Kill switch for entire vehicle tracking feature
vehicle_map_enabled = true

# Toggle real-time WebSocket updates
real_time_updates = true

# Toggle geofence alerts (enable after geofence setup complete)
geofence_alerts = false

# Toggle history visualization
history_trail = true

# Performance guardrail: max vehicles on map
max_vehicles_on_map = 500

# Per-org rollout (feature flag by org_id)
[feature_flags.per_org]
beta_orgs = ["org-beta-1", "org-beta-2", "org-beta-3"]
```

### 13.3 Deployment Sequence

```
Phase 1: Infrastructure (Week 1)
═══════════════════════════════════
1. Run database migrations (online, zero downtime)
   → migrations/001_create_telemetry_indexes
   → migrations/002_create_materialized_view
   → migrations/003_add_geofence_alert_channels
   → migrations/004_enable_rls

2. Deploy tracking-service (port 8082)
   → Sidecar with existing services
   → Feature flag: vehicle_map_enabled = false (hidden)

3. Deploy web-bff route updates
   → New proxy routes for tracking endpoints
   → WebSocket proxy handler

Phase 2: Beta Rollout (Week 2-3)
═══════════════════════════════════
4. Deploy dashboard-web with map component
   → Feature flag: enabled for beta_orgs only
   → Internal testing team validates

5. Enable for 3 beta organizations
   → Monitor metrics for 1 week
   → Gather feedback, iterate

6. Performance validation
   → Run load tests against production-like staging
   → Verify P95 latency targets met

Phase 3: General Availability (Week 4-5)
═══════════════════════════════════════════
7. Enable feature flag for all organizations
   → Gradual rollout: 10% → 25% → 50% → 100%
   → Monitor error rates, latency metrics

8. Enable geofence alerts
   → geofence_alerts = true
   → Verify alert delivery pipeline

9. Post-launch monitoring (2 weeks)
   → Watch all Prometheus metrics
   → Address any performance regressions
   → Optimize hot spots

Rollback Plan:
═══════════════
- Flip feature flag: vehicle_map_enabled = false
- If DB issues: run migrations/*.down.sql
- If WS issues: fallback_to_polling = true (auto)
```

### 13.4 Docker Compose Updates

```yaml
# docker-compose.yml (additions)

services:
  tracking-service:
    build:
      context: ./services/tracking-service
      dockerfile: Dockerfile
    ports:
      - "8082:8082"
    environment:
      - DATABASE_URL=postgres://fleet_user:${POSTGRES_PASSWORD}@postgres:5432/fleet_db
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  # web-bff already configured with tracking routes
```

---

## 14. Risk Assessment

### 14.1 Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Materialized view refresh causes lock contention | Medium | High | Use `REFRESH CONCURRENTLY`, monitor refresh latency |
| WebSocket connections exhaust server memory | Medium | High | Connection limit per org, auto-evict idle connections |
| TimescaleDB chunk pruning fails (queries scan all chunks) | Low | High | Verify time predicates in all queries, EXPLAIN ANALYZE |
| PostGIS spatial index ineffective for large polygons | Low | Medium | Test with real geofence data, add bounding box pre-filter |
| Browser OOM with 500+ vehicles | Medium | High | Hard limit at 500, cluster at low zoom, canvas rendering |
| GPS device clock skew causes out-of-order inserts | High | Medium | Use server-side `created_at` for ordering, accept `time` as reported |
| Redis pub/sub message loss during partition | Low | Medium | Idempotent ingest (vehicle_id + time unique constraint) |

### 14.2 Operational Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Telemetry ingest overwhelmed during peak hours | Medium | Medium | Auto-scale fleet-core-service, rate limit per vehicle |
| Third-party map tile provider outage | Low | High | Self-host tile server (OpenStreetMap), fallback style |
| RLS policy misconfiguration leaks cross-org data | Low | Critical | Automated RLS audit tests, peer review all policy changes |
| Geofence false positives (boundary jitter) | Medium | Medium | Add hysteresis buffer (5m radius), debounce consecutive events |

### 14.3 Security Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| JWT token stolen → unauthorized location access | Low | High | Short-lived tokens (1h), MFA for admin roles |
| WebSocket connection hijacked | Low | High | WSS (TLS 1.3), JWT validation on upgrade |
| Telemetry spoofing (fake GPS positions) | Medium | Medium | Rate limiting, vehicle auth tokens, anomaly detection |
| Geofence data exfiltration (competitor intelligence) | Low | Medium | RLS, audit logging, rate limiting on list endpoints |

---

## 15. Implementation Timeline

### Week 1: Foundation

- [ ] Create tracking-service project structure
- [ ] Implement database migrations (indexes, materialized view, RLS)
- [ ] Implement telemetry ingest endpoint (fleet-core-service)
- [ ] Implement WebSocket hub + handler (tracking-service)
- [ ] Implement history query endpoint with time_bucket
- [ ] Run migration tests on staging database

### Week 2: Core Backend

- [ ] Implement geofence detection service (PostGIS ST_Contains)
- [ ] Implement geofence CRUD endpoints (fleet-core-service)
- [ ] Implement Redis pub/sub integration
- [ ] Implement web-bff proxy routes
- [ ] Implement WebSocket proxy in web-bff
- [ ] Add Prometheus metrics to all services
- [ ] Write unit tests (80%+ coverage)

### Week 3: Frontend Foundation

- [ ] Create tracking feature structure (React)
- [ ] Implement VehicleMap component (MapLibre)
- [ ] Implement canvas-based marker rendering
- [ ] Implement marker clustering (Supercluster)
- [ ] Implement Redux state + RTK Query endpoints
- [ ] Implement WebSocket service + Redux middleware
- [ ] Implement vehicle search + filter components

### Week 4: Frontend Features

- [ ] Implement history trail visualization
- [ ] Implement playback controls
- [ ] Implement vehicle detail panel
- [ ] Implement geofence overlay rendering
- [ ] Implement connection status indicator
- [ ] Implement HTTP polling fallback
- [ ] Write frontend unit + integration tests

### Week 5: Integration & Performance

- [ ] End-to-end integration testing
- [ ] Load testing (telemetry ingest, WebSocket fan-out)
- [ ] Performance profiling and optimization
- [ ] E2E tests (Playwright)
- [ ] Security review (RLS audit, auth validation)
- [ ] Documentation (API docs, runbooks)

### Week 6: Beta Launch

- [ ] Deploy to staging
- [ ] Enable for 3 beta organizations
- [ ] Monitor metrics for 1 week
- [ ] Collect feedback, iterate
- [ ] Fix bugs, optimize hot spots

### Week 7-8: General Availability

- [ ] Gradual rollout (10% → 25% → 50% → 100%)
- [ ] Enable geofence alerts
- [ ] Post-launch monitoring (2 weeks)
- [ ] Performance tuning
- [ ] Knowledge transfer to ops team

---

## 16. Appendix

### 16.1 Glossary

| Term | Definition |
|------|-----------|
| **Hypertable** | TimescaleDB auto-partitioned table by time |
| **time_bucket** | TimescaleDB function for grouping data into time intervals |
| **Materialized View** | Pre-computed query result stored as a table |
| **PostGIS** | PostgreSQL spatial extension for geographic operations |
| **ST_Contains** | PostGIS function: checks if geometry A contains geometry B |
| **ST_DWithin** | PostGIS function: checks if geometries are within distance N |
| **Fan-out** | Distributing a single message to multiple subscribers |
| **Supercluster** | Fast geospatial point clustering library |
| **RAF** | requestAnimationFrame — browser API for smooth animations |
| **RLS** | Row-Level Security — PostgreSQL access control at row level |
| **RTK Query** | Redux Toolkit Query — data fetching and caching layer |

### 16.2 Related Documents

- [Vehicle Tracking Architecture](../architecture/VEHICLE_TRACKING.md)
- [Vehicle Tracking Product Spec](../product-spec/VEHICLE_TRACKING.md)
- [FlyFleet OpenAPI Specification](../openapi.yaml)
- [Database Schema Documentation](../docs/database/)
- [Security Architecture](../docs/security/)

### 16.3 Configuration Reference

```toml
# tracking-service config.toml

[server]
port = 8082
host = "0.0.0.0"
read_timeout = "10s"
write_timeout = "10s"
idle_timeout = "120s"

[database]
url = "postgres://fleet_user:${POSTGRES_PASSWORD}@postgres:5432/fleet_db"
max_open_conns = 50
max_idle_conns = 10
conn_max_lifetime = "5m"

[redis]
url = "redis://redis:6379"
pool_size = 20

[websocket]
handshake_timeout = "10s"
read_buffer_size = 4096
write_buffer_size = 4096
max_message_size = 65536
ping_period = "30s"
pong_wait = "60s"

[geofence]
check_interval = "5s"
hysteresis_radius_m = 5.0  # Prevent boundary jitter

[materialized_view]
refresh_interval = "5s"

[rate_limit]
telemetry_per_vehicle_per_min = 1
locations_per_user_per_min = 60
history_per_user_per_min = 30
```

### 16.4 OpenAPI Specification Additions

```yaml
# openapi.yaml (additions for vehicle tracking)

paths:
  /api/v1/telemetry:
    post:
      summary: Ingest GPS position
      operationId: ingestTelemetry
      tags: [tracking]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/IngestRequest'
      responses:
        '204': { description: 'Position accepted' }
        '400': { $ref: '#/components/responses/BadRequest' }
        '429': { $ref: '#/components/responses/RateLimited' }

  /api/v1/vehicles/locations:
    get:
      summary: Get latest positions for all org vehicles
      operationId: getVehicleLocations
      tags: [tracking]
      parameters:
        - name: status_filter
          in: query
          schema: { type: string }
        - name: limit
          in: query
          schema: { type: integer, default: 500, maximum: 500 }
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/VehicleLocationsResponse'

  /api/v1/vehicles/{vehicle_id}/telemetry/history:
    get:
      summary: Get telemetry history for a vehicle
      operationId: getTelemetryHistory
      tags: [tracking]
      parameters:
        - name: vehicle_id
          in: path
          required: true
          schema: { type: string, format: uuid }
        - name: start
          in: query
          required: true
          schema: { type: string, format: date-time }
        - name: end
          in: query
          required: true
          schema: { type: string, format: date-time }
        - name: resolution
          in: query
          schema: { type: integer, default: 60 }
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HistoryResponse'

  /api/v1/geofences:
    get:
      summary: List geofences for organization
      operationId: listGeofences
      tags: [tracking, geofence]
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GeofenceListResponse'
    post:
      summary: Create a new geofence
      operationId: createGeofence
      tags: [tracking, geofence]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateGeofenceRequest'
      responses:
        '201':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GeofenceResponse'

  /ws/telemetry:
    get:
      summary: WebSocket endpoint for real-time position updates
      operationId: websocketTelemetry
      tags: [tracking]
      responses:
        '101': { description: 'Switching Protocols' }

components:
  schemas:
    IngestRequest:
      type: object
      required: [vehicle_id, latitude, longitude]
      properties:
        vehicle_id: { type: string, format: uuid }
        latitude: { type: number, minimum: -90, maximum: 90 }
        longitude: { type: number, minimum: -180, maximum: 180 }
        speed_kmh: { type: number, minimum: 0 }
        heading: { type: number, minimum: 0, maximum: 360 }
        ignition: { type: boolean }
        fuel_level: { type: number, minimum: 0, maximum: 100 }
        raw_data: { type: object }

    VehiclePosition:
      type: object
      properties:
        vehicle_id: { type: string, format: uuid }
        registration: { type: string }
        latitude: { type: number }
        longitude: { type: number }
        speed_kmh: { type: number }
        heading: { type: number }
        derived_status:
          type: string
          enum: [moving, idle, stopped, offline]
        last_update: { type: string, format: date-time }

    Geofence:
      type: object
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        type: { type: string, enum: [circle, polygon, corridor] }
        center_lat: { type: number }
        center_lng: { type: number }
        radius_m: { type: number }
        polygon:
          type: array
          items:
            type: array
            items: { type: number }
            minItems: 2
            maxItems: 2
        status: { type: string, enum: [active, inactive] }
        alert_channels:
          type: array
          items: { type: string, enum: [in_app, email, sms, webhook] }
```

---

**Document Status:** Draft — Ready for Review  
**Next Review:** April 14, 2026  
**Approvers:** Engineering Lead, Architecture Review Board, Product Owner  
**Stakeholders:** Engineering Team, DevOps, Product Team
