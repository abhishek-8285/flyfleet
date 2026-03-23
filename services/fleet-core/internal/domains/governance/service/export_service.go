package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/governance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/governance/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/governance/utils"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type ExportService interface {
	ProcessExportRequest(ctx context.Context, requestID uuid.UUID) (*models.GovernanceExport, error)
	GenerateDownloadToken(ctx context.Context, exportID uuid.UUID) (*models.DownloadToken, error)
	ValidateToken(ctx context.Context, token uuid.UUID) (*models.GovernanceExport, error)
}

type exportService struct {
	repo       repository.GovernanceRepository
	aggregator GovernanceDataAggregator
	masker     MaskingService
	generator  utils.ExportFileGenerator
	nats       *nats.Conn
	logger     *zap.Logger
}

func NewExportService(
	repo repository.GovernanceRepository,
	aggregator GovernanceDataAggregator,
	masker MaskingService,
	generator utils.ExportFileGenerator,
	nats *nats.Conn,
	logger *zap.Logger,
) ExportService {
	return &exportService{
		repo:       repo,
		aggregator: aggregator,
		masker:     masker,
		generator:  generator,
		nats:       nats,
		logger:     logger,
	}
}

func (s *exportService) ProcessExportRequest(ctx context.Context, requestID uuid.UUID) (*models.GovernanceExport, error) {
	startTime := time.Now()
	req, err := s.repo.GetRequestByIDInternal(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, fmt.Errorf("request not found")
	}

	if req.Status != models.RequestStatusApproved {
		return nil, fmt.Errorf("request not approved")
	}

	// Update status to PROCESSING
	s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusProcessing, nil, nil)

	s.logger.Info("processing export request", zap.String("id", requestID.String()))

	// 2. Aggregate Data
	data, err := s.aggregator.AggregateData(ctx, req.RequesterID, req.OrganizationID)
	if err != nil {
		s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusFailed, nil, nil)
		return nil, err
	}

	// 3. Apply Masking
	settings, _ := s.repo.GetSettings(ctx, req.OrganizationID)
	if settings != nil {
		data = s.masker.MaskData(data, settings.MaskingRules)
	}

	// 4. Generate File
	storagePath := fmt.Sprintf("/tmp/exports/%s/dsar_%s.%s", req.OrganizationID, requestID, strings.ToLower(req.Format))
	var size int64
	var checksum string
	
	switch strings.ToUpper(req.Format) {
	case "CSV":
		size, checksum, err = s.generator.GenerateCSV(data, storagePath)
	case "JSON":
		size, checksum, err = s.generator.GenerateJSON(data, storagePath)
	case "PDF":
		size, checksum, err = s.generator.GeneratePDF(data, storagePath)
	default:
		size, checksum, err = s.generator.GenerateJSON(data, storagePath)
	}

	if err != nil {
		s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusFailed, nil, nil)
		return nil, err
	}

	// 5. Store Metadata
	export := &models.GovernanceExport{
		ID:        uuid.New(),
		RequestID: requestID,
		FilePath:  storagePath,
		FileSize:  size,
		Checksum:  checksum,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // Metadata expires, file stays?
	}

	if err := s.repo.CreateExport(ctx, export); err != nil {
		return nil, err
	}

	// Phase 5: Export Data Lineage
	lineage := &models.ExportLineage{
		ID:              uuid.New(),
		RequestID:       requestID,
		TableName:       "aggregated_data",
		RecordCount:     1, // Placeholder
		ExportSizeBytes: size,
		CreatedAt:       time.Now(),
	}
	_ = s.repo.SaveExportLineage(ctx, lineage)

	// Phase 6: Compliance Report Storage
	report := &models.ComplianceReport{
		ID:               uuid.New(),
		RequestID:        requestID,
		ExportChecksum:   checksum,
		ExportSizeBytes:  size,
		ProcessingTimeMs: int(time.Since(startTime).Milliseconds()),
		TablesAccessed:   json.RawMessage(`["users", "vehicles", "telemetry", "trips", "fuel_logs"]`),
		CreatedAt:        time.Now(),
	}
	_ = s.repo.SaveComplianceReport(ctx, report)

	// Phase 9: Security Alert for Large Exports (> 100MB)
	if size > 100*1024*1024 && s.nats != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"alert_type":      "LARGE_EXPORT_DETECTED",
			"organization_id": req.OrganizationID.String(),
			"user_id":         req.RequesterID.String(),
			"timestamp":       time.Now().Unix(),
		})
		s.nats.Publish("governance.security.alert", payload)
	}

	// Update Request Status to COMPLETED
	s.repo.UpdateRequestStatus(ctx, requestID, req.OrganizationID, models.RequestStatusCompleted, nil, nil)
	
	s.publishEvent("governance.export.generated", export)

	return export, nil
}

func (s *exportService) publishEvent(subject string, data interface{}) {
	if s.nats == nil {
		return
	}
	payload, _ := json.Marshal(data)
	s.nats.Publish(subject, payload)
}

func (s *exportService) GenerateDownloadToken(ctx context.Context, exportID uuid.UUID) (*models.DownloadToken, error) {
	token := &models.DownloadToken{
		Token:     uuid.New(),
		ExportID:  exportID,
		ExpiresAt: time.Now().Add(4 * time.Hour), // Default link TTL
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateDownloadToken(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *exportService) ValidateToken(ctx context.Context, token uuid.UUID) (*models.GovernanceExport, error) {
	export, err := s.repo.GetExportByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if export == nil {
		return nil, fmt.Errorf("token invalid or expired")
	}
	return export, nil
}
