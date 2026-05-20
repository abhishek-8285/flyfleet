package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/route/models"
	"github.com/flyfleet/fleet-core/internal/domains/route/service"
	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DTOs for request payloads (separate from models to prevent mass assignment)
type createOptimizationRequestDTO struct {
	Name              string                   `json:"name" validate:"required"`
	Description       string                   `json:"description"`
	StartTime         time.Time                `json:"start_time" validate:"required"`
	EndTime           time.Time                `json:"end_time" validate:"required"`
	StartTimeWindow   *time.Time               `json:"start_time_window,omitempty"`
	EndTimeWindow     *time.Time               `json:"end_time_window,omitempty"`
	Constraints       models.RouteConstraints  `json:"constraints" validate:"required"`
	Stops             []models.Stop            `json:"stops" validate:"required,dive"`
	Vehicles          []models.VehicleProfile  `json:"vehicles" validate:"required,dive"`
	Depots            []models.Depot           `json:"depots,omitempty" validate:"dive"`
	Priorities        models.RoutePriorities   `json:"priorities" validate:"required"`
	OptimizationGoals models.OptimizationGoals `json:"optimization_goals" validate:"required"`
}

// RouteOptimizationHandler handles HTTP requests for route optimization
type RouteOptimizationHandler struct {
	service  service.RouteOptimizationService
	logger   *zap.Logger
	validate *validator.Validate
}

// NewRouteOptimizationHandler creates a new instance of RouteOptimizationHandler
func NewRouteOptimizationHandler(svc service.RouteOptimizationService, logger *zap.Logger) *RouteOptimizationHandler {
	return &RouteOptimizationHandler{
		service:  svc,
		logger:   logger,
		validate: validator.New(),
	}
}

// getTenantID extracts organization ID from context and returns error if missing
func (h *RouteOptimizationHandler) getTenantID(r *http.Request) uuid.UUID {
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == uuid.Nil {
		h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware",
			zap.String("path", r.URL.Path))
	}
	return tenantID
}

// requireResourceOwnership checks that a resource belongs to the user's organization
func (h *RouteOptimizationHandler) requireResourceOwnership(w http.ResponseWriter, r *http.Request, resourceTenantID uuid.UUID) bool {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant ID required", nil)
		return false
	}
	if resourceTenantID != tenantID {
		h.logger.Warn("unauthorized cross-org access attempt",
			zap.String("requested_org", tenantID.String()),
			zap.String("resource_org", resourceTenantID.String()),
			zap.String("path", r.URL.Path))
		h.writeErrorResponse(w, http.StatusNotFound, "Resource not found", nil)
		return false
	}
	return true
}

// RegisterRoutes registers all route optimization API routes
func (h *RouteOptimizationHandler) RegisterRoutes(r chi.Router) {
	r.Route("/requests", func(r chi.Router) {
		r.Post("/", h.CreateOptimizationRequest)
		r.Get("/", h.ListOptimizationRequests)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetOptimizationRequest)
			r.Put("/", h.UpdateOptimizationRequest)
			r.Delete("/", h.DeleteOptimizationRequest)
		})
	})

	r.Route("/results", func(r chi.Router) {
		r.Get("/", h.ListOptimizationResults)
		r.Get("/{id}", h.GetOptimizationResult)
		r.Get("/request/{request_id}", h.GetOptimizationResultByRequestID)
	})

	r.Route("/jobs", func(r chi.Router) {
		r.Post("/", h.SubmitOptimizationJob)
		r.Get("/", h.ListOptimizationJobs)
		r.Get("/{id}", h.GetOptimizationJob)
	})

	r.Post("/recalculate", h.RecalculateRoute)
}

// RecalculateRoute handles POST /api/v1/routes/recalculate
func (h *RouteOptimizationHandler) RecalculateRoute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ResultID   string  `json:"result_id" binding:"required"`
		VehicleID  string  `json:"vehicle_id" binding:"required"`
		CurrentLat float64 `json:"current_lat" binding:"required"`
		CurrentLon float64 `json:"current_lon" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// SECURITY: First verify the result belongs to the authenticated org
	result, err := h.service.GetOptimizationResult(r.Context(), req.ResultID)
	if err != nil {
		if err == service.ErrNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Optimization result not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to verify result ownership", err)
		}
		return
	}

	if !h.requireResourceOwnership(w, r, result.TenantID) {
		return
	}

	// Now perform the recalculation
	recalculated, err := h.service.RecalculateRoute(r.Context(), req.ResultID, req.VehicleID, req.CurrentLat, req.CurrentLon)
	if err != nil {
		h.logger.Error("Failed to recalculate route", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to recalculate route", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, recalculated)
}

// CreateOptimizationRequest handles POST /api/v1/routes/requests
func (h *RouteOptimizationHandler) CreateOptimizationRequest(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant ID required", nil)
		return
	}

	var dto createOptimizationRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.validate.Struct(dto); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	// Map DTO to model
	req := &models.RouteOptimizationRequest{
		TenantID:          tenantID,
		Name:              dto.Name,
		Description:       dto.Description,
		StartTime:         dto.StartTime,
		EndTime:           dto.EndTime,
		StartTimeWindow:   dto.StartTimeWindow,
		EndTimeWindow:     dto.EndTimeWindow,
		Constraints:       dto.Constraints,
		Stops:             dto.Stops,
		Vehicles:          dto.Vehicles,
		Depots:            dto.Depots,
		Priorities:        dto.Priorities,
		OptimizationGoals: dto.OptimizationGoals,
	}

	if err := h.service.CreateOptimizationRequest(r.Context(), req); err != nil {
		h.logger.Error("Failed to create optimization request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create optimization request", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, req)
}

// GetOptimizationRequest handles GET /api/v1/routes/requests/{id}
func (h *RouteOptimizationHandler) GetOptimizationRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	request, err := h.service.GetOptimizationRequest(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Optimization request not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch optimization request", err)
		}
		return
	}

	// SECURITY: Verify request belongs to authenticated organization
	if !h.requireResourceOwnership(w, r, request.TenantID) {
		return
	}

	h.writeJSONResponse(w, http.StatusOK, request)
}

// UpdateOptimizationRequest handles PUT /api/v1/routes/requests/{id}
func (h *RouteOptimizationHandler) UpdateOptimizationRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// First, fetch the request to verify ownership
	request, err := h.service.GetOptimizationRequest(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Optimization request not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch optimization request", err)
		}
		return
	}

	// SECURITY: Verify ownership before update
	if !h.requireResourceOwnership(w, r, request.TenantID) {
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Prevent tenant_id tampering
	if _, ok := updates["tenant_id"]; ok {
		h.writeErrorResponse(w, http.StatusBadRequest, "Cannot change tenant_id", nil)
		return
	}

	if err := h.service.UpdateOptimizationRequest(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update optimization request", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Optimization request updated"})
}

// DeleteOptimizationRequest handles DELETE /api/v1/routes/requests/{id}
func (h *RouteOptimizationHandler) DeleteOptimizationRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Verify ownership before delete
	request, err := h.service.GetOptimizationRequest(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Optimization request not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch optimization request", err)
		}
		return
	}

	if !h.requireResourceOwnership(w, r, request.TenantID) {
		return
	}

	if err := h.service.DeleteOptimizationRequest(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete optimization request", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Optimization request deleted"})
}

// ListOptimizationRequests handles GET /api/v1/routes/requests
func (h *RouteOptimizationHandler) ListOptimizationRequests(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// SECURITY: Always use org from authenticated context, ignore any tenant_id query parameter
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Tenant ID required", nil)
		return
	}

	limit := 10 // default
	if lStr := query.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if oStr := query.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	requests, totalCount, err := h.service.ListOptimizationRequests(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list optimization requests", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"requests":       requests,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(requests),
	})
}

// GetOptimizationResult handles GET /api/v1/routes/results/{id}
func (h *RouteOptimizationHandler) GetOptimizationResult(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.service.GetOptimizationResult(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Optimization result not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch optimization result", err)
		}
		return
	}

	// SECURITY: Verify result belongs to authenticated organization
	if !h.requireResourceOwnership(w, r, result.TenantID) {
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// GetOptimizationResultByRequestID handles GET /api/v1/routes/results/request/{request_id}
func (h *RouteOptimizationHandler) GetOptimizationResultByRequestID(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "request_id")

	result, err := h.service.GetOptimizationResultByRequestID(r.Context(), requestID)
	if err != nil {
		if err == service.ErrNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "Optimization result not found", nil)
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch optimization result", err)
		}
		return
	}

	// SECURITY: Verify result belongs to authenticated organization
	if !h.requireResourceOwnership(w, r, result.TenantID) {
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// ListOptimizationResults handles GET /api/v1/routes/results
func (h *RouteOptimizationHandler) ListOptimizationResults(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	requestID := query.Get("request_id")
	if requestID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "request_id query parameter is required", nil)
		return
	}

	limit := 10 // default
	if lStr := query.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if oStr := query.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	results, totalCount, err := h.service.ListOptimizationResults(r.Context(), requestID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list optimization results", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"results":        results,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(results),
	})
}

// SubmitOptimizationJob handles POST /api/v1/routes/jobs
func (h *RouteOptimizationHandler) SubmitOptimizationJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RequestID string `json:"request_id" binding:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON or missing request_id", err)
		return
	}

	if err := h.service.SubmitOptimizationJob(r.Context(), req.RequestID); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to submit optimization job", err)
		return
	}

	h.writeJSONResponse(w, http.StatusAccepted, map[string]string{
		"status":     "accepted",
		"message":    "Optimization job submitted successfully",
		"request_id": req.RequestID,
	})
}

// GetOptimizationJob handles GET /api/v1/routes/jobs/{id}
func (h *RouteOptimizationHandler) GetOptimizationJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	job, err := h.service.GetOptimizationJob(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Optimization job not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, job)
}

// ListOptimizationJobs handles GET /api/v1/routes/jobs
func (h *RouteOptimizationHandler) ListOptimizationJobs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	requestID := query.Get("request_id")
	status := query.Get("status") // optional filter

	if requestID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "request_id query parameter is required", nil)
		return
	}

	limit := 10 // default
	if lStr := query.Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if oStr := query.Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	jobs, totalCount, err := h.service.ListOptimizationJobs(r.Context(), requestID, status, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list optimization jobs", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"jobs":           jobs,
		"total_count":    totalCount,
		"limit":          limit,
		"offset":         offset,
		"returned_count": len(jobs),
	})
}

// Helper methods

func (h *RouteOptimizationHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *RouteOptimizationHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := map[string]interface{}{
		"error": message,
	}
	if err != nil {
		var errDetail string
		if statusCode >= 500 {
			errDetail = "Internal server error"
		} else {
			errDetail = err.Error()
		}
		resp["details"] = errDetail
	}
	h.writeJSONResponse(w, statusCode, resp)
}
