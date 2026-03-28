package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AnalyticsService struct {
	repo   *repository.MaintenanceRepositoryImpl
	logger *zap.Logger
}

func NewAnalyticsService(repo *repository.MaintenanceRepositoryImpl, logger *zap.Logger) *AnalyticsService {
	return &AnalyticsService{repo: repo, logger: logger}
}

// GetCostPerVehicle retrieves total maintenance cost per vehicle
func (s *AnalyticsService) GetCostPerVehicle(ctx context.Context, orgID uuid.UUID) ([]models.CostPerVehicle, error) {
	query := `
		SELECT vehicle_id, SUM(cost) as total_cost
		FROM maintenance_logs
		WHERE organization_id = $1
		GROUP BY vehicle_id
		ORDER BY total_cost DESC`

	rows, err := s.repo.GetDB().Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.CostPerVehicle
	for rows.Next() {
		var c models.CostPerVehicle
		err := rows.Scan(&c.VehicleID, &c.TotalCost)
		if err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	return results, nil
}

// GetVendorRepairComparison compares repair costs across vendors
func (s *AnalyticsService) GetVendorRepairComparison(ctx context.Context, orgID uuid.UUID) ([]models.VendorCostComparison, error) {
	query := `
		SELECT v.id, v.name, SUM(l.cost) as total_cost, COUNT(l.id) as job_count, AVG(l.cost) as avg_cost
		FROM maintenance_logs l
		JOIN maintenance_vendors v ON l.performed_by = v.id
		WHERE l.organization_id = $1
		GROUP BY v.id, v.name
		ORDER BY avg_cost ASC`

	rows, err := s.repo.GetDB().Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.VendorCostComparison
	for rows.Next() {
		var v models.VendorCostComparison
		err := rows.Scan(&v.VendorID, &v.VendorName, &v.TotalCost, &v.JobCount, &v.AvgCost)
		if err != nil {
			return nil, err
		}
		results = append(results, v)
	}
	return results, nil
}

// UpdateVehicleLifecycleMetrics calculates and updates lifecycle metrics for a vehicle
func (s *AnalyticsService) UpdateVehicleLifecycleMetrics(ctx context.Context, vehicleID uuid.UUID) error {
	// 1. Get total cost and event count
	var totalCost float64
	var eventCount int
	err := s.repo.GetDB().QueryRow(ctx,
		"SELECT COALESCE(SUM(cost), 0), COUNT(*) FROM maintenance_logs WHERE vehicle_id = $1",
		vehicleID).Scan(&totalCost, &eventCount)
	if err != nil {
		return err
	}

	// 2. Get total downtime hours
	var downtimeMinutes int
	err = s.repo.GetDB().QueryRow(ctx,
		"SELECT COALESCE(SUM(downtime_minutes), 0) FROM vehicle_downtime WHERE vehicle_id = $1",
		vehicleID).Scan(&downtimeMinutes)
	if err != nil {
		return err
	}
	downtimeHours := float64(downtimeMinutes) / 60.0

	// 3. Simple replacement recommendation logic
	recommendation := "RETAIN"
	if totalCost > 50000 || downtimeHours > 200 {
		recommendation = "EVALUATE FOR REPLACEMENT"
	}
	if totalCost > 100000 || downtimeHours > 500 {
		recommendation = "REPLACE IMMEDIATELY"
	}

	query := `
		INSERT INTO vehicle_lifecycle_metrics (vehicle_id, total_maintenance_cost, maintenance_events, downtime_hours, replacement_recommendation, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (vehicle_id) DO UPDATE SET
			total_maintenance_cost = $2,
			maintenance_events = $3,
			downtime_hours = $4,
			replacement_recommendation = $5,
			updated_at = NOW()`

	_, err = s.repo.GetDB().Exec(ctx, query, vehicleID, totalCost, eventCount, downtimeHours, recommendation)
	return err
}
