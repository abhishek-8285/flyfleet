package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, COALESCE(name, ''), COALESCE(phone, ''), COALESCE(avatar_url, ''), COALESCE(mfa_enabled, false), COALESCE(mfa_secret, ''),
		       COALESCE(mfa_backup_codes, '[]'::jsonb), COALESCE(status, 'active'), COALESCE(email_verified, false),
		       COALESCE(failed_logins, 0), locked_until, last_login_at,
		       COALESCE(last_login_ip, '')::text, COALESCE(permission_mask, 0),
		       COALESCE(sidebar_collapsed, false), COALESCE(language, 'en'), COALESCE(theme, 'light'),
		       COALESCE(must_reset_password, false),
		       created_at, updated_at
		FROM users WHERE email = $1`, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.AvatarURL,
		&u.MFAEnabled, &u.MFASecret, &u.MFABackupCodes, &u.Status, &u.EmailVerified,
		&u.FailedLogins, &u.LockedUntil, &u.LastLoginAt, &u.LastLoginIP,
		&u.PermissionMask, &u.SidebarCollapsed, &u.Language, &u.Theme, &u.MustResetPassword, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, COALESCE(name, ''), COALESCE(phone, ''), COALESCE(avatar_url, ''), COALESCE(mfa_enabled, false), COALESCE(mfa_secret, ''),
		       COALESCE(mfa_backup_codes, '[]'::jsonb), COALESCE(status, 'active'), COALESCE(email_verified, false),
		       COALESCE(failed_logins, 0), locked_until, last_login_at,
		       COALESCE(last_login_ip, '')::text, COALESCE(permission_mask, 0),
		       COALESCE(sidebar_collapsed, false), COALESCE(language, 'en'), COALESCE(theme, 'light'),
		       COALESCE(must_reset_password, false),
		       created_at, updated_at
		FROM users WHERE id = $1`, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.AvatarURL,
		&u.MFAEnabled, &u.MFASecret, &u.MFABackupCodes, &u.Status, &u.EmailVerified,
		&u.FailedLogins, &u.LockedUntil, &u.LastLoginAt, &u.LastLoginIP,
		&u.PermissionMask, &u.SidebarCollapsed, &u.Language, &u.Theme, &u.MustResetPassword, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET password_hash = $1, failed_logins = 0, locked_until = NULL, updated_at = NOW() WHERE id = $2",
		hash, id,
	)
	return err
}

func (r *UserRepository) UpdateMFASecret(ctx context.Context, id uuid.UUID, secret string, enabled bool) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET mfa_secret = $1, mfa_enabled = $2, updated_at = NOW() WHERE id = $3",
		secret, enabled, id,
	)
	return err
}

func (r *UserRepository) UpdateMFABackupCodes(ctx context.Context, id uuid.UUID, codes []string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET mfa_backup_codes = $1, updated_at = NOW() WHERE id = $2",
		codes, id,
	)
	return err
}

func (r *UserRepository) IncrementFailedLogins(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET failed_logins = failed_logins + 1, updated_at = NOW() WHERE id = $1",
		id,
	)
	return err
}

func (r *UserRepository) ResetFailedLogins(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET failed_logins = 0, locked_until = NULL, updated_at = NOW() WHERE id = $1",
		id,
	)
	return err
}

func (r *UserRepository) LockAccount(ctx context.Context, id uuid.UUID, until time.Time) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET locked_until = $1, status = 'locked', updated_at = NOW() WHERE id = $2",
		until, id,
	)
	return err
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET last_login_at = NOW(), last_login_ip = $1 WHERE id = $2",
		ip, id,
	)
	return err
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, name, phone, avatarURL, theme, language string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET 
			name = COALESCE(NULLIF($1, ''), name), 
			phone = COALESCE(NULLIF($2, ''), phone), 
			avatar_url = COALESCE(NULLIF($3, ''), avatar_url),
			theme = COALESCE(NULLIF($4, ''), theme),
			language = COALESCE(NULLIF($5, ''), language),
			updated_at = NOW() 
		WHERE id = $6`,
		name, phone, avatarURL, theme, language, id,
	)
	return err
}

func (r *UserRepository) UpdatePreferences(ctx context.Context, id, tenantID uuid.UUID, sidebarCollapsed *bool, language *string, theme *string) error {
	var err error
	if tenantID != uuid.Nil {
		_, err = r.pool.Exec(ctx,
			`UPDATE users SET
				sidebar_collapsed = COALESCE($1, sidebar_collapsed),
				language = COALESCE($2, language),
				theme = COALESCE($3, theme),
				updated_at = NOW()
			WHERE id = $4 AND id IN (SELECT user_id FROM user_tenants WHERE tenant_id = $5)`,
			sidebarCollapsed, language, theme, id, tenantID,
		)
	} else {
		_, err = r.pool.Exec(ctx,
			`UPDATE users SET
				sidebar_collapsed = COALESCE($1, sidebar_collapsed),
				language = COALESCE($2, language),
				theme = COALESCE($3, theme),
				updated_at = NOW()
			WHERE id = $4`,
			sidebarCollapsed, language, theme, id,
		)
	}
	return err
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, status models.UserStatus) error {
	var err error
	if tenantID != uuid.Nil {
		_, err = r.pool.Exec(ctx, `
			UPDATE users SET status = $1, updated_at = NOW() 
			WHERE id = $2 AND id IN (SELECT user_id FROM user_tenants WHERE tenant_id = $3)`,
			status, id, tenantID,
		)
	} else {
		_, err = r.pool.Exec(ctx,
			"UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2",
			status, id,
		)
	}
	return err
}

func (r *UserRepository) CreateUser(ctx context.Context, u *models.User) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name, phone, avatar_url, status, email_verified, must_reset_password)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`,
		u.Email, u.PasswordHash, u.Name, u.Phone, u.AvatarURL, u.Status, u.EmailVerified, u.MustResetPassword,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) UpdatePasswordAndClearResetFlag(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET password_hash = $1, must_reset_password = false, failed_logins = 0, locked_until = NULL, updated_at = NOW() WHERE id = $2",
		hash, id,
	)
	return err
}

func (r *UserRepository) ListUsers(ctx context.Context, params models.ListUsersParams) (*models.ListUsersResponse, error) {
	query := `
		SELECT u.id, u.email, COALESCE(u.name, '') as name, COALESCE(u.status, 'active') as status, u.last_login_at, u.created_at
		FROM users u`
	
	countQuery := `SELECT COUNT(*) FROM users u`
	
	whereClause := " WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if params.TenantID != uuid.Nil {
		query += " JOIN user_tenants uo ON u.id = uo.user_id"
		countQuery += " JOIN user_tenants uo ON u.id = uo.user_id"
			whereClause += " AND uo.tenant_id = $" + strconv.Itoa(argCount)
		args = append(args, params.TenantID)
		argCount++
	}

	if params.Search != "" {
		whereClause += " AND (u.email ILIKE $" + strconv.Itoa(argCount) + " OR u.name ILIKE $" + strconv.Itoa(argCount) + ")"
		args = append(args, "%"+params.Search+"%")
		argCount++
	}
	if params.Status != "" {
		whereClause += " AND u.status = $" + strconv.Itoa(argCount)
		args = append(args, params.Status)
		argCount++
	}

	query += whereClause
	countQuery += whereClause

	// Count
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	query += " ORDER BY u.created_at DESC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, params.PageSize, (params.Page-1)*params.PageSize)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.UserListItem
	for rows.Next() {
		var u models.UserListItem
		var lastLoginAt *time.Time
		var createdAt time.Time
		err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Status, &lastLoginAt, &createdAt)
		if err != nil {
			return nil, err
		}
		if lastLoginAt != nil {
			s := lastLoginAt.Format(time.RFC3339)
			u.LastLoginAt = &s
		}
		u.CreatedAt = createdAt.Format(time.RFC3339)
		users = append(users, u)
	}

	return &models.ListUsersResponse{
		Users:    users,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, id, tenantID uuid.UUID) error {
	var err error
	if tenantID != uuid.Nil {
		_, err = r.pool.Exec(ctx, `
			UPDATE users SET status = 'disabled', updated_at = NOW() 
			WHERE id = $1 AND id IN (SELECT user_id FROM user_tenants WHERE tenant_id = $2)`,
			id, tenantID,
		)
	} else {
		_, err = r.pool.Exec(ctx,
			"UPDATE users SET status = 'disabled', updated_at = NOW() WHERE id = $1",
			id,
		)
	}
	return err
}

// MarkEmailVerified marks the user's email as verified.
func (r *UserRepository) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE users SET email_verified = true, updated_at = NOW() WHERE id = $1", id,
	)
	return err
}

// GetTenants returns all tenants the user belongs to.
func (r *UserRepository) GetTenants(ctx context.Context, userID uuid.UUID) ([]models.TenantInfo, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT o.id, COALESCE(o.name, '') as name, COALESCE(uo.role, 'member') as role
		FROM user_tenants uo
		JOIN tenants o ON uo.tenant_id = o.id
		WHERE uo.user_id = $1 AND uo.is_active = true
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []models.TenantInfo
	for rows.Next() {
		var tenant models.TenantInfo
		if err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Role); err != nil {
			continue
		}
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}
