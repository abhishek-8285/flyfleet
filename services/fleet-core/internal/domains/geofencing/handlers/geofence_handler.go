package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/geofencing/models"
	"github.com/flyfleet/fleet-core/internal/domains/geofencing/service"
)

// DTOs for request payloads (separate from models to avoid mass assignment)

type createGeofenceRequestDTO struct {
	Name          string                 `json:"name" binding:"required,max=100"`
	Type          string                 `json:"type" binding:"required,oneof=circle polygon route_corridor"`
	Center        *models.Location       `json:"center,omitempty"`
	RadiusMeters  *int                   `json:"radius,omitempty"`
	Coordinates   []models.Location      `json:"coordinates,omitempty"`
	RoutePolyline *string                `json:"route_polyline,omitempty"`
	BufferMeters  *int                   `json:"buffer_meters,omitempty"`
	EntryAlert    bool                   `json:"entry_alert"`
	ExitAlert     bool                   `json:"exit_alert"`
	SpeedLimit    *int                   `json:"speed_limit,omitempty"`
	Vehicles      []uuid.UUID            `json:"vehicles"`
	ActiveHours   *models.ActiveHours    `json:"active_hours,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type updateGeofenceRequestDTO struct {
	Name          *string                `json:"name,omitempty"`
	Type          *string                `json:"type,omitempty"`
	Center        *models.Location       `json:"center,omitempty"`
	RadiusMeters  *int                   `json:"radius,omitempty"`
	Coordinates   []models.Location      `json:"coordinates,omitempty"`
	RoutePolyline *string                `json:"route_polyline,omitempty"`
	BufferMeters  *int                   `json:"buffer_meters,omitempty"`
	EntryAlert    *bool                  `json:"entry_alert,omitempty"`
	ExitAlert     *bool                  `json:"exit_alert,omitempty"`
	SpeedLimit    *int                   `json:"speed_limit,omitempty"`
	Vehicles      []uuid.UUID            `json:"vehicles,omitempty"`
	ActiveHours   *models.ActiveHours    `json:"active_hours,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Active        *bool                  `json:"active,omitempty"`
}

type GeofenceHandler struct {
	service  *service.GeofenceService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewGeofenceHandler(svc *service.GeofenceService, logger *zap.Logger) *GeofenceHandler {
	return &GeofenceHandler{
		service:  svc,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *GeofenceHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListGeofences)
	r.Post("/", h.CreateGeofence)
	r.Get("/events", h.GetGeofenceEvents)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetGeofence)
		r.Put("/", h.UpdateGeofence)
		r.Delete("/", h.DeleteGeofence)
	})
}

// CreateGeofence handles creating a new geofence
func (h *GeofenceHandler) CreateGeofence(w http.ResponseWriter, r *http.Request) {
	var dto createGeofenceRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("Invalid request body for creating geofence", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.requireAuthContext(w, orgID, userID) {
		return
	}

	// Map DTO to model
	req := &models.CreateGeofenceRequest{
		Name:          dto.Name,
		Type:          dto.Type,
		Center:        dto.Center,
		RadiusMeters:  dto.RadiusMeters,
		Coordinates:   dto.Coordinates,
		RoutePolyline: dto.RoutePolyline,
		BufferMeters:  dto.BufferMeters,
		EntryAlert:    dto.EntryAlert,
		ExitAlert:     dto.ExitAlert,
		SpeedLimit:    dto.SpeedLimit,
		Vehicles:      dto.Vehicles,
		ActiveHours:   dto.ActiveHours,
		Metadata:      dto.Metadata,
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	geofence, err := h.service.Create(r.Context(), orgID, userID, req)
	if err != nil {
		h.logger.Error("Failed to create geofence", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{"data": geofence})
}

// GetGeofence retrieves a geofence by ID
func (h *GeofenceHandler) GetGeofence(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid geofence ID", zap.String("id", idStr), zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid geofence ID", nil)
		return
	}

	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.requireAuthContext(w, orgID, userID) {
		return
	}

	geofence, err := h.service.GetByID(r.Context(), id, orgID)
	if err != nil {
		h.logger.Error("Failed to get geofence", zap.String("id", idStr), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Geofence not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, geofence)
}

// ListGeofences retrieves geofences with pagination and filtering
func (h *GeofenceHandler) ListGeofences(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var req models.ListGeofencesRequest

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	req.Page = page

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
	req.Limit = limit

	req.Search = query.Get("search")
	req.Type = query.Get("type")

	activeStr := query.Get("active")
	if activeStr != "" {
		activeBool := activeStr == "true"
		req.Active = &activeBool
	}

	req.Sort = query.Get("sort")

	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.requireAuthContext(w, orgID, userID) {
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	response, err := h.service.List(r.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("Failed to list geofences", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Calculate pagination metadata
	totalPages := 0
	if response.Limit > 0 {
		totalPages = int((response.Total + int64(response.Limit) - 1) / int64(response.Limit))
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data": response.Geofences,
		"meta": map[string]interface{}{
			"total":       response.Total,
			"page":        response.Page,
			"limit":       response.Limit,
			"total_pages": totalPages,
			"has_next":    response.Page < totalPages,
			"has_prev":    response.Page > 1,
		},
	})
}

// UpdateGeofence updates an existing geofence
func (h *GeofenceHandler) UpdateGeofence(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid geofence ID", zap.String("id", idStr), zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid geofence ID", nil)
		return
	}

	var dto updateGeofenceRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.logger.Error("Invalid request body for updating geofence", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.requireAuthContext(w, orgID, userID) {
		return
	}

	// Map DTO to model
	req := &models.UpdateGeofenceRequest{
		Name:          dto.Name,
		Type:          dto.Type,
		Center:        dto.Center,
		RadiusMeters:  dto.RadiusMeters,
		Coordinates:   dto.Coordinates,
		RoutePolyline: dto.RoutePolyline,
		BufferMeters:  dto.BufferMeters,
		EntryAlert:    dto.EntryAlert,
		ExitAlert:     dto.ExitAlert,
		SpeedLimit:    dto.SpeedLimit,
		Vehicles:      dto.Vehicles,
		ActiveHours:   dto.ActiveHours,
		Metadata:      dto.Metadata,
		Active:        dto.Active,
	}

	if err := h.validate.Struct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	updatedGeofence, err := h.service.Update(r.Context(), id, orgID, req)
	if err != nil {
		h.logger.Error("Failed to update geofence", zap.String("id", idStr), zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": updatedGeofence})
}

// DeleteGeofence deletes a geofence (marks as inactive)
func (h *GeofenceHandler) DeleteGeofence(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid geofence ID", zap.String("id", idStr), zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid geofence ID", nil)
		return
	}

	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.requireAuthContext(w, orgID, userID) {
		return
	}

	if err := h.service.Delete(r.Context(), id, orgID); err != nil {
		h.logger.Error("Failed to delete geofence", zap.String("id", idStr), zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetGeofenceEvents retrieves geofence events with filtering
func (h *GeofenceHandler) GetGeofenceEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	var filter models.GeofenceEventFilter
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.requireAuthContext(w, orgID, userID) {
		return
	}

	geofenceIDStr := query.Get("geofence_id")
	if geofenceIDStr != "" {
		geofenceID, err := uuid.Parse(geofenceIDStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid geofence ID", nil)
			return
		}
		filter.GeofenceID = geofenceID
	}

	vehicleIDStr := query.Get("vehicle_id")
	if vehicleIDStr != "" {
		vehicleID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID", nil)
			return
		}
		filter.VehicleID = vehicleID
	}

	filter.EventType = query.Get("event_type")

	page, err := strconv.Atoi(query.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	filter.Page = page

	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}
	filter.Limit = limit
	filter.OrgID = orgID
	if err := h.validate.Struct(filter); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	events, total, err := h.service.GetEvents(r.Context(), &filter)
	if err != nil {
		h.logger.Error("Failed to get geofence events", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"events":   events,
		"total":    total,
		"page":     filter.Page,
		"limit":    filter.Limit,
		"has_more": int64((filter.Page * filter.Limit)) < total,
	})
}

// Helper methods

func (h *GeofenceHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *GeofenceHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := map[string]interface{}{
		"error": message,
	}
	if err != nil && statusCode == http.StatusInternalServerError {
		resp["details"] = err.Error()
	}
	h.writeJSONResponse(w, statusCode, resp)
}

func (h *GeofenceHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}
	// SECURITY FIX: Removed fallback to X-Organization-ID header to prevent header spoofing
	// The org ID must come from validated JWT context only
	h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
	return uuid.Nil
}

func (h *GeofenceHandler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}
	// SECURITY FIX: Removed fallback to X-User-ID header to prevent header spoofing
	// The user ID must come from validated JWT context only
	h.logger.Warn("Missing user_id in context - request likely bypassed authentication middleware")
	return uuid.Nil
}

func (h *GeofenceHandler) requireAuthContext(w http.ResponseWriter, orgID uuid.UUID, userID uuid.UUID) bool {
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "missing organization id", nil)
		return false
	}
	if userID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "missing user id", nil)
		return false
	}
	return true
}
