package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (r *SessionRepository) StoreToken(ctx context.Context, token *models.AuthToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO auth_tokens (user_id, org_id, token_hash, token_type, device_id, device_info,
		                         ip_address, user_agent, expires_at, last_active_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		token.UserID, token.OrgID, hashToken(token.TokenHash), token.TokenType,
		token.DeviceID, token.DeviceInfo, token.IPAddress, token.UserAgent,
		token.ExpiresAt, token.LastActiveAt,
	)
	return err
}

func (r *SessionRepository) GetActiveTokens(ctx context.Context, userID, orgID uuid.UUID) ([]models.AuthToken, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, org_id, device_id, device_info, ip_address, user_agent,
		       expires_at, last_active_at, created_at
		FROM auth_tokens
		WHERE user_id = $1 AND org_id = $2 AND revoked_at IS NULL AND expires_at > NOW()
		ORDER BY last_active_at DESC`, userID, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.AuthToken
	for rows.Next() {
		var t models.AuthToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.OrgID, &t.DeviceID, &t.DeviceInfo,
			&t.IPAddress, &t.UserAgent, &t.ExpiresAt, &t.LastActiveAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func (r *SessionRepository) RevokeToken(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, "UPDATE auth_tokens SET revoked_at = $1 WHERE id = $2", now, id)
	return err
}

func (r *SessionRepository) RevokeAllUserTokens(ctx context.Context, userID, orgID uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		"UPDATE auth_tokens SET revoked_at = $1 WHERE user_id = $2 AND org_id = $3",
		now, userID, orgID,
	)
	return err
}

func (r *SessionRepository) RevokeAllTokensEverywhere(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx,
		"UPDATE auth_tokens SET revoked_at = $1 WHERE user_id = $2",
		now, userID,
	)
	return err
}

func (r *SessionRepository) UpdateLastActive(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE auth_tokens SET last_active_at = NOW() WHERE id = $1", id,
	)
	return err
}

// GetSessionOwner returns the user ID that owns the given session.
func (r *SessionRepository) GetSessionOwner(ctx context.Context, sessionID string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx,
		"SELECT user_id FROM sessions WHERE id = $1", sessionID,
	).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}
