package service

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/auth_v2/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth_v2/repository"
	"github.com/google/uuid"
)

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *AuthService) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*repository.Organization, error) {
	return s.repo.GetOrganizationByID(ctx, id)
}

func (s *AuthService) ListOrganizations(ctx context.Context, limit, offset int32) ([]repository.Organization, int64, error) {
	return s.repo.ListOrganizations(ctx, limit, offset)
}

func (s *AuthService) GetStats(ctx context.Context) (*repository.UserStats, error) {
	return s.repo.GetStats(ctx)
}

func (s *AuthService) GetAuditLogs(ctx context.Context, limit, offset int32) ([]repository.AuditLog, int64, error) {
	return s.repo.GetAuditLogs(ctx, limit, offset)
}

func (s *AuthService) GetSecurityTelemetry(ctx context.Context) (map[string]any, error) {
	return s.repo.GetSecurityTelemetry(ctx)
}
