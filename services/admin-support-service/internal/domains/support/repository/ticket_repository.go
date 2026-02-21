package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/admin-support-service/internal/domains/support/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type TicketRepository interface {
	CreateTicket(ctx context.Context, ticket *models.SupportTicket) error
	GetTicketByID(ctx context.Context, id uuid.UUID) (*models.SupportTicket, error)
	GetTicketByTicketID(ctx context.Context, ticketID string) (*models.SupportTicket, error)
	ListTickets(ctx context.Context, filter models.TicketFilter) ([]*models.SupportTicket, int, error)
	UpdateTicket(ctx context.Context, ticket *models.SupportTicket) error
	DeleteTicket(ctx context.Context, id uuid.UUID) error
	UpdateTicketStatus(ctx context.Context, id uuid.UUID, status string) error
	AssignTicket(ctx context.Context, id uuid.UUID, assignedTo uuid.UUID) error
	GetRecentActivity(ctx context.Context, orgID uuid.UUID, limit int) ([]*models.SupportTicket, error)
	GetDashboardSummary(ctx context.Context, orgID uuid.UUID) (*models.DashboardSummary, error)
	AddComment(ctx context.Context, comment *models.IncidentComment) error
	GetTicketComments(ctx context.Context, ticketID uuid.UUID) ([]*models.IncidentComment, error)
	GetOrgMetrics(ctx context.Context, orgID uuid.UUID) (*models.OrgSupportMetrics, error)
	UpdateOrgMetrics(ctx context.Context, metrics *models.OrgSupportMetrics) error
}

type ticketRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewTicketRepository(db *pgxpool.Pool, logger *zap.Logger) TicketRepository {
	return &ticketRepository{db: db, logger: logger}
}

func (r *ticketRepository) CreateTicket(ctx context.Context, ticket *models.SupportTicket) error {
	// Validate required fields
	if ticket == nil {
		return fmt.Errorf("ticket cannot be nil")
	}
	if ticket.Title == "" {
		return fmt.Errorf("ticket title is required")
	}
	if ticket.OrgID == uuid.Nil {
		return fmt.Errorf("organization ID is required")
	}
	if ticket.ReporterID == uuid.Nil {
		return fmt.Errorf("reporter ID is required")
	}

	query := `
		INSERT INTO support_tickets (
			id, ticket_id, org_id, title, description, status, priority, category, 
			assigned_to, reporter_id, created_at, created_by, updated_at, updated_by, tags, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	ticket.CreatedAt = now
	ticket.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		ticket.ID, ticket.TicketID, ticket.OrgID, ticket.Title, ticket.Description,
		ticket.Status, ticket.Priority, ticket.Category, ticket.AssignedTo,
		ticket.ReporterID, ticket.CreatedAt, ticket.CreatedBy, ticket.UpdatedAt, ticket.UpdatedBy, ticket.Tags, ticket.Metadata,
	).Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to create ticket", zap.Error(err))
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	// Log audit trail
	if err := r.logAudit(ctx, nil, ticket.ID, ticket.ReporterID, "CREATE", nil, map[string]interface{}{
		"title": ticket.Title, "status": ticket.Status,
	}, "Ticket created"); err != nil {
		r.logger.Warn("Failed to log audit for ticket creation", zap.Error(err))
		// Continue with the operation even if audit logging fails
	}

	return nil
}

func (r *ticketRepository) GetTicketByID(ctx context.Context, id uuid.UUID) (*models.SupportTicket, error) {
	query := `
		SELECT id, ticket_id, org_id, title, description, status, priority, category,
		       assigned_to, reporter_id, created_at, created_by, updated_at, updated_by, resolved_at, 
		       sla_deadline, tags, metadata
		FROM support_tickets 
		WHERE id = $1`

	var ticket models.SupportTicket
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ticket.ID, &ticket.TicketID, &ticket.OrgID, &ticket.Title, &ticket.Description,
		&ticket.Status, &ticket.Priority, &ticket.Category, &ticket.AssignedTo,
		&ticket.ReporterID, &ticket.CreatedAt, &ticket.CreatedBy, &ticket.UpdatedAt, &ticket.UpdatedBy, &ticket.ResolvedAt,
		&ticket.SLADeadline, &ticket.Tags, &ticket.Metadata,
	)

	if err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (r *ticketRepository) GetTicketByTicketID(ctx context.Context, ticketID string) (*models.SupportTicket, error) {
	query := `
		SELECT id, ticket_id, org_id, title, description, status, priority, category,
		       assigned_to, reporter_id, created_at, created_by, updated_at, updated_by, resolved_at, 
		       sla_deadline, tags, metadata
		FROM support_tickets 
		WHERE ticket_id = $1`

	var ticket models.SupportTicket
	err := r.db.QueryRow(ctx, query, ticketID).Scan(
		&ticket.ID, &ticket.TicketID, &ticket.OrgID, &ticket.Title, &ticket.Description,
		&ticket.Status, &ticket.Priority, &ticket.Category, &ticket.AssignedTo,
		&ticket.ReporterID, &ticket.CreatedAt, &ticket.CreatedBy, &ticket.UpdatedAt, &ticket.UpdatedBy, &ticket.ResolvedAt,
		&ticket.SLADeadline, &ticket.Tags, &ticket.Metadata,
	)

	if err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (r *ticketRepository) ListTickets(ctx context.Context, filter models.TicketFilter) ([]*models.SupportTicket, int, error) {
	// Count query
	countQuery := "SELECT COUNT(*) FROM support_tickets WHERE 1=1"
	countArgs := []interface{}{}
	countArgIndex := 1
	var args []interface{}
	var argIndex int

	if filter.Status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", countArgIndex)
		countArgs = append(countArgs, filter.Status)
		// countArgIndex++ // Not needed since we only have one filter condition
	}

	var totalCount int
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Main query
	query := `
		SELECT id, ticket_id, org_id, title, description, status, priority, category,
		       assigned_to, reporter_id, created_at, created_by, updated_at, updated_by, resolved_at, 
		       sla_deadline, tags, metadata
		FROM support_tickets 
		WHERE 1=1`

	args = []interface{}{}
	argIndex = 1

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	// Add pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	
	offset := (filter.Page - 1) * filter.PageSize
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []*models.SupportTicket
	for rows.Next() {
		var ticket models.SupportTicket
		err := rows.Scan(
			&ticket.ID, &ticket.TicketID, &ticket.OrgID, &ticket.Title, &ticket.Description,
			&ticket.Status, &ticket.Priority, &ticket.Category, &ticket.AssignedTo,
			&ticket.ReporterID, &ticket.CreatedAt, &ticket.CreatedBy, &ticket.UpdatedAt, &ticket.UpdatedBy, &ticket.ResolvedAt,
			&ticket.SLADeadline, &ticket.Tags, &ticket.Metadata,
		)
		if err != nil {
			return nil, 0, err
		}
		tickets = append(tickets, &ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return tickets, totalCount, nil
}

func (r *ticketRepository) UpdateTicket(ctx context.Context, ticket *models.SupportTicket) error {
	query := `
		UPDATE support_tickets 
		SET title = $1, description = $2, status = $3, priority = $4, category = $5,
		    assigned_to = $6, updated_at = NOW(), tags = $7, metadata = $8
		WHERE id = $9`

	_, err := r.db.Exec(ctx, query,
		ticket.Title, ticket.Description, ticket.Status, ticket.Priority, ticket.Category,
		ticket.AssignedTo, ticket.Tags, ticket.Metadata, ticket.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update ticket", zap.Error(err), zap.String("ticket_id", ticket.ID.String()))
		return fmt.Errorf("failed to update ticket: %w", err)
	}

	return nil
}

func (r *ticketRepository) DeleteTicket(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM support_tickets WHERE id = $1"
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *ticketRepository) UpdateTicketStatus(ctx context.Context, id uuid.UUID, status string) error {
	// Validate inputs
	if id == uuid.Nil {
		return fmt.Errorf("ticket ID is required")
	}
	if status == "" {
		return fmt.Errorf("status is required")
	}

	// Get current ticket for audit
	currentTicket, err := r.GetTicketByID(ctx, id)
	if err != nil {
		return err
	}

	query := "UPDATE support_tickets SET status = $1, updated_at = NOW()"
	
	// Set resolved_at if status is RESOLVED
	if status == "RESOLVED" {
		query += ", resolved_at = NOW()"
	}
	
	query += " WHERE id = $2"

	_, err = r.db.Exec(ctx, query, status, id)
	if err != nil {
		return err
	}

	// Log audit trail
	if err := r.logAudit(ctx, currentTicket, id, uuid.Nil, "STATUS_CHANGE", 
		map[string]interface{}{"status": currentTicket.Status},
		map[string]interface{}{"status": status},
		fmt.Sprintf("Status changed from %s to %s", currentTicket.Status, status)); err != nil {
		r.logger.Warn("Failed to log audit for status change", zap.Error(err))
		// Continue with the operation even if audit logging fails
	}

	return nil
}

func (r *ticketRepository) AssignTicket(ctx context.Context, id uuid.UUID, assignedTo uuid.UUID) error {
	query := "UPDATE support_tickets SET assigned_to = $1, updated_at = NOW() WHERE id = $2"
	_, err := r.db.Exec(ctx, query, assignedTo, id)
	return err
}

func (r *ticketRepository) GetRecentActivity(ctx context.Context, orgID uuid.UUID, limit int) ([]*models.SupportTicket, error) {
	query := `
		SELECT id, ticket_id, org_id, title, description, status, priority, category,
		       assigned_to, reporter_id, created_at, updated_at, resolved_at, 
		       sla_deadline, tags, metadata
		FROM support_tickets 
		WHERE org_id = $1 AND status != 'RESOLVED'
		ORDER BY updated_at DESC 
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*models.SupportTicket
	for rows.Next() {
		var ticket models.SupportTicket
		err := rows.Scan(
			&ticket.ID, &ticket.TicketID, &ticket.OrgID, &ticket.Title, &ticket.Description,
			&ticket.Status, &ticket.Priority, &ticket.Category, &ticket.AssignedTo,
			&ticket.ReporterID, &ticket.CreatedAt, &ticket.UpdatedAt, &ticket.ResolvedAt,
			&ticket.SLADeadline, &ticket.Tags, &ticket.Metadata,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *ticketRepository) GetDashboardSummary(ctx context.Context, orgID uuid.UUID) (*models.DashboardSummary, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status != 'RESOLVED') as open_tickets,
			COUNT(*) FILTER (WHERE status = 'RESOLVED' AND DATE(resolved_at) = CURRENT_DATE) as resolved_today,
			COUNT(*) FILTER (WHERE sla_deadline < NOW() AND status != 'RESOLVED') as sla_breaches,
			COUNT(*) FILTER (WHERE priority = 'CRITICAL') as active_errors,
			AVG(EXTRACT(EPOCH FROM (resolved_at - created_at))/3600) as avg_resolution_hrs
		FROM support_tickets 
		WHERE org_id = $1`

	var summary models.DashboardSummary
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&summary.OpenTickets, &summary.ResolvedToday, &summary.SLABreaches,
		&summary.ActiveErrors, &summary.AvgResolutionHrs,
	)

	if err != nil {
		return nil, err
	}

	return &summary, nil
}

func (r *ticketRepository) AddComment(ctx context.Context, comment *models.IncidentComment) error {
	// Validate required fields
	if comment == nil {
		return fmt.Errorf("comment cannot be nil")
	}
	if comment.TicketID == uuid.Nil {
		return fmt.Errorf("ticket ID is required")
	}
	if comment.Content == "" {
		return fmt.Errorf("comment content is required")
	}

	query := `
		INSERT INTO incident_comments (
			id, ticket_id, author_id, author_name, comment_type, content, 
			is_private, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		comment.ID, comment.TicketID, comment.AuthorID, comment.AuthorName,
		comment.CommentType, comment.Content, comment.IsPrivate,
		comment.CreatedAt, comment.UpdatedAt,
	)

	return err
}

func (r *ticketRepository) GetTicketComments(ctx context.Context, ticketID uuid.UUID) ([]*models.IncidentComment, error) {
	query := `
		SELECT id, ticket_id, author_id, author_name, comment_type, content, 
		       is_private, created_at, updated_at
		FROM incident_comments 
		WHERE ticket_id = $1 
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.IncidentComment
	for rows.Next() {
		var comment models.IncidentComment
		err := rows.Scan(
			&comment.ID, &comment.TicketID, &comment.AuthorID, &comment.AuthorName,
			&comment.CommentType, &comment.Content, &comment.IsPrivate,
			&comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *ticketRepository) GetOrgMetrics(ctx context.Context, orgID uuid.UUID) (*models.OrgSupportMetrics, error) {
	query := `
		SELECT org_id, open_tickets, resolved_tickets, avg_resolution_time, 
		       sla_breaches, critical_issues, last_calculated, updated_at
		FROM org_support_metrics 
		WHERE org_id = $1`

	var metrics models.OrgSupportMetrics
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&metrics.OrgID, &metrics.OpenTickets, &metrics.ResolvedTickets,
		&metrics.AvgResolutionTime, &metrics.SLABreaches, &metrics.CriticalIssues,
		&metrics.LastCalculated, &metrics.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &metrics, nil
}

func (r *ticketRepository) UpdateOrgMetrics(ctx context.Context, metrics *models.OrgSupportMetrics) error {
	query := `
		INSERT INTO org_support_metrics (
			org_id, open_tickets, resolved_tickets, avg_resolution_time,
			sla_breaches, critical_issues, last_calculated, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (org_id) 
		DO UPDATE SET 
			open_tickets = $2, resolved_tickets = $3, avg_resolution_time = $4,
			sla_breaches = $5, critical_issues = $6, last_calculated = $7, updated_at = $8`

	now := time.Now()
	metrics.LastCalculated = now
	metrics.UpdatedAt = now

	_, err := r.db.Exec(ctx, query,
		metrics.OrgID, metrics.OpenTickets, metrics.ResolvedTickets,
		metrics.AvgResolutionTime, metrics.SLABreaches, metrics.CriticalIssues,
		metrics.LastCalculated, metrics.UpdatedAt,
	)

	return err
}

func (r *ticketRepository) logAudit(ctx context.Context, oldTicket *models.SupportTicket, ticketID uuid.UUID, userID uuid.UUID, action string, _, newValue map[string]interface{}, details string) error {
	query := `
		INSERT INTO support_audit_log (
			id, ticket_id, user_id, action, old_value, new_value, details, created_at, ip_address
		) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW(), '')`

	// Convert oldTicket to JSON if provided
	var oldTicketJSON map[string]interface{}
	if oldTicket != nil {
		oldTicketJSON = map[string]interface{}{
			"title":       oldTicket.Title,
			"status":      oldTicket.Status,
			"priority":    oldTicket.Priority,
			"category":    oldTicket.Category,
			"assigned_to": oldTicket.AssignedTo,
			"updated_at":  oldTicket.UpdatedAt,
		}
	}

	_, err := r.db.Exec(ctx, query, ticketID, userID, action, oldTicketJSON, newValue, details)
	if err != nil {
		r.logger.Error("Failed to log audit", zap.Error(err), 
			zap.String("ticket_id", ticketID.String()), 
			zap.String("action", action))
		return fmt.Errorf("failed to log audit: %w", err)
	}
	return nil
}