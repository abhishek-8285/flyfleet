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

func (r *IndiaIntegrationRepositoryImpl) CreateLocalPermit(ctx context.Context, permit *models.LocalPermit) error {
	query := `
		INSERT INTO local_permits (
			id, vehicle_id, driver_id, tenant_id, permit_number, permit_type, permit_category,
			permit_sub_category, valid_from, valid_to, issuing_authority, jurisdiction, route_details,
			number_of_trips, max_load_capacity, number_of_seats, fare, status, remarks,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`

	_, err := r.db.Exec(ctx, query,
		permit.ID,
		permit.VehicleID,
		permit.DriverID,
		permit.TenantID,
		permit.PermitNumber,
		permit.PermitType,
		permit.PermitCategory,
		permit.PermitSubCategory,
		permit.ValidFrom,
		permit.ValidTo,
		permit.IssuingAuthority,
		permit.Jurisdiction,
		permit.RouteDetails,
		permit.NumberOfTrips,
		permit.MaxLoadCapacity,
		permit.NumberOfSeats,
		permit.Fare,
		permit.Status,
		permit.Remarks,
		permit.CreatedAt,
		permit.CreatedBy,
		permit.UpdatedAt,
		permit.UpdatedBy,
	)

	return err
}

// GetLocalPermit retrieves a local permit by ID
func (r *IndiaIntegrationRepositoryImpl) GetLocalPermit(ctx context.Context, id string) (*models.LocalPermit, error) {
	query := `
		SELECT id, vehicle_id, driver_id, tenant_id, permit_number, permit_type, permit_category,
		       permit_sub_category, valid_from, valid_to, issuing_authority, jurisdiction, route_details,
		       number_of_trips, max_load_capacity, number_of_seats, fare, status, remarks,
		       created_at, created_by, updated_at, updated_by
		FROM local_permits
		WHERE id = $1`

	var permit models.LocalPermit

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&permit.ID,
		&permit.VehicleID,
		&permit.DriverID,
		&permit.TenantID,
		&permit.PermitNumber,
		&permit.PermitType,
		&permit.PermitCategory,
		&permit.PermitSubCategory,
		&permit.ValidFrom,
		&permit.ValidTo,
		&permit.IssuingAuthority,
		&permit.Jurisdiction,
		&permit.RouteDetails,
		&permit.NumberOfTrips,
		&permit.MaxLoadCapacity,
		&permit.NumberOfSeats,
		&permit.Fare,
		&permit.Status,
		&permit.Remarks,
		&permit.CreatedAt,
		&permit.CreatedBy,
		&permit.UpdatedAt,
		&permit.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get local permit: %w", err)
	}

	return &permit, nil
}

// GetLocalPermitByNumber retrieves a local permit by permit number
func (r *IndiaIntegrationRepositoryImpl) GetLocalPermitByNumber(ctx context.Context, permitNumber string) (*models.LocalPermit, error) {
	query := `
		SELECT id, vehicle_id, driver_id, tenant_id, permit_number, permit_type, permit_category,
		       permit_sub_category, valid_from, valid_to, issuing_authority, jurisdiction, route_details,
		       number_of_trips, max_load_capacity, number_of_seats, fare, status, remarks,
		       created_at, created_by, updated_at, updated_by
		FROM local_permits
		WHERE permit_number = $1`

	var permit models.LocalPermit

	row := r.db.QueryRow(ctx, query, permitNumber)
	err := row.Scan(
		&permit.ID,
		&permit.VehicleID,
		&permit.DriverID,
		&permit.TenantID,
		&permit.PermitNumber,
		&permit.PermitType,
		&permit.PermitCategory,
		&permit.PermitSubCategory,
		&permit.ValidFrom,
		&permit.ValidTo,
		&permit.IssuingAuthority,
		&permit.Jurisdiction,
		&permit.RouteDetails,
		&permit.NumberOfTrips,
		&permit.MaxLoadCapacity,
		&permit.NumberOfSeats,
		&permit.Fare,
		&permit.Status,
		&permit.Remarks,
		&permit.CreatedAt,
		&permit.CreatedBy,
		&permit.UpdatedAt,
		&permit.UpdatedBy,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("local permit not found")
		}
		return nil, fmt.Errorf("failed to get local permit by number: %w", err)
	}

	return &permit, nil
}

// UpdateLocalPermit updates an existing local permit
func (r *IndiaIntegrationRepositoryImpl) UpdateLocalPermit(ctx context.Context, id string, updates map[string]interface{}) error {
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

	query := fmt.Sprintf("UPDATE local_permits SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteLocalPermit deletes a local permit
func (r *IndiaIntegrationRepositoryImpl) DeleteLocalPermit(ctx context.Context, id string) error {
	query := "DELETE FROM local_permits WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete local permit: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("local permit not found")
	}

	return nil
}

// ListLocalPermits retrieves local permits with filtering
func (r *IndiaIntegrationRepositoryImpl) ListLocalPermits(ctx context.Context, vehicleID *uuid.UUID, permitType *string, status *string, limit, offset int) ([]*models.LocalPermit, int, error) {
	countQuery := "SELECT COUNT(*) FROM local_permits WHERE true"
	args := []interface{}{}
	argIndex := 1

	whereClause := "WHERE true"
	if vehicleID != nil {
		whereClause += fmt.Sprintf(" AND vehicle_id = $%d", argIndex)
		args = append(args, vehicleID)
		argIndex++
	}
	if permitType != nil {
		whereClause += fmt.Sprintf(" AND permit_type = $%d", argIndex)
		args = append(args, permitType)
		argIndex++
	}
	if status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	var totalCount int
	err := r.db.QueryRow(ctx, fmt.Sprintf("%s %s", countQuery, whereClause), args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count local permits: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, vehicle_id, driver_id, tenant_id, permit_number, permit_type, permit_category,
		       permit_sub_category, valid_from, valid_to, issuing_authority, jurisdiction, route_details,
		       number_of_trips, max_load_capacity, number_of_seats, fare, status, remarks,
		       created_at, created_by, updated_at, updated_by
		FROM local_permits
		%s
		ORDER BY valid_from DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list local permits: %w", err)
	}
	defer rows.Close()

	var permits []*models.LocalPermit
	for rows.Next() {
		var permit models.LocalPermit

		err := rows.Scan(
			&permit.ID,
			&permit.VehicleID,
			&permit.DriverID,
			&permit.TenantID,
			&permit.PermitNumber,
			&permit.PermitType,
			&permit.PermitCategory,
			&permit.PermitSubCategory,
			&permit.ValidFrom,
			&permit.ValidTo,
			&permit.IssuingAuthority,
			&permit.Jurisdiction,
			&permit.RouteDetails,
			&permit.NumberOfTrips,
			&permit.MaxLoadCapacity,
			&permit.NumberOfSeats,
			&permit.Fare,
			&permit.Status,
			&permit.Remarks,
			&permit.CreatedAt,
			&permit.CreatedBy,
			&permit.UpdatedAt,
			&permit.UpdatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan local permit: %w", err)
		}

		permits = append(permits, &permit)
	}

	return permits, totalCount, nil
}

// CreateGSTCompliance creates a new GST compliance record
