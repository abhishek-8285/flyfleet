package service

import (
	"context"
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/monitoring/models"
	"github.com/flyfleet/fleet-core/internal/domains/monitoring/repository"
	"github.com/flyfleet/fleet-core/internal/platform/observability"
	"go.uber.org/zap"
)

type SyntheticCheckWorker struct {
	repo   repository.MonitoringRepository
	logger *zap.Logger
	client *http.Client
}

func NewSyntheticCheckWorker(repo repository.MonitoringRepository, logger *zap.Logger) *SyntheticCheckWorker {
	return &SyntheticCheckWorker{
		repo:   repo,
		logger: logger,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *SyntheticCheckWorker) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runChecks(ctx)
		}
	}
}

func (w *SyntheticCheckWorker) runChecks(ctx context.Context) {
	ctx, span := observability.StartSpan(ctx, "SyntheticChecks.Run")
	defer span.End()

	checks := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"AuthCheck", w.checkAuth},
		{"VehicleLookup", w.checkVehicleLookup},
		{"WorkOrderCreation", w.checkWorkOrderCreation},
		{"TelemetryIngestion", w.checkTelemetryIngestion},
	}

	for _, c := range checks {
		start := time.Now()
		status := "UP"
		var errMsg string

		err := c.fn(ctx)
		if err != nil {
			status = "DOWN"
			errMsg = err.Error()
			w.logger.Error("Synthetic check failed", zap.String("check", c.name), zap.Error(err))
		}

		duration := time.Since(start).Milliseconds()

		checkResult := &models.SyntheticCheck{
			CheckName:      c.name,
			Status:         status,
			ResponseTimeMS: int(duration),
			Error:          errMsg,
			Timestamp:      time.Now(),
		}

		if err := w.repo.SaveSyntheticCheck(ctx, checkResult); err != nil {
			w.logger.Error("Failed to save synthetic check result", zap.Error(err))
		}
	}
}

func (w *SyntheticCheckWorker) checkAuth(ctx context.Context) error {
	// Simulated Auth request
	return nil // Mocked for now
}

func (w *SyntheticCheckWorker) checkVehicleLookup(ctx context.Context) error {
	// Simulated Vehicle Lookup
	return nil // Mocked for now
}

func (w *SyntheticCheckWorker) checkWorkOrderCreation(ctx context.Context) error {
	// Simulated Work Order Creation
	return nil // Mocked for now
}

func (w *SyntheticCheckWorker) checkTelemetryIngestion(ctx context.Context) error {
	// Simulated Telemetry Ingestion
	return nil // Mocked for now
}
