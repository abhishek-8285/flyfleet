package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core/internal/domains/emergency/service"
	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EmergencyHandlers struct {
	service          *service.EmergencyService
	analyticsService *service.EmergencyAnalyticsService
	logger           *zap.Logger
	validator        *validator.Validate
}

func NewEmergencyHandlers(s *service.EmergencyService, sa *service.EmergencyAnalyticsService, logger *zap.Logger) *EmergencyHandlers {
	return &EmergencyHandlers{
		service:          s,
		analyticsService: sa,
		logger:           logger,
		validator:        validator.New(),
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
	r.Get("/incidents/{id}/timeline", h.GetIncidentTimeline)

	// Analytics
	r.Get("/analytics", h.GetIncidentAnalytics)
	r.Get("/analytics/dashboard", h.GetSafetyDashboard)

	// Search & Export
	r.Get("/incidents/search", h.SearchIncidents)
	r.Get("/incidents/export", h.ExportIncidents)

	// Administrative commands
	r.Group(func(r chi.Router) {
		// Require admin or super_admin role for fleet-wide administrative actions
		r.Use(middleware.RequireRole(h.logger, "admin", "super_admin"))
		r.Post("/fleet-halt", h.FleetHalt)
		r.Post("/fleet-resume", h.FleetResume)
		r.Post("/contacts", h.CreateContact)
		r.Get("/contacts", h.ListContacts)
		r.Get("/export", h.Export)
	})

	// Health
	r.Get("/health", h.Health)

	// Phase 7: Drill simulation endpoint
	r.Post("/drill", h.TriggerDrill)
}

func (h *EmergencyHandlers) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	health, err := h.service.CheckHealth(ctx)
	if err != nil {
		h.logger.Error("Health check failed", zap.Error(err))
		http.Error(w, "Failed to perform health check", http.StatusInternalServerError)
		return
	}

	status := "healthy"
	for _, v := range health {
		if v != "healthy" && v != "not_configured" {
			status = "unhealthy"
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    status,
		"details":   health,
		"timestamp": time.Now(),
	})
}

func (h *EmergencyHandlers) Export(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID := middleware.GetOrgID(ctx)
	if orgID == uuid.Nil {
		http.Error(w, "Unauthorized: missing organization context", http.StatusUnauthorized)
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	start := time.Now().AddDate(0, 0, -30) // Default 30 days
	end := time.Now()

	if startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = t
		}
	}
	if endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = t
		}
	}

	incidents, err := h.service.ExportIncidents(ctx, orgID, start, end)
	if err != nil {
		h.logger.Error("Failed to export incidents", zap.Error(err))
		http.Error(w, "Failed to export incidents", http.StatusInternalServerError)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=incidents_export_%s.csv", time.Now().Format("20060102")))

		fmt.Fprintf(w, "ID,VehicleID,Timestamp,Status,Severity,Type,Latitude,Longitude,Address\n")
		for _, inc := range incidents {
			addr := ""
			if inc.LocationAddress != nil {
				addr = *inc.LocationAddress
			}
			fmt.Fprintf(w, "%s,%s,%s,%s,%s,%s,%f,%f,\"%s\"\n",
				inc.ID, inc.VehicleID, inc.Timestamp.Format(time.RFC3339),
				inc.Status, inc.Severity, inc.IncidentType, inc.Latitude, inc.Longitude, addr)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incidents)
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
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: organization context required", http.StatusUnauthorized)
		return
	}
	h.logger.Info("Triggering AIS140 SOS", zap.String("tenant_id", orgID.String()))

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
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: organization context required", http.StatusUnauthorized)
		return
	}
	incident, err := h.service.TriggerSOS(r.Context(), req)
	if err != nil {
		if err == service.ErrDuplicateSOS {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.logger.Error("Failed to trigger SOS", zap.Error(err))
		http.Error(w, "Failed to trigger SOS", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusCreated, incident)
}

func (h *EmergencyHandlers) ListActiveIncidents(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: organization context required", http.StatusUnauthorized)
		return
	}
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
	userID := middleware.GetUserID(r.Context())
	h.logger.Info("Context info for acknowledge", zap.String("tenant_id", orgID.String()), zap.String("user_id", userID.String()))

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
	userID := middleware.GetUserID(r.Context())
	status := models.SOSStatus(req.Status)
	h.logger.Info("Context info for resolve", zap.String("tenant_id", orgID.String()), zap.String("user_id", userID.String()), zap.String("status", string(status)))

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
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: organization context required", http.StatusUnauthorized)
		return
	}
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
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: organization context required", http.StatusUnauthorized)
		return
	}
	userID := middleware.GetUserID(r.Context())
	if userID == uuid.Nil {
		h.logger.Warn("Missing user_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: user context required", http.StatusUnauthorized)
		return
	}
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

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": analytics})
}

func (h *EmergencyHandlers) SearchIncidents(w http.ResponseWriter, r *http.Request) {
	orgID := getActorOrgID(r)

	// Parse query parameters
	statusStr := r.URL.Query().Get("status")
	severityStr := r.URL.Query().Get("severity")
	incidentTypeStr := r.URL.Query().Get("incident_type")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	query := r.URL.Query().Get("query")

	// Build search params
	params := service.IncidentSearchParams{}

	// Parse optional filters
	if statusStr != "" {
		status := models.SOSStatus(statusStr)
		params.Status = &status
	}
	if severityStr != "" {
		severity := models.SOSSeverity(severityStr)
		params.Severity = &severity
	}
	if incidentTypeStr != "" {
		incidentType := models.IncidentType(incidentTypeStr)
		params.IncidentType = &incidentType
	}
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			params.StartDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			params.EndDate = &parsed
		}
	}
	if query != "" {
		params.Query = &query
	}

	// Search incidents
	incidents, err := h.service.SearchIncidents(r.Context(), orgID, params)
	if err != nil {
		h.logger.Error("Failed to search incidents", zap.Error(err))
		http.Error(w, "Failed to search incidents", http.StatusInternalServerError)
		return
	}

	// Return results with total count
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":  incidents,
		"total": len(incidents),
	})
}

func (h *EmergencyHandlers) ExportIncidents(w http.ResponseWriter, r *http.Request) {
	orgID := getActorOrgID(r)

	// Parse date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "CSV"
	}

	startDate, endDate := time.Now().AddDate(0, 0, -30), time.Now()
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	// Fetch incidents
	incidents, err := h.service.ExportIncidents(r.Context(), orgID, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to export incidents", zap.Error(err))
		http.Error(w, "Failed to export incidents", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"incidents_%s.csv\"", startDate.Format("2006-01-02")))

	// Write CSV
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(incidents) // For now, return JSON; can implement proper CSV later
}

func (h *EmergencyHandlers) CreateContact(w http.ResponseWriter, r *http.Request) {
	orgID := getActorOrgID(r)

	var contact models.EmergencyContact
	if err := json.NewDecoder(r.Body).Decode(&contact); err != nil {
		h.logger.Error("Failed to decode create contact request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set tenant_id from context
	contact.OrgID = orgID

	if err := h.service.CreateContact(r.Context(), &contact); err != nil {
		h.logger.Error("Failed to create contact", zap.Error(err))
		http.Error(w, "Failed to create contact", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message":    "Contact created successfully",
		"contact_id": contact.ID,
	})
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

// TriggerDrill handles POST /api/v1/emergency/drill — Phase 7: Drill Simulation.
// It creates a simulated SOS incident but suppresses all real authority notifications.
func (h *EmergencyHandlers) TriggerDrill(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Emergency drill triggered")

	var req models.EmergencyDrillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// SECURITY: Require tenant_id from authenticated context ONLY (prevent spoofing)
	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		http.Error(w, "Unauthorized: organization context required", http.StatusUnauthorized)
		return
	}

	// SECURITY FIX: Do NOT accept tenant_id from request body - force use of authenticated org
	sosReq := models.CreateSOSRequest{
		OrgID:          orgID,
		VehicleID:      req.VehicleID,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		IncidentType:   models.IncidentType(req.IncidentType),
		SimulationMode: true,
	}

	incident, err := h.service.TriggerDrill(r.Context(), sosReq)
	if err != nil {
		h.logger.Error("Drill simulation failed", zap.Error(err))
		http.Error(w, "Failed to trigger drill", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Drill simulation completed", zap.String("incident_id", incident.ID.String()),
		zap.String("tenant_id", orgID.String()))
	writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"status":       "drill_triggered",
		"incident_id":  incident.ID,
		"simulation":   true,
		"organization": orgID.String(),
	})
}

func (h *EmergencyHandlers) GetIncidentTimeline(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	incidentID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid incident ID")
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	if orgID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
		h.respondWithError(w, http.StatusUnauthorized, "Organization ID required")
		return
	}

	timeline, err := h.service.GetIncidentTimeline(r.Context(), incidentID, orgID)
	if err != nil {
		if err == service.ErrIncidentNotFound {
			h.respondWithError(w, http.StatusNotFound, "Incident not found")
		} else if err == service.ErrUnauthorized {
			h.respondWithError(w, http.StatusForbidden, "Unauthorized access to incident")
		} else {
			h.logger.Error("Failed to get incident timeline", zap.Error(err))
			h.respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, timeline)
}

func (h *EmergencyHandlers) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *EmergencyHandlers) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
