package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/service"
)

type FuelHandler struct {
	svc      service.FuelService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewFuelHandler(svc service.FuelService, logger *zap.Logger) *FuelHandler {
	return &FuelHandler{
		svc:      svc,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *FuelHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListLogs)
	r.Post("/", h.CreateLog)
	r.Get("/stats", h.GetStats)
	r.Get("/settings", h.GetSettings)
	r.Put("/settings", h.UpdateSettings)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetLog)
		r.Put("/", h.UpdateLog)
		r.Delete("/", h.DeleteLog)
		r.Patch("/status", h.ChangeStatus)
	})
}

// CreateLog godoc
func (h *FuelHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}

	var req models.CreateFuelLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	log, err := h.svc.CreateLog(r.Context(), orgID, userID, req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, log)
}

// GetLog godoc
func (h *FuelHandler) GetLog(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid ID", nil)
		return
	}

	log, err := h.svc.GetLog(r.Context(), id, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, log)
}

// ListLogs godoc
func (h *FuelHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}

	query := r.URL.Query()
	filter := models.FuelLogFilter{
		OrganizationID: orgID,
		Search:         query.Get("search"),
	}

	if vID := query.Get("vehicle_id"); vID != "" {
		if u, err := uuid.Parse(vID); err == nil {
			filter.VehicleID = &u
		}
	}

	if dID := query.Get("driver_id"); dID != "" {
		if u, err := uuid.Parse(dID); err == nil {
			filter.DriverID = &u
		}
	}

	status := query["status"]
	for _, s := range status {
		filter.Status = append(filter.Status, models.FuelLogStatus(s))
	}

	res, err := h.svc.ListLogs(r.Context(), filter)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, res)
}

// UpdateLog godoc
func (h *FuelHandler) UpdateLog(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid ID", nil)
		return
	}

	var req models.UpdateFuelLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	log, err := h.svc.UpdateLog(r.Context(), id, orgID, req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, log)
}

// DeleteLog godoc
func (h *FuelHandler) DeleteLog(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid ID", nil)
		return
	}

	if err := h.svc.DeleteLog(r.Context(), id, orgID); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ChangeStatus godoc
func (h *FuelHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid ID", nil)
		return
	}

	var req models.ChangeFuelLogStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.svc.ChangeStatus(r.Context(), id, orgID, req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "status updated"})
}

// GetStats godoc
func (h *FuelHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	stats, err := h.svc.GetStats(r.Context(), orgID, models.FuelLogFilter{})
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, stats)
}

// GetSettings godoc
func (h *FuelHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}
	settings, err := h.svc.GetSettings(r.Context(), orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, settings)
}

// UpdateSettings godoc
func (h *FuelHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	userID := h.getUserID(r)
	if !h.ensureAuthIDs(w, orgID, userID) {
		return
	}

	var req models.UpdateFuelSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	settings, err := h.svc.UpdateSettings(r.Context(), orgID, req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, settings)
}

// Helper methods

func (h *FuelHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *FuelHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := map[string]interface{}{
		"error": message,
	}
	if err != nil && statusCode == http.StatusInternalServerError {
		resp["details"] = err.Error()
	}
	h.writeJSONResponse(w, statusCode, resp)
}

func (h *FuelHandler) getOrgID(r *http.Request) uuid.UUID {
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
	if id, err := uuid.Parse(r.Header.Get("X-Organization-ID")); err == nil {
		return id
	}
	return uuid.Nil
}

func (h *FuelHandler) getUserID(r *http.Request) uuid.UUID {
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
	if id, err := uuid.Parse(r.Header.Get("X-User-ID")); err == nil {
		return id
	}
	return uuid.Nil
}

func (h *FuelHandler) ensureAuthIDs(w http.ResponseWriter, orgID uuid.UUID, userID uuid.UUID) bool {
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
