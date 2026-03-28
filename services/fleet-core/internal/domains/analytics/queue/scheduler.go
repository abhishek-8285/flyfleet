package queue

import (
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const (
	TypeRefreshFleetHourlyStats  = "analytics:refresh:fleet_hourly"
	TypeRefreshVehicleDailyStats = "analytics:refresh:vehicle_daily"
	TypeRefreshDriverDailyStats  = "analytics:refresh:driver_daily"
	TypeAnalyticsRetentionCycle  = "analytics:retention:cycle"
	TypeAnalyticsReconciliation  = "analytics:ops:reconciliation"
)

// StartScheduler initializes and runs the asynq scheduler for recurring background tasks
func StartScheduler(redisAddr string, logger *zap.Logger) (*asynq.Scheduler, error) {
	location, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, err
	}

	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisAddr},
		&asynq.SchedulerOpts{
			Location: location,
			Logger:   logger.Sugar(),
		},
	)

	// Refresh fleet hourly stats at the top of every hour
	if _, err := scheduler.Register("@hourly", asynq.NewTask(TypeRefreshFleetHourlyStats, nil)); err != nil {
		return nil, err
	}

	// Refresh vehicle daily stats every day at 00:05 UTC
	if _, err := scheduler.Register("5 0 * * *", asynq.NewTask(TypeRefreshVehicleDailyStats, nil)); err != nil {
		return nil, err
	}

	// Refresh driver daily stats every day at 00:10 UTC
	if _, err := scheduler.Register("10 0 * * *", asynq.NewTask(TypeRefreshDriverDailyStats, nil)); err != nil {
		return nil, err
	}

	// Phase 8: Data Retention Cycle (Daily at 01:00 UTC)
	if _, err := scheduler.Register("0 1 * * *", asynq.NewTask(TypeAnalyticsRetentionCycle, nil)); err != nil {
		return nil, err
	}

	// Phase 8: Data Reconciliation (Daily at 02:00 UTC)
	if _, err := scheduler.Register("0 2 * * *", asynq.NewTask(TypeAnalyticsReconciliation, nil)); err != nil {
		return nil, err
	}

	logger.Info("Starting analytics asynq scheduler")
	if err := scheduler.Start(); err != nil {
		return nil, err
	}

	return scheduler, nil
}
