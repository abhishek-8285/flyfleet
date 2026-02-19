package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/admin-support-service/internal/domains/settings/models"
	"github.com/flyfleet/admin-support-service/internal/domains/settings/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Mock Repo duplication for handlers test
type MockRepo struct {
	LastAuditLog *models.AuditLog
}

func (m *MockRepo) GetPlatformSettings(ctx context.Context) (*models.PlatformSettings, error) {
	return &models.PlatformSettings{SystemName: "Test"}, nil
}
func (m *MockRepo) UpdatePlatformSettings(ctx context.Context, s *models.PlatformSettings) error {
	return nil
}
func (m *MockRepo) ListFeatureFlags(ctx context.Context) ([]models.FeatureFlag, error) { return nil, nil }
func (m *MockRepo) GetFeatureFlag(ctx context.Context, id uuid.UUID) (*models.FeatureFlag, error) {
	return nil, nil
}
func (m *MockRepo) CreateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error { return nil }
func (m *MockRepo) UpdateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error { return nil }
func (m *MockRepo) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error        { return nil }
func (m *MockRepo) GetLocalizationSettings(ctx context.Context) ([]models.LocalizationSettings, error) {
	return nil, nil
}
func (m *MockRepo) UpdateLocalizationSettings(ctx context.Context, l *models.LocalizationSettings) error {
	return nil
}
func (m *MockRepo) ListTemplates(ctx context.Context) ([]models.Template, error)                   { return nil, nil }
func (m *MockRepo) GetTemplate(ctx context.Context, id uuid.UUID) (*models.Template, error)        { return nil, nil }
func (m *MockRepo) CreateTemplate(ctx context.Context, t *models.Template) error                   { return nil }
func (m *MockRepo) UpdateTemplate(ctx context.Context, t *models.Template) error                   { return nil }
func (m *MockRepo) ListNotificationChannels(ctx context.Context) ([]models.NotificationChannel, error) {
	return nil, nil
}
func (m *MockRepo) UpdateNotificationChannel(ctx context.Context, c *models.NotificationChannel) error {
	return nil
}
func (m *MockRepo) GetSystemLimits(ctx context.Context) (*models.SystemLimit, error) { return nil, nil }
func (m *MockRepo) UpdateSystemLimits(ctx context.Context, l *models.SystemLimit) error {
	return nil
}
func (m *MockRepo) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	m.LastAuditLog = log
	return nil
}

func TestUpdatePlatformSettings_Integration(t *testing.T) {
	// Setup
	repo := &MockRepo{}
	logger := zap.NewNop()
	svc := service.NewSettingsService(repo, logger)
	h := NewSettingsHandlers(svc, logger)

	// Router with Chi
	r := chi.NewRouter()
	r.Route("/settings", func(r chi.Router) {
		h.RegisterRoutes(r)
	})

	// User ID
	userID := uuid.New()
	userIP := "192.168.1.100"

	// Request
	payload := map[string]string{
		"system_name":   "New System Name",
		"support_email": "support@test.com",
		"timezone": 		 "UTC",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("PUT", "/settings/platform", bytes.NewBuffer(body))
	req.Header.Set("X-User-ID", userID.String())
	req.Header.Set("X-Forwarded-For", userIP)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Execute
	r.ServeHTTP(rr, req)

	// Assert Response
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	// Assert Audit Log capture
	if repo.LastAuditLog == nil {
		t.Fatal("Audit log was not created")
	}
	if repo.LastAuditLog.ActorID != userID {
		t.Errorf("Audit Log ActorID = %v, want %v", repo.LastAuditLog.ActorID, userID)
	}
	if repo.LastAuditLog.ActorIP != userIP {
		t.Errorf("Audit Log ActorIP = %v, want %v", repo.LastAuditLog.ActorIP, userIP)
	}
	if repo.LastAuditLog.Action != "UPDATE" {
		t.Errorf("Audit Log Action = %v, want UPDATE", repo.LastAuditLog.Action)
	}
}
