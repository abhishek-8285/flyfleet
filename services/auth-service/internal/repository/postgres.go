package repository

import (
	"context"
	"encoding/json"
	"fmt"
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
	query := `INSERT INTO audit_logs (org_id, user_id, module, action, resource, entity_id, severity, status, details, before_state, after_state, metadata, ip_address, user_agent) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := r.db.Exec(ctx, query,
		log.OrgID, log.UserID, log.Module, log.Action, log.Resource, log.EntityID,
		log.Severity, log.Status, log.Details, log.BeforeState, log.AfterState,
		log.Metadata, log.IPAddress, log.UserAgent)
	return err
}

func (r *postgresRepository) GetAuditLogByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	query := `SELECT id, org_id, user_id, action, resource, severity, status, details, ip_address, user_agent, created_at FROM audit_logs WHERE id = $1`
	var log models.AuditLog
	err := r.db.QueryRow(ctx, query, id).Scan(
		&log.ID, &log.OrgID, &log.UserID, &log.Action, &log.Resource, &log.Severity, &log.Status, &log.Details, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *postgresRepository) ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.AuditLog, int, error) {
	offset := (page - 1) * size

	// Base query
	query := `SELECT id, org_id, user_id, module, action, resource, entity_id, severity, status, details, before_state, after_state, metadata, ip_address, user_agent, created_at 
			  FROM audit_logs WHERE org_id = $1`
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE org_id = $1`

	args := []any{orgID}
	placeholder := 2

	if module, ok := filters["module"].(string); ok && module != "" {
		query += fmt.Sprintf(" AND module = $%d", placeholder)
		countQuery += fmt.Sprintf(" AND module = $%d", placeholder)
		args = append(args, module)
		placeholder++
	}

	if action, ok := filters["action"].(string); ok && action != "" {
		query += fmt.Sprintf(" AND action = $%d", placeholder)
		countQuery += fmt.Sprintf(" AND action = $%d", placeholder)
		args = append(args, action)
		placeholder++
	}

	if severity, ok := filters["severity"].(string); ok && severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", placeholder)
		countQuery += fmt.Sprintf(" AND severity = $%d", placeholder)
		args = append(args, severity)
		placeholder++
	}

	if userID, ok := filters["user_id"].(uuid.UUID); ok && userID != uuid.Nil {
		query += fmt.Sprintf(" AND user_id = $%d", placeholder)
		countQuery += fmt.Sprintf(" AND user_id = $%d", placeholder)
		args = append(args, userID)
		placeholder++
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", placeholder, placeholder+1)

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, size, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		err := rows.Scan(
			&l.ID, &l.OrgID, &l.UserID, &l.Module, &l.Action, &l.Resource, &l.EntityID,
			&l.Severity, &l.Status, &l.Details, &l.BeforeState, &l.AfterState,
			&l.Metadata, &l.IPAddress, &l.UserAgent, &l.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	return logs, total, nil
}

func (r *postgresRepository) GetAuditSettings(ctx context.Context, orgID uuid.UUID) (*models.AuditSettings, error) {
	query := `SELECT id, org_id, retention_days_online, retention_days_archival, masking_rules, tracked_event_types, alert_on_critical, updated_at 
			  FROM audit_settings WHERE org_id = $1`
	var s models.AuditSettings
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&s.ID, &s.OrgID, &s.RetentionDaysOnline, &s.RetentionDaysArchival, &s.MaskingRules, &s.TrackedEventTypes, &s.AlertOnCritical, &s.UpdatedAt,
	)
	if err != nil {
		// If not found, return default
		return &models.AuditSettings{
			OrgID: orgID,
			RetentionDaysOnline: 90,
			RetentionDaysArchival: 2555,
			MaskingRules: []string{"password", "secret", "token", "phone", "email_local"},
			TrackedEventTypes: []string{"MFA_POLICY_UPDATE", "MFA_EXCEPTION_GRANT", "MFA_OVERRIDE_TOGGLE"},
		}, nil
	}
	return &s, nil
}

func (r *postgresRepository) UpdateAuditSettings(ctx context.Context, s *models.AuditSettings) error {
	query := `INSERT INTO audit_settings (org_id, retention_days_online, retention_days_archival, masking_rules, tracked_event_types, alert_on_critical, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, NOW())
			  ON CONFLICT (org_id) DO UPDATE SET
			  retention_days_online = $2, retention_days_archival = $3, masking_rules = $4, tracked_event_types = $5, alert_on_critical = $6, updated_at = NOW()`
	_, err := r.db.Exec(ctx, query, s.OrgID, s.RetentionDaysOnline, s.RetentionDaysArchival, s.MaskingRules, s.TrackedEventTypes, s.AlertOnCritical)
	return err
}

func (r *postgresRepository) GetUserRoleInOrg(ctx context.Context, userID, orgID uuid.UUID) (string, error) {
	query := `SELECT role FROM user_organizations WHERE user_id = $1 AND org_id = $2`
	var role string
	err := r.db.QueryRow(ctx, query, userID, orgID).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
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
		SELECT o.id, o.name, COALESCE(o.domain, ''), o.default_theme, o.status, o.created_at, o.updated_at
		FROM organizations o
		JOIN user_organizations uo ON o.id = uo.org_id
		WHERE uo.user_id = $1 AND o.is_deleted = FALSE`
	
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []models.Organization
	for rows.Next() {
		var o models.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Domain, &o.DefaultTheme, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, o)
	}
	return orgs, nil
}

func (r *postgresRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	query := `SELECT id, name, domain, default_theme, status, created_at, updated_at FROM organizations WHERE id = $1 AND is_deleted = FALSE`
	var o models.Organization
	err := r.db.QueryRow(ctx, query, id).Scan(&o.ID, &o.Name, &o.Domain, &o.DefaultTheme, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *postgresRepository) CreateOrganization(ctx context.Context, org *models.Organization) error {
	query := `INSERT INTO organizations (id, name, domain, default_theme, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, org.ID, org.Name, org.Domain, org.DefaultTheme, org.Status, time.Now(), time.Now())
	return err
}

func (r *postgresRepository) UpdateOrganization(ctx context.Context, org *models.Organization) error {
	query := `UPDATE organizations SET name=$1, domain=$2, default_theme=$3, status=$4, updated_at=NOW() WHERE id=$5`
	_, err := r.db.Exec(ctx, query, org.Name, org.Domain, org.DefaultTheme, org.Status, org.ID)
	return err
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

// Vehicles
func (r *postgresRepository) ListVehicles(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Vehicle, int, error) {
	offset := (page - 1) * size
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM vehicles WHERE org_id = $1 AND is_deleted = FALSE`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `SELECT id, org_id, current_driver_id, name, license_plate, make, model, year, fuel_type, vehicle_no, vin, odometer, service_due, notes, status, created_at, updated_at 
			  FROM vehicles WHERE org_id = $1 AND is_deleted = FALSE ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()

	var vehicles []models.Vehicle
	for rows.Next() {
		var v models.Vehicle
		if err := rows.Scan(&v.ID, &v.OrgID, &v.CurrentDriverID, &v.Name, &v.LicensePlate, &v.Make, &v.Model, &v.Year, &v.FuelType, &v.VehicleNo, &v.VIN, &v.Odometer, &v.ServiceDue, &v.Notes, &v.Status, &v.CreatedAt, &v.UpdatedAt); err != nil { return nil, 0, err }
		vehicles = append(vehicles, v)
	}
	return vehicles, total, nil
}

func (r *postgresRepository) CreateVehicle(ctx context.Context, v *models.Vehicle) error {
	query := `INSERT INTO vehicles (id, org_id, current_driver_id, name, license_plate, make, model, year, fuel_type, vehicle_no, vin, odometer, service_due, notes, status, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`
	_, err := r.db.Exec(ctx, query, v.ID, v.OrgID, v.CurrentDriverID, v.Name, v.LicensePlate, v.Make, v.Model, v.Year, v.FuelType, v.VehicleNo, v.VIN, v.Odometer, v.ServiceDue, v.Notes, v.Status, v.CreatedAt, v.UpdatedAt)
	return err
}

func (r *postgresRepository) GetVehicleByID(ctx context.Context, id uuid.UUID) (*models.Vehicle, error) {
	var v models.Vehicle
	query := `SELECT id, org_id, current_driver_id, name, license_plate, make, model, year, fuel_type, vehicle_no, vin, odometer, service_due, notes, status, created_at, updated_at 
			  FROM vehicles WHERE id = $1 AND is_deleted = FALSE`
	err := r.db.QueryRow(ctx, query, id).Scan(&v.ID, &v.OrgID, &v.CurrentDriverID, &v.Name, &v.LicensePlate, &v.Make, &v.Model, &v.Year, &v.FuelType, &v.VehicleNo, &v.VIN, &v.Odometer, &v.ServiceDue, &v.Notes, &v.Status, &v.CreatedAt, &v.UpdatedAt)
	if err != nil { return nil, err }
	return &v, nil
}

func (r *postgresRepository) UpdateVehicle(ctx context.Context, v *models.Vehicle) error {
	query := `UPDATE vehicles SET name=$1, license_plate=$2, make=$3, model=$4, year=$5, fuel_type=$6, vehicle_no=$7, vin=$8, odometer=$9, service_due=$10, notes=$11, status=$12, updated_at=NOW() WHERE id=$13`
	_, err := r.db.Exec(ctx, query, v.Name, v.LicensePlate, v.Make, v.Model, v.Year, v.FuelType, v.VehicleNo, v.VIN, v.Odometer, v.ServiceDue, v.Notes, v.Status, v.ID)
	return err
}

func (r *postgresRepository) DeleteVehicle(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE vehicles SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *postgresRepository) UpdateVehicleStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE vehicles SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

// Devices
func (r *postgresRepository) ListDevices(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.Device, int, error) {
	offset := (page - 1) * size
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE org_id = $1 AND is_deleted = FALSE`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `SELECT id, org_id, vehicle_id, serial_number, type, status, last_ping_at, created_at, updated_at FROM devices WHERE org_id = $1 AND is_deleted = FALSE ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil { return nil, 0, err }
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.OrgID, &d.VehicleID, &d.SerialNumber, &d.Type, &d.Status, &d.LastPingAt, &d.CreatedAt, &d.UpdatedAt); err != nil { return nil, 0, err }
		devices = append(devices, d)
	}
	return devices, total, nil
}

func (r *postgresRepository) GetDeviceByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	var d models.Device
	query := `SELECT id, org_id, vehicle_id, serial_number, type, status, last_ping_at, created_at, updated_at FROM devices WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&d.ID, &d.OrgID, &d.VehicleID, &d.SerialNumber, &d.Type, &d.Status, &d.LastPingAt, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *postgresRepository) CreateDevice(ctx context.Context, d *models.Device) error {
	query := `INSERT INTO devices (id, org_id, vehicle_id, serial_number, type, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, d.ID, d.OrgID, d.VehicleID, d.SerialNumber, d.Type, d.Status, d.CreatedAt, d.UpdatedAt)
	return err
}

func (r *postgresRepository) AssignDeviceToVehicle(ctx context.Context, deviceID, vehicleID uuid.UUID) error {
	// Transaction to ensure 1:1 mapping (optional but safe)
	tx, err := r.db.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	// Ensure device is not already assigned (or overwrite based on business logic)
	// Here we overwrite:
	if _, err := tx.Exec(ctx, `UPDATE devices SET vehicle_id = $1, updated_at = NOW() WHERE id = $2`, vehicleID, deviceID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *postgresRepository) UnassignDevice(ctx context.Context, deviceID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE devices SET vehicle_id = NULL, updated_at = NOW() WHERE id = $1`, deviceID)
	return err
}

func (r *postgresRepository) AssignDriverToVehicle(ctx context.Context, driverID, vehicleID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE vehicles SET current_driver_id = $1, updated_at = NOW() WHERE id = $2`, driverID, vehicleID)
	return err
}

func (r *postgresRepository) SoftDeleteDevice(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE devices SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1`, id)
	return err
}

// Documents
func (r *postgresRepository) CreateDocument(ctx context.Context, d *models.Document) error {
	query := `INSERT INTO documents (id, vehicle_id, type, doc_number, issue_date, expiry_date, file_url, status, metadata, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.Exec(ctx, query, d.ID, d.VehicleID, d.Type, d.DocNumber, d.IssueDate, d.ExpiryDate, d.FileURL, d.Status, d.Metadata, d.CreatedAt, d.UpdatedAt)
	return err
}

func (r *postgresRepository) GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	var d models.Document
	query := `SELECT id, vehicle_id, type, doc_number, issue_date, expiry_date, file_url, status, metadata, created_at, updated_at 
			  FROM documents WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&d.ID, &d.VehicleID, &d.Type, &d.DocNumber, &d.IssueDate, &d.ExpiryDate, &d.FileURL, &d.Status, &d.Metadata, &d.CreatedAt, &d.UpdatedAt)
	if err != nil { return nil, err }
	return &d, nil
}

func (r *postgresRepository) ListDocumentsByVehicleID(ctx context.Context, vehicleID uuid.UUID) ([]models.Document, error) {
	query := `SELECT id, vehicle_id, type, doc_number, issue_date, expiry_date, file_url, status, metadata, created_at, updated_at 
			  FROM documents WHERE vehicle_id = $1 ORDER BY expiry_date ASC`
	rows, err := r.db.Query(ctx, query, vehicleID)
	if err != nil { return nil, err }
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var d models.Document
		if err := rows.Scan(&d.ID, &d.VehicleID, &d.Type, &d.DocNumber, &d.IssueDate, &d.ExpiryDate, &d.FileURL, &d.Status, &d.Metadata, &d.CreatedAt, &d.UpdatedAt); err != nil { return nil, err }
		docs = append(docs, d)
	}
	return docs, nil
}

func (r *postgresRepository) UpdateDocument(ctx context.Context, d *models.Document) error {
	query := `UPDATE documents SET type=$1, doc_number=$2, issue_date=$3, expiry_date=$4, file_url=$5, status=$6, metadata=$7, updated_at=NOW() WHERE id=$8`
	_, err := r.db.Exec(ctx, query, d.Type, d.DocNumber, d.IssueDate, d.ExpiryDate, d.FileURL, d.Status, d.Metadata, d.ID)
	return err
}

func (r *postgresRepository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM documents WHERE id = $1`, id)
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
	query := `SELECT id, org_id, require_mfa, allowed_methods, max_attempts, lockout_minutes, 
			  grace_period_days, remember_device_days, enforce_at_login, admin_only_mode, 
			  remember_device_allowed, mfa_challenge_timeout, version 
			  FROM mfa_policies WHERE org_id = $1`
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&p.ID, &p.OrgID, &p.RequireMFA, &p.AllowedMethods, &p.MaxAttempts, &p.LockoutMinutes,
		&p.GracePeriodDays, &p.RememberDeviceDays, &p.EnforceAtLogin, &p.AdminOnlyMode,
		&p.RememberDeviceAllowed, &p.MFAChallengeTimeout, &p.Version,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *postgresRepository) UpdateMFAPolicy(ctx context.Context, p *models.MFAPolicy, updaterID uuid.UUID, reason string) error {
	// Optimistic concurrency check if version > 0
	if p.Version > 0 {
		query := `UPDATE mfa_policies SET 
				  require_mfa = $1, allowed_methods = $2, max_attempts = $3, lockout_minutes = $4,
				  grace_period_days = $5, remember_device_days = $6, enforce_at_login = $7, 
				  admin_only_mode = $8, remember_device_allowed = $9, mfa_challenge_timeout = $10,
				  version = version + 1, updated_at = NOW()
				  WHERE org_id = $11 AND version = $12`
		res, err := r.db.Exec(ctx, query, 
			p.RequireMFA, p.AllowedMethods, p.MaxAttempts, p.LockoutMinutes,
			p.GracePeriodDays, p.RememberDeviceDays, p.EnforceAtLogin, 
			p.AdminOnlyMode, p.RememberDeviceAllowed, p.MFAChallengeTimeout,
			p.OrgID, p.Version,
		)
		if err != nil { return err }
		if res.RowsAffected() == 0 {
			return fmt.Errorf("policy update conflict: version mismatch")
		}
	} else {
		query := `INSERT INTO mfa_policies (
				  org_id, require_mfa, allowed_methods, max_attempts, lockout_minutes,
				  grace_period_days, remember_device_days, enforce_at_login, 
				  admin_only_mode, remember_device_allowed, mfa_challenge_timeout, version
				  ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 1)
				  ON CONFLICT (org_id) DO UPDATE SET 
				  require_mfa = $2, allowed_methods = $3, max_attempts = $4, lockout_minutes = $5,
				  grace_period_days = $6, remember_device_days = $7, enforce_at_login = $8, 
				  admin_only_mode = $9, remember_device_allowed = $10, mfa_challenge_timeout = $11,
				  version = mfa_policies.version + 1, updated_at = NOW()`
		_, err := r.db.Exec(ctx, query, 
			p.OrgID, p.RequireMFA, p.AllowedMethods, p.MaxAttempts, p.LockoutMinutes,
			p.GracePeriodDays, p.RememberDeviceDays, p.EnforceAtLogin, 
			p.AdminOnlyMode, p.RememberDeviceAllowed, p.MFAChallengeTimeout,
		)
		if err != nil { return err }
	}

	// Create a history snapshot
	snapshot, _ := json.Marshal(p)
	historyQuery := `INSERT INTO mfa_policy_history (org_id, snapshot, changed_by, change_reason, version) 
					VALUES ($1, $2, $3, $4, (SELECT version FROM mfa_policies WHERE org_id = $1))`
	
	var uid *uuid.UUID
	if updaterID != uuid.Nil { uid = &updaterID }
	_, _ = r.db.Exec(ctx, historyQuery, p.OrgID, string(snapshot), uid, reason)

	return nil
}

func (r *postgresRepository) GetMFAPolicyHistory(ctx context.Context, orgID uuid.UUID) ([]models.MFAPolicyHistory, error) {
	query := `SELECT id, org_id, snapshot, changed_by, version, created_at 
			  FROM mfa_policy_history WHERE org_id = $1 ORDER BY version DESC`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil { return nil, err }
	defer rows.Close()

	var history []models.MFAPolicyHistory
	for rows.Next() {
		var h models.MFAPolicyHistory
		var changedBy *uuid.UUID
		err := rows.Scan(&h.ID, &h.OrgID, &h.Snapshot, &changedBy, &h.Version, &h.CreatedAt)
		if err != nil { return nil, err }
		if changedBy != nil { h.ChangedBy = *changedBy }
		history = append(history, h)
	}
	return history, nil
}

func (r *postgresRepository) RollbackMFAPolicy(ctx context.Context, historyID uuid.UUID) error {
	var h models.MFAPolicyHistory
	err := r.db.QueryRow(ctx, "SELECT org_id, snapshot FROM mfa_policy_history WHERE id = $1", historyID).Scan(&h.OrgID, &h.Snapshot)
	if err != nil { return err }

	var p models.MFAPolicy
	if err := json.Unmarshal([]byte(h.Snapshot), &p); err != nil {
		return err
	}

	// Update the policy using the snapshot, but keep incrementing version
	query := `UPDATE mfa_policies SET 
			  require_mfa = $1, allowed_methods = $2, max_attempts = $3, lockout_minutes = $4,
			  grace_period_days = $5, remember_device_days = $6, enforce_at_login = $7, 
			  admin_only_mode = $8, remember_device_allowed = $9, mfa_challenge_timeout = $10,
			  version = version + 1, updated_at = NOW()
			  WHERE org_id = $11`
	_, err = r.db.Exec(ctx, query, 
		p.RequireMFA, p.AllowedMethods, p.MaxAttempts, p.LockoutMinutes,
		p.GracePeriodDays, p.RememberDeviceDays, p.EnforceAtLogin, 
		p.AdminOnlyMode, p.RememberDeviceAllowed, p.MFAChallengeTimeout,
		h.OrgID,
	)
	return err
}

func (r *postgresRepository) GetMFAOrgOverride(ctx context.Context, orgID uuid.UUID) (*models.MFAOrgOverride, error) {
	var o models.MFAOrgOverride
	query := `SELECT id, org_id, enabled, allowed_methods, max_attempts, lockout_minutes, 
			  remember_device_days, grace_period_days, created_at, updated_at 
			  FROM mfa_org_overrides WHERE org_id = $1`
	err := r.db.QueryRow(ctx, query, orgID).Scan(
		&o.ID, &o.OrgID, &o.Enabled, &o.AllowedMethods, &o.MaxAttempts, &o.LockoutMinutes,
		&o.RememberDeviceDays, &o.GracePeriodDays, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *postgresRepository) UpdateMFAOrgOverride(ctx context.Context, o *models.MFAOrgOverride) error {
	query := `INSERT INTO mfa_org_overrides (
			  org_id, enabled, allowed_methods, max_attempts, lockout_minutes, 
			  remember_device_days, grace_period_days
			  ) VALUES ($1, $2, $3, $4, $5, $6, $7)
			  ON CONFLICT (org_id) DO UPDATE SET 
			  enabled = $2, allowed_methods = $3, max_attempts = $4, lockout_minutes = $5,
			  remember_device_days = $6, grace_period_days = $7, updated_at = NOW()`
	_, err := r.db.Exec(ctx, query, 
		o.OrgID, o.Enabled, o.AllowedMethods, o.MaxAttempts, o.LockoutMinutes,
		o.RememberDeviceDays, o.GracePeriodDays,
	)
	return err
}

func (r *postgresRepository) AddMFAException(ctx context.Context, exc *models.MFAException) error {
	query := `INSERT INTO mfa_exceptions (user_id, org_id, reason, granted_by, expires_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, exc.UserID, exc.OrgID, exc.Reason, exc.GrantedBy, exc.ExpiresAt)
	return err
}

func (r *postgresRepository) GetMFAExceptionByID(ctx context.Context, id uuid.UUID) (*models.MFAException, error) {
	var e models.MFAException
	query := `SELECT id, user_id, org_id, reason, granted_by, expires_at, created_at FROM mfa_exceptions WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&e.ID, &e.UserID, &e.OrgID, &e.Reason, &e.GrantedBy, &e.ExpiresAt, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
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

// RBAC Operations

func (r *postgresRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	query := `SELECT id, name, description, organization_id, is_system, created_at, updated_at FROM roles WHERE id = $1`
	var role models.Role
	err := r.db.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.OrganizationID, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *postgresRepository) CreateRole(ctx context.Context, role *models.Role) error {
	query := `INSERT INTO roles (id, name, description, organization_id, is_system, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, role.ID, role.Name, role.Description, role.OrganizationID, role.IsSystem, time.Now(), time.Now())
	return err
}

func (r *postgresRepository) UpdateRole(ctx context.Context, role *models.Role) error {
	query := `UPDATE roles SET name = $1, description = $2, updated_at = NOW() WHERE id = $3 AND is_system = FALSE`
	_, err := r.db.Exec(ctx, query, role.Name, role.Description, role.ID)
	return err
}

func (r *postgresRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1 AND is_system = FALSE`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *postgresRepository) ListRolesByOrg(ctx context.Context, orgID uuid.UUID) ([]models.Role, error) {
	query := `SELECT id, name, description, organization_id, is_system, created_at, updated_at FROM roles WHERE organization_id = $1 OR is_system = TRUE ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.OrganizationID, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *postgresRepository) ListPermissions(ctx context.Context) ([]models.PermissionDefinition, error) {
	query := `SELECT id, module, action, description FROM permissions ORDER BY module, action ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.PermissionDefinition
	for rows.Next() {
		var perm models.PermissionDefinition
		err := rows.Scan(&perm.ID, &perm.Module, &perm.Action, &perm.Description)
		if err != nil {
			return nil, err
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (r *postgresRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]models.PermissionDefinition, error) {
	query := `
		SELECT p.id, p.module, p.action, p.description
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.module, p.action ASC
	`
	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []models.PermissionDefinition
	for rows.Next() {
		var perm models.PermissionDefinition
		err := rows.Scan(&perm.ID, &perm.Module, &perm.Action, &perm.Description)
		if err != nil {
			return nil, err
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (r *postgresRepository) UpdateRolePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Clear existing permissions
	if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
		return err
	}

	// Insert new permissions
	for _, permID := range permissionIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`, roleID, permID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *postgresRepository) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	query := `INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at) VALUES ($1, $2, $3, NOW()) ON CONFLICT (user_id, role_id) DO NOTHING`
	_, err := r.db.Exec(ctx, query, userID, roleID, assignedBy)
	return err
}

func (r *postgresRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	_, err := r.db.Exec(ctx, query, userID, roleID)
	return err
}

func (r *postgresRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]models.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.organization_id, r.is_system, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.name ASC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.OrganizationID, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}
func (r *postgresRepository) ListEntityAuditHistory(ctx context.Context, entityType, entityID string) ([]models.EntityAuditHistory, error) {
	query := `SELECT id, entity_type, entity_id, change_data, actor, timestamp FROM entity_audit_history 
			  WHERE entity_type = $1 AND entity_id = $2 ORDER BY timestamp DESC`
	rows, err := r.db.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.EntityAuditHistory
	for rows.Next() {
		var h models.EntityAuditHistory
		if err := rows.Scan(&h.ID, &h.EntityType, &h.EntityID, &h.ChangeData, &h.Actor, &h.Timestamp); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func (r *postgresRepository) CreateEvidenceExport(ctx context.Context, export *models.EvidenceExport) error {
	query := `INSERT INTO evidence_exports (org_id, module, date_range_start, date_range_end, format, status, requested_by) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	return r.db.QueryRow(ctx, query, 
		export.OrgID, export.Module, export.DateRangeStart, export.DateRangeEnd, 
		export.Format, export.Status, export.RequestedBy).Scan(&export.ID)
}

func (r *postgresRepository) ListEvidenceExports(ctx context.Context, orgID uuid.UUID) ([]models.EvidenceExport, error) {
	query := `SELECT id, org_id, module, date_range_start, date_range_end, format, status, requested_by, COALESCE(file_url, ''), COALESCE(error_message, ''), created_at 
			  FROM evidence_exports WHERE org_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exports []models.EvidenceExport
	for rows.Next() {
		var e models.EvidenceExport
		if err := rows.Scan(&e.ID, &e.OrgID, &e.Module, &e.DateRangeStart, &e.DateRangeEnd, &e.Format, &e.Status, &e.RequestedBy, &e.FileURL, &e.ErrorMessage, &e.CreatedAt); err != nil {
			return nil, err
		}
		exports = append(exports, e)
	}
	return exports, nil
}

func (r *postgresRepository) GetEvidenceExportByID(ctx context.Context, id uuid.UUID) (*models.EvidenceExport, error) {
	query := `SELECT id, org_id, module, date_range_start, date_range_end, format, status, requested_by, COALESCE(file_url, ''), COALESCE(error_message, ''), created_at 
			  FROM evidence_exports WHERE id = $1`
	var e models.EvidenceExport
	err := r.db.QueryRow(ctx, query, id).Scan(&e.ID, &e.OrgID, &e.Module, &e.DateRangeStart, &e.DateRangeEnd, &e.Format, &e.Status, &e.RequestedBy, &e.FileURL, &e.ErrorMessage, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *postgresRepository) UpdateEvidenceExportStatus(ctx context.Context, id uuid.UUID, status, fileURL, errorMsg string) error {
	query := `UPDATE evidence_exports SET status = $1, file_url = $2, error_message = $3 WHERE id = $4`
	_, err := r.db.Exec(ctx, query, status, fileURL, errorMsg, id)
	return err
}

func (r *postgresRepository) GetAuditPolicy(ctx context.Context, orgID uuid.UUID) (*models.AuditPolicy, error) {
	query := `SELECT id, org_id, retention_years, archive_policy, masking_rules, updated_by, updated_at 
			  FROM audit_policies WHERE org_id = $1`
	var p models.AuditPolicy
	err := r.db.QueryRow(ctx, query, orgID).Scan(&p.ID, &p.OrgID, &p.RetentionYears, &p.ArchivePolicy, &p.MaskingRules, &p.UpdatedBy, &p.UpdatedAt)
	if err != nil {
		// Return defaults if not found
		return &models.AuditPolicy{
			OrgID: orgID,
			RetentionYears: 7,
			ArchivePolicy: "MOVE_TO_COLD_STORAGE_1Y",
			MaskingRules: []string{"credit_card", "ssn", "phone", "email"},
		}, nil
	}
	return &p, nil
}

func (r *postgresRepository) UpdateAuditPolicy(ctx context.Context, p *models.AuditPolicy) error {
	query := `INSERT INTO audit_policies (org_id, retention_years, archive_policy, masking_rules, updated_by, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, NOW()) 
			  ON CONFLICT (org_id) DO UPDATE SET 
			  retention_years = $2, archive_policy = $3, masking_rules = $4, updated_by = $5, updated_at = NOW()`
	_, err := r.db.Exec(ctx, query, p.OrgID, p.RetentionYears, p.ArchivePolicy, p.MaskingRules, p.UpdatedBy)
	return err
}
