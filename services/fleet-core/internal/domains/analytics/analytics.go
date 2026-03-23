package analytics

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/google/uuid"
)

// AnalyticsService defines the service interface for analytics operations
type AnalyticsService interface {
	// Predictive Model operations
	CreatePredictiveModel(ctx context.Context, model *models.PredictiveModel) error
	GetPredictiveModel(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictiveModel, error)
	GetPredictiveModelByVersion(ctx context.Context, name, version string, orgID uuid.UUID) (*models.PredictiveModel, error)
	UpdatePredictiveModel(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error
	DeletePredictiveModel(ctx context.Context, id string, orgID uuid.UUID) error
	ListPredictiveModels(ctx context.Context, orgID uuid.UUID, modelType *string, status *string, cursor string, limit, offset int) ([]*models.PredictiveModel, int, error)
	
	// Prediction Result operations
	CreatePredictionResult(ctx context.Context, result *models.PredictionResult) error
	GetPredictionResult(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictionResult, error)
	GetPredictionResultsByModel(ctx context.Context, modelID string, orgID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error)
	GetPredictionResultsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.PredictionResult, int, error)
	UpdatePredictionResult(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error
	DeletePredictionResult(ctx context.Context, id string, orgID uuid.UUID) error
	ListPredictionResults(ctx context.Context, orgID uuid.UUID, modelID *string, anomalyDetected *bool, cursor string, limit, offset int) ([]*models.PredictionResult, int, error)
	
	// Analytics Job operations
	CreateAnalyticsJob(ctx context.Context, job *models.AnalyticsJob) error
	GetAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID) (*models.AnalyticsJob, error)
	UpdateAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error
	DeleteAnalyticsJob(ctx context.Context, id string, orgID uuid.UUID) error
	ListAnalyticsJobs(ctx context.Context, jobType *string, status *string, orgID uuid.UUID, cursor string, limit, offset int) ([]*models.AnalyticsJob, int, error)
	
	// Fleet Analytics operations
	CreateFleetAnalytics(ctx context.Context, analytics *models.FleetAnalytics) error
	GetFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.FleetAnalytics, error)
	GetFleetAnalyticsByPeriod(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) (*models.FleetAnalytics, error)
	UpdateFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error
	DeleteFleetAnalytics(ctx context.Context, id string, orgID uuid.UUID) error
	ListFleetAnalytics(ctx context.Context, orgID uuid.UUID, startDate, endDate *time.Time, cursor string, limit, offset int) ([]*models.FleetAnalytics, int, error)
	
	// Driver Behavior Analytics operations
	CreateDriverBehaviorAnalytics(ctx context.Context, analytics *models.DriverBehaviorAnalytics) error
	GetDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.DriverBehaviorAnalytics, error)
	GetDriverBehaviorAnalyticsByDriver(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error)
	GetDriverBehaviorAnalyticsByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error)
	UpdateDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error
	DeleteDriverBehaviorAnalytics(ctx context.Context, id string, orgID uuid.UUID) error
	ListDriverBehaviorAnalytics(ctx context.Context, orgID uuid.UUID, periodStart, periodEnd *time.Time, cursor string, limit, offset int) ([]*models.DriverBehaviorAnalytics, int, error)
	
	// Predictive Maintenance Analytics operations
	CreatePredictiveMaintenanceAnalytics(ctx context.Context, analytics *models.PredictiveMaintenanceAnalytics) error
	GetPredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictiveMaintenanceAnalytics, error)
	GetPredictiveMaintenanceAnalyticsByVehicle(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error)
	UpdatePredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error
	DeletePredictiveMaintenanceAnalytics(ctx context.Context, id string, orgID uuid.UUID) error
	ListPredictiveMaintenanceAnalytics(ctx context.Context, orgID uuid.UUID, vehicleID *uuid.UUID, cursor string, limit, offset int) ([]*models.PredictiveMaintenanceAnalytics, int, error)
	
	// Pre-aggregated stats operations
	GetVehicleDailyStats(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.VehicleDailyStats, string, error)
	GetDriverDailyStats(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.DriverDailyStats, string, error)
	GetFleetHourlyStats(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time, cursor string, limit int) ([]*models.FleetHourlyStats, string, error)

	// Cache Invalidation operations
	InvalidateVehicleCache(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID) error
	InvalidateDriverCache(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID) error
	InvalidateFleetCache(ctx context.Context, orgID uuid.UUID) error
	InvalidateByPattern(ctx context.Context, pattern string) error
	IncrementFleetVersion(ctx context.Context, orgID uuid.UUID) error
	SetRaw(ctx context.Context, key string, data []byte, ttl time.Duration) error
	
	HealthCheck(ctx context.Context) error

	// Phase 8: Production-Grade Operations & ML Lifecycle
	DeployModel(ctx context.Context, deployment *models.ModelDeployment) error
	RunInference(ctx context.Context, modelID string, entityID uuid.UUID, orgID uuid.UUID, input map[string]interface{}) (*models.PredictionResult, error)
	
	CreateDriverRiskEvent(ctx context.Context, event *models.DriverRiskEvent) error
	GetDriverRiskEvents(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time) ([]*models.DriverRiskEvent, error)
	
	GetAnalyticsQuota(ctx context.Context, orgID uuid.UUID) (*models.AnalyticsQuota, error)
	CheckQuota(ctx context.Context, orgID uuid.UUID, credits int64) (bool, error)
	
	RunReconciliation(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) error
	RunHistoricalBackfill(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) error
	
	CreateVersionedFeatureVector(ctx context.Context, vector []*models.MLFeatureStore) error
	GetLatestFeatures(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) ([]*models.MLFeatureStore, error)
	GetFeaturesByVersion(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID, version string) ([]*models.MLFeatureStore, error)
	GetDynamicRateLimit(ctx context.Context, orgID uuid.UUID) (int, error)
	CheckModelHealth(ctx context.Context, modelID string, orgID uuid.UUID) error
	LogAccess(ctx context.Context, log *models.AnalyticsAccessLog)
	GetExecutiveDashboard(ctx context.Context, orgID uuid.UUID, period string) ([]*models.MaterializedFleetReport, error)
	RunRetentionCycle(ctx context.Context) error
}
