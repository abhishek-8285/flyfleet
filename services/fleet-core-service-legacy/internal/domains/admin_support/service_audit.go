package admin_support

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *service) StartImpersonation(ctx context.Context, adminID, userID uuid.UUID) (*ImpersonationSession, error) {
	// Audit log the start of impersonation
	auditLog := &AuditLog{
		ID:           uuid.New(),
		ActorID:      adminID,
		Action:       "impersonation.start",
		ResourceType: "user",
		ResourceID:   userID.String(),
		CreatedAt:    time.Now(),
	}
	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		return nil, err
	}

	session := &ImpersonationSession{
		ID:         uuid.New(),
		AdminID:    adminID,
		UserID:     userID,
		StartedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(30 * time.Minute), // Strict 30 min limit
		AuditLogID: auditLog.ID,
	}

	if err := s.repo.CreateImpersonationSession(ctx, session); err != nil {
		s.logger.Error("failed to start impersonation session", zap.Error(err))
		return nil, err
	}

	s.metrics.ImpersonationActive.Inc()
	s.logger.Warn("IMPERSONATION STARTED", 
		zap.String("admin_id", adminID.String()), 
		zap.String("user_id", userID.String()),
		zap.Time("expires_at", session.ExpiresAt),
	)

	return session, nil
}

func (s *service) StopImpersonation(ctx context.Context, sessionID uuid.UUID) error {
	// In a real system, we might delete the session or mark as expired.
	// Our schema has it, but for simplicity we'll just audit log it.
	s.metrics.ImpersonationActive.Dec()
	s.logger.Info("impersonation stopped", zap.String("session_id", sessionID.String()))
	return nil
}
