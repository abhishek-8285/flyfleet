//go:build integration
// +build integration

package worker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/alerts/models"
	"github.com/flyfleet/fleet-core/internal/domains/alerts/provider_clients"
	"github.com/flyfleet/fleet-core/internal/domains/alerts/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MockRepo for worker testing
type mockWorkerRepo struct {
	repository.Repository
	logs    []models.NotificationLog
	dlq     []models.DLQEntry
	updates int
}

func (m *mockWorkerRepo) ListPendingLogs(ctx context.Context, limit int) ([]models.NotificationLog, error) {
	var pending []models.NotificationLog
	for _, l := range m.logs {
		if l.Status == models.StatusQueued && l.ScheduledAt.Before(time.Now().Add(time.Second)) {
			pending = append(pending, l)
		}
	}
	if len(pending) > limit {
		return pending[:limit], nil
	}
	return pending, nil
}

func (m *mockWorkerRepo) UpdateLog(ctx context.Context, log *models.NotificationLog) error {
	m.updates++
	for i, l := range m.logs {
		if l.ID == log.ID {
			m.logs[i] = *log
			return nil
		}
	}
	return nil
}

func (m *mockWorkerRepo) AddToDLQ(ctx context.Context, entry *models.DLQEntry) error {
	m.dlq = append(m.dlq, *entry)
	return nil
}

func TestWorker_ProcessNotification_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := &mockWorkerRepo{
		logs: []models.NotificationLog{
			{
				ID:          uuid.New(),
				Channel:     models.ChannelEmail,
				Status:      models.StatusQueued,
				ScheduledAt: time.Now(),
				MaxRetries:  3,
			},
		},
	}

	mockEmail := providers.NewMockProvider("mock-email")
	factory := providers.NewProviderFactory(nil, mockEmail, nil, nil)

	w := NewNotificationWorker(repo, factory, logger)

	ctx := context.Background()
	w.processNotification(ctx, &repo.logs[0])

	if repo.logs[0].Status != models.StatusSent {
		t.Errorf("Expected status SENT, got %s", repo.logs[0].Status)
	}
}

func TestWorker_ProcessNotification_Retry(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := &mockWorkerRepo{
		logs: []models.NotificationLog{
			{
				ID:          uuid.New(),
				Channel:     models.ChannelSMS,
				Status:      models.StatusQueued,
				ScheduledAt: time.Now(),
				MaxRetries:  3,
				RetryCount:  0,
			},
		},
	}

	// Provider that always fails
	failProvider := &failingProvider{name: "fail-sms"}
	factory := providers.NewProviderFactory(failProvider, nil, nil, nil)

	w := NewNotificationWorker(repo, factory, logger)

	ctx := context.Background()
	w.processNotification(ctx, &repo.logs[0])

	if repo.logs[0].Status != models.StatusQueued {
		t.Errorf("Expected status QUEUED for retry, got %s", repo.logs[0].Status)
	}
	if repo.logs[0].RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", repo.logs[0].RetryCount)
	}
}

type failingProvider struct {
	name string
}

func (p *failingProvider) Name() string { return p.name }
func (p *failingProvider) Send(ctx context.Context, log *models.NotificationLog) (map[string]any, error) {
	return nil, fmt.Errorf("hard failure")
}
