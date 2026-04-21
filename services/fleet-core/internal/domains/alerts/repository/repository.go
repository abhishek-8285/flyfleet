package repository

import (
	"context"

	"github.com/flyfleet/fleet-core/internal/domains/alerts/models"
	"github.com/google/uuid"
)

type Repository interface {
	// Logs
	CreateLog(ctx context.Context, log *models.NotificationLog) error
	UpdateLog(ctx context.Context, log *models.NotificationLog) error
	GetLog(ctx context.Context, id uuid.UUID) (*models.NotificationLog, error)
	ListPendingLogs(ctx context.Context, limit int) ([]models.NotificationLog, error)

	// Templates
	GetTemplate(ctx context.Context, name string, channel models.Channel, language string, tenantID uuid.UUID) (*models.Template, error)
	CreateTemplate(ctx context.Context, template *models.Template) error

	// Preferences
	GetUserPreferences(ctx context.Context, userID, tenantID uuid.UUID) ([]models.UserPreference, error)
	UpsertUserPreference(ctx context.Context, pref *models.UserPreference) error

	// DLQ
	AddToDLQ(ctx context.Context, entry *models.DLQEntry) error
}
