package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/repository"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NotificationService struct {
	nc     *nats.Conn
	repo   *repository.NotificationRepository
	logger *zap.Logger
}

func NewNotificationService(nc *nats.Conn, repo *repository.NotificationRepository, logger *zap.Logger) *NotificationService {
	return &NotificationService{
		nc:     nc,
		repo:   repo,
		logger: logger,
	}
}

// PublishEvent sends a billing notification to NATS
func (s *NotificationService) PublishEvent(ctx context.Context, eventType string, orgID uuid.UUID, data map[string]interface{}) error {
	event := models.NotificationEvent{
		Type:      eventType,
		OrgID:     orgID,
		Timestamp: time.Now(),
		Data:      data,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("billing.v2.%s", eventType)
	return s.nc.Publish(subject, payload)
}

type DunningService struct {
	repo         *repository.NotificationRepository
	notifService *NotificationService
	billService  *BillingService
	logger       *zap.Logger
}

func NewDunningService(repo *repository.NotificationRepository, notifService *NotificationService, billService *BillingService, logger *zap.Logger) *DunningService {
	return &DunningService{
		repo:         repo,
		notifService: notifService,
		billService:  billService,
		logger:       logger,
	}
}

// HandlePaymentFailure records the attempt and triggers retry logic (Dunning)
func (s *DunningService) HandlePaymentFailure(ctx context.Context, orgID uuid.UUID, invoiceID uuid.UUID, reason string, amount float64, currency string) error {
	// 1. Get current attempt count
	currentCount, err := s.repo.GetLatestAttemptNumber(ctx, invoiceID)
	if err != nil {
		return err
	}

	// 2. Record new failed attempt
	attempt := &models.PaymentAttempt{
		OrgID:         orgID,
		InvoiceID:     invoiceID,
		AttemptNumber: currentCount + 1,
		Status:        models.AttemptStatusFailed,
		FailureReason: reason,
		Amount:        amount,
		Currency:      currency,
	}

	if err := s.repo.RecordPaymentAttempt(ctx, attempt); err != nil {
		return err
	}

	// 3. Fire NATS notification for failure
	_ = s.notifService.PublishEvent(ctx, "payment_failed", orgID, map[string]interface{}{
		"invoice_id":     invoiceID,
		"attempt_number": attempt.AttemptNumber,
		"reason":         reason,
	})

	// 4. Implement simple dunning logic (e.g., mark as overdue if 3 attempts failed)
	if attempt.AttemptNumber >= 3 {
		s.logger.Warn("Dunning threshold reached for invoice", zap.String("invoice_id", invoiceID.String()), zap.Int("attempts", attempt.AttemptNumber))
		// Log a specific ledger/audit event or move to a separate collections queue
	}

	return nil
}

// HandlePaymentSuccess records the successful attempt
func (s *DunningService) HandlePaymentSuccess(ctx context.Context, orgID uuid.UUID, invoiceID uuid.UUID, amount float64, currency string) error {
	currentCount, _ := s.repo.GetLatestAttemptNumber(ctx, invoiceID)
	
	attempt := &models.PaymentAttempt{
		OrgID:         orgID,
		InvoiceID:     invoiceID,
		AttemptNumber: currentCount + 1,
		Status:        models.AttemptStatusSuccess,
		Amount:        amount,
		Currency:      currency,
	}

	if err := s.repo.RecordPaymentAttempt(ctx, attempt); err != nil {
		return err
	}

	// Fire NATS notification for success
	return s.notifService.PublishEvent(ctx, "payment_success", orgID, map[string]interface{}{
		"invoice_id": invoiceID,
		"amount":     amount,
	})
}
