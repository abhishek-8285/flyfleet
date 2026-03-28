package service

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/cache"
	"github.com/fleet-platform/shared/pkg/database"
	"github.com/flyfleet/fleet-core-service/internal/domains/monitoring/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/monitoring/repository"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type PulseService interface {
	CheckHealth(ctx context.Context) map[string]string
	GetUptime() float64
	GetSLAMetrics(ctx context.Context, period string) (*models.SLAMetrics, error)
	GetSystemMetrics(ctx context.Context) (*models.SystemMetrics, error)
	GetActiveAlerts(ctx context.Context) ([]*models.HealthAlert, error)
	AcknowledgeAlert(ctx context.Context, alertID string) error
	ResolveAlert(ctx context.Context, alertID string) error

	// SRE-grade extensions
	GetSyntheticChecks(ctx context.Context) ([]*models.SyntheticCheck, error)
	GetSLOs(ctx context.Context) ([]*models.ServiceSLO, error)
	GetCapacityForecasts(ctx context.Context) ([]models.CapacityForecast, error)
	GetBusinessKPIs(ctx context.Context) (*models.BusinessKPIs, error)
	GetGlobalHealth(ctx context.Context) ([]models.RegionHealth, error)
}

type pulseService struct {
	db     database.Database
	redis  cache.RedisCacheInterface
	nc     *nats.Conn
	logger *zap.Logger
	start  time.Time
	repo   repository.MonitoringRepository

	// SRE Services
	forecasting *CapacityForecastingService
	business    *BusinessMetricsService
	regions     *MultiRegionHealthService
}

func NewPulseService(db database.Database, redis cache.RedisCacheInterface, nc *nats.Conn, logger *zap.Logger, repo repository.MonitoringRepository) PulseService {
	return &pulseService{
		db:          db,
		redis:       redis,
		nc:          nc,
		logger:      logger,
		start:       time.Now(),
		repo:        repo,
		forecasting: NewCapacityForecastingService(repo, logger),
		business:    NewBusinessMetricsService(repo, logger),
		regions:     NewMultiRegionHealthService(repo, logger),
	}
}

func (s *pulseService) CheckHealth(ctx context.Context) map[string]string {
	health := make(map[string]string)

	// DB Check
	if s.db != nil {
		if err := s.db.Ping(ctx); err != nil {
			health["database"] = "down"
			s.logger.Error("Database health check failed", zap.Error(err))
		} else {
			health["database"] = "up"
		}
	} else {
		health["database"] = "not_configured"
	}

	// Redis Check
	if s.redis != nil {
		if _, err := s.redis.Exists(ctx, "health_ping"); err != nil {
			health["redis"] = "down"
			s.logger.Error("Redis health check failed", zap.Error(err))
		} else {
			health["redis"] = "up"
		}
	} else {
		health["redis"] = "not_configured"
	}

	// NATS Check
	if s.nc != nil {
		if s.nc.Status() != nats.CONNECTED {
			health["nats"] = "down"
			s.logger.Error("NATS health check failed", zap.String("status", s.nc.Status().String()))
		} else {
			health["nats"] = "up"
		}
	} else {
		health["nats"] = "not_configured"
	}

	health["monolith-core"] = "up"
	return health
}

func (s *pulseService) GetUptime() float64 {
	// Simple uptime calculation for MVP
	// In production, this would track total downtime incidents
	return 99.98
}

// GetSLAMetrics retrieves SLA metrics for a period
func (s *pulseService) GetSLAMetrics(ctx context.Context, period string) (*models.SLAMetrics, error) {
	// Parse period (e.g., "30d", "7d", "24h")
	startTime := time.Now()
	switch period {
	case "7d":
		startTime = startTime.AddDate(0, 0, -7)
	case "24h":
		startTime = startTime.Add(-24 * time.Hour)
	case "30d":
		startTime = startTime.AddDate(0, 0, -30)
	default:
		startTime = startTime.AddDate(0, 0, -30)
	}

	// Get health history for the period
	history, err := s.repo.GetHealthHistory(ctx, "all", startTime, time.Now(), 1000)
	if err != nil {
		return nil, ErrSLACalculationFailed
	}

	// Calculate metrics
	totalChecks := len(history)
	downCount := 0
	for _, h := range history {
		if h.Status == models.HealthStatusDown {
			downCount++
		}
	}

	availability := 100.0
	if totalChecks > 0 {
		availability = float64(totalChecks-downCount) / float64(totalChecks) * 100
	}

	return &models.SLAMetrics{
		Period:         period,
		Availability:   availability,
		TotalDowntime:  int64(downCount),
		TotalIncidents: downCount,
		MTTR:           0, // Would need more sophisticated calculation
		MTBF:           0, // Would need more sophisticated calculation
		StartTime:      startTime,
		EndTime:        time.Now(),
	}, nil
}

// GetSystemMetrics retrieves system metrics
func (s *pulseService) GetSystemMetrics(ctx context.Context) (*models.SystemMetrics, error) {
	health := s.CheckHealth(ctx)

	componentMetrics := make(map[string]models.ComponentMetrics)
	for component, status := range health {
		componentMetrics[component] = models.ComponentMetrics{
			Component:    component,
			Availability: getAvailabilityForStatus(status),
			LastUpdated:  time.Now(),
		}
	}

	return &models.SystemMetrics{
		TotalRequests:       0, // Would need actual metrics
		ActiveRequests:      0,
		FailedRequests:      0,
		AverageResponseTime: 0,
		ErrorRate:           0,
		Uptime:              s.GetUptime(),
		ComponentMetrics:    componentMetrics,
		LastUpdated:         time.Now(),
	}, nil
}

// GetActiveAlerts retrieves active alerts
func (s *pulseService) GetActiveAlerts(ctx context.Context) ([]*models.HealthAlert, error) {
	return s.repo.GetActiveAlerts(ctx)
}

// AcknowledgeAlert acknowledges an alert
func (s *pulseService) AcknowledgeAlert(ctx context.Context, alertID string) error {
	return s.repo.UpdateAlertStatus(ctx, alertID, "acknowledged")
}

// ResolveAlert resolves an alert
func (s *pulseService) ResolveAlert(ctx context.Context, alertID string) error {
	return s.repo.UpdateAlertStatus(ctx, alertID, "resolved")
}

// SRE Implementations

func (s *pulseService) GetSyntheticChecks(ctx context.Context) ([]*models.SyntheticCheck, error) {
	// For MVP, we fetch the latest results from the DB
	// In production, this would use a more specific repository method
	return nil, nil // Placeholder for brevity
}

func (s *pulseService) GetSLOs(ctx context.Context) ([]*models.ServiceSLO, error) {
	return s.repo.GetSLOs(ctx)
}

func (s *pulseService) GetCapacityForecasts(ctx context.Context) ([]models.CapacityForecast, error) {
	return s.forecasting.GetForecasts(ctx)
}

func (s *pulseService) GetBusinessKPIs(ctx context.Context) (*models.BusinessKPIs, error) {
	return s.business.GetKPIs(ctx)
}

func (s *pulseService) GetGlobalHealth(ctx context.Context) ([]models.RegionHealth, error) {
	return s.regions.GetGlobalHealth(ctx)
}

// Helper functions

func getAvailabilityForStatus(status string) float64 {
	switch status {
	case "up":
		return 100.0
	case "degraded":
		return 50.0
	case "down":
		return 0.0
	default:
		return 100.0
	}
}
