package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/models"
	"github.com/jackc/pgx/v5"
)

func (r *IndiaIntegrationRepositoryImpl) CreateVehicleVerification(ctx context.Context, verification *models.VehicleVerification) error {
	query := `
		INSERT INTO vehicle_verifications (
			id, vehicle_id, org_id, registration_number, status, verification_date,
			chassis_number, engine_number, model, manufacturer, fuel_type, body_type,
			cubic_capacity, seating_capacity, color, manufacturing_year, owner_name,
			owner_address, owner_city, owner_district, owner_state, owner_pin, rc_status,
			insurance_company, insurance_policy, insurance_expiry, puc_number, puc_expiry,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32)`

	_, err := r.db.Exec(ctx, query,
		verification.ID,
		verification.VehicleID,
		verification.OrganizationID,
		verification.RegistrationNumber,
		verification.Status,
		verification.VerificationDate,
		verification.ChassisNumber,
		verification.EngineNumber,
		verification.Model,
		verification.Manufacturer,
		verification.FuelType,
		verification.BodyType,
		verification.CubicCapacity,
		verification.SeatingCapacity,
		verification.Color,
		verification.ManufacturingYear,
		verification.OwnerName,
		verification.OwnerAddress,
		verification.OwnerCity,
		verification.OwnerDistrict,
		verification.OwnerState,
		verification.OwnerPIN,
		verification.RCStatus,
		verification.InsuranceCompany,
		verification.InsurancePolicy,
		verification.InsuranceExpiry,
		verification.PUCNumber,
		verification.PUCExpiry,
		verification.CreatedAt,
		verification.CreatedBy,
		verification.UpdatedAt,
		verification.UpdatedBy,
	)

	return err
}

// GetVehicleVerification retrieves a vehicle verification by ID
func (r *IndiaIntegrationRepositoryImpl) GetVehicleVerification(ctx context.Context, id string) (*models.VehicleVerification, error) {
	query := `
		SELECT id, vehicle_id, org_id, registration_number, status, verification_date,
		       chassis_number, engine_number, model, manufacturer, fuel_type, body_type,
		       cubic_capacity, seating_capacity, color, manufacturing_year, owner_name,
		       owner_address, owner_city, owner_district, owner_state, owner_pin, rc_status,
		       insurance_company, insurance_policy, insurance_expiry, puc_number, puc_expiry,
		       created_at, created_by, updated_at, updated_by
		FROM vehicle_verifications
		WHERE id = $1`

	var verification models.VehicleVerification

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&verification.ID,
		&verification.VehicleID,
		&verification.OrganizationID,
		&verification.RegistrationNumber,
		&verification.Status,
		&verification.VerificationDate,
		&verification.ChassisNumber,
		&verification.EngineNumber,
		&verification.Model,
		&verification.Manufacturer,
		&verification.FuelType,
		&verification.BodyType,
		&verification.CubicCapacity,
		&verification.SeatingCapacity,
		&verification.Color,
		&verification.ManufacturingYear,
		&verification.OwnerName,
		&verification.OwnerAddress,
		&verification.OwnerCity,
		&verification.OwnerDistrict,
		&verification.OwnerState,
		&verification.OwnerPIN,
		&verification.RCStatus,
		&verification.InsuranceCompany,
		&verification.InsurancePolicy,
		&verification.InsuranceExpiry,
		&verification.PUCNumber,
		&verification.PUCExpiry,
		&verification.CreatedAt,
		&verification.CreatedBy,
		&verification.UpdatedAt,
		&verification.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get vehicle verification: %w", err)
	}

	return &verification, nil
}

// GetVehicleVerificationByRegNumber retrieves a vehicle verification by registration number
func (r *IndiaIntegrationRepositoryImpl) GetVehicleVerificationByRegNumber(ctx context.Context, regNumber string) (*models.VehicleVerification, error) {
	query := `
		SELECT id, vehicle_id, org_id, registration_number, status, verification_date,
		       chassis_number, engine_number, model, manufacturer, fuel_type, body_type,
		       cubic_capacity, seating_capacity, color, manufacturing_year, owner_name,
		       owner_address, owner_city, owner_district, owner_state, owner_pin, rc_status,
		       insurance_company, insurance_policy, insurance_expiry, puc_number, puc_expiry,
		       created_at, created_by, updated_at, updated_by
		FROM vehicle_verifications
		WHERE registration_number = $1`

	var verification models.VehicleVerification

	row := r.db.QueryRow(ctx, query, regNumber)
	err := row.Scan(
		&verification.ID,
		&verification.VehicleID,
		&verification.OrganizationID,
		&verification.RegistrationNumber,
		&verification.Status,
		&verification.VerificationDate,
		&verification.ChassisNumber,
		&verification.EngineNumber,
		&verification.Model,
		&verification.Manufacturer,
		&verification.FuelType,
		&verification.BodyType,
		&verification.CubicCapacity,
		&verification.SeatingCapacity,
		&verification.Color,
		&verification.ManufacturingYear,
		&verification.OwnerName,
		&verification.OwnerAddress,
		&verification.OwnerCity,
		&verification.OwnerDistrict,
		&verification.OwnerState,
		&verification.OwnerPIN,
		&verification.RCStatus,
		&verification.InsuranceCompany,
		&verification.InsurancePolicy,
		&verification.InsuranceExpiry,
		&verification.PUCNumber,
		&verification.PUCExpiry,
		&verification.CreatedAt,
		&verification.CreatedBy,
		&verification.UpdatedAt,
		&verification.UpdatedBy,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("vehicle verification not found")
		}
		return nil, fmt.Errorf("failed to get vehicle verification by registration number: %w", err)
	}

	return &verification, nil
}

// UpdateVehicleVerification updates an existing vehicle verification
func (r *IndiaIntegrationRepositoryImpl) UpdateVehicleVerification(ctx context.Context, id string, updates map[string]interface{}) error {
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

	query := fmt.Sprintf("UPDATE vehicle_verifications SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteVehicleVerification deletes a vehicle verification
func (r *IndiaIntegrationRepositoryImpl) DeleteVehicleVerification(ctx context.Context, id string) error {
	query := "DELETE FROM vehicle_verifications WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete vehicle verification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("vehicle verification not found")
	}

	return nil
}

// ListVehicleVerifications retrieves vehicle verifications with filtering
func (r *IndiaIntegrationRepositoryImpl) ListVehicleVerifications(ctx context.Context, regNumber *string, status *string, limit, offset int) ([]*models.VehicleVerification, int, error) {
	countQuery := "SELECT COUNT(*) FROM vehicle_verifications WHERE true"
	args := []interface{}{}
	argIndex := 1

	whereClause := "WHERE true"
	if regNumber != nil {
		whereClause += fmt.Sprintf(" AND registration_number = $%d", argIndex)
		args = append(args, regNumber)
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
		return nil, 0, fmt.Errorf("failed to count vehicle verifications: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, vehicle_id, org_id, registration_number, status, verification_date,
		       chassis_number, engine_number, model, manufacturer, fuel_type, body_type,
		       cubic_capacity, seating_capacity, color, manufacturing_year, owner_name,
		       owner_address, owner_city, owner_district, owner_state, owner_pin, rc_status,
		       insurance_company, insurance_policy, insurance_expiry, puc_number, puc_expiry,
		       created_at, created_by, updated_at, updated_by
		FROM vehicle_verifications
		%s
		ORDER BY verification_date DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list vehicle verifications: %w", err)
	}
	defer rows.Close()

	var verifications []*models.VehicleVerification
	for rows.Next() {
		var verification models.VehicleVerification

		err := rows.Scan(
			&verification.ID,
			&verification.VehicleID,
			&verification.OrganizationID,
			&verification.RegistrationNumber,
			&verification.Status,
			&verification.VerificationDate,
			&verification.ChassisNumber,
			&verification.EngineNumber,
			&verification.Model,
			&verification.Manufacturer,
			&verification.FuelType,
			&verification.BodyType,
			&verification.CubicCapacity,
			&verification.SeatingCapacity,
			&verification.Color,
			&verification.ManufacturingYear,
			&verification.OwnerName,
			&verification.OwnerAddress,
			&verification.OwnerCity,
			&verification.OwnerDistrict,
			&verification.OwnerState,
			&verification.OwnerPIN,
			&verification.RCStatus,
			&verification.InsuranceCompany,
			&verification.InsurancePolicy,
			&verification.InsuranceExpiry,
			&verification.PUCNumber,
			&verification.PUCExpiry,
			&verification.CreatedAt,
			&verification.CreatedBy,
			&verification.UpdatedAt,
			&verification.UpdatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan vehicle verification: %w", err)
		}

		verifications = append(verifications, &verification)
	}

	return verifications, totalCount, nil
}

// CreateDriverVerification creates a new driver verification record
