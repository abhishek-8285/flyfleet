package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/settings/middleware"
	"github.com/flyfleet/fleet-core-service/internal/domains/settings/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/settings/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/settings/service"
	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// testAdminRoleMiddleware injects an admin role into the request context for tests.
func testAdminRoleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxkeys.RoleKey, "admin")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Mock Repo duplication for handlers test
type MockRepo struct {
	LastAuditLog     *models.AuditLog
	AuditLogs        []models.AuditLog
	PlatformSettings *models.PlatformSettings
}

func (m *MockRepo) GetPlatformSettings(ctx context.Context) (*models.PlatformSettings, error) {
	return &models.PlatformSettings{
		ID:              models.PlatformSettingsSingletonID,
		SystemName:      "Test",
		SupportEmail:    "test@fleet-system.com",
		DefaultTimezone: "UTC",
		DefaultCurrency: "USD",
		DefaultLanguage: "en-US",
	}, nil
}
func (m *MockRepo) UpdatePlatformSettings(ctx context.Context, s *models.PlatformSettings) error {
	m.PlatformSettings = s
	return nil
}
func (m *MockRepo) ListFeatureFlags(ctx context.Context) ([]models.FeatureFlag, error) {
	return nil, nil
}
func (m *MockRepo) GetFeatureFlag(ctx context.Context, id uuid.UUID) (*models.FeatureFlag, error) {
	return nil, nil
}
func (m *MockRepo) CreateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error { return nil }
func (m *MockRepo) UpdateFeatureFlag(ctx context.Context, f *models.FeatureFlag) error { return nil }
func (m *MockRepo) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error          { return nil }
func (m *MockRepo) GetLocalizationSettings(ctx context.Context) ([]models.LocalizationSettings, error) {
	return nil, nil
}
func (m *MockRepo) UpdateLocalizationSettings(ctx context.Context, l *models.LocalizationSettings) error {
	return nil
}
func (m *MockRepo) ListTemplates(ctx context.Context) ([]models.Template, error) { return nil, nil }
func (m *MockRepo) GetTemplate(ctx context.Context, id uuid.UUID) (*models.Template, error) {
	return nil, nil
}
func (m *MockRepo) CreateTemplate(ctx context.Context, t *models.Template) error { return nil }
func (m *MockRepo) UpdateTemplate(ctx context.Context, t *models.Template) error { return nil }
func (m *MockRepo) ListNotificationChannels(ctx context.Context) ([]models.NotificationChannel, error) {
	return nil, nil
}
func (m *MockRepo) UpdateNotificationChannel(ctx context.Context, c *models.NotificationChannel) error {
	return nil
}
func (m *MockRepo) GetSystemLimits(ctx context.Context) (*models.SystemLimit, error) {
	return &models.SystemLimit{
		ID:             uuid.New(),
		APIRateLimit:   10000,
		BurstLimit:     500,
		AuditRetention: "1y",
	}, nil
}
func (m *MockRepo) UpdateSystemLimits(ctx context.Context, l *models.SystemLimit) error {
	return nil
}
func (m *MockRepo) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	m.LastAuditLog = log
	m.AuditLogs = append(m.AuditLogs, *log)
	return nil
}

func (m *MockRepo) ListAuditLogs(ctx context.Context, query models.AuditLogQuery) ([]models.AuditLog, error) {
	logs := m.AuditLogs
	if query.EntityType != "" {
		filtered := make([]models.AuditLog, 0, len(logs))
		for _, log := range logs {
			if log.EntityType == query.EntityType {
				filtered = append(filtered, log)
			}
		}
		logs = filtered
	}
	if query.Action != "" {
		filtered := make([]models.AuditLog, 0, len(logs))
		for _, log := range logs {
			if log.Action == query.Action {
				filtered = append(filtered, log)
			}
		}
		logs = filtered
	}
	start := query.Offset
	if start > len(logs) {
		return []models.AuditLog{}, nil
	}
	end := start + query.Limit
	if end > len(logs) {
		end = len(logs)
	}
	return append([]models.AuditLog{}, logs[start:end]...), nil
}

func (m *MockRepo) WithTx(ctx context.Context, fn func(repository.Repository) error) error {
	return fn(m)
}

func (m *MockRepo) CleanupOldAuditLogs(ctx context.Context, retentionPeriod string) (int64, error) {
	return 0, nil
}

func TestUpdatePlatformSettings_Integration(t *testing.T) {
	// Setup
	repo := &MockRepo{}
	logger := zap.NewNop()
	svc := service.NewSettingsService(repo, logger)
	h := NewSettingsHandlers(svc, logger)

	// Router with Middleware
	router := http.NewServeMux()
	router.HandleFunc("/settings/platform", h.UpdatePlatformSettings)
	handler := testAdminRoleMiddleware(middleware.UserContextMiddleware(router))

	// User ID
	userID := uuid.New()
	userIP := "192.168.1.100"

	// Request - use proper JSON field names matching the struct
	payload := map[string]interface{}{
		"system_name":      "New System Name",
		"support_email":    "support@test.com",
		"default_timezone": "UTC",
		"default_currency": "USD",
		"default_language": "en-US",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("PUT", "/settings/platform", bytes.NewBuffer(body))
	req.Header.Set("X-User-ID", userID.String())
	req.Header.Set("X-Forwarded-For", userIP)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(rr, req)

	// Debug: print response body if error
	if rr.Code != http.StatusOK {
		t.Logf("Response body: %s", rr.Body.String())
	}

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

func TestGetSettingsExport_ReturnsSnapshot(t *testing.T) {
	repo := &MockRepo{}
	logger := zap.NewNop()
	svc := service.NewSettingsService(repo, logger)
	h := NewSettingsHandlers(svc, logger)

	router := http.NewServeMux()
	router.HandleFunc("/settings/export", h.GetSettingsExport)
	handler := testAdminRoleMiddleware(middleware.UserContextMiddleware(router))

	req := httptest.NewRequest(http.MethodGet, "/settings/export", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode export response: %v", err)
	}
	if _, ok := payload["platform_settings"]; !ok {
		t.Fatal("missing platform_settings in export payload")
	}
	if _, ok := payload["exported_at"]; !ok {
		t.Fatal("missing exported_at in export payload")
	}
}

func TestImportSettings_ReturnsResult(t *testing.T) {
	repo := &MockRepo{}
	logger := zap.NewNop()
	svc := service.NewSettingsService(repo, logger)
	h := NewSettingsHandlers(svc, logger)

	router := http.NewServeMux()
	router.HandleFunc("/settings/import", h.ImportSettings)
	handler := testAdminRoleMiddleware(middleware.UserContextMiddleware(router))

	payload := map[string]any{
		"platform_settings": map[string]any{
			"system_name":      "Import System",
			"support_email":    "ops@example.com",
			"default_timezone": "UTC",
		},
		"feature_flags": []map[string]any{
			{"key": "new-flow", "scope": "GLOBAL", "enabled": true},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/settings/import", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := out["platform_updated"]; !ok {
		t.Fatalf("expected import result payload, got %v", out)
	}
}

func TestGetAuditLogs_ReturnsFilteredList(t *testing.T) {
	repo := &MockRepo{
		AuditLogs: []models.AuditLog{
			{EntityType: "FEATURE_FLAG", Action: "CREATE"},
			{EntityType: "SETTINGS_IMPORT", Action: "IMPORT"},
		},
	}
	logger := zap.NewNop()
	svc := service.NewSettingsService(repo, logger)
	h := NewSettingsHandlers(svc, logger)

	router := http.NewServeMux()
	router.HandleFunc("/settings/audit", h.GetAuditLogs)
	handler := testAdminRoleMiddleware(middleware.UserContextMiddleware(router))

	req := httptest.NewRequest(http.MethodGet, "/settings/audit?entity_type=settings_import&limit=20&offset=0", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode audit list response: %v", err)
	}
	logs, ok := payload["logs"].([]any)
	if !ok {
		t.Fatalf("expected logs array in response, got %v", payload)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 filtered audit log, got %d", len(logs))
	}
}

func TestUpdatePlatformSettings_InvalidPayloadReturnsBadRequest(t *testing.T) {
	repo := &MockRepo{}
	logger := zap.NewNop()
	svc := service.NewSettingsService(repo, logger)
	h := NewSettingsHandlers(svc, logger)

	router := http.NewServeMux()
	router.HandleFunc("/settings/platform", h.UpdatePlatformSettings)
	handler := testAdminRoleMiddleware(middleware.UserContextMiddleware(router))

	payload := map[string]string{
		"system_name":      "Invalid Email System",
		"support_email":    "not-an-email",
		"default_timezone": "UTC",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/settings/platform", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for validation error, got %d", rr.Code)
	}
}

func TestUpdateSystemLimits_NegativeRateLimitReturnsBadRequest(t *testing.T) {
	repo := &MockRepo{}
	logger := zap.NewNop()
	svc := service.NewSettingsService(repo, logger)
	h := NewSettingsHandlers(svc, logger)

	router := http.NewServeMux()
	router.HandleFunc("/settings/limits", h.UpdateSystemLimits)
	handler := testAdminRoleMiddleware(middleware.UserContextMiddleware(router))

	payload := map[string]any{
		"api_rate_limit": -1,
		"burst_limit":    10,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/settings/limits", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for validation error, got %d", rr.Code)
	}
}

func TestImportSettings_InvalidPayloadReturnsBadRequest(t *testing.T) {
	repo := &MockRepo{}
	logger := zap.NewNop()
	svc := service.NewSettingsService(repo, logger)
	h := NewSettingsHandlers(svc, logger)

	router := http.NewServeMux()
	router.HandleFunc("/settings/import", h.ImportSettings)
	handler := testAdminRoleMiddleware(middleware.UserContextMiddleware(router))

	req := httptest.NewRequest(http.MethodPost, "/settings/import", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid payload, got %d", rr.Code)
	}
}

func TestImportSettings_DuplicateFeatureFlagKeyReturnsBadRequest(t *testing.T) {
	repo := &MockRepo{}
	logger := zap.NewNop()
	svc := service.NewSettingsService(repo, logger)
	h := NewSettingsHandlers(svc, logger)

	router := http.NewServeMux()
	router.HandleFunc("/settings/import", h.ImportSettings)
	handler := testAdminRoleMiddleware(middleware.UserContextMiddleware(router))

	payload := map[string]any{
		"feature_flags": []map[string]any{
			{"key": "dispatch-v2", "scope": "GLOBAL", "enabled": true},
			{"key": " DISPATCH-v2 ", "scope": "GLOBAL", "enabled": false},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/settings/import", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for duplicate feature flag key, got %d", rr.Code)
	}
}

func TestGetAuditLogs_InvalidLimitReturnsBadRequest(t *testing.T) {
	repo := &MockRepo{}
	logger := zap.NewNop()
	svc := service.NewSettingsService(repo, logger)
	h := NewSettingsHandlers(svc, logger)

	router := http.NewServeMux()
	router.HandleFunc("/settings/audit", h.GetAuditLogs)
	handler := testAdminRoleMiddleware(middleware.UserContextMiddleware(router))

	req := httptest.NewRequest(http.MethodGet, "/settings/audit?limit=abc", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid limit, got %d", rr.Code)
	}
}
