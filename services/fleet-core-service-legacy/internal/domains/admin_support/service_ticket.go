package admin_support

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *service) CreateTicket(ctx context.Context, orgID uuid.UUID, subject string, priority TicketPriority, createdBy uuid.UUID) (*SupportTicket, error) {
	deadline := s.calculateSLADeadline(priority)
	ticket := &SupportTicket{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Subject:        subject,
		Priority:       priority,
		Status:         StatusOpen,
		CreatedBy:      createdBy,
		SLADeadline:    deadline,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.CreateTicket(ctx, ticket); err != nil {
		s.logger.Error("failed to create support ticket", zap.Error(err))
		return nil, err
	}

	s.metrics.TicketsOpenTotal.Inc()
	s.logger.Info("support ticket created", zap.String("ticket_id", ticket.ID.String()), zap.String("org_id", orgID.String()))
	return ticket, nil
}

func (s *service) GetTicket(ctx context.Context, id uuid.UUID) (*SupportTicket, error) {
	return s.repo.GetTicket(ctx, id)
}

func (s *service) ListTickets(ctx context.Context, orgID *uuid.UUID) ([]*SupportTicket, error) {
	if orgID != nil {
		return s.repo.ListTicketsByOrg(ctx, *orgID)
	}
	return s.repo.ListAllTickets(ctx)
}

func (s *service) AssignTicket(ctx context.Context, ticketID uuid.UUID, adminID uuid.UUID) error {
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return err
	}

	ticket.AssignedTo = &adminID
	ticket.Status = StatusInProgress
	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		s.logger.Error("failed to assign ticket", zap.Error(err))
		return err
	}

	s.logger.Info("support ticket assigned", zap.String("ticket_id", ticketID.String()), zap.String("admin_id", adminID.String()))
	return nil
}

func (s *service) ReplyToTicket(ctx context.Context, ticketID uuid.UUID, senderID uuid.UUID, message string) error {
	msg := &TicketMessage{
		ID:        uuid.New(),
		TicketID:  ticketID,
		SenderID:  senderID,
		Message:   message,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateTicketMessage(ctx, msg); err != nil {
		return err
	}

	// Also update ticket status if needed (e.g., to WAITING_CUSTOMER if sent by admin)
	// For now, just update updated_at
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err == nil {
		ticket.UpdatedAt = time.Now()
		_ = s.repo.UpdateTicket(ctx, ticket)
	}

	return nil
}

func (s *service) ResolveTicket(ctx context.Context, ticketID uuid.UUID) error {
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return err
	}

	ticket.Status = StatusResolved
	if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
		return err
	}

	s.metrics.TicketsOpenTotal.Dec()
	s.metrics.TicketsResolvedTotal.Inc()
	s.logger.Info("support ticket resolved", zap.String("ticket_id", ticketID.String()))
	return nil
}

func (s *service) AddTicketAttachment(ctx context.Context, ticketID uuid.UUID, fileURL string, uploadedBy uuid.UUID) (*SupportTicketAttachment, error) {
	att := &SupportTicketAttachment{
		ID:         uuid.New(),
		TicketID:   ticketID,
		FileURL:    fileURL,
		UploadedBy: uploadedBy,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateTicketAttachment(ctx, att); err != nil {
		s.logger.Error("failed to add ticket attachment", zap.Error(err))
		return nil, err
	}

	s.logger.Info("ticket attachment added", zap.String("ticket_id", ticketID.String()), zap.String("file_url", fileURL))
	return att, nil
}

func (s *service) ListTicketAttachments(ctx context.Context, ticketID uuid.UUID) ([]*SupportTicketAttachment, error) {
	return s.repo.ListTicketAttachments(ctx, ticketID)
}

func (s *service) calculateSLADeadline(priority TicketPriority) time.Time {
	now := time.Now()
	switch priority {
	case PriorityCritical:
		return now.Add(1 * time.Hour)
	case PriorityHigh:
		return now.Add(6 * time.Hour)
	case PriorityMedium:
		return now.Add(24 * time.Hour)
	case PriorityLow:
		fallthrough
	default:
		return now.Add(72 * time.Hour)
	}
}
