package repository

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/governance/models"
	"github.com/google/uuid"
)

type GovernanceRepository interface {
	CreateRequest(ctx context.Context, req *models.GovernanceRequest) error
	GetRequestByID(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*models.GovernanceRequest, error)
	GetRequestByIDInternal(ctx context.Context, id uuid.UUID) (*models.GovernanceRequest, error)
	ListRequests(ctx context.Context, filter models.GovernanceRequestFilter) (*models.GovernanceRequestListResponse, error)
	UpdateRequestStatus(ctx context.Context, id uuid.UUID, orgID uuid.UUID, status models.GovernanceRequestStatus, comments *string, expiryDate *time.Time) error
	UpdateVerificationStatus(ctx context.Context, id, orgID uuid.UUID, status string, verifiedBy uuid.UUID) error

	GetSettings(ctx context.Context, orgID uuid.UUID) (*models.GovernanceSettings, error)
	UpsertSettings(ctx context.Context, settings *models.GovernanceSettings) error

	CreateAuditLog(ctx context.Context, log *models.GovernanceAuditLog) error
	ListAuditLogs(ctx context.Context, orgID uuid.UUID, page, pageSize int) ([]models.GovernanceAuditLog, int, error)

	ExpireRequests(ctx context.Context) ([]models.ExpiredRequestInfo, error)

	CreateExport(ctx context.Context, export *models.GovernanceExport) error
	GetExportByRequestID(ctx context.Context, requestID uuid.UUID) (*models.GovernanceExport, error)
	CreateDownloadToken(ctx context.Context, token *models.DownloadToken) error
	GetExportByToken(ctx context.Context, token uuid.UUID) (*models.GovernanceExport, error)

	CheckLegalHold(ctx context.Context, entityID uuid.UUID) (bool, error)
	SaveExportLineage(ctx context.Context, lineage *models.ExportLineage) error
	SaveComplianceReport(ctx context.Context, report *models.ComplianceReport) error
	ListAllSettings(ctx context.Context) ([]models.GovernanceSettings, error)
	EnforceRetentionPolicy(ctx context.Context, orgID uuid.UUID, tableName string, retentionDays int) (int64, error)
}
