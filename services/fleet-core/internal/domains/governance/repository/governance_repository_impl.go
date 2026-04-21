package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/governance/models"
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
			id, tenant_id, requester_id, request_type, data_scope,
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

func (r *pgGovernanceRepository) GetRequestByID(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*models.GovernanceRequest, error) {
	query := `
		SELECT 
			id, tenant_id, requester_id, request_type, data_scope,
			purpose, format, delivery_method, status, expiry_date, comments,
			created_at, created_by, updated_at, updated_by
		FROM governance_requests
		WHERE id = $1 AND tenant_id = $2
	`
	var req models.GovernanceRequest
	err := r.db.QueryRow(ctx, query, id, tenantID).Scan(
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

func (r *pgGovernanceRepository) GetRequestByIDInternal(ctx context.Context, id uuid.UUID) (*models.GovernanceRequest, error) {
	query := `
		SELECT 
			id, tenant_id, requester_id, request_type, data_scope,
			purpose, format, delivery_method, status, expiry_date, comments,
			created_at, created_by, updated_at, updated_by
		FROM governance_requests
		WHERE id = $1
	`
	var req models.GovernanceRequest
	err := r.db.QueryRow(ctx, query, id).Scan(
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
	where := []string{"tenant_id = $1"}
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
			id, tenant_id, requester_id, request_type, data_scope,
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

func (r *pgGovernanceRepository) UpdateRequestStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status models.GovernanceRequestStatus, comments *string, expiryDate *time.Time) error {
	query := `
		UPDATE governance_requests
		SET status = $1, comments = COALESCE($2, comments), expiry_date = COALESCE($3, expiry_date), updated_at = NOW()
		WHERE id = $4 AND tenant_id = $5
	`
	_, err := r.db.Exec(ctx, query, status, comments, expiryDate, id, tenantID)
	return err
}

func (r *pgGovernanceRepository) UpdateVerificationStatus(ctx context.Context, id, tenantID uuid.UUID, status string, verifiedBy uuid.UUID) error {
	query := `
		UPDATE governance_requests
		SET verification_status = $1,
		    verified_at = CASE WHEN $1 = 'VERIFIED' THEN NOW() ELSE verified_at END,
		    verified_by = $2,
		    updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4
	`
	_, err := r.db.Exec(ctx, query, status, verifiedBy, id, tenantID)
	return err
}

func (r *pgGovernanceRepository) GetSettings(ctx context.Context, tenantID uuid.UUID) (*models.GovernanceSettings, error) {
	query := `
		SELECT tenant_id, retention_policies, masking_rules, auto_expiry_enabled, download_ttl_hours, created_at, updated_at
		FROM governance_settings WHERE tenant_id = $1
	`
	var s models.GovernanceSettings
	err := r.db.QueryRow(ctx, query, tenantID).Scan(
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
			tenant_id, retention_policies, masking_rules, auto_expiry_enabled, download_ttl_hours,
			created_at, updated_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), $6, $7)
		ON CONFLICT (tenant_id) DO UPDATE SET
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
			id, tenant_id, user_id, action, module, change_detail, ip_address
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		log.ID, log.OrganizationID, log.UserID, log.Action, log.Module, log.ChangeDetail, log.IPAddress,
	)
	return err
}

func (r *pgGovernanceRepository) ListAuditLogs(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]models.GovernanceAuditLog, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// Count total
	countQuery := `
		SELECT COUNT(*) FROM governance_audit_logs
		WHERE tenant_id = $1
	`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := `
		SELECT id, tenant_id, user_id, action, module, change_detail, ip_address, created_at
		FROM governance_audit_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := []models.GovernanceAuditLog{}
	for rows.Next() {
		var log models.GovernanceAuditLog
		err := rows.Scan(
			&log.ID, &log.OrganizationID, &log.UserID, &log.Action, &log.Module, &log.ChangeDetail, &log.IPAddress, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *pgGovernanceRepository) ExpireRequests(ctx context.Context) ([]models.ExpiredRequestInfo, error) {
	query := `
		UPDATE governance_requests
		SET status = 'EXPIRED', updated_at = NOW()
		WHERE status = 'APPROVED'
		AND expiry_date IS NOT NULL
		AND expiry_date < NOW()
		RETURNING id, tenant_id
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expired []models.ExpiredRequestInfo
	for rows.Next() {
		var info models.ExpiredRequestInfo
		if err := rows.Scan(&info.ID, &info.OrganizationID); err != nil {
			return expired, err
		}
		expired = append(expired, info)
	}

	return expired, nil
}

func (r *pgGovernanceRepository) CreateExport(ctx context.Context, e *models.GovernanceExport) error {
	query := `
		INSERT INTO governance_exports (id, request_id, file_path, file_size, checksum, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, e.ID, e.RequestID, e.FilePath, e.FileSize, e.Checksum, e.CreatedAt, e.ExpiresAt)
	return err
}

func (r *pgGovernanceRepository) GetExportByRequestID(ctx context.Context, requestID uuid.UUID) (*models.GovernanceExport, error) {
	query := `
		SELECT id, request_id, file_path, file_size, checksum, created_at, expires_at
		FROM governance_exports WHERE request_id = $1
	`
	var e models.GovernanceExport
	err := r.db.QueryRow(ctx, query, requestID).Scan(&e.ID, &e.RequestID, &e.FilePath, &e.FileSize, &e.Checksum, &e.CreatedAt, &e.ExpiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *pgGovernanceRepository) CreateDownloadToken(ctx context.Context, t *models.DownloadToken) error {
	query := `
		INSERT INTO governance_download_tokens (token, export_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query, t.Token, t.ExportID, t.ExpiresAt, t.CreatedAt)
	return err
}

func (r *pgGovernanceRepository) GetExportByToken(ctx context.Context, token uuid.UUID) (*models.GovernanceExport, error) {
	query := `
		SELECT e.id, e.request_id, e.file_path, e.file_size, e.checksum, e.created_at, e.expires_at
		FROM governance_exports e
		JOIN governance_download_tokens t ON e.id = t.export_id
		WHERE t.token = $1 AND t.expires_at > NOW()
	`
	var e models.GovernanceExport
	err := r.db.QueryRow(ctx, query, token).Scan(&e.ID, &e.RequestID, &e.FilePath, &e.FileSize, &e.Checksum, &e.CreatedAt, &e.ExpiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *pgGovernanceRepository) CheckLegalHold(ctx context.Context, entityID uuid.UUID) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM legal_holds WHERE entity_id = $1 AND (expires_at IS NULL OR expires_at > NOW())`
	err := r.db.QueryRow(ctx, query, entityID).Scan(&count)
	return count > 0, err
}

func (r *pgGovernanceRepository) SaveExportLineage(ctx context.Context, lineage *models.ExportLineage) error {
	query := `
		INSERT INTO governance_export_lineage (id, request_id, table_name, record_count, export_size_bytes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, lineage.ID, lineage.RequestID, lineage.TableName, lineage.RecordCount, lineage.ExportSizeBytes, lineage.CreatedAt)
	return err
}

func (r *pgGovernanceRepository) SaveComplianceReport(ctx context.Context, report *models.ComplianceReport) error {
	query := `
		INSERT INTO governance_compliance_reports (id, request_id, export_checksum, export_size_bytes, processing_time_ms, tables_accessed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, report.ID, report.RequestID, report.ExportChecksum, report.ExportSizeBytes, report.ProcessingTimeMs, report.TablesAccessed, report.CreatedAt)
	return err
}

func (r *pgGovernanceRepository) ListAllSettings(ctx context.Context) ([]models.GovernanceSettings, error) {
	query := `SELECT tenant_id, retention_policies, auto_expiry_enabled, download_ttl_hours FROM governance_settings`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.GovernanceSettings
	for rows.Next() {
		var s models.GovernanceSettings
		if err := rows.Scan(&s.OrganizationID, &s.RetentionPolicies, &s.AutoExpiryEnabled, &s.DownloadTTLHours); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func (r *pgGovernanceRepository) EnforceRetentionPolicy(ctx context.Context, tenantID uuid.UUID, tableName string, retentionDays int) (int64, error) {
	// Only allow known tables to prevent SQL injection and specify entity mapping
	entityMap := map[string]string{
		"telemetry": "device_id",
		"trips":     "driver_id",
		"fuel_logs": "driver_id",
	}

	entityCol, approved := entityMap[tableName]
	if !approved {
		return 0, fmt.Errorf("table %s not allowed for retention enforcement", tableName)
	}

	query := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE tenant_id = $1 
		AND created_at < NOW() - INTERVAL '1 day' * $2
		AND %s NOT IN (
			SELECT entity_id FROM legal_holds 
			WHERE tenant_id = $1 
			AND (expires_at IS NULL OR expires_at > NOW())
		)
	`, tableName, entityCol)

	tag, err := r.db.Exec(ctx, query, tenantID, retentionDays)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
