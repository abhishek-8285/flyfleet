package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/route/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/route/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

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

	result, err := h.service.RecalculateRoute(r.Context(), req.ResultID, req.VehicleID, req.CurrentLat, req.CurrentLon)
	if err != nil {
		h.logger.Error("Failed to recalculate route", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to recalculate route", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// CreateOptimizationRequest handles POST /api/v1/routes/requests
func (h *RouteOptimizationHandler) CreateOptimizationRequest(w http.ResponseWriter, r *http.Request) {
	var request models.RouteOptimizationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := h.validate.Struct(&request); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	if err := h.service.CreateOptimizationRequest(r.Context(), &request); err != nil {
		h.logger.Error("Failed to create optimization request", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create optimization request", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, request)
}

// GetOptimizationRequest handles GET /api/v1/routes/requests/{id}
func (h *RouteOptimizationHandler) GetOptimizationRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	request, err := h.service.GetOptimizationRequest(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Optimization request not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, request)
}

// UpdateOptimizationRequest handles PUT /api/v1/routes/requests/{id}
func (h *RouteOptimizationHandler) UpdateOptimizationRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
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

	if err := h.service.DeleteOptimizationRequest(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete optimization request", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success", "message": "Optimization request deleted"})
}

// ListOptimizationRequests handles GET /api/v1/routes/requests
func (h *RouteOptimizationHandler) ListOptimizationRequests(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	orgIDStr := query.Get("org_id")
	var orgID uuid.UUID
	var err error

	if orgIDStr == "" {
		orgID = h.getOrgID(r)
	} else {
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid org_id", err)
			return
		}
	}

	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "org_id query parameter or context is required", nil)
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

	requests, totalCount, err := h.service.ListOptimizationRequests(r.Context(), orgID, limit, offset)
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
		h.writeErrorResponse(w, http.StatusNotFound, "Optimization result not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// GetOptimizationResultByRequestID handles GET /api/v1/routes/results/request/{request_id}
func (h *RouteOptimizationHandler) GetOptimizationResultByRequestID(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "request_id")

	result, err := h.service.GetOptimizationResultByRequestID(r.Context(), requestID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "Optimization result not found", err)
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

func (h *RouteOptimizationHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
		if id, ok := val.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}
