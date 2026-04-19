package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvitationRepository struct {
	pool *pgxpool.Pool
}

func NewInvitationRepository(pool *pgxpool.Pool) *InvitationRepository {
	return &InvitationRepository{pool: pool}
}

func generateToken() (string, error) {
	b := make([]byte, 48)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (r *InvitationRepository) Create(ctx context.Context, inv *models.Invitation) error {
	token, err := generateToken()
	if err != nil {
		return err
	}
	inv.Token = token

	return r.pool.QueryRow(ctx, `
		INSERT INTO invitations (tenant_id, email, invited_by, token, role_ids, message, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`,
		inv.OrgID, inv.Email, inv.InvitedBy, inv.Token, inv.RoleIDs, inv.Message,
		inv.Status, inv.ExpiresAt,
	).Scan(&inv.ID, &inv.CreatedAt, &inv.UpdatedAt)
}

func (r *InvitationRepository) GetByToken(ctx context.Context, token string) (*models.Invitation, error) {
	var inv models.Invitation
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, invited_by, token, role_ids, message, status,
		       expires_at, accepted_at, accepted_user_id, created_at, updated_at
		FROM invitations WHERE token = $1`, token).Scan(
		&inv.ID, &inv.OrgID, &inv.Email, &inv.InvitedBy, &inv.Token, &inv.RoleIDs,
		&inv.Message, &inv.Status, &inv.ExpiresAt, &inv.AcceptedAt, &inv.AcceptedUserID,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

func (r *InvitationRepository) Accept(ctx context.Context, id, userID uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE invitations SET status = 'accepted', accepted_at = $1, accepted_user_id = $2, updated_at = NOW()
		WHERE id = $3`, now, userID, id)
	return err
}

func (r *InvitationRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE invitations SET status = 'revoked', updated_at = NOW() WHERE id = $1", id)
	return err
}

func (r *InvitationRepository) Resend(ctx context.Context, id uuid.UUID, newExpiry time.Time) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE invitations SET expires_at = $1, updated_at = NOW() WHERE id = $2",
		newExpiry, id)
	return err
}

func (r *InvitationRepository) ListPending(ctx context.Context, orgID uuid.UUID) ([]models.Invitation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, email, invited_by, token, role_ids, message, status,
		       expires_at, accepted_at, accepted_user_id, created_at, updated_at
		FROM invitations WHERE tenant_id = $1 AND status = 'pending' ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []models.Invitation
	for rows.Next() {
		var inv models.Invitation
		if err := rows.Scan(&inv.ID, &inv.OrgID, &inv.Email, &inv.InvitedBy, &inv.Token,
			&inv.RoleIDs, &inv.Message, &inv.Status, &inv.ExpiresAt, &inv.AcceptedAt,
			&inv.AcceptedUserID, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}
	return invites, nil
}

func (r *InvitationRepository) ExistsForEmail(ctx context.Context, orgID uuid.UUID, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM invitations WHERE tenant_id = $1 AND email = $2 AND status = 'pending')`,
		orgID, email).Scan(&exists)
	return exists, err
}
