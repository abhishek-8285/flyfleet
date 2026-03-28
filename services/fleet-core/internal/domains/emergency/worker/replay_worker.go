package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type EmergencyReplayWorker struct {
	repo   repository.EmergencyRepository
	logger *zap.Logger
}

func NewEmergencyReplayWorker(repo repository.EmergencyRepository, logger *zap.Logger) *EmergencyReplayWorker {
	return &EmergencyReplayWorker{repo: repo, logger: logger}
}

// ReplayIncident reconstructs the state of an incident from its timeline events.
func (w *EmergencyReplayWorker) ReplayIncident(ctx context.Context, incidentID uuid.UUID) (*models.SOSIncident, error) {
	events, err := w.repo.GetTimelineEvents(ctx, incidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline events for replay: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for incident: %s", incidentID)
	}

	// Reconstruct state
	incident := &models.SOSIncident{ID: incidentID}

	for _, event := range events {
		switch event.EventType {
		case "SOS_TRIGGERED":
			var data map[string]interface{}
			json.Unmarshal(event.EventData, &data)
			if s, ok := data["severity"].(string); ok {
				incident.Severity = models.SOSSeverity(s)
			}
			incident.CreatedAt = event.CreatedAt
		case "ACKNOWLEDGED":
			incident.Status = models.StatusAcknowledged
		case "RESOLVED":
			incident.Status = models.StatusResolved
		case "FALSE_ALARM":
			incident.Status = models.StatusFalseAlarm
		}
	}

	w.logger.Info("Incident replay completed",
		zap.String("incident_id", incidentID.String()),
		zap.String("final_status", string(incident.Status)))

	return incident, nil
}
