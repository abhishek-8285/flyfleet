package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/customer/models"
	"github.com/flyfleet/fleet-core/internal/domains/customer/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DisputeService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewDisputeService(repo repository.CustomerV2Repository, logger *zap.Logger) *DisputeService {
	return &DisputeService{
		repo:   repo,
		logger: logger,
	}
}

// OpenDispute creates a new ticket regarding a shipment or invoice issue
func (s *DisputeService) OpenDispute(ctx context.Context, customerID uuid.UUID, shipmentID *uuid.UUID, invoiceID *uuid.UUID, reason, description string) (*models.Dispute, error) {
	s.logger.Info("Opening a new customer dispute", zap.String("customer_id", customerID.String()))

	desc := description // Copy string
	dispute := &models.Dispute{
		ID:          uuid.New(),
		CustomerID:  customerID,
		ShipmentID:  shipmentID,
		InvoiceID:   invoiceID,
		Status:      "OPEN",
		Reason:      reason,
		Description: &desc,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Pseudo: _ = s.repo.CreateDispute(ctx, dispute)
	return dispute, nil
}

// AddDisputeMessage records a communication thread
func (s *DisputeService) AddDisputeMessage(ctx context.Context, disputeID uuid.UUID, senderID uuid.UUID, senderType, message string) (*models.DisputeMessage, error) {
	s.logger.Info("Adding message to dispute", zap.String("dispute_id", disputeID.String()), zap.String("sender_type", senderType))

	msg := &models.DisputeMessage{
		ID:         uuid.New(),
		DisputeID:  disputeID,
		SenderID:   senderID,
		SenderType: senderType,
		Message:    message,
		SentAt:     time.Now(),
	}

	// Pseudo: _ = s.repo.CreateDisputeMessage(ctx, msg)
	return msg, nil
}
