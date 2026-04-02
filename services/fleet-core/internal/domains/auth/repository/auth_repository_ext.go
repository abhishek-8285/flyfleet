package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth/models"
	"github.com/google/uuid"
)

func (r *AuthRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	query := "SELECT id, name, plan as type, COALESCE(logo_url, '') as logo_url, COALESCE(domain, '') as domain, COALESCE(default_theme, '') as default_theme, COALESCE(status, '') as status, created_at, updated_at FROM organizations WHERE id = $1"
	start := time.Now()
	var org models.Organization
	err := r.db.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Type, &org.LogoURL, &org.Domain, &org.DefaultTheme, &org.Status, &org.CreatedAt, &org.UpdatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *AuthRepository) ListOrganizations(ctx context.Context, limit, offset int32) ([]models.Organization, int32, error) {
	start1 := time.Now()
	var total int32
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM organizations").Scan(&total)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM organizations", time.Since(start1))
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT id, name, plan as type, COALESCE(logo_url, '') as logo_url, COALESCE(domain, '') as domain, COALESCE(default_theme, '') as default_theme, COALESCE(status, '') as status, created_at, updated_at FROM organizations ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	start2 := time.Now()
	rows, err := r.db.Query(ctx, query, limit, offset)
	metrics.RecordDBQuery(query, time.Since(start2))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orgs []models.Organization
	for rows.Next() {
		var org models.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Type, &org.LogoURL, &org.Domain, &org.DefaultTheme, &org.Status, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, 0, err
		}
		orgs = append(orgs, org)
	}
	return orgs, total, nil
}

func (r *AuthRepository) GetOrganizationsByEmail(ctx context.Context, email string) ([]models.Organization, error) {
	query := `
		SELECT
			o.id,
			o.name,
			o.plan as type,
			COALESCE(o.logo_url, '') as logo_url,
			COALESCE(o.domain, '') as domain,
			COALESCE(o.default_theme, '') as default_theme,
			COALESCE(o.status, '') as status,
			o.created_at,
			o.updated_at
		FROM organizations o
		JOIN user_organizations uo ON o.id = uo.org_id
		JOIN users u ON uo.user_id = u.id
		WHERE u.email = $1
	`
	// Security: Bypass RLS for organization resolution
	// Use a transaction to ensure set_config and the query use the SAME connection
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, _ = tx.Exec(ctx, "SELECT set_config('app.current_org_id', '00000000-0000-0000-0000-000000000001', false)")

	start := time.Now()
	rows, err := tx.Query(ctx, query, email)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user organizations: %w", err)
	}
	defer rows.Close()

	var orgs []models.Organization
	for rows.Next() {
		var org models.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Type, &org.LogoURL, &org.Domain, &org.DefaultTheme, &org.Status, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, org)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return orgs, nil
}

type UserStats struct {
	TotalUsers   int32
	ActiveUsers  int32
	TotalDrivers int32
	TotalOrgs    int32
	ActiveOrgs   int32
	PendingOrgs  int32
	InactiveOrgs int32
}

func (r *AuthRepository) GetStats(ctx context.Context) (*UserStats, error) {
	stats := &UserStats{}

	start := time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM users", time.Since(start))

	start = time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&stats.ActiveUsers)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM users WHERE status = 'active'", time.Since(start))

	start = time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM drivers").Scan(&stats.TotalDrivers)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM drivers", time.Since(start))

	start = time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM organizations").Scan(&stats.TotalOrgs)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM organizations", time.Since(start))

	start = time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE status = 'active'").Scan(&stats.ActiveOrgs)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM organizations WHERE status = 'active'", time.Since(start))

	start = time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE status = 'pending'").Scan(&stats.PendingOrgs)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM organizations WHERE status = 'pending'", time.Since(start))

	start = time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE status = 'inactive'").Scan(&stats.InactiveOrgs)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM organizations WHERE status = 'inactive'", time.Since(start))

	return stats, nil
}

type AuditLog struct {
	ID        uuid.UUID
	Action    string
	Entity    string
	EntityID  string
	ActorID   uuid.UUID
	CreatedAt time.Time
	Metadata  []byte // JSON
}

func (r *AuthRepository) GetAuditLogs(ctx context.Context, limit, offset int32) ([]AuditLog, int64, error) {
	start1 := time.Now()
	var total int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs").Scan(&total)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM audit_logs", time.Since(start1))
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT id, action, resource_type, resource_id, user_id, created_at, details FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	start2 := time.Now()
	rows, err := r.db.Query(ctx, query, limit, offset)
	metrics.RecordDBQuery(query, time.Since(start2))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		if err := rows.Scan(&log.ID, &log.Action, &log.Entity, &log.EntityID, &log.ActorID, &log.CreatedAt, &log.Metadata); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}

	return logs, total, nil
}

func (r *AuthRepository) GetSecurityTelemetry(ctx context.Context) (map[string]any, error) {
	telemetry := make(map[string]any)

	var failedLogins, activeSessions, blockedIPs, activeThreats int

	start := time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM security_events WHERE event_type = 'LOGIN_FAILED' AND created_at > NOW() - INTERVAL '24 hours'").Scan(&failedLogins)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM security_events WHERE event_type = 'LOGIN_FAILED' AND created_at > NOW() - INTERVAL '24 hours'", time.Since(start))

	start = time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM active_sessions WHERE expires_at > NOW()").Scan(&activeSessions)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM active_sessions WHERE expires_at > NOW()", time.Since(start))

	start = time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM ip_blocks WHERE expires_at > NOW()").Scan(&blockedIPs)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM ip_blocks WHERE expires_at > NOW()", time.Since(start))

	start = time.Now()
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM threat_alerts WHERE status = 'open'").Scan(&activeThreats)
	metrics.RecordDBQuery("SELECT COUNT(*) FROM threat_alerts WHERE status = 'open'", time.Since(start))

	telemetry["failed_logins_24h"] = failedLogins
	telemetry["active_sessions"] = activeSessions
	telemetry["blocked_ips"] = blockedIPs
	telemetry["active_threats"] = activeThreats

	return telemetry, nil
}
