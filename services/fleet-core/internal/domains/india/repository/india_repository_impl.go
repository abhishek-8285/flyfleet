package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/india/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/india/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IndiaIntegrationRepositoryImpl implements the IndiaIntegrationRepository interface
type IndiaIntegrationRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewIndiaIntegrationRepository creates a new instance of IndiaIntegrationRepositoryImpl
func NewIndiaIntegrationRepository(db *pgxpool.Pool) *IndiaIntegrationRepositoryImpl {
	return &IndiaIntegrationRepositoryImpl{
		db: db,
	}
}

// CreateFASTagTransaction creates a new FASTag transaction
func (r *IndiaIntegrationRepositoryImpl) CreateFASTagTransaction(ctx context.Context, transaction *models.FASTagTransaction) error {
	query := `
		INSERT INTO fastag_transactions (
			id, tag_number, vehicle_id, driver_id, organization_id, transaction_id, transaction_type,
			amount, toll_plaza, toll_plaza_id, lane_id, transaction_time, transaction_date,
			settlement_date, status, error_code, error_description, reconciled,
			reconciliation_date, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`

	_, err := r.db.Exec(ctx, query,
		transaction.ID,
		transaction.TagNumber,
		transaction.VehicleID,
		transaction.DriverID,
		transaction.OrganizationID,
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
		SELECT id, tag_number, vehicle_id, driver_id, organization_id, transaction_id, transaction_type,
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
		&transaction.OrganizationID,
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
		SELECT id, tag_number, vehicle_id, driver_id, organization_id, transaction_id, transaction_type,
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
			&transaction.OrganizationID,
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
func (r *IndiaIntegrationRepositoryImpl) CreateEWayBill(ctx context.Context, bill *models.EWayBill) error {
	query := `
		INSERT INTO e_way_bills (
			id, eway_bill_number, eway_bill_date, eway_bill_valid_till, document_number, document_date, 
			document_type, goods_description, goods_value, goods_tax_rate, goods_unit, goods_quantity,
			from_place, from_pincode, from_state_code, to_place, to_pincode, to_state_code,
			transporter_id, transporter_name, transport_mode, transport_document_number, 
			transport_document_date, transit_type, vehicle_number, vehicle_type, remaining_distance,
			user_gstin, recipient_gstin, generated_by, generated_at, cancelled, cancelled_by, 
			cancelled_on, cancelled_reason, created_by, updated_by, updated_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39)`

	_, err := r.db.Exec(ctx, query,
		bill.ID,
		bill.EWayBillNumber,
		bill.EWayBillDate,
		bill.EWayBillValidTill,
		bill.DocumentNumber,
		bill.DocumentDate,
		bill.DocumentType,
		bill.GoodsDescription,
		bill.GoodsValue,
		bill.GoodsTaxRate,
		bill.GoodsUnit,
		bill.GoodsQuantity,
		bill.FromPlace,
		bill.FromPincode,
		bill.FromStateCode,
		bill.ToPlace,
		bill.ToPincode,
		bill.ToStateCode,
		bill.TransporterID,
		bill.TransporterName,
		bill.TransportMode,
		bill.TransportDocumentNumber,
		bill.TransportDocumentDate,
		bill.TransitType,
		bill.VehicleNumber,
		bill.VehicleType,
		bill.RemainingDistance,
		bill.UserGSTIN,
		bill.RecipientGSTIN,
		bill.GeneratedBy,
		bill.GeneratedAt,
		bill.Cancelled,
		bill.CancelledBy,
		bill.CancelledOn,
		bill.CancelledReason,
		bill.CreatedBy,
		bill.UpdatedBy,
		bill.UpdatedAt,
		bill.CreatedAt,
	)

	return err
}

// GetEWayBill retrieves an E-Way Bill by ID
func (r *IndiaIntegrationRepositoryImpl) GetEWayBill(ctx context.Context, id string) (*models.EWayBill, error) {
	query := `
		SELECT id, eway_bill_number, eway_bill_date, eway_bill_valid_till, document_number, document_date,
		       document_type, goods_description, goods_value, goods_tax_rate, goods_unit, goods_quantity,
		       from_place, from_pincode, from_state_code, to_place, to_pincode, to_state_code,
		       transporter_id, transporter_name, transport_mode, transport_document_number,
		       transport_document_date, transit_type, vehicle_number, vehicle_type, remaining_distance,
		       user_gstin, recipient_gstin, generated_by, generated_at, cancelled, cancelled_by,
		       cancelled_on, cancelled_reason, created_by, updated_by, updated_at, created_at
		FROM e_way_bills
		WHERE id = $1`

	var bill models.EWayBill

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&bill.ID,
		&bill.EWayBillNumber,
		&bill.EWayBillDate,
		&bill.EWayBillValidTill,
		&bill.DocumentNumber,
		&bill.DocumentDate,
		&bill.DocumentType,
		&bill.GoodsDescription,
		&bill.GoodsValue,
		&bill.GoodsTaxRate,
		&bill.GoodsUnit,
		&bill.GoodsQuantity,
		&bill.FromPlace,
		&bill.FromPincode,
		&bill.FromStateCode,
		&bill.ToPlace,
		&bill.ToPincode,
		&bill.ToStateCode,
		&bill.TransporterID,
		&bill.TransporterName,
		&bill.TransportMode,
		&bill.TransportDocumentNumber,
		&bill.TransportDocumentDate,
		&bill.TransitType,
		&bill.VehicleNumber,
		&bill.VehicleType,
		&bill.RemainingDistance,
		&bill.UserGSTIN,
		&bill.RecipientGSTIN,
		&bill.GeneratedBy,
		&bill.GeneratedAt,
		&bill.Cancelled,
		&bill.CancelledBy,
		&bill.CancelledOn,
		&bill.CancelledReason,
		&bill.CreatedBy,
		&bill.UpdatedBy,
		&bill.UpdatedAt,
		&bill.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get e-way bill: %w", err)
	}

	return &bill, nil
}

// GetEWayBillByNumber retrieves an E-Way Bill by its number
func (r *IndiaIntegrationRepositoryImpl) GetEWayBillByNumber(ctx context.Context, number string) (*models.EWayBill, error) {
	query := `
		SELECT id, eway_bill_number, eway_bill_date, eway_bill_valid_till, document_number, document_date,
		       document_type, goods_description, goods_value, goods_tax_rate, goods_unit, goods_quantity,
		       from_place, from_pincode, from_state_code, to_place, to_pincode, to_state_code,
		       transporter_id, transporter_name, transport_mode, transport_document_number,
		       transport_document_date, transit_type, vehicle_number, vehicle_type, remaining_distance,
		       user_gstin, recipient_gstin, generated_by, generated_at, cancelled, cancelled_by,
		       cancelled_on, cancelled_reason, created_by, updated_by, updated_at, created_at
		FROM e_way_bills
		WHERE eway_bill_number = $1`

	var bill models.EWayBill

	row := r.db.QueryRow(ctx, query, number)
	err := row.Scan(
		&bill.ID,
		&bill.EWayBillNumber,
		&bill.EWayBillDate,
		&bill.EWayBillValidTill,
		&bill.DocumentNumber,
		&bill.DocumentDate,
		&bill.DocumentType,
		&bill.GoodsDescription,
		&bill.GoodsValue,
		&bill.GoodsTaxRate,
		&bill.GoodsUnit,
		&bill.GoodsQuantity,
		&bill.FromPlace,
		&bill.FromPincode,
		&bill.FromStateCode,
		&bill.ToPlace,
		&bill.ToPincode,
		&bill.ToStateCode,
		&bill.TransporterID,
		&bill.TransporterName,
		&bill.TransportMode,
		&bill.TransportDocumentNumber,
		&bill.TransportDocumentDate,
		&bill.TransitType,
		&bill.VehicleNumber,
		&bill.VehicleType,
		&bill.RemainingDistance,
		&bill.UserGSTIN,
		&bill.RecipientGSTIN,
		&bill.GeneratedBy,
		&bill.GeneratedAt,
		&bill.Cancelled,
		&bill.CancelledBy,
		&bill.CancelledOn,
		&bill.CancelledReason,
		&bill.CreatedBy,
		&bill.UpdatedBy,
		&bill.UpdatedAt,
		&bill.CreatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("e-way bill not found")
		}
		return nil, fmt.Errorf("failed to get e-way bill by number: %w", err)
	}

	return &bill, nil
}

// UpdateEWayBill updates an existing E-Way Bill
func (r *IndiaIntegrationRepositoryImpl) UpdateEWayBill(ctx context.Context, id string, updates map[string]interface{}) error {
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

	query := fmt.Sprintf("UPDATE e_way_bills SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteEWayBill deletes an E-Way Bill
func (r *IndiaIntegrationRepositoryImpl) DeleteEWayBill(ctx context.Context, id string) error {
	query := "DELETE FROM e_way_bills WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete e-way bill: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("e-way bill not found")
	}

	return nil
}

// ListEWayBills retrieves E-Way Bills with filtering
func (r *IndiaIntegrationRepositoryImpl) ListEWayBills(ctx context.Context, vehicleNumber *string, transporterID *string, startDate, endDate *time.Time, limit, offset int) ([]*models.EWayBill, int, error) {
	countQuery := "SELECT COUNT(*) FROM e_way_bills WHERE true"
	args := []interface{}{}
	argIndex := 1

	whereClause := "WHERE true"
	if vehicleNumber != nil {
		whereClause += fmt.Sprintf(" AND vehicle_number = $%d", argIndex)
		args = append(args, vehicleNumber)
		argIndex++
	}
	if transporterID != nil {
		whereClause += fmt.Sprintf(" AND transporter_id = $%d", argIndex)
		args = append(args, transporterID)
		argIndex++
	}
	if startDate != nil {
		whereClause += fmt.Sprintf(" AND generated_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}
	if endDate != nil {
		whereClause += fmt.Sprintf(" AND generated_at <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	var totalCount int
	err := r.db.QueryRow(ctx, fmt.Sprintf("%s %s", countQuery, whereClause), args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count e-way bills: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, eway_bill_number, eway_bill_date, eway_bill_valid_till, document_number, document_date,
		       document_type, goods_description, goods_value, goods_tax_rate, goods_unit, goods_quantity,
		       from_place, from_pincode, from_state_code, to_place, to_pincode, to_state_code,
		       transporter_id, transporter_name, transport_mode, transport_document_number,
		       transport_document_date, transit_type, vehicle_number, vehicle_type, remaining_distance,
		       user_gstin, recipient_gstin, generated_by, generated_at, cancelled, cancelled_by,
		       cancelled_on, cancelled_reason, created_by, updated_by, updated_at, created_at
		FROM e_way_bills
		%s
		ORDER BY generated_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list e-way bills: %w", err)
	}
	defer rows.Close()

	var bills []*models.EWayBill
	for rows.Next() {
		var bill models.EWayBill

		err := rows.Scan(
			&bill.ID,
			&bill.EWayBillNumber,
			&bill.EWayBillDate,
			&bill.EWayBillValidTill,
			&bill.DocumentNumber,
			&bill.DocumentDate,
			&bill.DocumentType,
			&bill.GoodsDescription,
			&bill.GoodsValue,
			&bill.GoodsTaxRate,
			&bill.GoodsUnit,
			&bill.GoodsQuantity,
			&bill.FromPlace,
			&bill.FromPincode,
			&bill.FromStateCode,
			&bill.ToPlace,
			&bill.ToPincode,
			&bill.ToStateCode,
			&bill.TransporterID,
			&bill.TransporterName,
			&bill.TransportMode,
			&bill.TransportDocumentNumber,
			&bill.TransportDocumentDate,
			&bill.TransitType,
			&bill.VehicleNumber,
			&bill.VehicleType,
			&bill.RemainingDistance,
			&bill.UserGSTIN,
			&bill.RecipientGSTIN,
			&bill.GeneratedBy,
			&bill.GeneratedAt,
			&bill.Cancelled,
			&bill.CancelledBy,
			&bill.CancelledOn,
			&bill.CancelledReason,
			&bill.CreatedBy,
			&bill.UpdatedBy,
			&bill.UpdatedAt,
			&bill.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan e-way bill: %w", err)
		}

		bills = append(bills, &bill)
	}

	return bills, totalCount, nil
}

// CreateVehicleVerification creates a new vehicle verification record
func (r *IndiaIntegrationRepositoryImpl) CreateVehicleVerification(ctx context.Context, verification *models.VehicleVerification) error {
	query := `
		INSERT INTO vehicle_verifications (
			id, vehicle_id, organization_id, registration_number, status, verification_date,
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
		SELECT id, vehicle_id, organization_id, registration_number, status, verification_date,
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
		SELECT id, vehicle_id, organization_id, registration_number, status, verification_date,
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
		SELECT id, vehicle_id, organization_id, registration_number, status, verification_date,
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
func (r *IndiaIntegrationRepositoryImpl) CreateDriverVerification(ctx context.Context, verification *models.DriverVerification) error {
	query := `
		INSERT INTO driver_verifications (
			id, driver_id, organization_id, license_number, status, verification_date,
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
		SELECT id, driver_id, organization_id, license_number, status, verification_date,
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
		SELECT id, driver_id, organization_id, license_number, status, verification_date,
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
		SELECT id, driver_id, organization_id, license_number, status, verification_date,
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
func (r *IndiaIntegrationRepositoryImpl) CreateLocalPermit(ctx context.Context, permit *models.LocalPermit) error {
	query := `
		INSERT INTO local_permits (
			id, vehicle_id, driver_id, organization_id, permit_number, permit_type, permit_category,
			permit_sub_category, valid_from, valid_to, issuing_authority, jurisdiction, route_details,
			number_of_trips, max_load_capacity, number_of_seats, fare, status, remarks,
			created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`

	_, err := r.db.Exec(ctx, query,
		permit.ID,
		permit.VehicleID,
		permit.DriverID,
		permit.OrganizationID,
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
		SELECT id, vehicle_id, driver_id, organization_id, permit_number, permit_type, permit_category,
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
		&permit.OrganizationID,
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
		SELECT id, vehicle_id, driver_id, organization_id, permit_number, permit_type, permit_category,
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
		&permit.OrganizationID,
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
		SELECT id, vehicle_id, driver_id, organization_id, permit_number, permit_type, permit_category,
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
			&permit.OrganizationID,
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
func (r *IndiaIntegrationRepositoryImpl) CreateIntegrationConfig(ctx context.Context, config *models.IntegrationConfig) error {
	query := `
		INSERT INTO integration_configs (
			id, organization_id, service_type, api_endpoint, api_key, username, password,
			client_id, client_secret, environment, active, last_sync_at, next_sync_at,
			sync_interval, created_at, created_by, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	// Encrypt sensitive fields
	apiKey, _ := utils.Encrypt(config.APIKey)
	username, _ := utils.Encrypt(config.Username)
	password, _ := utils.Encrypt(config.Password)
	clientSecret, _ := utils.Encrypt(config.ClientSecret)

	_, err := r.db.Exec(ctx, query,
		config.ID,
		config.OrganizationID,
		config.ServiceType,
		config.APIEndpoint,
		apiKey,
		username,
		password,
		config.ClientID,
		clientSecret,
		config.Environment,
		config.Active,
		config.LastSyncAt,
		config.NextSyncAt,
		config.SyncInterval,
		config.CreatedAt,
		config.CreatedBy,
		config.UpdatedAt,
		config.UpdatedBy,
	)

	return err
}

// GetIntegrationConfig retrieves an integration configuration by ID
func (r *IndiaIntegrationRepositoryImpl) GetIntegrationConfig(ctx context.Context, id string) (*models.IntegrationConfig, error) {
	query := `
		SELECT id, organization_id, service_type, api_endpoint, api_key, username, password,
		       client_id, client_secret, environment, active, last_sync_at, next_sync_at,
		       sync_interval, created_at, created_by, updated_at, updated_by
		FROM integration_configs
		WHERE id = $1`

	var config models.IntegrationConfig

	row := r.db.QueryRow(ctx, query, id)
	err := row.Scan(
		&config.ID,
		&config.OrganizationID,
		&config.ServiceType,
		&config.APIEndpoint,
		&config.APIKey,
		&config.Username,
		&config.Password,
		&config.ClientID,
		&config.ClientSecret,
		&config.Environment,
		&config.Active,
		&config.LastSyncAt,
		&config.NextSyncAt,
		&config.SyncInterval,
		&config.CreatedAt,
		&config.CreatedBy,
		&config.UpdatedAt,
		&config.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get integration config: %w", err)
	}

	// Decrypt sensitive fields
	config.APIKey, _ = utils.Decrypt(config.APIKey)
	config.Username, _ = utils.Decrypt(config.Username)
	config.Password, _ = utils.Decrypt(config.Password)
	config.ClientSecret, _ = utils.Decrypt(config.ClientSecret)

	return &config, nil
}

// GetIntegrationConfigByServiceType retrieves an integration configuration by service type and organization
func (r *IndiaIntegrationRepositoryImpl) GetIntegrationConfigByServiceType(ctx context.Context, orgID uuid.UUID, serviceType string) (*models.IntegrationConfig, error) {
	query := `
		SELECT id, organization_id, service_type, api_endpoint, api_key, username, password,
		       client_id, client_secret, environment, active, last_sync_at, next_sync_at,
		       sync_interval, created_at, created_by, updated_at, updated_by
		FROM integration_configs
		WHERE organization_id = $1 AND service_type = $2`

	var config models.IntegrationConfig

	row := r.db.QueryRow(ctx, query, orgID, serviceType)
	err := row.Scan(
		&config.ID,
		&config.OrganizationID,
		&config.ServiceType,
		&config.APIEndpoint,
		&config.APIKey,
		&config.Username,
		&config.Password,
		&config.ClientID,
		&config.ClientSecret,
		&config.Environment,
		&config.Active,
		&config.LastSyncAt,
		&config.NextSyncAt,
		&config.SyncInterval,
		&config.CreatedAt,
		&config.CreatedBy,
		&config.UpdatedAt,
		&config.UpdatedBy,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.New("integration config not found")
		}
		return nil, fmt.Errorf("failed to get integration config by service type: %w", err)
	}

	// Decrypt sensitive fields
	config.APIKey, _ = utils.Decrypt(config.APIKey)
	config.Username, _ = utils.Decrypt(config.Username)
	config.Password, _ = utils.Decrypt(config.Password)
	config.ClientSecret, _ = utils.Decrypt(config.ClientSecret)

	return &config, nil
}

// UpdateIntegrationConfig updates an existing integration configuration
func (r *IndiaIntegrationRepositoryImpl) UpdateIntegrationConfig(ctx context.Context, id string, updates map[string]interface{}) error {
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

		// Encrypt sensitive fields if present in updates
		if key == "api_key" || key == "username" || key == "password" || key == "client_secret" {
			if strVal, ok := value.(string); ok {
				encryptedVal, _ := utils.Encrypt(strVal)
				value = encryptedVal
			}
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

	query := fmt.Sprintf("UPDATE integration_configs SET %s WHERE id = $%d", setClause, argIndex+1)
	args = append(args, id)

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// DeleteIntegrationConfig deletes an integration configuration
func (r *IndiaIntegrationRepositoryImpl) DeleteIntegrationConfig(ctx context.Context, id string) error {
	query := "DELETE FROM integration_configs WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete integration config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("integration config not found")
	}

	return nil
}

// ListIntegrationConfigs retrieves integration configurations with filtering
func (r *IndiaIntegrationRepositoryImpl) ListIntegrationConfigs(ctx context.Context, orgID uuid.UUID, serviceType *string, active *bool, limit, offset int) ([]*models.IntegrationConfig, int, error) {
	countQuery := "SELECT COUNT(*) FROM integration_configs WHERE organization_id = $1"
	args := []interface{}{orgID}
	argIndex := 2

	whereClause := "WHERE organization_id = $1"
	if serviceType != nil {
		whereClause += fmt.Sprintf(" AND service_type = $%d", argIndex)
		args = append(args, serviceType)
		argIndex++
	}
	if active != nil {
		whereClause += fmt.Sprintf(" AND active = $%d", argIndex)
		args = append(args, active)
		argIndex++
	}

	var totalCount int
	err := r.db.QueryRow(ctx, fmt.Sprintf("%s %s", countQuery, whereClause), args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count integration configs: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, organization_id, service_type, api_endpoint, api_key, username, password,
		       client_id, client_secret, environment, active, last_sync_at, next_sync_at,
		       sync_interval, created_at, created_by, updated_at, updated_by
		FROM integration_configs
		%s
		ORDER BY service_type, created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list integration configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.IntegrationConfig
	for rows.Next() {
		var config models.IntegrationConfig

		err := rows.Scan(
			&config.ID,
			&config.OrganizationID,
			&config.ServiceType,
			&config.APIEndpoint,
			&config.APIKey,
			&config.Username,
			&config.Password,
			&config.ClientID,
			&config.ClientSecret,
			&config.Environment,
			&config.Active,
			&config.LastSyncAt,
			&config.NextSyncAt,
			&config.SyncInterval,
			&config.CreatedAt,
			&config.CreatedBy,
			&config.UpdatedAt,
			&config.UpdatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan integration config: %w", err)
		}

		// Decrypt sensitive fields
		config.APIKey, _ = utils.Decrypt(config.APIKey)
		config.Username, _ = utils.Decrypt(config.Username)
		config.Password, _ = utils.Decrypt(config.Password)
		config.ClientSecret, _ = utils.Decrypt(config.ClientSecret)

		configs = append(configs, &config)
	}

	return configs, totalCount, nil
}

// HealthCheck performs a health check on the database
func (r *IndiaIntegrationRepositoryImpl) HealthCheck(ctx context.Context) error {
	err := r.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

// Ping checks if the database connection is alive
func (r *IndiaIntegrationRepositoryImpl) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}
// CreateIntegrationLog creates a new integration audit log
func (r *IndiaIntegrationRepositoryImpl) CreateIntegrationLog(ctx context.Context, log *models.IntegrationLog) error {
	query := `
		INSERT INTO india_integration_logs (
			id, service_type, request_payload, response_payload, status, error_message, organization_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	reqJSON, _ := json.Marshal(log.RequestPayload)
	resJSON, _ := json.Marshal(log.ResponsePayload)

	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(ctx, query,
		log.ID,
		log.ServiceType,
		reqJSON,
		resJSON,
		log.Status,
		log.ErrorMessage,
		log.OrganizationID,
		log.CreatedAt,
	)

	return err
}
// BulkCreateVehicleVerification creates multiple vehicle verifications in one transaction
func (r *IndiaIntegrationRepositoryImpl) BulkCreateVehicleVerification(ctx context.Context, verifications []*models.VehicleVerification) error {
	if len(verifications) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO vehicle_verifications (
			id, vehicle_id, organization_id, registration_number, status, verification_date,
			chassis_number, engine_number, model, manufacturer, fuel_type, body_type,
			cubic_capacity, seating_capacity, color, manufacturing_year, owner_name,
			rc_status, insurance_expiry, puc_expiry, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	for _, v := range verifications {
		_, err := tx.Exec(ctx, query,
			v.ID, v.VehicleID, v.OrganizationID, v.RegistrationNumber, v.Status, v.VerificationDate,
			v.ChassisNumber, v.EngineNumber, v.Model, v.Manufacturer, v.FuelType, v.BodyType,
			v.CubicCapacity, v.SeatingCapacity, v.Color, v.ManufacturingYear, v.OwnerName,
			v.RCStatus, v.InsuranceExpiry, v.PUCExpiry, v.CreatedAt,
		)
		if err != nil {
			return err
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

	query := `
		INSERT INTO driver_verifications (
			id, driver_id, organization_id, license_number, status, verification_date,
			name, date_of_birth, validity_to, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	for _, v := range verifications {
		_, err := tx.Exec(ctx, query,
			v.ID, v.DriverID, v.OrganizationID, v.LicenseNumber, v.Status, v.VerificationDate,
			v.Name, v.DateOfBirth, v.ValidityTo, v.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetExpiringVehicleVerifications fetches vehicle verifications expiring within X days
func (r *IndiaIntegrationRepositoryImpl) GetExpiringVehicleVerifications(ctx context.Context, days int) ([]*models.VehicleVerification, error) {
	query := `
		SELECT id, vehicle_id, organization_id, registration_number, status, verification_date,
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
		SELECT id, driver_id, organization_id, license_number, status, verification_date,
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
		SELECT id, vehicle_id, driver_id, organization_id, permit_number, permit_type, permit_category,
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
		SELECT id, organization_id, gstin, status, legal_name, trade_name, address,
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
