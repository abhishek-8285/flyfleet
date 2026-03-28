package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/monitoring/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MonitoringRepository defines the interface for monitoring data operations
type MonitoringRepository interface {
	// Health history operations
	SaveHealthHistory(ctx context.Context, history *models.HealthHistory) error
	GetHealthHistory(ctx context.Context, component string, startTime, endTime time.Time, limit int) ([]*models.HealthHistory, error)
	GetLatestHealthStatus(ctx context.Context, component string) (*models.HealthHistory, error)

	// Alert operations
	SaveAlert(ctx context.Context, alert *models.HealthAlert) error
	GetActiveAlerts(ctx context.Context) ([]*models.HealthAlert, error)
	UpdateAlertStatus(ctx context.Context, alertID, status string) error
	GetAlertThresholds(ctx context.Context) ([]*models.AlertThreshold, error)
	SaveAlertThreshold(ctx context.Context, threshold *models.AlertThreshold) error

	// SRE-grade operations
	SaveSyntheticCheck(ctx context.Context, check *models.SyntheticCheck) error
	GetSLOs(ctx context.Context) ([]*models.ServiceSLO, error)
	GetSLO(ctx context.Context, serviceName string) (*models.ServiceSLO, error)
	SaveSLO(ctx context.Context, slo *models.ServiceSLO) error
	GetErrorBudget(ctx context.Context, serviceName string) (*models.ErrorBudget, error)
	SaveErrorBudget(ctx context.Context, budget *models.ErrorBudget) error

	// Alert grouping & Remediation
	GetOrCreateAlertGroup(ctx context.Context, groupKey string, component, metric, severity string) (*models.AlertGroup, error)
	UpdateAlertGroup(ctx context.Context, group *models.AlertGroup) error
	SaveRemediationAction(ctx context.Context, action *models.RemediationAction) error

	// Incidents
	CreateIncident(ctx context.Context, incident *models.Incident) error
	UpdateIncident(ctx context.Context, incident *models.Incident) error
}

type monitoringRepository struct {
	db *pgxpool.Pool
}

// NewMonitoringRepository creates a new monitoring repository
func NewMonitoringRepository(db *pgxpool.Pool) MonitoringRepository {
	return &monitoringRepository{db: db}
}

// SaveHealthHistory inserts a new health history record
func (r *monitoringRepository) SaveHealthHistory(ctx context.Context, history *models.HealthHistory) error {
	query := `
		INSERT INTO health_history (
			id, timestamp, component, status, response_time_ms, error, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	id := uuid.New().String()
	_, err := r.db.Exec(ctx, query,
		id, history.Timestamp, history.Component, history.Status,
		history.ResponseTime, history.Error, history.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save health history: %w", err)
	}
	return nil
}

// GetHealthHistory retrieves health history for a component
func (r *monitoringRepository) GetHealthHistory(ctx context.Context, component string, startTime, endTime time.Time, limit int) ([]*models.HealthHistory, error) {
	query := `
		SELECT id, timestamp, component, status, response_time_ms, error, created_at
		FROM health_history
		WHERE component = $1
		  AND timestamp >= $2
		  AND timestamp <= $3
		ORDER BY timestamp DESC
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query, component, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query health history: %w", err)
	}
	defer rows.Close()

	var history []*models.HealthHistory
	for rows.Next() {
		var h models.HealthHistory
		err := rows.Scan(
			&h.ID, &h.Timestamp, &h.Component, &h.Status,
			&h.ResponseTime, &h.Error, &h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan health history: %w", err)
		}
		history = append(history, &h)
	}

	return history, nil
}

// GetLatestHealthStatus retrieves the latest health status for a component
func (r *monitoringRepository) GetLatestHealthStatus(ctx context.Context, component string) (*models.HealthHistory, error) {
	query := `
		SELECT id, timestamp, component, status, response_time_ms, error, created_at
		FROM health_history
		WHERE component = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var h models.HealthHistory
	err := r.db.QueryRow(ctx, query, component).Scan(
		&h.ID, &h.Timestamp, &h.Component, &h.Status,
		&h.ResponseTime, &h.Error, &h.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get latest health status: %w", err)
	}

	return &h, nil
}

// SaveAlert inserts a new health alert
func (r *monitoringRepository) SaveAlert(ctx context.Context, alert *models.HealthAlert) error {
	query := `
		INSERT INTO health_alerts (
			id, alert_type, component, metric, current_value, threshold,
			severity, message, status, triggered_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	id := uuid.New().String()
	_, err := r.db.Exec(ctx, query,
		id, alert.AlertType, alert.Component, alert.Metric, alert.CurrentValue,
		alert.Threshold, alert.Severity, alert.Message, alert.Status, alert.TriggeredAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save alert: %w", err)
	}
	return nil
}

// GetActiveAlerts retrieves all active alerts
func (r *monitoringRepository) GetActiveAlerts(ctx context.Context) ([]*models.HealthAlert, error) {
	query := `
		SELECT id, alert_type, component, metric, current_value, threshold,
		       severity, message, status, triggered_at, acknowledged_at, resolved_at
		FROM health_alerts
		WHERE status = 'active'
		ORDER BY triggered_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.HealthAlert
	for rows.Next() {
		var a models.HealthAlert
		err := rows.Scan(
			&a.ID, &a.AlertType, &a.Component, &a.Metric, &a.CurrentValue,
			&a.Threshold, &a.Severity, &a.Message, &a.Status,
			&a.TriggeredAt, &a.AcknowledgedAt, &a.ResolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, &a)
	}

	return alerts, nil
}

// UpdateAlertStatus updates the status of an alert
func (r *monitoringRepository) UpdateAlertStatus(ctx context.Context, alertID, status string) error {
	now := time.Now()

	var query string
	if status == "acknowledged" {
		query = `UPDATE health_alerts SET status = $1, acknowledged_at = $2 WHERE id = $3`
		_, err := r.db.Exec(ctx, query, status, now, alertID)
		if err != nil {
			return fmt.Errorf("failed to acknowledge alert: %w", err)
		}
	} else if status == "resolved" {
		query = `UPDATE health_alerts SET status = $1, resolved_at = $2 WHERE id = $3`
		_, err := r.db.Exec(ctx, query, status, now, alertID)
		if err != nil {
			return fmt.Errorf("failed to resolve alert: %w", err)
		}
	} else {
		query = `UPDATE health_alerts SET status = $1 WHERE id = $2`
		_, err := r.db.Exec(ctx, query, status, alertID)
		if err != nil {
			return fmt.Errorf("failed to update alert status: %w", err)
		}
	}

	return nil
}

// GetAlertThresholds retrieves all alert thresholds
func (r *monitoringRepository) GetAlertThresholds(ctx context.Context) ([]*models.AlertThreshold, error) {
	query := `
		SELECT component, metric, threshold, operator, severity, enabled, notification_channels
		FROM alert_thresholds
		WHERE enabled = true
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert thresholds: %w", err)
	}
	defer rows.Close()

	var thresholds []*models.AlertThreshold
	for rows.Next() {
		var t models.AlertThreshold
		err := rows.Scan(
			&t.Component, &t.Metric, &t.Threshold, &t.Operator,
			&t.Severity, &t.Enabled, &t.NotificationChannels,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan threshold: %w", err)
		}
		thresholds = append(thresholds, &t)
	}

	return thresholds, nil
}

// SaveAlertThreshold saves or updates an alert threshold
func (r *monitoringRepository) SaveAlertThreshold(ctx context.Context, threshold *models.AlertThreshold) error {
	query := `
		INSERT INTO alert_thresholds (
			component, metric, threshold, operator, severity, enabled, notification_channels
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (component, metric) DO UPDATE SET
			threshold = EXCLUDED.threshold,
			operator = EXCLUDED.operator,
			severity = EXCLUDED.severity,
			enabled = EXCLUDED.enabled,
			notification_channels = EXCLUDED.notification_channels
	`

	_, err := r.db.Exec(ctx, query,
		threshold.Component, threshold.Metric, threshold.Threshold,
		threshold.Operator, threshold.Severity, threshold.Enabled, threshold.NotificationChannels,
	)

	if err != nil {
		return fmt.Errorf("failed to save alert threshold: %w", err)
	}
	return nil
}

// SaveSyntheticCheck inserts a new synthetic check result
func (r *monitoringRepository) SaveSyntheticCheck(ctx context.Context, check *models.SyntheticCheck) error {
	query := `
		INSERT INTO synthetic_checks (id, check_name, status, response_time_ms, error, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if check.ID == uuid.Nil {
		check.ID = uuid.New()
	}
	_, err := r.db.Exec(ctx, query, check.ID, check.CheckName, check.Status, check.ResponseTimeMS, check.Error, check.Timestamp)
	return err
}

// GetSLOs retrieves all service SLOs
func (r *monitoringRepository) GetSLOs(ctx context.Context) ([]*models.ServiceSLO, error) {
	query := `SELECT id, service_name, availability_target, latency_target_ms, error_rate_target, period_days, created_at, updated_at FROM service_slos`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slos []*models.ServiceSLO
	for rows.Next() {
		var s models.ServiceSLO
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.AvailabilityTarget, &s.LatencyTargetMS, &s.ErrorRateTarget, &s.PeriodDays, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		slos = append(slos, &s)
	}
	return slos, nil
}

// GetSLO retrieves a specific service SLO
func (r *monitoringRepository) GetSLO(ctx context.Context, serviceName string) (*models.ServiceSLO, error) {
	query := `SELECT id, service_name, availability_target, latency_target_ms, error_rate_target, period_days, created_at, updated_at FROM service_slos WHERE service_name = $1`
	var s models.ServiceSLO
	err := r.db.QueryRow(ctx, query, serviceName).Scan(&s.ID, &s.ServiceName, &s.AvailabilityTarget, &s.LatencyTargetMS, &s.ErrorRateTarget, &s.PeriodDays, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SaveSLO saves or updates a service SLO
func (r *monitoringRepository) SaveSLO(ctx context.Context, slo *models.ServiceSLO) error {
	query := `
		INSERT INTO service_slos (id, service_name, availability_target, latency_target_ms, error_rate_target, period_days, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (service_name) DO UPDATE SET
			availability_target = EXCLUDED.availability_target,
			latency_target_ms = EXCLUDED.latency_target_ms,
			error_rate_target = EXCLUDED.error_rate_target,
			period_days = EXCLUDED.period_days,
			updated_at = NOW()
	`
	if slo.ID == uuid.Nil {
		slo.ID = uuid.New()
	}
	_, err := r.db.Exec(ctx, query, slo.ID, slo.ServiceName, slo.AvailabilityTarget, slo.LatencyTargetMS, slo.ErrorRateTarget, slo.PeriodDays)
	return err
}

// GetErrorBudget retrieves the error budget for a service
func (r *monitoringRepository) GetErrorBudget(ctx context.Context, serviceName string) (*models.ErrorBudget, error) {
	query := `SELECT id, service_name, allowed_downtime_minutes, consumed_minutes, period_start, period_end, updated_at FROM error_budgets WHERE service_name = $1`
	var b models.ErrorBudget
	err := r.db.QueryRow(ctx, query, serviceName).Scan(&b.ID, &b.ServiceName, &b.AllowedDowntimeMinutes, &b.ConsumedMinutes, &b.PeriodStart, &b.PeriodEnd, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// SaveErrorBudget saves or updates an error budget
func (r *monitoringRepository) SaveErrorBudget(ctx context.Context, budget *models.ErrorBudget) error {
	query := `
		INSERT INTO error_budgets (id, service_name, allowed_downtime_minutes, consumed_minutes, period_start, period_end, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (service_name) DO UPDATE SET
			allowed_downtime_minutes = EXCLUDED.allowed_downtime_minutes,
			consumed_minutes = EXCLUDED.consumed_minutes,
			period_start = EXCLUDED.period_start,
			period_end = EXCLUDED.period_end,
			updated_at = NOW()
	`
	if budget.ID == uuid.Nil {
		budget.ID = uuid.New()
	}
	_, err := r.db.Exec(ctx, query, budget.ID, budget.ServiceName, budget.AllowedDowntimeMinutes, budget.ConsumedMinutes, budget.PeriodStart, budget.PeriodEnd)
	return err
}

// GetOrCreateAlertGroup retrieves or creates an alert group
func (r *monitoringRepository) GetOrCreateAlertGroup(ctx context.Context, groupKey string, component, metric, severity string) (*models.AlertGroup, error) {
	query := `
		INSERT INTO alert_groups (id, group_key, component, metric_name, severity, status, alert_count, first_triggered, last_triggered)
		VALUES ($1, $2, $3, $4, $5, 'ACTIVE', 1, NOW(), NOW())
		ON CONFLICT (group_key) DO UPDATE SET
			alert_count = alert_groups.alert_count + 1,
			last_triggered = NOW()
		RETURNING id, group_key, component, metric_name, severity, status, alert_count, first_triggered, last_triggered
	`
	var g models.AlertGroup
	err := r.db.QueryRow(ctx, query, uuid.New(), groupKey, component, metric, severity).Scan(
		&g.ID, &g.GroupKey, &g.Component, &g.MetricName, &g.Severity, &g.Status, &g.AlertCount, &g.FirstTriggered, &g.LastTriggered,
	)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// UpdateAlertGroup updates an alert group
func (r *monitoringRepository) UpdateAlertGroup(ctx context.Context, group *models.AlertGroup) error {
	query := `UPDATE alert_groups SET status = $1, alert_count = $2, last_triggered = $3 WHERE group_key = $4`
	_, err := r.db.Exec(ctx, query, group.Status, group.AlertCount, group.LastTriggered, group.GroupKey)
	return err
}

// SaveRemediationAction saves a remediation action
func (r *monitoringRepository) SaveRemediationAction(ctx context.Context, action *models.RemediationAction) error {
	query := `
		INSERT INTO remediation_actions (id, group_key, action_type, status, result_metadata, triggered_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if action.ID == uuid.Nil {
		action.ID = uuid.New()
	}
	_, err := r.db.Exec(ctx, query, action.ID, action.GroupKey, action.ActionType, action.Status, action.ResultMetadata, action.TriggeredAt, action.CompletedAt)
	return err
}

// CreateIncident creates a new incident
func (r *monitoringRepository) CreateIncident(ctx context.Context, incident *models.Incident) error {
	query := `
		INSERT INTO incidents (id, title, severity, status, group_key, service_name, detected_at, external_ref_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if incident.ID == uuid.Nil {
		incident.ID = uuid.New()
	}
	_, err := r.db.Exec(ctx, query, incident.ID, incident.Title, incident.Severity, incident.Status, incident.GroupKey, incident.ServiceName, incident.DetectedAt, incident.ExternalRefID)
	return err
}

// UpdateIncident updates an incident
func (r *monitoringRepository) UpdateIncident(ctx context.Context, incident *models.Incident) error {
	query := `
		UPDATE incidents SET status = $1, acknowledged_at = $2, resolved_at = $3, external_ref_id = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query, incident.Status, incident.AcknowledgedAt, incident.ResolvedAt, incident.ExternalRefID, incident.ID)
	return err
}
