package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/fleet-platform/shared/pkg/metrics"
	"github.com/flyfleet/fleet-core/internal/domains/auth/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type AuthRepository struct {
	db     *pgxpool.Pool
	logger logger.Logger
}

func NewAuthRepository(db *pgxpool.Pool, logger logger.Logger) *AuthRepository {
	return &AuthRepository{db: db, logger: logger}
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	start0 := time.Now()
	var dbName string
	_ = r.db.QueryRow(ctx, "SELECT current_database()").Scan(&dbName)
	metrics.RecordDBQuery("SELECT current_database()", time.Since(start0))
	r.logger.Info("GetUserByEmail DEBUG", zap.String("email", email), zap.String("db", dbName))

	// Security: Bypass RLS for initial user lookup during login
	// Use a transaction to ensure set_config and the query use the SAME connection
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, _ = tx.Exec(ctx, "SELECT set_config('flyfleet.bypass_rls', 'true', true)")

	query := `
		SELECT u.id, u.org_id, u.email, COALESCE(r.name, u.role, 'user') as role, u.password_hash, 
		u.created_at, u.created_by, u.updated_at, u.updated_by, u.mfa_enabled, u.mfa_methods, 
		u.mfa_secret, u.last_password_change, u.failed_login_attempts, u.locked_until 
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1
	`

	start2 := time.Now()
	var user models.User
	err = tx.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.OrgID, &user.Email, &user.Role, &user.PasswordHash,
		&user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
		&user.MfaEnabled, &user.MfaMethods, &user.MfaSecret, &user.LastPasswordChange,
		&user.FailedLoginAttempts, &user.LockedUntil,
	)
	r.logger.Info("GetUserByEmail RESULT", zap.String("email", email), zap.String("role", user.Role))
	metrics.RecordDBQuery(query, time.Since(start2))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &user, nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, org_id, email, password_hash, role, is_active, created_at, created_by, updated_at, updated_by, mfa_enabled, mfa_methods, last_password_change)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	var targetOrgID interface{} = user.OrgID
	if user.OrgID == uuid.Nil {
		targetOrgID = nil
	}

	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		user.ID, targetOrgID, user.Email, user.PasswordHash, user.Role, true,
		user.CreatedAt, user.CreatedBy, user.UpdatedAt, user.UpdatedBy,
		user.MfaEnabled, user.MfaMethods, user.LastPasswordChange,
	)
	metrics.RecordDBQuery(query, time.Since(start))

	if err != nil {
		r.logger.Error("SQL Exec failed in CreateUser", zap.Error(err), zap.String("query", query), zap.String("email", user.Email))
	}

	return err
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT u.id, u.org_id, u.email, COALESCE(r.name, u.role, 'user') as role, 
		u.created_at, u.created_by, u.updated_at, u.updated_by, u.mfa_enabled, u.mfa_methods, 
		u.mfa_secret, u.last_password_change, u.failed_login_attempts, u.locked_until 
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`

	start := time.Now()
	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.OrgID, &user.Email, &user.Role,
		&user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
		&user.MfaEnabled, &user.MfaMethods, &user.MfaSecret, &user.LastPasswordChange,
		&user.FailedLoginAttempts, &user.LockedUntil,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	return &user, nil
}

func (r *AuthRepository) UpdateUserLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error {
	query := "UPDATE users SET failed_login_attempts = $1, locked_until = $2, updated_at = NOW() WHERE id = $3"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, attempts, lockedUntil, id)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) ResetLoginAttempts(ctx context.Context, id uuid.UUID) error {
	query := "UPDATE users SET failed_login_attempts = 0, locked_until = NULL, updated_at = NOW() WHERE id = $1"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

// RefreshToken operations
func (r *AuthRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	query := "INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, revoked, ip_address, user_agent, last_used_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt, token.Revoked, token.IPAddress, token.UserAgent, token.LastUsedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, tokenStr string) (*models.RefreshToken, error) {
	query := "SELECT id, user_id, token, expires_at, created_at, revoked, revoked_at, ip_address::text, user_agent, last_used_at FROM refresh_tokens WHERE token = $1 AND expires_at > NOW()"
	start := time.Now()
	var rt models.RefreshToken
	err := r.db.QueryRow(ctx, query, tokenStr).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked, &rt.RevokedAt, &rt.IPAddress, &rt.UserAgent, &rt.LastUsedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	// Update last_used_at
	updateQuery := "UPDATE refresh_tokens SET last_used_at = NOW() WHERE id = $1"
	start2 := time.Now()
	_, _ = r.db.Exec(ctx, updateQuery, rt.ID)
	metrics.RecordDBQuery(updateQuery, time.Since(start2))

	return &rt, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, tokenStr string) error {
	now := time.Time{}
	query := "UPDATE refresh_tokens SET revoked = true, revoked_at = $1 WHERE token = $2"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, now, tokenStr)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

// TryRevokeRefreshToken atomically revokes the token only if it's not already revoked.
// Returns the token row if successfully revoked, nil if already revoked (no rows affected).
// This prevents the race condition where concurrent refresh requests all see revoked=false.
func (r *AuthRepository) TryRevokeRefreshToken(ctx context.Context, tokenStr string) (*models.RefreshToken, error) {
	now := time.Now()
	query := `UPDATE refresh_tokens 
			  SET revoked = true, revoked_at = $1 
			  WHERE token = $2 AND revoked = false AND expires_at > NOW()
			  RETURNING id, user_id, token, expires_at, created_at, revoked, revoked_at, ip_address::text, user_agent, last_used_at`
	start := time.Now()
	var rt models.RefreshToken
	err := r.db.QueryRow(ctx, query, now, tokenStr).Scan(
		&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked, &rt.RevokedAt, &rt.IPAddress, &rt.UserAgent, &rt.LastUsedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Token already revoked or expired — not an error
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	return &rt, nil
}

func (r *AuthRepository) ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]models.RefreshToken, error) {
	query := `SELECT id, user_id, token, expires_at, created_at, revoked, revoked_at, ip_address, user_agent, last_used_at
			  FROM refresh_tokens
			  WHERE user_id = $1 AND revoked = false AND expires_at > NOW()
			  ORDER BY last_used_at DESC NULLS LAST`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, userID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	var sessions []models.RefreshToken
	for rows.Next() {
		var rt models.RefreshToken
		if err := rows.Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked, &rt.RevokedAt, &rt.IPAddress, &rt.UserAgent, &rt.LastUsedAt); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, rt)
	}
	return sessions, rows.Err()
}

func (r *AuthRepository) RevokeSessionByID(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error {
	query := "UPDATE refresh_tokens SET revoked = true, revoked_at = NOW() WHERE id = $1 AND user_id = $2"
	start := time.Now()
	res, err := r.db.Exec(ctx, query, sessionID, userID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("session not found or not owned by user")
	}
	return nil
}

func (r *AuthRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := "UPDATE refresh_tokens SET revoked = true, revoked_at = $1 WHERE user_id = $2"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, now, userID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) CleanupExpiredTokens(ctx context.Context) error {
	query := "DELETE FROM refresh_tokens WHERE expires_at < NOW() - INTERVAL '30 days'"
	start := time.Now()
	_, err := r.db.Exec(ctx, query)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error {
	query := "INSERT INTO security_events (id, user_id, event_type, ip_address, user_agent, metadata, actor_id, on_behalf_of) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, uuid.New(), event.UserID, event.EventType, event.IPAddress, event.UserAgent, event.Metadata, event.ActorID, event.OnBehalfOf)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

// MFA Operations
func (r *AuthRepository) UpdateMFASecret(ctx context.Context, userID uuid.UUID, secret string) error {
	query := "UPDATE users SET mfa_secret = $1, updated_at = NOW() WHERE id = $2"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, secret, userID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) EnableMFA(ctx context.Context, userID uuid.UUID, secret string) error {
	query := `UPDATE users SET mfa_enabled = true,
			  mfa_methods = CASE
			    WHEN 'totp' = ANY(mfa_methods) THEN mfa_methods
			    ELSE array_append(mfa_methods, 'totp')
			  END,
			  mfa_secret = $1,
			  updated_at = NOW()
			  WHERE id = $2`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, secret, userID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	query := "UPDATE users SET mfa_enabled = false, mfa_secret = NULL, updated_at = NOW() WHERE id = $1"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, userID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

// Password Reset Operations
func (r *AuthRepository) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	query := "INSERT INTO password_reset_tokens (id, user_id, token, expires_at, created_at, used) VALUES ($1, $2, $3, $4, $5, false)"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, uuid.New(), userID, token, expiresAt, time.Now())
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) GetPasswordResetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	query := "SELECT id, user_id, token, expires_at, created_at, used FROM password_reset_tokens WHERE token = $1 AND used = false AND expires_at > NOW()"
	start := time.Now()
	var prt models.PasswordResetToken
	err := r.db.QueryRow(ctx, query, token).Scan(
		&prt.ID, &prt.UserID, &prt.Token, &prt.ExpiresAt, &prt.CreatedAt, &prt.Used,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired reset token")
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	return &prt, nil
}

func (r *AuthRepository) UsePasswordResetToken(ctx context.Context, token string) error {
	query := "UPDATE password_reset_tokens SET used = true, used_at = NOW() WHERE token = $1"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, token)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := "UPDATE users SET password_hash = $1, last_password_change = NOW(), updated_at = NOW(), failed_login_attempts = 0, locked_until = NULL WHERE id = $2"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, passwordHash, userID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) CleanupExpiredResetTokens(ctx context.Context) error {
	query := "DELETE FROM password_reset_tokens WHERE expires_at < NOW() - INTERVAL '24 hours'"
	start := time.Now()
	_, err := r.db.Exec(ctx, query)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) GetRolePermissions(ctx context.Context, role string) ([]string, error) {
	query := `SELECT p.resource || ':' || p.action FROM permissions p JOIN roles r ON p.role_id = r.id WHERE r.name = $1`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, role)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, p)
	}
	return permissions, rows.Err()
}

func (r *AuthRepository) SaveRecoveryCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error {
	start := time.Now()
	tx, err := r.db.Begin(ctx)
	metrics.RecordDBQuery("BEGIN", time.Since(start))
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete old codes
	start1 := time.Now()
	_, err = tx.Exec(ctx, "DELETE FROM mfa_recovery_codes WHERE user_id = $1", userID)
	metrics.RecordDBQuery("DELETE FROM mfa_recovery_codes WHERE user_id = $1", time.Since(start1))
	if err != nil {
		return err
	}

	// Batch insert recovery codes to avoid N+1 queries
	if len(codeHashes) > 0 {
		var args []interface{}
		var placeholders []string
		for i, hash := range codeHashes {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
			args = append(args, userID, hash)
		}
		batchQuery := fmt.Sprintf(
			"INSERT INTO mfa_recovery_codes (user_id, code_hash) VALUES %s",
			strings.Join(placeholders, ", "),
		)
		start2 := time.Now()
		_, err = tx.Exec(ctx, batchQuery, args...)
		metrics.RecordDBQuery(batchQuery, time.Since(start2))
		if err != nil {
			return fmt.Errorf("failed to batch insert recovery codes: %w", err)
		}
	}

	start3 := time.Now()
	err = tx.Commit(ctx)
	metrics.RecordDBQuery("COMMIT", time.Since(start3))
	return err
}

func (r *AuthRepository) GetUnusedRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]models.MfaRecoveryCode, error) {
	query := "SELECT id, user_id, code_hash, used, created_at FROM mfa_recovery_codes WHERE user_id = $1 AND used = false"
	start := time.Now()
	rows, err := r.db.Query(ctx, query, userID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []models.MfaRecoveryCode
	for rows.Next() {
		var c models.MfaRecoveryCode
		if err := rows.Scan(&c.ID, &c.UserID, &c.CodeHash, &c.Used, &c.CreatedAt); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

func (r *AuthRepository) MarkRecoveryCodeUsed(ctx context.Context, codeID uuid.UUID) error {
	query := "UPDATE mfa_recovery_codes SET used = true WHERE id = $1"
	start := time.Now()
	_, err := r.db.Exec(ctx, query, codeID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) GetLastLogin(ctx context.Context, userID uuid.UUID) (*models.LoginHistory, error) {
	return nil, nil
}

func (r *AuthRepository) LogLoginAttempt(ctx context.Context, h *models.LoginHistory) error {
	return nil
}

func (r *AuthRepository) GetOidcProviderByName(ctx context.Context, name string) (*models.OidcProvider, error) {
	query := `SELECT id, name, issuer, client_id, client_secret_hash, enabled, created_at, updated_at
			  FROM oidc_providers WHERE name = $1`
	start := time.Now()
	var p models.OidcProvider
	err := r.db.QueryRow(ctx, query, name).Scan(
		&p.ID, &p.Name, &p.Issuer, &p.ClientID, &p.ClientSecretHash, &p.Enabled, &p.CreatedAt, &p.UpdatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *AuthRepository) GetUserByOidcLink(ctx context.Context, providerID uuid.UUID, providerUserID string) (*models.User, error) {
	query := `SELECT u.id, u.org_id, u.email, COALESCE(CAST(u.role_id AS VARCHAR), u.role) as role, u.password_hash, u.mfa_enabled, u.mfa_secret, u.failed_login_attempts, u.locked_until, u.created_at
			  FROM users u
			  JOIN user_oidc_links l ON u.id = l.user_id
			  WHERE l.provider_id = $1 AND l.provider_user_id = $2`
	start := time.Now()
	var u models.User
	err := r.db.QueryRow(ctx, query, providerID, providerUserID).Scan(
		&u.ID, &u.OrgID, &u.Email, &u.Role, &u.PasswordHash, &u.MfaEnabled, &u.MfaSecret, &u.FailedLoginAttempts, &u.LockedUntil, &u.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepository) LinkUserToOidc(ctx context.Context, link *models.UserOidcLink) error {
	query := `INSERT INTO user_oidc_links (id, user_id, provider_id, provider_user_id, email, last_login_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	if link.ID == uuid.Nil {
		link.ID = uuid.New()
	}
	start := time.Now()
	_, err := r.db.Exec(ctx, query, link.ID, link.UserID, link.ProviderID, link.ProviderUserID, link.Email, link.LastLoginAt)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) UpdateOidcLinkLastLogin(ctx context.Context, linkID uuid.UUID) error {
	query := `UPDATE user_oidc_links SET last_login_at = NOW() WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, linkID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) GetActiveSigningKeys(ctx context.Context) ([]models.SigningKey, error) {
	query := `SELECT kid, private_key_enc, public_key, status, algorithm, created_at, activated_at, expires_at
			  FROM signing_keys WHERE status IN ('active', 'previous')`
	start := time.Now()
	rows, err := r.db.Query(ctx, query)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.SigningKey
	for rows.Next() {
		var k models.SigningKey
		err := rows.Scan(&k.KID, &k.PrivateKeyEnc, &k.PublicKey, &k.Status, &k.Algorithm, &k.CreatedAt, &k.ActivatedAt, &k.ExpiresAt)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (r *AuthRepository) GetSigningKey(ctx context.Context, kid string) (*models.SigningKey, error) {
	query := `SELECT kid, private_key_enc, public_key, status, algorithm, created_at, activated_at, expires_at
			  FROM signing_keys WHERE kid = $1`
	start := time.Now()
	var k models.SigningKey
	err := r.db.QueryRow(ctx, query, kid).Scan(&k.KID, &k.PrivateKeyEnc, &k.PublicKey, &k.Status, &k.Algorithm, &k.CreatedAt, &k.ActivatedAt, &k.ExpiresAt)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}

func (r *AuthRepository) SaveSigningKey(ctx context.Context, k *models.SigningKey) error {
	query := `INSERT INTO signing_keys (kid, private_key_enc, public_key, status, algorithm, created_at, activated_at, expires_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, k.KID, k.PrivateKeyEnc, k.PublicKey, k.Status, k.Algorithm, k.CreatedAt, k.ActivatedAt, k.ExpiresAt)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) UpdateSigningKeyStatus(ctx context.Context, kid string, status string) error {
	query := `UPDATE signing_keys SET status = $1 WHERE kid = $2`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, status, kid)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

// API Keys & M2M
func (r *AuthRepository) SaveAPIKey(ctx context.Context, k *models.APIKey) error {
	query := `INSERT INTO api_keys (id, org_id, name, prefix, secret_hash, created_at, expires_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, k.ID, k.OrgID, k.Name, k.Prefix, k.SecretHash, k.CreatedAt, k.ExpiresAt)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) GetAPIKeyByPrefix(ctx context.Context, prefix string) (*models.APIKey, error) {
	query := `SELECT id, org_id, name, prefix, secret_hash, created_at, expires_at, last_used_at, revoked
			  FROM api_keys WHERE prefix = $1 AND revoked = false`
	start := time.Now()
	var k models.APIKey
	err := r.db.QueryRow(ctx, query, prefix).Scan(
		&k.ID, &k.OrgID, &k.Name, &k.Prefix, &k.SecretHash, &k.CreatedAt, &k.ExpiresAt, &k.LastUsedAt, &k.Revoked,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &k, nil
}

func (r *AuthRepository) UpdateAPIKeyLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) GetServiceAccount(ctx context.Context, id uuid.UUID) (*models.ServiceAccount, error) {
	query := `SELECT id, org_id, name, secret_hash, created_at FROM service_accounts WHERE id = $1`
	start := time.Now()
	var sa models.ServiceAccount
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sa.ID, &sa.OrgID, &sa.Name, &sa.SecretHash, &sa.CreatedAt,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &sa, nil
}

// Compliance & Admin
func (r *AuthRepository) RevokeAllTokensGlobal(ctx context.Context) error {
	query := `UPDATE refresh_tokens SET revoked = true, revoked_at = NOW() WHERE revoked = false`
	start := time.Now()
	_, err := r.db.Exec(ctx, query)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) AnonymizeUser(ctx context.Context, userID uuid.UUID) error {
	// Pattern: Right to be Forgotten (Compliance)
	// We keep the record for referential integrity but overwrite sensitive data
	query := `UPDATE users SET
				email = 'deleted-' || id || '@flyfleet.internal',
				password_hash = 'GDPR_DELETED',
				mfa_secret = NULL,
				mfa_enabled = false,
				updated_at = NOW(),
				is_deleted = true
			  WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, userID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]models.APIKey, error) {
	query := `SELECT id, org_id, name, prefix, secret_hash, created_at, expires_at, last_used_at, revoked FROM api_keys WHERE org_id = $1 ORDER BY created_at DESC`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.OrgID, &k.Name, &k.Prefix, &k.SecretHash, &k.CreatedAt, &k.ExpiresAt, &k.LastUsedAt, &k.Revoked); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *AuthRepository) RevokeAPIKey(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET revoked = true WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) CreateServiceAccount(ctx context.Context, sa *models.ServiceAccount) error {
	query := `INSERT INTO service_accounts (id, org_id, name, description, client_id, client_secret_hash, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, sa.ID, sa.OrgID, sa.Name, sa.Description, sa.ClientID, sa.ClientSecretHash, sa.Role, sa.CreatedAt, sa.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) ListServiceAccounts(ctx context.Context, orgID uuid.UUID) ([]models.ServiceAccount, error) {
	query := `SELECT id, org_id, name, description, client_id, client_secret_hash, role, created_at, updated_at, last_used_at FROM service_accounts WHERE org_id = $1 ORDER BY created_at DESC`
	start := time.Now()
	rows, err := r.db.Query(ctx, query, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.ServiceAccount
	for rows.Next() {
		var sa models.ServiceAccount
		if err := rows.Scan(&sa.ID, &sa.OrgID, &sa.Name, &sa.Description, &sa.ClientID, &sa.ClientSecretHash, &sa.Role, &sa.CreatedAt, &sa.UpdatedAt, &sa.LastUsedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, sa)
	}
	return accounts, nil
}

func (r *AuthRepository) DeleteServiceAccount(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM service_accounts WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}
func (r *AuthRepository) CreateRole(ctx context.Context, role *models.Role) error {
	query := `INSERT INTO roles (id, name, organization_id, description, is_system, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, role.ID, role.Name, role.OrgID, role.Description, role.IsSystem, role.CreatedAt, role.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID, orgID uuid.UUID) error {
	query := `INSERT INTO user_roles (user_id, role_id, organization_id, granted_at)
			  VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, userID, roleID, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) GetUserRoles(ctx context.Context, userID uuid.UUID, orgID uuid.UUID) ([]models.Role, error) {
	query := `SELECT r.id, r.name, uo.org_id, r.description, r.is_system, r.created_at, r.updated_at
			  FROM roles r
			  JOIN user_organizations uo ON r.name = uo.role
			  WHERE uo.user_id = $1 AND uo.org_id = $2`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, userID, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.OrgID, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *AuthRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID, orgID uuid.UUID) ([]string, error) {
	query := `SELECT DISTINCT p.resource || ':' || p.action
			  FROM permissions p
			  JOIN user_organizations uo ON p.role_id = (SELECT id FROM roles WHERE name = uo.role)
			  WHERE uo.user_id = $1 AND uo.org_id = $2`

	start := time.Now()
	rows, err := r.db.Query(ctx, query, userID, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, rows.Err()
}

func (r *AuthRepository) CheckPermission(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, permission string) (bool, error) {
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return false, nil
	}
	query := `SELECT EXISTS (
				SELECT 1
				FROM permissions p
				JOIN user_organizations uo ON p.role_id = (SELECT id FROM roles WHERE name = uo.role)
				WHERE uo.user_id = $1 AND uo.org_id = $2 AND p.resource = $3 AND p.action = $4
			  )`
	start := time.Now()
	var exists bool
	err := r.db.QueryRow(ctx, query, userID, orgID, parts[0], parts[1]).Scan(&exists)
	metrics.RecordDBQuery(query, time.Since(start))
	return exists, err
}

func (r *AuthRepository) GetInviteByCode(ctx context.Context, code string) (*models.Invite, error) {
	query := `SELECT id, code, org_id, role, expiry, created_at, created_by, used FROM invites WHERE code = $1 LIMIT 1`
	start := time.Now()
	var invite models.Invite
	err := r.db.QueryRow(ctx, query, code).Scan(
		&invite.ID, &invite.Code, &invite.OrgID, &invite.Role, &invite.Expiry, &invite.CreatedAt, &invite.CreatedBy, &invite.Used,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *AuthRepository) CreateInvite(ctx context.Context, invite *models.Invite) error {
	query := `
		INSERT INTO invites (id, code, org_id, role, expiry, created_at, created_by, used)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query,
		invite.ID, invite.Code, invite.OrgID, invite.Role,
		invite.Expiry, invite.CreatedAt, invite.CreatedBy, invite.Used,
	)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) MarkInviteUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE invites SET used = true WHERE id = $1`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, id)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) CreateOrganization(ctx context.Context, org *models.Organization) error {
	query := `INSERT INTO organizations (id, name, type, logo_url, domain, default_theme, status, max_users_limit, max_api_keys_limit, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, org.ID, org.Name, org.Type, org.LogoURL, org.Domain, org.DefaultTheme, org.Status, org.MaxUsersLimit, org.MaxApiKeysLimit, org.CreatedAt, org.UpdatedAt)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}

func (r *AuthRepository) RevokeAllOrgTokens(ctx context.Context, orgID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens SET revoked = true, revoked_at = NOW() 
		WHERE user_id IN (SELECT id FROM users WHERE org_id = $1)
	`
	start := time.Now()
	_, err := r.db.Exec(ctx, query, orgID)
	metrics.RecordDBQuery(query, time.Since(start))
	return err
}
