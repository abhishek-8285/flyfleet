package workers

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/finance/repository"
	financeService "github.com/flyfleet/fleet-core/internal/domains/finance/service"
	"go.uber.org/zap"
)

// ReceiptProcessingWorker processes uploaded receipts asynchronously via OCR.
type ReceiptProcessingWorker struct {
	repo   *repository.FinanceRepository
	ocrSvc *financeService.ReceiptOCRService
	logger *zap.Logger
}

func NewReceiptProcessingWorker(repo *repository.FinanceRepository, ocrSvc *financeService.ReceiptOCRService, logger *zap.Logger) *ReceiptProcessingWorker {
	return &ReceiptProcessingWorker{repo: repo, ocrSvc: ocrSvc, logger: logger}
}

func (w *ReceiptProcessingWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processQueue(ctx)
		}
	}
}

func (w *ReceiptProcessingWorker) processQueue(ctx context.Context) {
	if w.repo == nil || w.ocrSvc == nil {
		return
	}

	items, err := w.repo.ListPendingExpenseItemsForOCR(ctx, 100)
	if err != nil {
		w.logger.Error("ReceiptProcessingWorker: failed to fetch pending expense items", zap.Error(err))
		return
	}
	if len(items) == 0 {
		return
	}

	w.logger.Info("ReceiptProcessingWorker: processing pending receipts", zap.Int("count", len(items)))
	for i := range items {
		item := items[i]
		if err := w.ocrSvc.ProcessExpenseItem(ctx, &item); err != nil {
			w.logger.Warn("ReceiptProcessingWorker: OCR processing failed",
				zap.String("expense_item_id", item.ID.String()),
				zap.Error(err))
		}
	}
}
