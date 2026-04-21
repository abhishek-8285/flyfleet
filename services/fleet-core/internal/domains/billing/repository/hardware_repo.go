package repository

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HardwareRepository struct {
	db *pgxpool.Pool
}

func NewHardwareRepository(db *pgxpool.Pool) *HardwareRepository {
	return &HardwareRepository{db: db}
}

func (r *HardwareRepository) ListProducts(ctx context.Context) ([]models.HardwareProduct, error) {
	query := `SELECT id, name, model_code, category, one_time_price, lease_monthly_price, currency, created_at FROM billing_hardware_products`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.HardwareProduct
	for rows.Next() {
		var p models.HardwareProduct
		err := rows.Scan(&p.ID, &p.Name, &p.ModelCode, &p.Category, &p.OneTimePrice, &p.LeaseMonthlyPrice, &p.Currency, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *HardwareRepository) GetAssignmentByDevice(ctx context.Context, deviceID string) (*models.DeviceAssignment, error) {
	query := `SELECT id, tenant_id, device_id, product_id, billing_type, status, activated_at, invoice_id, created_at, updated_at 
	          FROM billing_device_assignments WHERE device_id = $1`

	var a models.DeviceAssignment
	err := r.db.QueryRow(ctx, query, deviceID).Scan(
		&a.ID, &a.TenantID, &a.DeviceID, &a.ProductID, &a.BillingType, &a.Status, &a.ActivatedAt, &a.InvoiceID, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *HardwareRepository) CreateAssignment(ctx context.Context, a *models.DeviceAssignment) error {
	query := `INSERT INTO billing_device_assignments (id, tenant_id, device_id, product_id, billing_type, status) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, a.ID, a.TenantID, a.DeviceID, a.ProductID, a.BillingType, a.Status)
	return err
}

func (r *HardwareRepository) ActivateAssignment(ctx context.Context, deviceID string) error {
	query := `UPDATE billing_device_assignments SET status = 'ACTIVE', activated_at = NOW(), updated_at = NOW() WHERE device_id = $1`
	_, err := r.db.Exec(ctx, query, deviceID)
	return err
}
