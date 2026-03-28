package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/dto"
	"github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NotificationService struct {
	repo   repository.Repository
	logger *zap.Logger
}

func NewNotificationService(repo repository.Repository, logger *zap.Logger) *NotificationService {
	return &NotificationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *NotificationService) ListLogs(ctx context.Context, page, limit int, status, channel, search string) ([]dto.NotificationLogDTO, int, error) {
	logs, total, err := s.repo.ListLogs(ctx, page, limit, status, channel, search)
	if err != nil {
		return nil, 0, err
	}
	// Convert models to DTOs
	dtoLogs := make([]dto.NotificationLogDTO, len(logs))
	for i := range logs {
		dtoLogs[i] = *dto.ToNotificationLogDTO(&logs[i])
	}
	return dtoLogs, total, nil
}

func (s *NotificationService) GetLogDetail(ctx context.Context, id uuid.UUID) (*dto.NotificationLogDTO, error) {
	log, err := s.repo.GetLog(ctx, id)
	if err != nil || log == nil {
		return nil, err
	}
	return dto.ToNotificationLogDTO(log), nil
}

func (s *NotificationService) SendNotification(ctx context.Context, req dto.SendNotificationRequest) (uuid.UUID, error) {
	userID := s.getUserID(ctx)
	newLog := models.NotificationLog{
		ID:         uuid.Nil, // Set by repo
		UserEmail:  req.Recipient,
		Channel:    models.NotificationChannel(req.Channel),
		Type:       "TRANSACTIONAL", // Default
		Subject:    req.Subject,
		Status:     models.StatusQueued,
		RetryCount: 0,
		Metadata:   req.Metadata,
		CreatedAt:  time.Now(),
		CreatedBy:  userID,
		UpdatedAt:  time.Now(),
		UpdatedBy:  userID,
	}

	if req.Priority == "HIGH" {
		if newLog.Metadata == nil {
			newLog.Metadata = make(map[string]any)
		}
		newLog.Metadata["priority"] = "HIGH"
	}

	// Save log
	err := s.repo.CreateLog(ctx, &newLog)
	if err != nil {
		return uuid.Nil, err
	}

	// Add event
	err = s.repo.AddEvent(ctx, &models.NotificationEvent{
		NotificationID: newLog.ID,
		EventType:      "QUEUED",
		Message:        fmt.Sprintf("Notification queued via API. Subject: %s", req.Subject),
		CreatedBy:      userID,
	})
	if err != nil {
		s.logger.Error("Failed to add queue event", zap.Error(err))
	}

	// Mock Delivery Trigger
	s.logger.Info("Notification queued", zap.String("id", newLog.ID.String()), zap.String("recipient", req.Recipient))

	return newLog.ID, nil
}

func (s *NotificationService) getUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value("user_id")
	if val == nil {
		return uuid.Nil
	}
	uidStr, ok := val.(string)
	if !ok {
		return uuid.Nil
	}
	uid, _ := uuid.Parse(uidStr)
	return uid
}

func (s *NotificationService) ResendNotification(ctx context.Context, id uuid.UUID, reason string, highPriority bool) (uuid.UUID, error) {
	// 1. Fetch original log
	original, err := s.repo.GetLog(ctx, id)
	if err != nil {
		return uuid.Nil, err
	}
	if original == nil {
		return uuid.Nil, fmt.Errorf("notification not found")
	}

	// 2. Prepare new log entry based on original
	newLog := *original
	newLog.ID = uuid.Nil // Will be set by repository
	newLog.Status = models.StatusQueued
	newLog.RetryCount = 0
	corrID := fmt.Sprintf("resend-%s", original.ID.String())
	newLog.CorrelationID = &corrID
	if newLog.Metadata == nil {
		newLog.Metadata = make(map[string]any)
	}
	newLog.Metadata["resend_reason"] = reason
	newLog.Metadata["original_id"] = original.ID.String()
	if highPriority {
		newLog.Metadata["priority"] = "HIGH"
	}

	// 3. Save new log
	err = s.repo.CreateLog(ctx, &newLog)
	if err != nil {
		return uuid.Nil, err
	}

	// 4. Add "Resend Initiated" event
	err = s.repo.AddEvent(ctx, &models.NotificationEvent{
		NotificationID: newLog.ID,
		EventType:      "QUEUED",
		Message:        fmt.Sprintf("Resend initiated by admin. Reason: %s", reason),
	})
	if err != nil {
		s.logger.Error("Failed to add resend event", zap.Error(err), zap.String("new_id", newLog.ID.String()))
	}

	// 5. In a real system, we would trigger the actual delivery (Email/SMS/Push) here.
	// For this prototype, we'll just log the initiation.
	s.logger.Info("Notification resend initiated", zap.String("original_id", id.String()), zap.String("new_id", newLog.ID.String()))

	return newLog.ID, nil
}

func (s *NotificationService) GetSettings(ctx context.Context) ([]dto.SettingsResponse, error) {
	settings, err := s.repo.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	return dto.ToSettingsResponse(settings), nil
}

func (s *NotificationService) UpdateSettings(ctx context.Context, updates []dto.SettingUpdate) error {
	for i := range updates {
		// Convert DTO to model
		model := updates[i].ToDomain()
		if err := s.repo.UpdateSettings(ctx, model); err != nil {
			return err
		}
	}
	return nil
}

func (s *NotificationService) ListPreferences(ctx context.Context, page, limit int, search string) ([]dto.UserPreferenceResponse, int, error) {
	prefs, total, err := s.repo.ListPreferences(ctx, page, limit, search)
	if err != nil {
		return nil, 0, err
	}

	res := make([]dto.UserPreferenceResponse, len(prefs))
	for i, p := range prefs {
		res[i] = dto.UserPreferenceResponse{
			UserID:         p.UserID,
			UserName:       p.UserName,
			UserEmail:      p.UserEmail,
			EmailEnabled:   p.EmailEnabled,
			SmsEnabled:     p.SmsEnabled,
			PushEnabled:    p.PushEnabled,
			InAppEnabled:   p.InAppEnabled,
			LastModifiedAt: p.LastModifiedAt.Format(time.RFC3339),
		}
	}
	return res, total, nil
}

func (s *NotificationService) GetUserPreference(ctx context.Context, userId uuid.UUID) (*dto.UserPreferenceResponse, error) {
	p, err := s.repo.GetUserPreference(ctx, userId)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}

	return &dto.UserPreferenceResponse{
		UserID:         p.UserID,
		UserName:       p.UserName,
		UserEmail:      p.UserEmail,
		EmailEnabled:   p.EmailEnabled,
		SmsEnabled:     p.SmsEnabled,
		PushEnabled:    p.PushEnabled,
		InAppEnabled:   p.InAppEnabled,
		LastModifiedAt: p.LastModifiedAt.Format(time.RFC3339),
	}, nil
}

func (s *NotificationService) UpdateUserPreference(ctx context.Context, userId uuid.UUID, req dto.PreferenceUpdateRequest, changedBy uuid.UUID) error {
	old, err := s.repo.GetUserPreference(ctx, userId)
	if err != nil {
		return err
	}

	newPref := &models.UserNotificationPreference{
		UserID:         userId,
		EmailEnabled:   req.EmailEnabled,
		SmsEnabled:     req.SmsEnabled,
		PushEnabled:    req.PushEnabled,
		InAppEnabled:   req.InAppEnabled,
		LastModifiedBy: changedBy,
	}

	err = s.repo.UpdateUserPreference(ctx, newPref)
	if err != nil {
		return err
	}

	// Audit
	audit := &models.PreferenceAudit{
		UserID:       userId,
		ChangedBy:    changedBy,
		OldValues:    old,
		NewValues:    newPref,
		ChangeReason: req.Reason,
	}
	return s.repo.AddPreferenceAudit(ctx, audit)
}

func (s *NotificationService) GetGlobalDefaults(ctx context.Context) (*dto.GlobalDefaultResponse, error) {
	defaults, err := s.repo.GetGlobalDefaults(ctx)
	if err != nil {
		return nil, err
	}
	return dto.ToGlobalDefaultResponse(defaults), nil
}

func (s *NotificationService) UpdateGlobalDefaults(ctx context.Context, req dto.GlobalDefaultUpdateRequest, updatedBy uuid.UUID) error {
	d := &models.GlobalNotificationDefault{
		EmailEnabled: req.EmailEnabled,
		SmsEnabled:   req.SmsEnabled,
		PushEnabled:  req.PushEnabled,
		InAppEnabled: req.InAppEnabled,
		UpdatedBy:    updatedBy,
	}
	return s.repo.UpdateGlobalDefaults(ctx, d)
}

func (s *NotificationService) GetPreferenceAudit(ctx context.Context, userId uuid.UUID) ([]models.PreferenceAudit, error) {
	return s.repo.GetPreferenceAudit(ctx, userId)
}
