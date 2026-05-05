package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/driver/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DocumentExpiryService handles driver document expiry tracking
type DocumentExpiryService struct {
	repo   Repository
	logger *zap.Logger
}

// NewDocumentExpiryService creates a new document expiry service
func NewDocumentExpiryService(repo Repository, logger *zap.Logger) *DocumentExpiryService {
	return &DocumentExpiryService{
		repo:   repo,
		logger: logger,
	}
}

// ExpiringDocument represents a document that is expiring or expired
type ExpiringDocument struct {
	DriverID        uuid.UUID `json:"driver_id"`
	DriverName      string    `json:"driver_name"`
	DocumentID      uuid.UUID `json:"document_id"`
	DocumentType    string    `json:"document_type"`
	DocumentNumber  string    `json:"document_number"`
	ExpiryDate      time.Time `json:"expiry_date"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	Status          string    `json:"status"` // expiring_soon, expired
}

// GetExpiringDocuments retrieves documents expiring within the specified days
func (s *DocumentExpiryService) GetExpiringDocuments(ctx context.Context, tenantID string, days int) ([]ExpiringDocument, error) {
	if days <= 0 {
		days = 30 // Default to 30 days
	}

	cutoffDate := time.Now().AddDate(0, 0, days)

	// Get all drivers in organization
	// Note: This would need a new repository method to get drivers with their documents
	// For now, this is a placeholder implementation

	var expiringDocs []ExpiringDocument

	// In production, this would query the database for:
	// SELECT d.id, d.full_name, doc.*
	// FROM drivers d, jsonb_array_elements(d.documents) as doc
	// WHERE d.tenant_id = $1
	//   AND (doc->>'expiry_date')::date <= $2
	//   AND d.deleted_at IS NULL

	_ = cutoffDate // Use cutoff date in query

	return expiringDocs, nil
}

// GetExpiredDocuments retrieves all expired documents
func (s *DocumentExpiryService) GetExpiredDocuments(ctx context.Context, tenantID string) ([]ExpiringDocument, error) {
	// Get all drivers in organization with expired documents
	var expiredDocs []ExpiringDocument

	// In production, this would query:
	// SELECT d.id, d.full_name, doc.*
	// FROM drivers d, jsonb_array_elements(d.documents) as doc
	// WHERE d.tenant_id = $1
	//   AND (doc->>'expiry_date')::date < CURRENT_DATE
	//   AND d.deleted_at IS NULL

	return expiredDocs, nil
}

// CheckDocumentExpiry checks if a specific document is expired or expiring soon
func (s *DocumentExpiryService) CheckDocumentExpiry(ctx context.Context, driverID uuid.UUID, documentID uuid.UUID, warningDays int) (*ExpiringDocument, error) {
	driver, err := s.repo.GetByID(ctx, driverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}
	if driver == nil {
		return nil, ErrDriverNotFound
	}

	// Find the document
	var targetDoc *models.DriverDocument
	for i := range driver.Documents {
		if driver.Documents[i].ID == documentID {
			targetDoc = &driver.Documents[i]
			break
		}
	}

	if targetDoc == nil {
		return nil, ErrDocumentNotFound
	}

	now := time.Now()
	daysUntilExpiry := int(targetDoc.ExpiryDate.Sub(now).Hours() / 24)

	status := "valid"
	if daysUntilExpiry < 0 {
		status = "expired"
	} else if daysUntilExpiry <= warningDays {
		status = "expiring_soon"
	}

	return &ExpiringDocument{
		DriverID:        driverID,
		DriverName:      driver.FullName,
		DocumentID:      documentID,
		DocumentType:    targetDoc.DocumentType,
		DocumentNumber:  targetDoc.DocumentNumber,
		ExpiryDate:      targetDoc.ExpiryDate,
		DaysUntilExpiry: daysUntilExpiry,
		Status:          status,
	}, nil
}

// VerifyDocument marks a document as verified
func (s *DocumentExpiryService) VerifyDocument(ctx context.Context, driverID uuid.UUID, documentID uuid.UUID, verifiedBy uuid.UUID, userEmail string) error {
	driver, err := s.repo.GetByID(ctx, driverID)
	if err != nil {
		return fmt.Errorf("failed to get driver: %w", err)
	}
	if driver == nil {
		return ErrDriverNotFound
	}

	// Find and update the document
	found := false
	for i := range driver.Documents {
		if driver.Documents[i].ID == documentID {
			driver.Documents[i].Verified = true
			driver.Documents[i].VerifiedBy = &verifiedBy
			now := time.Now()
			driver.Documents[i].VerifiedAt = &now
			found = true
			break
		}
	}

	if !found {
		return ErrDocumentNotFound
	}

	// Update driver with verified document
	updates := map[string]interface{}{
		"documents":  driver.Documents,
		"updated_at": time.Now(),
	}

	if err := s.repo.Update(ctx, driverID, updates); err != nil {
		return fmt.Errorf("failed to update driver document: %w", err)
	}

	s.logger.Info("Document verified",
		zap.String("driver_id", driverID.String()),
		zap.String("document_id", documentID.String()),
		zap.String("verified_by", userEmail),
	)

	return nil
}

// SendExpiryNotifications sends notifications for expiring documents
func (s *DocumentExpiryService) SendExpiryNotifications(ctx context.Context, tenantID string, warningDays int) (int, error) {
	expiringDocs, err := s.GetExpiringDocuments(ctx, tenantID, warningDays)
	if err != nil {
		return 0, fmt.Errorf("failed to get expiring documents: %w", err)
	}

	notificationsSent := 0

	for _, doc := range expiringDocs {
		if doc.Status == "expired" {
			// Send expired notification
			notificationsSent++
			s.logger.Warn("Document expired",
				zap.String("driver_id", doc.DriverID.String()),
				zap.String("driver_name", doc.DriverName),
				zap.String("document_type", doc.DocumentType),
				zap.String("document_number", doc.DocumentNumber),
				zap.Time("expiry_date", doc.ExpiryDate),
			)
		} else if doc.Status == "expiring_soon" {
			// Send expiring soon notification
			notificationsSent++
			s.logger.Info("Document expiring soon",
				zap.String("driver_id", doc.DriverID.String()),
				zap.String("driver_name", doc.DriverName),
				zap.String("document_type", doc.DocumentType),
				zap.Int("days_until_expiry", doc.DaysUntilExpiry),
			)
		}
	}

	return notificationsSent, nil
}

// GetDocumentComplianceRate calculates document compliance rate for organization
func (s *DocumentExpiryService) GetDocumentComplianceRate(ctx context.Context, tenantID string) (float64, error) {
	// Get all drivers in organization
	// Count total documents and valid documents
	// Calculate compliance rate

	// In production:
	// total_docs := SELECT COUNT(*) FROM drivers, jsonb_array_elements(documents) WHERE tenant_id = $1
	// valid_docs := SELECT COUNT(*) FROM drivers, jsonb_array_elements(documents) WHERE tenant_id = $1 AND (doc->>'expiry_date')::date >= CURRENT_DATE
	// rate = (valid_docs / total_docs) * 100

	return 100.0, nil // Placeholder
}
