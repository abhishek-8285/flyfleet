package worker

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type mockRepoForReplay struct{}

func (m *mockRepoForReplay) GetTimelineEvents(ctx context.Context, incidentID uuid.UUID) ([]models.TimelineEvent, error) {
	return []models.TimelineEvent{
		{
			ID:         uuid.New(),
			IncidentID: incidentID,
			EventType:  "SOS_TRIGGERED",
			EventData:  []byte(`{"severity":"CRITICAL"}`),
			CreatedAt:  time.Now().Add(-10 * time.Minute),
		},
		{
			ID:         uuid.New(),
			IncidentID: incidentID,
			EventType:  "ACKNOWLEDGED",
			CreatedAt:  time.Now().Add(-5 * time.Minute),
		},
		{
			ID:         uuid.New(),
			IncidentID: incidentID,
			EventType:  "RESOLVED",
			CreatedAt:  time.Now(),
		},
	}, nil
}

func (m *mockRepoForReplay) CreateIncident(ctx context.Context, incident *models.SOSIncident) error {
	return nil
}
func (m *mockRepoForReplay) GetIncidentByID(ctx context.Context, id uuid.UUID) (*models.SOSIncident, error) {
	return nil, nil
}
func (m *mockRepoForReplay) UpdateIncidentStatus(ctx context.Context, id uuid.UUID, status models.SOSStatus, userID *uuid.UUID) error {
	return nil
}
func (m *mockRepoForReplay) GetActiveIncidentsByOrg(ctx context.Context, orgID uuid.UUID) ([]models.SOSIncident, error) {
	return nil, nil
}
func (m *mockRepoForReplay) ListActiveIncidents(ctx context.Context) ([]models.SOSIncident, error) {
	return nil, nil
}
func (m *mockRepoForReplay) GetEmergencyContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error) {
	return nil, nil
}
func (m *mockRepoForReplay) CreateEscalation(ctx context.Context, escalation *models.EmergencyEscalation) error {
	return nil
}
func (m *mockRepoForReplay) ListResponseUnits(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyResponseUnit, error) {
	return nil, nil
}
func (m *mockRepoForReplay) LogIncidentAction(ctx context.Context, log *models.IncidentLog) error {
	return nil
}
func (m *mockRepoForReplay) LogTimelineEvent(ctx context.Context, event *models.IncidentTimeline) error {
	return nil
}
func (m *mockRepoForReplay) GetTimeline(ctx context.Context, incidentID uuid.UUID) ([]models.IncidentTimeline, error) {
	return nil, nil
}
func (m *mockRepoForReplay) CreateMedia(ctx context.Context, media *models.IncidentMedia) error {
	return nil
}
func (m *mockRepoForReplay) UpdateDriverFeedback(ctx context.Context, incidentID uuid.UUID, feedback, notes string) error {
	return nil
}
func (m *mockRepoForReplay) GetIncidentMetrics(ctx context.Context, orgID uuid.UUID, start, end time.Time) (*repository.IncidentMetrics, error) {
	return nil, nil
}
func (m *mockRepoForReplay) CreateVehicleSnapshot(ctx context.Context, snapshot *models.IncidentVehicleSnapshot) error {
	return nil
}
func (m *mockRepoForReplay) RecordContactNotification(ctx context.Context, incidentID, contactID uuid.UUID) error {
	return nil
}
func (m *mockRepoForReplay) CreateTimelineEvent(ctx context.Context, event *models.TimelineEvent) error {
	return nil
}
func (m *mockRepoForReplay) GetAuthoritiesByLocation(ctx context.Context, lat, lng float64) ([]models.EmergencyResponseUnit, error) {
	return nil, nil
}
func (m *mockRepoForReplay) IncrementDeviceEventRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error {
	return nil
}
func (m *mockRepoForReplay) GetRecentSOSByVehicle(ctx context.Context, vehicleID uuid.UUID, minutes int, incidentType models.IncidentType) (*models.SOSIncident, error) {
	return nil, nil
}
func (m *mockRepoForReplay) GetIncidentAuditHistory(ctx context.Context, incidentID uuid.UUID) (*models.IncidentAuditHistory, error) {
	return nil, nil
}
func (m *mockRepoForReplay) UpdateEscalationStatus(ctx context.Context, id uuid.UUID, status string, respondedAt *time.Time) error {
	return nil
}
func (m *mockRepoForReplay) CreateContact(ctx context.Context, contact *models.EmergencyContact) error {
	return nil
}
func (m *mockRepoForReplay) ListContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error) {
	return nil, nil
}
func (m *mockRepoForReplay) UpdateContactResponse(ctx context.Context, incidentID, contactID uuid.UUID, status string, notes string) error {
	return nil
}
func (m *mockRepoForReplay) GetPendingDeviceEvents(ctx context.Context, limit int) ([]models.DeviceEventQueueItem, error) {
	return nil, nil
}
func (m *mockRepoForReplay) CreateResponseUnit(ctx context.Context, unit *models.EmergencyResponseUnit) error {
	return nil
}
func (m *mockRepoForReplay) UpdateVehicleHaltState(ctx context.Context, vehicleID uuid.UUID, orgID uuid.UUID, halted bool) error {
	return nil
}
func (m *mockRepoForReplay) GetVehicleHaltState(ctx context.Context, vehicleID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockRepoForReplay) GetMediaByIncident(ctx context.Context, incidentID uuid.UUID) ([]models.IncidentMedia, error) {
	return nil, nil
}
func (m *mockRepoForReplay) GetVehicleSnapshot(ctx context.Context, incidentID uuid.UUID) (*models.IncidentVehicleSnapshot, error) {
	return nil, nil
}
func (m *mockRepoForReplay) GetIncidentsByDay(ctx context.Context, orgID uuid.UUID, days int) ([]models.DailyIncidentCount, error) {
	return nil, nil
}
func (m *mockRepoForReplay) GetIncidentTypeDistribution(ctx context.Context, orgID uuid.UUID) ([]models.IncidentTypeCount, error) {
	return nil, nil
}
func (m *mockRepoForReplay) GetVehicleSafetyScores(ctx context.Context, orgID uuid.UUID) ([]models.VehicleSafetyScore, error) {
	return nil, nil
}
func (m *mockRepoForReplay) ArchiveOldSnapshots(ctx context.Context, olderThan time.Time) (int64, error) {
	return 0, nil
}
func (m *mockRepoForReplay) ArchiveOldIncidents(ctx context.Context, olderThan time.Time) (int64, error) {
	return 0, nil
}
func (m *mockRepoForReplay) LogSLAViolation(ctx context.Context, violation *models.SLAViolation) error {
	return nil
}
func (m *mockRepoForReplay) ArchiveOldMedia(ctx context.Context, olderThan time.Time) (int64, error) {
	return 0, nil
}
func (m *mockRepoForReplay) GetTopLocations(ctx context.Context, orgID uuid.UUID, limit int) ([]string, error) {
	return nil, nil
}
func (m *mockRepoForReplay) GetIncidentsCount(ctx context.Context, orgID uuid.UUID, since time.Time) (int, error) {
	return 0, nil
}
func (m *mockRepoForReplay) GetStalledIncidents(ctx context.Context, stalledFor time.Duration) ([]models.SOSIncident, error) {
	return nil, nil
}
func (m *mockRepoForReplay) EnqueueDeviceEvent(ctx context.Context, vehicleID uuid.UUID, eventType string, payload []byte) error {
	return nil
}
func (m *mockRepoForReplay) MarkDeviceEventDelivered(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockRepoForReplay) LogAuthorityDelivery(ctx context.Context, log *models.AuthorityDeliveryLog) error {
	return nil
}
func (m *mockRepoForReplay) UpdateAuthorityDelivery(ctx context.Context, id uuid.UUID, status string, responseCode int, responseBody string) error {
	return nil
}
func (m *mockRepoForReplay) CreateIncidentSnapshot(ctx context.Context, snapshot *models.IncidentSnapshot) error {
	return nil
}
func (m *mockRepoForReplay) GetCriticalIncidentsForSnapshot(ctx context.Context, since time.Time) ([]models.SOSIncident, error) {
	return nil, nil
}
func (m *mockRepoForReplay) ExportIncidents(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]models.SOSIncident, error) {
	return nil, nil
}
func (m *mockRepoForReplay) Ping(ctx context.Context) error { return nil }

func TestEmergencyReplayWorker_ReplayIncident(t *testing.T) {
	repo := &mockRepoForReplay{}
	logger, _ := zap.NewDevelopment()
	worker := NewEmergencyReplayWorker(repo, logger)

	incidentID := uuid.New()
	incident, err := worker.ReplayIncident(context.Background(), incidentID)
	if err != nil {
		t.Fatalf("Replay failed: %v", err)
	}

	if incident.Severity != models.SeverityCritical {
		t.Errorf("Expected severity CRITICAL, got %v", incident.Severity)
	}

	if incident.Status != models.StatusResolved {
		t.Errorf("Expected status RESOLVED, got %v", incident.Status)
	}
}
