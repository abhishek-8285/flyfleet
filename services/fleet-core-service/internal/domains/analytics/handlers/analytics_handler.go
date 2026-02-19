package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AnalyticsHandler handles HTTP requests for analytics operations
type AnalyticsHandler struct {
	service service.AnalyticsService
	logger  *zap.Logger
}

// NewAnalyticsHandler creates a new instance of AnalyticsHandler
func NewAnalyticsHandler(service service.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers the analytics routes
func (h *AnalyticsHandler) RegisterRoutes(r chi.Router) {
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

	// Legacy/Placeholder endpoints
	r.Get("/vehicle-analytics", h.GetVehicleAnalytics)
	r.Get("/driver-analytics", h.GetDriverAnalytics)
	r.Get("/trip-analytics", h.GetTripAnalytics)
	r.Get("/forecast-analytics", h.GetForecastAnalytics)

	// Health Check
	r.Get("/health", h.HealthCheck)
}

// CreatePredictiveModel handles the creation of a new predictive model
func (h *AnalyticsHandler) CreatePredictiveModel(w http.ResponseWriter, r *http.Request) {
	var model models.PredictiveModel
	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if err := validatePredictiveModel(&model); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

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

	model, err := h.service.GetPredictiveModel(r.Context(), id)
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

	if name == "" || version == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Name and version are required", nil)
		return
	}

	model, err := h.service.GetPredictiveModelByVersion(r.Context(), name, version)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, model)
}

// UpdatePredictiveModel handles updating an existing predictive model
func (h *AnalyticsHandler) UpdatePredictiveModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate updates
	if err := validatePredictiveModelUpdates(updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdatePredictiveModel(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeletePredictiveModel handles deleting a predictive model
func (h *AnalyticsHandler) DeletePredictiveModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeletePredictiveModel(r.Context(), id); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// ListPredictiveModels handles listing predictive models with filtering
func (h *AnalyticsHandler) ListPredictiveModels(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	modelType := parseStringParam(query.Get("type"))
	status := parseStringParam(query.Get("status"))

	limit, offset := parsePaginationParams(query)

	models, totalCount, err := h.service.ListPredictiveModels(r.Context(), modelType, status, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        models,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreatePredictionResult handles creating a new prediction result
func (h *AnalyticsHandler) CreatePredictionResult(w http.ResponseWriter, r *http.Request) {
	var result models.PredictionResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if err := validatePredictionResult(&result); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

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

	result, err := h.service.GetPredictionResult(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// GetPredictionResultsByModel handles retrieving prediction results by model ID
func (h *AnalyticsHandler) GetPredictionResultsByModel(w http.ResponseWriter, r *http.Request) {
	modelID := chi.URLParam(r, "modelId")

	limit, offset := parsePaginationParams(r.URL.Query())

	results, totalCount, err := h.service.GetPredictionResultsByModel(r.Context(), modelID, limit, offset)
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

	results, totalCount, err := h.service.ListPredictionResults(r.Context(), orgID, modelID, anomalyDetected, limit, offset)
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

// CreateAnalyticsJob handles creating a new analytics job
func (h *AnalyticsHandler) CreateAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	var job models.AnalyticsJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if err := validateAnalyticsJob(&job); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

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

	job, err := h.service.GetAnalyticsJob(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, job)
}

// UpdateAnalyticsJob handles updating an existing analytics job
func (h *AnalyticsHandler) UpdateAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate updates
	if err := validateAnalyticsJobUpdates(updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdateAnalyticsJob(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteAnalyticsJob handles deleting an analytics job
func (h *AnalyticsHandler) DeleteAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteAnalyticsJob(r.Context(), id); err != nil {
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

	jobs, totalCount, err := h.service.ListAnalyticsJobs(r.Context(), jobType, status, orgID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data":        jobs,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateFleetAnalytics handles creating new fleet analytics
func (h *AnalyticsHandler) CreateFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	var analytics models.FleetAnalytics
	if err := json.NewDecoder(r.Body).Decode(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if err := validateFleetAnalytics(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
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

	analytics, err := h.service.GetFleetAnalytics(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analytics)
}

// GetFleetAnalyticsByPeriod handles retrieving fleet analytics by organization and date period
func (h *AnalyticsHandler) GetFleetAnalyticsByPeriod(w http.ResponseWriter, r *http.Request) {
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

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate updates
	if err := validateFleetAnalyticsUpdates(updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdateFleetAnalytics(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteFleetAnalytics handles deleting fleet analytics
func (h *AnalyticsHandler) DeleteFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteFleetAnalytics(r.Context(), id); err != nil {
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

	analytics, totalCount, err := h.service.ListFleetAnalytics(r.Context(), orgID, startDate, endDate, limit, offset)
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

// CreateDriverBehaviorAnalytics handles creating new driver behavior analytics
func (h *AnalyticsHandler) CreateDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	var analytics models.DriverBehaviorAnalytics
	if err := json.NewDecoder(r.Body).Decode(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate required fields
	if err := validateDriverBehaviorAnalytics(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

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

	analytics, err := h.service.GetDriverBehaviorAnalytics(r.Context(), id)
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
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid driver ID", err)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.GetDriverBehaviorAnalyticsByDriver(r.Context(), driverID, limit, offset)
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

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate updates
	if err := validateDriverBehaviorAnalyticsUpdates(updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdateDriverBehaviorAnalytics(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteDriverBehaviorAnalytics handles deleting driver behavior analytics
func (h *AnalyticsHandler) DeleteDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteDriverBehaviorAnalytics(r.Context(), id); err != nil {
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

	analytics, totalCount, err := h.service.ListDriverBehaviorAnalytics(r.Context(), orgID, periodStart, periodEnd, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	page := offset/limit + 1
	totalPages := (totalCount + limit - 1) / limit

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data": analytics,
		"meta": map[string]interface{}{
			"total":       totalCount,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
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

	// Validate required fields
	if err := validatePredictiveMaintenanceAnalytics(&analytics); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
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

	analytics, err := h.service.GetPredictiveMaintenanceAnalytics(r.Context(), id)
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
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid vehicle ID", err)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.GetPredictiveMaintenanceAnalyticsByVehicle(r.Context(), vehicleID, limit, offset)
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

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate updates
	if err := validatePredictiveMaintenanceAnalyticsUpdates(updates); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.UpdatePredictiveMaintenanceAnalytics(r.Context(), id, updates); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeletePredictiveMaintenanceAnalytics handles deleting predictive maintenance analytics
func (h *AnalyticsHandler) DeletePredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeletePredictiveMaintenanceAnalytics(r.Context(), id); err != nil {
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

	analytics, totalCount, err := h.service.ListPredictiveMaintenanceAnalytics(r.Context(), orgID, vehicleID, limit, offset)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	page := offset/limit + 1
	totalPages := (totalCount + limit - 1) / limit

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"data": analytics,
		"meta": map[string]interface{}{
			"total":       totalCount,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
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

// Validation functions
func validatePredictiveModel(model *models.PredictiveModel) error {
	if model.Name == "" {
		return ValidationError("model name is required")
	}
	if model.Version == "" {
		return ValidationError("model version is required")
	}
	if model.ModelType == "" {
		return ValidationError("model type is required")
	}
	if model.Status == "" {
		return ValidationError("model status is required")
	}
	if model.TrainedAt.IsZero() {
		return ValidationError("trained at timestamp is required")
	}
	return nil
}

func validatePredictiveModelUpdates(updates map[string]interface{}) error {
	if len(updates) == 0 {
		return ValidationError("updates are required")
	}
	return nil
}

func validatePredictionResult(result *models.PredictionResult) error {
	if result.ModelID == "" {
		return ValidationError("model ID is required")
	}
	if result.OrganizationID == uuid.Nil {
		return ValidationError("organization ID is required")
	}
	if result.ConfidenceScore < 0 || result.ConfidenceScore > 1 {
		return ValidationError("confidence score must be between 0 and 1")
	}
	return nil
}

func validateAnalyticsJob(job *models.AnalyticsJob) error {
	if job.Name == "" {
		return ValidationError("job name is required")
	}
	if job.Type == "" {
		return ValidationError("job type is required")
	}
	if job.Status == "" {
		return ValidationError("job status is required")
	}
	if job.OrganizationID == uuid.Nil {
		return ValidationError("organization ID is required")
	}
	if job.CreatedBy == uuid.Nil {
		return ValidationError("created by ID is required")
	}
	return nil
}

func validateAnalyticsJobUpdates(updates map[string]interface{}) error {
	if len(updates) == 0 {
		return ValidationError("updates are required")
	}
	return nil
}

func validateFleetAnalytics(analytics *models.FleetAnalytics) error {
	if analytics.OrganizationID == uuid.Nil {
		return ValidationError("organization ID is required")
	}
	if analytics.StartDate.IsZero() || analytics.EndDate.IsZero() {
		return ValidationError("start and end dates are required")
	}
	if analytics.StartDate.After(analytics.EndDate) {
		return ValidationError("start date cannot be after end date")
	}
	return nil
}

func validateFleetAnalyticsUpdates(updates map[string]interface{}) error {
	if len(updates) == 0 {
		return ValidationError("updates are required")
	}
	return nil
}

func validateDriverBehaviorAnalytics(analytics *models.DriverBehaviorAnalytics) error {
	if analytics.DriverID == uuid.Nil {
		return ValidationError("driver ID is required")
	}
	if analytics.OrganizationID == uuid.Nil {
		return ValidationError("organization ID is required")
	}
	if analytics.AnalysisPeriodStart.IsZero() || analytics.AnalysisPeriodEnd.IsZero() {
		return ValidationError("analysis period start and end dates are required")
	}
	if analytics.AnalysisPeriodStart.After(analytics.AnalysisPeriodEnd) {
		return ValidationError("analysis period start cannot be after end date")
	}
	return nil
}

func validateDriverBehaviorAnalyticsUpdates(updates map[string]interface{}) error {
	if len(updates) == 0 {
		return ValidationError("updates are required")
	}
	return nil
}

func validatePredictiveMaintenanceAnalytics(analytics *models.PredictiveMaintenanceAnalytics) error {
	if analytics.VehicleID == uuid.Nil {
		return ValidationError("vehicle ID is required")
	}
	if analytics.OrganizationID == uuid.Nil {
		return ValidationError("organization ID is required")
	}
	if analytics.AnalysisDate.IsZero() {
		return ValidationError("analysis date is required")
	}
	return nil
}

func validatePredictiveMaintenanceAnalyticsUpdates(updates map[string]interface{}) error {
	if len(updates) == 0 {
		return ValidationError("updates are required")
	}
	return nil
}

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
		limit = 100
	}

	return limit, offset
}

func (h *AnalyticsHandler) GetVehicleAnalytics(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Vehicle analytics endpoint",
		"data":    []interface{}{},
	})
}

func (h *AnalyticsHandler) GetDriverAnalytics(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Driver analytics endpoint",
		"data":    []interface{}{},
	})
}

func (h *AnalyticsHandler) GetTripAnalytics(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Trip analytics endpoint",
		"data":    []interface{}{},
	})
}

func (h *AnalyticsHandler) GetForecastAnalytics(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Forecast analytics endpoint",
		"data":    []interface{}{},
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
