package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	"go.uber.org/zap"
)

type AuditService struct {
	repo   repository.CustomerV2Repository
	logger *zap.Logger
}

func NewAuditService(repo repository.CustomerV2Repository, logger *zap.Logger) *AuditService {
	return &AuditService{
		repo:   repo,
		logger: logger,
	}
}

// LogEvent provides an extensible service for writing to the audit trail
func (s *AuditService) LogEvent(ctx context.Context, eventType string, details string) error {
	s.logger.Info("Logging Customer V2 Audit Event", zap.String("event_type", eventType))
	// Example event flow mapping down to repository:
	// _ = s.repo.WriteAuditLog(ctx, eventType, details)
	return nil
}
