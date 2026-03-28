//go:build integration
// +build integration

package service_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	"github.com/flyfleet/fleet-core-service/internal/domains/emergency/service"
	trackingModels "github.com/flyfleet/fleet-core-service/internal/domains/tracking/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockEmergencyRepo is a mock implementation of repository.EmergencyRepository
type MockEmergencyRepo struct {
	mock.Mock
}

func (m *MockEmergencyRepo) CreateIncident(ctx context.Context, incident *models.SOSIncident) error {
	args := m.Called(ctx, incident)
	return args.Error(0)
}

func (m *MockEmergencyRepo) GetIncidentByID(ctx context.Context, id uuid.UUID) (*models.SOSIncident, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SOSIncident), args.Error(1)
}

func (m *MockEmergencyRepo) UpdateIncidentStatus(ctx context.Context, id uuid.UUID, status models.SOSStatus, resolvedBy *uuid.UUID) error {
	args := m.Called(ctx, id, status, resolvedBy)
	return args.Error(0)
}

func (m *MockEmergencyRepo) LogTimelineEvent(ctx context.Context, event *models.IncidentTimeline) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEmergencyRepo) GetActiveIncidentsByOrg(ctx context.Context, orgID uuid.UUID) ([]models.SOSIncident, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]models.SOSIncident), args.Error(1)
}

func (m *MockEmergencyRepo) UpdateDriverFeedback(ctx context.Context, id uuid.UUID, feedback, notes string) error {
	args := m.Called(ctx, id, feedback, notes)
	return args.Error(0)
}

func (m *MockEmergencyRepo) LogIncidentAction(ctx context.Context, log *models.IncidentLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockEmergencyRepo) GetRecentSOSByVehicle(ctx context.Context, vehicleID uuid.UUID, minutes int, incidentType models.IncidentType) (*models.SOSIncident, error) {
	args := m.Called(ctx, vehicleID, minutes, incidentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SOSIncident), args.Error(1)
}

func (m *MockEmergencyRepo) GetIncidentMetrics(ctx context.Context, orgID uuid.UUID, start, end time.Time) (*repository.IncidentMetrics, error) {
	args := m.Called(ctx, orgID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.IncidentMetrics), args.Error(1)
}

func (m *MockEmergencyRepo) GetIncidentAuditHistory(ctx context.Context, id uuid.UUID) (*models.IncidentAuditHistory, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) CreateEscalation(ctx context.Context, esc *models.EmergencyEscalation) error {
	return nil
}
func (m *MockEmergencyRepo) UpdateEscalationStatus(ctx context.Context, id uuid.UUID, status string, respondedAt *time.Time) error {
	return nil
}
func (m *MockEmergencyRepo) CreateContact(ctx context.Context, contact *models.EmergencyContact) error {
	return nil
}
func (m *MockEmergencyRepo) ListContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) CreateResponseUnit(ctx context.Context, unit *models.EmergencyResponseUnit) error {
	return nil
}
func (m *MockEmergencyRepo) ListResponseUnits(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyResponseUnit, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) UpdateVehicleHaltState(ctx context.Context, vehicleID, orgID uuid.UUID, halted bool) error {
	return nil
}
func (m *MockEmergencyRepo) GetVehicleHaltState(ctx context.Context, vehicleID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockEmergencyRepo) GetMediaByIncident(ctx context.Context, id uuid.UUID) ([]models.IncidentMedia, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) GetVehicleSnapshot(ctx context.Context, id uuid.UUID) (*models.IncidentVehicleSnapshot, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) GetIncidentsByDay(ctx context.Context, orgID uuid.UUID, days int) ([]models.DailyIncidentCount, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) GetIncidentTypeDistribution(ctx context.Context, orgID uuid.UUID) ([]models.IncidentTypeCount, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) GetVehicleSafetyScores(ctx context.Context, orgID uuid.UUID) ([]models.VehicleSafetyScore, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) ArchiveOldSnapshots(ctx context.Context, olderThan time.Time) (int64, error) {
	return 0, nil
}
func (m *MockEmergencyRepo) ArchiveOldIncidents(ctx context.Context, olderThan time.Time) (int64, error) {
	return 0, nil
}
func (m *MockEmergencyRepo) GetEmergencyContacts(ctx context.Context, orgID uuid.UUID) ([]models.EmergencyContact, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.EmergencyContact), args.Error(1)
}

func (m *MockEmergencyRepo) CreateMedia(ctx context.Context, media *models.IncidentMedia) error {
	return nil
}
func (m *MockEmergencyRepo) CreateVehicleSnapshot(ctx context.Context, snapshot *models.IncidentVehicleSnapshot) error {
	return nil
}
func (m *MockEmergencyRepo) RecordContactNotification(ctx context.Context, incidentID, contactID uuid.UUID) error {
	return nil
}
func (m *MockEmergencyRepo) UpdateContactResponse(ctx context.Context, incidentID, contactID uuid.UUID, status, notes string) error {
	return nil
}
func (m *MockEmergencyRepo) GetTopLocations(ctx context.Context, orgID uuid.UUID, limit int) ([]string, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) GetIncidentsCount(ctx context.Context, orgID uuid.UUID, since time.Time) (int, error) {
	return 0, nil
}
func (m *MockEmergencyRepo) GetStalledIncidents(ctx context.Context, stalledFor time.Duration) ([]models.SOSIncident, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) EnqueueDeviceEvent(ctx context.Context, vehicleID uuid.UUID, eventType string, payload []byte) error {
	return nil
}
func (m *MockEmergencyRepo) GetPendingDeviceEvents(ctx context.Context, limit int) ([]models.DeviceEventQueueItem, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) MarkDeviceEventDelivered(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockEmergencyRepo) IncrementDeviceEventRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error {
	return nil
}
func (m *MockEmergencyRepo) LogAuthorityDelivery(ctx context.Context, log *models.AuthorityDeliveryLog) error {
	return nil
}
func (m *MockEmergencyRepo) UpdateAuthorityDelivery(ctx context.Context, id uuid.UUID, status string, responseCode int, responseBody string) error {
	return nil
}
func (m *MockEmergencyRepo) CreateTimelineEvent(ctx context.Context, event *models.TimelineEvent) error {
	return nil
}
func (m *MockEmergencyRepo) GetTimelineEvents(ctx context.Context, incidentID uuid.UUID) ([]models.TimelineEvent, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) GetAuthoritiesByLocation(ctx context.Context, lat, lng float64) ([]models.EmergencyResponseUnit, error) {
	return nil, nil
}
func (m *MockEmergencyRepo) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockTrackingRepo struct {
	mock.Mock
}

func (m *MockTrackingRepo) GetLiveLocation(ctx context.Context, vehicleID uuid.UUID) (*trackingModels.LiveLocation, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*trackingModels.LiveLocation), args.Error(1)
}

func (m *MockTrackingRepo) SetLiveLocation(ctx context.Context, loc *trackingModels.LiveLocation) error {
	return nil
}

func TestEmergencyService_OrgIDIsolation(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	repo := new(MockEmergencyRepo)
	trackingRepo := new(MockTrackingRepo)
	logger := zap.NewNop()
	svc := service.NewEmergencyService(repo, trackingRepo, nil, rdb, logger, nil, nil, nil, nil, nil)

	orgA := uuid.New()
	orgB := uuid.New()
	incidentID := uuid.New()
	userID := uuid.New()

	incident := &models.SOSIncident{
		ID:     incidentID,
		OrgID:  orgA,
		Status: models.StatusTriggered,
	}

	repo.On("GetIncidentByID", mock.Anything, incidentID).Return(incident, nil)

	t.Run("Acknowledge with correct OrgID", func(t *testing.T) {
		repo.On("UpdateIncidentStatus", mock.Anything, incidentID, models.StatusAcknowledged, mock.Anything).Return(nil)
		repo.On("LogTimelineEvent", mock.Anything, mock.Anything).Return(nil)
		repo.On("LogIncidentAction", mock.Anything, mock.Anything).Return(nil)

		err := svc.AcknowledgeIncident(context.Background(), incidentID, orgA, userID, "Looks real")
		assert.NoError(t, err)
	})

	t.Run("Acknowledge with wrong OrgID should fail", func(t *testing.T) {
		err := svc.AcknowledgeIncident(context.Background(), incidentID, orgB, userID, "Hack attempt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}

func TestEmergencyService_SimulationMode(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	repo := new(MockEmergencyRepo)
	trackingRepo := new(MockTrackingRepo)
	logger := zap.NewNop()
	svc := service.NewEmergencyService(repo, trackingRepo, nil, rdb, logger, nil, nil, nil, nil, nil)

	orgID := uuid.New()
	vehicleID := uuid.New()

	repo.On("GetRecentSOSByVehicle", mock.Anything, vehicleID, mock.Anything, mock.Anything).Return(nil, nil)
	repo.On("CreateIncident", mock.Anything, mock.MatchedBy(func(i *models.SOSIncident) bool {
		return i.SimulationMode == true
	})).Return(nil)
	repo.On("LogTimelineEvent", mock.Anything, mock.Anything).Return(nil)
	repo.On("LogIncidentAction", mock.Anything, mock.Anything).Return(nil)
	trackingRepo.On("GetLiveLocation", mock.Anything, vehicleID).Return(nil, nil)

	req := models.CreateSOSRequest{
		VehicleID:    vehicleID.String(),
		OrgID:        orgID,
		IncidentType: models.TypePanic,
		Latitude:     19.0,
		Longitude:    72.0,
	}

	incident, err := svc.TriggerSimulatedSOS(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, incident.SimulationMode)

	var meta map[string]interface{}
	json.Unmarshal(incident.Metadata, &meta)
	assert.Equal(t, true, meta["simulated"])
}
