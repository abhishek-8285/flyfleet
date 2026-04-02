package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateMessage creates a new driver message
func (r *DriverRepository) CreateMessage(ctx context.Context, msg *models.DriverMessage) error {
	query := `
		INSERT INTO driver_messages (
			id, driver_id, trip_id, direction, type, content, attachment_url, is_read, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`

	id := uuid.New().String()
	now := time.Now()

	start := time.Now()
	err := r.db.QueryRow(ctx, query,
		id, msg.DriverID, msg.TripID, msg.Direction, msg.Type,
		msg.Content, msg.AttachmentURL, msg.IsRead, now,
	).Scan(&msg.ID, &msg.CreatedAt)
	metrics.RecordDBQuery(query, time.Since(start))

	return err
}

// GetMessages retrieves messages for a driver
func (r *DriverRepository) GetMessages(ctx context.Context, driverID string, limit, offset int) ([]*models.DriverMessage, error) {
	query := `
		SELECT id, driver_id, trip_id, direction, type, content, attachment_url, is_read, created_at
		FROM driver_messages
		WHERE driver_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, driverID, limit, offset)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.DriverMessage
	for rows.Next() {
		var msg models.DriverMessage
		err := rows.Scan(
			&msg.ID, &msg.DriverID, &msg.TripID, &msg.Direction, &msg.Type,
			&msg.Content, &msg.AttachmentURL, &msg.IsRead, &msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// CreateExpense creates a new driver expense record
func (r *DriverRepository) CreateExpense(ctx context.Context, expense *models.DriverExpense) error {
	query := `
		INSERT INTO driver_expenses (
			id, driver_id, trip_id, type, amount, currency, receipt_url, 
			notes, status, transaction_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()

	start := time.Now()
	err := r.db.QueryRow(ctx, query,
		id, expense.DriverID, expense.TripID, expense.Type, expense.Amount,
		expense.Currency, expense.ReceiptURL, expense.Notes, expense.Status,
		expense.TransactionDate, now, now,
	).Scan(&expense.ID, &expense.CreatedAt, &expense.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))

	return err
}

// ListExpenses retrieves expenses for a driver
func (r *DriverRepository) ListExpenses(ctx context.Context, driverID string) ([]*models.DriverExpense, error) {
	query := `
		SELECT id, driver_id, trip_id, type, amount, currency, receipt_url,
			   notes, status, transaction_date, created_at, updated_at
		FROM driver_expenses
		WHERE driver_id = $1
		ORDER BY transaction_date DESC`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, driverID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*models.DriverExpense
	for rows.Next() {
		var expense models.DriverExpense
		err := rows.Scan(
			&expense.ID, &expense.DriverID, &expense.TripID, &expense.Type, &expense.Amount,
			&expense.Currency, &expense.ReceiptURL, &expense.Notes, &expense.Status,
			&expense.TransactionDate, &expense.CreatedAt, &expense.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, &expense)
	}

	return expenses, nil
}

// CreateDutyStatusLog creates a new duty status log
func (r *DriverRepository) CreateDutyStatusLog(ctx context.Context, log *models.DutyStatusLog) error {
	query := `
		INSERT INTO duty_status_logs (
			id, driver_id, status, location, started_at, ended_at, 
			duration_minutes, remarks, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	id := uuid.New().String()
	now := time.Now()
	locationJSON, _ := json.Marshal(log.Location)

	start := time.Now()
	err := r.db.QueryRow(ctx, query,
		id, log.DriverID, log.Status, locationJSON, log.StartedAt,
		log.EndedAt, log.DurationMinutes, log.Remarks, now, now,
	).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))

	return err
}

// GetLatestDutyStatus retrieves the latest duty status for a driver
func (r *DriverRepository) GetLatestDutyStatus(ctx context.Context, driverID string) (*models.DutyStatusLog, error) {
	query := `
		SELECT id, driver_id, status, location, started_at, ended_at,
			   duration_minutes, remarks, created_at, updated_at
		FROM duty_status_logs
		WHERE driver_id = $1
		ORDER BY started_at DESC
		LIMIT 1`

	var log models.DutyStatusLog
	var locationJSON []byte

	start := time.Now()
	err := r.db.QueryRow(ctx, query, driverID).Scan(
		&log.ID, &log.DriverID, &log.Status, &locationJSON, &log.StartedAt,
		&log.EndedAt, &log.DurationMinutes, &log.Remarks, &log.CreatedAt, &log.UpdatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	json.Unmarshal(locationJSON, &log.Location)
	return &log, nil
}

// EndDutyStatusLog updates a duty status log with end time and duration
func (r *DriverRepository) EndDutyStatusLog(ctx context.Context, logID string, durationMinutes int) error {
	query := `
		UPDATE duty_status_logs
		SET ended_at = NOW(), duration_minutes = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, durationMinutes, logID)
	return err
}

// GetDutyStatusLogs retrieves duty status logs for a driver within a date range
func (r *DriverRepository) GetDutyStatusLogs(ctx context.Context, driverID string, startDate, endDate time.Time) ([]*models.DutyStatusLog, error) {
	query := `
		SELECT id, driver_id, status, location, started_at, ended_at,
			   duration_minutes, remarks, created_at, updated_at
		FROM duty_status_logs
		WHERE driver_id = $1 AND started_at BETWEEN $2 AND $3
		ORDER BY started_at DESC
	`

	rows, err := r.db.Query(ctx, query, driverID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.DutyStatusLog
	for rows.Next() {
		var log models.DutyStatusLog
		var locationJSON []byte
		err := rows.Scan(
			&log.ID, &log.DriverID, &log.Status, &locationJSON, &log.StartedAt,
			&log.EndedAt, &log.DurationMinutes, &log.Remarks, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(locationJSON, &log.Location)
		logs = append(logs, &log)
	}
	return logs, nil
}
