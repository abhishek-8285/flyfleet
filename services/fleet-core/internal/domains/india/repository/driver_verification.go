package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/india/models"
	"github.com/jackc/pgx/v5"
)

func (r *IndiaIntegrationRepositoryImpl) CreateDriverVerification(ctx context.Context, verification *models.DriverVerification) error {
	query := `
		INSERT INTO driver_verifications (
			id, driver_id, tenant_id, license_number, status, verification_date,
			name, father_name, date_of_birth, gender, address, city, district, state, pin,
			license_class, issue_date, validity_from, validity_to, last_transaction,
			enrollment_number, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)`

	licenseClassJSON, err := json.Marshal(verification.LicenseClass)
	if err != nil {
		return fmt.Errorf("failed to marshal license class: %w", err)
	}

	_, err = r.db.Exec(ctx, query,
		verification.ID,
		verification.DriverID,
		verification.OrganizationID,
		verification.LicenseNumber,
		verification.Status,
		verification.VerificationDate,
		verification.Name,
		verification.FatherName,
		verification.DateOfBirth,
		verification.Gender,
		verification.Address,
		verification.City,
		verification.District,
		verification.State,
		verification.PIN,
		licenseClassJSON,
		verification.IssueDate,
		verification.ValidityFrom,
		verification.ValidityTo,
		verification.LastTransaction,
		verification.EnrollmentNumber,
		verification.CreatedAt,
		verification.CreatedBy,
		verification.UpdatedAt,
		verification.UpdatedBy,
	)

	return err
}

// GetDriverVerification retrieves a driver verification by ID
func (r *IndiaIntegrationRepositoryImpl) GetDriverVerification(ctx context.Context, id string) (*models.DriverVerification, error) {
	query := `
		SELECT id, driver_id, tenant_id, license_number, status, verification_date,
		       name, father_name, date_of_birth, gender, address, city, district, state, pin,
		       license_class, issue_date, validity_from, validity_to, last_transaction,
		       enrollment_number, created_at, created_by, updated_at, updated_by
		FROM driver_verifications
		WHERE id = $1`

	var verification models.DriverVerification
	var licenseClassJSON []byte

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&verification.ID,
		&verification.DriverID,
		&verification.OrganizationID,
		&verification.LicenseNumber,
		&verification.Status,
		&verification.VerificationDate,
		&verification.Name,
		&verification.FatherName,
		&verification.DateOfBirth,
		&verification.Gender,
		&verification.Address,
		&verification.City,
		&verification.District,
		&verification.State,
		&verification.PIN,
		&licenseClassJSON,
		&verification.IssueDate,
		&verification.ValidityFrom,
		&verification.ValidityTo,
		&verification.LastTransaction,
		&verification.EnrollmentNumber,
		&verification.CreatedAt,
		&verification.CreatedBy,
		&verification.UpdatedAt,
		&verification.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get driver verification: %w", err)
	}

	// Unmarshal license class
	if err := json.Unmarshal(licenseClassJSON, &verification.LicenseClass); err != nil {
		return nil, fmt.Errorf("failed to unmarshal license class: %w", err)
	}

	return &verification, nil
}

// GetDriverVerificationByLicenseNumber retrieves a driver verification by license number
func (r *IndiaIntegrationRepositoryImpl) GetDriverVerificationByLicenseNumber(ctx context.Context, licenseNumber string) (*models.DriverVerification, error) {
	query := `
		SELECT id, driver_id, tenant_id, license_number, status, verification_date,
		       name, father_name, date_of_birth, gender, address, city, district, state, pin,
		       license_class, issue_date, validity_from, validity_to, last_transaction,
		       enrollment_number, created_at, created_by, updated_at, updated_by
		FROM driver_verifications
		WHERE license_number = $1`

	var verification models.DriverVerification
	var licenseClassJSON []byte

	row := r.db.QueryRow(ctx, query, licenseNumber)
	err := row.Scan(
		&verification.ID,
		&verification.DriverID,
		&verification.OrganizationID,
		&verification.LicenseNumber,
		&verification.Status,
		&verification.VerificationDate,
		&verification.Name,
		&verification.FatherName,
		&verification.DateOfBirth,
		&verification.Gender,
		&verification.Address,
		&verification.City,
		&verification.District,
		&verification.State,
		&verification.PIN,
		&licenseClassJSON,
		&verification.IssueDate,
		&verification.ValidityFrom,
		&verification.ValidityTo,
		&verification.LastTransaction,
		&verification.EnrollmentNumber,
		&verification.CreatedAt,
		&verification.CreatedBy,
		&verification.UpdatedAt,
		&verification.UpdatedBy,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("driver verification not found")
		}
		return nil, fmt.Errorf("failed to get driver verification by license number: %w", err)
	}

	// Unmarshal license class
	if err := json.Unmarshal(licenseClassJSON, &verification.LicenseClass); err != nil {
		return nil, fmt.Errorf("failed to unmarshal license class: %w", err)
	}

	return &verification, nil
}

// UpdateDriverVerification updates an existing driver verification
func (r *IndiaIntegrationRepositoryImpl) UpdateDriverVerification(ctx context.Context, id string, updates map[string]interface{}) error {
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
		// Handle special case for license_class field which needs JSON marshaling
		if key == "license_class" {
			licenseClassBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal license_class: %w", err)
			}
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, licenseClassBytes)
		} else {
			setClause += fmt.Sprintf("%s = $%d", key, argIndex)
			args = append(args, value)
		}
		argIndex++
	}

	if setClause != "" {
		setClause += ", "
	}
	setClause += fmt.Sprintf("updated_at = $%d", argIndex)
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE driver_verifications SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteDriverVerification deletes a driver verification
func (r *IndiaIntegrationRepositoryImpl) DeleteDriverVerification(ctx context.Context, id string) error {
	query := "DELETE FROM driver_verifications WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete driver verification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("driver verification not found")
	}

	return nil
}

// ListDriverVerifications retrieves driver verifications with filtering
func (r *IndiaIntegrationRepositoryImpl) ListDriverVerifications(ctx context.Context, licenseNumber *string, status *string, limit, offset int) ([]*models.DriverVerification, int, error) {
	countQuery := "SELECT COUNT(*) FROM driver_verifications WHERE true"
	args := []interface{}{}
	argIndex := 1

	whereClause := "WHERE true"
	if licenseNumber != nil {
		whereClause += fmt.Sprintf(" AND license_number = $%d", argIndex)
		args = append(args, licenseNumber)
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
		return nil, 0, fmt.Errorf("failed to count driver verifications: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, driver_id, tenant_id, license_number, status, verification_date,
		       name, father_name, date_of_birth, gender, address, city, district, state, pin,
		       license_class, issue_date, validity_from, validity_to, last_transaction,
		       enrollment_number, created_at, created_by, updated_at, updated_by
		FROM driver_verifications
		%s
		ORDER BY verification_date DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list driver verifications: %w", err)
	}
	defer rows.Close()

	var verifications []*models.DriverVerification
	for rows.Next() {
		var verification models.DriverVerification
		var licenseClassJSON []byte

		err := rows.Scan(
			&verification.ID,
			&verification.DriverID,
			&verification.OrganizationID,
			&verification.LicenseNumber,
			&verification.Status,
			&verification.VerificationDate,
			&verification.Name,
			&verification.FatherName,
			&verification.DateOfBirth,
			&verification.Gender,
			&verification.Address,
			&verification.City,
			&verification.District,
			&verification.State,
			&verification.PIN,
			&licenseClassJSON,
			&verification.IssueDate,
			&verification.ValidityFrom,
			&verification.ValidityTo,
			&verification.LastTransaction,
			&verification.EnrollmentNumber,
			&verification.CreatedAt,
			&verification.CreatedBy,
			&verification.UpdatedAt,
			&verification.UpdatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan driver verification: %w", err)
		}

		// Unmarshal license class
		if err := json.Unmarshal(licenseClassJSON, &verification.LicenseClass); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal license class: %w", err)
		}

		verifications = append(verifications, &verification)
	}

	return verifications, totalCount, nil
}

// CreateLocalPermit creates a new local permit
