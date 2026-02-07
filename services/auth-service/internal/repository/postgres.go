package repository

import (
	"context"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
)

// User operations
// User operations
func (r *postgresRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, name, COALESCE(phone, ''), mfa_enabled, COALESCE(mfa_secret, ''), status, failed_logins, locked_until, verified, COALESCE(theme, ''), created_at, updated_at FROM users WHERE email = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Phone,
		&user.MFAEnabled, &user.MFASecret, &user.Status, &user.FailedLogins, &user.LockedUntil, &user.Verified, &user.Theme, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, nil
	}
	return &user, nil
}

func (r *postgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, email, password_hash, name, phone, status, theme, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.Name, user.Phone, user.Status, user.Theme, time.Now(), time.Now())
	return err
}

func (r *postgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `SELECT id, email, password_hash, name, phone, mfa_enabled, mfa_secret, status, failed_logins, locked_until, verified, theme, created_at, updated_at FROM users WHERE id = $1`
	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Phone,
		&user.MFAEnabled, &user.MFASecret, &user.Status, &user.FailedLogins, &user.LockedUntil, &user.Verified, &user.Theme, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *postgresRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `UPDATE users SET email=$1, name=$2, phone=$3, status=$4, theme=$5, updated_at=NOW() WHERE id=$6`
	_, err := r.db.Exec(ctx, query, user.Email, user.Name, user.Phone, user.Status, user.Theme, user.ID)
	return err
}

func (r *postgresRepository) UpdateTheme(ctx context.Context, userID uuid.UUID, theme string) error {
	query := `UPDATE users SET theme=$1, theme_updated_at=NOW(), updated_at=NOW() WHERE id=$2`
	_, err := r.db.Exec(ctx, query, theme, userID)
	return err
}

// MFA operations
func (r *postgresRepository) SaveMFACode(ctx context.Context, mfa *models.MFACode) error {
	query := `INSERT INTO mfa_codes (user_id, code, method, expires_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, mfa.UserID, mfa.Code, mfa.Method, mfa.ExpiresAt)
	return err
}
// ... existing MFA code ...

// Security & Audit operations

func (r *postgresRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	query := `INSERT INTO audit_logs (user_id, action, resource, details, ip_address, user_agent) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, log.UserID, log.Action, log.Resource, log.Details, log.IPAddress, log.UserAgent)
	return err
}

func (r *postgresRepository) GetMFACode(ctx context.Context, userID uuid.UUID, code string) (*models.MFACode, error) {
	query := `SELECT id, user_id, code, method, expires_at FROM mfa_codes WHERE user_id = $1 AND code = $2 AND expires_at > NOW()`
	row := r.db.QueryRow(ctx, query, userID, code)
	var mfa models.MFACode
	err := row.Scan(&mfa.ID, &mfa.UserID, &mfa.Code, &mfa.Method, &mfa.ExpiresAt)
	if err != nil {
		return nil, nil
	}
	return &mfa, nil
}

func (r *postgresRepository) DeleteMFACode(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM mfa_codes WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Token operations
func (r *postgresRepository) SaveToken(ctx context.Context, token *models.AuthToken) error {
	query := `INSERT INTO auth_tokens (user_id, access_token, refresh_token, expires_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, token.UserID, token.AccessToken, token.RefreshToken, token.ExpiresAt)
	return err
}

func (r *postgresRepository) GetTokenByRefresh(ctx context.Context, refreshToken string) (*models.AuthToken, error) {
	query := `SELECT id, user_id, access_token, refresh_token, expires_at FROM auth_tokens WHERE refresh_token = $1`
	var t models.AuthToken
	err := r.db.QueryRow(ctx, query, refreshToken).Scan(&t.ID, &t.UserID, &t.AccessToken, &t.RefreshToken, &t.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *postgresRepository) DeleteToken(ctx context.Context, accessToken string) error {
	query := `DELETE FROM auth_tokens WHERE access_token = $1`
	_, err := r.db.Exec(ctx, query, accessToken)
	return err
}

func (r *postgresRepository) DeleteTokensByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM auth_tokens WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *postgresRepository) GetTokensByUserID(ctx context.Context, userID uuid.UUID) ([]models.AuthToken, error) {
	query := `SELECT id, user_id, access_token, refresh_token, expires_at FROM auth_tokens WHERE user_id = $1 AND expires_at > NOW()`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.AuthToken
	for rows.Next() {
		var t models.AuthToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.AccessToken, &t.RefreshToken, &t.ExpiresAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// Organization operations
func (r *postgresRepository) GetOrganizationsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Organization, error) {
	query := `
		SELECT o.id, o.name, COALESCE(o.domain, ''), o.status, o.created_at, o.updated_at
		FROM organizations o
		JOIN user_organizations uo ON o.id = uo.org_id
		WHERE uo.user_id = $1`
	
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []models.Organization
	for rows.Next() {
		var o models.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Domain, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, o)
	}
	return orgs, nil
}

func (r *postgresRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	query := `SELECT id, name, domain, status, created_at, updated_at FROM organizations WHERE id = $1`
	var o models.Organization
	err := r.db.QueryRow(ctx, query, id).Scan(&o.ID, &o.Name, &o.Domain, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *postgresRepository) AddUserToOrganization(ctx context.Context, userID, orgID uuid.UUID, role string) error {
	query := `INSERT INTO user_organizations (user_id, org_id, role) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, userID, orgID, role)
	return err
}

// Security & Audit operations
func (r *postgresRepository) LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error {
	query := `INSERT INTO security_events (user_id, event_type, ip_address, user_agent, metadata) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, event.UserID, event.EventType, event.IPAddress, event.UserAgent, event.Metadata)
	return err
}

func (r *postgresRepository) GetSecurityEvents(ctx context.Context, userID uuid.UUID) ([]models.SecurityEvent, error) {
	query := `SELECT id, user_id, event_type, ip_address, user_agent, metadata, created_at FROM security_events WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.SecurityEvent
	for rows.Next() {
		var e models.SecurityEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.EventType, &e.IPAddress, &e.UserAgent, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *postgresRepository) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `UPDATE users SET failed_logins = failed_logins + 1, updated_at = NOW() WHERE id = $1 RETURNING failed_logins`
	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *postgresRepository) ResetFailedLogins(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET failed_logins = 0, locked_until = NULL, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *postgresRepository) LockAccount(ctx context.Context, userID uuid.UUID, duration time.Duration) error {
	lockedUntil := time.Now().Add(duration)
	query := `UPDATE users SET locked_until = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, lockedUntil, userID)
	return err
}

// Admin operations
func (r *postgresRepository) ListOrganizations(ctx context.Context, page, size int) ([]models.Organization, int, error) {
	offset := (page - 1) * size
	
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM organizations WHERE is_deleted = FALSE`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get data
	query := `
		SELECT id, name, domain, default_theme, status, created_at, updated_at 
		FROM organizations 
		WHERE is_deleted = FALSE 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`
	
	rows, err := r.db.Query(ctx, query, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orgs []models.Organization
	for rows.Next() {
		var o models.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Domain, &o.DefaultTheme, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		orgs = append(orgs, o)
	}
	return orgs, total, nil
}

func (r *postgresRepository) ListUsers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.User, int, error) {
	offset := (page - 1) * size

	// Get total count (scoped to org)
	var total int
	countQuery := `
		SELECT COUNT(u.id) 
		FROM users u
		JOIN user_organizations uo ON u.id = uo.user_id
		WHERE uo.org_id = $1 AND u.is_deleted = FALSE`
	
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get data
	query := `
		SELECT u.id, u.email, u.name, u.phone, u.status, u.role_id, u.verified, u.theme, u.created_at, u.updated_at
		FROM users u
		JOIN user_organizations uo ON u.id = uo.user_id
		WHERE uo.org_id = $1 AND u.is_deleted = FALSE
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.Status, &u.RoleID, &u.Verified, &u.Theme, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *postgresRepository) SoftDeleteOrganization(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE organizations SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *postgresRepository) SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Devices
func (r *postgresRepository) ListDevices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Device, int, error) {
	offset := (page - 1) * size
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = FALSE`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `SELECT id, org_id, serial_number, type, status, last_ping_at, created_at, updated_at FROM devices WHERE org_id = $1 AND is_deleted = FALSE ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.OrgID, &d.SerialNumber, &d.Type, &d.Status, &d.LastPingAt, &d.CreatedAt, &d.UpdatedAt); err != nil { return nil, 0, err }
		devices = append(devices, d)
	}
	return devices, total, nil
}

func (r *postgresRepository) CreateDevice(ctx context.Context, d *models.Device) error {
	query := `INSERT INTO devices (id, org_id, serial_number, type, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, d.ID, d.OrgID, d.SerialNumber, d.Type, d.Status, d.CreatedAt, d.UpdatedAt)
	return err
}

func (r *postgresRepository) SoftDeleteDevice(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE devices SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1`, id)
	return err
}
// Security Suite Implementation

func (r *postgresRepository) GetMFADashboardSummary(ctx context.Context, orgID uuid.UUID) (*models.MFADashboardSummary, error) {
	var summary models.MFADashboardSummary
	
	// Complex aggregation query
	query := `
		SELECT 
			(SELECT COUNT(*) FROM users u JOIN user_organizations uo ON u.id = uo.user_id WHERE uo.org_id = $1) as total,
			(SELECT COUNT(*) FROM users u JOIN user_organizations uo ON u.id = uo.user_id WHERE uo.org_id = $1 AND u.mfa_enabled = TRUE) as enabled,
			(SELECT COUNT(*) FROM users u JOIN user_organizations uo ON u.id = uo.user_id WHERE uo.org_id = $1 AND u.mfa_enabled = FALSE AND u.status = 'active') as pending,
			(SELECT COUNT(*) FROM users u JOIN user_organizations uo ON u.id = uo.user_id WHERE uo.org_id = $1 AND u.mfa_enabled = FALSE) as disabled,
			(SELECT COUNT(*) FROM mfa_failures WHERE org_id = $1 AND created_at > NOW() - INTERVAL '24 hours') as failures,
			(SELECT COUNT(*) FROM users u JOIN user_organizations uo ON u.id = uo.user_id WHERE uo.org_id = $1 AND u.locked_until > NOW()) as locked
	`
	
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&summary.TotalUsers, &summary.MFAEnabled, &summary.MFAPending, 
		&summary.MFADisabled, &summary.FailedAttempts24h, &summary.LockedAccounts,
	)
	return &summary, err
}

func (r *postgresRepository) GetMFAAdoptionTrend(ctx context.Context, orgID uuid.UUID, days int) ([]map[string]any, error) {
	query := `
		SELECT CAST(created_at AS DATE) as date, COUNT(*) as enabled
		FROM users u
		JOIN user_organizations uo ON u.id = uo.user_id
		WHERE uo.org_id = $1 AND u.mfa_enabled = TRUE AND u.created_at > NOW() - $2 * INTERVAL '1 day'
		GROUP BY date
		ORDER BY date ASC
	`
	rows, err := r.db.Query(ctx, query, orgID, days)
	if err != nil { return nil, err }
	defer rows.Close()

	var trend []map[string]any
	for rows.Next() {
		var date time.Time
		var count int
		if err := rows.Scan(&date, &count); err != nil { return nil, err }
		trend = append(trend, map[string]any{
			"date": date.Format("2006-01-02"),
			"enabled": count,
		})
	}
	return trend, nil
}

func (r *postgresRepository) GetMFAMethodDistribution(ctx context.Context, orgID uuid.UUID) (map[string]int, error) {
	query := `
		SELECT method, COUNT(*) 
		FROM mfa_events 
		WHERE org_id = $1 AND event_type = 'VERIFY_SUCCESS'
		GROUP BY method
	`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil { return nil, err }
	defer rows.Close()

	dist := make(map[string]int)
	for rows.Next() {
		var method string
		var count int
		if err := rows.Scan(&method, &count); err != nil { return nil, err }
		dist[method] = count
	}
	return dist, nil
}

func (r *postgresRepository) GetMFAEvents(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.MFAEvent, int, error) {
	offset := (page - 1) * size
	query := `SELECT id, user_id, org_id, event_type, status, method, ip_address, user_agent, location, details, created_at 
			  FROM mfa_events WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()

	var events []models.MFAEvent
	for rows.Next() {
		var e models.MFAEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.OrgID, &e.EventType, &e.Status, &e.Method, &e.IPAddress, &e.UserAgent, &e.Location, &e.Details, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}
	
	var total int
	r.db.QueryRow(ctx, "SELECT COUNT(*) FROM mfa_events WHERE org_id = $1", orgID).Scan(&total)
	return events, total, nil
}

func (r *postgresRepository) GetTopFailureLocations(ctx context.Context, orgID uuid.UUID) ([]map[string]any, error) {
	query := `SELECT location, COUNT(*) as count FROM mfa_failures WHERE org_id = $1 GROUP BY location ORDER BY count DESC LIMIT 5`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil { return nil, err }
	defer rows.Close()

	var locs []map[string]any
	for rows.Next() {
		var loc string
		var count int
		rows.Scan(&loc, &count)
		locs = append(locs, map[string]any{"location": loc, "count": count})
	}
	return locs, nil
}

func (r *postgresRepository) GetTopFailureUsers(ctx context.Context, orgID uuid.UUID) ([]map[string]any, error) {
	query := `SELECT u.email, COUNT(f.id) as count 
			  FROM mfa_failures f JOIN users u ON f.user_id = u.id 
			  WHERE f.org_id = $1 GROUP BY u.email ORDER BY count DESC LIMIT 5`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil { return nil, err }
	defer rows.Close()

	var users []map[string]any
	for rows.Next() {
		var email string
		var count int
		rows.Scan(&email, &count)
		users = append(users, map[string]any{"user": email, "count": count})
	}
	return users, nil
}

func (r *postgresRepository) GetLockedUsers(ctx context.Context, orgID uuid.UUID) ([]models.User, error) {
	query := `SELECT u.id, u.email, u.name, u.status, u.locked_until 
			  FROM users u JOIN user_organizations uo ON u.id = uo.user_id 
			  WHERE uo.org_id = $1 AND u.locked_until > NOW()`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil { return nil, err }
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Email, &u.Name, &u.Status, &u.LockedUntil)
		users = append(users, u)
	}
	return users, nil
}

func (r *postgresRepository) UnlockUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET locked_until = NULL, failed_logins = 0 WHERE id = $1", userID)
	return err
}

func (r *postgresRepository) ResetMFA(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "UPDATE users SET mfa_enabled = FALSE, mfa_secret = NULL WHERE id = $1", userID)
	return err
}

func (r *postgresRepository) GetMFAPolicy(ctx context.Context, orgID uuid.UUID) (*models.MFAPolicy, error) {
	var p models.MFAPolicy
	query := `SELECT id, org_id, require_mfa, allowed_methods, max_attempts, lockout_minutes FROM mfa_policies WHERE org_id = $1`
	err := r.db.QueryRow(ctx, query, orgID).Scan(&p.ID, &p.OrgID, &p.RequireMFA, &p.AllowedMethods, &p.MaxAttempts, &p.LockoutMinutes)
	if err != nil { return nil, err }
	return &p, nil
}

func (r *postgresRepository) UpdateMFAPolicy(ctx context.Context, p *models.MFAPolicy) error {
	query := `INSERT INTO mfa_policies (org_id, require_mfa, allowed_methods, max_attempts, lockout_minutes) 
			  VALUES ($1, $2, $3, $4, $5) 
			  ON CONFLICT (org_id) DO UPDATE SET require_mfa = $2, allowed_methods = $3, max_attempts = $4, lockout_minutes = $5, updated_at = NOW()`
	_, err := r.db.Exec(ctx, query, p.OrgID, p.RequireMFA, p.AllowedMethods, p.MaxAttempts, p.LockoutMinutes)
	return err
}

func (r *postgresRepository) AddMFAException(ctx context.Context, exc *models.MFAException) error {
	query := `INSERT INTO mfa_exceptions (user_id, org_id, reason, granted_by, expires_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, exc.UserID, exc.OrgID, exc.Reason, exc.GrantedBy, exc.ExpiresAt)
	return err
}

func (r *postgresRepository) GetMFAExceptions(ctx context.Context, orgID uuid.UUID) ([]models.MFAException, error) {
	query := `SELECT id, user_id, org_id, reason, granted_by, expires_at, created_at FROM mfa_exceptions WHERE org_id = $1`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil { return nil, err }
	defer rows.Close()

	var excs []models.MFAException
	for rows.Next() {
		var e models.MFAException
		rows.Scan(&e.ID, &e.UserID, &e.OrgID, &e.Reason, &e.GrantedBy, &e.ExpiresAt, &e.CreatedAt)
		excs = append(excs, e)
	}
	return excs, nil
}

func (r *postgresRepository) DeleteMFAException(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM mfa_exceptions WHERE id = $1", id)
	return err
}

func (r *postgresRepository) LogMFAEvent(ctx context.Context, e *models.MFAEvent) error {
	query := `INSERT INTO mfa_events (user_id, org_id, event_type, status, method, ip_address, user_agent, location, details) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query, e.UserID, e.OrgID, e.EventType, e.Status, e.Method, e.IPAddress, e.UserAgent, e.Location, e.Details)
	return err
}

func (r *postgresRepository) LogMFAFailure(ctx context.Context, f *models.MFAFailure) error {
	query := `INSERT INTO mfa_failures (user_id, org_id, method, failure_reason, ip_address, location) 
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, f.UserID, f.OrgID, f.Method, f.FailureReason, f.IPAddress, f.Location)
	return err
}

func (r *postgresRepository) CreateMFALockout(ctx context.Context, l *models.MFALockout) error {
	query := `INSERT INTO mfa_lockouts (user_id, org_id, reason, expires_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, l.UserID, l.OrgID, l.Reason, l.ExpiresAt)
	return err
}
