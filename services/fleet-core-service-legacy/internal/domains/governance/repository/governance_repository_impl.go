package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgGovernanceRepository struct {
	db *pgxpool.Pool
}

func NewGovernanceRepository(db *pgxpool.Pool) GovernanceRepository {
	return &pgGovernanceRepository{db: db}
}

func (r *pgGovernanceRepository) CreateRequest(ctx context.Context, req *models.GovernanceRequest) error {
	query := `
		INSERT INTO governance_requests (
			id, organization_id, requester_id, request_type, data_scope,
			purpose, format, delivery_method, status, expiry_date, comments,
			created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		req.ID, req.OrganizationID, req.RequesterID, req.RequestType, req.DataScope,
		req.Purpose, req.Format, req.DeliveryMethod, req.Status, req.ExpiryDate, req.Comments,
		req.CreatedBy, req.UpdatedBy,
	)
	return err
}

func (r *pgGovernanceRepository) GetRequestByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.GovernanceRequest, error) {
	query := `
		SELECT 
			id, organization_id, requester_id, request_type, data_scope,
			purpose, format, delivery_method, status, expiry_date, comments,
			created_at, created_by, updated_at, updated_by
		FROM governance_requests
		WHERE id = $1 AND organization_id = $2
	`
	var req models.GovernanceRequest
	err := r.db.QueryRow(ctx, query, id, orgID).Scan(
		&req.ID, &req.OrganizationID, &req.RequesterID, &req.RequestType, &req.DataScope,
		&req.Purpose, &req.Format, &req.DeliveryMethod, &req.Status, &req.ExpiryDate, &req.Comments,
		&req.CreatedAt, &req.CreatedBy, &req.UpdatedAt, &req.UpdatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

func (r *pgGovernanceRepository) ListRequests(ctx context.Context, filter models.GovernanceRequestFilter) (*models.GovernanceRequestListResponse, error) {
	where := []string{"organization_id = $1"}
	args := []interface{}{filter.OrganizationID}
	argIdx := 2

	if filter.Type != nil {
		where = append(where, fmt.Sprintf("request_type = $%d", argIdx))
		args = append(args, *filter.Type)
		argIdx++
	}

	if len(filter.Status) > 0 {
		statusList := []string{}
		for _, s := range filter.Status {
			statusList = append(statusList, string(s))
		}
		where = append(where, fmt.Sprintf("status = ANY($%d)", argIdx))
		args = append(args, statusList)
		argIdx++
	}

	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(purpose ILIKE $%d OR data_scope ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM governance_requests WHERE %s", strings.Join(where, " AND "))
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	limit := filter.PageSize
	query := fmt.Sprintf(`
		SELECT 
			id, organization_id, requester_id, request_type, data_scope,
			purpose, format, delivery_method, status, expiry_date, comments,
			created_at, created_by, updated_at, updated_by
		FROM governance_requests
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), argIdx, argIdx+1)

	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []models.GovernanceRequest{}
	for rows.Next() {
		var req models.GovernanceRequest
		err := rows.Scan(
			&req.ID, &req.OrganizationID, &req.RequesterID, &req.RequestType, &req.DataScope,
			&req.Purpose, &req.Format, &req.DeliveryMethod, &req.Status, &req.ExpiryDate, &req.Comments,
			&req.CreatedAt, &req.CreatedBy, &req.UpdatedAt, &req.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return &models.GovernanceRequestListResponse{
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Requests: requests,
	}, nil
}

func (r *pgGovernanceRepository) UpdateRequestStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.GovernanceRequestStatus, comments *string) error {
	query := `
		UPDATE governance_requests 
		SET status = $1, comments = COALESCE($2, comments), updated_at = NOW() 
		WHERE id = $3 AND organization_id = $4
	`
	_, err := r.db.Exec(ctx, query, status, comments, id, orgID)
	return err
}

func (r *pgGovernanceRepository) GetSettings(ctx context.Context, orgID uuid.UUID) (*models.GovernanceSettings, error) {
	query := `
		SELECT organization_id, retention_policies, masking_rules, auto_expiry_enabled, download_ttl_hours, created_at, updated_at
		FROM governance_settings WHERE organization_id = $1
	`
	var s models.GovernanceSettings
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&s.OrganizationID, &s.RetentionPolicies, &s.MaskingRules, &s.AutoExpiryEnabled, &s.DownloadTTLHours, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *pgGovernanceRepository) UpsertSettings(ctx context.Context, s *models.GovernanceSettings) error {
	query := `
		INSERT INTO governance_settings (
			organization_id, retention_policies, masking_rules, auto_expiry_enabled, download_ttl_hours, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, NOW(), $6, $7)
		ON CONFLICT (organization_id) DO UPDATE SET
			retention_policies = EXCLUDED.retention_policies,
			masking_rules = EXCLUDED.masking_rules,
			auto_expiry_enabled = EXCLUDED.auto_expiry_enabled,
			download_ttl_hours = EXCLUDED.download_ttl_hours,
			updated_at = NOW(),
			updated_by = EXCLUDED.updated_by
	`
	_, err := r.db.Exec(ctx, query,
		s.OrganizationID, s.RetentionPolicies, s.MaskingRules, s.AutoExpiryEnabled, s.DownloadTTLHours,
		s.CreatedBy, s.UpdatedBy,
	)
	return err
}

func (r *pgGovernanceRepository) CreateAuditLog(ctx context.Context, log *models.GovernanceAuditLog) error {
	query := `
		INSERT INTO governance_audit_logs (
			id, organization_id, user_id, action, module, change_detail, ip_address
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		log.ID, log.OrganizationID, log.UserID, log.Action, log.Module, log.ChangeDetail, log.IPAddress,
	)
	return err
}

func (r *pgGovernanceRepository) ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]models.GovernanceAuditLog, int, error) {
	// ... implementation for audit logs list if needed
	return nil, 0, nil
}
