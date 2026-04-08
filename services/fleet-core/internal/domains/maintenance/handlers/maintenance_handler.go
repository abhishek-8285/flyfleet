package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core/internal/domains/maintenance/service"
)

type MaintenanceHandler struct {
	service *service.MaintenanceService
	logger  *zap.Logger
}

func NewMaintenanceHandler(svc *service.MaintenanceService, logger *zap.Logger) *MaintenanceHandler {
	return &MaintenanceHandler{
		service: svc,
		logger:  logger,
	}
}

func (h *MaintenanceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/schedules", func(r chi.Router) {
		r.Post("/", h.CreateSchedule)
		r.Get("/", h.ListSchedules)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetSchedule)
			r.Put("/", h.UpdateSchedule)
			r.Delete("/", h.DeleteSchedule)
		})
	})

	r.Route("/logs", func(r chi.Router) {
		r.Post("/", h.CreateLog)
		r.Get("/", h.ListLogs)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetLog)
			r.Put("/", h.UpdateLog)
			r.Delete("/", h.DeleteLog)
		})
	})

	r.Get("/upcoming", h.GetUpcomingMaintenance)
	r.Get("/overdue", h.GetOverdueMaintenance)
	r.Get("/summary", h.GetMaintenanceSummary)

	// Note: This route was originally /vehicles/{vehicle_id}/maintenance-history
	// If we want to keep it under /maintenance, we could change it or register it elsewhere
	r.Get("/vehicles/{vehicle_id}/history", h.GetVehicleMaintenanceHistory)
}

// CreateSchedule godoc
func (h *MaintenanceHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req models.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	userID := h.getUserID(r)
	if userID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User ID required", nil)
		return
	}

	schedule, err := h.service.CreateSchedule(r.Context(), orgID, &req, userID)
	if err != nil {
		h.logger.Error("failed to create maintenance schedule", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, schedule)
}

// GetSchedule godoc
func (h *MaintenanceHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	schedule, err := h.service.GetSchedule(r.Context(), id, orgID)
	if err != nil {
		h.logger.Error("failed to get maintenance schedule", zap.String("schedule_id", id), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Maintenance schedule not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, schedule)
}

// ListSchedules godoc
func (h *MaintenanceHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	query := r.URL.Query()
	filter := models.MaintenanceListFilter{
		OrgID:    orgID,
		Page:     1,
		PageSize: 20,
		OrderBy:  "due_date",
		OrderDir: "asc",
	}

	if page := query.Get("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum >= 1 {
			filter.Page = pageNum
		}
	}
	if pageSize := query.Get("page_size"); pageSize != "" {
		if size, err := strconv.Atoi(pageSize); err == nil && size >= 1 && size <= 100 {
			filter.PageSize = size
		}
	}

	if types := query["type"]; len(types) > 0 {
		filter.Type = types
	}
	if statuses := query["status"]; len(statuses) > 0 {
		filter.Status = statuses
	}
	if priorities := query["priority"]; len(priorities) > 0 {
		filter.Priority = priorities
	}

	if vehicleIDStr := query.Get("vehicle_id"); vehicleIDStr != "" {
		if vehicleID, err := uuid.Parse(vehicleIDStr); err == nil {
			filter.VehicleID = &vehicleID
		}
	}

	if assignedToStr := query.Get("assigned_to"); assignedToStr != "" {
		if assignedTo, err := uuid.Parse(assignedToStr); err == nil {
			filter.AssignedTo = &assignedTo
		}
	}

	if fromDateStr := query.Get("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			filter.FromDate = fromDate
		}
	}

	if toDateStr := query.Get("to_date"); toDateStr != "" {
		if toDate, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			filter.ToDate = toDate
		}
	}

	schedules, totalCount, err := h.service.ListSchedules(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list maintenance schedules", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list maintenance schedules", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        schedules,
		"total":       totalCount,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	})
}

// UpdateSchedule godoc
func (h *MaintenanceHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	var req models.UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	schedule, err := h.service.UpdateSchedule(r.Context(), id, orgID, &req)
	if err != nil {
		h.logger.Error("failed to update maintenance schedule", zap.String("schedule_id", id), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Maintenance schedule not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, schedule)
}

// DeleteSchedule godoc
func (h *MaintenanceHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	if err := h.service.DeleteSchedule(r.Context(), id, orgID); err != nil {
		h.logger.Error("failed to delete maintenance schedule", zap.String("schedule_id", id), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Maintenance schedule not found", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateLog godoc
func (h *MaintenanceHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	var req models.CreateLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	logEntry, err := h.service.CreateLog(r.Context(), orgID, &req)
	if err != nil {
		h.logger.Error("failed to create maintenance log", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, logEntry)
}

// GetLog godoc
func (h *MaintenanceHandler) GetLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	logEntry, err := h.service.GetLog(r.Context(), id, orgID)
	if err != nil {
		h.logger.Error("failed to get maintenance log", zap.String("log_id", id), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Maintenance log not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, logEntry)
}

// ListLogs godoc
func (h *MaintenanceHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	query := r.URL.Query()
	filter := models.MaintenanceListFilter{
		OrgID:    orgID,
		Page:     1,
		PageSize: 20,
		OrderBy:  "start_date",
		OrderDir: "desc",
	}

	if page := query.Get("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum >= 1 {
			filter.Page = pageNum
		}
	}
	if pageSize := query.Get("page_size"); pageSize != "" {
		if size, err := strconv.Atoi(pageSize); err == nil && size >= 1 && size <= 100 {
			filter.PageSize = size
		}
	}

	if types := query["type"]; len(types) > 0 {
		filter.Type = types
	}
	if statuses := query["status"]; len(statuses) > 0 {
		filter.Status = statuses
	}

	if vehicleIDStr := query.Get("vehicle_id"); vehicleIDStr != "" {
		if vehicleID, err := uuid.Parse(vehicleIDStr); err == nil {
			filter.VehicleID = &vehicleID
		}
	}

	if performedByStr := query.Get("performed_by"); performedByStr != "" {
		if performedBy, err := uuid.Parse(performedByStr); err == nil {
			filter.AssignedTo = &performedBy
		}
	}

	if fromDateStr := query.Get("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			filter.FromDate = fromDate
		}
	}

	if toDateStr := query.Get("to_date"); toDateStr != "" {
		if toDate, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			filter.ToDate = toDate
		}
	}

	logs, totalCount, err := h.service.ListLogs(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list maintenance logs", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list maintenance logs", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        logs,
		"total":       totalCount,
		"page":        filter.Page,
		"page_size":   filter.PageSize,
		"total_pages": (totalCount + filter.PageSize - 1) / filter.PageSize,
	})
}

// UpdateLog godoc
func (h *MaintenanceHandler) UpdateLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	var req models.UpdateLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	logEntry, err := h.service.UpdateLog(r.Context(), id, orgID, &req)
	if err != nil {
		h.logger.Error("failed to update maintenance log", zap.String("log_id", id), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Maintenance log not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, logEntry)
}

// DeleteLog godoc
func (h *MaintenanceHandler) DeleteLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	if err := h.service.DeleteLog(r.Context(), id, orgID); err != nil {
		h.logger.Error("failed to delete maintenance log", zap.String("log_id", id), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Maintenance log not found", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetVehicleMaintenanceHistory godoc
func (h *MaintenanceHandler) GetVehicleMaintenanceHistory(w http.ResponseWriter, r *http.Request) {
	vehicleIDStr := chi.URLParam(r, "vehicle_id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID", nil)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limitVal, err := strconv.Atoi(limitStr); err == nil && limitVal > 0 && limitVal <= 100 {
			limit = limitVal
		}
	}

	logs, err := h.service.GetMaintenanceHistory(r.Context(), vehicleID, orgID, limit)
	if err != nil {
		h.logger.Error("failed to get maintenance history", zap.String("vehicle_id", vehicleIDStr), zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get maintenance history", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, logs)
}

// GetUpcomingMaintenance godoc
func (h *MaintenanceHandler) GetUpcomingMaintenance(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	days := 30
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if daysVal, err := strconv.Atoi(daysStr); err == nil && daysVal > 0 && daysVal <= 365 {
			days = daysVal
		}
	}

	schedules, err := h.service.GetUpcomingSchedules(r.Context(), orgID, days)
	if err != nil {
		h.logger.Error("failed to get upcoming maintenance", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get upcoming maintenance", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, schedules)
}

// GetOverdueMaintenance godoc
func (h *MaintenanceHandler) GetOverdueMaintenance(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	schedules, err := h.service.GetOverdueSchedules(r.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get overdue maintenance", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get overdue maintenance", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, schedules)
}

// GetMaintenanceSummary godoc
func (h *MaintenanceHandler) GetMaintenanceSummary(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Organization ID required", nil)
		return
	}

	query := r.URL.Query()
	fromDateStr := query.Get("from_date")
	toDateStr := query.Get("to_date")

	if fromDateStr == "" || toDateStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Both from_date and to_date are required", nil)
		return
	}

	fromDate, err := time.Parse(time.RFC3339, fromDateStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid from_date format", err)
		return
	}

	toDate, err := time.Parse(time.RFC3339, toDateStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid to_date format", err)
		return
	}

	summary, err := h.service.GetMaintenanceSummary(r.Context(), orgID, fromDate, toDate)
	if err != nil {
		h.logger.Error("failed to get maintenance summary", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get maintenance summary", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, summary)
}

// Helper methods

func (h *MaintenanceHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *MaintenanceHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := map[string]interface{}{
		"error": message,
	}
	if err != nil && statusCode == http.StatusInternalServerError {
		resp["details"] = err.Error()
	}
	h.writeJSONResponse(w, statusCode, resp)
}

func (h *MaintenanceHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func (h *MaintenanceHandler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(userIDStr); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
