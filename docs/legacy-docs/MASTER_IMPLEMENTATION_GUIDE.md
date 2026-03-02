# 🚀 Master Implementation Guide - FlyFleet Production Platform

**Based on doc.txn (14,498 lines) - Complete Architecture Blueprint**

## 📊 Current Status

### ✅ Phase 1: Core Services (COMPLETE)
- Auth Service (8080) - Running
- Vehicle Service (8081) - Running
- Tracking Service (8082) - Running
- Driver Service (8083) - Running
- Trip Service (8084) - Running
- Alert Service (8085) - Running
- PostgreSQL + PostGIS - Running
- Redis - Running
- NATS - Running
- Centrifugo - Running
- Web Admin (React) - Running

### 🎯 Next Phase: Production Enhancement

---

## 🔥 PHASE 2: TIMESCALEDB + TELEMETRY PIPELINE

### **Week 5: TimescaleDB Hypertables Setup**

#### Task 2.1: Enable TimescaleDB Extension
**Priority:** CRITICAL  
**File:** `database/migrations/005_timescaledb_setup.sql`

```sql
-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- GPS Locations Hypertable (from doc.txn lines 3274-3287)
CREATE TABLE IF NOT EXISTS gps_locations (
    time TIMESTAMPTZ NOT NULL,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    device_id UUID,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    altitude DECIMAL(8, 2),
    speed DECIMAL(5, 2), -- km/h
    heading DECIMAL(5, 2), -- degrees 0-360
    accuracy DECIMAL(5, 2), -- meters
    ignition BOOLEAN,
    odometer DECIMAL(10, 2), -- total km
    battery_level INTEGER, -- percentage 0-100
    satellites INTEGER,
    hdop DECIMAL(4, 2),
    raw_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Convert to hypertable with 1-day chunks
SELECT create_hypertable('gps_locations', 'time', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Indexes for fast queries
CREATE INDEX IF NOT EXISTS idx_gps_vehicle_time 
    ON gps_locations (vehicle_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_gps_device_time 
    ON gps_locations (device_id, time DESC) WHERE device_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_gps_location 
    ON gps_locations USING GIST (ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography);

-- Compression policy (90% space savings after 7 days)
SELECT add_compression_policy('gps_locations', INTERVAL '7 days', if_not_exists => TRUE);

-- Retention policy (delete data older than 2 years)
SELECT add_retention_policy('gps_locations', INTERVAL '2 years', if_not_exists => TRUE);

-- OBD Readings Hypertable
CREATE TABLE IF NOT EXISTS obd_readings (
    time TIMESTAMPTZ NOT NULL,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    device_id UUID,
    engine_rpm INTEGER,
    coolant_temp INTEGER, -- Celsius
    fuel_level INTEGER, -- percentage 0-100
    throttle_position INTEGER, -- percentage 0-100
    engine_load INTEGER, -- percentage 0-100
    intake_temp INTEGER, -- Celsius
    maf DECIMAL(6, 2), -- Mass Air Flow g/s
    dtc_codes TEXT[], -- Diagnostic Trouble Codes
    vin VARCHAR(17),
    raw_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

SELECT create_hypertable('obd_readings', 'time',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

CREATE INDEX IF NOT EXISTS idx_obd_vehicle_time 
    ON obd_readings (vehicle_id, time DESC);

SELECT add_compression_policy('obd_readings', INTERVAL '7 days', if_not_exists => TRUE);
SELECT add_retention_policy('obd_readings', INTERVAL '2 years', if_not_exists => TRUE);

-- Fuel Events Table
CREATE TABLE IF NOT EXISTS fuel_events (
    time TIMESTAMPTZ NOT NULL,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    location GEOGRAPHY(POINT, 4326),
    fuel_added_liters DECIMAL(8, 2),
    cost DECIMAL(10, 2),
    odometer_reading DECIMAL(10, 2),
    station_name VARCHAR(255),
    receipt_photo_url TEXT,
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

SELECT create_hypertable('fuel_events', 'time',
    chunk_time_interval => INTERVAL '1 month',
    if_not_exists => TRUE
);

CREATE INDEX IF NOT EXISTS idx_fuel_vehicle_time 
    ON fuel_events (vehicle_id, time DESC);

-- Continuous Aggregates for Analytics (from doc.txn lines 3290-3310)

-- Hourly GPS rollups
CREATE MATERIALIZED VIEW IF NOT EXISTS gps_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    vehicle_id,
    COUNT(*) AS point_count,
    AVG(speed) AS avg_speed,
    MAX(speed) AS max_speed,
    AVG(odometer) AS avg_odometer,
    MAX(odometer) - MIN(odometer) AS distance_km,
    AVG(battery_level) AS avg_battery
FROM gps_locations
GROUP BY bucket, vehicle_id
WITH NO DATA;

-- Refresh policy for continuous aggregate
SELECT add_continuous_aggregate_policy('gps_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- Daily vehicle summary
CREATE MATERIALIZED VIEW IF NOT EXISTS vehicle_daily_stats
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS day,
    vehicle_id,
    COUNT(*) AS gps_points,
    SUM(CASE WHEN ignition = true THEN 1 ELSE 0 END) AS ignition_on_count,
    MAX(odometer) - MIN(odometer) AS distance_km,
    AVG(speed) AS avg_speed,
    MAX(speed) AS max_speed,
    AVG(battery_level) AS avg_battery
FROM gps_locations
GROUP BY day, vehicle_id
WITH NO DATA;

SELECT add_continuous_aggregate_policy('vehicle_daily_stats',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Grant permissions
GRANT SELECT, INSERT ON gps_locations TO fleet;
GRANT SELECT, INSERT ON obd_readings TO fleet;
GRANT SELECT, INSERT ON fuel_events TO fleet;
GRANT SELECT ON gps_hourly TO fleet;
GRANT SELECT ON vehicle_daily_stats TO fleet;
```

**Validation:**
```bash
cd /home/abhishek/Desktop/flyfleet
psql postgres://fleet:fleet@localhost:5433/fleetdb -f database/migrations/005_timescaledb_setup.sql

# Verify hypertables
psql postgres://fleet:fleet@localhost:5433/fleetdb -c "SELECT * FROM timescaledb_information.hypertables;"

# Test insertion
psql postgres://fleet:fleet@localhost:5433/fleetdb -c "
INSERT INTO gps_locations (time, vehicle_id, latitude, longitude, speed, ignition)
VALUES (NOW(), '00000000-0000-0000-0000-000000000001'::uuid, 28.6139, 77.2090, 45.5, true);
"

# Query test
psql postgres://fleet:fleet@localhost:5433/fleetdb -c "
SELECT * FROM gps_locations WHERE time > NOW() - INTERVAL '1 hour';
"
```

**Success Metrics:**
- ✅ TimescaleDB extension enabled
- ✅ 3 hypertables created (gps_locations, obd_readings, fuel_events)
- ✅ Compression policies active
- ✅ Continuous aggregates refreshing
- ✅ Test data inserted and queried successfully

---

#### Task 2.2: GPS Telemetry Ingestion API
**Priority:** CRITICAL  
**File:** `services/microservices/tracking-service/internal/handlers/telemetry_handler.go`

```go
package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/nats-io/nats.go"
    "go.uber.org/zap"
)

// GPSPayload from device (doc.txn lines 3400-3420)
type GPSPayload struct {
    DeviceID    string    `json:"device_id" binding:"required"`
    Timestamp   int64     `json:"timestamp" binding:"required"`
    Latitude    float64   `json:"latitude" binding:"required,min=-90,max=90"`
    Longitude   float64   `json:"longitude" binding:"required,min=-180,max=180"`
    Altitude    *float64  `json:"altitude"`
    Speed       float64   `json:"speed"`
    Heading     float64   `json:"heading"`
    Accuracy    float64   `json:"accuracy"`
    Ignition    bool      `json:"ignition"`
    Odometer    *float64  `json:"odometer"`
    Battery     *int      `json:"battery_level"`
    Satellites  *int      `json:"satellites"`
    HDOP        *float64  `json:"hdop"`
    RawData     json.RawMessage `json:"raw_data,omitempty"`
}

type TelemetryHandler struct {
    db     *pgxpool.Pool
    nc     *nats.Conn
    logger *zap.Logger
}

func NewTelemetryHandler(db *pgxpool.Pool, nc *nats.Conn, logger *zap.Logger) *TelemetryHandler {
    return &TelemetryHandler{
        db:     db,
        nc:     nc,
        logger: logger,
    }
}

// IngestGPS - High-throughput GPS ingestion endpoint
func (h *TelemetryHandler) IngestGPS(c *gin.Context) {
    var payload GPSPayload
    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get vehicle_id from device_id
    var vehicleID uuid.UUID
    err := h.db.QueryRow(context.Background(),
        `SELECT vehicle_id FROM devices WHERE device_id = $1 AND status = 'active'`,
        payload.DeviceID,
    ).Scan(&vehicleID)

    if err != nil {
        h.logger.Error("Device not found or inactive", zap.String("device_id", payload.DeviceID), zap.Error(err))
        c.JSON(http.StatusNotFound, gin.H{"error": "Device not registered"})
        return
    }

    // Convert Unix timestamp to time.Time
    timestamp := time.Unix(payload.Timestamp, 0)

    // Insert to TimescaleDB hypertable (async for performance)
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        _, err := h.db.Exec(ctx, `
            INSERT INTO gps_locations (
                time, vehicle_id, device_id, latitude, longitude, altitude,
                speed, heading, accuracy, ignition, odometer, battery_level,
                satellites, hdop, raw_data
            ) VALUES (
                $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
            )`,
            timestamp, vehicleID, payload.DeviceID, payload.Latitude, payload.Longitude,
            payload.Altitude, payload.Speed, payload.Heading, payload.Accuracy,
            payload.Ignition, payload.Odometer, payload.Battery, payload.Satellites,
            payload.HDOP, payload.RawData,
        )

        if err != nil {
            h.logger.Error("Failed to insert GPS data", zap.Error(err))
            return
        }

        // Update vehicle current location (for live map)
        _, err = h.db.Exec(ctx, `
            UPDATE vehicles 
            SET current_location = ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
                last_seen = $3,
                current_speed = $4,
                ignition_status = $5
            WHERE id = $6`,
            payload.Longitude, payload.Latitude, timestamp, payload.Speed, payload.Ignition, vehicleID,
        )

        if err != nil {
            h.logger.Error("Failed to update vehicle location", zap.Error(err))
        }
    }()

    // Publish to NATS for real-time consumers (alert engine, live map)
    eventData, _ := json.Marshal(map[string]interface{}{
        "vehicle_id": vehicleID.String(),
        "device_id":  payload.DeviceID,
        "latitude":   payload.Latitude,
        "longitude":  payload.Longitude,
        "speed":      payload.Speed,
        "ignition":   payload.Ignition,
        "timestamp":  timestamp.Unix(),
    })

    if err := h.nc.Publish("fleet.telemetry.gps", eventData); err != nil {
        h.logger.Error("Failed to publish GPS event", zap.Error(err))
    }

    // Respond immediately (don't wait for DB write)
    c.JSON(http.StatusAccepted, gin.H{
        "status":     "accepted",
        "vehicle_id": vehicleID.String(),
        "timestamp":  timestamp.Unix(),
    })
}

// IngestOBD - OBD-II data ingestion
func (h *TelemetryHandler) IngestOBD(c *gin.Context) {
    var payload struct {
        DeviceID         string   `json:"device_id" binding:"required"`
        Timestamp        int64    `json:"timestamp" binding:"required"`
        EngineRPM        *int     `json:"engine_rpm"`
        CoolantTemp      *int     `json:"coolant_temp"`
        FuelLevel        *int     `json:"fuel_level"`
        ThrottlePosition *int     `json:"throttle_position"`
        EngineLoad       *int     `json:"engine_load"`
        IntakeTemp       *int     `json:"intake_temp"`
        MAF              *float64 `json:"maf"`
        DTCCodes         []string `json:"dtc_codes"`
        VIN              *string  `json:"vin"`
        RawData          json.RawMessage `json:"raw_data,omitempty"`
    }

    if err := c.ShouldBindJSON(&payload); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var vehicleID uuid.UUID
    err := h.db.QueryRow(context.Background(),
        `SELECT vehicle_id FROM devices WHERE device_id = $1 AND status = 'active'`,
        payload.DeviceID,
    ).Scan(&vehicleID)

    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Device not registered"})
        return
    }

    timestamp := time.Unix(payload.Timestamp, 0)

    // Insert to OBD hypertable
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        _, err := h.db.Exec(ctx, `
            INSERT INTO obd_readings (
                time, vehicle_id, device_id, engine_rpm, coolant_temp, fuel_level,
                throttle_position, engine_load, intake_temp, maf, dtc_codes, vin, raw_data
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
            timestamp, vehicleID, payload.DeviceID, payload.EngineRPM, payload.CoolantTemp,
            payload.FuelLevel, payload.ThrottlePosition, payload.EngineLoad, payload.IntakeTemp,
            payload.MAF, payload.DTCCodes, payload.VIN, payload.RawData,
        )

        if err != nil {
            h.logger.Error("Failed to insert OBD data", zap.Error(err))
        }
    }()

    // Check for DTC codes and trigger alerts
    if len(payload.DTCCodes) > 0 {
        eventData, _ := json.Marshal(map[string]interface{}{
            "vehicle_id": vehicleID.String(),
            "dtc_codes":  payload.DTCCodes,
            "timestamp":  timestamp.Unix(),
        })
        h.nc.Publish("fleet.alerts.dtc", eventData)
    }

    c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

// GetVehicleHistory - Query GPS history from TimescaleDB
func (h *TelemetryHandler) GetVehicleHistory(c *gin.Context) {
    vehicleID := c.Param("id")
    startTime := c.DefaultQuery("start", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
    endTime := c.DefaultQuery("end", time.Now().Format(time.RFC3339))

    start, _ := time.Parse(time.RFC3339, startTime)
    end, _ := time.Parse(time.RFC3339, endTime)

    rows, err := h.db.Query(context.Background(), `
        SELECT time, latitude, longitude, speed, heading, ignition, odometer
        FROM gps_locations
        WHERE vehicle_id = $1 AND time BETWEEN $2 AND $3
        ORDER BY time ASC`,
        vehicleID, start, end,
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    type Point struct {
        Timestamp int64   `json:"timestamp"`
        Latitude  float64 `json:"latitude"`
        Longitude float64 `json:"longitude"`
        Speed     float64 `json:"speed"`
        Heading   float64 `json:"heading"`
        Ignition  bool    `json:"ignition"`
        Odometer  *float64 `json:"odometer,omitempty"`
    }

    var points []Point
    for rows.Next() {
        var p Point
        var t time.Time
        if err := rows.Scan(&t, &p.Latitude, &p.Longitude, &p.Speed, &p.Heading, &p.Ignition, &p.Odometer); err != nil {
            continue
        }
        p.Timestamp = t.Unix()
        points = append(points, p)
    }

    c.JSON(http.StatusOK, gin.H{
        "vehicle_id": vehicleID,
        "start":      start.Unix(),
        "end":        end.Unix(),
        "points":     points,
        "count":      len(points),
    })
}
```

**Add routes in tracking-service main.go:**
```go
// Device telemetry endpoints (API key auth)
device := r.Group("/api/v1/device")
device.Use(deviceAuthMiddleware())
{
    device.POST("/telemetry/gps", telemetryHandler.IngestGPS)
    device.POST("/telemetry/obd", telemetryHandler.IngestOBD)
}

// Web admin endpoints (JWT auth)
api := r.Group("/api/v1")
api.Use(jwtAuthMiddleware())
{
    api.GET("/vehicles/:id/history", telemetryHandler.GetVehicleHistory)
}
```

**Validation:**
```bash
# Test GPS ingestion
curl -X POST http://localhost:8082/api/v1/device/telemetry/gps \
  -H "Authorization: Bearer DEVICE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "DEV001",
    "timestamp": 1738713600,
    "latitude": 28.6139,
    "longitude": 77.2090,
    "speed": 45.5,
    "heading": 180,
    "accuracy": 10,
    "ignition": true,
    "odometer": 12500.5,
    "battery_level": 85,
    "satellites": 12
  }'

# Verify data in TimescaleDB
psql postgres://fleet:fleet@localhost:5433/fleetdb -c "
SELECT time, vehicle_id, latitude, longitude, speed, ignition 
FROM gps_locations 
ORDER BY time DESC LIMIT 10;
"

# Check NATS message published
nats sub fleet.telemetry.gps
```

**Success Metrics:**
- ✅ GPS endpoints accepting 1000+ req/sec
- ✅ Data appearing in TimescaleDB within 100ms
- ✅ NATS events published for real-time consumers
- ✅ Vehicle current location updated on live map

---

### **Week 6: Geofencing Engine**

#### Task 2.3: Geofence Database Schema
**File:** `database/migrations/006_geofences.sql`

```sql
-- Geofences table (from doc.txn lines 220-245)
CREATE TABLE IF NOT EXISTS geofences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'circle', 'polygon', 'route_corridor'
    geometry GEOGRAPHY(GEOMETRY, 4326) NOT NULL,
    radius_meters INTEGER, -- for circle type
    metadata JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_geofences_org ON geofences(org_id) WHERE is_active = true;
CREATE INDEX idx_geofences_geometry ON geofences USING GIST(geometry);

-- Geofence assignments (which vehicles should trigger which geofences)
CREATE TABLE IF NOT EXISTS geofence_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    geofence_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
    vehicle_id UUID REFERENCES vehicles(id) ON DELETE CASCADE,
    vehicle_group_id UUID, -- for bulk assignments
    assignment_type VARCHAR(20) NOT NULL, -- 'enter', 'exit', 'dwell'
    dwell_minutes INTEGER, -- for dwell type
    notification_settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_assignments_geofence ON geofence_assignments(geofence_id) WHERE is_active = true;
CREATE INDEX idx_assignments_vehicle ON geofence_assignments(vehicle_id) WHERE is_active = true;

-- Geofence events log
CREATE TABLE IF NOT EXISTS geofence_events (
    time TIMESTAMPTZ NOT NULL,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    geofence_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
    event_type VARCHAR(20) NOT NULL, -- 'enter', 'exit', 'dwell'
    location GEOGRAPHY(POINT, 4326),
    duration_minutes INTEGER, -- for exit/dwell events
    metadata JSONB DEFAULT '{}'
);

-- TimescaleDB hypertable for geofence events
SELECT create_hypertable('geofence_events', 'time',
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

CREATE INDEX idx_geofence_events_vehicle ON geofence_events(vehicle_id, time DESC);
CREATE INDEX idx_geofence_events_geofence ON geofence_events(geofence_id, time DESC);

GRANT ALL ON geofences TO fleet;
GRANT ALL ON geofence_assignments TO fleet;
GRANT SELECT, INSERT ON geofence_events TO fleet;
```

**Validation:**
```bash
psql postgres://fleet:fleet@localhost:5433/fleetdb -f database/migrations/006_geofences.sql

# Test geofence creation (circular)
psql postgres://fleet:fleet@localhost:5433/fleetdb -c "
INSERT INTO geofences (org_id, name, type, geometry, radius_meters)
VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,
    'Warehouse Delhi',
    'circle',
    ST_SetSRID(ST_MakePoint(77.2090, 28.6139), 4326)::geography,
    500
);
"

# Test geofence query (check if point is inside)
psql postgres://fleet:fleet@localhost:5433/fleetdb -c "
SELECT id, name, 
    ST_Distance(
        geometry,
        ST_SetSRID(ST_MakePoint(77.2100, 28.6140), 4326)::geography
    ) AS distance_meters
FROM geofences
WHERE ST_DWithin(
    geometry,
    ST_SetSRID(ST_MakePoint(77.2100, 28.6140), 4326)::geography,
    COALESCE(radius_meters, 0)
);
"
```

---

#### Task 2.4: Geofencing Service Implementation
**File:** `services/geofencing-service/internal/engine/geofence_engine.go`

```go
package engine

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/nats-io/nats.go"
    "go.uber.org/zap"
)

type GeofenceEngine struct {
    db     *pgxpool.Pool
    nc     *nats.Conn
    logger *zap.Logger
    
    // Cache of vehicle last known geofence state
    vehicleStates map[uuid.UUID]*VehicleState
}

type VehicleState struct {
    VehicleID      uuid.UUID
    InsideGeofences map[uuid.UUID]time.Time // geofence_id -> entry_time
    LastChecked    time.Time
}

type GPSEvent struct {
    VehicleID string  `json:"vehicle_id"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Timestamp int64   `json:"timestamp"`
    Ignition  bool    `json:"ignition"`
}

func NewGeofenceEngine(db *pgxpool.Pool, nc *nats.Conn, logger *zap.Logger) *GeofenceEngine {
    return &GeofenceEngine{
        db:            db,
        nc:            nc,
        logger:        logger,
        vehicleStates: make(map[uuid.UUID]*VehicleState),
    }
}

// Start - Subscribe to GPS events and process geofence checks
func (ge *GeofenceEngine) Start(ctx context.Context) error {
    ge.logger.Info("Starting geofence engine...")

    // Subscribe to GPS telemetry stream
    sub, err := ge.nc.Subscribe("fleet.telemetry.gps", func(msg *nats.Msg) {
        var event GPSEvent
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            ge.logger.Error("Failed to parse GPS event", zap.Error(err))
            return
        }

        vehicleID, err := uuid.Parse(event.VehicleID)
        if err != nil {
            return
        }

        // Process geofence checks for this GPS point
        if err := ge.processGPSEvent(ctx, vehicleID, event); err != nil {
            ge.logger.Error("Failed to process geofence event",
                zap.String("vehicle_id", event.VehicleID),
                zap.Error(err))
        }
    })

    if err != nil {
        return fmt.Errorf("failed to subscribe to GPS events: %w", err)
    }

    ge.logger.Info("Geofence engine started", zap.String("subject", "fleet.telemetry.gps"))

    // Wait for context cancellation
    <-ctx.Done()
    sub.Unsubscribe()
    return nil
}

func (ge *GeofenceEngine) processGPSEvent(ctx context.Context, vehicleID uuid.UUID, event GPSEvent) error {
    // Query all active geofences assigned to this vehicle
    rows, err := ge.db.Query(ctx, `
        SELECT g.id, g.name, g.type, g.geometry, g.radius_meters,
               ga.assignment_type, ga.dwell_minutes
        FROM geofences g
        JOIN geofence_assignments ga ON ga.geofence_id = g.id
        WHERE ga.vehicle_id = $1 
          AND ga.is_active = true 
          AND g.is_active = true
    `, vehicleID)

    if err != nil {
        return err
    }
    defer rows.Close()

    point := fmt.Sprintf("ST_SetSRID(ST_MakePoint(%f, %f), 4326)::geography", event.Longitude, event.Latitude)
    
    // Get or create vehicle state
    state, exists := ge.vehicleStates[vehicleID]
    if !exists {
        state = &VehicleState{
            VehicleID:       vehicleID,
            InsideGeofences: make(map[uuid.UUID]time.Time),
            LastChecked:     time.Unix(event.Timestamp, 0),
        }
        ge.vehicleStates[vehicleID] = state
    }

    currentlyInside := make(map[uuid.UUID]bool)

    for rows.Next() {
        var (
            geofenceID     uuid.UUID
            name           string
            geoType        string
            geometry       string
            radiusMeters   *int
            assignmentType string
            dwellMinutes   *int
        )

        if err := rows.Scan(&geofenceID, &name, &geoType, &geometry, &radiusMeters, &assignmentType, &dwellMinutes); err != nil {
            continue
        }

        // Check if vehicle is inside this geofence using PostGIS
        var isInside bool
        query := fmt.Sprintf(`
            SELECT ST_DWithin(
                '%s'::geography,
                %s,
                COALESCE($1, 0)
            )
        `, geometry, point)

        err = ge.db.QueryRow(ctx, query, radiusMeters).Scan(&isInside)
        if err != nil {
            continue
        }

        if isInside {
            currentlyInside[geofenceID] = true

            // Check for ENTER event
            if _, wasInside := state.InsideGeofences[geofenceID]; !wasInside {
                if assignmentType == "enter" || assignmentType == "both" {
                    ge.triggerGeofenceEvent(ctx, vehicleID, geofenceID, "enter", event)
                }
                state.InsideGeofences[geofenceID] = time.Unix(event.Timestamp, 0)
            } else {
                // Check for DWELL event
                if assignmentType == "dwell" && dwellMinutes != nil {
                    entryTime := state.InsideGeofences[geofenceID]
                    duration := time.Unix(event.Timestamp, 0).Sub(entryTime)
                    if duration >= time.Duration(*dwellMinutes)*time.Minute {
                        ge.triggerGeofenceEvent(ctx, vehicleID, geofenceID, "dwell", event)
                        // Reset entry time to avoid duplicate dwell alerts
                        state.InsideGeofences[geofenceID] = time.Unix(event.Timestamp, 0)
                    }
                }
            }
        }
    }

    // Check for EXIT events
    for geofenceID, entryTime := range state.InsideGeofences {
        if !currentlyInside[geofenceID] {
            duration := time.Unix(event.Timestamp, 0).Sub(entryTime)
            ge.triggerGeofenceEvent(ctx, vehicleID, geofenceID, "exit", event)
            delete(state.InsideGeofences, geofenceID)
            
            // Log to TimescaleDB
            _, _ = ge.db.Exec(ctx, `
                INSERT INTO geofence_events (time, vehicle_id, geofence_id, event_type, location, duration_minutes)
                VALUES ($1, $2, $3, 'exit', ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography, $6)
            `, time.Unix(event.Timestamp, 0), vehicleID, geofenceID, event.Longitude, event.Latitude, int(duration.Minutes()))
        }
    }

    state.LastChecked = time.Unix(event.Timestamp, 0)
    return nil
}

func (ge *GeofenceEngine) triggerGeofenceEvent(ctx context.Context, vehicleID, geofenceID uuid.UUID, eventType string, event GPSEvent) {
    // Log to TimescaleDB
    _, err := ge.db.Exec(ctx, `
        INSERT INTO geofence_events (time, vehicle_id, geofence_id, event_type, location)
        VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography)
    `, time.Unix(event.Timestamp, 0), vehicleID, geofenceID, eventType, event.Longitude, event.Latitude)

    if err != nil {
        ge.logger.Error("Failed to log geofence event", zap.Error(err))
    }

    // Publish to NATS for alert service
    eventData, _ := json.Marshal(map[string]interface{}{
        "vehicle_id":   vehicleID.String(),
        "geofence_id":  geofenceID.String(),
        "event_type":   eventType,
        "latitude":     event.Latitude,
        "longitude":    event.Longitude,
        "timestamp":    event.Timestamp,
    })

    if err := ge.nc.Publish("fleet.geofence.event", eventData); err != nil {
        ge.logger.Error("Failed to publish geofence event", zap.Error(err))
    }

    ge.logger.Info("Geofence event triggered",
        zap.String("vehicle_id", vehicleID.String()),
        zap.String("geofence_id", geofenceID.String()),
        zap.String("type", eventType))
}
```

**Success Metrics:**
- ✅ Real-time geofence detection (<5 second latency)
- ✅ Enter/exit/dwell events logged to TimescaleDB
- ✅ NATS events published for alert service
- ✅ No false positives with proper radius handling

---

## 🔥 PHASE 3: ANDROID DRIVER APP

### **Week 7-8: Mobile App Development**

#### Task 3.1: Android Project Structure
**Directory:** `android/driver-app/`

```bash
cd /home/abhishek/Desktop/flyfleet
mkdir -p android/driver-app
cd android/driver-app

# Create new Android project with Kotlin + Jetpack Compose
```

**File:** `android/driver-app/build.gradle.kts`

```kotlin
plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
    id("kotlin-kapt")
    id("com.google.dagger.hilt.android")
}

android {
    namespace = "com.flyfleet.driver"
    compileSdk = 34

    defaultConfig {
        applicationId = "com.flyfleet.driver"
        minSdk = 24
        targetSdk = 34
        versionCode = 1
        versionName = "1.0.0"

        buildConfigField("String", "API_BASE_URL", "\"https://api.flyfleet.in\"")
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    composeOptions {
        kotlinCompilerExtensionVersion = "1.5.3"
    }
}

dependencies {
    // Compose
    implementation(platform("androidx.compose:compose-bom:2024.01.00"))
    implementation("androidx.compose.ui:ui")
    implementation("androidx.compose.material3:material3")
    implementation("androidx.compose.ui:ui-tooling-preview")
    implementation("androidx.activity:activity-compose:1.8.2")
    implementation("androidx.navigation:navigation-compose:2.7.6")

    // Lifecycle
    implementation("androidx.lifecycle:lifecycle-runtime-ktx:2.7.0")
    implementation("androidx.lifecycle:lifecycle-viewmodel-compose:2.7.0")

    // Location Services
    implementation("com.google.android.gms:play-services-location:21.1.0")

    // Networking
    implementation("com.squareup.retrofit2:retrofit:2.9.0")
    implementation("com.squareup.retrofit2:converter-gson:2.9.0")
    implementation("com.squareup.okhttp3:okhttp:4.12.0")
    implementation("com.squareup.okhttp3:logging-interceptor:4.12.0")

    // Room Database (offline sync)
    implementation("androidx.room:room-runtime:2.6.1")
    implementation("androidx.room:room-ktx:2.6.1")
    kapt("androidx.room:room-compiler:2.6.1")

    // Dependency Injection
    implementation("com.google.dagger:hilt-android:2.48")
    kapt("com.google.dagger:hilt-compiler:2.48")

    // QR Code Scanner
    implementation("com.google.mlkit:barcode-scanning:17.2.0")
    implementation("androidx.camera:camera-camera2:1.3.1")
    implementation("androidx.camera:camera-lifecycle:1.3.1")
    implementation("androidx.camera:camera-view:1.3.1")

    // WorkManager (background sync)
    implementation("androidx.work:work-runtime-ktx:2.9.0")
}
```

---

#### Task 3.2: Foreground Location Tracking Service
**File:** `android/driver-app/app/src/main/java/com/flyfleet/driver/service/LocationTrackingService.kt`

```kotlin
package com.flyfleet.driver.service

import android.Manifest
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Intent
import android.content.pm.PackageManager
import android.location.Location
import android.os.Build
import android.os.IBinder
import android.os.Looper
import androidx.core.app.ActivityCompat
import androidx.core.app.NotificationCompat
import com.flyfleet.driver.R
import com.flyfleet.driver.data.local.dao.GPSQueueDao
import com.flyfleet.driver.data.local.entity.GPSQueueItem
import com.flyfleet.driver.data.remote.ApiService
import com.google.android.gms.location.*
import dagger.hilt.android.AndroidEntryPoint
import kotlinx.coroutines.*
import javax.inject.Inject

@AndroidEntryPoint
class LocationTrackingService : Service() {

    @Inject
    lateinit var apiService: ApiService

    @Inject
    lateinit var gpsQueueDao: GPSQueueDao

    private lateinit var fusedLocationClient: FusedLocationProviderClient
    private lateinit var locationCallback: LocationCallback
    private val serviceScope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    companion object {
        const val CHANNEL_ID = "location_tracking_channel"
        const val NOTIFICATION_ID = 1001
        const val ACTION_START_TRACKING = "START_TRACKING"
        const val ACTION_STOP_TRACKING = "STOP_TRACKING"
    }

    override fun onCreate() {
        super.onCreate()
        fusedLocationClient = LocationServices.getFusedLocationProviderClient(this)
        createNotificationChannel()
        startForeground(NOTIFICATION_ID, createNotification())
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_START_TRACKING -> startLocationUpdates()
            ACTION_STOP_TRACKING -> stopSelf()
        }
        return START_STICKY
    }

    private fun startLocationUpdates() {
        val locationRequest = LocationRequest.Builder(
            Priority.PRIORITY_HIGH_ACCURACY,
            10000L // 10 seconds
        ).apply {
            setMinUpdateIntervalMillis(5000L) // 5 seconds fastest
            setMaxUpdateDelayMillis(15000L) // 15 seconds max delay
        }.build()

        locationCallback = object : LocationCallback() {
            override fun onLocationResult(result: LocationResult) {
                result.lastLocation?.let { location ->
                    handleNewLocation(location)
                }
            }
        }

        if (ActivityCompat.checkSelfPermission(this, Manifest.permission.ACCESS_FINE_LOCATION) 
            == PackageManager.PERMISSION_GRANTED) {
            fusedLocationClient.requestLocationUpdates(
                locationRequest,
                locationCallback,
                Looper.getMainLooper()
            )
        }
    }

    private fun handleNewLocation(location: Location) {
        serviceScope.launch {
            val gpsItem = GPSQueueItem(
                deviceId = getDeviceId(),
                timestamp = System.currentTimeMillis() / 1000,
                latitude = location.latitude,
                longitude = location.longitude,
                altitude = if (location.hasAltitude()) location.altitude else null,
                speed = if (location.hasSpeed()) location.speed.toDouble() * 3.6 else 0.0, // m/s to km/h
                heading = if (location.hasBearing()) location.bearing.toDouble() else 0.0,
                accuracy = if (location.hasAccuracy()) location.accuracy.toDouble() else 0.0,
                satellites = null, // Not available from FusedLocation
                isSynced = false
            )

            // Save to local Room database
            gpsQueueDao.insert(gpsItem)

            // Try to upload immediately if online
            tryUploadGPSData()
        }
    }

    private suspend fun tryUploadGPSData() {
        val pendingItems = gpsQueueDao.getUnsyncedItems(limit = 50)
        
        if (pendingItems.isEmpty()) return

        try {
            // Upload batch
            val response = apiService.uploadGPSBatch(pendingItems.map { item ->
                mapOf(
                    "device_id" to item.deviceId,
                    "timestamp" to item.timestamp,
                    "latitude" to item.latitude,
                    "longitude" to item.longitude,
                    "altitude" to item.altitude,
                    "speed" to item.speed,
                    "heading" to item.heading,
                    "accuracy" to item.accuracy,
                    "ignition" to true // TODO: Get from OBD/vehicle state
                )
            })

            if (response.isSuccessful) {
                // Mark as synced
                gpsQueueDao.markAsSynced(pendingItems.map { it.id })
                
                // Delete old synced items (keep last 24 hours)
                val cutoffTime = System.currentTimeMillis() / 1000 - (24 * 3600)
                gpsQueueDao.deleteOldSyncedItems(cutoffTime)
            }
        } catch (e: Exception) {
            // Will retry on next location update
            e.printStackTrace()
        }
    }

    private fun getDeviceId(): String {
        val prefs = getSharedPreferences("flyfleet", MODE_PRIVATE)
        return prefs.getString("device_id", "UNKNOWN") ?: "UNKNOWN"
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "Location Tracking",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "FlyFleet is tracking your location"
            }
            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }

    private fun createNotification(): Notification {
        val stopIntent = Intent(this, LocationTrackingService::class.java).apply {
            action = ACTION_STOP_TRACKING
        }
        val stopPendingIntent = PendingIntent.getService(
            this, 0, stopIntent, PendingIntent.FLAG_IMMUTABLE
        )

        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("FlyFleet Tracking Active")
            .setContentText("Your location is being tracked")
            .setSmallIcon(R.drawable.ic_location)
            .setOngoing(true)
            .addAction(R.drawable.ic_stop, "Stop", stopPendingIntent)
            .build()
    }

    override fun onDestroy() {
        super.onDestroy()
        fusedLocationClient.removeLocationUpdates(locationCallback)
        serviceScope.cancel()
    }

    override fun onBind(intent: Intent?): IBinder? = null
}
```

**Success Metrics:**
- ✅ Foreground service running with persistent notification
- ✅ GPS updates every 10 seconds (5s fastest)
- ✅ Offline queue with Room database
- ✅ Auto-sync when online with batch upload
- ✅ Battery optimized with smart intervals

---

## 🔥 PHASE 4: PRODUCTION DEPLOYMENT

### **Week 9: Docker Production Stack**

#### Task 4.1: Production Docker Compose
**File:** `/home/abhishek/Desktop/flyfleet/docker-compose.prod.yml`

This is extracted from doc.txn lines 13825-14053 with complete production configuration including:
- Traefik reverse proxy with automatic HTTPS
- PostgreSQL with TimescaleDB
- Redis with persistence
- NATS JetStream
- Centrifugo WebSocket
- All 6 microservices
- Prometheus + Grafana monitoring
- Automated backups

**(File too large, will create in separate step)**

---

## 📊 Implementation Priority Matrix

| Week | Task | Priority | Status |
|------|------|----------|--------|
| 5 | TimescaleDB Setup | 🔴 CRITICAL | ⏳ Pending |
| 5 | GPS Telemetry API | 🔴 CRITICAL | ⏳ Pending |
| 6 | Geofencing Engine | 🟠 HIGH | ⏳ Pending |
| 6 | Geofence UI (Web Admin) | 🟠 HIGH | ⏳ Pending |
| 7 | Android Location Service | 🟠 HIGH | ⏳ Pending |
| 7 | Device Enrollment Flow | 🟠 HIGH | ⏳ Pending |
| 8 | Offline Sync + Room DB | 🟡 MEDIUM | ⏳ Pending |
| 8 | OBD Integration | 🟡 MEDIUM | ⏳ Pending |
| 9 | Production Docker Stack | 🔴 CRITICAL | ⏳ Pending |
| 9 | Monitoring Setup | 🟠 HIGH | ⏳ Pending |

---

## 🚀 Quick Start - Next Steps

```bash
# Step 1: Setup TimescaleDB
cd /home/abhishek/Desktop/flyfleet
psql postgres://fleet:fleet@localhost:5433/fleetdb -f database/migrations/005_timescaledb_setup.sql

# Step 2: Update tracking-service with telemetry handlers
# (Add TelemetryHandler from Task 2.2)

# Step 3: Test GPS ingestion
curl -X POST http://localhost:8082/api/v1/device/telemetry/gps \
  -H "Content-Type: application/json" \
  -d '{"device_id":"TEST001","timestamp":1738713600,"latitude":28.6139,"longitude":77.2090,"speed":45,"ignition":true}'

# Step 4: Setup geofencing
psql postgres://fleet:fleet@localhost:5433/fleetdb -f database/migrations/006_geofences.sql

# Step 5: Create geofencing-service
mkdir -p services/geofencing-service
# (Implement GeofenceEngine from Task 2.4)
```

---

## 📚 Architecture Reference (doc.txn)

- **Lines 1-500**: Technology stack, microservices overview
- **Lines 111-327**: Complete database schemas
- **Lines 328-1028**: Full API specifications
- **Lines 1029-1256**: Android architecture
- **Lines 1460-2302**: UI screen specs (React)
- **Lines 2746-3145**: Phased implementation
- **Lines 3200-5000**: Deployment scripts
- **Lines 3866-4555**: Android source code
- **Lines 5000-8000**: Go API implementation
- **Lines 13000-14498**: Production infrastructure (Docker, monitoring, runbooks)

---

## ✅ Next Action Required

**Which implementation should I start with?**

1. **TimescaleDB + Telemetry Pipeline** (Week 5) - Foundation for all tracking
2. **Geofencing Engine** (Week 6) - High-value feature for alerts
3. **Android Driver App** (Week 7-8) - Mobile data collection
4. **Production Docker Stack** (Week 9) - Deployment infrastructure

Let me know which priority you want to tackle first, and I'll implement it step-by-step!
