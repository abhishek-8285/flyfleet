package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core/internal/domains/customer/models"
	dispatchRepo "github.com/flyfleet/fleet-core/internal/domains/dispatch/repository"
	tripModels "github.com/flyfleet/fleet-core/internal/domains/trip/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─── Port interfaces (hexagonal: dispatch defines its own ports) ─────────────

// TripServicePort defines what the dispatch service needs from the trip domain.
// The real *tripService.TripService implements this interface.
type TripServicePort interface {
	Create(ctx context.Context, tenantID string, req *tripModels.CreateTripRequest) (*tripModels.Trip, error)
	GetTripByID(ctx context.Context, id string) (*tripModels.Trip, error)
	Fail(ctx context.Context, id, reason string) error
	GetStats(ctx context.Context) (*tripModels.TripStats, error)
}

// CustomerRepositoryPort defines what the dispatch service needs from the
// customer domain. The real customerV2Repository implements this interface.
type CustomerRepositoryPort interface {
	GetShipmentByID(ctx context.Context, id uuid.UUID) (*models.Shipment, error)
	GetShipmentByTripID(ctx context.Context, tripID uuid.UUID) (*models.Shipment, error)
	UpdateShipment(ctx context.Context, s *models.Shipment) error
	CountPendingShipments(ctx context.Context, tenantID uuid.UUID) (int, error)
}

// EventBus defines the interface for publishing domain events (best-effort).
type EventBus interface {
	Publish(subject string, data interface{}) error
}

// ─── Public API ───────────────────────────────────────────────────────────────

// DispatchService is the orchestrator that coordinates booking → trip → vehicle
// → driver across domain boundaries.
type DispatchService interface {
	Assign(ctx context.Context, tenantID uuid.UUID, req AssignRequest) (*AssignResult, error)
	Redispatch(ctx context.Context, tenantID uuid.UUID, req RedispatchRequest) (*RedispatchResult, error)
	GetDashboard(ctx context.Context, tenantID uuid.UUID) (*DashboardData, error)
}

type AssignRequest struct {
	ShipmentID uuid.UUID `json:"shipment_id"`
	VehicleID  uuid.UUID `json:"vehicle_id"`
	DriverID   uuid.UUID `json:"driver_id"`
}

type AssignResult struct {
	TripID         string `json:"trip_id"`
	Status         string `json:"status"`
	ShipmentStatus string `json:"shipment_status"`
}

type RedispatchRequest struct {
	TripID       string    `json:"trip_id"`
	NewVehicleID uuid.UUID `json:"new_vehicle_id"`
	NewDriverID  uuid.UUID `json:"new_driver_id"`
	Reason       string    `json:"reason"`
}

type RedispatchResult struct {
	NewTripID string `json:"new_trip_id"`
	Status    string `json:"status"`
}

type DashboardData struct {
	PendingShipments  int `json:"pending_shipments"`
	ScheduledTrips    int `json:"scheduled_trips"`
	InProgressTrips   int `json:"in_progress_trips"`
	CompletedTrips    int `json:"completed_trips"`
	AvailableVehicles int `json:"available_vehicles"`
	AvailableDrivers  int `json:"available_drivers"`
}

// ─── Implementation ───────────────────────────────────────────────────────────

type dispatchService struct {
	tripSvc      TripServicePort
	customerRepo CustomerRepositoryPort
	dispatchRepo dispatchRepo.DispatchRepository
	eventBus     EventBus
	logger       *zap.Logger
}

func NewDispatchService(
	tripSvc TripServicePort,
	customerRepo CustomerRepositoryPort,
	dispatchRepo dispatchRepo.DispatchRepository,
	eb EventBus,
	logger *zap.Logger,
) DispatchService {
	return &dispatchService{
		tripSvc:      tripSvc,
		customerRepo: customerRepo,
		dispatchRepo: dispatchRepo,
		eventBus:     eb,
		logger:       logger,
	}
}

// ─── Assign ───────────────────────────────────────────────────────────────────

// Assign creates a real trip from a confirmed shipment and links them.
// It validates shipment status, checks vehicle availability, creates the trip
// via TripService, updates the shipment link, records an audit entry, and
// publishes domain events.
func (s *dispatchService) Assign(ctx context.Context, tenantID uuid.UUID, req AssignRequest) (*AssignResult, error) {
	// 1. Fetch the shipment
	shipment, err := s.customerRepo.GetShipmentByID(ctx, req.ShipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shipment: %w", err)
	}
	if shipment == nil {
		return nil, fmt.Errorf("shipment %s not found", req.ShipmentID)
	}

	// 2. Validate shipment status — must be CONFIRMED
	if shipment.Status != "CONFIRMED" {
		return nil, fmt.Errorf("shipment must be in CONFIRMED status, got: %s", shipment.Status)
	}

	// 3. Validate shipment has no existing trip
	if shipment.TripID != nil && *shipment.TripID != uuid.Nil {
		return nil, fmt.Errorf("shipment already assigned to trip %s", shipment.TripID)
	}

	// 4. Check vehicle availability (no active trips)
	if s.dispatchRepo != nil {
		activeTrips, err := s.dispatchRepo.GetActiveTripsForVehicle(ctx, req.VehicleID)
		if err != nil {
			s.logger.Warn("failed to check vehicle availability, proceeding anyway", zap.Error(err))
		} else if len(activeTrips) > 0 {
			return nil, fmt.Errorf("vehicle %s has %d active trips, cannot assign", req.VehicleID, len(activeTrips))
		}
	}

	// 5. Build CreateTripRequest from shipment data, then inject selected vehicle/driver
	tripReq, err := buildTripRequestFromShipment(shipment)
	if err != nil {
		return nil, fmt.Errorf("failed to build trip request: %w", err)
	}
	tripReq.VehicleID = req.VehicleID.String()
	tripReq.DriverID = req.DriverID.String()

	// 6. Create the trip via TripService
	trip, err := s.tripSvc.Create(ctx, tenantID.String(), tripReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create trip: %w", err)
	}

	// 7. Link shipment to trip
	tripUUID, err := uuid.Parse(trip.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid trip ID returned: %w", err)
	}
	shipment.TripID = &tripUUID
	shipment.Status = "ASSIGNED"
	if err := s.customerRepo.UpdateShipment(ctx, shipment); err != nil {
		s.logger.Error("failed to link shipment to trip",
			zap.String("trip_id", trip.ID),
			zap.String("shipment_id", req.ShipmentID.String()),
			zap.Error(err))
		// Don't fail the assignment — the trip is created, just log the link failure
	}

	// 8. Audit: record dispatch action
	if s.dispatchRepo != nil {
		assignmentTx := &dispatchRepo.AssignmentTx{
			ShipmentID: req.ShipmentID,
			TripID:     trip.ID,
			VehicleID:  req.VehicleID,
			DriverID:   req.DriverID,
			TenantID:   tenantID,
			ActionType: dispatchRepo.ActionManualAssign,
		}
		if _, err := s.dispatchRepo.ExecuteAssignment(ctx, assignmentTx); err != nil {
			s.logger.Warn("failed to record dispatch action", zap.Error(err))
		}
	}

	// 9. Publish events (best-effort)
	s.publishEvent("trip.assigned", map[string]interface{}{
		"trip_id":     trip.ID,
		"shipment_id": req.ShipmentID.String(),
		"vehicle_id":  req.VehicleID.String(),
		"driver_id":   req.DriverID.String(),
		"tenant_id":   tenantID.String(),
	})
	s.publishEvent("shipment.assigned", map[string]interface{}{
		"shipment_id": req.ShipmentID.String(),
		"trip_id":     trip.ID,
		"tenant_id":   tenantID.String(),
	})

	s.logger.Info("shipment dispatched to vehicle",
		zap.String("trip_id", trip.ID),
		zap.String("shipment_id", req.ShipmentID.String()),
		zap.String("vehicle_id", req.VehicleID.String()),
	)

	return &AssignResult{
		TripID:         trip.ID,
		Status:         "SCHEDULED",
		ShipmentStatus: "ASSIGNED",
	}, nil
}

// ─── Redispatch ───────────────────────────────────────────────────────────────

// Redispatch fails the existing trip (if in CREATED or SCHEDULED) and creates
// a new one with a different vehicle/driver, re-linking the shipment.
func (s *dispatchService) Redispatch(ctx context.Context, tenantID uuid.UUID, req RedispatchRequest) (*RedispatchResult, error) {
	// 1. Fetch the existing trip to get source data for the new trip
	trip, err := s.tripSvc.GetTripByID(ctx, req.TripID)
	if err != nil {
		return nil, fmt.Errorf("trip %s not found: %w", req.TripID, err)
	}
	if trip == nil {
		return nil, fmt.Errorf("trip %s not found", req.TripID)
	}

	// 2. Validate trip is in a state that can be redispatched
	if trip.Status != tripModels.TripStatusCreated && trip.Status != tripModels.TripStatusScheduled {
		return nil, fmt.Errorf("trip must be in CREATED or SCHEDULED status to redispatch, got: %s", trip.Status)
	}

	// 3. Find the shipment linked to this trip
	var shipmentID *uuid.UUID
	var shipment *models.Shipment
	if s.customerRepo != nil {
		tripUUID, parseErr := uuid.Parse(req.TripID)
		if parseErr == nil {
			shipment, _ = s.customerRepo.GetShipmentByTripID(ctx, tripUUID)
			if shipment != nil {
				shipmentID = &shipment.ID
				}
		}
	}

	// 4. Check new vehicle availability
	if s.dispatchRepo != nil {
		activeTrips, err := s.dispatchRepo.GetActiveTripsForVehicle(ctx, req.NewVehicleID)
		if err != nil {
			s.logger.Warn("failed to check new vehicle availability", zap.Error(err))
		} else if len(activeTrips) > 0 {
			return nil, fmt.Errorf("new vehicle %s has %d active trips", req.NewVehicleID, len(activeTrips))
		}
	}

	// 5. Fail the existing trip
	if err := s.tripSvc.Fail(ctx, req.TripID, fmt.Sprintf("Redispatched: %s", req.Reason)); err != nil {
		s.logger.Warn("failed to fail original trip during redispatch",
			zap.String("trip_id", req.TripID), zap.Error(err))
	}

	// 6. Create a new trip with the same data, new vehicle/driver
	tripReq, err := buildTripRequestFromExistingTrip(trip, req.NewVehicleID, req.NewDriverID)
	if err != nil {
		return nil, fmt.Errorf("failed to build redispatch trip request: %w", err)
	}

	newTrip, err := s.tripSvc.Create(ctx, tenantID.String(), tripReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create new trip: %w", err)
	}

	// 7. Re-link shipment to the new trip
	if shipment != nil {
		newTripUUID, _ := uuid.Parse(newTrip.ID)
		shipment.TripID = &newTripUUID
		shipment.Status = "ASSIGNED"
		if err := s.customerRepo.UpdateShipment(ctx, shipment); err != nil {
			s.logger.Warn("failed to re-link shipment to new trip",
				zap.String("trip_id", newTrip.ID),
				zap.String("shipment_id", shipment.ID.String()),
				zap.Error(err))
		}
	}

	// 8. Audit: record redispatch action
	if s.dispatchRepo != nil {
		assignmentTx := &dispatchRepo.AssignmentTx{
			TripID:     newTrip.ID,
			VehicleID:  req.NewVehicleID,
			DriverID:   req.NewDriverID,
			TenantID:   tenantID,
			ActionType: dispatchRepo.ActionRedispatch,
		}
		if shipmentID != nil {
			assignmentTx.ShipmentID = *shipmentID
		}
		if _, err := s.dispatchRepo.ExecuteAssignment(ctx, assignmentTx); err != nil {
			s.logger.Warn("failed to record redispatch action", zap.Error(err))
		}
	}

	// 9. Publish events
	s.publishEvent("trip.reassigned", map[string]interface{}{
		"old_trip_id": req.TripID,
		"new_trip_id": newTrip.ID,
		"shipment_id": shipmentID,
		"vehicle_id":  req.NewVehicleID.String(),
		"driver_id":   req.NewDriverID.String(),
		"tenant_id":   tenantID.String(),
	})

	s.logger.Info("trip redispatched",
		zap.String("old_trip_id", req.TripID),
		zap.String("new_trip_id", newTrip.ID),
	)

	return &RedispatchResult{
		NewTripID: newTrip.ID,
		Status:    "SCHEDULED",
	}, nil
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

// GetDashboard returns a dispatch dashboard summary with real counts from the
// trip and customer domains. Degrades gracefully when dependencies are
// unavailable (e.g. in tests with nil deps).
func (s *dispatchService) GetDashboard(ctx context.Context, tenantID uuid.UUID) (*DashboardData, error) {
	data := &DashboardData{}

	if s.tripSvc != nil {
		stats, err := s.tripSvc.GetStats(ctx)
		if err != nil {
			s.logger.Warn("failed to get trip stats for dashboard", zap.Error(err))
		} else if stats != nil {
			data.InProgressTrips = stats.ActiveTrips
		}
	}

	if s.customerRepo != nil {
		pending, err := s.customerRepo.CountPendingShipments(ctx, tenantID)
		if err != nil {
			s.logger.Warn("failed to count pending shipments for dashboard", zap.Error(err))
		} else {
			data.PendingShipments = pending
		}
	}

	return data, nil
}

// ─── Event Publishing ─────────────────────────────────────────────────────────

func (s *dispatchService) publishEvent(subject string, data interface{}) {
	if s.eventBus == nil {
		return
	}
	if err := s.eventBus.Publish(subject, data); err != nil {
		s.logger.Warn("failed to publish event",
			zap.String("subject", subject), zap.Error(err))
	}
}

// ─── Trip Request Builders ────────────────────────────────────────────────────

// buildTripRequestFromShipment maps a Shipment into a CreateTripRequest.
func buildTripRequestFromShipment(sh *models.Shipment) (*tripModels.CreateTripRequest, error) {
	pickupGeo := parseLocationByte(sh.Origin)
	dropGeo := parseLocationByte(sh.Destination)

	waypoints := []tripModels.Waypoint{
		{
			Sequence: 1,
			Type:     "pickup",
			Location:  dereferenceLocation(pickupGeo),
		},
		{
			Sequence: 2,
			Type:     "dropoff",
			Location:  dereferenceLocation(dropGeo),
		},
	}

	var distanceKm float64
	if sh.DistanceKm != nil {
		distanceKm = *sh.DistanceKm
	}
	estimatedDuration := int(distanceKm * 2.5) // rough ETA: 2.5 min/km
	if estimatedDuration == 0 {
		estimatedDuration = 60
	}

	customerName := fmt.Sprintf("Customer %s", sh.CustomerID.String()[:8])

	return &tripModels.CreateTripRequest{
		Title:                fmt.Sprintf("Dispatch shipment %s", sh.ID.String()[:8]),
		Description:          fmt.Sprintf("Auto-generated trip for shipment %s", sh.ID),
		PlannedRoute:         waypoints,
		Waypoints:            waypoints,
		PlannedDistance:      distanceKm,
		EstimatedDurationMin: estimatedDuration,
		CargoDetails: tripModels.CargoDetails{
			Description: fmt.Sprintf("Shipment %s", sh.ID),
		},
		ScheduledStart:    time.Now(),
		CustomerName:      customerName,
		CustomerPhone:     "+910000000000",
		PickupGeo:         pickupGeo,
		DropGeo:           dropGeo,
		GSTClassification: "E",
		TripType:          tripModels.TripTypeDelivery,
	}, nil
}

// dereferenceLocation safely converts a *Location to a value, returning a zero
// Location if the pointer is nil (Waypoint.Location is a value type).
func dereferenceLocation(loc *tripModels.Location) tripModels.Location {
	if loc == nil {
		return tripModels.Location{}
	}
	return *loc
}

// buildTripRequestFromExistingTrip creates a CreateTripRequest from an existing
// trip's data, overriding only the vehicle and driver.
func buildTripRequestFromExistingTrip(t *tripModels.Trip, newVehicleID, newDriverID uuid.UUID) (*tripModels.CreateTripRequest, error) {
	waypoints := t.Waypoints
	if len(waypoints) == 0 {
		waypoints = t.PlannedRoute
	}
	if len(waypoints) == 0 {
		waypoints = []tripModels.Waypoint{
			{Type: "pickup", Location: dereferenceLocation(t.PickupGeo)},
			{Type: "dropoff", Location: dereferenceLocation(t.DropGeo)},
		}
	}

	req := &tripModels.CreateTripRequest{
		VehicleID:            newVehicleID.String(),
		DriverID:             newDriverID.String(),
		Title:                t.Title,
		Description:          t.Description,
		PlannedRoute:         t.PlannedRoute,
		Waypoints:            waypoints,
		PlannedDistance:      t.PlannedDistance,
		EstimatedDurationMin: t.EstimatedDurationMin,
		CargoDetails:         t.CargoDetails,
		ScheduledStart:       time.Now(),
		PickupGeo:            t.PickupGeo,
		DropGeo:              t.DropGeo,
		EWayBillNo:           t.EWayBillNo,
		GSTClassification:    t.GSTClassification,
		CustomerName:         t.CustomerName,
		CustomerPhone:        t.CustomerPhone,
		CustomerEmail:        t.CustomerEmail,
		TripType:             t.TripType,
	}
	return req, nil
}

// parseLocationByte unmarshals a JSON byte blob from the Shipment.Origin /
// Destination columns into a tripModels.Location pointer.
func parseLocationByte(raw []byte) *tripModels.Location {
	if len(raw) == 0 {
		return nil
	}
	var loc struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Address   string  `json:"address"`
	}
	if err := json.Unmarshal(raw, &loc); err != nil {
		return nil
	}
	return &tripModels.Location{
		Latitude:  loc.Latitude,
		Longitude: loc.Longitude,
		Address:   loc.Address,
	}
}
