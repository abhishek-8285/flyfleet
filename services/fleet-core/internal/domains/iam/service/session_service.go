package service

import (
	"context"
	"errors"

	redis2 "github.com/flyfleet/fleet-core/internal/domains/iam/redis"
	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SessionService struct {
	sessionRepo *repository.SessionRepository
	redis       *redis2.Client
	log         *zap.Logger
}

func NewSessionService(
	sessionRepo *repository.SessionRepository,
	redis *redis2.Client,
	log *zap.Logger,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		redis:       redis,
		log:         log,
	}
}

// ListSessions returns all active sessions for a user in an organization.
// It merges DB-persisted tokens with Redis-cached session data.
func (s *SessionService) ListSessions(ctx context.Context, userID, tenantID uuid.UUID) ([]models.Session, error) {
	// Get active tokens from DB
	tokens, err := s.sessionRepo.GetActiveTokens(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	sessions := make([]models.Session, 0, len(tokens))
	for _, t := range tokens {
		session := models.Session{
			ID:         t.ID.String(),
			UserID:     t.UserID,
			TenantID:      t.TenantID,
			Device:     t.DeviceInfo,
			IP:         t.IPAddress,
			CreatedAt:  t.CreatedAt,
			LastActive: t.LastActiveAt,
		}

		// Try to enrich from Redis cache
		if data, err := s.redis.GetSession(ctx, t.ID.String()); err == nil && data != nil {
			session.Device = data.Device
			session.IP = data.IP
			session.LastActive = data.LastActive
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// RevokeSession revokes a single session by ID, verifying ownership first.
// Returns "not_owner" error if the session belongs to a different user.
func (s *SessionService) RevokeSession(ctx context.Context, userID, tenantID uuid.UUID, sessionID string) error {
	sUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return err
	}

	// Verify ownership: check that the session belongs to this user and org
	tokens, err := s.sessionRepo.GetActiveTokens(ctx, userID, tenantID)
	if err != nil {
		return err
	}

	var found bool
	for _, t := range tokens {
		if t.ID == sUUID {
			found = true
			break
		}
	}
	if !found {
		// Check if the session exists but belongs to another user
		// We query by session ID to see who owns it
		ownerID, err := s.sessionRepo.GetSessionOwner(ctx, sUUID.String())
		if err == nil && ownerID != "" && ownerID != userID.String() {
			return errors.New("not_owner")
		}
		return errors.New("session_not_found")
	}

	// Revoke in DB
	if err := s.sessionRepo.RevokeToken(ctx, sUUID); err != nil {
		return err
	}

	// Delete from Redis cache
	s.redis.DeleteSession(ctx, sessionID)

	s.log.Info("session_revoked",
		zap.String("session_id", sessionID),
		zap.String("user_id", userID.String()),
	)
	return nil
}

// RevokeAllSessions revokes every session for a user across all organizations.
func (s *SessionService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	if err := s.sessionRepo.RevokeAllTokensEverywhere(ctx, userID); err != nil {
		return err
	}

	s.log.Info("all_sessions_revoked", zap.String("user_id", userID.String()))
	return nil
}
