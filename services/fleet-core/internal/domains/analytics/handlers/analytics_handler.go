package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"errors"
	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics"
	analyticsMiddleware "github.com/flyfleet/fleet-core-service/internal/domains/analytics/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Rate limiters for expensive endpoints
var (
	// Fleet analytics: 10 requests per minute (expensive aggregation)
	fleetAnalyticsLimiter = NewRateLimiter(10, time.Minute)
	// Driver behavior: 30 requests per minute
	driverBehaviorLimiter = NewRateLimiter(30, time.Minute)
	// Predictions: 60 requests per minute
	predictionLimiter = NewRateLimiter(60, time.Minute)

	initLimitersOnce sync.Once
)

// AnalyticsHandler handles HTTP requests for analytics operations
type AnalyticsHandler struct {
	service          analytics.AnalyticsService
	guard            *analyticsMiddleware.TenantGuard
	logger           *zap.Logger
	rateLimiter      *analyticsMiddleware.RateLimiter
	qualityValidator *analyticsMiddleware.DataQualityValidator
}

// NewAnalyticsHandler creates a new instance of AnalyticsHandler
func NewAnalyticsHandler(service analytics.AnalyticsService, guard *analyticsMiddleware.TenantGuard, logger *zap.Logger, rateLimiter *analyticsMiddleware.RateLimiter, qualityValidator *analyticsMiddleware.DataQualityValidator) *AnalyticsHandler {
	// Start cleanup goroutine for rate limiters only once
	initLimitersOnce.Do(func() {
		stopCh := make(chan struct{})
		fleetAnalyticsLimiter.StartCleanup(time.Minute, stopCh)
		driverBehaviorLimiter.StartCleanup(time.Minute, stopCh)
		predictionLimiter.StartCleanup(time.Minute, stopCh)
	})

	return &AnalyticsHandler{
		service:          service,
		guard:            guard,
		logger:           logger,
		rateLimiter:      rateLimiter,
		qualityValidator: qualityValidator,
	}
}

// RegisterRoutes registers the analytics routes
func (h *AnalyticsHandler) RegisterRoutes(r chi.Router) {
	// h.service.SubscribeToInvalidations(context.Background())

	r.Route("/analytics", func(r chi.Router) {
		r.Use(h.rateLimiter.OrgRateLimit)

		// Predictive Models
		r.Route("/predictive-models", func(r chi.Router) {
			r.Post("/", h.CreatePredictiveModel)
			r.Get("/", h.ListPredictiveModels)
			r.Get("/version", h.GetPredictiveModelByVersion)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetPredictiveModel)
				r.Put("/", h.UpdatePredictiveModel)
				r.Delete("/", h.DeletePredictiveModel)
			})
		})

		// Prediction Results
		r.Route("/predictions", func(r chi.Router) {
			r.Post("/", h.CreatePredictionResult)
			r.Get("/", h.ListPredictionResults)
			r.Get("/org", h.GetPredictionResultsByOrg)
			r.Get("/{id}", h.GetPredictionResult)
			r.Get("/model/{modelId}", h.GetPredictionResultsByModel)
		})

		// Analytics Jobs
		r.Route("/jobs", func(r chi.Router) {
			r.Post("/", h.CreateAnalyticsJob)
			r.Get("/", h.ListAnalyticsJobs)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetAnalyticsJob)
				r.Put("/", h.UpdateAnalyticsJob)
				r.Delete("/", h.DeleteAnalyticsJob)
			})
		})

		// Phase 8: Operations & ML Lifecycle
		r.Route("/deployments", func(r chi.Router) {
			r.Post("/", h.DeployModel)
		})

		r.Route("/risk-events", func(r chi.Router) {
			r.Post("/", h.CreateDriverRiskEvent)
			r.Get("/", h.GetDriverRiskEvents)
		})

		r.Route("/quotas", func(r chi.Router) {
			r.Get("/", h.GetAnalyticsQuota)
		})

		r.Route("/ops", func(r chi.Router) {
			r.Post("/reconciliation", h.RunReconciliation)
			r.Post("/backfill", h.RunHistoricalBackfill)
		})

		// Fleet Analytics
		r.Route("/fleet", func(r chi.Router) {
			r.Post("/", h.CreateFleetAnalytics)
			r.Get("/", h.ListFleetAnalytics)
			r.Get("/period", h.GetFleetAnalyticsByPeriod)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetFleetAnalytics)
				r.Put("/", h.UpdateFleetAnalytics)
				r.Delete("/", h.DeleteFleetAnalytics)
			})
		})

		// Driver Behavior
		r.Route("/driver-behavior", func(r chi.Router) {
			r.Post("/", h.CreateDriverBehaviorAnalytics)
			r.Get("/", h.ListDriverBehaviorAnalytics)
			r.Get("/driver", h.GetDriverBehaviorAnalyticsByDriver)
			r.Get("/org", h.GetDriverBehaviorAnalyticsByOrg)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetDriverBehaviorAnalytics)
				r.Put("/", h.UpdateDriverBehaviorAnalytics)
				r.Delete("/", h.DeleteDriverBehaviorAnalytics)
			})
		})

		// Predictive Maintenance
		r.Route("/predictive-maintenance", func(r chi.Router) {
			r.Post("/", h.CreatePredictiveMaintenanceAnalytics)
			r.Get("/", h.ListPredictiveMaintenanceAnalytics)
			r.Get("/vehicle", h.GetPredictiveMaintenanceAnalyticsByVehicle)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetPredictiveMaintenanceAnalytics)
				r.Put("/", h.UpdatePredictiveMaintenanceAnalytics)
				r.Delete("/", h.DeletePredictiveMaintenanceAnalytics)
			})
		})

		// Fleet Analytics with Ownership Guard
		r.Route("/analytics/fleet", func(r chi.Router) {
			r.Use(h.guard.ResourceOwnershipGuard())
			r.Get("/summary", h.GetFleetSummary)
			r.Get("/hourly", h.GetFleetHourlyStats)
		})

		// Stats routes with cursor pagination
		r.Group(func(r chi.Router) {
			r.Use(h.rateLimiter.OrgRateLimit)
			r.Use(h.guard.ResourceOwnershipGuard())
			r.Get("/stats/vehicle/{vehicleID}", h.GetVehicleDailyStats)
			r.Get("/stats/driver/{driverID}", h.GetDriverDailyStats)
		})

		// Health Check
		r.Get("/health", h.HealthCheck)
	})
}

// CreatePredictiveModel handles the creation of a new predictive model
func (h *AnalyticsHandler) CreatePredictiveModel(w http.ResponseWriter, r *http.Request) {
	var model models.PredictiveModel
	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate using validator
	if err := ValidateRequest(&model); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	// Ensure organization ID matches context if applicable (PredictiveModel is global-ish but might be org-bound)
	// For now, focusing on the ones we know are org-bound

	// Set ID and timestamps
	model.ID = uuid.New().String()
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()

	if err := h.service.CreatePredictiveModel(r.Context(), &model); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, model)
}

// GetPredictiveModel handles retrieving a predictive model by ID
func (h *AnalyticsHandler) GetPredictiveModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	model, err := h.service.GetPredictiveModel(r.Context(), id, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, model)
}

// GetPredictiveModelByVersion handles retrieving a predictive model by name and version
func (h *AnalyticsHandler) GetPredictiveModelByVersion(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	version := r.URL.Query().Get("version")
	orgID := h.getOrgID(r)

	if name == "" || version == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Name and version are required", nil)
		return
	}

	model, err := h.service.GetPredictiveModelByVersion(r.Context(), name, version, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, model)
}

// UpdatePredictiveModel handles updating an existing predictive model
func (h *AnalyticsHandler) UpdatePredictiveModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	var req models.UpdatePredictiveModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	// Map to only the fields we allow updating
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Version != nil {
		updates["version"] = *req.Version
	}
	updates["updated_at"] = time.Now()

	if err := h.service.UpdatePredictiveModel(r.Context(), id, orgID, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeletePredictiveModel handles deleting a predictive model
func (h *AnalyticsHandler) DeletePredictiveModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	if err := h.service.DeletePredictiveModel(r.Context(), id, orgID); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListPredictiveModels handles listing predictive models with filtering
func (h *AnalyticsHandler) ListPredictiveModels(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	orgID := h.getOrgID(r)
	modelType := parseStringParam(query.Get("type"))
	status := parseStringParam(query.Get("status"))

	limit, offset := parsePaginationParams(query)
	cursor := query.Get("cursor")

	modelsList, totalCount, err := h.service.ListPredictiveModels(r.Context(), orgID, modelType, status, cursor, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var nextCursor string
	if len(modelsList) == limit && len(modelsList) > 0 {
		last := modelsList[len(modelsList)-1]
		nextCursor = utils.EncodeCursor(last.CreatedAt, last.ID)
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        modelsList,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
		"next_cursor": nextCursor,
	})
}

// CreatePredictionResult handles creating a new prediction result
func (h *AnalyticsHandler) CreatePredictionResult(w http.ResponseWriter, r *http.Request) {
	// Rate limiting
	orgID := h.getOrgID(r)
	if !predictionLimiter.Allow(orgID.String()) {
		h.writeErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded: Maximum 60 requests per minute", nil)
		return
	}

	var result models.PredictionResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate using validator
	if err := ValidateRequest(&result); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	// Ensure organization ID matches the one from context
	result.OrganizationID = orgID

	// Set ID and timestamps
	result.ID = uuid.New().String()
	result.CreatedAt = time.Now()
	result.UpdatedAt = time.Now()
	result.PredictedAt = time.Now()

	if err := h.service.CreatePredictionResult(r.Context(), &result); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, result)
}

// GetPredictionResult handles retrieving a prediction result by ID
func (h *AnalyticsHandler) GetPredictionResult(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	result, err := h.service.GetPredictionResult(r.Context(), id, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// GetPredictionResultsByModel handles retrieving prediction results by model ID
func (h *AnalyticsHandler) GetPredictionResultsByModel(w http.ResponseWriter, r *http.Request) {
	modelID := chi.URLParam(r, "modelId")
	orgID := h.getOrgID(r)

	limit, offset := parsePaginationParams(r.URL.Query())

	results, totalCount, err := h.service.GetPredictionResultsByModel(r.Context(), modelID, orgID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        results,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetPredictionResultsByOrg handles retrieving prediction results by organization
func (h *AnalyticsHandler) GetPredictionResultsByOrg(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		orgID = h.getOrgID(r)
	}

	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	results, totalCount, err := h.service.GetPredictionResultsByOrg(r.Context(), orgID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        results,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// ListPredictionResults handles listing prediction results with filtering
func (h *AnalyticsHandler) ListPredictionResults(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	orgIDStr := query.Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		orgID = h.getOrgID(r)
	}

	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}

	modelID := parseStringParam(query.Get("modelId"))
	anomalyDetected := parseBoolParam(query.Get("anomalyDetected"))

	limit, offset := parsePaginationParams(query)
	cursor := query.Get("cursor")

	results, totalCount, err := h.service.ListPredictionResults(r.Context(), orgID, modelID, anomalyDetected, cursor, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var nextCursor string
	if len(results) == limit && len(results) > 0 {
		last := results[len(results)-1]
		nextCursor = utils.EncodeCursor(last.CreatedAt, last.ID)
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        results,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
		"next_cursor": nextCursor,
	})
}

// CreateAnalyticsJob handles creating a new analytics job
func (h *AnalyticsHandler) CreateAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	// Rate limiting (shared with fleet analytics as they are both heavy)
	orgID := h.getOrgID(r)
	if !fleetAnalyticsLimiter.Allow(orgID.String()) {
		h.writeErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded: Maximum 10 requests per minute", nil)
		return
	}

	var job models.AnalyticsJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate using validator
	if err := ValidateRequest(&job); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	// Ensure organization ID matches context
	job.OrganizationID = orgID

	// Set ID and timestamps
	job.ID = uuid.New().String()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	if err := h.service.CreateAnalyticsJob(r.Context(), &job); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, job)
}

// GetAnalyticsJob handles retrieving an analytics job by ID
func (h *AnalyticsHandler) GetAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	job, err := h.service.GetAnalyticsJob(r.Context(), id, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, job)
}

// UpdateAnalyticsJob handles updating an existing analytics job
func (h *AnalyticsHandler) UpdateAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	var req models.UpdateAnalyticsJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	updates := make(map[string]interface{})
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Progress != nil {
		updates["progress"] = *req.Progress
	}
	if req.ErrorDetails != nil {
		updates["error_details"] = *req.ErrorDetails
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Parameters != nil {
		updates["parameters"] = req.Parameters
	}
	updates["updated_at"] = time.Now()

	if err := h.service.UpdateAnalyticsJob(r.Context(), id, orgID, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteAnalyticsJob handles deleting an analytics job
func (h *AnalyticsHandler) DeleteAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	if err := h.service.DeleteAnalyticsJob(r.Context(), id, orgID); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListAnalyticsJobs handles listing analytics jobs with filtering
func (h *AnalyticsHandler) ListAnalyticsJobs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	jobType := parseStringParam(query.Get("jobType"))
	status := parseStringParam(query.Get("status"))

	orgIDStr := query.Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		orgID = h.getOrgID(r)
	}

	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}

	limit, offset := parsePaginationParams(query)
	cursor := query.Get("cursor")

	jobs, totalCount, err := h.service.ListAnalyticsJobs(r.Context(), jobType, status, orgID, cursor, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var nextCursor string
	if len(jobs) == limit && len(jobs) > 0 {
		last := jobs[len(jobs)-1]
		nextCursor = utils.EncodeCursor(last.CreatedAt, last.ID)
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        jobs,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
		"next_cursor": nextCursor,
	})
}

// CreateFleetAnalytics handles creating new fleet analytics
func (h *AnalyticsHandler) CreateFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	var analytics models.FleetAnalytics
	if err := json.NewDecoder(r.Body).Decode(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate using validator
	if err := ValidateRequest(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	// Set ID and timestamps
	analytics.ID = uuid.New().String()
	analytics.CreatedAt = time.Now()
	analytics.UpdatedAt = time.Now()
	analytics.GeneratedAt = time.Now()

	if err := h.service.CreateFleetAnalytics(r.Context(), &analytics); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, analytics)
}

// GetFleetAnalytics handles retrieving fleet analytics by ID
func (h *AnalyticsHandler) GetFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	analytics, err := h.service.GetFleetAnalytics(r.Context(), id, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analytics)
}

// GetFleetAnalyticsByPeriod handles retrieving fleet analytics by organization and date period
func (h *AnalyticsHandler) GetFleetAnalyticsByPeriod(w http.ResponseWriter, r *http.Request) {
	// Rate limiting: 10 requests per minute per org
	orgIDStr := r.URL.Query().Get("orgId")
	if orgIDStr == "" {
		orgIDStr = h.getOrgID(r).String()
	}
	if !fleetAnalyticsLimiter.Allow(orgIDStr) {
		h.writeErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded: Maximum 10 requests per minute", nil)
		return
	}

	query := r.URL.Query()
	orgIDStr = query.Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		orgID = h.getOrgID(r)
	}

	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}

	startDateStr := query.Get("startDate")
	endDateStr := query.Get("endDate")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid start date format, expected YYYY-MM-DD", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid end date format, expected YYYY-MM-DD", err)
		return
	}

	analytics, err := h.service.GetFleetAnalyticsByPeriod(r.Context(), orgID, startDate, endDate)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analytics)
}

// UpdateFleetAnalytics handles updating existing fleet analytics
func (h *AnalyticsHandler) UpdateFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdateFleetAnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	updates := make(map[string]interface{})
	if req.TotalVehicles != nil {
		updates["total_vehicles"] = *req.TotalVehicles
	}
	if req.ActiveVehicles != nil {
		updates["active_vehicles"] = *req.ActiveVehicles
	}
	if req.TotalDistance != nil {
		updates["total_distance"] = *req.TotalDistance
	}
	if req.AverageSpeed != nil {
		updates["average_speed"] = *req.AverageSpeed
	}
	if req.TotalFuelConsumption != nil {
		updates["total_fuel_consumption"] = *req.TotalFuelConsumption
	}
	if req.AverageFuelEfficiency != nil {
		updates["average_fuel_efficiency"] = *req.AverageFuelEfficiency
	}
	if req.SafetyScore != nil {
		updates["safety_score"] = *req.SafetyScore
	}
	updates["updated_at"] = time.Now()

	if err := h.service.UpdateFleetAnalytics(r.Context(), id, h.getOrgID(r), updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteFleetAnalytics handles deleting fleet analytics
func (h *AnalyticsHandler) DeleteFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	if err := h.service.DeleteFleetAnalytics(r.Context(), id, orgID); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListFleetAnalytics handles listing fleet analytics with filtering
func (h *AnalyticsHandler) ListFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	orgIDStr := query.Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		orgID = h.getOrgID(r)
	}

	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}

	var startDate, endDate *time.Time

	if startDateStr := query.Get("startDate"); startDateStr != "" {
		date, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid start date format, expected YYYY-MM-DD", err)
			return
		}
		startDate = &date
	}

	if endDateStr := query.Get("endDate"); endDateStr != "" {
		date, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid end date format, expected YYYY-MM-DD", err)
			return
		}
		endDate = &date
	}

	limit, offset := parsePaginationParams(query)
	cursor := query.Get("cursor")

	analyticsList, totalCount, err := h.service.ListFleetAnalytics(r.Context(), orgID, startDate, endDate, cursor, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var nextCursor string
	if len(analyticsList) == limit && len(analyticsList) > 0 {
		last := analyticsList[len(analyticsList)-1]
		nextCursor = utils.EncodeCursor(last.StartDate, last.ID)
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        analyticsList,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
		"next_cursor": nextCursor,
	})
}

// CreateDriverBehaviorAnalytics handles creating new driver behavior analytics
func (h *AnalyticsHandler) CreateDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	// Rate limiting
	orgID := h.getOrgID(r)
	if !driverBehaviorLimiter.Allow(orgID.String()) {
		h.writeErrorResponse(w, http.StatusTooManyRequests, "Rate limit exceeded", nil)
		return
	}

	var analytics models.DriverBehaviorAnalytics
	if err := json.NewDecoder(r.Body).Decode(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate using validator
	if err := ValidateRequest(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	// Ensure organization ID matches context
	analytics.OrganizationID = orgID

	// Set ID and timestamps
	analytics.ID = uuid.New().String()
	analytics.CreatedAt = time.Now()
	analytics.UpdatedAt = time.Now()
	analytics.GeneratedAt = time.Now()

	if err := h.service.CreateDriverBehaviorAnalytics(r.Context(), &analytics); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, analytics)
}

// GetDriverBehaviorAnalytics handles retrieving driver behavior analytics by ID
func (h *AnalyticsHandler) GetDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	analytics, err := h.service.GetDriverBehaviorAnalytics(r.Context(), id, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analytics)
}

// GetDriverBehaviorAnalyticsByDriver handles retrieving driver behavior analytics by driver ID
func (h *AnalyticsHandler) GetDriverBehaviorAnalyticsByDriver(w http.ResponseWriter, r *http.Request) {
	driverIDStr := r.URL.Query().Get("driverId")
	driverID, err := uuid.Parse(driverIDStr)
	orgID := h.getOrgID(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID", err)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.GetDriverBehaviorAnalyticsByDriver(r.Context(), driverID, orgID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetDriverBehaviorAnalyticsByOrg handles retrieving driver behavior analytics by organization
func (h *AnalyticsHandler) GetDriverBehaviorAnalyticsByOrg(w http.ResponseWriter, r *http.Request) {
	orgIDStr := r.URL.Query().Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		orgID = h.getOrgID(r)
	}

	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.GetDriverBehaviorAnalyticsByOrg(r.Context(), orgID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// UpdateDriverBehaviorAnalytics handles updating existing driver behavior analytics
func (h *AnalyticsHandler) UpdateDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdateDriverBehaviorAnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	updates := make(map[string]interface{})
	if req.SafetyScore != nil {
		updates["safety_score"] = *req.SafetyScore
	}
	if req.EfficiencyScore != nil {
		updates["efficiency_score"] = *req.EfficiencyScore
	}
	if req.ComfortScore != nil {
		updates["comfort_score"] = *req.ComfortScore
	}
	if req.ImprovementSuggestions != nil {
		updates["improvement_suggestions"] = req.ImprovementSuggestions
	}
	if req.BehaviorTrend != nil {
		updates["behavior_trend"] = *req.BehaviorTrend
	}
	updates["updated_at"] = time.Now()

	if err := h.service.UpdateDriverBehaviorAnalytics(r.Context(), id, h.getOrgID(r), updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteDriverBehaviorAnalytics handles deleting driver behavior analytics
func (h *AnalyticsHandler) DeleteDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	if err := h.service.DeleteDriverBehaviorAnalytics(r.Context(), id, orgID); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListDriverBehaviorAnalytics handles listing driver behavior analytics with filtering
func (h *AnalyticsHandler) ListDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	orgIDStr := query.Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		orgID = h.getOrgID(r)
	}

	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}

	var periodStart, periodEnd *time.Time

	if periodStartStr := query.Get("periodStart"); periodStartStr != "" {
		date, err := time.Parse("2006-01-02", periodStartStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid period start date format, expected YYYY-MM-DD", err)
			return
		}
		periodStart = &date
	}

	if periodEndStr := query.Get("periodEnd"); periodEndStr != "" {
		date, err := time.Parse("2006-01-02", periodEndStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid period end date format, expected YYYY-MM-DD", err)
			return
		}
		periodEnd = &date
	}

	limit, offset := parsePaginationParams(query)
	cursor := query.Get("cursor")

	analyticsList, totalCount, err := h.service.ListDriverBehaviorAnalytics(r.Context(), orgID, periodStart, periodEnd, cursor, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	page := offset/limit + 1
	totalPages := (totalCount + limit - 1) / limit

	var nextCursor string
	if len(analyticsList) == limit && len(analyticsList) > 0 {
		last := analyticsList[len(analyticsList)-1]
		nextCursor = utils.EncodeCursor(last.AnalysisPeriodStart, last.ID)
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data": analyticsList,
		"meta": map[string]interface{}{
			"total":       totalCount,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
			"next_cursor": nextCursor,
		},
	})
}

// CreatePredictiveMaintenanceAnalytics handles creating new predictive maintenance analytics
func (h *AnalyticsHandler) CreatePredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	var analytics models.PredictiveMaintenanceAnalytics
	if err := json.NewDecoder(r.Body).Decode(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate using validator
	if err := ValidateRequest(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	// Set ID and timestamps
	analytics.ID = uuid.New().String()
	analytics.CreatedAt = time.Now()
	analytics.UpdatedAt = time.Now()
	analytics.GeneratedAt = time.Now()

	if err := h.service.CreatePredictiveMaintenanceAnalytics(r.Context(), &analytics); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, analytics)
}

// GetPredictiveMaintenanceAnalytics handles retrieving predictive maintenance analytics by ID
func (h *AnalyticsHandler) GetPredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	analytics, err := h.service.GetPredictiveMaintenanceAnalytics(r.Context(), id, orgID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analytics)
}

// GetPredictiveMaintenanceAnalyticsByVehicle handles retrieving predictive maintenance analytics by vehicle ID
func (h *AnalyticsHandler) GetPredictiveMaintenanceAnalyticsByVehicle(w http.ResponseWriter, r *http.Request) {
	vehicleIDStr := r.URL.Query().Get("vehicleId")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	orgID := h.getOrgID(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID", err)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.GetPredictiveMaintenanceAnalyticsByVehicle(r.Context(), vehicleID, orgID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// UpdatePredictiveMaintenanceAnalytics handles updating existing predictive maintenance analytics
func (h *AnalyticsHandler) UpdatePredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdatePredictiveMaintenanceAnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := ValidateRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error(), nil)
		return
	}

	updates := make(map[string]interface{})
	if req.NextServiceDue != nil {
		updates["next_service_due"] = *req.NextServiceDue
	}
	if req.ServicePriority != nil {
		updates["service_priority"] = *req.ServicePriority
	}
	if req.CostEstimation != nil {
		updates["cost_estimation"] = *req.CostEstimation
	}
	if req.ReliabilityScore != nil {
		updates["reliability_score"] = *req.ReliabilityScore
	}
	if req.Recommendations != nil {
		updates["recommendations"] = req.Recommendations
	}
	updates["updated_at"] = time.Now()

	if err := h.service.UpdatePredictiveMaintenanceAnalytics(r.Context(), id, h.getOrgID(r), updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeletePredictiveMaintenanceAnalytics handles deleting predictive maintenance analytics
func (h *AnalyticsHandler) DeletePredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	orgID := h.getOrgID(r)

	if err := h.service.DeletePredictiveMaintenanceAnalytics(r.Context(), id, orgID); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListPredictiveMaintenanceAnalytics handles listing predictive maintenance analytics with filtering
func (h *AnalyticsHandler) ListPredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	orgIDStr := query.Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		orgID = h.getOrgID(r)
	}

	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid organization ID", nil)
		return
	}

	var vehicleID *uuid.UUID

	if vehicleIDStr := query.Get("vehicleId"); vehicleIDStr != "" {
		parsedUUID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID", err)
			return
		}
		vehicleID = &parsedUUID
	}

	limit, offset := parsePaginationParams(query)
	cursor := query.Get("cursor")

	analyticsList, totalCount, err := h.service.ListPredictiveMaintenanceAnalytics(r.Context(), orgID, vehicleID, cursor, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	page := offset/limit + 1
	totalPages := (totalCount + limit - 1) / limit

	var nextCursor string
	if len(analyticsList) == limit && len(analyticsList) > 0 {
		last := analyticsList[len(analyticsList)-1]
		nextCursor = utils.EncodeCursor(last.CreatedAt, last.ID)
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data": analyticsList,
		"meta": map[string]interface{}{
			"total":       totalCount,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
			"next_cursor": nextCursor,
		},
	})
}

// HealthCheck handles health check requests
func (h *AnalyticsHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.service.HealthCheck(r.Context()); err != nil {
		h.writeErrorResponse(w, http.StatusServiceUnavailable, "Service unavailable", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// No manual validation functions needed anymore as we use validator/v10

// ValidationError represents a validation error
type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}

// parseStringParam parses a string parameter, returning nil if empty
func parseStringParam(param string) *string {
	if param == "" {
		return nil
	}
	return &param
}

// parseBoolParam parses a boolean parameter, returning nil if empty
func parseBoolParam(param string) *bool {
	if param == "" {
		return nil
	}
	value, err := strconv.ParseBool(param)
	if err != nil {
		return nil
	}
	return &value
}

// parsePaginationParams parses limit and offset from query parameters
func parsePaginationParams(values map[string][]string) (int, int) {
	DefaultQueryLimit := 10
	limit := DefaultQueryLimit // default limit
	if limitStrs, ok := values["limit"]; ok && len(limitStrs) > 0 {
		limitStr := limitStrs[0]
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // default offset
	if offsetStrs, ok := values["offset"]; ok && len(offsetStrs) > 0 {
		offsetStr := offsetStrs[0]
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Limit maximum limit to prevent abuse
	MaxQueryLimit := 100
	if limit > MaxQueryLimit {
		limit = MaxQueryLimit
	}

	// Limit maximum offset to prevent performance issues (max 10,000 records)
	MaxQueryOffset := 10000
	if offset > MaxQueryOffset {
		offset = MaxQueryOffset
	}

	return limit, offset
}

// GetVehicleDailyStats handles retrieving daily stats for a vehicle
func (h *AnalyticsHandler) GetVehicleDailyStats(w http.ResponseWriter, r *http.Request) {
	vehicleIDStr := chi.URLParam(r, "vehicleID")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID", err)
		return
	}

	query := r.URL.Query()
	startDate, _ := time.Parse(time.RFC3339, query.Get("startDate"))
	endDate, _ := time.Parse(time.RFC3339, query.Get("endDate"))
	cursor := query.Get("cursor")
	limit, _ := strconv.Atoi(query.Get("limit"))

	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, 0, -7)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	orgID := h.getOrgID(r)
	stats, nextCursor, err := h.service.GetVehicleDailyStats(r.Context(), vehicleID, orgID, startDate, endDate, cursor, limit)
	if err != nil {
		if errors.Is(err, repository.ErrQueryTooExpensive) {
			h.writeErrorResponse(w, http.StatusBadRequest, "Query too expensive, please narrow date range", nil)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":   stats,
		"cursor": nextCursor,
	})
}

// GetDriverDailyStats handles retrieving daily stats for a driver
func (h *AnalyticsHandler) GetDriverDailyStats(w http.ResponseWriter, r *http.Request) {
	driverIDStr := chi.URLParam(r, "driverID")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID", err)
		return
	}

	query := r.URL.Query()
	startDate, _ := time.Parse(time.RFC3339, query.Get("startDate"))
	endDate, _ := time.Parse(time.RFC3339, query.Get("endDate"))
	cursor := query.Get("cursor")
	limit, _ := strconv.Atoi(query.Get("limit"))

	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, 0, -7)
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	orgID := h.getOrgID(r)
	stats, nextCursor, err := h.service.GetDriverDailyStats(r.Context(), driverID, orgID, startDate, endDate, cursor, limit)
	if err != nil {
		if errors.Is(err, repository.ErrQueryTooExpensive) {
			h.writeErrorResponse(w, http.StatusBadRequest, "Query too expensive, please narrow date range", nil)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":   stats,
		"cursor": nextCursor,
	})
}

// GetFleetSummary handles the retrieval of fleet-wide aggregate stats
func (h *AnalyticsHandler) GetFleetSummary(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	startDate, _ := time.Parse(time.RFC3339, query.Get("startDate"))
	endDate, _ := time.Parse(time.RFC3339, query.Get("endDate"))

	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, 0, -7) // Last 7 days
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	orgID := h.getOrgID(r)
	summary, err := h.service.GetFleetAnalyticsByPeriod(r.Context(), orgID, startDate, endDate)
	if err != nil {
		if errors.Is(err, repository.ErrQueryTooExpensive) {
			h.writeErrorResponse(w, http.StatusBadRequest, "Query too expensive, please narrow date range", nil)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, summary)
}

// GetFleetHourlyStats handles retrieving hourly stats for the fleet
func (h *AnalyticsHandler) GetFleetHourlyStats(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	startDate, _ := time.Parse(time.RFC3339, query.Get("startDate"))
	endDate, _ := time.Parse(time.RFC3339, query.Get("endDate"))
	cursor := query.Get("cursor")
	limit, _ := strconv.Atoi(query.Get("limit"))

	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, 0, -1) // Last 24 hours
	}
	if endDate.IsZero() {
		endDate = time.Now()
	}

	orgID := h.getOrgID(r)
	stats, nextCursor, err := h.service.GetFleetHourlyStats(r.Context(), orgID, startDate, endDate, cursor, limit)
	if err != nil {
		if errors.Is(err, repository.ErrQueryTooExpensive) {
			h.writeErrorResponse(w, http.StatusBadRequest, "Query too expensive, please narrow date range", nil)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":   stats,
		"cursor": nextCursor,
	})
}

func (h *AnalyticsHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error("Failed to encode JSON response", zap.Error(err))
		}
	}
}

func (h *AnalyticsHandler) writeErrorResponse(w http.ResponseWriter, status int, message string, err error) {
	resp := map[string]string{"error": message}
	if err != nil && status >= 500 {
		h.logger.Error(message, zap.Error(err))
	}
	h.writeJSONResponse(w, status, resp)
}

func (h *AnalyticsHandler) getOrgID(r *http.Request) uuid.UUID {
	orgID, ok := r.Context().Value(middleware.OrgIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return orgID
}

// Phase 8 Handlers

func (h *AnalyticsHandler) DeployModel(w http.ResponseWriter, r *http.Request) {
	var deployment models.ModelDeployment
	if err := json.NewDecoder(r.Body).Decode(&deployment); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid payload", err)
		return
	}
	deployment.OrganizationID = h.getOrgID(r)
	if err := h.service.DeployModel(r.Context(), &deployment); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Deployment failed", err)
		return
	}
	h.writeJSONResponse(w, http.StatusCreated, deployment)
}

func (h *AnalyticsHandler) CreateDriverRiskEvent(w http.ResponseWriter, r *http.Request) {
	var event models.DriverRiskEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid payload", err)
		return
	}
	event.OrganizationID = h.getOrgID(r)
	if err := h.service.CreateDriverRiskEvent(r.Context(), &event); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create risk event", err)
		return
	}
	h.writeJSONResponse(w, http.StatusCreated, event)
}

func (h *AnalyticsHandler) GetDriverRiskEvents(w http.ResponseWriter, r *http.Request) {
	driverIDStr := r.URL.Query().Get("driver_id")
	driverID, _ := uuid.Parse(driverIDStr)
	start, _ := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
	end, _ := time.Parse(time.RFC3339, r.URL.Query().Get("end"))

	events, err := h.service.GetDriverRiskEvents(r.Context(), driverID, h.getOrgID(r), start, end)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch events", err)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, events)
}

func (h *AnalyticsHandler) GetAnalyticsQuota(w http.ResponseWriter, r *http.Request) {
	quota, err := h.service.GetAnalyticsQuota(r.Context(), h.getOrgID(r))
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch quota", err)
		return
	}
	h.writeJSONResponse(w, http.StatusOK, quota)
}

func (h *AnalyticsHandler) RunReconciliation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	h.service.RunReconciliation(r.Context(), h.getOrgID(r), req.Start, req.End)
	h.writeJSONResponse(w, http.StatusAccepted, map[string]string{"status": "triggered"})
}

func (h *AnalyticsHandler) RunHistoricalBackfill(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	h.service.RunHistoricalBackfill(r.Context(), h.getOrgID(r), req.Start, req.End)
	h.writeJSONResponse(w, http.StatusAccepted, map[string]string{"status": "triggered"})
}
