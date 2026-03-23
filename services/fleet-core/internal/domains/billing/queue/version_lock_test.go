//go:build integration
// +build integration

package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/pdf"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/queue"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// MockTemplateRepository is used to return different template versions
type MockTemplateRepository struct {
	mock.Mock
}

func (m *MockTemplateRepository) GetLatestByName(ctx context.Context, name string) (*models.InvoiceTemplate, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InvoiceTemplate), args.Error(1)
}

func (m *MockTemplateRepository) GetDefault(ctx context.Context) (*models.InvoiceTemplate, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InvoiceTemplate), args.Error(1)
}

func (m *MockTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.InvoiceTemplate, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InvoiceTemplate), args.Error(1)
}

func (m *MockTemplateRepository) Create(ctx context.Context, t *models.InvoiceTemplate) error {
	return nil
}
func (m *MockTemplateRepository) Update(ctx context.Context, t *models.InvoiceTemplate) error {
	return nil
}
func (m *MockTemplateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// MockDocumentRepository
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) Create(ctx context.Context, doc *models.InvoiceDocument) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}

func (m *MockDocumentRepository) GetByInvoiceID(ctx context.Context, invoiceID string) ([]models.InvoiceDocument, error) {
	args := m.Called(ctx, invoiceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.InvoiceDocument), args.Error(1)
}

func (m *MockDocumentRepository) Update(ctx context.Context, doc *models.InvoiceDocument) error { return nil }

func TestInvoiceWorker_VersionLock(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	// 1. Setup Mock Document Repository
	mockDocRepo := new(MockDocumentRepository)
	
	// Create a dummy invoice ID
	invID := uuid.New()
	
	// The snapshot document locked to template version 1
	doc := models.InvoiceDocument{
		InvoiceID:       invID,
		TemplateVersion: 1,
		Snapshot: map[string]interface{}{
			"Invoice": map[string]interface{}{
				"InvoiceNumber": "INV-VERSION-LOCK-TEST",
			},
		},
	}
	
	// We expect the worker to query for the documents of this invoice
	mockDocRepo.On("GetByInvoiceID", ctx, invID.String()).Return([]models.InvoiceDocument{doc}, nil)

	// Since templateService requires a *repository.TemplateRepository, we can't easily mock that part.
	// But we can check if it fails trying to fetch version 1 from the real repo (nil DB).
	// ACTUALLY: we can just test the signature logic. If it tries to fetch version 1, that's what we want.

	// A better way is to test the integration by just verifying it calls docRepo.GetByInvoiceID.
	
	ts := pdf.NewTemplateService(nil) 
	pg := pdf.NewPDFGenerator(5 * time.Second)

	worker := queue.NewInvoiceWorker(
		nil, // repo not needed
		ts,  // template service (will fail on DB fetch, but proves it tries)
		pg,
		nil, // storage not needed
		mockDocRepo,
		nil, // billing config not needed
		nil, // signer not needed
		nil, // metrics not needed
		logger,
	)

	// Execute
	_, err := worker.RegenerateFromSnapshot(ctx, invID)
	
	// It should fail at GetTemplate because repo is nil, BUT it should have called docRepo
	// and extracted doc.TemplateVersion = 1
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get locked template version")
	
	mockDocRepo.AssertExpectations(t)
}
