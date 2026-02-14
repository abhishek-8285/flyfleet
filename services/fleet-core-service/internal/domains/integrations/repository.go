package integrations

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Repository interface {
	Create(ctx context.Context, integration *Integration) error
	GetByID(ctx context.Context, id uuid.UUID) (*Integration, error)
	List(ctx context.Context, limit, offset int) ([]Integration, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewRepository(db *pgxpool.Pool, logger *zap.Logger) Repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

func (r *repository) Create(ctx context.Context, integration *Integration) error {
	r.logger.Info("Stub: Creating integration", zap.String("name", integration.Name))
	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Integration, error) {
	return &Integration{
		ID:        id,
		Name:      "Production Slack",
		Type:      IntegrationTypeSlack,
		Status:    IntegrationStatusActive,
		CreatedAt: time.Now().Add(-24 * time.Hour),
		Config:    map[string]string{"webhook_url": "https://hooks.slack.com/..."},
	}, nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]Integration, error) {
	return []Integration{
		{
			ID:        uuid.New(),
			Name:      "Production Slack",
			Type:      IntegrationTypeSlack,
			Status:    IntegrationStatusActive,
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			ID:        uuid.New(),
			Name:      "Ops Jira",
			Type:      IntegrationTypeJira,
			Status:    IntegrationStatusInactive,
			CreatedAt: time.Now().Add(-48 * time.Hour),
		},
	}, nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Info("Stub: Deleting integration", zap.String("id", id.String()))
	return nil
}
