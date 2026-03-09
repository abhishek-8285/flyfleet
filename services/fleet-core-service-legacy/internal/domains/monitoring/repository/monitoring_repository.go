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
