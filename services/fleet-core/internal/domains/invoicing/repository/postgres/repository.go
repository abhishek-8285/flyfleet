package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyfleet/fleet-core/internal/domains/invoicing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvoiceRepository struct {
	pool *pgxpool.Pool
}

func NewInvoiceRepository(pool *pgxpool.Pool) *InvoiceRepository {
	return &InvoiceRepository{pool: pool}
}

func (r *InvoiceRepository) Create(ctx context.Context, inv *invoicing.Invoice) (*invoicing.Invoice, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO invoices (
			tenant_id, invoice_number, customer_name, customer_email, customer_phone, billing_address,
			invoice_date, due_date, status, subtotal, tax_amount, total_amount, total, balance_due, amount_paid, currency, notes,
			template_type, trip_id, vehicle_no, driver_name, driver_mobile, pickup_location, drop_location,
			pickup_time, drop_time, map_url, billed_distance, place_of_supply, reverse_charge,
			gstin_seller, pan_seller, eway_bill_no, eway_bill_date, irn, ack_no, ack_date,
			bank_name, account_no, ifsc_code, branch_name, beneficiary_name,
			created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		inv.TenantID, inv.InvoiceNumber, inv.CustomerName, inv.CustomerEmail, inv.CustomerPhone, inv.BillingAddress,
		inv.InvoiceDate, inv.DueDate, inv.Status, inv.Subtotal, inv.TaxAmount, inv.TotalAmount, inv.TotalAmount, inv.BalanceDue, inv.AmountPaid, inv.Currency, inv.Notes,
		inv.TemplateType, inv.TripID, inv.VehicleNo, inv.DriverName, inv.DriverMobile, inv.PickupLocation, inv.DropLocation,
		inv.PickupTime, inv.DropTime, inv.MapUrl, inv.BilledDistance, inv.PlaceOfSupply, inv.ReverseCharge,
		inv.GstinSeller, inv.PanSeller, inv.EWayBillNo, inv.EWayBillDate, inv.IRN, inv.AckNo, inv.AckDate,
		inv.BankName, inv.AccountNo, inv.IFSC, inv.Branch, inv.Beneficiary,
		inv.CreatedBy, inv.UpdatedBy,
	).Scan(&inv.ID, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert invoice: %w", err)
	}

	// Batch insert items
	if len(inv.Items) > 0 {
		var args []interface{}
		var placeholders []string
		for i, item := range inv.Items {
			item.ID = uuid.New()
			pBase := i * 15
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", 
				pBase+1, pBase+2, pBase+3, pBase+4, pBase+5, pBase+6, pBase+7, pBase+8, pBase+9, pBase+10, pBase+11, pBase+12, pBase+13, pBase+14, pBase+15))
			args = append(args,
				item.ID,
				inv.TenantID,
				inv.ID,
				item.Description,
				item.Quantity,
				item.UnitPrice,
				item.TaxPercent,
				item.Amount,
				item.Amount,
				item.HSNSAC,
				item.Unit,
				item.ListPrice,
				item.DiscountAmount,
				inv.CreatedBy,
				inv.UpdatedBy,
			)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO invoice_items (id, tenant_id, invoice_id, description, quantity, unit_price, tax_percent, amount, total, hsn_sac, unit, list_price, discount_amount, created_by, updated_by) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		_, err = tx.Exec(ctx, batchQuery, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to batch insert invoice items: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return inv, nil
}

func (r *InvoiceRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*invoicing.Invoice, error) {
	query := `
		SELECT id, tenant_id, COALESCE(invoice_number, '') as invoice_number, 
		       COALESCE(customer_name, '') as customer_name, 
		       COALESCE(customer_email, '') as customer_email, 
		       COALESCE(customer_phone, '') as customer_phone, 
		       COALESCE(billing_address, '') as billing_address,
		       invoice_date, due_date, status, COALESCE(subtotal, 0), COALESCE(tax_amount, 0), COALESCE(total_amount, 0), COALESCE(amount_paid, 0), COALESCE(balance_due, 0),
		       COALESCE(currency, 'INR') as currency, COALESCE(notes, '') as notes,
		       COALESCE(template_type, 'TAX_INVOICE'), trip_id, 
		       COALESCE(vehicle_no, ''), COALESCE(driver_name, ''), COALESCE(driver_mobile, ''),
		       COALESCE(pickup_location, ''), COALESCE(drop_location, ''), pickup_time, drop_time,
		       COALESCE(map_url, ''), COALESCE(billed_distance, 0),
		       COALESCE(place_of_supply, ''), reverse_charge,
		       COALESCE(gstin_seller, ''), COALESCE(pan_seller, ''),
		       COALESCE(eway_bill_no, ''), eway_bill_date, COALESCE(irn, ''),
		       COALESCE(ack_no, ''), ack_date,
		       COALESCE(bank_name, ''), COALESCE(account_no, ''), COALESCE(ifsc_code, ''),
		       COALESCE(branch_name, ''), COALESCE(beneficiary_name, ''),
		       created_at, COALESCE(created_by, '00000000-0000-0000-0000-000000000000') as created_by, 
		       updated_at, COALESCE(updated_by, '00000000-0000-0000-0000-000000000000') as updated_by
		FROM invoices
		WHERE id = $1 AND tenant_id = $2 AND is_deleted = false
	`
	inv := &invoicing.Invoice{}
	err := r.pool.QueryRow(ctx, query, id, tenantID).Scan(
		&inv.ID, &inv.TenantID, &inv.InvoiceNumber, &inv.CustomerName, &inv.CustomerEmail, &inv.CustomerPhone, &inv.BillingAddress,
		&inv.InvoiceDate, &inv.DueDate, &inv.Status, &inv.Subtotal, &inv.TaxAmount, &inv.TotalAmount, &inv.AmountPaid, &inv.BalanceDue,
		&inv.Currency, &inv.Notes,
		&inv.TemplateType, &inv.TripID, &inv.VehicleNo, &inv.DriverName, &inv.DriverMobile,
		&inv.PickupLocation, &inv.DropLocation, &inv.PickupTime, &inv.DropTime,
		&inv.MapUrl, &inv.BilledDistance,
		&inv.PlaceOfSupply, &inv.ReverseCharge,
		&inv.GstinSeller, &inv.PanSeller,
		&inv.EWayBillNo, &inv.EWayBillDate, &inv.IRN,
		&inv.AckNo, &inv.AckDate,
		&inv.BankName, &inv.AccountNo, &inv.IFSC,
		&inv.Branch, &inv.Beneficiary,
		&inv.CreatedAt, &inv.CreatedBy, &inv.UpdatedAt, &inv.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}

	// Fetch items
	itemQuery := `
		SELECT id, invoice_id, COALESCE(description, ''), COALESCE(quantity, 0), COALESCE(unit_price, 0), COALESCE(tax_percent, 0), COALESCE(amount, 0), COALESCE(total, 0),
		       COALESCE(hsn_sac, ''), COALESCE(unit, ''), COALESCE(list_price, 0), COALESCE(discount_amount, 0),
		       created_at, COALESCE(created_by, '00000000-0000-0000-0000-000000000000') as created_by, updated_at, COALESCE(updated_by, '00000000-0000-0000-0000-000000000000') as updated_by
		FROM invoice_items
		WHERE invoice_id = $1 AND tenant_id = $2 AND is_deleted = false
	`
	rows, err := r.pool.Query(ctx, itemQuery, id, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item invoicing.InvoiceItem
		err := rows.Scan(
			&item.ID, &item.InvoiceID, &item.Description, &item.Quantity, &item.UnitPrice, &item.TaxPercent, &item.Amount, &item.Total,
			&item.HSNSAC, &item.Unit, &item.ListPrice, &item.DiscountAmount,
			&item.CreatedAt, &item.CreatedBy, &item.UpdatedAt, &item.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		inv.Items = append(inv.Items, item)
	}

	return inv, nil
}

func (r *InvoiceRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]invoicing.Invoice, int64, error) {
	countQuery := `SELECT COUNT(*) FROM invoices WHERE tenant_id = $1 AND is_deleted = false`
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, tenant_id, COALESCE(invoice_number, '') as invoice_number, 
		       COALESCE(customer_name, '') as customer_name, 
		       status, COALESCE(total_amount, 0), invoice_date, due_date, created_at, 
		       COALESCE(created_by, '00000000-0000-0000-0000-000000000000') as created_by, 
		       updated_at, COALESCE(updated_by, '00000000-0000-0000-0000-000000000000') as updated_by
		FROM invoices
		WHERE tenant_id = $1 AND is_deleted = false
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []invoicing.Invoice
	for rows.Next() {
		var inv invoicing.Invoice
		err := rows.Scan(&inv.ID, &inv.TenantID, &inv.InvoiceNumber, &inv.CustomerName, &inv.Status, &inv.TotalAmount, &inv.InvoiceDate, &inv.DueDate, &inv.CreatedAt, &inv.CreatedBy, &inv.UpdatedAt, &inv.UpdatedBy)
		if err != nil {
			return nil, 0, err
		}
		invoices = append(invoices, inv)
	}

	return invoices, total, nil
}

func (r *InvoiceRepository) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status invoicing.InvoiceStatus, updatedBy uuid.UUID) error {
	query := `UPDATE invoices SET status = $1, updated_at = CURRENT_TIMESTAMP, updated_by = $2 WHERE id = $3 AND tenant_id = $4 AND is_deleted = false`
	_, err := r.pool.Exec(ctx, query, status, updatedBy, id, tenantID)
	return err
}

func (r *InvoiceRepository) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `UPDATE invoices SET is_deleted = true, deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND tenant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, tenantID)
	return err
}
