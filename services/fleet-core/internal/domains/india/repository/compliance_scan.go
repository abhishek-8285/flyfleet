package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/india/models"
)

func (r *IndiaIntegrationRepositoryImpl) GetExpiringVehicleVerifications(ctx context.Context, days int) ([]*models.VehicleVerification, error) {
	query := `
		SELECT id, vehicle_id, org_id, registration_number, status, verification_date,
		       chassis_number, engine_number, model, manufacturer, fuel_type, body_type,
		       cubic_capacity, seating_capacity, color, manufacturing_year, owner_name,
		       rc_status, insurance_expiry, puc_expiry, created_at
		FROM vehicle_verifications
		WHERE (insurance_expiry <= $1 OR puc_expiry <= $1) AND status = 'verified'`

	threshold := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	rows, err := r.db.Query(ctx, query, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.VehicleVerification
	for rows.Next() {
		var v models.VehicleVerification
		err := rows.Scan(
			&v.ID, &v.VehicleID, &v.OrganizationID, &v.RegistrationNumber, &v.Status, &v.VerificationDate,
			&v.ChassisNumber, &v.EngineNumber, &v.Model, &v.Manufacturer, &v.FuelType, &v.BodyType,
			&v.CubicCapacity, &v.SeatingCapacity, &v.Color, &v.ManufacturingYear, &v.OwnerName,
			&v.RCStatus, &v.InsuranceExpiry, &v.PUCExpiry, &v.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, &v)
	}
	return results, nil
}

// GetExpiringDriverVerifications fetches driver verifications expiring within X days
func (r *IndiaIntegrationRepositoryImpl) GetExpiringDriverVerifications(ctx context.Context, days int) ([]*models.DriverVerification, error) {
	query := `
		SELECT id, driver_id, org_id, license_number, status, verification_date,
		       name, date_of_birth, validity_to, created_at
		FROM driver_verifications
		WHERE validity_to <= $1 AND status = 'verified'`

	threshold := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	rows, err := r.db.Query(ctx, query, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.DriverVerification
	for rows.Next() {
		var d models.DriverVerification
		err := rows.Scan(
			&d.ID, &d.DriverID, &d.OrganizationID, &d.LicenseNumber, &d.Status, &d.VerificationDate,
			&d.Name, &d.DateOfBirth, &d.ValidityTo, &d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, &d)
	}
	return results, nil
}

// GetExpiringLocalPermits fetches local permits expiring within X days
func (r *IndiaIntegrationRepositoryImpl) GetExpiringLocalPermits(ctx context.Context, days int) ([]*models.LocalPermit, error) {
	query := `
		SELECT id, vehicle_id, driver_id, org_id, permit_number, permit_type, permit_category,
		       permit_sub_category, valid_from, valid_to, issuing_authority, jurisdiction, route_details,
		       number_of_trips, max_load_capacity, number_of_seats, fare, status, remarks,
		       created_at
		FROM local_permits
		WHERE valid_to <= $1 AND status = 'active'`

	threshold := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	rows, err := r.db.Query(ctx, query, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.LocalPermit
	for rows.Next() {
		var p models.LocalPermit
		err := rows.Scan(
			&p.ID, &p.VehicleID, &p.DriverID, &p.OrganizationID, &p.PermitNumber, &p.PermitType, &p.PermitCategory,
			&p.PermitSubCategory, &p.ValidFrom, &p.ValidTo, &p.IssuingAuthority, &p.Jurisdiction, &p.RouteDetails,
			&p.NumberOfTrips, &p.MaxLoadCapacity, &p.NumberOfSeats, &p.Fare, &p.Status, &p.Remarks,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, &p)
	}
	return results, nil
}

// GetNonCompliantGSTRecords fetches non-compliant GST records
func (r *IndiaIntegrationRepositoryImpl) GetNonCompliantGSTRecords(ctx context.Context) ([]*models.GSTCompliance, error) {
	query := `
		SELECT id, org_id, gstin, status, legal_name, trade_name, address,
		       centre_jurisdiction, state_jurisdiction, date_of_registration, date_of_cancellation,
		       last_updated, return_filing_frequency, return_filing_status, payment_status,
		       created_at
		FROM gst_compliances
		WHERE status IN ('cancelled', 'suspended')`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.GSTCompliance
	for rows.Next() {
		var g models.GSTCompliance
		err := rows.Scan(
			&g.ID, &g.OrganizationID, &g.GSTIN, &g.Status, &g.LegalName, &g.TradeName, &g.Address,
			&g.CentreJurisdiction, &g.StateJurisdiction, &g.DateOfRegistration, &g.DateOfCancellation,
			&g.LastUpdated, &g.ReturnFilingFrequency, &g.ReturnFilingStatus, &g.PaymentStatus,
			&g.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, &g)
	}
	return results, nil
}
