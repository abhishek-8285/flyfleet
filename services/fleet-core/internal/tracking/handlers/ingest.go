package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/tracking/models"
	telemetryv1 "github.com/flyfleet/proto/telemetry/v1"
)

// IngestHandler handles POST /api/v1/telemetry
type IngestHandler struct {
	ingestion ingestService
	logger    *zap.Logger
}

// ingestService is the interface satisfied by *service.IngestionService.
type ingestService interface {
	Ingest(ctx context.Context, event *models.TelemetryEvent) error
}

func NewIngestHandler(ingestion ingestService, logger *zap.Logger) *IngestHandler {
	return &IngestHandler{ingestion: ingestion, logger: logger}
}

type ingestRequest struct {
	VehicleID  string  `json:"vehicle_id"`
	DeviceID   string  `json:"device_id"`
	Latitude   float64 `json:"lat"`
	Longitude  float64 `json:"lng"`
	Speed      float64 `json:"speed"`
	Heading    float64 `json:"heading"`
	Ignition   bool    `json:"ignition"`
	FuelLevel  float64 `json:"fuel_level"`
	Odometer   float64 `json:"odometer"`
	SequenceNo int64   `json:"sequence_no"`
	RecordedAt string  `json:"recorded_at"`
	RawData    string  `json:"raw_data"`
	EventID    string  `json:"event_id"`
	TenantID   string  `json:"tenant_id"`
}

func (h *IngestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event *models.TelemetryEvent
	contentType := r.Header.Get("Content-Type")
	h.logger.Debug("ingest handler: processing request", zap.String("content_type", contentType))

	isProtobuf := strings.HasPrefix(contentType, "application/x-protobuf") || 
		strings.HasPrefix(contentType, "application/octet-stream")

	if isProtobuf {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
			return
		}
		var payload telemetryv1.TelemetryPayload
		if err := proto.Unmarshal(bodyBytes, &payload); err != nil {
			http.Error(w, `{"error":"invalid protobuf payload"}`, http.StatusBadRequest)
			return
		}

		eventIDStr := r.Header.Get("X-Event-ID")
		var eventID uuid.UUID
		if eventIDStr != "" {
			eventID, _ = uuid.FromString(eventIDStr)
		}
		if eventID == uuid.Nil {
			eventID = uuid.Must(uuid.NewV4())
		}

		// Speed conversion: SpeedMph to km/h
		speedKmh := float64(payload.SpeedMph * 1.60934)
		// Odometer conversion: TotalOdometerMeters to km
		odometerKm := float64(payload.TotalOdometerMeters / 1000.0)

		// Create raw data JSON for telemetry active DTCs and engine data
		type dTC struct {
			SPN   string `json:"spn"`
			FMI   string `json:"fmi"`
			Count int32  `json:"count"`
		}
		dtcs := make([]dTC, 0, len(payload.ActiveDtcs))
		for _, d := range payload.ActiveDtcs {
			dtcs = append(dtcs, dTC{
				SPN:   d.Spn,
				FMI:   d.Fmi,
				Count: d.Count,
			})
		}

		rawMap := map[string]any{
			"hdop":                 payload.Hdop,
			"engine_rpm":           payload.EngineRpm,
			"engine_load_percent":  payload.EngineLoadPercent,
			"coolant_temp_celsius": payload.CoolantTempCelsius,
			"battery_voltage":      payload.BatteryVoltage,
			"cellular_signal_dbm":  payload.CellularSignalDbm,
			"active_dtcs":          dtcs,
		}
		rawDataBytes, _ := json.Marshal(rawMap)

		event = &models.TelemetryEvent{
			EventID:    eventID,
			DeviceID:   payload.DeviceId,
			Lat:        payload.Latitude,
			Lng:        payload.Longitude,
			Speed:      speedKmh,
			Heading:    float64(payload.HeadingDegrees),
			Ignition:   payload.IgnitionStatus,
			FuelLevel:  float64(payload.FuelLevelPercent),
			Odometer:   odometerKm,
			SequenceNo: 0,
			RawData:    string(rawDataBytes),
			Signature:  r.Header.Get("X-HMAC-Signature"),
			RecordedAt: time.UnixMilli(payload.TimestampMs).UTC(),
			ReceivedAt: time.Now().UTC(),
		}
	} else {
		var req ingestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}

		vehicleID, err := uuid.FromString(req.VehicleID)
		if err != nil {
			http.Error(w, `{"error":"invalid vehicle_id"}`, http.StatusBadRequest)
			return
		}

		eventID, _ := uuid.FromString(req.EventID)
		if eventID == uuid.Nil {
			eventID = uuid.Must(uuid.NewV4())
		}

		// Resolve TenantID: Try context first (User), fallback to body (Hardware)
		middlewareTenantID := middleware.GetTenantID(r.Context())
		tenantID, _ := uuid.FromString(middlewareTenantID.String())
		if tenantID == uuid.Nil && req.TenantID != "" {
			tenantID, _ = uuid.FromString(req.TenantID)
		}

		event = &models.TelemetryEvent{
			EventID:    eventID,
			VehicleID:  vehicleID,
			DeviceID:   req.DeviceID,
			TenantID:   tenantID,
			Lat:        req.Latitude,
			Lng:        req.Longitude,
			Speed:      req.Speed,
			Heading:    req.Heading,
			Ignition:   req.Ignition,
			FuelLevel:  req.FuelLevel,
			Odometer:   req.Odometer,
			SequenceNo: req.SequenceNo,
			RawData:    req.RawData,
			Signature:  r.Header.Get("X-HMAC-Signature"),
		}

		if req.RecordedAt != "" {
			if t, err := time.Parse(time.RFC3339, req.RecordedAt); err == nil {
				event.RecordedAt = t
			}
		}
	}

	if event.RecordedAt.IsZero() {
		event.RecordedAt = time.Now().UTC()
	}

	if err := h.ingestion.Ingest(r.Context(), event); err != nil {
		h.logger.Error("telemetry ingest failed", zap.Error(err))
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	h.logger.Info("telemetry event accepted", 
		zap.String("event_id", event.EventID.String()),
		zap.String("device_id", event.DeviceID))

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
}
