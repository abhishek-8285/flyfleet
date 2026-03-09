package admin_support

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// MaintenanceWorker manages maintenance window scheduling and activation.
type MaintenanceWorker struct {
	repo     Repository
	metrics  *Metrics
	logger   *zap.Logger
	interval time.Duration
	stopCh   chan struct{}
}

// NewMaintenanceWorker creates a new maintenance window worker.
func NewMaintenanceWorker(repo Repository, metrics *Metrics, logger *zap.Logger) *MaintenanceWorker {
	return &MaintenanceWorker{
		repo:     repo,
		metrics:  metrics,
		logger:   logger,
		interval: 1 * time.Minute, // Check every minute
		stopCh:   make(chan struct{}),
	}
}

// Start begins the maintenance window monitoring loop.
func (w *MaintenanceWorker) Start(ctx context.Context) {
	w.logger.Info("Maintenance window worker started", zap.Duration("interval", w.interval))

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Maintenance window worker stopping due to context cancellation")
			return
		case <-w.stopCh:
			w.logger.Info("Maintenance window worker stopping due to stop signal")
			return
		case <-ticker.C:
			w.checkMaintenanceWindows(ctx)
		}
	}
}

// Stop gracefully stops the worker.
func (w *MaintenanceWorker) Stop() {
	close(w.stopCh)
}

// checkMaintenanceWindows activates or deactivates maintenance windows based on current time.
func (w *MaintenanceWorker) checkMaintenanceWindows(ctx context.Context) {
	now := time.Now()

	windows, err := w.repo.ListMaintenanceWindows(ctx)
	if err != nil {
		w.logger.Error("failed to list maintenance windows", zap.Error(err))
		return
	}

	for _, window := range windows {
		// Check if window should be activated
		if now.After(window.StartTime) && now.Before(window.EndTime) {
			w.handleActiveWindow(window)
		} else if now.After(window.EndTime) {
			w.handleExpiredWindow(window)
		}
	}
}

// handleActiveWindow logs and handles an active maintenance window.
func (w *MaintenanceWorker) handleActiveWindow(window *MaintenanceWindow) {
	// In a real system, this would set a system flag to pause the affected service
	w.logger.Warn("MAINTENANCE WINDOW ACTIVE",
		zap.String("window_id", window.ID.String()),
		zap.String("service", window.Service),
		zap.Time("end_time", window.EndTime),
		zap.String("description", window.Description),
	)

	// Set system flag to indicate maintenance mode for the service
	flagKey := "maintenance_" + window.Service
	flag := &SystemFlag{
		Key:       flagKey,
		Value:     "true",
		UpdatedAt: time.Now(),
	}

	if err := w.repo.SetSystemFlag(context.Background(), flag); err != nil {
		w.logger.Error("failed to set maintenance flag", zap.Error(err))
	}
}

// handleExpiredWindow handles an expired maintenance window.
func (w *MaintenanceWorker) handleExpiredWindow(window *MaintenanceWindow) {
	// Clear the maintenance flag
	flagKey := "maintenance_" + window.Service
	flag := &SystemFlag{
		Key:       flagKey,
		Value:     "false",
		UpdatedAt: time.Now(),
	}

	if err := w.repo.SetSystemFlag(context.Background(), flag); err != nil {
		w.logger.Error("failed to clear maintenance flag", zap.Error(err))
	}

	w.logger.Info("MAINTENANCE WINDOW EXPIRED",
		zap.String("window_id", window.ID.String()),
		zap.String("service", window.Service),
	)
}
