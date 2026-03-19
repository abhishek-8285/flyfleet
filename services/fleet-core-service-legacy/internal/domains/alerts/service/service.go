package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/alert_router"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/preference_engine"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/alerts/template_engine"
	"go.uber.org/zap"
)

type AlertsService struct {
	repo             repository.Repository
	router           alert_router.AlertRouter
	preferenceEngine preference_engine.PreferenceEngine
	templateEngine   template_engine.TemplateEngine
	logger           *zap.Logger
}

func NewAlertsService(
	repo repository.Repository,
	router alert_router.AlertRouter,
	prefEngine preference_engine.PreferenceEngine,
	tmplEngine template_engine.TemplateEngine,
	logger *zap.Logger,
) *AlertsService {
	return &AlertsService{
		repo:             repo,
		router:           router,
		preferenceEngine: prefEngine,
		templateEngine:   tmplEngine,
		logger:           logger,
	}
}

func (s *AlertsService) TriggerAlert(ctx context.Context, req models.AlertRequest) error {
	// 1. Route to initial channels
	channels := s.router.Route(req.Priority)
	if len(channels) == 0 {
		return nil
	}

	// 2. Filter by user preferences
	if req.UserID != nil {
		prefs, err := s.repo.GetUserPreferences(ctx, *req.UserID)
		if err != nil {
			s.logger.Error("Failed to fetch user preferences", zap.Error(err), zap.String("user_id", req.UserID.String()))
		} else {
			channels = s.preferenceEngine.FilterChannels(req.UserID, channels, prefs)
		}
	}

	if len(channels) == 0 {
		s.logger.Info("All channels filtered out by user preferences", zap.String("user_id", fmt.Sprintf("%v", req.UserID)))
		return nil
	}

	// 3. For each channel, render and queue
	for _, channel := range channels {
		// a. Get Template
		tmpl, err := s.repo.GetTemplate(ctx, req.AlertType, channel, "en") // TODO: Support user language
		if err != nil {
			s.logger.Error("Failed to fetch template", zap.Error(err), 
				zap.String("type", req.AlertType), zap.String("channel", string(channel)))
			continue
		}
		if tmpl == nil {
			s.logger.Warn("Template not found", zap.String("type", req.AlertType), zap.String("channel", string(channel)))
			continue
		}

		// b. Render
		subject, body, err := s.templateEngine.Render(tmpl, req.Data)
		if err != nil {
			s.logger.Error("Failed to render template", zap.Error(err), 
				zap.String("type", req.AlertType), zap.String("channel", string(channel)))
			continue
		}

		// c. Determine Recipient (if user_id provided, we might need to fetch their email/phone from another domain in a real system)
		recipient := req.Recipient
		if recipient == "" && req.UserID != nil {
			recipient = fmt.Sprintf("user-%s@example.com", req.UserID.String())
		}

		// d. Create Log entry (QUEUED)
		log := &models.NotificationLog{
			UserID:      req.UserID,
			Recipient:   recipient,
			Channel:     channel,
			AlertType:   req.AlertType,
			Subject:     subject,
			Body:        body,
			Status:      models.StatusQueued,
			MaxRetries:  3,
			Metadata:    req.Metadata,
			ScheduledAt: time.Now(),
		}

		if err := s.repo.CreateLog(ctx, log); err != nil {
			s.logger.Error("Failed to create notification log", zap.Error(err), 
				zap.String("channel", string(channel)))
		} else {
			s.logger.Info("Notification queued", zap.String("id", log.ID.String()), 
				zap.String("channel", string(channel)))
		}
	}

	return nil
}
