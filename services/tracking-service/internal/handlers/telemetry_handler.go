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
        "odometer":   payload.Odometer,
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
