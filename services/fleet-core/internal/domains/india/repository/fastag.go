package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/india/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *IndiaIntegrationRepositoryImpl) CreateFASTagTransaction(ctx context.Context, transaction *models.FASTagTransaction) error {
	query := `
		INSERT INTO fastag_transactions (
			id, tag_number, vehicle_id, driver_id, tenant_id, transaction_id, transaction_type,
			amount, toll_plaza, toll_plaza_id, lane_id, transaction_time, transaction_date,
			settlement_date, status, error_code, error_description, reconciled,
			reconciliation_date, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`

	_, err := r.db.Exec(ctx, query,
		transaction.ID,
		transaction.TagNumber,
		transaction.VehicleID,
		transaction.DriverID,
		transaction.TenantID,
		transaction.TransactionID,
		transaction.TransactionType,
		transaction.Amount,
		transaction.TollPlaza,
		transaction.TollPlazaID,
		transaction.LaneID,
		transaction.TransactionTime,
		transaction.TransactionDate,
		transaction.SettlementDate,
		transaction.Status,
		transaction.ErrorCode,
		transaction.ErrorDescription,
		transaction.Reconciled,
		transaction.ReconciliationDate,
		transaction.CreatedAt,
		transaction.CreatedBy,
		transaction.UpdatedAt,
		transaction.UpdatedBy,
	)

	return err
}

// GetFASTagTransaction retrieves a FASTag transaction by ID
func (r *IndiaIntegrationRepositoryImpl) GetFASTagTransaction(ctx context.Context, id string) (*models.FASTagTransaction, error) {
	query := `
		SELECT id, tag_number, vehicle_id, driver_id, tenant_id, transaction_id, transaction_type,
		       amount, toll_plaza, toll_plaza_id, lane_id, transaction_time, transaction_date,
		       settlement_date, status, error_code, error_description, reconciled,
		       reconciliation_date, created_at, created_by, updated_at, updated_by
		FROM fastag_transactions
		WHERE id = $1`

	var transaction models.FASTagTransaction

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&transaction.ID,
		&transaction.TagNumber,
		&transaction.VehicleID,
		&transaction.DriverID,
		&transaction.TenantID,
		&transaction.TransactionID,
		&transaction.TransactionType,
		&transaction.Amount,
		&transaction.TollPlaza,
		&transaction.TollPlazaID,
		&transaction.LaneID,
		&transaction.TransactionTime,
		&transaction.TransactionDate,
		&transaction.SettlementDate,
		&transaction.Status,
		&transaction.ErrorCode,
		&transaction.ErrorDescription,
		&transaction.Reconciled,
		&transaction.ReconciliationDate,
		&transaction.CreatedAt,
		&transaction.CreatedBy,
		&transaction.UpdatedAt,
		&transaction.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("fastag transaction not found")
		}
		return nil, fmt.Errorf("failed to get fastag transaction: %w", err)
	}

	return &transaction, nil
}

// UpdateFASTagTransaction updates an existing FASTag transaction
func (r *IndiaIntegrationRepositoryImpl) UpdateFASTagTransaction(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setClause := ""
	args := []interface{}{}
	argIndex := 1

	for key, value := range updates {
		// Security check against SQL injection
		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(key) {
			continue
		}
		if key == "updated_at" {
			continue
		}
		if argIndex > 1 {
			setClause += ", "
		}
		setClause += fmt.Sprintf("%s = $%d", key, argIndex)
		args = append(args, value)
		argIndex++
	}

	if setClause != "" {
		setClause += ", "
	}
	setClause += fmt.Sprintf("updated_at = $%d", argIndex)
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE fastag_transactions SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteFASTagTransaction deletes a FASTag transaction
func (r *IndiaIntegrationRepositoryImpl) DeleteFASTagTransaction(ctx context.Context, id string) error {
	query := "DELETE FROM fastag_transactions WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete fastag transaction: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("fastag transaction not found")
	}

	return nil
}

// ListFASTagTransactions retrieves FASTag transactions with filtering
func (r *IndiaIntegrationRepositoryImpl) ListFASTagTransactions(ctx context.Context, vehicleID *uuid.UUID, tagNumber *string, startDate, endDate *time.Time, limit, offset int) ([]*models.FASTagTransaction, int, error) {
	countQuery := "SELECT COUNT(*) FROM fastag_transactions WHERE true"
	args := []interface{}{}
	argIndex := 1

	whereClause := "WHERE true"
	if vehicleID != nil {
		whereClause += fmt.Sprintf(" AND vehicle_id = $%d", argIndex)
		args = append(args, vehicleID)
		argIndex++
	}
	if tagNumber != nil {
		whereClause += fmt.Sprintf(" AND tag_number = $%d", argIndex)
		args = append(args, tagNumber)
		argIndex++
	}
	if startDate != nil {
		whereClause += fmt.Sprintf(" AND transaction_time >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}
	if endDate != nil {
		whereClause += fmt.Sprintf(" AND transaction_time <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	var totalCount int
	err := r.db.QueryRow(ctx, fmt.Sprintf("%s %s", countQuery, whereClause), args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count fastag transactions: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, tag_number, vehicle_id, driver_id, tenant_id, transaction_id, transaction_type,
		       amount, toll_plaza, toll_plaza_id, lane_id, transaction_time, transaction_date,
		       settlement_date, status, error_code, error_description, reconciled,
		       reconciliation_date, created_at, created_by, updated_at, updated_by
		FROM fastag_transactions
		%s
		ORDER BY transaction_time DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list fastag transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.FASTagTransaction
	for rows.Next() {
		var transaction models.FASTagTransaction

		err := rows.Scan(
			&transaction.ID,
			&transaction.TagNumber,
			&transaction.VehicleID,
			&transaction.DriverID,
			&transaction.TenantID,
			&transaction.TransactionID,
			&transaction.TransactionType,
			&transaction.Amount,
			&transaction.TollPlaza,
			&transaction.TollPlazaID,
			&transaction.LaneID,
			&transaction.TransactionTime,
			&transaction.TransactionDate,
			&transaction.SettlementDate,
			&transaction.Status,
			&transaction.ErrorCode,
			&transaction.ErrorDescription,
			&transaction.Reconciled,
			&transaction.ReconciliationDate,
			&transaction.CreatedAt,
			&transaction.CreatedBy,
			&transaction.UpdatedAt,
			&transaction.UpdatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan fastag transaction: %w", err)
		}

		transactions = append(transactions, &transaction)
	}

	return transactions, totalCount, nil
}

// CreateEWayBill creates a new E-Way Bill
