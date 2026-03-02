package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/organizations/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type OrganizationRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewOrganizationRepository(db *pgxpool.Pool, logger *zap.Logger) *OrganizationRepository {
	return &OrganizationRepository{
		db:     db,
		logger: logger,
	}
}

func (r *OrganizationRepository) ListOrganizations(ctx context.Context, page, limit int, searchQuery string) ([]models.Organization, int, error) {
	offset := (page - 1) * limit
	
	query := `
		SELECT id, name, domain, status, created_at, created_by, updated_at, updated_by
		FROM organizations
		WHERE is_deleted = false
	`
	
	countQuery := `SELECT COUNT(*) FROM organizations WHERE is_deleted = false`
	
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
	
	var orgs []models.Organization
	for rows.Next() {
		var org models.Organization
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Domain,
			&org.Status,
			&org.CreatedAt,
			&org.CreatedBy,
			&org.UpdatedAt,
			&org.UpdatedBy,
		)
		if err != nil {
			return nil, 0, err
		}
		orgs = append(orgs, org)
	}
	
	return orgs, total, nil
}

func (r *OrganizationRepository) GetOrganizationStats(ctx context.Context) (models.OrganizationStats, error) {
	var stats models.OrganizationStats
	
	// Use a transaction or multiple queries
	queries := map[string]*int{
		"SELECT COUNT(*) FROM organizations WHERE is_deleted = false": &stats.TotalOrgs,
		"SELECT COUNT(*) FROM organizations WHERE status = 'active' AND is_deleted = false": &stats.ActiveOrgs,
		"SELECT COUNT(*) FROM organizations WHERE status = 'pending' AND is_deleted = false": &stats.PendingOrgs,
		"SELECT COUNT(*) FROM organizations WHERE status = 'inactive' AND is_deleted = false": &stats.InactiveOrgs,
		"SELECT COUNT(*) FROM users": &stats.TotalUsers,
		"SELECT COUNT(*) FROM vehicles": &stats.TotalVehicles,
	}
	
	for q, ptr := range queries {
		err := r.db.QueryRow(ctx, q).Scan(ptr)
		if err != nil {
			r.logger.Warn("Failed to fetch stat", zap.String("query", q), zap.Error(err))
		}
	}
	
	stats.TotalDevices = stats.TotalVehicles // Assuming 1:1 for now
	
	// Active Alerts
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM threat_alerts WHERE status = 'open'").Scan(&stats.ActiveAlerts)
	if err != nil {
		r.logger.Warn("Failed to fetch alert stats", zap.Error(err))
	}
	
	return stats, nil
}

func (r *OrganizationRepository) GetDashboardStats(ctx context.Context) (models.DashboardStats, error) {
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

func (r *OrganizationRepository) CreateOrganization(ctx context.Context, req models.CreateOrganizationRequest, creatorID uuid.UUID) (models.CreateOrganizationResponse, error) {
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
		INSERT INTO organizations (id, name, domain, status, plan, contact_email, default_theme, created_at, created_by, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, name, domain, status, created_at
	`

	var resp models.CreateOrganizationResponse
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
		r.logger.Error("Failed to create organization", zap.Error(err))
		return models.CreateOrganizationResponse{}, err
	}

	return resp, nil
}

func (r *OrganizationRepository) DeleteOrganization(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) error {
	query := `UPDATE organizations SET is_deleted = true, updated_at = $2, updated_by = $3 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, time.Now().UTC(), updatedBy)
	if err != nil {
		r.logger.Error("Failed to delete organization", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}
