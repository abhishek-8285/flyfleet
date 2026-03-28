package repository

import (
	"context"
	"time"

	"errors"
	"github.com/flyfleet/fleet-core-service/internal/domains/analytics/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrDuplicateJob  = errors.New("analytics job with this idempotency key already exists for this organization")
	ErrInvalidTenant = errors.New("invalid or missing tenant context for analytics operation")
)

// AnalyticsRepository defines the interface for analytics data access
type AnalyticsRepository interface {
	// Predictive Model operations
	CreatePredictiveModel(ctx context.Context, model *models.PredictiveModel) error
	GetPredictiveModel(ctx context.Context, id string, orgID uuid.UUID) (*models.PredictiveModel, error)
	GetPredictiveModelByVersion(ctx context.Context, name, version string, orgID uuid.UUID) (*models.PredictiveModel, error)
	UpdatePredictiveModel(ctx context.Context, id string, orgID uuid.UUID, updates map[string]interface{}) error
	DeletePredictiveModel(ctx context.Context, id string, orgID uuid.UUID) error
	ListPredictiveModels(ctx context.Context, orgID uuid.UUID, modelType *string, status *string, cursor string, limit, offset int) ([]*models.PredictiveModel, int, error)

	// Prediction Result operations
	CreatePredictionResult(ctx context.Context, result *models.PredictionResult) error
	CreatePredictionResults(ctx context.Context, results []*models.PredictionResult) error
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

	RefreshVehicleDailyStats(ctx context.Context) error
	RefreshDriverDailyStats(ctx context.Context) error
	RefreshFleetHourlyStats(ctx context.Context) error

	// Audit Logging
	CreateQueryLog(ctx context.Context, log *models.AnalyticsQueryLog) error

	// Phase 8: Production-Grade Operations & ML Lifecycle
	SetTenantContext(ctx context.Context, tx pgx.Tx, orgID uuid.UUID) error

	SaveModelDeployment(ctx context.Context, deployment *models.ModelDeployment) error
	GetActiveModelDeployment(ctx context.Context, modelID string, orgID uuid.UUID) (*models.ModelDeployment, error)

	SaveDriverRiskEvent(ctx context.Context, event *models.DriverRiskEvent) error
	GetDriverRiskEvents(ctx context.Context, driverID uuid.UUID, orgID uuid.UUID, startDate, endDate time.Time) ([]*models.DriverRiskEvent, error)

	GetAnalyticsQuota(ctx context.Context, orgID uuid.UUID) (*models.AnalyticsQuota, error)
	UpdateUsedCredits(ctx context.Context, orgID uuid.UUID, credits int64) error

	// Versioned Feature Store
	CreateVersionedFeatureVector(ctx context.Context, vector []*models.MLFeatureStore) error
	GetLatestFeatures(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID) ([]*models.MLFeatureStore, error)
	GetFeaturesByVersion(ctx context.Context, orgID uuid.UUID, entityType string, entityID uuid.UUID, version string) ([]*models.MLFeatureStore, error)

	// Phase 10: Final System Maturity
	SaveQueryPlan(ctx context.Context, plan *models.AnalyticsQueryPlan) error

	// Phase 11: Production-Grade Hardening
	CreateAccessLog(ctx context.Context, log *models.AnalyticsAccessLog) error
	GetMaterializedReports(ctx context.Context, orgID uuid.UUID, reportType string, period string) ([]*models.MaterializedFleetReport, error)
	GetCapacityStats(ctx context.Context) (map[string]int64, error)
	PurgeAgedData(ctx context.Context, table string, retentionPeriod time.Duration) (int64, error)

	HealthCheck(ctx context.Context) error
	Ping(ctx context.Context) error
}
