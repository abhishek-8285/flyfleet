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
