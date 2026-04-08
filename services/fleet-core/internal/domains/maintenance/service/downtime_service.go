package service

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core/internal/domains/maintenance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DowntimeService struct {
	repo   *repository.MaintenanceRepositoryImpl
	logger *zap.Logger
}

func NewDowntimeService(repo *repository.MaintenanceRepositoryImpl, logger *zap.Logger) *DowntimeService {
	return &DowntimeService{repo: repo, logger: logger}
}

// CalculateVehicleMetrics calculates MTTR and MTBF for a vehicle
func (s *DowntimeService) CalculateVehicleMetrics(ctx context.Context, vehicleID uuid.UUID) (*models.DowntimeMetrics, error) {
	downtimes, err := s.repo.GetDowntimeByVehicle(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	metrics := &models.DowntimeMetrics{VehicleID: vehicleID}
	if len(downtimes) == 0 {
		return metrics, nil
	}

	var totalDowntime int
	var repairCount int

	for _, d := range downtimes {
		if d.DowntimeEnd != nil {
			totalDowntime += d.DowntimeMinutes
			repairCount++
		}
	}

	if repairCount > 0 {
		metrics.MTTR = float64(totalDowntime) / float64(repairCount)
	}
	metrics.TotalDowntime = totalDowntime
	metrics.MaintenanceCount = len(downtimes)

	// Calculate MTBF (Mean Time Between Failures)
	// Simplified: (Last repair date - First repair date) / Number of failures
	if len(downtimes) > 1 {
		first := downtimes[len(downtimes)-1].DowntimeStart
		last := downtimes[0].DowntimeStart
		days := last.Sub(first).Hours() / 24
		metrics.MTBF = days / float64(len(downtimes)-1)
	}

	return metrics, nil
}

// RecordDowntime records a new downtime event for a vehicle
func (s *DowntimeService) RecordDowntime(ctx context.Context, downtime *models.VehicleDowntime) error {
	query := `
		INSERT INTO vehicle_downtime (id, organization_id, vehicle_id, downtime_start, downtime_end, downtime_minutes, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`

	if downtime.ID == uuid.Nil {
		downtime.ID = uuid.New()
	}

	if downtime.DowntimeEnd != nil {
		downtime.DowntimeMinutes = int(downtime.DowntimeEnd.Sub(downtime.DowntimeStart).Minutes())
	}

	_, err := s.repo.GetDB().Exec(ctx, query,
		downtime.ID, downtime.OrgID, downtime.VehicleID,
		downtime.DowntimeStart, downtime.DowntimeEnd, downtime.DowntimeMinutes,
		downtime.Reason)
	return err
}
