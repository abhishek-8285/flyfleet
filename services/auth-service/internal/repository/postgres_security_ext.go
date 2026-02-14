package repository

import (
	"context"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
)

// IP Rules

func (r *postgresRepository) ListIPRules(ctx context.Context, orgID uuid.UUID) ([]models.IPRule, error) {
	var query string
	var args []interface{}

	if orgID == uuid.Nil {
		query = `SELECT id, org_id, cidr, action, description, added_by, created_at, updated_at 
				  FROM ip_rules ORDER BY created_at DESC`
	} else {
		query = `SELECT id, org_id, cidr, action, description, added_by, created_at, updated_at 
				  FROM ip_rules WHERE org_id = $1 ORDER BY created_at DESC`
		args = append(args, orgID)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.IPRule
	for rows.Next() {
		var rule models.IPRule
		if err := rows.Scan(&rule.ID, &rule.OrgID, &rule.CIDR, &rule.Action, &rule.Description, &rule.AddedBy, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *postgresRepository) GetIPRuleByID(ctx context.Context, id uuid.UUID) (*models.IPRule, error) {
	var rule models.IPRule
	query := `SELECT id, org_id, cidr, action, description, added_by, created_at, updated_at 
			  FROM ip_rules WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&rule.ID, &rule.OrgID, &rule.CIDR, &rule.Action, &rule.Description, &rule.AddedBy, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *postgresRepository) CreateIPRule(ctx context.Context, rule *models.IPRule) error {
	rule.ID = uuid.New()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	query := `INSERT INTO ip_rules (id, org_id, cidr, action, description, added_by, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, query, rule.ID, rule.OrgID, rule.CIDR, rule.Action, rule.Description, rule.AddedBy, rule.CreatedAt, rule.UpdatedAt)
	return err
}

func (r *postgresRepository) DeleteIPRule(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM ip_rules WHERE id = $1`, id)
	return err
}

// Access Reviews

func (r *postgresRepository) ListAccessReviews(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AccessReview, int, error) {
	offset := (page - 1) * size
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM access_reviews WHERE org_id = $1`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, org_id, title, status, due_date, total_users, reviewed_users, created_by, created_at, updated_at 
			  FROM access_reviews WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []models.AccessReview
	for rows.Next() {
		var r models.AccessReview
		if err := rows.Scan(&r.ID, &r.OrgID, &r.Title, &r.Status, &r.DueDate, &r.TotalUsers, &r.ReviewedUsers, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, r)
	}
	return reviews, total, nil
}

func (r *postgresRepository) GetAccessReviewByID(ctx context.Context, id uuid.UUID) (*models.AccessReview, error) {
	var review models.AccessReview
	query := `SELECT id, org_id, title, status, due_date, total_users, reviewed_users, created_by, created_at, updated_at 
			  FROM access_reviews WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&review.ID, &review.OrgID, &review.Title, &review.Status, &review.DueDate, &review.TotalUsers, &review.ReviewedUsers, &review.CreatedBy, &review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *postgresRepository) CreateAccessReview(ctx context.Context, review *models.AccessReview) error {
	review.ID = uuid.New()
	review.CreatedAt = time.Now()
	review.UpdatedAt = time.Now()
	query := `INSERT INTO access_reviews (id, org_id, title, status, due_date, total_users, reviewed_users, created_by, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, query, review.ID, review.OrgID, review.Title, review.Status, review.DueDate, review.TotalUsers, review.ReviewedUsers, review.CreatedBy, review.CreatedAt, review.UpdatedAt)
	return err
}

func (r *postgresRepository) UpdateAccessReviewStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE access_reviews SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (r *postgresRepository) ListReviewItems(ctx context.Context, reviewID uuid.UUID) ([]models.ReviewItem, error) {
	query := `SELECT ri.id, ri.review_id, ri.user_id, ri.role_id, ri.current_status, ri.reviewer_id, ri.reviewed_at, ri.notes,
			  u.id, u.email, u.name, u.status,
			  ro.id, ro.name
			  FROM review_items ri
			  LEFT JOIN users u ON ri.user_id = u.id
			  LEFT JOIN roles ro ON ri.role_id = ro.id
			  WHERE ri.review_id = $1`
	rows, err := r.db.Query(ctx, query, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ReviewItem
	for rows.Next() {
		var item models.ReviewItem
		var user models.User
		var role models.Role
		if err := rows.Scan(
			&item.ID, &item.ReviewID, &item.UserID, &item.RoleID, &item.CurrentStatus, &item.ReviewerID, &item.ReviewedAt, &item.Notes,
			&user.ID, &user.Email, &user.Name, &user.Status,
			&role.ID, &role.Name,
		); err != nil {
			return nil, err
		}
		item.User = &user
		item.Role = &role
		items = append(items, item)
	}
	return items, nil
}

func (r *postgresRepository) UpdateReviewItem(ctx context.Context, itemID uuid.UUID, status string, reviewerID uuid.UUID, notes string) error {
	now := time.Now()
	_, err := r.db.Exec(ctx, `UPDATE review_items SET current_status = $1, reviewer_id = $2, reviewed_at = $3, notes = $4 WHERE id = $5`,
		status, reviewerID, now, notes, itemID)
	return err
}

// Sessions - Extended for Admin management

func (r *postgresRepository) ListAllSessions(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.AuthToken, int, error) {
	offset := (page - 1) * size
	var total int
	countQuery := `SELECT COUNT(*) FROM auth_tokens t 
				   JOIN users u ON t.user_id = u.id
				   JOIN user_organizations uo ON u.id = uo.user_id
				   WHERE uo.org_id = $1 AND t.expires_at > NOW()`
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT t.id, t.user_id, t.access_token, t.refresh_token, t.device_id, 
					 t.ip_address, t.user_agent, t.last_active, t.is_active,
					 t.expires_at, t.created_at
			  FROM auth_tokens t
			  JOIN users u ON t.user_id = u.id
			  JOIN user_organizations uo ON u.id = uo.user_id
			  WHERE uo.org_id = $1 AND t.expires_at > NOW()
			  ORDER BY t.created_at DESC
			  LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tokens []models.AuthToken
	for rows.Next() {
		var t models.AuthToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.AccessToken, &t.RefreshToken, &t.DeviceID, 
							&t.IPAddress, &t.UserAgent, &t.LastActive, &t.IsActive,
							&t.ExpiresAt, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		tokens = append(tokens, t)
	}
	return tokens, total, nil
}

func (r *postgresRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM auth_tokens WHERE id = $1`, sessionID)
	return err
}

func (r *postgresRepository) RevokeAllSessionsForOrg(ctx context.Context, orgID uuid.UUID) error {
	query := `DELETE FROM auth_tokens WHERE user_id IN (
				SELECT u.id FROM users u
				JOIN user_organizations uo ON u.id = uo.user_id
				WHERE uo.org_id = $1
			  )`
	_, err := r.db.Exec(ctx, query, orgID)
	return err
}

// Security Production Expansion Implementation

func (r *postgresRepository) LogLoginEvent(ctx context.Context, event *models.LoginEvent) error {
	event.ID = uuid.New()
	event.CreatedAt = time.Now()
	query := `INSERT INTO login_events (id, org_id, user_id, email, ip_address, device, browser, location, status, risk_level, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.Exec(ctx, query, event.ID, event.OrgID, event.UserID, event.Email, event.IPAddress, event.Device, event.Browser, event.Location, event.Status, event.RiskLevel, event.CreatedAt)
	return err
}

func (r *postgresRepository) ListLoginEvents(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.LoginEvent, int, error) {
	offset := (page - 1) * size
	var total int
	
	countQuery := `SELECT COUNT(*) FROM login_events WHERE org_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, org_id, user_id, email, ip_address, device, browser, location, status, risk_level, created_at
			  FROM login_events WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []models.LoginEvent
	for rows.Next() {
		var e models.LoginEvent
		if err := rows.Scan(&e.ID, &e.OrgID, &e.UserID, &e.Email, &e.IPAddress, &e.Device, &e.Browser, &e.Location, &e.Status, &e.RiskLevel, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		events = append(events, e)
	}
	return events, total, nil
}

func (r *postgresRepository) GetLoginEventByID(ctx context.Context, id uuid.UUID) (*models.LoginEvent, error) {
	var e models.LoginEvent
	query := `SELECT id, org_id, user_id, email, ip_address, device, browser, location, status, risk_level, created_at
			  FROM login_events WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&e.ID, &e.OrgID, &e.UserID, &e.Email, &e.IPAddress, &e.Device, &e.Browser, &e.Location, &e.Status, &e.RiskLevel, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *postgresRepository) CreateThreatAlert(ctx context.Context, alert *models.ThreatAlert) error {
	alert.ID = uuid.New()
	alert.CreatedAt = time.Now()
	query := `INSERT INTO threat_alerts (id, org_id, type, severity, ip, user_id, description, status, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, query, alert.ID, alert.OrgID, alert.Type, alert.Severity, alert.IP, alert.UserID, alert.Description, alert.Status, alert.CreatedAt)
	return err
}

func (r *postgresRepository) ListThreatAlerts(ctx context.Context, orgID uuid.UUID, page, size int, filters map[string]any) ([]models.ThreatAlert, int, error) {
	offset := (page - 1) * size
	var total int
	
	countQuery := `SELECT COUNT(*) FROM threat_alerts WHERE org_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, org_id, type, severity, ip, user_id, description, status, created_at
			  FROM threat_alerts WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []models.ThreatAlert
	for rows.Next() {
		var a models.ThreatAlert
		if err := rows.Scan(&a.ID, &a.OrgID, &a.Type, &a.Severity, &a.IP, &a.UserID, &a.Description, &a.Status, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		alerts = append(alerts, a)
	}
	return alerts, total, nil
}

func (r *postgresRepository) GetThreatAlertByID(ctx context.Context, id uuid.UUID) (*models.ThreatAlert, error) {
	var a models.ThreatAlert
	query := `SELECT id, org_id, type, severity, ip, user_id, description, status, created_at
			  FROM threat_alerts WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&a.ID, &a.OrgID, &a.Type, &a.Severity, &a.IP, &a.UserID, &a.Description, &a.Status, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *postgresRepository) UpdateThreatAlertStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE threat_alerts SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *postgresRepository) ListMFAUsers(ctx context.Context, orgID uuid.UUID, page, size int) ([]models.MFAUser, int, error) {
	offset := (page - 1) * size
	var total int
	
	countQuery := `SELECT COUNT(*) FROM mfa_users WHERE org_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT user_id, org_id, status, method, trusted_devices_count, last_reset, updated_at
			  FROM mfa_users WHERE org_id = $1 ORDER BY updated_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, orgID, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.MFAUser
	for rows.Next() {
		var u models.MFAUser
		if err := rows.Scan(&u.UserID, &u.OrgID, &u.Status, &u.Method, &u.TrustedDevicesCount, &u.LastReset, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *postgresRepository) GetMFAUser(ctx context.Context, userID uuid.UUID) (*models.MFAUser, error) {
	var u models.MFAUser
	query := `SELECT user_id, org_id, status, method, trusted_devices_count, last_reset, updated_at
			  FROM mfa_users WHERE user_id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&u.UserID, &u.OrgID, &u.Status, &u.Method, &u.TrustedDevicesCount, &u.LastReset, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *postgresRepository) UpdateMFAUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE mfa_users SET status = $1, updated_at = NOW() WHERE user_id = $2`, status, userID)
	return err
}

func (r *postgresRepository) GetSecurityTelemetry(ctx context.Context) (map[string]any, error) {
	var totalUsers int
	var mfaUsers int
	var failedLogins24h int
	var totalLogins24h int

	// Total and MFA users
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE mfa_enabled = true").Scan(&mfaUsers)

	// Logins in last 24h
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM login_events WHERE status = 'failed' AND created_at > NOW() - INTERVAL '24 hours'").Scan(&failedLogins24h)
	_ = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM login_events WHERE created_at > NOW() - INTERVAL '24 hours'").Scan(&totalLogins24h)

	mfaAdoption := 0.0
	if totalUsers > 0 {
		mfaAdoption = (float64(mfaUsers) / float64(totalUsers)) * 100.0
	}

	loginFailureRate := 0.0
	if totalLogins24h > 0 {
		loginFailureRate = (float64(failedLogins24h) / float64(totalLogins24h)) * 100.0
	}

	return map[string]any{
		"mfa_adoption":       mfaAdoption,
		"failed_login_ratio": loginFailureRate,
		"total_users":        totalUsers,
		"mfa_users":          mfaUsers,
	}, nil
}
