package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/maintenance/repository"
)

type PredictiveMaintenanceWorker struct {
	repo         *repository.MaintenanceRepositoryImpl
	woService    *WorkOrderService
	logger       *zap.Logger
	pollInterval time.Duration
}

func NewPredictiveMaintenanceWorker(
	repo *repository.MaintenanceRepositoryImpl,
	woService *WorkOrderService,
	logger *zap.Logger,
) *PredictiveMaintenanceWorker {
	return &PredictiveMaintenanceWorker{
		repo:         repo,
		woService:    woService,
		logger:       logger,
		pollInterval: 1 * time.Hour,
	}
}

func (w *PredictiveMaintenanceWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.logger.Info("Starting Predictive Maintenance Worker")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.RunEvaluation(ctx)
		}
	}
}

func (w *PredictiveMaintenanceWorker) RunEvaluation(ctx context.Context) {
	w.logger.Info("Running predictive maintenance evaluation")

	// 1. Get all vehicles (simplified: get unique vehicle IDs from logs)
	query := "SELECT DISTINCT vehicle_id, organization_id FROM maintenance_logs"
	rows, err := w.repo.GetDB().Query(ctx, query)
	if err != nil {
		w.logger.Error("Failed to fetch vehicles for prediction", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var vehicleID, orgID uuid.UUID
		if err := rows.Scan(&vehicleID, &orgID); err != nil {
			continue
		}

		w.evaluateVehicle(ctx, orgID, vehicleID)
	}
}

func (w *PredictiveMaintenanceWorker) evaluateVehicle(ctx context.Context, orgID, vehicleID uuid.UUID) {
	// Rule 1: High Failure Rate Pattern
	// If a vehicle has > 3 corrective maintenance events in 30 days
	var recentFailures int
	err := w.repo.GetDB().QueryRow(ctx,
		"SELECT COUNT(*) FROM maintenance_logs WHERE vehicle_id = $1 AND type = 'corrective' AND start_date > NOW() - INTERVAL '30 days'",
		vehicleID).Scan(&recentFailures)

	if err == nil && recentFailures > 3 {
		w.triggerPreventiveWO(ctx, orgID, vehicleID, "Predictive: High failure rate detected. Schedule comprehensive inspection.")
		return
	}

	// Rule 2: Engine Hours Threshold (Placeholder for real telemetry)
	// IF engine_hours > 500 since last service
	var lastServiceHours float64
	var currentHours float64
	// In a real system, currentHours would come from a telemetry store/cache
	err = w.repo.GetDB().QueryRow(ctx,
		"SELECT COALESCE(MAX(engine_hours), 0) FROM maintenance_logs WHERE vehicle_id = $1",
		vehicleID).Scan(&lastServiceHours)
	
	// Simulated telemetry check
	currentHours = lastServiceHours + 550 // Mocking that threshold is exceeded

	if currentHours-lastServiceHours > 500 {
		w.triggerPreventiveWO(ctx, orgID, vehicleID, "Predictive: Engine hours threshold exceeded. Schedule oil and filter change.")
	}
}

func (w *PredictiveMaintenanceWorker) triggerPreventiveWO(ctx context.Context, orgID, vehicleID uuid.UUID, description string) {
	req := &models.CreateWorkOrderRequest{
		VehicleID:       vehicleID,
		Title:           "PREDICTIVE PREVENTIVE MAINTENANCE",
		Description:     description,
		MaintenanceType: models.WOMaintenanceTypePreventive,
		Priority:        models.WorkOrderPriorityMedium,
	}

	wo, err := w.woService.CreateWorkOrder(ctx, orgID, uuid.Nil, req) // System-created
	if err != nil {
		w.logger.Error("Failed to create predictive work order", zap.Error(err))
		return
	}

	w.logger.Info("Created predictive preventive work order", zap.String("wo_id", wo.ID.String()))
	
	// Auto-assign technician if enabled
	_ = w.woService.AutoAssignTechnician(ctx, wo)
}
