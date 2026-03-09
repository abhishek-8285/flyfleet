package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/service"
	"github.com/flyfleet/fleet-core-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EmergencyHandlers struct {
	service        *service.EmergencyService
	analyticsService *service.EmergencyAnalyticsService
	logger         *zap.Logger
	validator      *validator.Validate
}

func NewEmergencyHandlers(s *service.EmergencyService, sa *service.EmergencyAnalyticsService, logger *zap.Logger) *EmergencyHandlers {
	return &EmergencyHandlers{
		service:        s,
		analyticsService: sa,
		logger:         logger,
		validator:      validator.New(),
	}
}

func (h *EmergencyHandlers) RegisterRoutes(r chi.Router) {
	r.Post("/sos-trigger", h.TriggerSOS)
	r.Post("/sos-trigger/ais140", h.TriggerAIS140SOS)
	r.Post("/sos-trigger/simulate", h.TriggerSimulatedSOS)
	r.Get("/incidents", h.ListActiveIncidents)
	r.Get("/incidents/{id}", h.GetIncident)
	r.Post("/incidents/{id}/acknowledge", h.AcknowledgeIncident)
	r.Post("/incidents/{id}/resolve", h.ResolveIncident)
	r.Post("/incidents/{id}/false-alarm", h.MarkAsFalseAlarm)
	r.Post("/incidents/{id}/media", h.UploadMedia)
	r.Post("/incidents/{id}/feedback", h.SubmitDriverFeedback)
	
	// Analytics
	r.Get("/analytics", h.GetIncidentAnalytics)
	r.Get("/analytics/dashboard", h.GetSafetyDashboard)
	
	// Search & Export
	r.Get("/incidents/search", h.SearchIncidents)
	r.Get("/incidents/export", h.ExportIncidents)
	
	// Administrative commands
	r.Group(func(r chi.Router) {
		// r.Use(auth.RequireAdmin) // Placeholder for Admin middleware
		r.Post("/fleet-halt", h.FleetHalt)
		r.Post("/fleet-resume", h.FleetResume)
		r.Post("/contacts", h.CreateContact)
		r.Get("/contacts", h.ListContacts)
	})
}

func (h *EmergencyHandlers) TriggerAIS140SOS(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("TriggerAIS140SOS request received")
	var req struct {
		DeviceID  string  `json:"device_id" validate:"required"`
		VehicleID string  `json:"vehicle_id" validate:"required"`
		Lat       float64 `json:"lat" validate:"required"`
		Lng       float64 `json:"lng" validate:"required"`
		Timestamp int64   `json:"timestamp" validate:"required"`
		PanicFlag bool    `json:"panic_flag" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode AIS140 request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("AIS140 validation failed", zap.Error(err))
		http.Error(w, "400 Invalid AIS140 payload", http.StatusBadRequest)
		return
	}

	// Panic flag must be true for this endpoint
	if !req.PanicFlag {
		h.logger.Warn("AIS140 panic_flag is false")
		http.Error(w, "400 Invalid AIS140 payload: panic_flag must be true", http.StatusBadRequest)
		return
	}

	// Validate coordinates
	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		h.logger.Warn("AIS140 invalid coordinates", zap.Float64("lat", req.Lat), zap.Float64("lng", req.Lng))
		http.Error(w, "400 Invalid AIS140 payload: invalid coordinates", http.StatusBadRequest)
		return
	}

	// Convert to CreateSOSRequest
	orgID := middleware.GetOrgID(r.Context())
	h.logger.Info("Triggering AIS140 SOS", zap.String("org_id", orgID.String()))
	
	createReq := models.CreateSOSRequest{
		DeviceID:     req.DeviceID,
		VehicleID:    req.VehicleID,
		IncidentType: models.TypePanic,
		PanicSource:  models.SourceAIS140,
		Latitude:     req.Lat,
		Longitude:    req.Lng,
		OrgID:        orgID,
	}

	incident, err := h.service.TriggerSOS(r.Context(), createReq)
	if err != nil {
		h.logger.Error("Failed to trigger AIS140 SOS in service", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusAccepted, incident)
}

func (h *EmergencyHandlers) TriggerSOS(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSOSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Panic recovered in TriggerSOS", zap.Any("panic", r))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		if headerID := r.Header.Get("X-Organization-ID"); headerID != "" {
			if id, err := uuid.Parse(headerID); err == nil {
				orgID = id
			}
		}
	}
	req.OrgID = orgID
	incident, err := h.service.TriggerSOS(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to trigger SOS", zap.Error(err))
		http.Error(w, "Failed to trigger SOS", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusCreated, incident)
}

func (h *EmergencyHandlers) ListActiveIncidents(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	incidents, err := h.service.ListActiveIncidents(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to list incidents", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": incidents})
}

func (h *EmergencyHandlers) AcknowledgeIncident(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	h.logger.Info("AcknowledgeIncident request received", zap.String("id", idStr))
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid ID format in acknowledge", zap.String("id", idStr), zap.Error(err))
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var req models.UpdateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode acknowledge request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		if headerID := r.Header.Get("X-Organization-ID"); headerID != "" {
			if id, err := uuid.Parse(headerID); err == nil {
				orgID = id
			}
		}
	}
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		if headerID := r.Header.Get("X-User-ID"); headerID != "" {
			if id, err := uuid.Parse(headerID); err == nil {
				userID = id
			}
		}
	}
	h.logger.Info("Context info for acknowledge", zap.String("org_id", orgID.String()), zap.String("user_id", userID.String()))

	if err := h.service.AcknowledgeIncident(r.Context(), id, orgID, userID, req.Notes); err != nil {
		h.logger.Error("Failed to acknowledge incident in service", zap.String("id", id.String()), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Incident acknowledged"})
}

func (h *EmergencyHandlers) ResolveIncident(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	h.logger.Info("ResolveIncident request received", zap.String("id", idStr))
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid ID format in resolve", zap.String("id", idStr), zap.Error(err))
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var req models.UpdateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode resolve request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		if headerID := r.Header.Get("X-Organization-ID"); headerID != "" {
			if id, err := uuid.Parse(headerID); err == nil {
				orgID = id
			}
		}
	}
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		if headerID := r.Header.Get("X-User-ID"); headerID != "" {
			if id, err := uuid.Parse(headerID); err == nil {
				userID = id
			}
		}
	}
	status := models.SOSStatus(req.Status)
	h.logger.Info("Context info for resolve", zap.String("org_id", orgID.String()), zap.String("user_id", userID.String()), zap.String("status", string(status)))

	if err := h.service.ResolveIncident(r.Context(), id, orgID, userID, req.Notes, status); err != nil {
		h.logger.Error("Failed to resolve incident in service", zap.String("id", id.String()), zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Incident resolved"})
}

func (h *EmergencyHandlers) GetIncident(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	incident, err := h.service.GetIncident(r.Context(), id, orgID)
	if err != nil {
		h.logger.Error("Failed to get incident", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if incident == nil {
		http.Error(w, "Incident not found", http.StatusNotFound)
		return
	}

	writeJSONResponse(w, http.StatusOK, incident)
}

func (h *EmergencyHandlers) MarkAsFalseAlarm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if err := h.service.ResolveIncident(r.Context(), id, orgID, userID, req.Notes, models.StatusFalseAlarm); err != nil {
		h.logger.Error("Failed to mark as false alarm", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Marked as false alarm"})
}

func (h *EmergencyHandlers) FleetHalt(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Reason string `json:"reason" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	orgID := getActorOrgID(r)
	userName := "Admin" // Simplified

	if err := h.service.FleetHalt(r.Context(), orgID, req.Reason, userName); err != nil {
		h.logger.Error("Failed to issue fleet halt", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Fleet-wide halt issued"})
}

func (h *EmergencyHandlers) FleetResume(w http.ResponseWriter, r *http.Request) {
	orgID := getActorOrgID(r)
	userName := "Admin"

	if err := h.service.FleetResume(r.Context(), orgID, userName); err != nil {
		h.logger.Error("Failed to issue fleet resume", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Fleet-wide resume issued"})
}

func (h *EmergencyHandlers) GetSafetyDashboard(w http.ResponseWriter, r *http.Request) {
	orgID := getActorOrgID(r)
	data, err := h.analyticsService.GetSafetyDashboard(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get safety dashboard", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, http.StatusOK, data)
}

func (h *EmergencyHandlers) GetIncidentAnalytics(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	analytics, err := h.analyticsService.GetIncidentAnalytics(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to get incident analytics", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, http.StatusOK, analytics)
}

func (h *EmergencyHandlers) SearchIncidents(w http.ResponseWriter, r *http.Request) {
	// Dummy search for now
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": []string{}, "total": 0})
}

func (h *EmergencyHandlers) ExportIncidents(w http.ResponseWriter, r *http.Request) {
	// Dummy export for now
	format := r.URL.Query().Get("format")
	if format == "" { format = "CSV" }
	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Export " + format + " initiated", "status": "PENDING"})
}

func (h *EmergencyHandlers) CreateContact(w http.ResponseWriter, r *http.Request) {
	// Simplified create contact
	writeJSONResponse(w, http.StatusCreated, map[string]string{"message": "Contact created"})
}

func (h *EmergencyHandlers) ListContacts(w http.ResponseWriter, r *http.Request) {
	orgID := getActorOrgID(r)
	contacts, err := h.service.GetEmergencyContacts(r.Context(), orgID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": contacts})
}

func getActorOrgID(r *http.Request) uuid.UUID {
	return middleware.GetOrgID(r.Context())
}

func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *EmergencyHandlers) UploadMedia(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	incidentID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	var req models.IncidentMedia
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.ID = uuid.New()
	req.IncidentID = incidentID
	req.CreatedAt = time.Now()

	userID := middleware.GetUserID(r.Context())
	if userID != uuid.Nil {
		req.UploadedBy = &userID
	}

	orgID := middleware.GetOrgID(r.Context())
	if err := h.service.AddMedia(r.Context(), req, orgID); err != nil {
		h.logger.Error("Failed to add media to incident", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusCreated, req)
}

func (h *EmergencyHandlers) TriggerSimulatedSOS(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSOSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.SimulationMode = true
	req.OrgID = middleware.GetOrgID(r.Context())
	incident, err := h.service.TriggerSimulatedSOS(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to trigger simulated SOS", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusAccepted, incident)
}

func (h *EmergencyHandlers) SubmitDriverFeedback(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	incidentID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Feedback string `json:"feedback"`
		Notes    string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	if err := h.service.UpdateDriverFeedback(r.Context(), incidentID, orgID, req.Feedback, req.Notes); err != nil {
		h.logger.Error("Failed to update driver feedback", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "feedback_updated"})
}
