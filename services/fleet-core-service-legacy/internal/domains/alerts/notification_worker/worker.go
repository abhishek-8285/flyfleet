package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/provider_clients"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/repository"
	"go.uber.org/zap"
)

type NotificationWorker struct {
	repo            repository.Repository
	providerFactory providers.ProviderFactory
	logger          *zap.Logger
	stopChan        chan struct{}
}

func NewNotificationWorker(
	repo repository.Repository,
	providerFactory providers.ProviderFactory,
	logger *zap.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		repo:            repo,
		providerFactory: providerFactory,
		logger:          logger,
		stopChan:        make(chan struct{}),
	}
}

func (w *NotificationWorker) Start() {
	ticker := time.NewTicker(5 * time.Second) // Poll every 5 seconds
	defer ticker.Stop()

	w.logger.Info("Notification worker started")

	for {
		select {
		case <-ticker.C:
			w.processBatch()
		case <-w.stopChan:
			w.logger.Info("Notification worker stopping")
			return
		}
	}
}

func (w *NotificationWorker) Stop() {
	close(w.stopChan)
}

func (w *NotificationWorker) processBatch() {
	ctx := context.Background()
	logs, err := w.repo.ListPendingLogs(ctx, 10) // Process 10 at a time
	if err != nil {
		w.logger.Error("Failed to list pending logs", zap.Error(err))
		return
	}

	for _, log := range logs {
		w.processNotification(ctx, &log)
	}
}

func (w *NotificationWorker) processNotification(ctx context.Context, log *models.NotificationLog) {
	// 1. Get Provider Client
	provider, err := w.providerFactory.GetClient(log.Channel)
	if err != nil {
		w.handlePermanentFailure(ctx, log, fmt.Sprintf("Unsupported channel/provider: %v", err))
		return
	}

	// 2. Attempt Delivery
	w.logger.Debug("Attempting delivery", zap.String("id", log.ID.String()), zap.String("channel", string(log.Channel)))
	
	resp, err := provider.Send(ctx, log)
	log.Provider = provider.Name()
	log.ProviderResponse = resp

	if err != nil {
		w.logger.Warn("Delivery failed", zap.Error(err), zap.String("id", log.ID.String()))
		w.handleRetry(ctx, log, err)
	} else {
		w.logger.Info("Delivery successful", zap.String("id", log.ID.String()))
		w.handleSuccess(ctx, log)
	}
}

func (w *NotificationWorker) handleRetry(ctx context.Context, log *models.NotificationLog, deliverErr error) {
	log.RetryCount++
	log.ErrorMessage = deliverErr.Error()

	if log.RetryCount >= log.MaxRetries {
		w.handlePermanentFailure(ctx, log, fmt.Sprintf("Max retries exceeded: %v", deliverErr))
		return
	}

	// Exponential backoff: 2^retry_count * 1 minute
	backoff := time.Duration(1<<uint(log.RetryCount-1)) * time.Minute
	log.ScheduledAt = time.Now().Add(backoff)
	log.Status = models.StatusQueued

	if err := w.repo.UpdateLog(ctx, log); err != nil {
		w.logger.Error("Failed to update log during retry", zap.Error(err), zap.String("id", log.ID.String()))
	}
}

func (w *NotificationWorker) handleSuccess(ctx context.Context, log *models.NotificationLog) {
	log.Status = models.StatusSent
	if err := w.repo.UpdateLog(ctx, log); err != nil {
		w.logger.Error("Failed to update log as SENT", zap.Error(err), zap.String("id", log.ID.String()))
	}
}

func (w *NotificationWorker) handlePermanentFailure(ctx context.Context, log *models.NotificationLog, reason string) {
	log.Status = models.StatusFailed
	log.ErrorMessage = reason

	if err := w.repo.UpdateLog(ctx, log); err != nil {
		w.logger.Error("Failed to update log as FAILED", zap.Error(err), zap.String("id", log.ID.String()))
	}

	// Add to DLQ
	dlq := &models.DLQEntry{
		NotificationID: log.ID,
		Reason:         reason,
		Payload: map[string]any{
			"recipient": log.Recipient,
			"channel":   log.Channel,
			"subject":   log.Subject,
			"body":      log.Body,
			"metadata":  log.Metadata,
		},
	}

	if err := w.repo.AddToDLQ(ctx, dlq); err != nil {
		w.logger.Error("Failed to add to DLQ", zap.Error(err), zap.String("id", log.ID.String()))
	}
}
