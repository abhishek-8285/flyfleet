package service

import (
	"context"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core/internal/domains/emergency/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EscalationConfig holds escalation configuration
type EscalationConfig struct {
	AcknowledgementTimeoutMin int // Time before escalating unacknowledged incidents
	ResolutionTimeoutMin      int // Time before escalating unresolved incidents
	MaxEscalationLevel        int // Maximum escalation level
}

// DefaultEscalationConfig returns default escalation configuration
func DefaultEscalationConfig() *EscalationConfig {
	return &EscalationConfig{
		AcknowledgementTimeoutMin: 15, // Escalate if not acknowledged in 15 minutes
		ResolutionTimeoutMin:      60, // Escalate if not resolved in 60 minutes
		MaxEscalationLevel:        3,  // Maximum 3 escalation levels
	}
}

// EscalationService handles incident escalation workflows
type EscalationService struct {
	repo   repository.EmergencyRepository
	config *EscalationConfig
	logger *zap.Logger
}

// NewEscalationService creates a new escalation service
func NewEscalationService(repo repository.EmergencyRepository, config *EscalationConfig, logger *zap.Logger) *EscalationService {
	if config == nil {
		config = DefaultEscalationConfig()
	}
	return &EscalationService{
		repo:   repo,
		config: config,
		logger: logger,
	}
}

// EscalationLevel represents escalation severity
type EscalationLevel int

const (
	Level1 EscalationLevel = iota + 1 // Initial contact
	Level2                            // Supervisor
	Level3                            // Management
)

// CheckAndEscalateIncidents checks for incidents requiring escalation
func (s *EscalationService) CheckAndEscalateIncidents(ctx context.Context) (int, error) {
	incidents, err := s.repo.ListActiveIncidents(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list active incidents: %w", err)
	}

	escalatedCount := 0
	now := time.Now()

	for _, incident := range incidents {
		needsEscalation, reason := s.checkIncidentEscalation(incident, now)
		if needsEscalation {
			if err := s.escalateIncident(ctx, incident.ID, reason); err != nil {
				s.logger.Error("Failed to escalate incident",
					zap.String("incident_id", incident.ID.String()),
					zap.Error(err),
				)
				continue
			}
			escalatedCount++
		}
	}

	return escalatedCount, nil
}

// checkIncidentEscalation checks if an incident needs escalation
func (s *EscalationService) checkIncidentEscalation(incident models.SOSIncident, now time.Time) (bool, string) {
	timeSinceTriggered := now.Sub(incident.Timestamp)

	// Check acknowledgement timeout
	if incident.Status == models.StatusTriggered {
		if int(timeSinceTriggered.Minutes()) > s.config.AcknowledgementTimeoutMin {
			return true, fmt.Sprintf("Not acknowledged within %d minutes", s.config.AcknowledgementTimeoutMin)
		}
	}

	// Check resolution timeout
	if incident.Status == models.StatusAcknowledged {
		if int(timeSinceTriggered.Minutes()) > s.config.ResolutionTimeoutMin {
			return true, fmt.Sprintf("Not resolved within %d minutes", s.config.ResolutionTimeoutMin)
		}
	}

	return false, ""
}

// escalateIncident escalates an incident to the next level
func (s *EscalationService) escalateIncident(ctx context.Context, incidentID uuid.UUID, reason string) error {
	// Log escalation action
	log := &models.IncidentLog{
		ID:         uuid.New(),
		IncidentID: incidentID,
		Action:     "ESCALATED",
		Notes:      reason,
		Timestamp:  time.Now(),
	}

	if err := s.repo.LogIncidentAction(ctx, log); err != nil {
		return fmt.Errorf("failed to log escalation: %w", err)
	}

	s.logger.Info("Incident escalated",
		zap.String("incident_id", incidentID.String()),
		zap.String("reason", reason),
	)

	// In production:
	// 1. Get current escalation level from metadata
	// 2. Determine next escalation level
	// 3. Notify next level contacts
	// 4. Update incident metadata with escalation info
	// 5. Send notifications (SMS, email, call)

	return nil
}

// GetEscalationPolicy returns the escalation policy for an incident
func (s *EscalationService) GetEscalationPolicy(incident models.SOSIncident) *EscalationPolicy {
	policy := &EscalationPolicy{
		CurrentLevel: Level1,
		NextEscalationAt: incident.Timestamp.Add(
			time.Duration(s.config.AcknowledgementTimeoutMin) * time.Minute,
		),
	}

	// Determine current level based on time and status
	timeSinceTriggered := time.Since(incident.Timestamp)

	if incident.Status == models.StatusAcknowledged {
		// Reset escalation timer after acknowledgement
		policy.NextEscalationAt = incident.Timestamp.Add(
			time.Duration(s.config.ResolutionTimeoutMin) * time.Minute,
		)
	}

	if int(timeSinceTriggered.Minutes()) > s.config.ResolutionTimeoutMin {
		policy.CurrentLevel = Level3
	} else if int(timeSinceTriggered.Minutes()) > s.config.AcknowledgementTimeoutMin {
		policy.CurrentLevel = Level2
	}

	return policy
}

// EscalationPolicy represents the escalation policy for an incident
type EscalationPolicy struct {
	CurrentLevel     EscalationLevel `json:"current_level"`
	NextEscalationAt time.Time       `json:"next_escalation_at"`
	MaxLevel         EscalationLevel `json:"max_level"`
	ContactsNotified []string        `json:"contacts_notified"`
}

// ManualEscalate manually escalates an incident
func (s *EscalationService) ManualEscalate(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, reason string, level EscalationLevel) error {
	incident, err := s.repo.GetIncidentByID(ctx, incidentID)
	if err != nil {
		return fmt.Errorf("failed to get incident: %w", err)
	}
	if incident == nil {
		return ErrIncidentNotFound
	}

	// Log manual escalation
	log := &models.IncidentLog{
		ID:         uuid.New(),
		IncidentID: incidentID,
		UserID:     &userID,
		Action:     fmt.Sprintf("MANUAL_ESCALATION_LEVEL_%d", level),
		Notes:      reason,
		Timestamp:  time.Now(),
	}

	if err := s.repo.LogIncidentAction(ctx, log); err != nil {
		return fmt.Errorf("failed to log manual escalation: %w", err)
	}

	s.logger.Info("Manual incident escalation",
		zap.String("incident_id", incidentID.String()),
		zap.Int("level", int(level)),
		zap.String("reason", reason),
	)

	return nil
}

// GetUnacknowledgedIncidents gets incidents that haven't been acknowledged
func (s *EscalationService) GetUnacknowledgedIncidents(ctx context.Context, minutes int) ([]models.SOSIncident, error) {
	incidents, err := s.repo.ListActiveIncidents(ctx)
	if err != nil {
		return nil, err
	}

	var unacknowledged []models.SOSIncident
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)

	for _, incident := range incidents {
		if incident.Status == models.StatusTriggered && incident.Timestamp.Before(cutoff) {
			unacknowledged = append(unacknowledged, incident)
		}
	}

	return unacknowledged, nil
}

// GetOverdueIncidents gets incidents that haven't been resolved within SLA
func (s *EscalationService) GetOverdueIncidents(ctx context.Context, minutes int) ([]models.SOSIncident, error) {
	incidents, err := s.repo.ListActiveIncidents(ctx)
	if err != nil {
		return nil, err
	}

	var overdue []models.SOSIncident
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)

	for _, incident := range incidents {
		if incident.Status == models.StatusAcknowledged && incident.Timestamp.Before(cutoff) {
			overdue = append(overdue, incident)
		}
	}

	return overdue, nil
}
