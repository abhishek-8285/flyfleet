package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/finance/repository"
	financeService "github.com/flyfleet/fleet-core-service/internal/domains/finance/service"
	"go.uber.org/zap"
)

// FraudRiskWorker periodically scans recent fuel logs for fraud signals.
type FraudRiskWorker struct {
	repo     *repository.FinanceRepository
	fraudSvc *financeService.FraudDetectionService
	alertSvc *financeService.AlertingService
	metrics  *financeService.FinanceMetrics
	logger   *zap.Logger
}

func NewFraudRiskWorker(repo *repository.FinanceRepository, fraudSvc *financeService.FraudDetectionService, alertSvc *financeService.AlertingService, metrics *financeService.FinanceMetrics, logger *zap.Logger) *FraudRiskWorker {
	return &FraudRiskWorker{repo: repo, fraudSvc: fraudSvc, alertSvc: alertSvc, metrics: metrics, logger: logger}
}

func (w *FraudRiskWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.scanForFraud(ctx)
		}
	}
}

func (w *FraudRiskWorker) scanForFraud(ctx context.Context) {
	w.logger.Info("Running fraud detection scan")

	// Fetch recent fuel logs (last 30 minutes)
	since := time.Now().Add(-30 * time.Minute)
	logs, err := w.repo.GetRecentFuelLogsForFraud(ctx, since)
	if err != nil {
		w.logger.Error("Failed to fetch fuel logs for fraud detection", zap.Error(err))
		return
	}

	w.logger.Info("Scanning fuel logs for fraud", zap.Int("count", len(logs)))

	for _, log := range logs {
		score, level, signals, err := w.fraudSvc.ScoreFuelLog(ctx, log.ID, log.OrganizationID)
		if err != nil {
			w.logger.Warn("Failed to score fuel log", zap.String("id", log.ID.String()), zap.Error(err))
			continue
		}

		if score > 0 {
			_, err := w.fraudSvc.CreateFraudIncident(ctx, "fuel_log", log.ID, log.OrganizationID, score, level, signals)
			if err != nil {
				w.logger.Error("Failed to create fraud incident", zap.Error(err))
				continue
			}

			w.metrics.FraudIncidentsTotal.WithLabelValues(
				log.OrganizationID.String(),
				string(level),
			).Inc()
			if w.alertSvc != nil {
				_ = w.alertSvc.Publish(ctx, "fuel.theft.detected", log.OrganizationID, map[string]interface{}{
					"fuel_log_id": log.ID.String(),
					"fraud_score": score,
					"risk_level":  string(level),
					"signals":     signals,
				})
			}

			w.logger.Warn("Fraud incident created",
				zap.String("log_id", log.ID.String()),
				zap.Float64("score", score),
				zap.String("risk_level", string(level)),
			)
		}
	}
}
