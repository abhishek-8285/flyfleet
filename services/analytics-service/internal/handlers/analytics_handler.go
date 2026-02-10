package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/flyfleet/analytics-service/internal/models"
	"github.com/flyfleet/analytics-service/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

// CreatePredictiveModel handles the creation of a new predictive model
func (h *AnalyticsHandler) CreatePredictiveModel(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var model models.PredictiveModel
	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validatePredictiveModel(&model); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set ID and timestamps
	model.ID = uuid.New().String()
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()

	if err := h.service.CreatePredictiveModel(ctx, &model); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model)
}

// GetPredictiveModel handles retrieving a predictive model by ID
func (h *AnalyticsHandler) GetPredictiveModel(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	model, err := h.service.GetPredictiveModel(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model)
}

// GetPredictiveModelByVersion handles retrieving a predictive model by name and version
func (h *AnalyticsHandler) GetPredictiveModelByVersion(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	name := r.URL.Query().Get("name")
	version := r.URL.Query().Get("version")

	if name == "" || version == "" {
		http.Error(w, "Name and version are required", http.StatusBadRequest)
		return
	}

	model, err := h.service.GetPredictiveModelByVersion(ctx, name, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model)
}

// UpdatePredictiveModel handles updating an existing predictive model
func (h *AnalyticsHandler) UpdatePredictiveModel(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate updates
	if err := validatePredictiveModelUpdates(updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdatePredictiveModel(ctx, id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeletePredictiveModel handles deleting a predictive model
func (h *AnalyticsHandler) DeletePredictiveModel(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeletePredictiveModel(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListPredictiveModels handles listing predictive models with filtering
func (h *AnalyticsHandler) ListPredictiveModels(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	modelType := parseStringParam(r.URL.Query().Get("type"))
	status := parseStringParam(r.URL.Query().Get("status"))

	limit, offset := parsePaginationParams(r.URL.Query())

	models, totalCount, err := h.service.ListPredictiveModels(ctx, modelType, status, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        models,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreatePredictionResult handles creating a new prediction result
func (h *AnalyticsHandler) CreatePredictionResult(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var result models.PredictionResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validatePredictionResult(&result); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set ID and timestamps
	result.ID = uuid.New().String()
	result.CreatedAt = time.Now()
	result.UpdatedAt = time.Now()
	result.PredictedAt = time.Now()

	if err := h.service.CreatePredictionResult(ctx, &result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// GetPredictionResult handles retrieving a prediction result by ID
func (h *AnalyticsHandler) GetPredictionResult(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	result, err := h.service.GetPredictionResult(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetPredictionResultsByModel handles retrieving prediction results by model ID
func (h *AnalyticsHandler) GetPredictionResultsByModel(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	modelID := vars["modelId"]

	limit, offset := parsePaginationParams(r.URL.Query())

	results, totalCount, err := h.service.GetPredictionResultsByModel(ctx, modelID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        results,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPredictionResultsByOrg handles retrieving prediction results by organization
func (h *AnalyticsHandler) GetPredictionResultsByOrg(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	orgIDStr := r.URL.Query().Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	results, totalCount, err := h.service.GetPredictionResultsByOrg(ctx, orgID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        results,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListPredictionResults handles listing prediction results with filtering
func (h *AnalyticsHandler) ListPredictionResults(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	orgIDStr := r.URL.Query().Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	modelID := parseStringParam(r.URL.Query().Get("modelId"))
	anomalyDetected := parseBoolParam(r.URL.Query().Get("anomalyDetected"))

	limit, offset := parsePaginationParams(r.URL.Query())

	results, totalCount, err := h.service.ListPredictionResults(ctx, orgID, modelID, anomalyDetected, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        results,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateAnalyticsJob handles creating a new analytics job
func (h *AnalyticsHandler) CreateAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var job models.AnalyticsJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validateAnalyticsJob(&job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set ID and timestamps
	job.ID = uuid.New().String()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	if err := h.service.CreateAnalyticsJob(ctx, &job); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

// GetAnalyticsJob handles retrieving an analytics job by ID
func (h *AnalyticsHandler) GetAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	job, err := h.service.GetAnalyticsJob(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// UpdateAnalyticsJob handles updating an existing analytics job
func (h *AnalyticsHandler) UpdateAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate updates
	if err := validateAnalyticsJobUpdates(updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateAnalyticsJob(ctx, id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteAnalyticsJob handles deleting an analytics job
func (h *AnalyticsHandler) DeleteAnalyticsJob(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteAnalyticsJob(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListAnalyticsJobs handles listing analytics jobs with filtering
func (h *AnalyticsHandler) ListAnalyticsJobs(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	jobType := parseStringParam(r.URL.Query().Get("jobType"))
	status := parseStringParam(r.URL.Query().Get("status"))

	orgIDStr := r.URL.Query().Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	jobs, totalCount, err := h.service.ListAnalyticsJobs(ctx, jobType, status, orgID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        jobs,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateFleetAnalytics handles creating new fleet analytics
func (h *AnalyticsHandler) CreateFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var analytics models.FleetAnalytics
	if err := json.NewDecoder(r.Body).Decode(&analytics); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validateFleetAnalytics(&analytics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set ID and timestamps
	analytics.ID = uuid.New().String()
	analytics.CreatedAt = time.Now()
	analytics.UpdatedAt = time.Now()
	analytics.GeneratedAt = time.Now()

	if err := h.service.CreateFleetAnalytics(ctx, &analytics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(analytics)
}

// GetFleetAnalytics handles retrieving fleet analytics by ID
func (h *AnalyticsHandler) GetFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	analytics, err := h.service.GetFleetAnalytics(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// GetFleetAnalyticsByPeriod handles retrieving fleet analytics by organization and date period
func (h *AnalyticsHandler) GetFleetAnalyticsByPeriod(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	orgIDStr := r.URL.Query().Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	analytics, err := h.service.GetFleetAnalyticsByPeriod(ctx, orgID, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// UpdateFleetAnalytics handles updating existing fleet analytics
func (h *AnalyticsHandler) UpdateFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate updates
	if err := validateFleetAnalyticsUpdates(updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateFleetAnalytics(ctx, id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteFleetAnalytics handles deleting fleet analytics
func (h *AnalyticsHandler) DeleteFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteFleetAnalytics(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListFleetAnalytics handles listing fleet analytics with filtering
func (h *AnalyticsHandler) ListFleetAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	orgIDStr := r.URL.Query().Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	var startDate, endDate *time.Time

	if startDateStr := r.URL.Query().Get("startDate"); startDateStr != "" {
		date, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		startDate = &date
	}

	if endDateStr := r.URL.Query().Get("endDate"); endDateStr != "" {
		date, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		endDate = &date
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.ListFleetAnalytics(ctx, orgID, startDate, endDate, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateDriverBehaviorAnalytics handles creating new driver behavior analytics
func (h *AnalyticsHandler) CreateDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var analytics models.DriverBehaviorAnalytics
	if err := json.NewDecoder(r.Body).Decode(&analytics); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validateDriverBehaviorAnalytics(&analytics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set ID and timestamps
	analytics.ID = uuid.New().String()
	analytics.CreatedAt = time.Now()
	analytics.UpdatedAt = time.Now()
	analytics.GeneratedAt = time.Now()

	if err := h.service.CreateDriverBehaviorAnalytics(ctx, &analytics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(analytics)
}

// GetDriverBehaviorAnalytics handles retrieving driver behavior analytics by ID
func (h *AnalyticsHandler) GetDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	analytics, err := h.service.GetDriverBehaviorAnalytics(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// GetDriverBehaviorAnalyticsByDriver handles retrieving driver behavior analytics by driver ID
func (h *AnalyticsHandler) GetDriverBehaviorAnalyticsByDriver(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	driverIDStr := r.URL.Query().Get("driverId")
	driverID, err := uuid.Parse(driverIDStr)
	if err != nil {
		http.Error(w, "Invalid driver ID", http.StatusBadRequest)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.GetDriverBehaviorAnalyticsByDriver(ctx, driverID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDriverBehaviorAnalyticsByOrg handles retrieving driver behavior analytics by organization
func (h *AnalyticsHandler) GetDriverBehaviorAnalyticsByOrg(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	orgIDStr := r.URL.Query().Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.GetDriverBehaviorAnalyticsByOrg(ctx, orgID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateDriverBehaviorAnalytics handles updating existing driver behavior analytics
func (h *AnalyticsHandler) UpdateDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate updates
	if err := validateDriverBehaviorAnalyticsUpdates(updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateDriverBehaviorAnalytics(ctx, id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteDriverBehaviorAnalytics handles deleting driver behavior analytics
func (h *AnalyticsHandler) DeleteDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteDriverBehaviorAnalytics(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListDriverBehaviorAnalytics handles listing driver behavior analytics with filtering
func (h *AnalyticsHandler) ListDriverBehaviorAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	orgIDStr := r.URL.Query().Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	var periodStart, periodEnd *time.Time

	if periodStartStr := r.URL.Query().Get("periodStart"); periodStartStr != "" {
		date, err := time.Parse("2006-01-02", periodStartStr)
		if err != nil {
			http.Error(w, "Invalid period start date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		periodStart = &date
	}

	if periodEndStr := r.URL.Query().Get("periodEnd"); periodEndStr != "" {
		date, err := time.Parse("2006-01-02", periodEndStr)
		if err != nil {
			http.Error(w, "Invalid period end date format, expected YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		periodEnd = &date
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.ListDriverBehaviorAnalytics(ctx, orgID, periodStart, periodEnd, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreatePredictiveMaintenanceAnalytics handles creating new predictive maintenance analytics
func (h *AnalyticsHandler) CreatePredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var analytics models.PredictiveMaintenanceAnalytics
	if err := json.NewDecoder(r.Body).Decode(&analytics); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validatePredictiveMaintenanceAnalytics(&analytics); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set ID and timestamps
	analytics.ID = uuid.New().String()
	analytics.CreatedAt = time.Now()
	analytics.UpdatedAt = time.Now()
	analytics.GeneratedAt = time.Now()

	if err := h.service.CreatePredictiveMaintenanceAnalytics(ctx, &analytics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(analytics)
}

// GetPredictiveMaintenanceAnalytics handles retrieving predictive maintenance analytics by ID
func (h *AnalyticsHandler) GetPredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	analytics, err := h.service.GetPredictiveMaintenanceAnalytics(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// GetPredictiveMaintenanceAnalyticsByVehicle handles retrieving predictive maintenance analytics by vehicle ID
func (h *AnalyticsHandler) GetPredictiveMaintenanceAnalyticsByVehicle(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vehicleIDStr := r.URL.Query().Get("vehicleId")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.GetPredictiveMaintenanceAnalyticsByVehicle(ctx, vehicleID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdatePredictiveMaintenanceAnalytics handles updating existing predictive maintenance analytics
func (h *AnalyticsHandler) UpdatePredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate updates
	if err := validatePredictiveMaintenanceAnalyticsUpdates(updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdatePredictiveMaintenanceAnalytics(ctx, id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeletePredictiveMaintenanceAnalytics handles deleting predictive maintenance analytics
func (h *AnalyticsHandler) DeletePredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeletePredictiveMaintenanceAnalytics(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListPredictiveMaintenanceAnalytics handles listing predictive maintenance analytics with filtering
func (h *AnalyticsHandler) ListPredictiveMaintenanceAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	orgIDStr := r.URL.Query().Get("orgId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	var vehicleID *uuid.UUID

	if vehicleIDStr := r.URL.Query().Get("vehicleId"); vehicleIDStr != "" {
		parsedUUID, err := uuid.Parse(vehicleIDStr)
		if err != nil {
			http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
			return
		}
		vehicleID = &parsedUUID
	}

	limit, offset := parsePaginationParams(r.URL.Query())

	analytics, totalCount, err := h.service.ListPredictiveMaintenanceAnalytics(ctx, orgID, vehicleID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":        analytics,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheck handles health check requests
func (h *AnalyticsHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := h.service.HealthCheck(ctx); err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
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
func (h *AnalyticsHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(statusCode)
json.NewEncoder(w).Encode(data)
}
