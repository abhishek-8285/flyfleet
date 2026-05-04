package service

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/document/models"
	"github.com/flyfleet/fleet-core/internal/domains/document/repository"
	"github.com/flyfleet/fleet-core/internal/pkg/storage"
	"github.com/google/uuid"
)

type DocumentService struct {
	repo    *repository.DocumentRepository
	storage storage.StorageProvider
}

func NewDocumentService(repo *repository.DocumentRepository, storage storage.StorageProvider) *DocumentService {
	return &DocumentService{
		repo:    repo,
		storage: storage,
	}
}

func (s *DocumentService) Upload(ctx context.Context, tenantID, userID uuid.UUID, req *models.UploadDocumentRequest, fileName string, fileSize int64, mimeType string, reader io.Reader) (*models.Document, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("storage provider not initialized")
	}

	// Generate a unique path: tenant/entity_type/entity_id/uuid_filename
	docID := uuid.New()
	ext := filepath.Ext(fileName)
	storagePath := fmt.Sprintf("%s/%s/%s/%s%s", 
		tenantID, 
		req.EntityType, 
		req.EntityID, 
		docID, 
		ext,
	)

	// Save to storage
	path, err := s.storage.Save(ctx, storagePath, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to save file to storage: %w", err)
	}

	// Create metadata record
	doc := &models.Document{
		ID:           docID,
		TenantID:     tenantID,
		EntityType:   req.EntityType,
		EntityID:     req.EntityID,
		DocumentType: req.DocumentType,
		FileName:     fileName,
		FilePath:     path,
		FileSize:     fileSize,
		MimeType:     mimeType,
		ExpiryDate:   req.ExpiryDate,
		Status:       models.DocStatusPending,
		UploadedBy:   &userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		// Cleanup storage on DB failure
		_ = s.storage.Delete(ctx, path)
		return nil, fmt.Errorf("failed to save document metadata: %w", err)
	}

	return doc, nil
}

func (s *DocumentService) GetDocuments(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.Document, error) {
	return s.repo.GetByEntity(ctx, entityType, entityID)
}

func (s *DocumentService) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	// In a real system, we might want to delete from storage too, 
	// but soft-delete in DB is safer for auditing.
	return s.repo.Delete(ctx, id)
}
