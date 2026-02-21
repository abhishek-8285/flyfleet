package integrations

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	CreateIntegration(ctx context.Context, req CreateIntegrationRequest) (*Integration, error)
	GetIntegration(ctx context.Context, id uuid.UUID) (*Integration, error)
	ListIntegrations(ctx context.Context, limit, offset int) ([]Integration, error)
	DeleteIntegration(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo   Repository
	logger *zap.Logger
}

func NewService(repo Repository, logger *zap.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

func (s *service) CreateIntegration(ctx context.Context, req CreateIntegrationRequest) (*Integration, error) {
	integration := &Integration{
		ID:        uuid.New(),
		Name:      req.Name,
		Type:      req.Type,
		Config:    req.Config,
		Status:    IntegrationStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, integration); err != nil {
		return nil, err
	}

	return integration, nil
}

func (s *service) GetIntegration(ctx context.Context, id uuid.UUID) (*Integration, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListIntegrations(ctx context.Context, limit, offset int) ([]Integration, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *service) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
