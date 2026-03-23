package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/fuel/repository"
	"go.uber.org/zap"
)

type FuelAnomalyWorker struct {
	repo   repository.FuelLogRepository
	logger *zap.Logger
}

func NewFuelAnomalyWorker(repo repository.FuelLogRepository, logger *zap.Logger) *FuelAnomalyWorker {
	return &FuelAnomalyWorker{
		repo:   repo,
		logger: logger,
	}
}

func (w *FuelAnomalyWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runDetection(ctx)
		}
	}
}

func (w *FuelAnomalyWorker) runDetection(ctx context.Context) {
	w.logger.Info("Running fuel anomaly detection")

	// Check logs from the last 24 hours
	since := time.Now().Add(-24 * time.Hour)
	logs, err := w.repo.GetLogsForAnomalyDetection(ctx, since)
	if err != nil {
		w.logger.Error("failed to get logs for anomaly detection", zap.Error(err))
		return
	}

	// Group logs by org+vehicle to preserve tenant isolation.
	vehicleLogs := make(map[string][]models.FuelLog)
	for _, log := range logs {
		key := log.OrganizationID.String() + ":" + log.VehicleID.String()
		vehicleLogs[key] = append(vehicleLogs[key], log)
	}

	for vehicleID, history := range vehicleLogs {
		if len(history) < 2 {
			continue
		}

		for i := 1; i < len(history); i++ {
			prev := history[i-1]
			curr := history[i]

			// Pattern 1: Frequent fill-ups (e.g., < 4 hours apart)
			timeDiff := curr.TransactionDate.Sub(prev.TransactionDate)
			if timeDiff > 0 && timeDiff < 4*time.Hour {
				w.logger.Warn("Frequent fill-up detected",
					zap.String("vehicle_id", vehicleID),
					zap.String("log_id", curr.ID.String()),
					zap.Duration("time_diff", timeDiff))
				_ = w.repo.MarkAsAnomalous(ctx, curr.ID, curr.OrganizationID, "Frequent fill-up (< 4h)")
			}

			// Pattern 2: Abnormal efficiency (km per liter)
			dist := curr.Odometer - prev.Odometer
			if dist > 0 {
				efficiency := dist / curr.Quantity
				// For heavy trucks, < 2 km/l is suspicious
				if efficiency < 2.0 {
					w.logger.Warn("Low fuel efficiency detected (suspicion of theft)",
						zap.String("vehicle_id", vehicleID),
						zap.Float64("efficiency", efficiency))
					_ = w.repo.MarkAsAnomalous(ctx, curr.ID, curr.OrganizationID, fmt.Sprintf("Suspiciously low efficiency: %.2f km/l", efficiency))
				}
			} else if dist < 0 {
				// Odometer tampered or error
				_ = w.repo.MarkAsAnomalous(ctx, curr.ID, curr.OrganizationID, "Odometer reversal detected")
			}
		}
	}
}
