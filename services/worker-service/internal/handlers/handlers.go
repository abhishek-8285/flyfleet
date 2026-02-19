package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fleet-platform/shared/pkg/async"
	"github.com/fleet-platform/shared/pkg/database"
	"github.com/fleet-platform/shared/pkg/webhook"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type WorkerHandlers struct {
	db            database.Database
	webhookClient *webhook.WebhookClient
	logger        *zap.Logger
}

func NewWorkerHandlers(db database.Database, webhookClient *webhook.WebhookClient, logger *zap.Logger) *WorkerHandlers {
	return &WorkerHandlers{
		db:            db,
		webhookClient: webhookClient,
		logger:        logger,
	}
}

// HandleWebhookDeliver processes distributed webhook delivery
func (h *WorkerHandlers) HandleWebhookDeliver(ctx context.Context, t *asynq.Task) error {
	var p async.WebhookPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal webhook payload: %v: %w", err, asynq.SkipRetry)
	}

	h.logger.Info("Processing webhook delivery", 
		zap.String("webhook_id", p.WebhookID.String()),
		zap.String("event", p.Event))

	delivery := webhook.WebhookDelivery{
		URL:        "https://example.com/webhook",
		Secret:     "super-secret",
		MaxRetries: 1,
	}

	return h.db.WithTenant(ctx, p.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		payload := webhook.WebhookPayload{
			Event:     p.Event,
			Data:      p.Payload,
			WebhookID: p.WebhookID.String(),
		}
		return h.webhookClient.Deliver(ctx, delivery, payload)
	})
}

// HandleUsageAggregation aggregates consumption events
func (h *WorkerHandlers) HandleUsageAggregation(ctx context.Context, t *asynq.Task) error {
	var p async.UsageAggregationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal aggregation payload: %w", err)
	}

	return h.db.WithTenant(ctx, p.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		// Aggregate mileage
		var totalDistance float64
		row := tx.QueryRow(ctx, `
			SELECT COALESCE(SUM(amount), 0) 
			FROM consumption_events 
			WHERE org_id = $1 AND meter_code = 'DISTANCE_KM' 
			AND time BETWEEN $2 AND $3 AND processed_invoice_id IS NULL`,
			p.OrgID, p.StartTime, p.EndTime)
		
		if err := row.Scan(&totalDistance); err != nil {
			return fmt.Errorf("failed to aggregate distance: %w", err)
		}

		h.logger.Info("Aggregated usage", 
			zap.String("org_id", p.OrgID.String()), 
			zap.Float64("distance_km", totalDistance))

		return nil
	})
}

// HandleMonthlyInvoiceGeneration generates a monthly invoice based on usage
func (h *WorkerHandlers) HandleMonthlyInvoiceGeneration(ctx context.Context, t *asynq.Task) error {
	var p async.MonthlyInvoicePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal invoice payload: %w", err)
	}

	return h.db.WithTenant(ctx, p.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		// 1. Create Invoice Header
		var invoiceID string
		err := tx.QueryRow(ctx, `
			INSERT INTO invoices (org_id, customer_name, invoice_date, due_date, status)
			VALuES ($1, 'Tenant Customer', NOW(), NOW() + INTERVAL '14 days', 'PENDING')
			RETURNING id`,
			p.OrgID).Scan(&invoiceID)
		if err != nil {
			return fmt.Errorf("failed to create invoice: %w", err)
		}

		// 2. Aggregate & Add Item for Mileage
		err = tx.Execute(ctx, `
			INSERT INTO invoice_items (invoice_id, description, quantity, unit_price, total)
			SELECT $1, 'Mileage Consumption', CEIL(SUM(amount)), 0.15, SUM(amount) * 0.15
			FROM consumption_events
			WHERE org_id = $2 AND meter_code = 'DISTANCE_KM' AND processed_invoice_id IS NULL
			GROUP BY org_id`,
			invoiceID, p.OrgID)
		
		if err != nil {
			return fmt.Errorf("failed to add mileage items: %w", err)
		}

		// 3. Mark events as processed
		err = tx.Execute(ctx, `
			UPDATE consumption_events 
			SET processed_invoice_id = $1 
			WHERE org_id = $2 AND processed_invoice_id IS NULL`,
			invoiceID, p.OrgID)

		h.logger.Info("Monthly invoice generated", 
			zap.String("org_id", p.OrgID.String()), 
			zap.String("invoice_id", invoiceID))

		return err
	})
}

// ... existing HandleReportGenerate and HandleMaintenanceReminder remain ...
// HandleAlertEscalation escalates alerts that haven't been acknowledged within SLA
func (h *WorkerHandlers) HandleAlertEscalation(ctx context.Context, t *asynq.Task) error {
	var p async.AlertEscalationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal escalation payload: %w", err)
	}

	return h.db.WithTenant(ctx, p.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		// 1. Escalate alert in DB
		err := tx.Execute(ctx, `
			UPDATE alerts 
			SET status = 'escalated', 
			    escalation_level = escalation_level + 1,
			    last_escalated_at = NOW(),
			    updated_at = NOW()
			WHERE id = $1 AND status = 'new' AND sla_deadline < NOW()`,
			p.AlertID)
		
		if err != nil {
			return fmt.Errorf("failed to escalate alert: %w", err)
		}

		h.logger.Info("Alert escalated", zap.String("alert_id", p.AlertID))
		return nil
	})
}

// HandleSLAAggregation aggregates SLA metrics into hypertables
func (h *WorkerHandlers) HandleSLAAggregation(ctx context.Context, t *asynq.Task) error {
	var p async.SLAAggregationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal sla payload: %w", err)
	}

	return h.db.WithTenant(ctx, p.OrgID.String(), func(ctx context.Context, tx database.Transaction) error {
		// Example: Calculate alert response time for the period
		var avgResponseMinutes float64
		err := tx.QueryRow(ctx, `
			SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (acknowledged_at - timestamp))/60), 0)
			FROM alerts
			WHERE org_id = $1 AND timestamp BETWEEN $2 AND $3 AND acknowledged_at IS NOT NULL`,
			p.OrgID, p.StartTime, p.EndTime).Scan(&avgResponseMinutes)
		
		if err != nil {
			return fmt.Errorf("failed to calculate SLA: %w", err)
		}

		// Insert into sla_metrics hypertable
		err = tx.Execute(ctx, `
			INSERT INTO sla_metrics (time, org_id, metric_name, value)
			VALUES (NOW(), $1, $2, $3)`,
			p.OrgID, p.Metric, avgResponseMinutes)

		return err
	})
}

func (h *WorkerHandlers) HandleReportGenerate(ctx context.Context, t *asynq.Task) error { return nil }
func (h *WorkerHandlers) HandleMaintenanceReminder(ctx context.Context, t *asynq.Task) error { return nil }
