package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *IndiaIntegrationRepositoryImpl) CreateGSTCompliance(ctx context.Context, compliance *models.GSTCompliance) error {
	query := `
		INSERT INTO gst_compliances (
			id, organization_id, gstin, status, legal_name, trade_name, address,
			centre_jurisdiction, state_jurisdiction, date_of_registration, date_of_cancellation,
			last_updated, return_filing_frequency, return_filing_status, payment_status,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`

	_, err := r.db.Exec(ctx, query,
		compliance.ID,
		compliance.OrganizationID,
		compliance.GSTIN,
		compliance.Status,
		compliance.LegalName,
		compliance.TradeName,
		compliance.Address,
		compliance.CentreJurisdiction,
		compliance.StateJurisdiction,
		compliance.DateOfRegistration,
		compliance.DateOfCancellation,
		compliance.LastUpdated,
		compliance.ReturnFilingFrequency,
		compliance.ReturnFilingStatus,
		compliance.PaymentStatus,
		compliance.CreatedAt,
		compliance.CreatedBy,
		compliance.UpdatedAt,
		compliance.UpdatedBy,
	)

	return err
}

// GetGSTCompliance retrieves a GST compliance record by ID
func (r *IndiaIntegrationRepositoryImpl) GetGSTCompliance(ctx context.Context, id string) (*models.GSTCompliance, error) {
	query := `
		SELECT id, organization_id, gstin, status, legal_name, trade_name, address,
		       centre_jurisdiction, state_jurisdiction, date_of_registration, date_of_cancellation,
		       last_updated, return_filing_frequency, return_filing_status, payment_status,
		       created_at, created_by, updated_at, updated_by
		FROM gst_compliances
		WHERE id = $1`

	var compliance models.GSTCompliance

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&compliance.ID,
		&compliance.OrganizationID,
		&compliance.GSTIN,
		&compliance.Status,
		&compliance.LegalName,
		&compliance.TradeName,
		&compliance.Address,
		&compliance.CentreJurisdiction,
		&compliance.StateJurisdiction,
		&compliance.DateOfRegistration,
		&compliance.DateOfCancellation,
		&compliance.LastUpdated,
		&compliance.ReturnFilingFrequency,
		&compliance.ReturnFilingStatus,
		&compliance.PaymentStatus,
		&compliance.CreatedAt,
		&compliance.CreatedBy,
		&compliance.UpdatedAt,
		&compliance.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get gst compliance: %w", err)
	}

	return &compliance, nil
}

// GetGSTComplianceByGSTIN retrieves a GST compliance record by GSTIN
func (r *IndiaIntegrationRepositoryImpl) GetGSTComplianceByGSTIN(ctx context.Context, gstin string) (*models.GSTCompliance, error) {
	query := `
		SELECT id, organization_id, gstin, status, legal_name, trade_name, address,
		       centre_jurisdiction, state_jurisdiction, date_of_registration, date_of_cancellation,
		       last_updated, return_filing_frequency, return_filing_status, payment_status,
		       created_at, created_by, updated_at, updated_by
		FROM gst_compliances
		WHERE gstin = $1`

	var compliance models.GSTCompliance

	row := r.db.QueryRow(ctx, query, gstin)
	err := row.Scan(
		&compliance.ID,
		&compliance.OrganizationID,
		&compliance.GSTIN,
		&compliance.Status,
		&compliance.LegalName,
		&compliance.TradeName,
		&compliance.Address,
		&compliance.CentreJurisdiction,
		&compliance.StateJurisdiction,
		&compliance.DateOfRegistration,
		&compliance.DateOfCancellation,
		&compliance.LastUpdated,
		&compliance.ReturnFilingFrequency,
		&compliance.ReturnFilingStatus,
		&compliance.PaymentStatus,
		&compliance.CreatedAt,
		&compliance.CreatedBy,
		&compliance.UpdatedAt,
		&compliance.UpdatedBy,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("gst compliance not found")
		}
		return nil, fmt.Errorf("failed to get gst compliance by gstin: %w", err)
	}

	return &compliance, nil
}

// UpdateGSTCompliance updates an existing GST compliance record
func (r *IndiaIntegrationRepositoryImpl) UpdateGSTCompliance(ctx context.Context, id string, updates map[string]interface{}) error {
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

	query := fmt.Sprintf("UPDATE gst_compliances SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteGSTCompliance deletes a GST compliance record
func (r *IndiaIntegrationRepositoryImpl) DeleteGSTCompliance(ctx context.Context, id string) error {
	query := "DELETE FROM gst_compliances WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete gst compliance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("gst compliance not found")
	}

	return nil
}

// ListGSTCompliances retrieves GST compliance records with filtering
func (r *IndiaIntegrationRepositoryImpl) ListGSTCompliances(ctx context.Context, orgID uuid.UUID, status *string, limit, offset int) ([]*models.GSTCompliance, int, error) {
	countQuery := "SELECT COUNT(*) FROM gst_compliances WHERE organization_id = $1"
	args := []interface{}{orgID}
	argIndex := 2

	whereClause := "WHERE organization_id = $1"
	if status != nil {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	var totalCount int
	err := r.db.QueryRow(ctx, fmt.Sprintf("%s %s", countQuery, whereClause), args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count gst compliances: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, organization_id, gstin, status, legal_name, trade_name, address,
		       centre_jurisdiction, state_jurisdiction, date_of_registration, date_of_cancellation,
		       last_updated, return_filing_frequency, return_filing_status, payment_status,
		       created_at, created_by, updated_at, updated_by
		FROM gst_compliances
		%s
		ORDER BY last_updated DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list gst compliances: %w", err)
	}
	defer rows.Close()

	var compliances []*models.GSTCompliance
	for rows.Next() {
		var compliance models.GSTCompliance

		err := rows.Scan(
			&compliance.ID,
			&compliance.OrganizationID,
			&compliance.GSTIN,
			&compliance.Status,
			&compliance.LegalName,
			&compliance.TradeName,
			&compliance.Address,
			&compliance.CentreJurisdiction,
			&compliance.StateJurisdiction,
			&compliance.DateOfRegistration,
			&compliance.DateOfCancellation,
			&compliance.LastUpdated,
			&compliance.ReturnFilingFrequency,
			&compliance.ReturnFilingStatus,
			&compliance.PaymentStatus,
			&compliance.CreatedAt,
			&compliance.CreatedBy,
			&compliance.UpdatedAt,
			&compliance.UpdatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan gst compliance: %w", err)
		}

		compliances = append(compliances, &compliance)
	}

	return compliances, totalCount, nil
}

// CreateIntegrationConfig creates a new integration configuration
