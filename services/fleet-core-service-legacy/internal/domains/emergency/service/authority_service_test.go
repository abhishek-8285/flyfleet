package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type mockEmergencyRepo struct {
	deliveryLogs []*models.AuthorityDeliveryLog
}

func (m *mockEmergencyRepo) LogAuthorityDelivery(ctx context.Context, log *models.AuthorityDeliveryLog) error {
	m.deliveryLogs = append(m.deliveryLogs, log)
	return nil
}

func (m *mockEmergencyRepo) UpdateAuthorityDelivery(ctx context.Context, id uuid.UUID, status string, code int, response string) error {
	for _, l := range m.deliveryLogs {
		if l.ID == id {
			l.DeliveryStatus = status
			l.UpdatedAt = time.Now()
		}
	}
	return nil
}

// Implement other interface methods as no-ops for this test
func (m *mockEmergencyRepo) CreateIncident(ctx context.Context, incident *models.SOSIncident) error { return nil }
func (m *mockEmergencyRepo) GetIncidentByID(ctx context.Context, id uuid.UUID) (*models.SOSIncident, error) { return nil, nil }
func (m *mockEmergencyRepo) UpdateIncidentStatus(ctx context.Context, id uuid.UUID, status models.SOSStatus, userID *uuid.UUID) error { return nil }
func (m *mockEmergencyRepo) GetActiveIncidentsByOrg(ctx context.Context, orgID uuid.UUID) ([]models.SOSIncident, error) { return nil, nil }
func (m *mockEmergencyRepo) GetEmergencyContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error) { return nil, nil }
func (m *mockEmergencyRepo) CreateEscalation(ctx context.Context, escalation *models.EmergencyEscalation) error { return nil }
func (m *mockEmergencyRepo) ListResponseUnits(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyResponseUnit, error) { return nil, nil }
func (m *mockEmergencyRepo) LogIncidentAction(ctx context.Context, log *models.IncidentLog) error { return nil }
func (m *mockEmergencyRepo) LogTimelineEvent(ctx context.Context, event *models.IncidentTimeline) error { return nil }
func (m *mockEmergencyRepo) GetTimeline(ctx context.Context, incidentID uuid.UUID) ([]models.IncidentTimeline, error) { return nil, nil }
func (m *mockEmergencyRepo) CreateMedia(ctx context.Context, media *models.IncidentMedia) error { return nil }
func (m *mockEmergencyRepo) UpdateDriverFeedback(ctx context.Context, incidentID uuid.UUID, feedback, notes string) error { return nil }
func (m *mockEmergencyRepo) GetIncidentMetrics(ctx context.Context, orgID uuid.UUID, start, end time.Time) (*repository.IncidentMetrics, error) { return nil, nil }
func (m *mockEmergencyRepo) CreateVehicleSnapshot(ctx context.Context, snapshot *models.IncidentVehicleSnapshot) error { return nil }
func (m *mockEmergencyRepo) RecordContactNotification(ctx context.Context, incidentID, contactID uuid.UUID) error { return nil }
func (m *mockEmergencyRepo) CreateTimelineEvent(ctx context.Context, event *models.TimelineEvent) error { return nil }
func (m *mockEmergencyRepo) GetTimelineEvents(ctx context.Context, incidentID uuid.UUID) ([]models.TimelineEvent, error) { return nil, nil }
func (m *mockEmergencyRepo) GetAuthoritiesByLocation(ctx context.Context, lat, lng float64) ([]models.EmergencyResponseUnit, error) { return nil, nil }
func (m *mockEmergencyRepo) IncrementDeviceEventRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error { return nil }
func (m *mockEmergencyRepo) GetIncidentTypeDistribution(ctx context.Context, orgID uuid.UUID) ([]models.IncidentTypeCount, error) { return nil, nil }
func (m *mockEmergencyRepo) GetVehicleSafetyScores(ctx context.Context, orgID uuid.UUID) ([]models.VehicleSafetyScore, error) { return nil, nil }
func (m *mockEmergencyRepo) ArchiveOldSnapshots(ctx context.Context, olderThan time.Time) (int64, error) { return 0, nil }
func (m *mockEmergencyRepo) ArchiveOldIncidents(ctx context.Context, olderThan time.Time) (int64, error) { return 0, nil }
func (m *mockEmergencyRepo) LogSLAViolation(ctx context.Context, violation *models.SLAViolation) error { return nil }
func (m *mockEmergencyRepo) ArchiveOldMedia(ctx context.Context, olderThan time.Time) (int64, error) { return 0, nil }
func (m *mockEmergencyRepo) UpdateContactResponse(ctx context.Context, incidentID, contactID uuid.UUID, status string, notes string) error { return nil }
func (m *mockEmergencyRepo) GetTopLocations(ctx context.Context, orgID uuid.UUID, limit int) ([]string, error) { return nil, nil }
func (m *mockEmergencyRepo) GetIncidentsCount(ctx context.Context, orgID uuid.UUID, since time.Time) (int, error) { return 0, nil }
func (m *mockEmergencyRepo) GetStalledIncidents(ctx context.Context, stalledFor time.Duration) ([]models.SOSIncident, error) { return nil, nil }
func (m *mockEmergencyRepo) EnqueueDeviceEvent(ctx context.Context, vehicleID uuid.UUID, eventType string, payload []byte) error { return nil }
func (m *mockEmergencyRepo) GetPendingDeviceEvents(ctx context.Context, limit int) ([]models.DeviceEventQueueItem, error) { return nil, nil }
func (m *mockEmergencyRepo) MarkDeviceEventDelivered(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockEmergencyRepo) CreateContact(ctx context.Context, contact *models.EmergencyContact) error { return nil }
func (m *mockEmergencyRepo) ListContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error) { return nil, nil }
func (m *mockEmergencyRepo) ListActiveIncidents(ctx context.Context) ([]models.SOSIncident, error) { return nil, nil }
func (m *mockEmergencyRepo) GetRecentSOSByVehicle(ctx context.Context, vehicleID uuid.UUID, minutes int, incidentType models.IncidentType) (*models.SOSIncident, error) { return nil, nil }
func (m *mockEmergencyRepo) GetIncidentAuditHistory(ctx context.Context, incidentID uuid.UUID) (*models.IncidentAuditHistory, error) { return nil, nil }
func (m *mockEmergencyRepo) UpdateEscalationStatus(ctx context.Context, id uuid.UUID, status string, respondedAt *time.Time) error { return nil }
func (m *mockEmergencyRepo) CreateResponseUnit(ctx context.Context, unit *models.EmergencyResponseUnit) error { return nil }
func (m *mockEmergencyRepo) UpdateVehicleHaltState(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, halted bool) error { return nil }
func (m *mockEmergencyRepo) GetVehicleHaltState(ctx context.Context, vehicleID uuid.UUID) (bool, error) { return false, nil }
func (m *mockEmergencyRepo) GetMediaByIncident(ctx context.Context, incidentID uuid.UUID) ([]models.IncidentMedia, error) { return nil, nil }
func (m *mockEmergencyRepo) GetVehicleSnapshot(ctx context.Context, incidentID uuid.UUID) (*models.IncidentVehicleSnapshot, error) { return nil, nil }
func (m *mockEmergencyRepo) GetIncidentsByDay(ctx context.Context, orgID uuid.UUID, days int) ([]models.DailyIncidentCount, error) { return nil, nil }
func (m *mockEmergencyRepo) CreateIncidentSnapshot(ctx context.Context, snapshot *models.IncidentSnapshot) error { return nil }
func (m *mockEmergencyRepo) GetCriticalIncidentsForSnapshot(ctx context.Context, since time.Time) ([]models.SOSIncident, error) { return nil, nil }
func (m *mockEmergencyRepo) ExportIncidents(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]models.SOSIncident, error) { return nil, nil }
func (m *mockEmergencyRepo) Ping(ctx context.Context) error { return nil }

func TestEmergencyAuthorityService_NotifyWithRetry(t *testing.T) {
	repo := &mockEmergencyRepo{}
	logger, _ := zap.NewDevelopment()
	service := NewEmergencyAuthorityService(repo, nil, logger)

	incident := &models.SOSIncident{
		ID:        uuid.New(),
		VehicleID: uuid.New(),
		Severity:  models.SeverityCritical,
		Latitude:  12.9716,
		Longitude: 77.5946,
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	// Test notification (NATS is nil, so attemptNotification will return true because s.nc == nil is handled)
	// Actually, look at the code:
	/*
	func (s *EmergencyAuthorityService) attemptNotification(...) bool {
		...
		if s.nc != nil {
			err := s.nc.Publish(...)
			if err != nil { return false }
		}
		return true // returns true if s.nc is nil
	}
	*/
	
	service.notifyWithRetry(ctx, incident, AuthorityPolice)

	if len(repo.deliveryLogs) != 1 {
		t.Errorf("Expected 1 delivery log, got %d", len(repo.deliveryLogs))
	}

	if repo.deliveryLogs[0].DeliveryStatus != "CONFIRMED" {
		t.Errorf("Expected status CONFIRMED, got %s", repo.deliveryLogs[0].DeliveryStatus)
	}
}
