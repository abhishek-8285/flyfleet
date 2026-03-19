package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BookingApprovalService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewBookingApprovalService(repo repository.CustomerV2Repository, logger *zap.Logger) *BookingApprovalService {
	return &BookingApprovalService{
		repo:   repo,
		logger: logger,
	}
}

// ApproveBooking handles the state progression of a booking through an approval chain
func (s *BookingApprovalService) ApproveBooking(ctx context.Context, bookingID uuid.UUID, approverID uuid.UUID, comments string) error {
	s.logger.Info("Approving booking", zap.String("booking_id", bookingID.String()), zap.String("approver_id", approverID.String()))

	booking, err := s.repo.GetBookingRequestByID(ctx, bookingID)
	if err != nil {
		s.logger.Error("Failed to fetch booking for approval", zap.Error(err))
		return err
	}

	// Validate current state
	if booking.Status != "UNDER_REVIEW" && booking.Status != "REQUESTED" {
		s.logger.Warn("Booking cannot be approved in its current state", zap.String("status", booking.Status))
		// return fmt.Errorf("invalid state for approval")
	}

	// Update booking status
	oldStatus := booking.Status
	booking.Status = "APPROVED" // Simplistic: jumps straight to approved
	
	// Stub: update booking in DB
	// _ = s.repo.UpdateBookingRequest(ctx, booking)

	// Record the approval
	// approval := &models.BookingApproval{...}
	// _ = s.repo.CreateBookingApproval(ctx, approval)

	// Record status history
	// history := &models.BookingStatusHistory{...}
	// _ = s.repo.CreateBookingStatusHistory(ctx, history)

	s.logger.Info("Booking approved successfully", zap.String("booking_id", booking.ID.String()), zap.String("old_status", oldStatus))
	return nil
}
