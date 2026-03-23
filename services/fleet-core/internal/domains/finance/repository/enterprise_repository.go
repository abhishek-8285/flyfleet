package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/finance/models"
	fuelModels "github.com/flyfleet/fleet-core-service/internal/domains/fuel/models"
	"github.com/google/uuid"
)

// =============================================================================
// GST E-Invoices
// =============================================================================

func (r *FinanceRepository) CreateGSTInvoice(ctx context.Context, inv *models.GSTInvoice) error {
	query := `
		INSERT INTO gst_invoices (
			id, invoice_number, irn, ack_number, ack_date, qr_code, signed_invoice,
			status, entity_type, entity_id, org_id, gstin, buyer_gstin,
			total_amount, taxable_amount, cgst_amount, sgst_amount, igst_amount,
			error_message, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		inv.ID, inv.InvoiceNumber, inv.IRN, inv.AckNumber, inv.AckDate,
		inv.QRCode, inv.SignedInvoice, string(inv.Status), inv.EntityType, inv.EntityID,
		inv.OrgID, inv.GSTIN, inv.BuyerGSTIN, inv.TotalAmount, inv.TaxableAmount,
		inv.CGSTAmount, inv.SGSTAmount, inv.IGSTAmount, inv.ErrorMessage, now, now,
	)
	return err
}

func (r *FinanceRepository) GetGSTInvoice(ctx context.Context, id uuid.UUID) (*models.GSTInvoice, error) {
	query := `SELECT id, invoice_number, irn, ack_number, ack_date, qr_code, signed_invoice,
	                 status, entity_type, entity_id, org_id, gstin, buyer_gstin,
	                 total_amount, taxable_amount, cgst_amount, sgst_amount, igst_amount,
	                 error_message, created_at, updated_at
	          FROM gst_invoices WHERE id = $1`

	var inv models.GSTInvoice
	var status string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&inv.ID, &inv.InvoiceNumber, &inv.IRN, &inv.AckNumber, &inv.AckDate,
		&inv.QRCode, &inv.SignedInvoice, &status, &inv.EntityType, &inv.EntityID,
		&inv.OrgID, &inv.GSTIN, &inv.BuyerGSTIN, &inv.TotalAmount, &inv.TaxableAmount,
		&inv.CGSTAmount, &inv.SGSTAmount, &inv.IGSTAmount, &inv.ErrorMessage,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	inv.Status = models.GSTInvoiceStatus(status)
	return &inv, nil
}

// =============================================================================
// TDS Records
// =============================================================================

func (r *FinanceRepository) CreateTDSRecord(ctx context.Context, rec *models.TDSRecord) error {
	query := `
		INSERT INTO tds_records (
			id, entity_type, entity_id, vendor_id, org_id, pan_number, tds_section,
			gross_amount, tds_rate, tds_amount, net_amount, deducted_at,
			deposited, financial_year, quarter, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`

	_, err := r.db.Exec(ctx, query,
		rec.ID, rec.EntityType, rec.EntityID, rec.VendorID, rec.OrgID, rec.PANNumber, rec.TDSSection,
		rec.GrossAmount, rec.TDSRate, rec.TDSAmount, rec.NetAmount, rec.DeductedAt,
		rec.Deposited, rec.FinancialYear, rec.Quarter, time.Now(),
	)
	return err
}

func (r *FinanceRepository) GetTDSRecords(ctx context.Context, orgID uuid.UUID, fy, quarter string) ([]models.TDSRecord, error) {
	query := `SELECT id, entity_type, entity_id, vendor_id, org_id, pan_number, tds_section,
	                 gross_amount, tds_rate, tds_amount, net_amount, deducted_at,
	                 deposited, financial_year, quarter, created_at
	          FROM tds_records WHERE org_id = $1`
	args := []interface{}{orgID}

	if fy != "" {
		query += fmt.Sprintf(" AND financial_year = $%d", len(args)+1)
		args = append(args, fy)
	}
	if quarter != "" {
		query += fmt.Sprintf(" AND quarter = $%d", len(args)+1)
		args = append(args, quarter)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.TDSRecord
	for rows.Next() {
		var rec models.TDSRecord
		if err := rows.Scan(
			&rec.ID, &rec.EntityType, &rec.EntityID, &rec.VendorID, &rec.OrgID, &rec.PANNumber, &rec.TDSSection,
			&rec.GrossAmount, &rec.TDSRate, &rec.TDSAmount, &rec.NetAmount, &rec.DeductedAt,
			&rec.Deposited, &rec.FinancialYear, &rec.Quarter, &rec.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}

// =============================================================================
// Vendor Payments
// =============================================================================

func (r *FinanceRepository) CreateVendorPayment(ctx context.Context, p *models.VendorPayment) error {
	query := `
		INSERT INTO vendor_payments (
			id, vendor_id, org_id, invoice_number, amount, tds_deducted, net_payable,
			currency, payment_status, payment_mode, payment_reference, bank_account,
			ifsc_code, payment_date, paid_at, gst_invoice_id, tds_record_id,
			notes, created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		p.ID, p.VendorID, p.OrgID, p.InvoiceNumber, p.Amount, p.TDSDeducted, p.NetPayable,
		p.Currency, string(p.PaymentStatus), p.PaymentMode, p.PaymentReference,
		p.BankAccount, p.IFSCCode, p.PaymentDate, p.PaidAt,
		p.GSTInvoiceID, p.TDSRecordID, p.Notes, p.CreatedBy, now, now,
	)
	return err
}

func (r *FinanceRepository) ListVendorPayments(ctx context.Context, orgID uuid.UUID, vendorID *uuid.UUID) ([]models.VendorPayment, error) {
	query := `SELECT id, vendor_id, org_id, invoice_number, amount, tds_deducted, net_payable,
	                 currency, payment_status, payment_mode, payment_reference, bank_account,
	                 ifsc_code, payment_date, paid_at, gst_invoice_id, tds_record_id,
	                 notes, created_by, created_at, updated_at
	          FROM vendor_payments WHERE org_id = $1`
	args := []interface{}{orgID}

	if vendorID != nil {
		query += " AND vendor_id = $2"
		args = append(args, *vendorID)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.VendorPayment
	for rows.Next() {
		var p models.VendorPayment
		var status string
		if err := rows.Scan(
			&p.ID, &p.VendorID, &p.OrgID, &p.InvoiceNumber, &p.Amount, &p.TDSDeducted, &p.NetPayable,
			&p.Currency, &status, &p.PaymentMode, &p.PaymentReference, &p.BankAccount,
			&p.IFSCCode, &p.PaymentDate, &p.PaidAt, &p.GSTInvoiceID, &p.TDSRecordID,
			&p.Notes, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		p.PaymentStatus = models.VendorPaymentStatus(status)
		payments = append(payments, p)
	}
	return payments, nil
}

// =============================================================================
// Expense Items
// =============================================================================

func (r *FinanceRepository) CreateExpenseItem(ctx context.Context, item *models.ExpenseItem) error {
	query := `
		INSERT INTO expense_items (
			id, org_id, report_id, invoice_number, category, amount, gst_amount, gst_rate,
			receipt_url, description, vendor_name, vendor_gstin,
			transaction_date, ocr_processed, ocr_confidence, created_at
		) VALUES (
			$1,
			COALESCE((SELECT org_id::uuid FROM expense_reports WHERE id = $2), '00000000-0000-0000-0000-000000000000'::uuid),
			$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
		)`

	_, err := r.db.Exec(ctx, query,
		item.ID, item.ReportID, item.InvoiceNumber, string(item.Category), item.Amount, item.GSTAmount, item.GSTRate,
		item.ReceiptURL, item.Description, item.VendorName, item.VendorGSTIN,
		item.TransactionDate, item.OCRProcessed, item.OCRConfidence, time.Now(),
	)
	return err
}

func (r *FinanceRepository) ListExpenseItems(ctx context.Context, reportID uuid.UUID) ([]models.ExpenseItem, error) {
	query := `SELECT id, report_id, invoice_number, category, amount, gst_amount, gst_rate,
	                 receipt_url, description, vendor_name, vendor_gstin,
	                 transaction_date, ocr_processed, ocr_confidence, created_at
	          FROM expense_items WHERE report_id = $1 ORDER BY created_at`

	rows, err := r.db.Query(ctx, query, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ExpenseItem
	for rows.Next() {
		var item models.ExpenseItem
		var category string
		if err := rows.Scan(
			&item.ID, &item.ReportID, &item.InvoiceNumber, &category, &item.Amount, &item.GSTAmount, &item.GSTRate,
			&item.ReceiptURL, &item.Description, &item.VendorName, &item.VendorGSTIN,
			&item.TransactionDate, &item.OCRProcessed, &item.OCRConfidence, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.Category = models.ExpenseItemCategory(category)
		items = append(items, item)
	}
	return items, nil
}

func (r *FinanceRepository) MarkExpenseItemOCRProcessed(ctx context.Context, id uuid.UUID, confidence float64) error {
	query := `UPDATE expense_items SET ocr_processed = TRUE, ocr_confidence = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, confidence, id)
	return err
}

// =============================================================================
// Fuel Card Reconciliation
// =============================================================================

func (r *FinanceRepository) CreateFuelCardReconciliation(ctx context.Context, rec *models.FuelCardReconciliation) error {
	query := `
		INSERT INTO fuel_card_reconciliation (
			id, card_transaction_id, internal_transaction_id, org_id,
			status, variance, card_amount, internal_amount,
			reconciled_by, reconciled_at, notes, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`

	_, err := r.db.Exec(ctx, query,
		rec.ID, rec.CardTransactionID, rec.InternalTransactionID, rec.OrgID,
		string(rec.Status), rec.Variance, rec.CardAmount, rec.InternalAmount,
		rec.ReconciledBy, rec.ReconciledAt, rec.Notes, time.Now(),
	)
	return err
}

func (r *FinanceRepository) UpdateReconciliationStatus(ctx context.Context, reconID, internalTxID, userID uuid.UUID, status models.ReconciliationStatus) error {
	query := `UPDATE fuel_card_reconciliation
	          SET status = $1, internal_transaction_id = $2, reconciled_by = $3, reconciled_at = NOW()
	          WHERE id = $4`
	_, err := r.db.Exec(ctx, query, string(status), internalTxID, userID, reconID)
	return err
}

// =============================================================================
// Fraud Incidents
// =============================================================================

func (r *FinanceRepository) CreateFraudIncident(ctx context.Context, inc *models.FraudIncident) error {
	query := `
		INSERT INTO fraud_incidents (
			id, entity_type, entity_id, org_id, vehicle_id, driver_id,
			fraud_score, risk_level, signals, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	_, err := r.db.Exec(ctx, query,
		inc.ID, inc.EntityType, inc.EntityID, inc.OrgID, inc.VehicleID, inc.DriverID,
		inc.FraudScore, string(inc.RiskLevel), inc.Signals, inc.Status, time.Now(),
	)
	return err
}

// GetRecentFuelLogsForFraud returns fuel logs updated since the given time.
func (r *FinanceRepository) GetRecentFuelLogsForFraud(ctx context.Context, since time.Time) ([]fuelModels.FuelLog, error) {
	query := `SELECT id, organization_id, vehicle_id, driver_id, station_name, fuel_type,
	                 quantity, unit_price, total_cost, odometer, transaction_date,
	                 status, gst_number, gst_rate, gst_amount, invoice_number, vendor_gstin,
	                 receipt_url, notes, anomaly_flag, anomaly_reason, created_by, created_at, updated_at
	          FROM fuel_logs WHERE updated_at >= $1 ORDER BY updated_at DESC LIMIT 200`

	rows, err := r.db.Query(ctx, query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []fuelModels.FuelLog
	for rows.Next() {
		var log fuelModels.FuelLog
		if err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.VehicleID, &log.DriverID, &log.StationName, &log.FuelType,
			&log.Quantity, &log.UnitPrice, &log.TotalCost, &log.Odometer, &log.TransactionDate,
			&log.Status, &log.GSTNumber, &log.GSTRate, &log.GSTAmount, &log.InvoiceNumber, &log.VendorGSTIN,
			&log.ReceiptURL, &log.Notes, &log.AnomalyFlag, &log.AnomalyReason, &log.CreatedBy, &log.CreatedAt, &log.UpdatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// =============================================================================
// Trip Profitability
// =============================================================================

func (r *FinanceRepository) UpsertTripProfitability(ctx context.Context, tp *models.TripProfitability) error {
	query := `
		INSERT INTO trip_profitability (
			id, trip_id, org_id, vehicle_id, driver_id, trip_revenue,
			fuel_cost, toll_cost, driver_cost, maintenance_cost, other_cost,
			profit_margin, distance_km, cost_per_km, calculated_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		ON CONFLICT (trip_id) DO UPDATE SET
			trip_revenue = EXCLUDED.trip_revenue, fuel_cost = EXCLUDED.fuel_cost,
			toll_cost = EXCLUDED.toll_cost, driver_cost = EXCLUDED.driver_cost,
			profit_margin = EXCLUDED.profit_margin, distance_km = EXCLUDED.distance_km,
			cost_per_km = EXCLUDED.cost_per_km, calculated_at = EXCLUDED.calculated_at,
			updated_at = EXCLUDED.updated_at`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		tp.ID, tp.TripID, tp.OrgID, tp.VehicleID, tp.DriverID, tp.TripRevenue,
		tp.FuelCost, tp.TollCost, tp.DriverCost, tp.MaintenanceCost, tp.OtherCost,
		tp.ProfitMargin, tp.DistanceKM, tp.CostPerKM, tp.CalculatedAt, now, now,
	)
	return err
}

func (r *FinanceRepository) GetTripProfitability(ctx context.Context, tripID uuid.UUID) (*models.TripProfitability, error) {
	query := `SELECT id, trip_id, org_id, vehicle_id, driver_id, trip_revenue,
	                 fuel_cost, toll_cost, driver_cost, maintenance_cost, other_cost,
	                 total_cost, profit, profit_margin, distance_km, cost_per_km, calculated_at
	          FROM trip_profitability WHERE trip_id = $1`

	var tp models.TripProfitability
	err := r.db.QueryRow(ctx, query, tripID).Scan(
		&tp.ID, &tp.TripID, &tp.OrgID, &tp.VehicleID, &tp.DriverID, &tp.TripRevenue,
		&tp.FuelCost, &tp.TollCost, &tp.DriverCost, &tp.MaintenanceCost, &tp.OtherCost,
		&tp.TotalCost, &tp.Profit, &tp.ProfitMargin, &tp.DistanceKM, &tp.CostPerKM, &tp.CalculatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &tp, nil
}

// GetTripFuelCost returns total fuel cost for a given trip ID.
func (r *FinanceRepository) GetTripFuelCost(ctx context.Context, tripID string) (float64, error) {
	query := `SELECT COALESCE(SUM(amount), 0) FROM fuel_transactions WHERE metadata->>'trip_id' = $1`
	var cost float64
	err := r.db.QueryRow(ctx, query, tripID).Scan(&cost)
	return cost, err
}

// GetTripTollCost returns total toll cost for a given trip ID.
func (r *FinanceRepository) GetTripTollCost(ctx context.Context, tripID string) (float64, error) {
	query := `SELECT COALESCE(SUM(amount), 0) FROM toll_transactions WHERE metadata->>'trip_id' = $1`
	var cost float64
	err := r.db.QueryRow(ctx, query, tripID).Scan(&cost)
	return cost, err
}

// GetTripDistance returns trip distance in km from the trips table.
func (r *FinanceRepository) GetTripDistance(ctx context.Context, tripID string) (float64, error) {
	query := `SELECT COALESCE(distance_km, 0) FROM trips WHERE id = $1`
	var dist float64
	err := r.db.QueryRow(ctx, query, tripID).Scan(&dist)
	if err != nil {
		return 0, nil // non-fatal, trips may not have this field
	}
	return dist, nil
}

// GetTripRevenue returns trip revenue from billing metadata.
func (r *FinanceRepository) GetTripRevenue(ctx context.Context, tripID string) (float64, error) {
	query := `SELECT COALESCE((cargo_details->>'value_inr')::DECIMAL, 0) FROM trips WHERE id = $1`
	var rev float64
	err := r.db.QueryRow(ctx, query, tripID).Scan(&rev)
	if err != nil {
		return 0, nil
	}
	return rev, err
}

func (r *FinanceRepository) ListPendingExpenseItemsForOCR(ctx context.Context, limit int) ([]models.ExpenseItem, error) {
	query := `SELECT id, report_id, invoice_number, category, amount, gst_amount, gst_rate,
	                 receipt_url, description, vendor_name, vendor_gstin,
	                 transaction_date, ocr_processed, ocr_confidence, created_at
	          FROM expense_items
	          WHERE receipt_url IS NOT NULL AND receipt_url <> '' AND ocr_processed = FALSE
	          ORDER BY created_at ASC
	          LIMIT $1`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.ExpenseItem, 0, limit)
	for rows.Next() {
		var item models.ExpenseItem
		var category string
		if err := rows.Scan(
			&item.ID, &item.ReportID, &item.InvoiceNumber, &category, &item.Amount, &item.GSTAmount, &item.GSTRate,
			&item.ReceiptURL, &item.Description, &item.VendorName, &item.VendorGSTIN,
			&item.TransactionDate, &item.OCRProcessed, &item.OCRConfidence, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.Category = models.ExpenseItemCategory(category)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *FinanceRepository) UpdateExpenseItemOCRData(ctx context.Context, item *models.ExpenseItem) error {
	query := `UPDATE expense_items
	          SET vendor_name = COALESCE($1, vendor_name),
	              vendor_gstin = COALESCE($2, vendor_gstin),
	              gst_amount = CASE WHEN $3 > 0 THEN $3 ELSE gst_amount END,
	              invoice_number = COALESCE($4, invoice_number),
	              ocr_processed = TRUE,
	              ocr_confidence = $5
	          WHERE id = $6`
	confidence := 0.0
	if item.OCRConfidence != nil {
		confidence = *item.OCRConfidence
	}
	_, err := r.db.Exec(ctx, query, item.VendorName, item.VendorGSTIN, item.GSTAmount, item.InvoiceNumber, confidence, item.ID)
	return err
}

// =============================================================================
// Finance Dashboard
// =============================================================================

func (r *FinanceRepository) RefreshFinanceDashboard(ctx context.Context) error {
	query := `REFRESH MATERIALIZED VIEW CONCURRENTLY finance_dashboard_metrics`
	_, err := r.db.Exec(ctx, query)
	return err
}

func (r *FinanceRepository) GetFinanceDashboardMetrics(ctx context.Context, orgID uuid.UUID) (*models.FinanceDashboardMetrics, error) {
	query := `SELECT org_id, metric_month, total_revenue, total_fuel_cost,
	                 total_expenses, profit_margin, avg_cost_per_km
	          FROM finance_dashboard_metrics WHERE org_id = $1`

	var m models.FinanceDashboardMetrics
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&m.OrgID, &m.MetricMonth, &m.TotalRevenue, &m.TotalFuelCost,
		&m.TotalExpenses, &m.ProfitMargin, &m.AvgCostPerKm,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
