package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth_v2/models"
	"github.com/google/uuid"
)

type AuthRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewAuthRepository(db *sql.DB, logger logger.Logger) *AuthRepository {
	return &AuthRepository{db: db, logger: logger}
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := "SELECT id, org_id, email, role, password_hash, created_at, created_by, updated_at, updated_by, mfa_enabled, mfa_methods, last_password_change FROM users WHERE email = $1 AND is_deleted = false"
	
	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.OrgID, &user.Email, &user.Role, &user.PasswordHash,
		&user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
		&user.MfaEnabled, &user.MfaMethods, &user.LastPasswordChange,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	
	return &user, nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, org_id, email, password_hash, role, status, created_at, created_by, updated_at, updated_by, mfa_enabled, mfa_methods, last_password_change)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.OrgID, user.Email, user.PasswordHash, user.Role, "active",
		user.CreatedAt, user.CreatedBy, user.UpdatedAt, user.UpdatedBy,
		user.MfaEnabled, user.MfaMethods, user.LastPasswordChange,
	)
	return err
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := "SELECT id, org_id, email, role, created_at, created_by, updated_at, updated_by, mfa_enabled, mfa_methods, last_password_change FROM users WHERE id = $1 AND is_deleted = false"
	
	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.OrgID, &user.Email, &user.Role,
		&user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
		&user.MfaEnabled, &user.MfaMethods, &user.LastPasswordChange,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	return &user, nil
}

func (r *AuthRepository) LogSecurityEvent(ctx context.Context, event *models.SecurityEvent) error {
	query := "INSERT INTO security_events (id, user_id, event_type, ip_address, user_agent, metadata) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := r.db.ExecContext(ctx, query, uuid.New(), event.UserID, event.EventType, event.IPAddress, event.UserAgent, event.Metadata)
	return err
}
