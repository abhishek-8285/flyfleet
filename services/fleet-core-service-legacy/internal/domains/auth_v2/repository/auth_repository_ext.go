package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Domain       string    `json:"domain"`
	DefaultTheme string    `json:"default_theme"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (r *AuthRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
	query := "SELECT id, name, domain, default_theme, status, created_at, updated_at FROM organizations WHERE id = $1 AND is_deleted = false"
	var org Organization
	err := r.db.QueryRow(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Domain, &org.DefaultTheme, &org.Status, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *AuthRepository) ListOrganizations(ctx context.Context, limit, offset int32) ([]Organization, int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE is_deleted = false").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT id, name, domain, default_theme, status, created_at, updated_at FROM organizations WHERE is_deleted = false ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orgs []Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Domain, &org.DefaultTheme, &org.Status, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, 0, err
		}
		orgs = append(orgs, org)
	}
	return orgs, total, nil
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

	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&stats.ActiveUsers)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM drivers").Scan(&stats.TotalDrivers)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM organizations").Scan(&stats.TotalOrgs)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE status = 'active'").Scan(&stats.ActiveOrgs)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE status = 'pending'").Scan(&stats.PendingOrgs)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM organizations WHERE status = 'inactive'").Scan(&stats.InactiveOrgs)

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
	var total int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := "SELECT id, action, resource_type, resource_id, user_id, created_at, details FROM audit_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	rows, err := r.db.Query(ctx, query, limit, offset)
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
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM security_events WHERE event_type = 'LOGIN_FAILED' AND created_at > NOW() - INTERVAL '24 hours'").Scan(&failedLogins)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM active_sessions WHERE expires_at > NOW()").Scan(&activeSessions)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM ip_blocks WHERE expires_at > NOW()").Scan(&blockedIPs)
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM threat_alerts WHERE status = 'open'").Scan(&activeThreats)

	telemetry["failed_logins_24h"] = failedLogins
	telemetry["active_sessions"] = activeSessions
	telemetry["blocked_ips"] = blockedIPs
	telemetry["active_threats"] = activeThreats

	return telemetry, nil
}
