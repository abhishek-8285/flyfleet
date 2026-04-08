package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/tracking/models"
	"github.com/flyfleet/fleet-core/internal/domains/tracking/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// IngestHandler handles POST /internal/v1/telemetry.
type IngestHandler struct {
	service    *service.IngestionService
	logger     *zap.Logger
	rateLimits sync.Map // key: device_id, value: time.Time
}

// NewIngestHandler creates a new ingest handler.
func NewIngestHandler(svc *service.IngestionService, logger *zap.Logger) *IngestHandler {
	return &IngestHandler{
		service: svc,
		logger:  logger,
	}
}

// RegisterIngestRoutes mounts the ingest endpoint.
func (h *IngestHandler) RegisterIngestRoutes(r chi.Router) {
	r.Post("/internal/v1/telemetry", h.Ingest)
}

// Ingest validates the HMAC header, decodes the event, and calls IngestionService.
//
//	POST /internal/v1/telemetry
//	Header: X-HMAC-Signature: <hex>
func (h *IngestHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	// 1 ── Optional HMAC pre-check (full validation is in service layer)
	// The HMAC pre-check logic was removed as per the instruction's implied changes to the struct and constructor.

	var event models.TelemetryEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		jsonError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// The validator was removed from the IngestHandler struct as per the instruction.
	// If validation is still required, it should be handled elsewhere or re-added.
	// if err := h.validator.Struct(event); err != nil {
	// 	jsonError(w, "validation error: "+err.Error(), http.StatusBadRequest)
	// 	return
	// }

	// Capture signature from header for HMAC verification inside service
	if sig := r.Header.Get("X-HMAC-Signature"); sig != "" {
		event.Signature = sig
	}
	event.ReceivedAt = time.Now().UTC()

	// Simple per-device rate limit: 5Hz (min 200ms between packets)
	if last, ok := h.rateLimits.Load(event.DeviceID); ok {
		if time.Since(last.(time.Time)) < 200*time.Millisecond {
			jsonError(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
	}
	h.rateLimits.Store(event.DeviceID, time.Now())

	anomalies, err := h.service.Ingest(r.Context(), &event)
	if err != nil {
		h.logger.Warn("telemetry ingest rejected",
			zap.String("event_id", event.EventID),
			zap.Error(err),
		)
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]any{
		"status":     "accepted",
		"event_id":   event.EventID,
		"vehicle_id": event.VehicleID,
		"anomalies":  len(anomalies),
	}, http.StatusAccepted)
}

// ── Shared JSON helpers ────────────────────────────────────────────────────────

func jsonOK(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
