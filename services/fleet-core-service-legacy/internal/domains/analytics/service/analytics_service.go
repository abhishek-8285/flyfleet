package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/repository"
	"github.com/google/uuid"
)

// AnalyticsService defines the service interface for analytics operations
type AnalyticsService interface {
	// Predictive Model operations
	CreatePredictiveModel(ctx context.Context, model *models.PredictiveModel) error
	GetPredictiveModel(ctx context.Context, id string) (*models.PredictiveModel, error)
	GetPredictiveModelByVersion(ctx context.Context, name, version string) (*models.PredictiveModel, error)
	UpdatePredictiveModel(ctx context.Context, id string, updates map[string]interface{}) error
	DeletePredictiveModel(ctx context.Context, id string) error
	ListPredictiveModels(ctx context.Context, modelType *string, status *string, limit, offset int) ([]*models.PredictiveModel, int, error)
	
	// Prediction Result operations
	CreatePredictionResult(ctx context.Context, result *models.PredictionResult) error
	GetPredictionResult(ctx context.Context, id string) (*models.PredictionResult, error)
	GetPredictionResultsByModel(ctx context.Context, modelID string, limit, offset int) ([]*models.PredictionResult, int, error)
	GetPredictionResultsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error)
	UpdatePredictionResult(ctx context.Context, id string, updates map[string]interface{}) error
	DeletePredictionResult(ctx context.Context, id string) error
	ListPredictionResults(ctx context.Context, orgID uuid.UUID, modelID *string, anomalyDetected *bool, limit, offset int) ([]*models.PredictionResult, int, error)
	
	// Analytics Job operations
	CreateAnalyticsJob(ctx context.Context, job *models.AnalyticsJob) error
	GetAnalyticsJob(ctx context.Context, id string) (*models.AnalyticsJob, error)
	UpdateAnalyticsJob(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteAnalyticsJob(ctx context.Context, id string) error
	ListAnalyticsJobs(ctx context.Context, jobType *string, status *string, orgID uuid.UUID, limit, offset int) ([]*models.AnalyticsJob, int, error)
	
	// Fleet Analytics operations
	CreateFleetAnalytics(ctx context.Context, analytics *models.FleetAnalytics) error
	GetFleetAnalytics(ctx context.Context, id string) (*models.FleetAnalytics, error)
	GetFleetAnalyticsByPeriod(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*models.FleetAnalytics, error)
	UpdateFleetAnalytics(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteFleetAnalytics(ctx context.Context, id string) error
	ListFleetAnalytics(ctx context.Context, orgID uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]*models.FleetAnalytics, int, error)
	
	// Driver Behavior Analytics operations
	CreateDriverBehaviorAnalytics(ctx context.Context, analytics *models.DriverBehaviorAnalytics) error
	GetDriverBehaviorAnalytics(ctx context.Context, id string) (*models.DriverBehaviorAnalytics, error)
	GetDriverBehaviorAnalyticsByDriver(ctx context.Context, driverID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error)
	GetDriverBehaviorAnalyticsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error)
	UpdateDriverBehaviorAnalytics(ctx context.Context, id string, updates map[string]interface{}) error
	DeleteDriverBehaviorAnalytics(ctx context.Context, id string) error
	ListDriverBehaviorAnalytics(ctx context.Context, orgID uuid.UUID, periodStart, periodEnd *time.Time, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error)
	
	// Predictive Maintenance Analytics operations
	CreatePredictiveMaintenanceAnalytics(ctx context.Context, analytics *models.PredictiveMaintenanceAnalytics) error
	GetPredictiveMaintenanceAnalytics(ctx context.Context, id string) (*models.PredictiveMaintenanceAnalytics, error)
	GetPredictiveMaintenanceAnalyticsByVehicle(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error)
	UpdatePredictiveMaintenanceAnalytics(ctx context.Context, id string, updates map[string]interface{}) error
	DeletePredictiveMaintenanceAnalytics(ctx context.Context, id string) error
	ListPredictiveMaintenanceAnalytics(ctx context.Context, orgID uuid.UUID, vehicleID *uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error)
	
	HealthCheck(ctx context.Context) error
}

// AnalyticsServiceImpl implements the AnalyticsService interface
type AnalyticsServiceImpl struct {
	repo repository.AnalyticsRepository
}

// NewAnalyticsService creates a new instance of AnalyticsService
func NewAnalyticsService(repo repository.AnalyticsRepository) *AnalyticsServiceImpl {
	return &AnalyticsServiceImpl{
		repo: repo,
	}
}

// Validation functions
func validatePredictiveModel(model *models.PredictiveModel) error {
	if model.Name == "" {
		return errors.New("model name is required")
	}
	if model.Version == "" {
		return errors.New("model version is required")
	}
	if model.ModelType == "" {
		return errors.New("model type is required")
	}
	if model.Status == "" {
		return errors.New("model status is required")
	}
	if model.TrainedAt.IsZero() {
		return errors.New("trained at timestamp is required")
	}
	return nil
}

func validatePredictionResult(result *models.PredictionResult) error {
	if result.ModelID == "" {
		return errors.New("model ID is required")
	}
	if result.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if result.ConfidenceScore < 0 || result.ConfidenceScore > 1 {
		return errors.New("confidence score must be between 0 and 1")
	}
	return nil
}

func validateAnalyticsJob(job *models.AnalyticsJob) error {
	if job.Name == "" {
		return errors.New("job name is required")
	}
	if job.Type == "" {
		return errors.New("job type is required")
	}
	if job.Status == "" {
		return errors.New("job status is required")
	}
	if job.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if job.CreatedBy == uuid.Nil {
		return errors.New("created by ID is required")
	}
	return nil
}

func validateFleetAnalytics(analytics *models.FleetAnalytics) error {
	if analytics.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if analytics.StartDate.IsZero() || analytics.EndDate.IsZero() {
		return errors.New("start and end dates are required")
	}
	if analytics.StartDate.After(analytics.EndDate) {
		return errors.New("start date cannot be after end date")
	}
	return nil
}

func validateDriverBehaviorAnalytics(analytics *models.DriverBehaviorAnalytics) error {
	if analytics.DriverID == uuid.Nil {
		return errors.New("driver ID is required")
	}
	if analytics.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if analytics.AnalysisPeriodStart.IsZero() || analytics.AnalysisPeriodEnd.IsZero() {
		return errors.New("analysis period start and end dates are required")
	}
	if analytics.AnalysisPeriodStart.After(analytics.AnalysisPeriodEnd) {
		return errors.New("analysis period start cannot be after end date")
	}
	return nil
}

func validatePredictiveMaintenanceAnalytics(analytics *models.PredictiveMaintenanceAnalytics) error {
	if analytics.VehicleID == uuid.Nil {
		return errors.New("vehicle ID is required")
	}
	if analytics.OrganizationID == uuid.Nil {
		return errors.New("organization ID is required")
	}
	if analytics.AnalysisDate.IsZero() {
		return errors.New("analysis date is required")
	}
	return nil
}

// CreatePredictiveModel creates a new predictive model with validation
func (s *AnalyticsServiceImpl) CreatePredictiveModel(ctx context.Context, model *models.PredictiveModel) error {
	if err := validatePredictiveModel(model); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	if model.CreatedAt.IsZero() {
		model.CreatedAt = time.Now()
	}
	if model.UpdatedAt.IsZero() {
		model.UpdatedAt = time.Now()
	}

	return s.repo.CreatePredictiveModel(ctx, model)
}

// GetPredictiveModel retrieves a predictive model by ID
func (s *AnalyticsServiceImpl) GetPredictiveModel(ctx context.Context, id string) (*models.PredictiveModel, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetPredictiveModel(ctx, id)
}

// GetPredictiveModelByVersion retrieves a predictive model by name and version
func (s *AnalyticsServiceImpl) GetPredictiveModelByVersion(ctx context.Context, name, version string) (*models.PredictiveModel, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if version == "" {
		return nil, errors.New("version is required")
	}

	return s.repo.GetPredictiveModelByVersion(ctx, name, version)
}

// UpdatePredictiveModel updates an existing predictive model with validation
func (s *AnalyticsServiceImpl) UpdatePredictiveModel(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdatePredictiveModel(ctx, id, updates)
}

// DeletePredictiveModel deletes a predictive model
func (s *AnalyticsServiceImpl) DeletePredictiveModel(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeletePredictiveModel(ctx, id)
}

// ListPredictiveModels retrieves predictive models with filtering
func (s *AnalyticsServiceImpl) ListPredictiveModels(ctx context.Context, modelType *string, status *string, limit, offset int) ([]*models.PredictiveModel, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListPredictiveModels(ctx, modelType, status, limit, offset)
}

// CreatePredictionResult creates a new prediction result with validation
func (s *AnalyticsServiceImpl) CreatePredictionResult(ctx context.Context, result *models.PredictionResult) error {
	if err := validatePredictionResult(result); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if result.ID == "" {
		result.ID = uuid.New().String()
	}
	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}
	if result.UpdatedAt.IsZero() {
		result.UpdatedAt = time.Now()
	}
	if result.PredictedAt.IsZero() {
		result.PredictedAt = time.Now()
	}

	return s.repo.CreatePredictionResult(ctx, result)
}

// GetPredictionResult retrieves a prediction result by ID
func (s *AnalyticsServiceImpl) GetPredictionResult(ctx context.Context, id string) (*models.PredictionResult, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetPredictionResult(ctx, id)
}

// GetPredictionResultsByModel retrieves prediction results by model ID
func (s *AnalyticsServiceImpl) GetPredictionResultsByModel(ctx context.Context, modelID string, limit, offset int) ([]*models.PredictionResult, int, error) {
	if modelID == "" {
		return nil, 0, errors.New("model ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetPredictionResultsByModel(ctx, modelID, limit, offset)
}

// GetPredictionResultsByOrg retrieves prediction results by organization
func (s *AnalyticsServiceImpl) GetPredictionResultsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, errors.New("organization ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetPredictionResultsByOrg(ctx, orgID, limit, offset)
}

// UpdatePredictionResult updates an existing prediction result with validation
func (s *AnalyticsServiceImpl) UpdatePredictionResult(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdatePredictionResult(ctx, id, updates)
}

// DeletePredictionResult deletes a prediction result
func (s *AnalyticsServiceImpl) DeletePredictionResult(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeletePredictionResult(ctx, id)
}

// ListPredictionResults retrieves prediction results with filtering
func (s *AnalyticsServiceImpl) ListPredictionResults(ctx context.Context, orgID uuid.UUID, modelID *string, anomalyDetected *bool, limit, offset int) ([]*models.PredictionResult, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, errors.New("organization ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListPredictionResults(ctx, orgID, modelID, anomalyDetected, limit, offset)
}

func (s *AnalyticsServiceImpl) CreateAnalyticsJob(ctx context.Context, job *models.AnalyticsJob) error {
	if err := validateAnalyticsJob(job); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	if job.Status == "" {
		job.Status = "pending"
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = time.Now()
	}

	if err := s.repo.CreateAnalyticsJob(ctx, job); err != nil {
		return err
	}

	// In a real implementation, this would publish the job to a queue for processing
	// For now, we'll simulate the processing by updating the job status
	go s.processAnalyticsJob(job.ID)

	return nil
}

// processAnalyticsJob simulates the processing of an analytics job
func (s *AnalyticsServiceImpl) processAnalyticsJob(jobID string) {
	ctx := context.Background()

	// Update job status to running
	err := s.repo.UpdateAnalyticsJob(ctx, jobID, map[string]interface{}{
		"status":      "running",
		"start_time":  time.Now(),
		"updated_at":  time.Now(),
	})
	if err != nil {
		fmt.Printf("Error updating analytics job status to running: %v\n", err)
		return
	}

	// Simulate processing time
	time.Sleep(5 * time.Second)

	// Get the job to check its type
	job, err := s.repo.GetAnalyticsJob(ctx, jobID)
	if err != nil {
		fmt.Printf("Error getting analytics job: %v\n", err)
		return
	}

	// Based on job type, simulate different results
	// For now, we'll just mark it as completed and update progress
	updates := map[string]interface{}{
		"status":     "completed",
		"progress":   100.0,
		"end_time":   time.Now(),
		"updated_at": time.Now(),
	}

	err = s.repo.UpdateAnalyticsJob(ctx, jobID, updates)
	if err != nil {
		fmt.Printf("Error updating analytics job status to completed: %v\n", err)
		return
	}

	// Generate results based on the job type and parameters
	s.generateAnalyticsResults(ctx, job)
}

func (s *AnalyticsServiceImpl) generateAnalyticsResults(ctx context.Context, job *models.AnalyticsJob) {
	// Get parameters from the job
	params := job.Parameters
	if params == nil {
		params = make(map[string]interface{})
	}

	switch job.Type {
	case "driver_behavior":
		// Try to get driver_id from params, otherwise generate a placeholder warning
		var driverID uuid.UUID
		if driverIDStr, ok := params["driver_id"].(string); ok {
			var err error
			driverID, err = uuid.Parse(driverIDStr)
			if err != nil {
				fmt.Printf("Warning: Invalid driver_id in job parameters: %s\n", driverIDStr)
				return
			}
		} else {
			fmt.Printf("Warning: driver_behavior job missing driver_id parameter\n")
			return
		}

		// Get analysis period from params or use defaults
		periodStart := time.Now().Add(-24 * time.Hour)
		periodEnd := time.Now()
		if startStr, ok := params["period_start"].(string); ok {
			if t, err := time.Parse(time.RFC3339, startStr); err == nil {
				periodStart = t
			}
		}
		if endStr, ok := params["period_end"].(string); ok {
			if t, err := time.Parse(time.RFC3339, endStr); err == nil {
				periodEnd = t
			}
		}

		analytics := &models.DriverBehaviorAnalytics{
			ID:                  uuid.New().String(),
			DriverID:            driverID,
			OrganizationID:      job.OrganizationID,
			AnalysisPeriodStart: periodStart,
			AnalysisPeriodEnd:   periodEnd,
			SafetyScore:         85.5,
			EfficiencyScore:     92.0,
			ComfortScore:        100.0,
			GeneratedAt:         time.Now(),
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
		_ = s.repo.CreateDriverBehaviorAnalytics(ctx, analytics)
		fmt.Printf("Generated driver behavior analytics for job %s, driver %s\n", job.ID, driverID.String())

	case "predictive_maintenance":
		// Try to get vehicle_id from params
		var vehicleID uuid.UUID
		if vehicleIDStr, ok := params["vehicle_id"].(string); ok {
			var err error
			vehicleID, err = uuid.Parse(vehicleIDStr)
			if err != nil {
				fmt.Printf("Warning: Invalid vehicle_id in job parameters: %s\n", vehicleIDStr)
				return
			}
		} else {
			fmt.Printf("Warning: predictive_maintenance job missing vehicle_id parameter\n")
			return
		}

		analytics := &models.PredictiveMaintenanceAnalytics{
			ID:               uuid.New().String(),
			VehicleID:        vehicleID,
			OrganizationID:   job.OrganizationID,
			AnalysisDate:     time.Now(),
			ReliabilityScore: 78.2,
			FailurePrediction: &models.FailurePrediction{
				Component:            "Brake Pads",
				PredictedFailureDate: time.Now().Add(15 * 24 * time.Hour),
				ConfidenceScore:      0.85,
				UrgencyLevel:         "medium",
			},
			NextServiceDue: time.Now().Add(15 * 24 * time.Hour),
			GeneratedAt:    time.Now(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		_ = s.repo.CreatePredictiveMaintenanceAnalytics(ctx, analytics)
		fmt.Printf("Generated predictive maintenance analytics for job %s, vehicle %s\n", job.ID, vehicleID.String())
	default:
		fmt.Printf("Warning: Unknown analytics job type: %s\n", job.Type)
	}
}

// GetAnalyticsJob retrieves an analytics job by ID
func (s *AnalyticsServiceImpl) GetAnalyticsJob(ctx context.Context, id string) (*models.AnalyticsJob, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetAnalyticsJob(ctx, id)
}

// UpdateAnalyticsJob updates an existing analytics job with validation
func (s *AnalyticsServiceImpl) UpdateAnalyticsJob(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateAnalyticsJob(ctx, id, updates)
}

// DeleteAnalyticsJob deletes an analytics job
func (s *AnalyticsServiceImpl) DeleteAnalyticsJob(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteAnalyticsJob(ctx, id)
}

// ListAnalyticsJobs retrieves analytics jobs with filtering
func (s *AnalyticsServiceImpl) ListAnalyticsJobs(ctx context.Context, jobType *string, status *string, orgID uuid.UUID, limit, offset int) ([]*models.AnalyticsJob, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, errors.New("organization ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListAnalyticsJobs(ctx, jobType, status, orgID, limit, offset)
}

// CreateFleetAnalytics creates new fleet analytics with validation
func (s *AnalyticsServiceImpl) CreateFleetAnalytics(ctx context.Context, analytics *models.FleetAnalytics) error {
	if err := validateFleetAnalytics(analytics); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if analytics.ID == "" {
		analytics.ID = uuid.New().String()
	}
	if analytics.GeneratedAt.IsZero() {
		analytics.GeneratedAt = time.Now()
	}
	if analytics.CreatedAt.IsZero() {
		analytics.CreatedAt = time.Now()
	}
	if analytics.UpdatedAt.IsZero() {
		analytics.UpdatedAt = time.Now()
	}

	return s.repo.CreateFleetAnalytics(ctx, analytics)
}

// GetFleetAnalytics retrieves fleet analytics by ID
func (s *AnalyticsServiceImpl) GetFleetAnalytics(ctx context.Context, id string) (*models.FleetAnalytics, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetFleetAnalytics(ctx, id)
}

// GetFleetAnalyticsByPeriod retrieves fleet analytics by organization and date period
func (s *AnalyticsServiceImpl) GetFleetAnalyticsByPeriod(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*models.FleetAnalytics, error) {
	if orgID == uuid.Nil {
		return nil, errors.New("organization ID is required")
	}
	if startDate.IsZero() || endDate.IsZero() {
		return nil, errors.New("start and end dates are required")
	}
	if startDate.After(endDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	return s.repo.GetFleetAnalyticsByPeriod(ctx, orgID, startDate, endDate)
}

// UpdateFleetAnalytics updates existing fleet analytics
func (s *AnalyticsServiceImpl) UpdateFleetAnalytics(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateFleetAnalytics(ctx, id, updates)
}

// DeleteFleetAnalytics deletes fleet analytics
func (s *AnalyticsServiceImpl) DeleteFleetAnalytics(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteFleetAnalytics(ctx, id)
}

// ListFleetAnalytics retrieves fleet analytics with filtering
func (s *AnalyticsServiceImpl) ListFleetAnalytics(ctx context.Context, orgID uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]*models.FleetAnalytics, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, errors.New("organization ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListFleetAnalytics(ctx, orgID, startDate, endDate, limit, offset)
}

// CreateDriverBehaviorAnalytics creates new driver behavior analytics with validation
func (s *AnalyticsServiceImpl) CreateDriverBehaviorAnalytics(ctx context.Context, analytics *models.DriverBehaviorAnalytics) error {
	if err := validateDriverBehaviorAnalytics(analytics); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if analytics.ID == "" {
		analytics.ID = uuid.New().String()
	}
	if analytics.GeneratedAt.IsZero() {
		analytics.GeneratedAt = time.Now()
	}
	if analytics.CreatedAt.IsZero() {
		analytics.CreatedAt = time.Now()
	}
	if analytics.UpdatedAt.IsZero() {
		analytics.UpdatedAt = time.Now()
	}

	return s.repo.CreateDriverBehaviorAnalytics(ctx, analytics)
}

// GetDriverBehaviorAnalytics retrieves driver behavior analytics by ID
func (s *AnalyticsServiceImpl) GetDriverBehaviorAnalytics(ctx context.Context, id string) (*models.DriverBehaviorAnalytics, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetDriverBehaviorAnalytics(ctx, id)
}

// GetDriverBehaviorAnalyticsByDriver retrieves driver behavior analytics by driver ID
func (s *AnalyticsServiceImpl) GetDriverBehaviorAnalyticsByDriver(ctx context.Context, driverID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	if driverID == uuid.Nil {
		return nil, 0, errors.New("driver ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetDriverBehaviorAnalyticsByDriver(ctx, driverID, limit, offset)
}

// GetDriverBehaviorAnalyticsByOrg retrieves driver behavior analytics by organization
func (s *AnalyticsServiceImpl) GetDriverBehaviorAnalyticsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, errors.New("organization ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetDriverBehaviorAnalyticsByOrg(ctx, orgID, limit, offset)
}

// UpdateDriverBehaviorAnalytics updates existing driver behavior analytics
func (s *AnalyticsServiceImpl) UpdateDriverBehaviorAnalytics(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdateDriverBehaviorAnalytics(ctx, id, updates)
}

// DeleteDriverBehaviorAnalytics deletes driver behavior analytics
func (s *AnalyticsServiceImpl) DeleteDriverBehaviorAnalytics(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeleteDriverBehaviorAnalytics(ctx, id)
}

// ListDriverBehaviorAnalytics retrieves driver behavior analytics with filtering
func (s *AnalyticsServiceImpl) ListDriverBehaviorAnalytics(ctx context.Context, orgID uuid.UUID, periodStart, periodEnd *time.Time, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, errors.New("organization ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListDriverBehaviorAnalytics(ctx, orgID, periodStart, periodEnd, limit, offset)
}

// CreatePredictiveMaintenanceAnalytics creates new predictive maintenance analytics with validation
func (s *AnalyticsServiceImpl) CreatePredictiveMaintenanceAnalytics(ctx context.Context, analytics *models.PredictiveMaintenanceAnalytics) error {
	if err := validatePredictiveMaintenanceAnalytics(analytics); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set default values if not provided
	if analytics.ID == "" {
		analytics.ID = uuid.New().String()
	}
	if analytics.GeneratedAt.IsZero() {
		analytics.GeneratedAt = time.Now()
	}
	if analytics.CreatedAt.IsZero() {
		analytics.CreatedAt = time.Now()
	}
	if analytics.UpdatedAt.IsZero() {
		analytics.UpdatedAt = time.Now()
	}

	return s.repo.CreatePredictiveMaintenanceAnalytics(ctx, analytics)
}

// GetPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics by ID
func (s *AnalyticsServiceImpl) GetPredictiveMaintenanceAnalytics(ctx context.Context, id string) (*models.PredictiveMaintenanceAnalytics, error) {
	if id == "" {
		return nil, errors.New("ID is required")
	}

	return s.repo.GetPredictiveMaintenanceAnalytics(ctx, id)
}

// GetPredictiveMaintenanceAnalyticsByVehicle retrieves predictive maintenance analytics by vehicle ID
func (s *AnalyticsServiceImpl) GetPredictiveMaintenanceAnalyticsByVehicle(ctx context.Context, vehicleID uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	if vehicleID == uuid.Nil {
		return nil, 0, errors.New("vehicle ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetPredictiveMaintenanceAnalyticsByVehicle(ctx, vehicleID, limit, offset)
}

// UpdatePredictiveMaintenanceAnalytics updates existing predictive maintenance analytics
func (s *AnalyticsServiceImpl) UpdatePredictiveMaintenanceAnalytics(ctx context.Context, id string, updates map[string]interface{}) error {
	if id == "" {
		return errors.New("ID is required")
	}
	if len(updates) == 0 {
		return errors.New("updates are required")
	}

	return s.repo.UpdatePredictiveMaintenanceAnalytics(ctx, id, updates)
}

// DeletePredictiveMaintenanceAnalytics deletes predictive maintenance analytics
func (s *AnalyticsServiceImpl) DeletePredictiveMaintenanceAnalytics(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("ID is required")
	}

	return s.repo.DeletePredictiveMaintenanceAnalytics(ctx, id)
}

// ListPredictiveMaintenanceAnalytics retrieves predictive maintenance analytics with filtering
func (s *AnalyticsServiceImpl) ListPredictiveMaintenanceAnalytics(ctx context.Context, orgID uuid.UUID, vehicleID *uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error) {
	if orgID == uuid.Nil {
		return nil, 0, errors.New("organization ID is required")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListPredictiveMaintenanceAnalytics(ctx, orgID, vehicleID, limit, offset)
}

// HealthCheck performs a health check on the service and its dependencies
func (s *AnalyticsServiceImpl) HealthCheck(ctx context.Context) error {
	return s.repo.HealthCheck(ctx)
}
