package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/alert-service/internal/models"
	"github.com/flyfleet/alert-service/internal/providers"
	"github.com/flyfleet/alert-service/internal/repository"
	"github.com/flyfleet/alert-service/internal/rule_engine"
)

type AlertService struct {
	repo                *repository.AlertRepository
	ruleEngine          *rule_engine.RuleEngine
	notificationService *providers.NotificationService
	logger              *zap.Logger
}

func NewAlertService(
	repo *repository.AlertRepository,
	ruleEngine *rule_engine.RuleEngine,
	notificationService *providers.NotificationService,
	logger *zap.Logger,
) *AlertService {
	return &AlertService{
		repo:                repo,
		ruleEngine:          ruleEngine,
		notificationService: notificationService,
		logger:              logger,
	}
}

// CreateRule creates a new alert rule
func (s *AlertService) CreateRule(ctx context.Context, orgID uuid.UUID, req *models.CreateRuleRequest, createdBy uuid.UUID) (*models.AlertRule, error) {
	// Validate rule type
	validTypes := map[string]bool{
		models.AlertTypeOverspeed:          true,
		models.AlertTypeGeofenceEntry:      true,
		models.AlertTypeGeofenceExit:       true,
		models.AlertTypeHarshAcceleration:  true,
		models.AlertTypeHarshBraking:       true,
		models.AlertTypeHarshCornering:     true,
		models.AlertTypeIdleTime:           true,
		models.AlertTypeFuelTheft:          true,
		models.AlertTypeLowBattery:         true,
		models.AlertTypeTemperatureExcursion: true,
		models.AlertTypeAccidentDetected:   true,
		models.AlertTypeMaintenanceDue:     true,
		models.AlertTypeDocumentExpiry:     true,
		models.AlertTypeDriverFatigue:      true,
		models.AlertTypeADASEvent:          true,
		models.AlertTypeVehicleOffline:     true,
	}

	if !validTypes[req.Type] {
		return nil, fmt.Errorf("invalid alert type: %s", req.Type)
	}

	// Validate severity
	validSeverities := map[string]bool{
		models.SeverityInfo:     true,
		models.SeverityWarning:  true,
		models.SeverityCritical: true,
	}

	if !validSeverities[req.Severity] {
		return nil, fmt.Errorf("invalid severity: %s", req.Severity)
	}

	// Validate that at least one action is specified
	if len(req.Actions) == 0 {
		return nil, fmt.Errorf("at least one action is required")
	}

	rule := &models.AlertRule{
		ID:          uuid.New().String(),
		OrgID:       orgID,
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Conditions:  req.Conditions,
		Severity:    req.Severity,
		Actions:     req.Actions,
		Vehicles:    req.Vehicles,
		DriverIDs:   req.DriverIDs,
		Enabled:     true,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateRule(ctx, rule); err != nil {
		s.logger.Error("failed to create alert rule", zap.Error(err))
		return nil, fmt.Errorf("failed to create alert rule: %w", err)
	}

	s.logger.Info("alert rule created", zap.String("rule_id", rule.ID))
	return rule, nil
}

// GetRule retrieves an alert rule by ID
func (s *AlertService) GetRule(ctx context.Context, id string) (*models.AlertRule, error) {
	rule, err := s.repo.GetRuleByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get alert rule", zap.String("rule_id", id), zap.Error(err))
		return nil, fmt.Errorf("alert rule not found: %w", err)
	}
	return rule, nil
}

// ListRules retrieves alert rules with filtering and pagination
func (s *AlertService) ListRules(ctx context.Context, orgID uuid.UUID, filter models.AlertListFilter) ([]*models.AlertRule, int, error) {
	rules, totalCount, err := s.repo.ListRules(ctx, orgID, filter)
	if err != nil {
		s.logger.Error("failed to list alert rules", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list alert rules: %w", err)
	}
	return rules, totalCount, nil
}

// UpdateRule modifies an existing alert rule
func (s *AlertService) UpdateRule(ctx context.Context, id string, req *models.UpdateRuleRequest) (*models.AlertRule, error) {
	// Validate that rule exists
	_, err := s.GetRule(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Conditions != nil {
		updates["conditions"] = *req.Conditions
	}
	if req.Severity != nil {
		// Validate severity
		validSeverities := map[string]bool{
			models.SeverityInfo:     true,
			models.SeverityWarning:  true,
			models.SeverityCritical: true,
		}
		if !validSeverities[*req.Severity] {
			return nil, fmt.Errorf("invalid severity: %s", *req.Severity)
		}
		updates["severity"] = *req.Severity
	}
	if req.Actions != nil {
		if len(*req.Actions) == 0 {
			return nil, fmt.Errorf("at least one action is required")
		}
		updates["actions"] = *req.Actions
	}
	if req.Vehicles != nil {
		updates["vehicles"] = *req.Vehicles
	}
	if req.DriverIDs != nil {
		updates["driver_ids"] = *req.DriverIDs
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()

		if err := s.repo.UpdateRule(ctx, id, updates); err != nil {
			s.logger.Error("failed to update alert rule", zap.String("rule_id", id), zap.Error(err))
			return nil, fmt.Errorf("failed to update alert rule: %w", err)
		}
	}

	return s.GetRule(ctx, id)
}

// DeleteRule removes an alert rule
func (s *AlertService) DeleteRule(ctx context.Context, id string) error {
	// Check if rule exists
	_, err := s.GetRule(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteRule(ctx, id); err != nil {
		s.logger.Error("failed to delete alert rule", zap.String("rule_id", id), zap.Error(err))
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	s.logger.Info("alert rule deleted", zap.String("rule_id", id))
	return nil
}

// EnableRule enables an alert rule
func (s *AlertService) EnableRule(ctx context.Context, id string) error {
	updates := map[string]interface{}{
		"enabled":    true,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateRule(ctx, id, updates); err != nil {
		s.logger.Error("failed to enable alert rule", zap.String("rule_id", id), zap.Error(err))
		return fmt.Errorf("failed to enable alert rule: %w", err)
	}

	s.logger.Info("alert rule enabled", zap.String("rule_id", id))
	return nil
}

// DisableRule disables an alert rule
func (s *AlertService) DisableRule(ctx context.Context, id string) error {
	updates := map[string]interface{}{
		"enabled":    false,
		"updated_at": time.Now(),
	}

	if err := s.repo.UpdateRule(ctx, id, updates); err != nil {
		s.logger.Error("failed to disable alert rule", zap.String("rule_id", id), zap.Error(err))
		return fmt.Errorf("failed to disable alert rule: %w", err)
	}

	s.logger.Info("alert rule disabled", zap.String("rule_id", id))
	return nil
}

// CreateAlert creates a new alert instance
func (s *AlertService) CreateAlert(ctx context.Context, alert *models.Alert) error {
	if err := s.repo.CreateAlert(ctx, alert); err != nil {
		s.logger.Error("failed to create alert", zap.Error(err))
		return fmt.Errorf("failed to create alert: %w", err)
	}

	// Send notifications for the alert
	if err := s.sendAlertNotifications(alert); err != nil {
		s.logger.Error("failed to send alert notifications", zap.Error(err))
		// Don't fail the alert creation if notifications fail
	}

	s.logger.Info("alert created", zap.String("alert_id", alert.ID))
	return nil
}

// GetAlert retrieves an alert by ID
func (s *AlertService) GetAlert(ctx context.Context, id string) (*models.Alert, error) {
	alert, err := s.repo.GetAlertByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get alert", zap.String("alert_id", id), zap.Error(err))
		return nil, fmt.Errorf("alert not found: %w", err)
	}
	s.populateAlertMetadata(alert)
	return alert, nil
}

// ListAlerts retrieves alerts with filtering and pagination
func (s *AlertService) ListAlerts(ctx context.Context, filter models.AlertListFilter) ([]*models.Alert, int, error) {
	alerts, totalCount, err := s.repo.ListAlerts(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list alerts", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list alerts: %w", err)
	}
	for _, alert := range alerts {
		s.populateAlertMetadata(alert)
	}
	return alerts, totalCount, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *AlertService) AcknowledgeAlert(ctx context.Context, id string, userID uuid.UUID, notes string) error {
	// Check if alert exists
	alert, err := s.GetAlert(ctx, id)
	if err != nil {
		return err
	}

	if alert.Status != models.AlertStatusNew {
		return fmt.Errorf("alert must be in 'new' status to acknowledge")
	}

	if err := s.repo.UpdateAlertStatus(ctx, id, models.AlertStatusAcknowledged, &userID); err != nil {
		s.logger.Error("failed to acknowledge alert", zap.String("alert_id", id), zap.Error(err))
		return fmt.Errorf("failed to acknowledge alert: %w", err)
	}

	// Update notes if provided
	if notes != "" {
		updates := map[string]interface{}{
			"notes":      notes,
			"updated_at": time.Now(),
		}
		s.repo.UpdateRule(ctx, id, updates) // This will update the alert's notes
	}

	s.logger.Info("alert acknowledged", zap.String("alert_id", id), zap.String("user_id", userID.String()))
	return nil
}

// ResolveAlert marks an alert as resolved
func (s *AlertService) ResolveAlert(ctx context.Context, id string, userID uuid.UUID, notes string) error {
	// Check if alert exists
	alert, err := s.GetAlert(ctx, id)
	if err != nil {
		return err
	}

	if alert.Status == models.AlertStatusResolved {
		return fmt.Errorf("alert is already resolved")
	}

	if err := s.repo.UpdateAlertStatus(ctx, id, models.AlertStatusResolved, &userID); err != nil {
		s.logger.Error("failed to resolve alert", zap.String("alert_id", id), zap.Error(err))
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	// Update notes if provided
	if notes != "" {
		updates := map[string]interface{}{
			"notes":      notes,
			"updated_at": time.Now(),
		}
		s.repo.UpdateRule(ctx, id, updates) // This will update the alert's notes
	}

	s.logger.Info("alert resolved", zap.String("alert_id", id), zap.String("user_id", userID.String()))
	return nil
}

// EscalateAlert escalates an alert to higher authorities
func (s *AlertService) EscalateAlert(ctx context.Context, id string, escalateTo []uuid.UUID, notes string) error {
	// Check if alert exists
	_, err := s.GetAlert(ctx, id)
	if err != nil {
		return err
	}

	// Update escalation information
	updates := map[string]interface{}{
		"status":       models.AlertStatusEscalated,
		"escalated_to": escalateTo,
		"notes":        notes,
		"updated_at":   time.Now(),
	}

	if err := s.repo.UpdateRule(ctx, id, updates); err != nil {
		s.logger.Error("failed to escalate alert", zap.String("alert_id", id), zap.Error(err))
		return fmt.Errorf("failed to escalate alert: %w", err)
	}

	s.logger.Info("alert escalated", zap.String("alert_id", id))
	return nil
}

// GetAlertSummary retrieves alert statistics
func (s *AlertService) GetAlertSummary(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) (*models.AlertSummary, error) {
	summary, err := s.repo.GetAlertSummary(ctx, orgID, fromDate, toDate)
	if err != nil {
		s.logger.Error("failed to get alert summary", zap.Error(err))
		return nil, fmt.Errorf("failed to get alert summary: %w", err)
	}
	return summary, nil
}

// GetAlertTrends retrieves alert trends over time
func (s *AlertService) GetAlertTrends(ctx context.Context, orgID uuid.UUID, fromDate, toDate time.Time) ([]*models.AlertTrend, error) {
	// This would aggregate alerts by date
	// Implementation would depend on specific requirements for trend analysis
	return []*models.AlertTrend{}, nil
}

// ProcessTelemetryEvent processes incoming telemetry data through the rule engine
func (s *AlertService) ProcessTelemetryEvent(ctx context.Context, eventType string, data []byte) error {
	switch eventType {
	case "gps":
		return s.ruleEngine.ProcessGPSEvent(data)
	case "obd":
		return s.ruleEngine.ProcessOBDEvent(data)
	case "fuel":
		return s.ruleEngine.ProcessFuelEvent(data)
	case "driver":
		return s.ruleEngine.ProcessDriverEvent(data)
	default:
		return fmt.Errorf("unsupported telemetry event type: %s", eventType)
	}
}

func (s *AlertService) populateAlertMetadata(alert *models.Alert) {
	if alert == nil {
		return
	}
	alert.Acknowledged = (alert.Status == models.AlertStatusAcknowledged || alert.Status == models.AlertStatusResolved)
	if alert.Message == "" {
		alert.Message = s.formatAlertMessage(alert, "default")
	}
}

// sendAlertNotifications sends notifications for an alert based on its actions
func (s *AlertService) sendAlertNotifications(alert *models.Alert) error {
	// Get the rule to determine actions
	rule, err := s.repo.GetRuleByID(context.Background(), alert.RuleID)
	if err != nil {
		return fmt.Errorf("failed to get rule for alert: %w", err)
	}

	var notifs []*providers.Notification

	// Process each action
	for _, action := range rule.Actions {
		var notification *providers.Notification

		switch action.Type {
		case models.ActionTypeSMS:
			notification = &providers.Notification{
				Type:      "sms",
				Recipient: action.Target,
				Body:      s.formatAlertMessage(alert, action.Template),
			}
		case models.ActionTypeEmail:
			notification = &providers.Notification{
				Type:      "email",
				Recipient: action.Target,
				Subject:   s.formatAlertSubject(alert, action.Template),
				Body:      s.formatAlertMessage(alert, action.Template),
			}
		case models.ActionTypePush:
			notification = &providers.Notification{
				Type:      "push",
				Recipient: action.Target,
				Subject:   s.formatAlertSubject(alert, action.Template),
				Body:      s.formatAlertMessage(alert, action.Template),
				Data:      map[string]interface{}{"alert_id": alert.ID, "type": alert.Type},
			}
		case models.ActionTypeNotification:
			notification = &providers.Notification{
				Type:      "notification",
				Recipient: action.Target,
				Body:      s.formatAlertMessage(alert, action.Template),
			}
		default:
			s.logger.Warn("unsupported action type", zap.String("type", action.Type))
			continue
		}

		if notification != nil {
			notifs = append(notifs, notification)
		}
	}

	// Send all notifications
	if len(notifs) > 0 {
		errors := s.notificationService.SendBulkNotifications(notifs)
		if len(errors) > 0 {
			return fmt.Errorf("some notifications failed: %v", errors)
		}
	}

	return nil
}

// formatAlertMessage formats an alert message using template and alert data
func (s *AlertService) formatAlertMessage(alert *models.Alert, templateName string) string {
	// This would retrieve the template from database and format it
	// For now, return a basic formatted message
	return fmt.Sprintf("Alert: %s - Severity: %s - Time: %s", 
		alert.Type, alert.Severity, alert.Timestamp.Format("2006-01-02 15:04:05"))
}

// formatAlertSubject formats an alert subject for email notifications
func (s *AlertService) formatAlertSubject(alert *models.Alert, templateName string) string {
	return fmt.Sprintf("[%s] %s Alert", alert.Severity, alert.Type)
}
func (s *AlertService) GetStats(ctx context.Context) (*models.AlertStats, error) {
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		s.logger.Error("failed to get alert stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	return stats, nil
}
