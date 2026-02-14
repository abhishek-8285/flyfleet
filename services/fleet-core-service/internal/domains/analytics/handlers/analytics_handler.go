package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AnalyticsHandler handles HTTP requests for analytics operations
type AnalyticsHandler struct {
	service service.AnalyticsService
}

// NewAnalyticsHandler creates a new instance of AnalyticsHandler
func NewAnalyticsHandler(service service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
	}
}

// RegisterRoutes registers the analytics routes with the provided router group
func (h *AnalyticsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	analytics := rg.Group("/analytics")
	{
		// Predictive Models
		analytics.POST("/predictive-models", h.CreatePredictiveModel)
		analytics.GET("/predictive-models/:id", h.GetPredictiveModel)
		analytics.PUT("/predictive-models/:id", h.UpdatePredictiveModel)
		analytics.DELETE("/predictive-models/:id", h.DeletePredictiveModel)
		analytics.GET("/predictive-models", h.ListPredictiveModels)
		analytics.GET("/predictive-models/version", h.GetPredictiveModelByVersion)

		// Prediction Results
		analytics.POST("/predictions", h.CreatePredictionResult)
		analytics.GET("/predictions/:id", h.GetPredictionResult)
		analytics.GET("/predictions/model/:modelId", h.GetPredictionResultsByModel)
		analytics.GET("/predictions/org", h.GetPredictionResultsByOrg)
		analytics.GET("/predictions", h.ListPredictionResults)

		// Analytics Jobs
		analytics.POST("/jobs", h.CreateAnalyticsJob)
		analytics.GET("/jobs/:id", h.GetAnalyticsJob)
		analytics.PUT("/jobs/:id", h.UpdateAnalyticsJob)
		analytics.DELETE("/jobs/:id", h.DeleteAnalyticsJob)
		analytics.GET("/jobs", h.ListAnalyticsJobs)

		// Fleet Analytics
		analytics.POST("/fleet", h.CreateFleetAnalytics)
		analytics.GET("/fleet/:id", h.GetFleetAnalytics)
		analytics.GET("/fleet/period", h.GetFleetAnalyticsByPeriod)
		analytics.PUT("/fleet/:id", h.UpdateFleetAnalytics)
		analytics.DELETE("/fleet/:id", h.DeleteFleetAnalytics)
		analytics.GET("/fleet", h.ListFleetAnalytics)

		// Driver Behavior
		analytics.POST("/driver-behavior", h.CreateDriverBehaviorAnalytics)
		analytics.GET("/driver-behavior/:id", h.GetDriverBehaviorAnalytics)
		analytics.GET("/driver-behavior/driver", h.GetDriverBehaviorAnalyticsByDriver)
		analytics.GET("/driver-behavior/org", h.GetDriverBehaviorAnalyticsByOrg)
		analytics.PUT("/driver-behavior/:id", h.UpdateDriverBehaviorAnalytics)
		analytics.DELETE("/driver-behavior/:id", h.DeleteDriverBehaviorAnalytics)
		analytics.GET("/driver-behavior", h.ListDriverBehaviorAnalytics)

		// Predictive Maintenance
		analytics.POST("/predictive-maintenance", h.CreatePredictiveMaintenanceAnalytics)
		analytics.GET("/predictive-maintenance/:id", h.GetPredictiveMaintenanceAnalytics)
		analytics.GET("/predictive-maintenance/vehicle", h.GetPredictiveMaintenanceAnalyticsByVehicle)
		analytics.PUT("/predictive-maintenance/:id", h.UpdatePredictiveMaintenanceAnalytics)
		analytics.DELETE("/predictive-maintenance/:id", h.DeletePredictiveMaintenanceAnalytics)
		analytics.GET("/predictive-maintenance", h.ListPredictiveMaintenanceAnalytics)


		// Legacy/Placeholder endpoints
		analytics.GET("/vehicle-analytics", h.GetVehicleAnalytics)
		analytics.GET("/driver-analytics", h.GetDriverAnalytics)
		analytics.GET("/trip-analytics", h.GetTripAnalytics)
		analytics.GET("/forecast-analytics", h.GetForecastAnalytics)

		// Health Check
		analytics.GET("/health", h.HealthCheck)
	}
}

// CreatePredictiveModel handles the creation of a new predictive model
func (h *AnalyticsHandler) CreatePredictiveModel(c *gin.Context) {
	var model models.PredictiveModel
	if err := c.ShouldBindJSON(&model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if err := validatePredictiveModel(&model); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set ID and timestamps
	model.ID = uuid.New().String()
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()

	if err := h.service.CreatePredictiveModel(c.Request.Context(), &model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model)
}

// GetPredictiveModel handles retrieving a predictive model by ID
func (h *AnalyticsHandler) GetPredictiveModel(c *gin.Context) {
	id := c.Param("id")

	model, err := h.service.GetPredictiveModel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model)
}

// GetPredictiveModelByVersion handles retrieving a predictive model by name and version
func (h *AnalyticsHandler) GetPredictiveModelByVersion(c *gin.Context) {
	name := c.Query("name")
	version := c.Query("version")

	if name == "" || version == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and version are required"})
		return
	}

	model, err := h.service.GetPredictiveModelByVersion(c.Request.Context(), name, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model)
}

// UpdatePredictiveModel handles updating an existing predictive model
func (h *AnalyticsHandler) UpdatePredictiveModel(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate updates
	if err := validatePredictiveModelUpdates(updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdatePredictiveModel(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeletePredictiveModel handles deleting a predictive model
func (h *AnalyticsHandler) DeletePredictiveModel(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeletePredictiveModel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListPredictiveModels handles listing predictive models with filtering
func (h *AnalyticsHandler) ListPredictiveModels(c *gin.Context) {
	modelType := parseStringParam(c.Query("type"))
	status := parseStringParam(c.Query("status"))

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	models, totalCount, err := h.service.ListPredictiveModels(c.Request.Context(), modelType, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        models,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreatePredictionResult handles creating a new prediction result
func (h *AnalyticsHandler) CreatePredictionResult(c *gin.Context) {
	var result models.PredictionResult
	if err := c.ShouldBindJSON(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if err := validatePredictionResult(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set ID and timestamps
	result.ID = uuid.New().String()
	result.CreatedAt = time.Now()
	result.UpdatedAt = time.Now()
	result.PredictedAt = time.Now()

	if err := h.service.CreatePredictionResult(c.Request.Context(), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetPredictionResult handles retrieving a prediction result by ID
func (h *AnalyticsHandler) GetPredictionResult(c *gin.Context) {
	id := c.Param("id")

	result, err := h.service.GetPredictionResult(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPredictionResultsByModel handles retrieving prediction results by model ID
func (h *AnalyticsHandler) GetPredictionResultsByModel(c *gin.Context) {
	modelID := c.Param("modelId")

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	results, totalCount, err := h.service.GetPredictionResultsByModel(c.Request.Context(), modelID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        results,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetPredictionResultsByOrg handles retrieving prediction results by organization
func (h *AnalyticsHandler) GetPredictionResultsByOrg(c *gin.Context) {
	orgIDStr := c.Query("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	results, totalCount, err := h.service.GetPredictionResultsByOrg(c.Request.Context(), orgID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        results,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// ListPredictionResults handles listing prediction results with filtering
func (h *AnalyticsHandler) ListPredictionResults(c *gin.Context) {
	orgIDStr := c.Query("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	modelID := parseStringParam(c.Query("modelId"))
	anomalyDetected := parseBoolParam(c.Query("anomalyDetected"))

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	results, totalCount, err := h.service.ListPredictionResults(c.Request.Context(), orgID, modelID, anomalyDetected, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        results,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateAnalyticsJob handles creating a new analytics job
func (h *AnalyticsHandler) CreateAnalyticsJob(c *gin.Context) {
	var job models.AnalyticsJob
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if err := validateAnalyticsJob(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set ID and timestamps
	job.ID = uuid.New().String()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	if err := h.service.CreateAnalyticsJob(c.Request.Context(), &job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// GetAnalyticsJob handles retrieving an analytics job by ID
func (h *AnalyticsHandler) GetAnalyticsJob(c *gin.Context) {
	id := c.Param("id")

	job, err := h.service.GetAnalyticsJob(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

// UpdateAnalyticsJob handles updating an existing analytics job
func (h *AnalyticsHandler) UpdateAnalyticsJob(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate updates
	if err := validateAnalyticsJobUpdates(updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateAnalyticsJob(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteAnalyticsJob handles deleting an analytics job
func (h *AnalyticsHandler) DeleteAnalyticsJob(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteAnalyticsJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListAnalyticsJobs handles listing analytics jobs with filtering
func (h *AnalyticsHandler) ListAnalyticsJobs(c *gin.Context) {
	jobType := parseStringParam(c.Query("jobType"))
	status := parseStringParam(c.Query("status"))

	orgIDStr := c.Query("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	jobs, totalCount, err := h.service.ListAnalyticsJobs(c.Request.Context(), jobType, status, orgID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        jobs,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateFleetAnalytics handles creating new fleet analytics
func (h *AnalyticsHandler) CreateFleetAnalytics(c *gin.Context) {
	var analytics models.FleetAnalytics
	if err := c.ShouldBindJSON(&analytics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if err := validateFleetAnalytics(&analytics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set ID and timestamps
	analytics.ID = uuid.New().String()
	analytics.CreatedAt = time.Now()
	analytics.UpdatedAt = time.Now()
	analytics.GeneratedAt = time.Now()

	if err := h.service.CreateFleetAnalytics(c.Request.Context(), &analytics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, analytics)
}

// GetFleetAnalytics handles retrieving fleet analytics by ID
func (h *AnalyticsHandler) GetFleetAnalytics(c *gin.Context) {
	id := c.Param("id")

	analytics, err := h.service.GetFleetAnalytics(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetFleetAnalyticsByPeriod handles retrieving fleet analytics by organization and date period
func (h *AnalyticsHandler) GetFleetAnalyticsByPeriod(c *gin.Context) {
	orgIDStr := c.Query("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format, expected YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format, expected YYYY-MM-DD"})
		return
	}

	analytics, err := h.service.GetFleetAnalyticsByPeriod(c.Request.Context(), orgID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// UpdateFleetAnalytics handles updating existing fleet analytics
func (h *AnalyticsHandler) UpdateFleetAnalytics(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate updates
	if err := validateFleetAnalyticsUpdates(updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateFleetAnalytics(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteFleetAnalytics handles deleting fleet analytics
func (h *AnalyticsHandler) DeleteFleetAnalytics(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteFleetAnalytics(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListFleetAnalytics handles listing fleet analytics with filtering
func (h *AnalyticsHandler) ListFleetAnalytics(c *gin.Context) {
	orgIDStr := c.Query("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var startDate, endDate *time.Time

	if startDateStr := c.Query("startDate"); startDateStr != "" {
		date, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format, expected YYYY-MM-DD"})
			return
		}
		startDate = &date
	}

	if endDateStr := c.Query("endDate"); endDateStr != "" {
		date, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format, expected YYYY-MM-DD"})
			return
		}
		endDate = &date
	}

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	analytics, totalCount, err := h.service.ListFleetAnalytics(c.Request.Context(), orgID, startDate, endDate, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateDriverBehaviorAnalytics handles creating new driver behavior analytics
func (h *AnalyticsHandler) CreateDriverBehaviorAnalytics(c *gin.Context) {
	var analytics models.DriverBehaviorAnalytics
	if err := c.ShouldBindJSON(&analytics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if err := validateDriverBehaviorAnalytics(&analytics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set ID and timestamps
	analytics.ID = uuid.New().String()
	analytics.CreatedAt = time.Now()
	analytics.UpdatedAt = time.Now()
	analytics.GeneratedAt = time.Now()

	if err := h.service.CreateDriverBehaviorAnalytics(c.Request.Context(), &analytics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, analytics)
}

// GetDriverBehaviorAnalytics handles retrieving driver behavior analytics by ID
func (h *AnalyticsHandler) GetDriverBehaviorAnalytics(c *gin.Context) {
	id := c.Param("id")

	analytics, err := h.service.GetDriverBehaviorAnalytics(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetDriverBehaviorAnalyticsByDriver handles retrieving driver behavior analytics by driver ID
func (h *AnalyticsHandler) GetDriverBehaviorAnalyticsByDriver(c *gin.Context) {
	driverIDStr := c.Query("driverId")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid driver ID"})
		return
	}

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	analytics, totalCount, err := h.service.GetDriverBehaviorAnalyticsByDriver(c.Request.Context(), driverID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetDriverBehaviorAnalyticsByOrg handles retrieving driver behavior analytics by organization
func (h *AnalyticsHandler) GetDriverBehaviorAnalyticsByOrg(c *gin.Context) {
	orgIDStr := c.Query("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	analytics, totalCount, err := h.service.GetDriverBehaviorAnalyticsByOrg(c.Request.Context(), orgID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// UpdateDriverBehaviorAnalytics handles updating existing driver behavior analytics
func (h *AnalyticsHandler) UpdateDriverBehaviorAnalytics(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate updates
	if err := validateDriverBehaviorAnalyticsUpdates(updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateDriverBehaviorAnalytics(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeleteDriverBehaviorAnalytics handles deleting driver behavior analytics
func (h *AnalyticsHandler) DeleteDriverBehaviorAnalytics(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteDriverBehaviorAnalytics(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListDriverBehaviorAnalytics handles listing driver behavior analytics with filtering
func (h *AnalyticsHandler) ListDriverBehaviorAnalytics(c *gin.Context) {
	orgIDStr := c.Query("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var periodStart, periodEnd *time.Time

	if periodStartStr := c.Query("periodStart"); periodStartStr != "" {
		date, err := time.Parse("2006-01-02", periodStartStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period start date format, expected YYYY-MM-DD"})
			return
		}
		periodStart = &date
	}

	if periodEndStr := c.Query("periodEnd"); periodEndStr != "" {
		date, err := time.Parse("2006-01-02", periodEndStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid period end date format, expected YYYY-MM-DD"})
			return
		}
		periodEnd = &date
	}

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	analytics, totalCount, err := h.service.ListDriverBehaviorAnalytics(c.Request.Context(), orgID, periodStart, periodEnd, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreatePredictiveMaintenanceAnalytics handles creating new predictive maintenance analytics
func (h *AnalyticsHandler) CreatePredictiveMaintenanceAnalytics(c *gin.Context) {
	var analytics models.PredictiveMaintenanceAnalytics
	if err := c.ShouldBindJSON(&analytics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if err := validatePredictiveMaintenanceAnalytics(&analytics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set ID and timestamps
	analytics.ID = uuid.New().String()
	analytics.CreatedAt = time.Now()
	analytics.UpdatedAt = time.Now()
	analytics.GeneratedAt = time.Now()

	if err := h.service.CreatePredictiveMaintenanceAnalytics(c.Request.Context(), &analytics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, analytics)
}

// GetPredictiveMaintenanceAnalytics handles retrieving predictive maintenance analytics by ID
func (h *AnalyticsHandler) GetPredictiveMaintenanceAnalytics(c *gin.Context) {
	id := c.Param("id")

	analytics, err := h.service.GetPredictiveMaintenanceAnalytics(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetPredictiveMaintenanceAnalyticsByVehicle handles retrieving predictive maintenance analytics by vehicle ID
func (h *AnalyticsHandler) GetPredictiveMaintenanceAnalyticsByVehicle(c *gin.Context) {
	vehicleIDStr := c.Query("vehicleId")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID"})
		return
	}

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	analytics, totalCount, err := h.service.GetPredictiveMaintenanceAnalyticsByVehicle(c.Request.Context(), vehicleID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// UpdatePredictiveMaintenanceAnalytics handles updating existing predictive maintenance analytics
func (h *AnalyticsHandler) UpdatePredictiveMaintenanceAnalytics(c *gin.Context) {
	id := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate updates
	if err := validatePredictiveMaintenanceAnalyticsUpdates(updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdatePredictiveMaintenanceAnalytics(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeletePredictiveMaintenanceAnalytics handles deleting predictive maintenance analytics
func (h *AnalyticsHandler) DeletePredictiveMaintenanceAnalytics(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeletePredictiveMaintenanceAnalytics(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// ListPredictiveMaintenanceAnalytics handles listing predictive maintenance analytics with filtering
func (h *AnalyticsHandler) ListPredictiveMaintenanceAnalytics(c *gin.Context) {
	orgIDStr := c.Query("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var vehicleID *uuid.UUID

	if vehicleIDStr := c.Query("vehicleId"); vehicleIDStr != "" {
		parsedUUID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vehicle ID"})
			return
		}
		vehicleID = &parsedUUID
	}

	limit, offset := parsePaginationParams(c.Request.URL.Query())

	analytics, totalCount, err := h.service.ListPredictiveMaintenanceAnalytics(c.Request.Context(), orgID, vehicleID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	})
}

// HealthCheck handles health check requests
func (h *AnalyticsHandler) HealthCheck(c *gin.Context) {
	if err := h.service.HealthCheck(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service unavailable"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
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

func (h *AnalyticsHandler) GetVehicleAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle analytics endpoint",
		"data":    []interface{}{},
	})
}

func (h *AnalyticsHandler) GetDriverAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Driver analytics endpoint",
		"data":    []interface{}{},
	})
}

func (h *AnalyticsHandler) GetTripAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Trip analytics endpoint",
		"data":    []interface{}{},
	})
}

func (h *AnalyticsHandler) GetForecastAnalytics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Forecast analytics endpoint",
		"data":    []interface{}{},
	})
}
