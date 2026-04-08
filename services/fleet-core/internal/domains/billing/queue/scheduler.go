package queue

import (
	"context"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// InvoicingProvider defines the interface for identifying orgs to bill
type InvoicingProvider interface {
	IdentifyOrganizationsForInvoicing(ctx context.Context) ([]uuid.UUID, error)
}

// BillingScheduler schedules recurring billing tasks
type BillingScheduler struct {
	scheduler *asynq.Scheduler
	provider  InvoicingProvider
	queue     *BillingQueueClient
	logger    *zap.Logger
}

func NewBillingScheduler(redisAddr string, p InvoicingProvider, q *BillingQueueClient, logger *zap.Logger) *BillingScheduler {
	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: redisAddr},
		&asynq.SchedulerOpts{
			Location: nil, // Use UTC
			Logger:   logger.Sugar(),
		},
	)

	return &BillingScheduler{
		scheduler: scheduler,
		provider:  p,
		queue:     q,
		logger:    logger,
	}
}

// RegisterRecurringTasks sets up the periodic billing checks
func (s *BillingScheduler) RegisterRecurringTasks() error {
	// 1. Every 6 hours, check for subscriptions that reached PeriodEnd
	// Note: We use a custom task for this check, which then enqueues individual invoice tasks
	entryID, err := s.scheduler.Register("0 */6 * * *", asynq.NewTask("billing:check_renewals", nil), asynq.Queue("billing"))
	if err != nil {
		return fmt.Errorf("failed to register billing check: %w", err)
	}
	s.logger.Info("Registered billing renewal check", zap.String("entry_id", entryID))

	// 2. Periodic Usage Aggregation (Optional if we do it in real-time)
	return nil
}

// Run starts the scheduler
func (s *BillingScheduler) Run() error {
	return s.scheduler.Run()
}

// HandleCheckRenewals identifies organizations that need invoicing and enqueues tasks
// This is called by the scheduler task handler OR we can just do it in a service method
func (s *BillingScheduler) RunBillingCycle(ctx context.Context) error {
	s.logger.Info("Starting recurring billing cycle check")

	// 1. Identify organizations that need invoicing
	orgIDs, err := s.provider.IdentifyOrganizationsForInvoicing(ctx)
	if err != nil {
		return fmt.Errorf("failed to identify orgs: %w", err)
	}

	s.logger.Info("Identified organizations for invoicing", zap.Int("count", len(orgIDs)))

	// 2. Enqueue individual tasks
	for _, orgID := range orgIDs {
		err := s.queue.EnqueueGenerateInvoice(ctx, models.GenerateInvoicePayload{OrgID: orgID})
		if err != nil {
			s.logger.Error("Failed to enqueue invoice task", zap.String("org_id", orgID.String()), zap.Error(err))
			continue
		}
	}

	return nil
}
