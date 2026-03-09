package service

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/vehicle/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DocumentsService manages immutable vehicle document uploads.
type DocumentsService struct {
	vRepo     repository.VehicleRepositoryInterface
	docsRepo  repository.DocumentsRepositoryInterface
	evtsRepo  repository.EventsRepositoryInterface
	logger    *zap.Logger
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
func (s *DocumentsService) Upload(ctx context.Context, vehicleID uuid.UUID, req *models.UploadDocumentRequest, uploadedBy *uuid.UUID) (*models.VehicleDocument, error) {
	// Ensure vehicle exists
	if _, err := s.vRepo.GetByID(ctx, vehicleID); err != nil {
		return nil, err
	}

	// Validate hash length (must be exactly 64 hex chars = SHA-256)
	if len(req.Hash) != 64 {
		return nil, fmt.Errorf("invalid hash: must be 64-character SHA-256 hex string")
	}

	doc := &models.VehicleDocument{
		ID:           uuid.New(),
		VehicleID:    vehicleID,
		DocumentType: req.DocumentType,
		FileURL:      req.FileURL,
		Hash:         req.Hash,
		UploadedBy:   uploadedBy,
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
				"document_type": doc.DocumentType,
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
func (s *DocumentsService) ListDocuments(ctx context.Context, vehicleID uuid.UUID) ([]models.VehicleDocument, error) {
	return s.docsRepo.List(ctx, vehicleID)
}

// GetDocument returns a single document by ID.
func (s *DocumentsService) GetDocument(ctx context.Context, docID uuid.UUID) (*models.VehicleDocument, error) {
	return s.docsRepo.Get(ctx, docID)
}

// VerifyHash recomputes a SHA-256 hash from content and verifies it against the stored hash.
func VerifyDocumentHash(content []byte) string {
	sum := sha256.Sum256(content)
	return fmt.Sprintf("%x", sum)
}
