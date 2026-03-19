package models

import (
	"time"
	"github.com/google/uuid"
)

type SupportTicket struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TicketID    string    `json:"ticket_id" db:"ticket_id"`
	OrgID       uuid.UUID `json:"org_id" db:"org_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	Priority    string    `json:"priority" db:"priority"`
	Category    string    `json:"category" db:"category"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty" db:"assigned_to"`
	ReporterID  uuid.UUID `json:"reporter_id" db:"reporter_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	SLADeadline *time.Time `json:"sla_deadline,omitempty" db:"sla_deadline"`
	Tags        []string  `json:"tags" db:"tags"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedBy   uuid.UUID              `json:"created_by" db:"created_by"`
	UpdatedBy   uuid.UUID              `json:"updated_by" db:"updated_by"`
}

type CreateTicketRequest struct {
	Title       string   `json:"title" validate:"required"`
	Description string   `json:"description"`
	Priority    string   `json:"priority" validate:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}

type UpdateTicketStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=OPEN IN_PROGRESS RESOLVED CLOSED ESCALATED"`
}

type AddCommentRequest struct {
	Content   string `json:"content" validate:"required"`
	IsPrivate bool   `json:"is_private"`
}

type AssignTicketRequest struct {
	AssignedTo uuid.UUID `json:"assigned_to" validate:"required"`
}

type TicketFilter struct {
	Status   string `json:"status"`
	Priority string `json:"priority"`
	Category string `json:"category"`
	Query    string `json:"q"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type DashboardSummary struct {
	OpenTickets      int `json:"open_tickets"`
	ResolvedToday    int `json:"resolved_today"`
	SLABreaches      int `json:"sla_breaches"`
	ActiveErrors     int `json:"active_errors"`
	AvgResolutionHrs float64 `json:"avg_resolution_hrs"`
}

type IncidentComment struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TicketID    uuid.UUID `json:"ticket_id" db:"ticket_id"`
	AuthorID    *uuid.UUID `json:"author_id,omitempty" db:"author_id"`
	AuthorName  string    `json:"author_name" db:"author_name"`
	CommentType string    `json:"comment_type" db:"comment_type"`
	Content     string    `json:"content" db:"content"`
	IsPrivate   bool      `json:"is_private" db:"is_private"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy   uuid.UUID `json:"updated_by" db:"updated_by"`
}

type OrgSupportMetrics struct {
	OrgID             uuid.UUID `json:"org_id" db:"org_id"`
	OpenTickets       int       `json:"open_tickets" db:"open_tickets"`
	ResolvedTickets   int       `json:"resolved_tickets" db:"resolved_tickets"`
	AvgResolutionTime float64   `json:"avg_resolution_time" db:"avg_resolution_time"`
	SLABreaches       int       `json:"sla_breaches" db:"sla_breaches"`
	CriticalIssues    int       `json:"critical_issues" db:"critical_issues"`
	LastCalculated    time.Time `json:"last_calculated" db:"last_calculated"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}