package repository

import (
	"context"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/google/uuid"
)

type GovernanceRepository interface {
	CreateRequest(ctx context.Context, req *models.GovernanceRequest) error
	GetRequestByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.GovernanceRequest, error)
	ListRequests(ctx context.Context, filter models.GovernanceRequestFilter) (*models.GovernanceRequestListResponse, error)
	UpdateRequestStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.GovernanceRequestStatus, comments *string) error
	
	GetSettings(ctx context.Context, orgID uuid.UUID) (*models.GovernanceSettings, error)
	UpsertSettings(ctx context.Context, settings *models.GovernanceSettings) error
	
	CreateAuditLog(ctx context.Context, log *models.GovernanceAuditLog) error
	ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]models.GovernanceAuditLog, int, error)
}
