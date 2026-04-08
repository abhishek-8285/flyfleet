//go:build integration
// +build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/tracking/models"
	"github.com/flyfleet/fleet-core/internal/domains/tracking/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ── In-memory mocks ────────────────────────────────────────────────────────────

type mockTelemetryRepo struct {
	events    map[string]*models.TelemetryEvent // key: eventID
	lastEvent map[uuid.UUID]*models.TelemetryEvent
}

func newMockTelemetryRepo() *mockTelemetryRepo {
	return &mockTelemetryRepo{
		events:    make(map[string]*models.TelemetryEvent),
		lastEvent: make(map[uuid.UUID]*models.TelemetryEvent),
	}
}

func (m *mockTelemetryRepo) IngestEvent(_ context.Context, _ pgx.Tx, e *models.TelemetryEvent) error {
	m.events[e.EventID] = e
	m.lastEvent[e.VehicleID] = e
	return nil
}

func (m *mockTelemetryRepo) IsDuplicate(_ context.Context, eventID, _ string) (bool, error) {
	_, ok := m.events[eventID]
	return ok, nil
}

func (m *mockTelemetryRepo) GetHistory(_ context.Context, _ uuid.UUID, _ models.HistoryQuery) ([]models.TelemetryEvent, error) {
	return nil, nil
}

func (m *mockTelemetryRepo) GetNearby(_ context.Context, _ uuid.UUID, _ models.NearbyQuery) ([]models.LiveLocation, error) {
	return nil, nil
}

func (m *mockTelemetryRepo) GetLatestByVehicle(_ context.Context, _ uuid.UUID, vid uuid.UUID) (*models.TelemetryEvent, error) {
	e, ok := m.lastEvent[vid]
	if !ok {
		return nil, nil
	}
	return e, nil
}

func (m *mockTelemetryRepo) GetTrackingConfig(_ context.Context, _ uuid.UUID) (map[string]any, error) {
	return nil, nil
}

type mockDevRegRepo struct {
	devices map[string]*models.DeviceRegistry
}

func newMockDevRegRepo() *mockDevRegRepo {
	return &mockDevRegRepo{devices: make(map[string]*models.DeviceRegistry)}
}

func (m *mockDevRegRepo) GetByDeviceID(_ context.Context, deviceID string) (*models.DeviceRegistry, error) {
	d, ok := m.devices[deviceID]
	if !ok {
		return nil, nil
	}
	return d, nil
}

func (m *mockDevRegRepo) UpsertDevice(_ context.Context, d *models.DeviceRegistry) error {
	m.devices[d.DeviceID] = d
	return nil
}

func (m *mockDevRegRepo) UpdateLastHeartbeat(_ context.Context, _ string) error { return nil }

func (m *mockDevRegRepo) CreateCommand(_ context.Context, deviceID, cmdType string, _ []byte, _ time.Duration) (string, error) {
	return uuid.NewString(), nil
}

func (m *mockDevRegRepo) GetPendingCommands(_ context.Context, _ string) ([]models.DeviceCommand, error) {
	return nil, nil
}

func (m *mockDevRegRepo) AckCommand(_ context.Context, _, _ string) error { return nil }

type mockOutboxRepo struct{}

func (m *mockOutboxRepo) AddToOutbox(_ context.Context, _ pgx.Tx, _ *models.TelemetryEvent) error {
	return nil
}
func (m *mockOutboxRepo) GetPending(_ context.Context, _ int) ([]repository.OutboxMessage, error) {
	return nil, nil
}
func (m *mockOutboxRepo) MarkProcessed(_ context.Context, _ int64) error       { return nil }
func (m *mockOutboxRepo) MarkFailed(_ context.Context, _ int64, _ error) error { return nil }

type mockMessageBroker struct{}

func (m *mockMessageBroker) Publish(_ string, _ any) error                           { return nil }
func (m *mockMessageBroker) PublishJS(_ string, _ any) error                         { return nil }
func (m *mockMessageBroker) Subscribe(_ string, _ func([]byte)) error                { return nil }
func (m *mockMessageBroker) SubscribeJS(_ string, _, _ string, _ func([]byte)) error { return nil }
func (m *mockMessageBroker) Close() error                                            { return nil }

// ── Setup helper ───────────────────────────────────────────────────────────────

func setupIngestionService() (*IngestionService, *mockTelemetryRepo, *mockDevRegRepo) {
	telemRepo := newMockTelemetryRepo()
	devRepo := newMockDevRegRepo()
	outbox := &mockOutboxRepo{}
	mb := &mockMessageBroker{}
	engine := NewAnomalyEngine(DefaultAnomalyConfig())
	cfg := DefaultIngestionConfig()
	cfg.BufferEnabled = false // Test sync path by default in existing tests
	svc := NewIngestionService(nil, mb, telemRepo, devRepo, outbox, nil, engine, nil, cfg, zap.NewNop())
	return svc, telemRepo, devRepo
}

func registeredDevice(vehicleID uuid.UUID) *models.DeviceRegistry {
	return &models.DeviceRegistry{
		DeviceID:  "DEV-001",
		VehicleID: vehicleID.String(),
		Status:    "active",
	}
}

func validEvent(deviceID string, vehicleID uuid.UUID) *models.TelemetryEvent {
	return &models.TelemetryEvent{
		EventID:    uuid.NewString(),
		DeviceID:   deviceID,
		VehicleID:  vehicleID,
		TenantID:   uuid.New(),
		Lat:        12.9716,
		Lng:        77.5946,
		Speed:      60,
		Heading:    90,
		Ignition:   true,
		EventType:  models.EventTypeGPS,
		RecordedAt: time.Now().UTC(),
	}
}

// ── Tests ──────────────────────────────────────────────────────────────────────

func TestIngestion_ValidEvent(t *testing.T) {
	svc, telemRepo, devRepo := setupIngestionService()
	vehicleID := uuid.New()
	_ = devRepo.UpsertDevice(context.Background(), registeredDevice(vehicleID))

	event := validEvent("DEV-001", vehicleID)
	_, err := svc.Ingest(context.Background(), event)

	require.NoError(t, err)
	assert.Contains(t, telemRepo.events, event.EventID)
}

func TestIngestion_DuplicateEvent(t *testing.T) {
	svc, _, devRepo := setupIngestionService()
	vehicleID := uuid.New()
	_ = devRepo.UpsertDevice(context.Background(), registeredDevice(vehicleID))

	event := validEvent("DEV-001", vehicleID)
	// First ingest — should succeed
	_, err := svc.Ingest(context.Background(), event)
	require.NoError(t, err)

	// Second ingest of the same event — should be rejected
	_, err = svc.Ingest(context.Background(), event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestIngestion_TimestampDrift(t *testing.T) {
	svc, _, devRepo := setupIngestionService()
	vehicleID := uuid.New()
	_ = devRepo.UpsertDevice(context.Background(), registeredDevice(vehicleID))

	event := validEvent("DEV-001", vehicleID)
	event.RecordedAt = time.Now().Add(-10 * time.Minute) // 10 min in past > 300 s

	_, err := svc.Ingest(context.Background(), event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "drift")
}

func TestIngestion_InvalidLatitude(t *testing.T) {
	svc, _, devRepo := setupIngestionService()
	vehicleID := uuid.New()
	_ = devRepo.UpsertDevice(context.Background(), registeredDevice(vehicleID))

	event := validEvent("DEV-001", vehicleID)
	event.Lat = 95.0 // > 90

	_, err := svc.Ingest(context.Background(), event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "latitude")
}

func TestIngestion_SpeedExceedsLimit(t *testing.T) {
	svc, _, devRepo := setupIngestionService()
	vehicleID := uuid.New()
	_ = devRepo.UpsertDevice(context.Background(), registeredDevice(vehicleID))

	event := validEvent("DEV-001", vehicleID)
	event.Speed = 400 // > 300 km/h

	_, err := svc.Ingest(context.Background(), event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "speed")
}

func TestIngestion_DeviceNotRegistered(t *testing.T) {
	svc, _, _ := setupIngestionService()
	// Do NOT register device

	event := validEvent("UNREGISTERED-DEV", uuid.New())
	_, err := svc.Ingest(context.Background(), event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestIngestion_PanicEventTriggersAnomaly(t *testing.T) {
	svc, _, devRepo := setupIngestionService()
	vehicleID := uuid.New()
	_ = devRepo.UpsertDevice(context.Background(), registeredDevice(vehicleID))

	event := validEvent("DEV-001", vehicleID)
	event.EventType = models.EventTypePanic

	anomalies, err := svc.Ingest(context.Background(), event)
	require.NoError(t, err)

	panicAnomaly := findAnomaly(anomalies, models.AnomalyPanic)
	assert.NotNil(t, panicAnomaly, "panic event should produce a PANIC anomaly")
}

func TestIngestion_OverspeedTriggersAnomaly(t *testing.T) {
	svc, _, devRepo := setupIngestionService()
	vehicleID := uuid.New()
	_ = devRepo.UpsertDevice(context.Background(), registeredDevice(vehicleID))

	event := validEvent("DEV-001", vehicleID)
	event.Speed = 150 // > 120 threshold

	anomalies, err := svc.Ingest(context.Background(), event)
	require.NoError(t, err)

	overspeed := findAnomaly(anomalies, models.AnomalyOverspeed)
	assert.NotNil(t, overspeed)
}
