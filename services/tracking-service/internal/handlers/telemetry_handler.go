package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fleet-platform/shared/pkg/database"
	"github.com/fleet-platform/shared/pkg/messaging"
	"github.com/fleet-platform/shared/pkg/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GPSPayload from device
type GPSPayload struct {
	DeviceID   string          `json:"device_id" validate:"required"`
	Timestamp  int64           `json:"timestamp" validate:"required"`
	Latitude   float64         `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude  float64         `json:"longitude" validate:"required,min=-180,max=180"`
	Altitude   *float64        `json:"altitude"`
	Speed      float64         `json:"speed"`
	Heading    float64         `json:"heading"`
	Accuracy   float64         `json:"accuracy"`
	Ignition   bool            `json:"ignition"`
	Odometer   *float64        `json:"odometer"`
	Battery    *int            `json:"battery_level"`
	Satellites *int            `json:"satellites"`
	HDOP       *float64        `json:"hdop"`
	RawData    json.RawMessage `json:"raw_data,omitempty"`
}

type TelemetryHandler struct {
	db        database.Database
	mb        messaging.MessageBroker
	logger    *zap.Logger
	validator *validator.Validate
}

func NewTelemetryHandler(db database.Database, mb messaging.MessageBroker, logger *zap.Logger) *TelemetryHandler {
	return &TelemetryHandler{
		db:        db,
		mb:        mb,
		logger:    logger,
		validator: validator.New(),
	}
}

func (h *TelemetryHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/device/telemetry", func(r chi.Router) {
		r.Post("/gps", h.IngestGPS)
		r.Post("/obd", h.IngestOBD)
	})
	r.Get("/api/v1/vehicles/{id}/history", h.GetVehicleHistory)

	// Legacy endpoints for backward compatibility
	r.Post("/telemetry/gps", h.IngestGPS)
	r.Post("/telemetry/obd", h.IngestOBD)
}

// Helper methods for JSON responses
func (h *TelemetryHandler) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode json response", zap.Error(err))
	}
}

func (h *TelemetryHandler) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}

// IngestGPS - High-throughput GPS ingestion endpoint (JetStream)
// Legacy in-process storage removed. Now persistent via NATS JetStream.
func (h *TelemetryHandler) IngestGPS(w http.ResponseWriter, r *http.Request) {
	var payload GPSPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(payload); err != nil {
		h.errorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	var vehicleID, orgID uuid.UUID
	row := h.db.QueryRow(r.Context(),
		`SELECT vehicle_id, org_id FROM devices WHERE device_id = $1 AND status = 'active'`,
		payload.DeviceID,
	)
	err := row.Scan(&vehicleID, &orgID)

	if err != nil {
		h.logger.Error("Device not found or inactive", zap.String("device_id", payload.DeviceID), zap.Error(err))
		h.errorJSON(w, "Device not registered", http.StatusNotFound)
		return
	}

	timestamp := time.Unix(payload.Timestamp, 0)

	// Pillar 4: Persistent Ingestion via JetStream
	// Storage handled by dedicated ingestion-service
	eventData := struct {
		Type      string             `json:"type"`
		Timestamp string             `json:"timestamp"`
		OrgID     string             `json:"org_id"`
		VehicleID string             `json:"vehicle_id"`
		Data      models.GPSTelemetry `json:"data"`
	}{
		Type:      "gps",
		Timestamp: timestamp.Format(time.RFC3339),
		OrgID:     orgID.String(),
		VehicleID: vehicleID.String(),
		Data: models.GPSTelemetry{
			Time:       timestamp,
			VehicleID:  vehicleID,
			OrgID:      orgID,
			Latitude:   payload.Latitude,
			Longitude:  payload.Longitude,
			Altitude:   payload.Altitude,
			Speed:      &payload.Speed,
			Heading:    &payload.Heading,
			Accuracy:   &payload.Accuracy,
			Ignition:   &payload.Ignition,
			Odometer:   payload.Odometer,
			Battery:    payload.Battery,
			Satellites: payload.Satellites,
			RawData:    nil, // Map json.RawMessage if needed
		},
	}

	if err := h.mb.PublishJS("telemetry.gps", eventData); err != nil {
		h.logger.Error("Failed to ingest GPS data to JetStream", zap.Error(err))
	}

	// Immediate response to device
	h.json(w, map[string]interface{}{
		"status":     "accepted",
		"vehicle_id": vehicleID.String(),
		"timestamp":  timestamp.Unix(),
	}, http.StatusAccepted)
}

// IngestOBD - OBD-II data ingestion (JetStream)
func (h *TelemetryHandler) IngestOBD(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		DeviceID   string          `json:"device_id" validate:"required"`
		Timestamp  int64           `json:"timestamp" validate:"required"`
		EngineRPM  *int            `json:"engine_rpm"`
		FuelLevel  *int            `json:"fuel_level"`
		RawData    json.RawMessage `json:"raw_data,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.errorJSON(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	var vehicleID, orgID uuid.UUID
	row := h.db.QueryRow(r.Context(),
		`SELECT vehicle_id, org_id FROM devices WHERE device_id = $1 AND status = 'active'`,
		payload.DeviceID,
	)
	err := row.Scan(&vehicleID, &orgID)
	if err != nil {
		h.errorJSON(w, "Device not registered", http.StatusNotFound)
		return
	}

	timestamp := time.Unix(payload.Timestamp, 0)

	// Pillar 4: Persistent Ingestion via JetStream
	eventData := struct {
		Type      string             `json:"type"`
		Timestamp string             `json:"timestamp"`
		OrgID     string             `json:"org_id"`
		VehicleID string             `json:"vehicle_id"`
		Data      models.OBDTelemetry `json:"data"`
	}{
		Type:      "obd",
		Timestamp: timestamp.Format(time.RFC3339),
		OrgID:     orgID.String(),
		VehicleID: vehicleID.String(),
		Data: models.OBDTelemetry{
			Time:      timestamp,
			VehicleID: vehicleID,
			OrgID:     orgID,
			RPM:       payload.EngineRPM,
			FuelLevel: nil, // Map correctly from payload if needed
		},
	}

	if err := h.mb.PublishJS("telemetry.obd", eventData); err != nil {
		h.logger.Error("Failed to ingest OBD data to JetStream", zap.Error(err))
	}

	h.json(w, map[string]string{"status": "accepted"}, http.StatusAccepted)
}

// GetVehicleHistory - Query GPS history
func (h *TelemetryHandler) GetVehicleHistory(w http.ResponseWriter, r *http.Request) {
	vehicleID := chi.URLParam(r, "id")
	
	rows, err := h.db.Query(r.Context(), `
        SELECT time, ST_Y(location::geometry) as lat, ST_X(location::geometry) as lng, speed
        FROM gps_locations
        WHERE vehicle_id = $1
        ORDER BY time ASC LIMIT 100`,
		vehicleID,
	)

	if err != nil {
		h.errorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.json(w, rows, http.StatusOK)
}
