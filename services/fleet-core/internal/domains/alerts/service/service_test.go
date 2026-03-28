package service

import (
	"context"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/alert_router"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/preference_engine"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/template_engine"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MockRepo for testing
type mockRepo struct {
	repository.Repository
	logs []models.NotificationLog
}

func (m *mockRepo) CreateLog(ctx context.Context, log *models.NotificationLog) error {
	m.logs = append(m.logs, *log)
	return nil
}

func (m *mockRepo) GetTemplate(ctx context.Context, name string, channel models.Channel, language string, orgID uuid.UUID) (*models.Template, error) {
	return &models.Template{
		Name:    name,
		Channel: channel,
		Subject: "Test Subject",
		Body:    "Hello {{.name}}",
	}, nil
}

func (m *mockRepo) GetUserPreferences(ctx context.Context, userID, orgID uuid.UUID) ([]models.UserPreference, error) {
	return nil, nil // All enabled by default
}

func TestTriggerAlert(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := &mockRepo{}
	router := alert_router.NewAlertRouter()
	prefEngine := preference_engine.NewPreferenceEngine()
	tmplEngine := template_engine.NewTemplateEngine()

	svc := NewAlertsService(repo, router, prefEngine, tmplEngine, logger)

	userID := uuid.New()
	req := models.AlertRequest{
		UserID:    &userID,
		AlertType: "OVERSPEED",
		Priority:  models.PriorityHigh,
		Data:      map[string]any{"name": "John Doe"},
	}

	err := svc.TriggerAlert(context.Background(), req)
	if err != nil {
		t.Fatalf("TriggerAlert failed: %v", err)
	}

	// High Priority should route to EMAIL, PUSH, IN_APP (according to our router)
	// EMAIL, PUSH, IN_APP = 3 logs
	if len(repo.logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(repo.logs))
	}

	for _, log := range repo.logs {
		if log.Body != "Hello John Doe" {
			t.Errorf("Body not rendered correctly, got %s", log.Body)
		}
		if log.Status != models.StatusQueued {
			t.Errorf("Expected status QUEUED, got %s", log.Status)
		}
	}
}
