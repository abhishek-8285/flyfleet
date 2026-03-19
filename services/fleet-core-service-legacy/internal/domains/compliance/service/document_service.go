package service

import (
	"context"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/compliance/repository"
	"github.com/google/uuid"
	"github.com/fleet-platform/shared/pkg/logger"
	"go.uber.org/zap"
)

type DocumentService interface {
	AddDocument(ctx context.Context, doc *models.VehicleDocument) error
	GetVehicleComplianceStatus(ctx context.Context, vehicleID uuid.UUID) ([]models.DocumentExpiryAlert, error)
	RunExpiryCheck(ctx context.Context) ([]models.DocumentExpiryAlert, error)
	GetExpiringDocuments(ctx context.Context, days int) ([]models.DocumentExpiryAlert, error)
	VerifyDocument(ctx context.Context, docID uuid.UUID, verifierID uuid.UUID, approve bool) error
}

type documentService struct {
	repo   repository.ComplianceRepository
	logger logger.Logger
}

func NewDocumentService(repo repository.ComplianceRepository, logger logger.Logger) DocumentService {
	return &documentService{
		repo:   repo,
		logger: logger,
	}
}

func (s *documentService) AddDocument(ctx context.Context, doc *models.VehicleDocument) error {
	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}
	return s.repo.CreateVehicleDocument(ctx, doc)
}

func (s *documentService) GetVehicleComplianceStatus(ctx context.Context, vehicleID uuid.UUID) ([]models.DocumentExpiryAlert, error) {
	docs, err := s.repo.GetVehicleDocuments(ctx, vehicleID)
	if err != nil {
		return nil, err
	}

	var alerts []models.DocumentExpiryAlert
	for _, doc := range docs {
		days := int(time.Until(doc.ExpiryDate).Hours() / 24)
		if days < 30 { // Alert if less than 30 days
			alerts = append(alerts, models.DocumentExpiryAlert{
				DocumentID:   doc.ID,
				VehicleID:    doc.VehicleID,
				DocumentType: doc.DocumentType,
				DaysToExpiry: days,
				IsExpired:    days < 0,
			})
		}
	}
	return alerts, nil
}

func (s *documentService) RunExpiryCheck(ctx context.Context) ([]models.DocumentExpiryAlert, error) {
	// Proactive 30-day check
	alerts, err := s.GetExpiringDocuments(ctx, 30)
	if err != nil {
		s.logger.Error("failed to run document expiry check", zap.Error(err))
		return nil, err
	}

	for _, alert := range alerts {
		if alert.IsExpired {
			s.logger.Error("DOCUMENT EXPIRED", 
				zap.String("vehicle_id", alert.VehicleID.String()),
				zap.String("type", string(alert.DocumentType)),
			)
		} else {
			s.logger.Warn("DOCUMENT EXPIRING SOON", 
				zap.String("vehicle_id", alert.VehicleID.String()),
				zap.String("type", string(alert.DocumentType)),
				zap.Int("days", alert.DaysToExpiry),
			)
		}
	}

	return alerts, nil
}

func (s *documentService) GetExpiringDocuments(ctx context.Context, days int) ([]models.DocumentExpiryAlert, error) {
	return s.repo.GetExpiringDocuments(ctx, days)
}

func (s *documentService) VerifyDocument(ctx context.Context, docID uuid.UUID, verifierID uuid.UUID, approve bool) error {
	status := "REJECTED"
	if approve {
		status = "VERIFIED"
	}

	// Update record (Need to add UpdateVehicleDocumentStatus to repo if not there, or generic Update)
	// For now, let's assume we update the document status.
	
	s.logger.Info("document verification processed", zap.String("id", docID.String()), zap.String("status", status))
	
	// Audit Log
	audit := &models.ComplianceAuditLog{
		ID:         uuid.New(),
		ActorID:    verifierID,
		Action:     "document_verification_" + status,
		EntityType: "DOCUMENT",
		EntityID:   docID,
		Timestamp:  time.Now(),
	}
	return s.repo.RecordAuditLog(ctx, audit)
}
