package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/tracking/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// AlertRepository handles alert data persistence
type AlertRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *pgxpool.Pool, logger *zap.Logger) *AlertRepository {
	return &AlertRepository{
		db:     db,
		logger: logger,
	}
}

// SaveAlert saves an alert to the database
func (r *AlertRepository) SaveAlert(ctx context.Context, alert *models.VehicleAlert) error {
	const query = `
		INSERT INTO vehicle_alerts (
			id, vehicle_id, tenant_id, alert_type, severity, description,
			latitude, longitude, alert_timestamp, metadata, processed, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
	`

	metadataJSON, _ := json.Marshal(alert.Metadata)

	_, err := r.db.Exec(ctx, query,
		alert.ID,
		alert.VehicleID,
		alert.TenantID,
		alert.Type,
		alert.Severity,
		alert.Description,
		alert.Latitude,
		alert.Longitude,
		alert.Timestamp,
		metadataJSON,
		alert.Processed,
	)

	if err != nil {
		return fmt.Errorf("failed to save alert: %w", err)
	}

	r.logger.Debug("alert saved",
		zap.String("alert_id", alert.ID),
		zap.String("type", alert.Type),
		zap.String("vehicle_id", alert.VehicleID))

	return nil
}

// GetAlerts retrieves alerts for a vehicle in a time range
func (r *AlertRepository) GetAlerts(ctx context.Context, vehicleID string, from, to time.Time, limit int) ([]*models.VehicleAlert, error) {
	const query = `
		SELECT id, vehicle_id, tenant_id, alert_type, severity, description,
		       latitude, longitude, alert_timestamp, metadata, processed, processed_at, created_at
		FROM vehicle_alerts
		WHERE vehicle_id = $1 AND alert_timestamp BETWEEN $2 AND $3
		ORDER BY alert_timestamp DESC
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query, vehicleID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.VehicleAlert
	for rows.Next() {
		var alert models.VehicleAlert
		var metadataJSON []byte
		var processedAt sql.NullTime

		if err := rows.Scan(
			&alert.ID, &alert.VehicleID, &alert.TenantID, &alert.Type, &alert.Severity,
			&alert.Description, &alert.Latitude, &alert.Longitude, &alert.Timestamp,
			&metadataJSON, &alert.Processed, &processedAt, &alert.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &alert.Metadata)
		}

		if processedAt.Valid {
			alert.ProcessedAt = &processedAt.Time
		}

		alerts = append(alerts, &alert)
	}

	return alerts, nil
}

// GetUnprocessedAlerts retrieves unprocessed alerts for a vehicle
func (r *AlertRepository) GetUnprocessedAlerts(ctx context.Context, vehicleID string) ([]*models.VehicleAlert, error) {
	const query = `
		SELECT id, vehicle_id, tenant_id, alert_type, severity, description,
		       latitude, longitude, alert_timestamp, metadata, processed, processed_at, created_at
		FROM vehicle_alerts
		WHERE vehicle_id = $1 AND processed = false
		ORDER BY alert_timestamp ASC
	`

	rows, err := r.db.Query(ctx, query, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query unprocessed alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*models.VehicleAlert
	for rows.Next() {
		var alert models.VehicleAlert
		var metadataJSON []byte
		var processedAt sql.NullTime

		if err := rows.Scan(
			&alert.ID, &alert.VehicleID, &alert.TenantID, &alert.Type, &alert.Severity,
			&alert.Description, &alert.Latitude, &alert.Longitude, &alert.Timestamp,
			&metadataJSON, &alert.Processed, &processedAt, &alert.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &alert.Metadata)
		}

		if processedAt.Valid {
			alert.ProcessedAt = &processedAt.Time
		}

		alerts = append(alerts, &alert)
	}

	return alerts, nil
}

// MarkAlertProcessed marks an alert as processed
func (r *AlertRepository) MarkAlertProcessed(ctx context.Context, alertID string) error {
	const query = `
		UPDATE vehicle_alerts
		SET processed = true, processed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, alertID)
	if err != nil {
		return fmt.Errorf("failed to mark alert processed: %w", err)
	}

	r.logger.Debug("alert marked as processed", zap.String("alert_id", alertID))
	return nil
}

// GetAlertCount returns the count of alerts for a vehicle in a time range
func (r *AlertRepository) GetAlertCount(ctx context.Context, vehicleID string, from, to time.Time) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM vehicle_alerts
		WHERE vehicle_id = $1 AND alert_timestamp BETWEEN $2 AND $3
	`

	var count int64
	err := r.db.QueryRow(ctx, query, vehicleID, from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	return count, nil
}

// DeleteOldAlerts removes alerts older than the specified date
func (r *AlertRepository) DeleteOldAlerts(ctx context.Context, before time.Time, processedOnly bool) (int64, error) {
	var query string
	if processedOnly {
		query = `DELETE FROM vehicle_alerts WHERE alert_timestamp < $1 AND processed = true`
	} else {
		query = `DELETE FROM vehicle_alerts WHERE alert_timestamp < $1`
	}

	result, err := r.db.Exec(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old alerts: %w", err)
	}

	return result.RowsAffected(), nil
}
