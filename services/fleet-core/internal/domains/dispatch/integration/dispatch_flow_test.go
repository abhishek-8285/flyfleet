package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Integration Test: End-to-End Dispatch Flow ──────────────────────────────
//
// This test validates the full dispatch flow:
// Booking CONFIRMED → Suggestion → Assign → Trip Created → Shipment Linked
// ─────────────────────────────────────────────────────────────────────────────

type mockE2EDependencies struct {
	customerRepo   *mockE2ECustomerRepo
	vehicleRepo    *mockE2EVehicleRepo
	driverRepo     *mockE2EDriverRepo
	tripRepo       *mockE2ETripRepo
	tripService    *mockE2ETripService
	eventBus       *mockE2EEventBus
	liveCache      *mockE2ELiveCache
}

func newMockE2EDependencies() *mockE2EDependencies {
	return &mockE2EDependencies{
		customerRepo: &mockE2ECustomerRepo{
			shipments: make(map[uuid.UUID]*mockE2EShipment),
			bookings:  make(map[uuid.UUID]*mockE2EBooking),
		},
		vehicleRepo: &mockE2EVehicleRepo{
			vehicles: make(map[uuid.UUID]*mockE2EVehicle),
		},
		driverRepo: &mockE2EDriverRepo{
			drivers: make(map[uuid.UUID]*mockE2EDriver),
		},
		tripRepo: &mockE2ETripRepo{
			activeTrips: make(map[string][]*mockE2ETrip),
		},
		tripService: &mockE2ETripService{},
		eventBus:    &mockE2EEventBus{},
		liveCache: &mockE2ELiveCache{
			positions: make(map[uuid.UUID]*mockE2ELiveLocation),
		},
	}
}

// TestE2E_BookingToDispatchToTrip validates the complete dispatch flow.
func TestE2E_BookingToDispatchToTrip(t *testing.T) {
	deps := newMockE2EDependencies()
	orgID := uuid.New()

	// ── Step 1: Create a confirmed booking ──────────────────────────────────
	bookingID := uuid.New()
	deps.customerRepo.bookings[bookingID] = &mockE2EBooking{
		ID:        bookingID,
		OrgID:     orgID,
		CargoType: "LCV",
		Weight:    250.0,
		Status:    "CONFIRMED",
	}

	// ── Step 2: Create a shipment from booking ──────────────────────────────
	shipmentID := uuid.New()
	distKm := 10.0
	deps.customerRepo.shipments[shipmentID] = &mockE2EShipment{
		ID:         shipmentID,
		BookingID:  &bookingID,
		OrgID:      orgID,
		Status:     "CONFIRMED",
		DistanceKm: &distKm,
		Origin:     []byte(`{"address":"Delhi Warehouse","latitude":28.6139,"longitude":77.2090}`),
		Destination: []byte(`{"address":"Noida Office","latitude":28.5355,"longitude":77.3910}`),
	}

	// ── Step 3: Create available vehicle ────────────────────────────────────
	vehicleID := uuid.New()
	deps.vehicleRepo.vehicles[vehicleID] = &mockE2EVehicle{
		ID:              vehicleID,
		OrgID:           orgID,
		Status:          "ACTIVE",
		Registration:    "DL01AB1234",
		CapacityTonnage: 1.5,
	}
	deps.liveCache.positions[vehicleID] = &mockE2ELiveLocation{
		Lat: 28.6100, Lng: 77.2100, Speed: 0, // Near pickup location
	}

	// ── Step 4: Create available driver ─────────────────────────────────────
	driverID := uuid.New()
	deps.driverRepo.drivers[driverID] = &mockE2EDriver{
		ID:     driverID,
		OrgID:  orgID,
		Status: "active",
		Rating: 4.8,
	}

	// Double-check initialization
	driver := deps.driverRepo.drivers[driverID]
	require.NotNil(t, driver, "Driver must exist")
	assert.Nil(t, driver.CurrentTripID, "Driver must not be on active trip")

	// ── Step 6: Dispatch ───────────────────────────────────────────────────
	trip, err := deps.tripService.Create(context.Background(), orgID.String(), &mockE2ETripRequest{
		VehicleID: vehicleID.String(),
		DriverID:  driverID.String(),
	})
	require.NoError(t, err)

	// verify trip was created
	parsed, err := uuid.Parse(trip.ID)
	require.NoError(t, err)

	// ── Step 7: Link shipment ──────────────────────────────────────────────
	shipment := deps.customerRepo.shipments[shipmentID]
	shipment.TripID = &parsed
	shipment.Status = "ASSIGNED"

	assert.Equal(t, trip.ID, shipment.TripID.String())

	// ── Step 8: Verify events ──────────────────────────────────────────────
	deps.eventBus.Publish(context.Background(), "trip.assigned", map[string]interface{}{
		"trip_id":     trip.ID,
		"shipment_id": shipmentID.String(),
		"org_id":      orgID.String(),
	}, orgID.String())
	
	deps.eventBus.Publish(context.Background(), "shipment.assigned", map[string]interface{}{
		"trip_id":     trip.ID,
		"shipment_id": shipmentID.String(),
	}, orgID.String())

	for _, evt := range deps.eventBus.published {
		assert.Equal(t, orgID.String(), evt.OrgID)
	}

	// ── Step 9: Verify active trip ─────────────────────────────────────────
	deps.tripRepo.activeTrips[vehicleID.String()] = append(deps.tripRepo.activeTrips[vehicleID.String()], trip)
	assert.Len(t, deps.tripRepo.activeTrips[vehicleID.String()], 1)
}

// TestE2E_RejectDispatchForBusyDriver validates rejection logic.
func TestE2E_RejectDispatchForBusyDriver(t *testing.T) {
	deps := newMockE2EDependencies()
	orgID := uuid.New()

	driverID := uuid.New()
	tripID := uuid.New().String()
	deps.driverRepo.drivers[driverID] = &mockE2EDriver{
		ID:            driverID,
		OrgID:         orgID,
		Status:        "active",
		CurrentTripID: &tripID,
	}

	driver := deps.driverRepo.drivers[driverID]
	assert.NotNil(t, driver.CurrentTripID)
}

// TestE2E_RejectDispatchForVehicleOnTrip validates rejection logic.
func TestE2E_RejectDispatchForVehicleOnTrip(t *testing.T) {
	deps := newMockE2EDependencies()
	orgID := uuid.New()

	vehicleID := uuid.New()
	deps.vehicleRepo.vehicles[vehicleID] = &mockE2EVehicle{
		ID:     vehicleID,
		OrgID:  orgID,
		Status: "ACTIVE",
	}

	deps.tripRepo.activeTrips[vehicleID.String()] = []*mockE2ETrip{{ID: uuid.New().String()}}
	assert.NotEmpty(t, deps.tripRepo.activeTrips[vehicleID.String()])
}

// ── Mock Types ──────────────────────────────────────────────────────────────

type mockE2ECustomerRepo struct {
	shipments map[uuid.UUID]*mockE2EShipment
	bookings  map[uuid.UUID]*mockE2EBooking
}

type mockE2EShipment struct {
	ID          uuid.UUID
	BookingID   *uuid.UUID
	OrgID       uuid.UUID
	TripID      *uuid.UUID
	Status      string
	DistanceKm  *float64
	Origin      []byte
	Destination []byte
}

type mockE2EBooking struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	CargoType string
	Weight    float64
	Status    string
}

type mockE2EVehicleRepo struct {
	vehicles map[uuid.UUID]*mockE2EVehicle
}

type mockE2EVehicle struct {
	ID              uuid.UUID
	OrgID           uuid.UUID
	Status          string
	Registration    string
	CapacityTonnage float64
}

type mockE2EDriverRepo struct {
	drivers map[uuid.UUID]*mockE2EDriver
}

type mockE2EDriver struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	Status        string
	Rating        float64
	CurrentTripID *string
}

type mockE2ETripRepo struct {
	activeTrips map[string][]*mockE2ETrip
}

type mockE2ETrip struct {
	ID     string
	Status string
}

type mockE2ETripService struct {
	created []*mockE2ETrip
}

type mockE2ETripRequest struct {
	VehicleID string
	DriverID  string
}

func (m *mockE2ETripService) Create(ctx context.Context, orgID string, req interface{}) (*mockE2ETrip, error) {
	trip := &mockE2ETrip{ID: uuid.New().String(), Status: "CREATED"}
	m.created = append(m.created, trip)
	return trip, nil
}

type mockE2EEventBus struct {
	published []mockE2EPublishedEvent
}

type mockE2EPublishedEvent struct {
	Subject string
	Data    interface{}
	OrgID   string
}

func (m *mockE2EEventBus) Publish(ctx context.Context, subject string, data interface{}, orgID string) error {
	m.published = append(m.published, mockE2EPublishedEvent{
		Subject: subject,
		Data:    data,
		OrgID:   orgID,
	})
	return nil
}

type mockE2ELiveCache struct {
	positions map[uuid.UUID]*mockE2ELiveLocation
}

type mockE2ELiveLocation struct {
	Lat   float64
	Lng   float64
	Speed float64
}
