package repository

import (
	"context"
	"time"

	"github.com/flyfleet/analytics-service/internal/models"
	"github.com/google/uuid"
)

// AnalyticsRepository defines the interface for analytics data access
type AnalyticsRepository interface {
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
	Ping(ctx context.Context) error
}