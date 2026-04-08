package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"go.uber.org/zap"
)

type DocumentProvider interface {
	GetDriversWithExpiringDocuments(ctx context.Context, days int) ([]*models.Driver, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, topic string, payload interface{}) error
}

type DocumentExpiryWorker struct {
	repo      DocumentProvider
	publisher EventPublisher
	logger    *zap.Logger
}

func NewDocumentExpiryWorker(repo DocumentProvider, publisher EventPublisher, logger *zap.Logger) *DocumentExpiryWorker {
	return &DocumentExpiryWorker{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

func (w *DocumentExpiryWorker) Run(ctx context.Context) {
	w.logger.Info("starting document expiry worker")
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.checkExpiries(ctx)
		}
	}
}

func (w *DocumentExpiryWorker) checkExpiries(ctx context.Context) {
	// Check for 30 days
	w.notifyForExpiry(ctx, 30, "driver.document_expiring")
	// Check for 0 days (expired)
	w.notifyForExpiry(ctx, 0, "driver.document_expired")
}

func (w *DocumentExpiryWorker) notifyForExpiry(ctx context.Context, days int, topic string) {
	drivers, err := w.repo.GetDriversWithExpiringDocuments(ctx, days)
	if err != nil {
		w.logger.Error("failed to get drivers with expiring documents", zap.Error(err), zap.Int("days", days))
		return
	}

	for _, d := range drivers {
		err := w.publisher.Publish(ctx, topic, d)
		if err != nil {
			w.logger.Error("failed to publish expiry event", zap.Error(err), zap.String("driver_id", d.ID))
		}
	}
}
