package admin_support

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Workers contains all background workers for the Admin Support Module.
// Each worker is designed as a long-running goroutine, accepting a done channel
// for graceful shutdown coordination.
type Workers struct {
	repo    Repository
	metrics *Metrics
	logger  *zap.Logger
}

func NewWorkers(repo Repository, metrics *Metrics, logger *zap.Logger) *Workers {
	return &Workers{
		repo:    repo,
		metrics: metrics,
		logger:  logger,
	}
}

// StartAll launches all background workers. It returns immediately; the workers
// run in the background until ctx is cancelled.
func (w *Workers) StartAll(ctx context.Context) {
	go w.runSLABreachMonitor(ctx)
	go w.runIncidentMonitor(ctx)
	go w.runMaintenanceScheduler(ctx)
	w.logger.Info("admin support background workers started")
}

// ─── SLA Breach Monitor ───────────────────────────────────────────────────────
// Runs every 5 minutes. Detects open tickets that have passed their SLA deadline
// and emits an escalation alert (log + metric increment).

func (w *Workers) runSLABreachMonitor(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Run once immediately on startup
	w.checkSLABreaches(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("SLA breach monitor stopped")
			return
		case <-ticker.C:
			w.checkSLABreaches(ctx)
		}
	}
}

func (w *Workers) checkSLABreaches(ctx context.Context) {
	tickets, err := w.repo.ListSLABreachedTickets(ctx)
	if err != nil {
		w.logger.Error("SLA breach monitor: failed to query tickets", zap.Error(err))
		return
	}

	if len(tickets) == 0 {
		return
	}

	for _, t := range tickets {
		overdueBy := time.Since(t.SLADeadline)
		w.logger.Warn("SLA breach detected",
			zap.String("ticket_id", t.ID.String()),
			zap.String("priority", string(t.Priority)),
			zap.String("org_id", t.OrganizationID.String()),
			zap.Duration("overdue_by", overdueBy),
		)
		w.metrics.TicketsOpenTotal.Inc() // Reuse existing gauge; a dedicated breach counter is Phase 3
	}

	w.logger.Info("SLA breach monitor: scan complete",
		zap.Int("breached_count", len(tickets)),
	)
}

// ─── Incident Monitor ─────────────────────────────────────────────────────────
// Runs every minute. Notifies if any SEV1 incident has been unresolved for
// longer than 15 minutes to trigger escalation paging.

const sev1EscalationThreshold = 15 * time.Minute

func (w *Workers) runIncidentMonitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Run once immediately on startup
	w.checkUnresolvedSEV1(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("incident monitor stopped")
			return
		case <-ticker.C:
			w.checkUnresolvedSEV1(ctx)
		}
	}
}

func (w *Workers) checkUnresolvedSEV1(ctx context.Context) {
	incidents, err := w.repo.ListUnresolvedSEV1Incidents(ctx)
	if err != nil {
		w.logger.Error("incident monitor: failed to query incidents", zap.Error(err))
		return
	}

	for _, inc := range incidents {
		duration := time.Since(inc.DetectedAt)
		if duration >= sev1EscalationThreshold {
			w.logger.Error("CRITICAL: Unresolved SEV1 incident exceeds threshold — escalating",
				zap.String("incident_id", inc.ID.String()),
				zap.Duration("open_for", duration),
				zap.String("description", inc.Description),
			)
			// In production: trigger PagerDuty re-page or webhook here
		}
	}

	if len(incidents) > 0 {
		w.logger.Warn("incident monitor: active SEV1 incidents detected",
			zap.Int("count", len(incidents)),
		)
	}
}

// ─── Maintenance Scheduler ────────────────────────────────────────────────────
// Runs every minute. Activates windows that have started, logs expired windows,
// and updates the system_flag "maintenance_active_<service>" accordingly.

func (w *Workers) runMaintenanceScheduler(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Run once immediately on startup
	w.processMaintenanceWindows(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("maintenance scheduler stopped")
			return
		case <-ticker.C:
			w.processMaintenanceWindows(ctx)
		}
	}
}

func (w *Workers) processMaintenanceWindows(ctx context.Context) {
	// 1. Activate windows currently in range
	active, err := w.repo.ListActiveMaintenanceWindows(ctx)
	if err != nil {
		w.logger.Error("maintenance scheduler: failed to query active windows", zap.Error(err))
		return
	}

	for _, win := range active {
		flagKey := "maintenance_active_" + win.Service
		if err := w.repo.SetSystemFlag(ctx, &SystemFlag{
			Key:       flagKey,
			Value:     "true",
			UpdatedAt: time.Now(),
		}); err != nil {
			w.logger.Error("maintenance scheduler: failed to set flag",
				zap.String("service", win.Service),
				zap.Error(err),
			)
		} else {
			w.logger.Info("maintenance window active",
				zap.String("service", win.Service),
				zap.Time("ends_at", win.EndTime),
			)
		}
	}

	// 2. Deactivate expired windows
	expiredCount, err := w.repo.DeactivateExpiredMaintenanceWindows(ctx)
	if err != nil {
		w.logger.Error("maintenance scheduler: expired window check failed", zap.Error(err))
		return
	}

	if expiredCount > 0 {
		w.logger.Info("maintenance scheduler: expired windows detected",
			zap.Int64("count", expiredCount),
		)
	}
}
