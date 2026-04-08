package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/fuel/repository"
	"go.uber.org/zap"
)

type FuelEfficiencyWorker struct {
	repo   repository.FuelLogRepository
	logger *zap.Logger
}

func NewFuelEfficiencyWorker(repo repository.FuelLogRepository, logger *zap.Logger) *FuelEfficiencyWorker {
	return &FuelEfficiencyWorker{
		repo:   repo,
		logger: logger,
	}
}

func (w *FuelEfficiencyWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.calculateEfficiency(ctx)
		}
	}
}

func (w *FuelEfficiencyWorker) calculateEfficiency(ctx context.Context) {
	w.logger.Info("Calculating daily fuel efficiency metrics")
	// Query recent fuel logs and calculate MPG/Kmpl
	// Update fuel_efficiency_history table (Phase 5)
}
