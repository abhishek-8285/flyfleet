package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/support/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/support/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TicketService interface {
	CreateTicket(ctx context.Context, req *models.CreateTicketRequest, orgID, reporterID uuid.UUID) (*models.SupportTicket, error)
	GetTicket(ctx context.Context, id uuid.UUID) (*models.SupportTicket, error)
	ListTickets(ctx context.Context, filter models.TicketFilter) ([]*models.SupportTicket, int, error)
	UpdateTicketStatus(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID) error
	AssignTicket(ctx context.Context, id, assignedTo, userID uuid.UUID) error
	AddComment(ctx context.Context, ticketID, userID uuid.UUID, req *models.AddCommentRequest, authorName string) error
	GetTicketComments(ctx context.Context, ticketID uuid.UUID) ([]*models.IncidentComment, error)
	GetRecentActivity(ctx context.Context, orgID uuid.UUID, limit int) ([]*models.SupportTicket, error)
	GetDashboardSummary(ctx context.Context, orgID uuid.UUID) (*models.DashboardSummary, error)
	GetOrgMetrics(ctx context.Context, orgID uuid.UUID) (*models.OrgSupportMetrics, error)
}

type ticketService struct {
	repo            repository.TicketRepository
	logger          *zap.Logger
	notificationCh  chan<- NotificationEvent // Channel for sending notifications
}

// NotificationEvent represents a notification to be sent
type NotificationEvent struct {
	Type       string    `json:"type"` // COMMENT_ADDED, STATUS_CHANGED, TICKET_ASSIGNED
	TicketID   string    `json:"ticket_id"`
	UserID     string    `json:"user_id"`
	Message    string    `json:"message"`
	Priority   string    `json:"priority"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewTicketService(repo repository.TicketRepository, logger *zap.Logger, notificationCh chan<- NotificationEvent) TicketService {
	return &ticketService{
		repo:            repo,
		logger:          logger,
		notificationCh:  notificationCh,
	}
}

func (s *ticketService) CreateTicket(ctx context.Context, req *models.CreateTicketRequest, orgID, reporterID uuid.UUID) (*models.SupportTicket, error) {
	// Generate unique ticket number using database sequence
	ticketNum, err := s.repo.GetNextTicketNumber(ctx)
	if err != nil {
		s.logger.Error("Failed to generate ticket number", zap.Error(err))
		return nil, fmt.Errorf("failed to generate ticket number: %w", err)
	}
	ticketID := fmt.Sprintf("INC-%d", ticketNum)

	userID := s.getUserID(ctx)

	// Set default priority if not provided
	priority := req.Priority
	if priority == "" {
		priority = "MEDIUM"
	}

	// Calculate SLA deadline based on priority
	slaDeadline := s.calculateSLADeadline(priority)

	ticket := &models.SupportTicket{
		ID:          uuid.New(),
		TicketID:    ticketID,
		OrgID:       orgID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "OPEN",
		Priority:    priority,
		Category:    req.Category,
		ReporterID:  reporterID,
		CreatedBy:   userID,
		UpdatedBy:   userID,
		Tags:        req.Tags,
		Metadata:    make(map[string]interface{}),
		SLADeadline: &slaDeadline,
	}

	err = s.repo.CreateTicket(ctx, ticket)
	if err != nil {
		s.logger.Error("Failed to create ticket", zap.Error(err))
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	return ticket, nil
}

// calculateSLADeadline calculates the SLA deadline based on priority
// CRITICAL: 4 hours, HIGH: 24 hours, MEDIUM: 72 hours, LOW: 168 hours (7 days)
func (s *ticketService) calculateSLADeadline(priority string) time.Time {
	slaHours := map[string]int{
		"CRITICAL": 4,
		"HIGH":     24,
		"MEDIUM":   72,
		"LOW":      168, // 7 days
	}

	hours, ok := slaHours[priority]
	if !ok {
		hours = 72 // Default to MEDIUM
	}

	return time.Now().Add(time.Duration(hours) * time.Hour)
}

func (s *ticketService) GetTicket(ctx context.Context, id uuid.UUID) (*models.SupportTicket, error) {
	ticket, err := s.repo.GetTicketByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get ticket", zap.Error(err), zap.String("ticket_id", id.String()))
		return nil, fmt.Errorf("ticket not found: %w", err)
	}
	return ticket, nil
}

func (s *ticketService) ListTickets(ctx context.Context, filter models.TicketFilter) ([]*models.SupportTicket, int, error) {
	tickets, totalCount, err := s.repo.ListTickets(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to list tickets", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list tickets: %w", err)
	}
	return tickets, totalCount, nil
}

func (s *ticketService) UpdateTicketStatus(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID) error {
	// Validate status
	validStatuses := map[string]bool{
		"OPEN": true, "IN_PROGRESS": true, "RESOLVED": true, 
		"CLOSED": true, "ESCALATED": true,
	}
	
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	err := s.repo.UpdateTicketStatus(ctx, id, status)
	if err != nil {
		s.logger.Error("Failed to update ticket status", zap.Error(err), 
			zap.String("ticket_id", id.String()), zap.String("status", status))
		return fmt.Errorf("failed to update ticket status: %w", err)
	}

	return nil
}

func (s *ticketService) AssignTicket(ctx context.Context, id, assignedTo, userID uuid.UUID) error {
	err := s.repo.AssignTicket(ctx, id, assignedTo)
	if err != nil {
		s.logger.Error("Failed to assign ticket", zap.Error(err), 
			zap.String("ticket_id", id.String()), zap.String("assigned_to", assignedTo.String()))
		return fmt.Errorf("failed to assign ticket: %w", err)
	}

	return nil
}

func (s *ticketService) AddComment(ctx context.Context, ticketID, userID uuid.UUID, req *models.AddCommentRequest, authorName string) error {
	comment := &models.IncidentComment{
		ID:          uuid.New(),
		TicketID:    ticketID,
		AuthorID:    &userID,
		AuthorName:  authorName,
		CommentType: "INTERNAL",
		Content:     req.Content,
		IsPrivate:   req.IsPrivate,
		CreatedBy:   userID,
		UpdatedBy:   userID,
	}

	err := s.repo.AddComment(ctx, comment)
	if err != nil {
		s.logger.Error("Failed to add comment", zap.Error(err))
		return fmt.Errorf("failed to add comment: %w", err)
	}

	// Send notification for non-private comments
	if !req.IsPrivate && s.notificationCh != nil {
		ticket, err := s.repo.GetTicketByID(ctx, ticketID)
		if err != nil {
			s.logger.Warn("Failed to get ticket for notification", zap.Error(err))
		} else {
			s.notificationCh <- NotificationEvent{
				Type:      "COMMENT_ADDED",
				TicketID:  ticket.TicketID,
				UserID:    userID.String(),
				Message:   fmt.Sprintf("New comment on ticket %s by %s", ticket.TicketID, authorName),
				Priority:  ticket.Priority,
				CreatedAt: time.Now(),
			}
		}
	}

	return nil
}

func (s *ticketService) getUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value("user_id")
	if val == nil {
		return uuid.Nil
	}
	uidStr, ok := val.(string)
	if !ok {
		return uuid.Nil
	}
	uid, _ := uuid.Parse(uidStr)
	return uid
}

func (s *ticketService) GetTicketComments(ctx context.Context, ticketID uuid.UUID) ([]*models.IncidentComment, error) {
	comments, err := s.repo.GetTicketComments(ctx, ticketID)
	if err != nil {
		s.logger.Error("Failed to get ticket comments", zap.Error(err), zap.String("ticket_id", ticketID.String()))
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	return comments, nil
}

func (s *ticketService) GetRecentActivity(ctx context.Context, orgID uuid.UUID, limit int) ([]*models.SupportTicket, error) {
	if limit <= 0 {
		limit = 10
	}
	
	tickets, err := s.repo.GetRecentActivity(ctx, orgID, limit)
	if err != nil {
		s.logger.Error("Failed to get recent activity", zap.Error(err))
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	return tickets, nil
}

func (s *ticketService) GetDashboardSummary(ctx context.Context, orgID uuid.UUID) (*models.DashboardSummary, error) {
	summary, err := s.repo.GetDashboardSummary(ctx, orgID)
	if err != nil {
		s.logger.Error("Failed to get dashboard summary", zap.Error(err))
		return nil, fmt.Errorf("failed to get dashboard summary: %w", err)
	}
	return summary, nil
}

func (s *ticketService) GetOrgMetrics(ctx context.Context, orgID uuid.UUID) (*models.OrgSupportMetrics, error) {
	metrics, err := s.repo.GetOrgMetrics(ctx, orgID)
	if err != nil {
		// Return default metrics if not found
		return &models.OrgSupportMetrics{
			OrgID:             orgID,
			OpenTickets:       0,
			ResolvedTickets:   0,
			SLABreaches:       0,
			CriticalIssues:    0,
		}, nil
	}
	return metrics, nil
}