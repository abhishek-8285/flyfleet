package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/models"
)

func (r *IndiaIntegrationRepositoryImpl) BulkCreateVehicleVerification(ctx context.Context, verifications []*models.VehicleVerification) error {
	if len(verifications) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Batch insert to avoid N+1 queries
	if len(verifications) > 0 {
		var args []interface{}
		var placeholders []string
		for i, v := range verifications {
			offset := i * 21
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9, offset+10,
				offset+11, offset+12, offset+13, offset+14, offset+15, offset+16, offset+17, offset+18, offset+19, offset+20, offset+21))
			args = append(args,
				v.ID, v.VehicleID, v.OrganizationID, v.RegistrationNumber, v.Status, v.VerificationDate,
				v.ChassisNumber, v.EngineNumber, v.Model, v.Manufacturer, v.FuelType, v.BodyType,
				v.CubicCapacity, v.SeatingCapacity, v.Color, v.ManufacturingYear, v.OwnerName,
				v.RCStatus, v.InsuranceExpiry, v.PUCExpiry, v.CreatedAt,
			)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO vehicle_verifications (id, vehicle_id, org_id, registration_number, status, verification_date, "+
				"chassis_number, engine_number, model, manufacturer, fuel_type, body_type, "+
				"cubic_capacity, seating_capacity, color, manufacturing_year, owner_name, "+
				"rc_status, insurance_expiry, puc_expiry, created_at) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		_, err := tx.Exec(ctx, batchQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to batch insert vehicle verifications: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// BulkCreateDriverVerification creates multiple driver verifications in one transaction
func (r *IndiaIntegrationRepositoryImpl) BulkCreateDriverVerification(ctx context.Context, verifications []*models.DriverVerification) error {
	if len(verifications) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Batch insert to avoid N+1 queries
	if len(verifications) > 0 {
		var args []interface{}
		var placeholders []string
		for i, v := range verifications {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*10+1, i*10+2, i*10+3, i*10+4, i*10+5, i*10+6, i*10+7, i*10+8, i*10+9, i*10+10))
			args = append(args,
				v.ID, v.DriverID, v.OrganizationID, v.LicenseNumber, v.Status, v.VerificationDate,
				v.Name, v.DateOfBirth, v.ValidityTo, v.CreatedAt,
			)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO driver_verifications (id, driver_id, org_id, license_number, status, verification_date, "+
				"name, date_of_birth, validity_to, created_at) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		_, err := tx.Exec(ctx, batchQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to batch insert driver verifications: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetExpiringVehicleVerifications fetches vehicle verifications expiring within X days
