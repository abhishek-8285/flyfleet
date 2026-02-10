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
}

type CreateTicketRequest struct {
	Title       string   `json:"title" validate:"required"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}

type UpdateTicketStatusRequest struct {
	Status string `json:"status" validate:"required"`
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