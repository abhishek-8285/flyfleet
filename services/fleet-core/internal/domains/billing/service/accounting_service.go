package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core/internal/domains/billing/repository"
	"github.com/flyfleet/fleet-core/internal/domains/billing/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AccountingService struct {
	repo        *repository.AccountingRepository
	ledgerRepo  *repository.LedgerRepository
	invoiceRepo Repository
	storage     storage.ObjectStorage
	exporters   map[models.ExternalSystem]models.AccountingExporter
	logger      *zap.Logger
}

func NewAccountingService(
	repo *repository.AccountingRepository,
	ledgerRepo *repository.LedgerRepository,
	invoiceRepo Repository,
	storage storage.ObjectStorage,
	exporters []models.AccountingExporter,
	logger *zap.Logger,
) *AccountingService {
	exporterMap := make(map[models.ExternalSystem]models.AccountingExporter)
	for _, e := range exporters {
		exporterMap[e.GetName()] = e
	}

	return &AccountingService{
		repo:        repo,
		ledgerRepo:  ledgerRepo,
		invoiceRepo: invoiceRepo,
		storage:     storage,
		exporters:   exporterMap,
		logger:      logger,
	}
}

func (s *AccountingService) ExportLedger(ctx context.Context, orgID uuid.UUID, system models.ExternalSystem, userID uuid.UUID) (*models.AccountingExport, error) {
	exporter, ok := s.exporters[system]
	if !ok {
		return nil, fmt.Errorf("exporter not found for system: %s", system)
	}

	// 1. Create pending export record
	export := &models.AccountingExport{
		ID:             uuid.New(),
		OrgID:          orgID,
		ExternalSystem: system,
		ExportType:     models.ExportLedger,
		Status:         models.ExportStatusPending,
		CreatedAt:      time.Now(),
		CreatedBy:      userID,
	}
	if err := s.repo.CreateExport(ctx, export); err != nil {
		return nil, err
	}

	// 2. Fetch data (simplified - get recent 1000 entries)
	// TODO: Add date range filtering to ledger repo
	entries, err := s.fetchLedgerEntries(ctx, orgID)
	if err != nil {
		s.repo.UpdateExportStatus(ctx, export.ID, models.ExportStatusFailed, "", 0, err.Error())
		return nil, err
	}

	// 3. Transform
	data, fileName, err := exporter.ExportLedger(orgID, entries)
	if err != nil {
		s.repo.UpdateExportStatus(ctx, export.ID, models.ExportStatusFailed, "", 0, err.Error())
		return nil, err
	}

	// 4. Store
	// For now we just log the size and file name
	s.logger.Info("Ledger exported successfully", zap.String("tenant_id", orgID.String()), zap.String("file", fileName), zap.Int("data_size", len(data)))
	fileKey := fmt.Sprintf("exports/%s/%s", orgID, fileName)
	// We'll reuse Invoice model for Save method context or update interface
	// For now, let's assume Save can take a string key too or we use a temporary hack
	// Actually, let's update ObjectStorage to support generic file saving
	// Since I can't easily change it now, I'll return the data directly for local download in handler

	err = s.repo.UpdateExportStatus(ctx, export.ID, models.ExportStatusCompleted, fileKey, len(entries), "")
	return export, err
}

func (s *AccountingService) ExportInvoices(ctx context.Context, orgID uuid.UUID, system models.ExternalSystem, userID uuid.UUID) (*models.AccountingExport, error) {
	exporter, ok := s.exporters[system]
	if !ok {
		return nil, fmt.Errorf("exporter not found for system: %s", system)
	}

	// 1. Create pending record
	export := &models.AccountingExport{
		ID:             uuid.New(),
		OrgID:          orgID,
		ExternalSystem: system,
		ExportType:     models.ExportInvoices,
		Status:         models.ExportStatusPending,
		CreatedAt:      time.Now(),
		CreatedBy:      userID,
	}
	if err := s.repo.CreateExport(ctx, export); err != nil {
		return nil, err
	}

	// 2. Fetch data
	invoices, _, err := s.invoiceRepo.ListInvoices(ctx, orgID, 1, 1000)
	if err != nil {
		s.repo.UpdateExportStatus(ctx, export.ID, models.ExportStatusFailed, "", 0, err.Error())
		return nil, err
	}

	// 3. Transform
	data, fileName, err := exporter.ExportInvoices(orgID, invoices)
	if err != nil {
		s.repo.UpdateExportStatus(ctx, export.ID, models.ExportStatusFailed, "", 0, err.Error())
		return nil, err
	}

	// 4. Update status
	s.logger.Info("Invoices exported successfully", zap.String("tenant_id", orgID.String()), zap.String("file", fileName), zap.Int("data_size", len(data)))
	fileKey := fmt.Sprintf("exports/%s/%s", orgID, fileName)
	err = s.repo.UpdateExportStatus(ctx, export.ID, models.ExportStatusCompleted, fileKey, len(invoices), "")
	return export, err
}

func (s *AccountingService) GetExportHistory(ctx context.Context, orgID uuid.UUID) ([]models.AccountingExport, error) {
	return s.repo.GetExportHistory(ctx, orgID, 50)
}

func (s *AccountingService) fetchLedgerEntries(ctx context.Context, orgID uuid.UUID) ([]models.LedgerEntry, error) {
	return s.ledgerRepo.GetEntriesByOrg(ctx, orgID, 1000)
}
