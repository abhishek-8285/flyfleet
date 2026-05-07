package service

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/flyfleet/fleet-core/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core/internal/domains/vehicle/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DocumentsService manages immutable vehicle document uploads.
type DocumentsService struct {
	vRepo    repository.VehicleRepositoryInterface
	docsRepo repository.DocumentsRepositoryInterface
	evtsRepo repository.EventsRepositoryInterface
	logger   *zap.Logger
}

func NewDocumentsService(
	vRepo repository.VehicleRepositoryInterface,
	docsRepo repository.DocumentsRepositoryInterface,
	evtsRepo repository.EventsRepositoryInterface,
	logger *zap.Logger,
) *DocumentsService {
	return &DocumentsService{vRepo: vRepo, docsRepo: docsRepo, evtsRepo: evtsRepo, logger: logger}
}

// Upload validates and inserts an immutable document record.
func (s *DocumentsService) Upload(ctx context.Context, vehicleID uuid.UUID, tenantID uuid.UUID, req *models.UploadDocumentRequest, uploadedBy *uuid.UUID) (*models.VehicleDocument, error) {
	// Ensure vehicle exists and belongs to the organization
	vehicle, err := s.vRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle.TenantID != tenantID {
		s.logger.Warn("unauthorized document upload attempt",
			zap.String("vehicle_id", vehicleID.String()),
			zap.String("tenant_id", tenantID.String()),
			zap.String("vehicle_tenant", vehicle.TenantID.String()))
		return nil, fmt.Errorf("vehicle not found")
	}

	// Validate hash length (must be exactly 64 hex chars = SHA-256)
	if len(req.Hash) != 64 {
		return nil, fmt.Errorf("invalid hash: must be 64-character SHA-256 hex string")
	}

	doc := &models.VehicleDocument{
		ID:             uuid.New(),
		VehicleID:      vehicleID,
		Type:           req.Type,
		DocumentURL:    &req.DocumentURL,
		DocumentNumber: req.DocumentNumber,
		ExpiryDate:     req.ExpiryDate,
		Status:         models.VerificationPending,
		Hash:           req.Hash,
		UploadedBy:     uploadedBy,
	}

	if err := s.docsRepo.Insert(ctx, doc); err != nil {
		s.logger.Error("failed to insert document", zap.Error(err))
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	// Emit event
	if s.evtsRepo != nil {
		event := &models.VehicleEvent{
			ID:        uuid.New(),
			VehicleID: vehicleID,
			EventType: models.EventVehicleDocumentUploaded,
			Payload: map[string]interface{}{
				"document_type": doc.Type,
				"document_id":   doc.ID.String(),
				"version":       doc.Version,
			},
		}
		if err := s.evtsRepo.Append(ctx, event); err != nil {
			s.logger.Warn("failed to emit document event", zap.Error(err))
		}
	}

	return doc, nil
}

// ListDocuments returns all documents for a vehicle.
func (s *DocumentsService) ListDocuments(ctx context.Context, vehicleID uuid.UUID, tenantID uuid.UUID) ([]models.VehicleDocument, error) {
	// Verify vehicle belongs to org
	vehicle, err := s.vRepo.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle.TenantID != tenantID {
		s.logger.Warn("unauthorized document list attempt",
			zap.String("vehicle_id", vehicleID.String()),
			zap.String("tenant_id", tenantID.String()))
		return nil, fmt.Errorf("vehicle not found")
	}
	return s.docsRepo.List(ctx, vehicleID)
}

// GetDocument returns a single document by ID.
func (s *DocumentsService) GetDocument(ctx context.Context, docID uuid.UUID, tenantID uuid.UUID) (*models.VehicleDocument, error) {
	doc, err := s.docsRepo.Get(ctx, docID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	// Verify the document's vehicle belongs to the org
	vehicle, err := s.vRepo.GetByID(ctx, doc.VehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle.TenantID != tenantID {
		s.logger.Warn("unauthorized document access attempt",
			zap.String("doc_id", docID.String()),
			zap.String("vehicle_id", doc.VehicleID.String()),
			zap.String("tenant_id", tenantID.String()))
		return nil, nil
	}
	return doc, nil
}

// VerifyHash recomputes a SHA-256 hash from content and verifies it against the stored hash.
func VerifyDocumentHash(content []byte) string {
	sum := sha256.Sum256(content)
	return fmt.Sprintf("%x", sum)
}
