package admin_support

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SLAMonitorWorker monitors support tickets for SLA violations.
type SLAMonitorWorker struct {
	repo    Repository
	metrics *Metrics
	logger  *zap.Logger
	interval time.Duration
	stopCh   chan struct{}
}

// NewSLAMonitorWorker creates a new SLA monitor worker.
func NewSLAMonitorWorker(repo Repository, metrics *Metrics, logger *zap.Logger) *SLAMonitorWorker {
	return &SLAMonitorWorker{
		repo:     repo,
		metrics:  metrics,
		logger:   logger,
		interval: 5 * time.Minute, // Run every 5 minutes
		stopCh:   make(chan struct{}),
	}
}

// Start begins the SLA monitoring loop.
func (w *SLAMonitorWorker) Start(ctx context.Context) {
	w.logger.Info("SLA monitor worker started", zap.Duration("interval", w.interval))
	
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("SLA monitor worker stopping due to context cancellation")
			return
		case <-w.stopCh:
			w.logger.Info("SLA monitor worker stopping due to stop signal")
			return
		case <-ticker.C:
			w.checkSLAViolations(ctx)
		}
	}
}

// Stop gracefully stops the worker.
func (w *SLAMonitorWorker) Stop() {
	close(w.stopCh)
}

// checkSLAViolations scans for tickets that have exceeded their SLA deadline.
func (w *SLAMonitorWorker) checkSLAViolations(ctx context.Context) {
	w.logger.Debug("checking for SLA violations")

	// Fetch all open tickets
	tickets, err := w.repo.ListAllTickets(ctx)
	if err != nil {
		w.logger.Error("failed to list tickets for SLA check", zap.Error(err))
		return
	}

	now := time.Now()
	violated := 0

	for _, ticket := range tickets {
		// Skip resolved/closed tickets
		if ticket.Status == StatusResolved || ticket.Status == StatusClosed {
			continue
		}

		// Check if SLA deadline has passed
		if now.After(ticket.SLADeadline) {
			violated++
			w.metrics.SLAExceededTotal.Inc()

			// Log the violation
			w.logger.Warn("SLA violation detected",
				zap.String("ticket_id", ticket.ID.String()),
				zap.String("org_id", ticket.OrganizationID.String()),
				zap.String("priority", string(ticket.Priority)),
				zap.Time("deadline", ticket.SLADeadline),
				zap.Duration("overdue", now.Sub(ticket.SLADeadline)),
			)

			// Auto-escalate if not already escalated
			if ticket.Status != StatusEscalated {
				w.escalateTicket(ctx, ticket)
			}
		}
	}

	if violated > 0 {
		w.logger.Warn("SLA violations detected", zap.Int("count", violated))
	}
}

// escalateTicket marks a ticket as escalated and logs the action.
func (w *SLAMonitorWorker) escalateTicket(ctx context.Context, ticket *SupportTicket) {
	ticket.Status = StatusEscalated
	if err := w.repo.UpdateTicket(ctx, ticket); err != nil {
		w.logger.Error("failed to escalate ticket", zap.Error(err))
		return
	}

	// Create an audit log for the escalation
	auditLog := &AuditLog{
		ID:           uuid.New(),
		ActorID:      uuid.Nil, // System action
		Action:       "ticket.escalate",
		ResourceType: "support_ticket",
		ResourceID:   ticket.ID.String(),
		BeforeState:  []byte(`{"status": "` + string(ticket.Status) + `"}`),
		AfterState:   []byte(`{"status": "ESCALATED"}`),
		CreatedAt:    time.Now(),
	}

	if err := w.repo.CreateAuditLog(ctx, auditLog); err != nil {
		w.logger.Error("failed to create audit log for ticket escalation", zap.Error(err))
	}

	// Post a system message to the ticket
	msg := &TicketMessage{
		ID:        uuid.New(),
		TicketID:  ticket.ID,
		SenderID:  uuid.Nil, // System sender
		Message:   "[SYSTEM] This ticket has been automatically escalated due to SLA violation.",
		CreatedAt: time.Now(),
	}

	if err := w.repo.CreateTicketMessage(ctx, msg); err != nil {
		w.logger.Error("failed to post escalation message", zap.Error(err))
	}

	w.logger.Info("ticket escalated", zap.String("ticket_id", ticket.ID.String()))
}
