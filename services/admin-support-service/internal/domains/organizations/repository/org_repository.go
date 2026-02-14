package repository

import (
	"context"
	"fmt"

	"github.com/flyfleet/admin-support-service/internal/domains/organizations/models"
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
		SELECT id, name, domain, status, created_at, updated_at
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
			&org.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		// Mock Plan and RenewalDate as they aren't in the schema yet
		org.Plan = "Enterprise"
		org.RenewalDate = org.CreatedAt.AddDate(1, 0, 0)
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
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalDrivers) // Fallback drivers to users
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM trips").Scan(&stats.ActiveTrips) // Active trips?
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM threat_alerts WHERE status = 'open'").Scan(&stats.AlertsCount)
	
	// Distances and fuel usually come from telemetry or aggregated data
	// Since we don't have that yet, use some realistic mock values
	stats.TotalDistanceToday = 1245.5
	stats.FuelConsumedToday = 450.2
	
	return stats, nil
}
