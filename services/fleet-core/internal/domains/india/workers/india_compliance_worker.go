package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/messaging"
	"github.com/flyfleet/fleet-core-service/internal/domains/india/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type IndiaComplianceWorker struct {
	repo   repository.IndiaIntegrationRepository
	mb     messaging.MessageBroker
	redis  *redis.Client
	logger *zap.Logger
}

func NewIndiaComplianceWorker(repo repository.IndiaIntegrationRepository, mb messaging.MessageBroker, rdb *redis.Client, logger *zap.Logger) *IndiaComplianceWorker {
	return &IndiaComplianceWorker{
		repo:   repo,
		mb:     mb,
		redis:  rdb,
		logger: logger,
	}
}

func (w *IndiaComplianceWorker) Start(ctx context.Context) {
	w.logger.Info("IndiaComplianceWorker started")
	// Run daily
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Initial scan
	w.Process(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("IndiaComplianceWorker stopped")
			return
		case <-ticker.C:
			w.Process(ctx)
		}
	}
}

func (w *IndiaComplianceWorker) Process(ctx context.Context) {
	w.logger.Info("IndiaComplianceWorker: starting compliance scan")

	// 1. Scan Vehicle Expiries (Insurance, PUC)
	vehicles, err := w.repo.GetExpiringVehicleVerifications(ctx, 30) // 30 days threshold
	if err != nil {
		w.logger.Error("failed to fetch expiring vehicle verifications", zap.Error(err))
	} else {
		for _, v := range vehicles {
			w.publishEvent("india.compliance.vehicle_expiry", v)

			// If insurance is already expired, mark it in Redis for the ingestion worker
			if v.InsuranceExpiry.Before(time.Now()) {
				key := fmt.Sprintf("compliance:expired_insurance:vehicle:%s", v.VehicleID.String())
				if w.redis != nil {
					w.redis.Set(ctx, key, "true", 0) // Permanent until next check or manual renewal
				}
			}
		}
	}

	// 2. Scan Driver License Expiries
	drivers, err := w.repo.GetExpiringDriverVerifications(ctx, 30)
	if err != nil {
		w.logger.Error("failed to fetch expiring driver verifications", zap.Error(err))
	} else {
		for _, d := range drivers {
			w.publishEvent("india.compliance.driver_expiry", d)
		}
	}

	// 3. Scan Permit Expiries
	permits, err := w.repo.GetExpiringLocalPermits(ctx, 30)
	if err != nil {
		w.logger.Error("failed to fetch expiring local permits", zap.Error(err))
	} else {
		for _, p := range permits {
			// We can use vehicle_expiry or a specific permit_expiry if needed,
			// but prompt said vehicle, driver, gst. I'll use vehicle_expiry for permits too or just log.
			w.publishEvent("india.compliance.vehicle_expiry", p)
		}
	}

	// 4. Scan GST Non-compliance
	gstRecords, err := w.repo.GetNonCompliantGSTRecords(ctx)
	if err != nil {
		w.logger.Error("failed to fetch non-compliant GST records", zap.Error(err))
	} else {
		for _, g := range gstRecords {
			w.publishEvent("india.compliance.gst_noncompliant", g)
		}
	}

	w.logger.Info("IndiaComplianceWorker: scan completed")
}

func (w *IndiaComplianceWorker) publishEvent(subject string, data interface{}) {
	if w.mb == nil {
		w.logger.Warn("message broker not available, skipping event publish", zap.String("subject", subject))
		return
	}

	if err := w.mb.PublishJS(subject, data); err != nil {
		w.logger.Error("failed to publish compliance event", zap.String("subject", subject), zap.Error(err))
	} else {
		w.logger.Info("compliance event published", zap.String("subject", subject))
	}
}
