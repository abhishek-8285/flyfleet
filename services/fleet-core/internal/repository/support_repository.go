package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type SupportTicket struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Subject        string     `json:"subject"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	AssignedTo     *uuid.UUID `json:"assigned_to"`
	ReporterID     *uuid.UUID `json:"reporter_id"`
	ReporterName   string     `json:"reporter_name"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
}

type TicketEvent struct {
	ID         uuid.UUID              `json:"id"`
	TicketID   uuid.UUID              `json:"ticket_id"`
	AuthorID   *uuid.UUID             `json:"author_id"`
	AuthorName string                 `json:"author_name"`
	EventType  string                 `json:"event_type"`
	Message    string                 `json:"message"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

type UserSession struct {
	ID         uuid.UUID `json:"sessionId"`
	UserID     uuid.UUID `json:"user_id"`
	Device     string    `json:"device"`
	Browser    string    `json:"browser"`
	IP         string    `json:"ip"`
	Location   string    `json:"location"`
	Current    bool      `json:"current"`
	LastActive time.Time `json:"last_active"`
}

type UserNote struct {
	Agent string    `json:"agent"`
	Time  time.Time `json:"time"`
	Note  string    `json:"note"`
}

type SecurityHistory struct {
	Time      time.Time `json:"time"`
	EventType string    `json:"eventType"`
	IP        string    `json:"ip"`
	Location  string    `json:"location"`
	RiskScore int       `json:"riskScore"`
	Result    string    `json:"result"`
}

type SupportRepository interface {
	GetTickets(ctx context.Context, orgID uuid.UUID) ([]*SupportTicket, error)
	GetTicketDetail(ctx context.Context, ticketID uuid.UUID) (*SupportTicket, []*TicketEvent, error)
	GetDashboardMetrics(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error)
	AddTicketEvent(ctx context.Context, event *TicketEvent) error
	LogAudit(ctx context.Context, log *AuditLog) error

	// User Ops
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*UserSession, error)
	GetUserNotes(ctx context.Context, userID uuid.UUID) ([]*UserNote, error)
	AddUserNote(ctx context.Context, userID uuid.UUID, agent string, note string) error
	GetLoginHistory(ctx context.Context, userID uuid.UUID) ([]*SecurityHistory, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error

	// Analytics
	GetActivityTrend(ctx context.Context, days int) ([]map[string]interface{}, error)
	GetRecentAuditFailures(ctx context.Context, limit int) ([]map[string]interface{}, error)
}

type AuditLog struct {
	OrganizationID uuid.UUID              `json:"organization_id"`
	UserID         uuid.UUID              `json:"user_id"`
	Action         string                 `json:"action"`
	TargetUserID   *uuid.UUID             `json:"target_user_id"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	Status         string                 `json:"status"`
	ErrorMessage   string                 `json:"error_message"`
	Metadata       map[string]interface{} `json:"metadata"`
}

type supportRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewSupportRepository(db *pgxpool.Pool, logger *zap.Logger) SupportRepository {
	return &supportRepository{db: db, logger: logger}
}

func (r *supportRepository) GetTickets(ctx context.Context, orgID uuid.UUID) ([]*SupportTicket, error) {
	query := `SELECT id, organization_id, subject, status, priority, reporter_name, created_at, assigned_to 
	          FROM support_tickets WHERE organization_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*SupportTicket
	for rows.Next() {
		t := &SupportTicket{}
		err := rows.Scan(&t.ID, &t.OrganizationID, &t.Subject, &t.Status, &t.Priority, &t.ReporterName, &t.CreatedAt, &t.AssignedTo)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, nil
}

func (r *supportRepository) GetTicketDetail(ctx context.Context, ticketID uuid.UUID) (*SupportTicket, []*TicketEvent, error) {
	t := &SupportTicket{}
	err := r.db.QueryRow(ctx, "SELECT id, organization_id, subject, description, status, priority, assigned_to, reporter_name, created_at FROM support_tickets WHERE id = $1", ticketID).
		Scan(&t.ID, &t.OrganizationID, &t.Subject, &t.Description, &t.Status, &t.Priority, &t.AssignedTo, &t.ReporterName, &t.CreatedAt)
	if err != nil {
		return nil, nil, err
	}

	rows, err := r.db.Query(ctx, "SELECT id, author_name, event_type, message, metadata, created_at FROM ticket_events WHERE ticket_id = $1 ORDER BY created_at DESC", ticketID)
	if err != nil {
		return t, nil, nil
	}
	defer rows.Close()

	var events []*TicketEvent
	for rows.Next() {
		e := &TicketEvent{TicketID: ticketID}
		err := rows.Scan(&e.ID, &e.AuthorName, &e.EventType, &e.Message, &e.Metadata, &e.CreatedAt)
		if err != nil {
			continue
		}
		events = append(events, e)
	}
	return t, events, nil
}

func (r *supportRepository) GetDashboardMetrics(ctx context.Context, orgID uuid.UUID) (map[string]interface{}, error) {
	var total, open, resolvedToday int
	var err error
	
	if orgID == uuid.Nil {
		err = r.db.QueryRow(ctx, "SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'OPEN'), COUNT(*) FILTER (WHERE status = 'RESOLVED' AND resolved_at >= CURRENT_DATE) FROM support_tickets").
			Scan(&total, &open, &resolvedToday)
	} else {
		err = r.db.QueryRow(ctx, "SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'OPEN'), COUNT(*) FILTER (WHERE status = 'RESOLVED' AND resolved_at >= CURRENT_DATE) FROM support_tickets WHERE organization_id = $1", orgID).
			Scan(&total, &open, &resolvedToday)
	}
	
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_incidents": total,
		"new_today":       open, // Simplified for metrics
		"critical_open":   open,
		"high_priority":   open,
		"in_progress":     open,
		"resolved_24h":    resolvedToday,
		"mttr_display":    "1h 15m", // Placeholder for now
	}, nil
}

func (r *supportRepository) AddTicketEvent(ctx context.Context, event *TicketEvent) error {
	query := `INSERT INTO ticket_events (ticket_id, author_id, author_name, event_type, message, metadata) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, event.TicketID, event.AuthorID, event.AuthorName, event.EventType, event.Message, event.Metadata)
	return err
}

func (r *supportRepository) LogAudit(ctx context.Context, l *AuditLog) error {
	query := `INSERT INTO system_audit_logs (organization_id, user_id, action, target_user_id, ip_address, user_agent, status, error_message, metadata) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query, l.OrganizationID, l.UserID, l.Action, l.TargetUserID, l.IPAddress, l.UserAgent, l.Status, l.ErrorMessage, l.Metadata)
	return err
}

func (r *supportRepository) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*UserSession, error) {
	query := `SELECT id, device, browser, ip_address, location, is_current, last_active FROM user_sessions WHERE user_id = $1 ORDER BY last_active DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*UserSession
	for rows.Next() {
		s := &UserSession{UserID: userID}
		err := rows.Scan(&s.ID, &s.Device, &s.Browser, &s.IP, &s.Location, &s.Current, &s.LastActive)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *supportRepository) GetUserNotes(ctx context.Context, userID uuid.UUID) ([]*UserNote, error) {
	query := `SELECT agent_name, created_at, note FROM user_notes WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*UserNote
	for rows.Next() {
		n := &UserNote{}
		err := rows.Scan(&n.Agent, &n.Time, &n.Note)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (r *supportRepository) AddUserNote(ctx context.Context, userID uuid.UUID, agent string, note string) error {
	query := `INSERT INTO user_notes (user_id, agent_name, note) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, userID, agent, note)
	return err
}

func (r *supportRepository) GetLoginHistory(ctx context.Context, userID uuid.UUID) ([]*SecurityHistory, error) {
	query := `SELECT created_at, event_type, ip_address, location, risk_score, result FROM security_history WHERE user_id = $1 ORDER BY created_at DESC LIMIT 10`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*SecurityHistory
	for rows.Next() {
		h := &SecurityHistory{}
		err := rows.Scan(&h.Time, &h.EventType, &h.IP, &h.Location, &h.RiskScore, &h.Result)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func (r *supportRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM user_sessions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, sessionID)
	return err
}

func (r *supportRepository) GetActivityTrend(ctx context.Context, days int) ([]map[string]interface{}, error) {
	query := `
		SELECT TO_CHAR(created_at, 'Dy') as day, COUNT(*) as count 
		FROM system_audit_logs 
		WHERE created_at > CURRENT_DATE - INTERVAL '1 day' * $1 
		GROUP BY TO_CHAR(created_at, 'Dy'), date_trunc('day', created_at)
		ORDER BY date_trunc('day', created_at) ASC
	`
	rows, err := r.db.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []map[string]interface{}
	for rows.Next() {
		var day string
		var count int
		if err := rows.Scan(&day, &count); err == nil {
			points = append(points, map[string]interface{}{"day": day, "count": count})
		}
	}
	return points, nil
}

func (r *supportRepository) GetRecentAuditFailures(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT action as errorCode, error_message as title, COUNT(*) as count, 'CRITICAL' as severity
		FROM system_audit_logs 
		WHERE status = 'FAILURE' 
		GROUP BY action, error_message 
		ORDER BY count DESC 
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var errors []map[string]interface{}
	for rows.Next() {
		var code, title, severity string
		var count int
		if err := rows.Scan(&code, &title, &count, &severity); err == nil {
			if title == "" {
				title = "System Operation Failed"
			}
			errors = append(errors, map[string]interface{}{
				"errorCode": code,
				"title":     title,
				"count":     count,
				"severity":  severity,
			})
		}
	}
	return errors, nil
}
