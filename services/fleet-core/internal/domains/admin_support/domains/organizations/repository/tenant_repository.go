package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type TenantRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewTenantRepository(db *pgxpool.Pool, logger *zap.Logger) *TenantRepository {
	return &TenantRepository{
		db:     db,
		logger: logger,
	}
}

func (r *TenantRepository) ListTenants(ctx context.Context, page, limit int, searchQuery string) ([]models.Tenant, int, error) {
	offset := (page - 1) * limit

	query := `
		SELECT id, name, domain, status, created_at, created_by, updated_at, updated_by
		FROM tenants
		WHERE is_deleted = false
	`

	countQuery := `SELECT COUNT(*) FROM tenants WHERE is_deleted = false`

	args := []interface{}{}
	argIdx := 1

	if searchQuery != "" {
		filter := fmt.Sprintf(" AND (name ILIKE $%d OR domain ILIKE $%d)", argIdx, argIdx+1)
		query += filter
		countQuery += filter
		args = append(args, "%"+searchQuery+"%", "%"+searchQuery+"%")
		argIdx += 2
	}

	query += fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var total int
	err := r.db.QueryRow(ctx, countQuery, args[:argIdx-1]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tenants []models.Tenant
	for rows.Next() {
		var tenant models.Tenant
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Domain,
			&tenant.Status,
			&tenant.CreatedAt,
			&tenant.CreatedBy,
			&tenant.UpdatedAt,
			&tenant.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		tenants = append(tenants, tenant)
	}

	return tenants, total, nil
}

func (r *TenantRepository) GetTenantStats(ctx context.Context) (models.TenantStats, error) {
	var stats models.TenantStats

	// Use a transaction or multiple queries
	queries := map[string]*int{
		"SELECT COUNT(*) FROM tenants WHERE is_deleted = false":                         &stats.TotalTenants,
		"SELECT COUNT(*) FROM tenants WHERE status = 'active' AND is_deleted = false":   &stats.ActiveTenants,
		"SELECT COUNT(*) FROM tenants WHERE status = 'pending' AND is_deleted = false":  &stats.PendingTenants,
		"SELECT COUNT(*) FROM tenants WHERE status = 'inactive' AND is_deleted = false": &stats.InactiveTenants,
		"SELECT COUNT(*) FROM users":    &stats.TotalUsers,
		"SELECT COUNT(*) FROM vehicles": &stats.TotalVehicles,
	}

	for q, ptr := range queries {
		err := r.db.QueryRow(ctx, q).Scan(ptr)
		if err != nil {
			r.logger.Warn("Failed to fetch stat", zap.String("query", q), zap.Error(err))
		}
	}

	// Get actual device count from devices table (not assuming 1:1 with vehicles)
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM devices").Scan(&stats.TotalDevices)
	if err != nil {
		r.logger.Warn("Failed to fetch device count", zap.Error(err))
		stats.TotalDevices = 0
	}

	// Active Alerts
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM threat_alerts WHERE status = 'open'").Scan(&stats.ActiveAlerts)
	if err != nil {
		r.logger.Warn("Failed to fetch alert stats", zap.Error(err))
	}

	return stats, nil
}

func (r *TenantRepository) GetDashboardStats(ctx context.Context) (models.DashboardStats, error) {
	var stats models.DashboardStats

	// Aggregate from multiple tables
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM vehicles").Scan(&stats.TotalVehicles)
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM vehicles WHERE status = 'active'").Scan(&stats.ActiveVehicles)
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalDrivers)
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM trips WHERE status = 'in_progress'").Scan(&stats.ActiveTrips)
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM threat_alerts WHERE status = 'open'").Scan(&stats.AlertsCount)

	// Real Aggregates for today
	_ = r.db.QueryRow(ctx, "SELECT COALESCE(SUM(actual_distance_km), 0) FROM trips WHERE started_at >= CURRENT_DATE").Scan(&stats.TotalDistanceToday)
	_ = r.db.QueryRow(ctx, "SELECT COALESCE(SUM(quantity), 0) FROM fuel_logs WHERE transaction_date >= CURRENT_DATE").Scan(&stats.FuelConsumedToday)

	return stats, nil
}

func (r *TenantRepository) CreateTenant(ctx context.Context, req models.CreateTenantRequest, creatorID uuid.UUID) (models.CreateTenantResponse, error) {
	id := uuid.New()
	now := time.Now().UTC()
	status := "active"
	plan := "enterprise"

	if req.SubscriptionTier != "" {
		plan = req.SubscriptionTier
	}

	defaultTheme := "system"
	if req.DefaultTheme != "" {
		defaultTheme = req.DefaultTheme
	}

	query := `
		INSERT INTO tenants (id, name, domain, status, plan, contact_email, default_theme, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, name, domain, status, created_at
	`

	var resp models.CreateTenantResponse
	err := r.db.QueryRow(ctx, query,
		id,
		req.Name,
		req.Domain,
		status,
		plan,
		req.ContactEmail,
		defaultTheme,
		now,
		creatorID,
		now,
		creatorID,
	).Scan(&resp.ID, &resp.Name, &resp.Domain, &resp.Status, &resp.CreatedAt)

	if err != nil {
		r.logger.Error("Failed to create tenant", zap.Error(err))
		return models.CreateTenantResponse{}, err
	}

	return resp, nil
}

func (r *TenantRepository) DeleteTenant(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	// Check for dependent records before deleting
	var userCount, vehicleCount int

	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM user_tenants WHERE tenant_id = $1 AND status = 'active'", id).Scan(&userCount)
	if err != nil {
		r.logger.Error("Failed to count users", zap.String("tenant_id", id.String()), zap.Error(err))
		return fmt.Errorf("failed to check dependencies: %w", err)
	}

	if userCount > 0 {
		return fmt.Errorf("cannot delete tenant with %d active users", userCount)
	}

	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM vehicles WHERE tenant_id = $1 AND deleted_at IS NULL", id).Scan(&vehicleCount)
	if err != nil {
		r.logger.Error("Failed to count vehicles", zap.String("tenant_id", id.String()), zap.Error(err))
		return fmt.Errorf("failed to check dependencies: %w", err)
	}

	if vehicleCount > 0 {
		return fmt.Errorf("cannot delete tenant with %d active vehicles", vehicleCount)
	}

	query := `UPDATE tenants SET is_deleted = true, updated_at = $2, updated_by = $3 WHERE id = $1`
	_, err = r.db.Exec(ctx, query, id, time.Now().UTC(), updatedBy)
	if err != nil {
		r.logger.Error("Failed to delete tenant", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}

func (r *TenantRepository) ListBranches(ctx context.Context, tenantID uuid.UUID) ([]models.Branch, error) {
	query := `
		SELECT id, name, code, location
		FROM branches
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		r.logger.Error("Failed to list branches", zap.String("tenant_id", tenantID.String()), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var branches []models.Branch
	for rows.Next() {
		var b models.Branch
		if err := rows.Scan(&b.ID, &b.Name, &b.Code, &b.Location); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}

	return branches, nil
}

func (r *TenantRepository) ListCostCenters(ctx context.Context, tenantID uuid.UUID) ([]models.CostCenter, error) {
	query := `
		SELECT id, name, code, description
		FROM cost_centers
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		r.logger.Error("Failed to list cost centers", zap.String("tenant_id", tenantID.String()), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var costCenters []models.CostCenter
	for rows.Next() {
		var c models.CostCenter
		if err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.Description); err != nil {
			return nil, err
		}
		costCenters = append(costCenters, c)
	}

	return costCenters, nil
}
