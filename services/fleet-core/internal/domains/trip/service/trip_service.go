package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	emergencyRepo "github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	tripmetrics "github.com/flyfleet/fleet-core-service/internal/domains/trip"
	"github.com/flyfleet/fleet-core-service/internal/domains/trip/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/trip/repository"
	"github.com/flyfleet/fleet-core-service/internal/middleware"
	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	"github.com/flyfleet/fleet-core-service/internal/platform/metrics"
)

// TripServiceInterface defines the interface for trip service
type TripServiceInterface interface {
	// CRUD
	Create(ctx context.Context, orgID string, req *models.CreateTripRequest) (*models.Trip, error)
	GetTripByID(ctx context.Context, id string) (*models.Trip, error)
	List(ctx context.Context, filter models.TripListFilter) ([]*models.Trip, int, error)
	Update(ctx context.Context, id string, req *models.UpdateTripRequest) (*models.Trip, error)
	Delete(ctx context.Context, id string) error

	// Lifecycle
	Approve(ctx context.Context, id string) error
	Start(ctx context.Context, id, driverID string, req *models.StartTripRequest) (*models.Trip, error)
	Pause(ctx context.Context, id, driverID string) (*models.Trip, error)
	Resume(ctx context.Context, id, driverID string) (*models.Trip, error)
	Complete(ctx context.Context, id, driverID string, req *models.CompleteTripRequest) (*models.Trip, error)
	Cancel(ctx context.Context, id, reason string) error
	Fail(ctx context.Context, id, reason string) error

	// Events & POD
	AddEvent(ctx context.Context, tripID, driverID string, req *models.AddEventRequest) error
	SubmitPOD(ctx context.Context, tripID, driverID string, req *models.SubmitPODRequest) error

	// Offers
	OfferTrip(ctx context.Context, tripID, driverID string) (*models.TripOffer, error)
	RespondToOffer(ctx context.Context, offerID, driverID, response, reason string) error
	GetDriverOffers(ctx context.Context, driverID string) ([]*models.TripOffer, error)

	// Analytics
	GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]*models.Trip, error)
	GetTripSummary(ctx context.Context, orgID string, startDate, endDate time.Time) (*models.TripSummary, error)
	GetDriverStats(ctx context.Context, orgID string) ([]models.DriverPerformance, error)
	GetTripStats(ctx context.Context, orgID string) ([]models.DailyTripTrend, error)
	GetStats(ctx context.Context) (*models.TripStats, error)

	// Telemetry
	GetTelemetry(ctx context.Context, tripID string) ([]models.TelemetryPoint, error)
}

type TripService struct {
	repo   repository.TripRepository
	logger *zap.Logger
	nc     *nats.Conn
	emRepo emergencyRepo.EmergencyRepository
}

func NewTripService(repo repository.TripRepository, logger *zap.Logger, opts ...func(*TripService)) *TripService {
	s := &TripService{
		repo:   repo,
		logger: logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithNATS sets the NATS connection for event publishing
func WithNATS(nc *nats.Conn) func(*TripService) {
	return func(s *TripService) {
		s.nc = nc
	}
}

// WithEmergencyRepository sets the emergency repository for halt checks
func WithEmergencyRepository(repo emergencyRepo.EmergencyRepository) func(*TripService) {
	return func(s *TripService) {
		s.emRepo = repo
	}
}

// publishEvent publishes a NATS event (best-effort, non-blocking)
func (s *TripService) publishEvent(subject string, data interface{}) {
	if s.nc == nil {
		return
	}
	payload, err := json.Marshal(data)
	if err != nil {
		s.logger.Warn("failed to marshal NATS event", zap.String("subject", subject), zap.Error(err))
		return
	}
	if err := s.nc.Publish(subject, payload); err != nil {
		s.logger.Warn("failed to publish NATS event", zap.String("subject", subject), zap.Error(err))
	}
}

// Create creates a new trip
func (s *TripService) Create(ctx context.Context, orgID string, req *models.CreateTripRequest) (*models.Trip, error) {
	// Business Rule: Duplicate trip detection (overlapping windows)
	endTime := req.ScheduledStart.Add(time.Duration(req.EstimatedDurationMin) * time.Minute)
	if req.PlannedETA != nil {
		endTime = *req.PlannedETA
	}
	overlapping, err := s.repo.CheckOverlappingTrips(ctx, req.VehicleID, req.DriverID, req.ScheduledStart, endTime)
	if err != nil {
		s.logger.Warn("failed to check overlapping trips", zap.Error(err))
	} else if overlapping {
		return nil, fmt.Errorf("overlapping trip detected for vehicle or driver in this time window")
	}

	trip := &models.Trip{
		ID:                   uuid.New().String(),
		OrgID:                orgID,
		VehicleID:            req.VehicleID,
		DriverID:             &req.DriverID,
		Status:               models.TripStatusCreated,
		Title:                req.Title,
		Description:          req.Description,
		PickupGeo:            req.PickupGeo,
		DropGeo:              req.DropGeo,
		PlannedRoute:         req.PlannedRoute,
		Waypoints:            req.Waypoints,
		PlannedDistance:      req.PlannedDistance,
		EstimatedDurationMin: req.EstimatedDurationMin,
		CargoDetails:         req.CargoDetails,
		ScheduledStart:       req.ScheduledStart,
		PlannedETA:           req.PlannedETA,
		EWayBillNo:           req.EWayBillNo,
		GSTClassification:    req.GSTClassification,
		CustomerName:         req.CustomerName,
		CustomerPhone:        req.CustomerPhone,
		CustomerEmail:        req.CustomerEmail,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := s.repo.Create(ctx, trip); err != nil {
		s.logger.Error("failed to create trip", zap.Error(err))
		return nil, fmt.Errorf("failed to create trip: %w", err)
	}

	middleware.RecordTripEvent(middleware.MetricTripCreated)
	s.logger.Info("trip created",
		zap.String("trip_id", trip.ID),
		zap.String("correlation_id", trip.CorrelationID),
		zap.String("org_id", trip.OrgID),
	)
	return trip, nil
}

// GetTripByID retrieves a trip by its ID (with tenant validation)
func (s *TripService) GetTripByID(ctx context.Context, id string) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get trip", zap.String("trip_id", id), zap.Error(err))
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != trip.OrgID {
			s.logger.Warn("unauthorized trip access attempt",
				zap.String("trip_id", id),
				zap.String("trip_tenant", trip.OrgID),
				zap.String("request_tenant", orgID))
			return nil, fmt.Errorf("trip not found") // Opaque error to prevent tenant enumeration
		}
	}
	// If no tenant context, allow (for backward compatibility or service-to-service calls)

	return trip, nil
}

// List retrieves trips with filtering
func (s *TripService) List(ctx context.Context, filter models.TripListFilter) ([]*models.Trip, int, error) {
	trips, totalCount, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("failed to list trips", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list trips: %w", err)
	}

	return trips, totalCount, nil
}

// Update modifies an existing trip
func (s *TripService) Update(ctx context.Context, id string, req *models.UpdateTripRequest) (*models.Trip, error) {
	// First get the existing trip to validate ownership
	existingTrip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != existingTrip.OrgID {
			s.logger.Warn("unauthorized trip update attempt",
				zap.String("trip_id", id),
				zap.String("trip_tenant", existingTrip.OrgID),
				zap.String("request_tenant", orgID))
			return nil, fmt.Errorf("trip not found")
		}
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.PlannedRoute != nil {
		updates["planned_route"] = *req.PlannedRoute
	}
	if req.Waypoints != nil {
		updates["waypoints"] = *req.Waypoints
	}
	if req.PlannedDistance != nil {
		updates["planned_distance"] = *req.PlannedDistance
	}
	if req.EstimatedDurationMin != nil {
		updates["estimated_duration_min"] = *req.EstimatedDurationMin
	}
	if req.CargoDetails != nil {
		updates["cargo_details"] = *req.CargoDetails
	}
	if req.ScheduledStart != nil {
		updates["scheduled_start"] = *req.ScheduledStart
	}
	if req.CustomerName != nil {
		updates["customer_name"] = *req.CustomerName
	}
	if req.CustomerPhone != nil {
		updates["customer_phone"] = *req.CustomerPhone
	}
	if req.CustomerEmail != nil {
		updates["customer_email"] = *req.CustomerEmail
	}

	if len(updates) == 0 {
		return existingTrip, nil
	}

	// Only allow updates to created or scheduled trips
	if existingTrip.Status != models.TripStatusCreated && existingTrip.Status != models.TripStatusScheduled {
		return nil, fmt.Errorf("can only update trips in CREATED or SCHEDULED status")
	}

	updates["updated_at"] = time.Now()

	if err := s.repo.Update(ctx, id, updates); err != nil {
		s.logger.Error("failed to update trip", zap.String("trip_id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to update trip: %w", err)
	}

	// Return updated trip
	return s.repo.GetByID(ctx, id)
}

// Delete removes a trip
func (s *TripService) Delete(ctx context.Context, id string) error {
	// First get the trip to validate it exists and can be deleted
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != trip.OrgID {
			s.logger.Warn("unauthorized trip delete attempt",
				zap.String("trip_id", id),
				zap.String("trip_tenant", trip.OrgID),
				zap.String("request_tenant", orgID))
			return fmt.Errorf("trip not found")
		}
	}

	// Only allow deletion of created or scheduled trips
	if trip.Status != models.TripStatusCreated && trip.Status != models.TripStatusScheduled {
		return fmt.Errorf("can only delete trips in CREATED or SCHEDULED status")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete trip", zap.String("trip_id", id), zap.Error(err))
		return fmt.Errorf("failed to delete trip: %w", err)
	}

	s.logger.Info("trip deleted", zap.String("trip_id", id))
	return nil
}

// Approve moves a trip from CREATED to SCHEDULED
func (s *TripService) Approve(ctx context.Context, id string) error {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal == nil {
		s.logger.Error("tenant context missing")
		return fmt.Errorf("unauthorized")
	}
	orgID, ok := orgIDVal.(string)
	if !ok || orgID != trip.OrgID {
		s.logger.Warn("unauthorized trip approval attempt",
			zap.String("trip_id", id),
			zap.String("trip_tenant", trip.OrgID),
			zap.Any("request_tenant", orgIDVal))
		return fmt.Errorf("unauthorized")
	}

	if trip.Status != models.TripStatusCreated {
		return fmt.Errorf("trip cannot be approved from status: %s", trip.Status)
	}

	if err := s.repo.Approve(ctx, id); err != nil {
		s.logger.Error("failed to approve trip", zap.String("trip_id", id), zap.Error(err))
		return fmt.Errorf("failed to approve trip: %w", err)
	}

	s.logger.Info("trip approved", zap.String("trip_id", id))
	return nil
}

// Start begins a trip
func (s *TripService) Start(ctx context.Context, id, driverID string, req *models.StartTripRequest) (*models.Trip, error) {
	// Get existing trip
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != trip.OrgID {
			s.logger.Warn("unauthorized trip start attempt",
				zap.String("trip_id", id),
				zap.String("trip_tenant", trip.OrgID),
				zap.String("request_tenant", orgID))
			return nil, fmt.Errorf("trip not found")
		}
	}

	// Validate authorization
	if trip.DriverID == nil || *trip.DriverID != driverID {
		return nil, fmt.Errorf("unauthorized: driver does not own this trip")
	}

	// Validate trip status — Spec: SCHEDULED → STARTED only
	if trip.Status != models.TripStatusScheduled {
		return nil, fmt.Errorf("trip cannot be started from status %s: must be in SCHEDULED state", trip.Status)
	}

	// Business Rule: One active trip per vehicle
	activeVehicleTrips, _ := s.repo.GetActiveTripsForVehicle(ctx, trip.VehicleID)
	for _, t := range activeVehicleTrips {
		if t.ID != id {
			return nil, fmt.Errorf("vehicle %s already has an active trip: %s", trip.VehicleID, t.ID)
		}
	}

	// Business Rule: One active trip per driver
	activeDriverTrips, _ := s.repo.GetActiveTripsForDriver(ctx, driverID)
	for _, t := range activeDriverTrips {
		if t.ID != id {
			return nil, fmt.Errorf("driver %s already has an active trip: %s", driverID, t.ID)
		}
	}

	// Business Rule: Emergency Halt validation
	if s.emRepo != nil {
		vehicleID, _ := uuid.Parse(trip.VehicleID)
		halted, err := s.emRepo.GetVehicleHaltState(ctx, vehicleID)
		if err == nil && halted {
			return nil, fmt.Errorf("vehicle %s is halted due to an active emergency: cannot start trip", trip.VehicleID)
		}
	}

	// Business Rule: Permit validation before START
	if !trip.PermitVerified {
		return nil, fmt.Errorf("vehicle permit must be verified before starting the trip")
	}

	// Business Rule: PUC & Insurance validity check (Placeholder)
	// In a real system, this would call VAHAN API or check a compliance service
	s.logger.Debug("verifying PUC and Insurance validity", zap.String("vehicle_id", trip.VehicleID))

	now := time.Now()
	updates := map[string]interface{}{
		"status":              models.TripStatusStarted,
		"start_time":          &now,
		"start_odometer":      req.OdometerReading,
		"fuel_level_snapshot": req.FuelLevel,
		"ais140_compliant":    true, // Placeholder for actual validation
		"updated_at":          now,
		"version":             trip.Version, // For optimistic locking
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		s.logger.Error("failed to start trip", zap.String("trip_id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to start trip: %w", err)
	}

	s.logger.Info("trip started", zap.String("trip_id", id))

	// Record metrics
	orgID := trip.OrgID
	tripType := "standard"   // Placeholder: derive from trip type/category if available
	vehicleType := "unknown" // Placeholder: derive from vehicle type if available
	tripmetrics.TripsStartedTotal.WithLabelValues(orgID, tripType, vehicleType).Inc()
	metrics.ActiveTrips.Inc()

	// Return updated trip
	return s.repo.GetByID(ctx, id)
}

// Pause temporarily stops a trip
func (s *TripService) Pause(ctx context.Context, id, driverID string) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != trip.OrgID {
			s.logger.Warn("unauthorized trip pause attempt",
				zap.String("trip_id", id),
				zap.String("trip_tenant", trip.OrgID),
				zap.String("request_tenant", orgID))
			return nil, fmt.Errorf("trip not found")
		}
	}

	if trip.DriverID == nil || *trip.DriverID != driverID {
		return nil, fmt.Errorf("unauthorized: driver does not own this trip")
	}

	if trip.Status != models.TripStatusStarted && trip.Status != models.TripStatusResumed {
		return nil, fmt.Errorf("trip must be started or resumed to pause")
	}

	updates := map[string]interface{}{
		"status":     models.TripStatusPaused,
		"updated_at": time.Now(),
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		s.logger.Error("failed to pause trip", zap.String("trip_id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to pause trip: %w", err)
	}

	s.logger.Info("trip paused", zap.String("trip_id", id))
	return s.repo.GetByID(ctx, id)
}

// Resume continues a paused trip
func (s *TripService) Resume(ctx context.Context, id, driverID string) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != trip.OrgID {
			s.logger.Warn("unauthorized trip resume attempt",
				zap.String("trip_id", id),
				zap.String("trip_tenant", trip.OrgID),
				zap.String("request_tenant", orgID))
			return nil, fmt.Errorf("trip not found")
		}
	}

	if trip.DriverID == nil || *trip.DriverID != driverID {
		return nil, fmt.Errorf("unauthorized: driver does not own this trip")
	}

	if trip.Status != models.TripStatusPaused {
		return nil, fmt.Errorf("trip must be paused to resume")
	}

	updates := map[string]interface{}{
		"status":     models.TripStatusResumed,
		"updated_at": time.Now(),
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		s.logger.Error("failed to resume trip", zap.String("trip_id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to resume trip: %w", err)
	}

	s.logger.Info("trip resumed", zap.String("trip_id", id))
	return s.repo.GetByID(ctx, id)
}

// Complete finishes a trip
func (s *TripService) Complete(ctx context.Context, id, driverID string, req *models.CompleteTripRequest) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != trip.OrgID {
			s.logger.Warn("unauthorized trip complete attempt",
				zap.String("trip_id", id),
				zap.String("trip_tenant", trip.OrgID),
				zap.String("request_tenant", orgID))
			return nil, fmt.Errorf("trip not found")
		}
	}

	if trip.DriverID == nil || *trip.DriverID != driverID {
		return nil, fmt.Errorf("unauthorized: driver does not own this trip")
	}

	if trip.Status != models.TripStatusStarted && trip.Status != models.TripStatusResumed {
		return nil, fmt.Errorf("trip must be started or resumed to complete")
	}

	now := time.Now()
	duration := now.Sub(*trip.StartTime)

	// Calculate billable distance if not provided
	actualDistance := req.TotalDistance
	if actualDistance == 0 && trip.StartOdometer != nil {
		actualDistance = req.OdometerReading - *trip.StartOdometer
	}

	updates := map[string]interface{}{
		"status":              models.TripStatusCompleted,
		"end_time":            &now,
		"end_odometer":        req.OdometerReading,
		"total_distance":      actualDistance,
		"actual_duration_min": int(duration.Minutes()),
		"fuel_used":           req.FuelUsed,
		"idling_time_minutes": req.IdlingTimeMinutes,
		"updated_at":          now,
		"version":             trip.Version,
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		s.logger.Error("failed to complete trip", zap.String("trip_id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to complete trip: %w", err)
	}

	// Record metrics: trip completed and active trips decrement
	metrics.TripsCompletedTotal.Inc()
	metrics.ActiveTrips.Dec()

	// Add POD documents
	for _, pod := range req.PODDocuments {
		pod.ID = uuid.New().String()
		pod.Timestamp = now
		pod.UploadedBy = driverID

		if err := s.repo.AddPODDocument(ctx, id, pod); err != nil {
			s.logger.Error("failed to add POD document", zap.String("trip_id", id), zap.Error(err))
			// Don't fail the completion for POD errors
		}
	}

	s.logger.Info("trip completed",
		zap.String("trip_id", id),
		zap.Float64("billable_km", actualDistance),
	)
	middleware.RecordTripEvent(middleware.MetricTripCompleted)

	// Financial Hooks & Event Emission
	s.publishEvent("trip.completed", map[string]interface{}{
		"trip_id":           id,
		"org_id":            trip.OrgID,
		"vehicle_id":        trip.VehicleID,
		"driver_id":         driverID,
		"billable_distance": actualDistance,
		"duration_minutes":  int(duration.Minutes()),
		"completed_at":      now,
	})

	return s.repo.GetByID(ctx, id)
}

// Cancel cancels a trip
func (s *TripService) Cancel(ctx context.Context, id, reason string) error {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != trip.OrgID {
			s.logger.Warn("unauthorized trip cancel attempt",
				zap.String("trip_id", id),
				zap.String("trip_tenant", trip.OrgID),
				zap.String("request_tenant", orgID))
			return fmt.Errorf("trip not found")
		}
	}

	// Spec: CREATED → CANCELLED and SCHEDULED → CANCELLED only.
	// Active trips (STARTED/PAUSED/RESUMED) must be FAILED, not CANCELLED.
	if trip.Status != models.TripStatusCreated && trip.Status != models.TripStatusScheduled {
		return fmt.Errorf("cannot cancel trip with status %s: only CREATED or SCHEDULED trips can be cancelled", trip.Status)
	}

	updates := map[string]interface{}{
		"status":              models.TripStatusCancelled,
		"cancellation_reason": reason,
		"updated_at":          time.Now(),
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		s.logger.Error("failed to cancel trip", zap.String("trip_id", id), zap.Error(err))
		return fmt.Errorf("failed to cancel trip: %w", err)
	}

	s.logger.Info("trip cancelled", zap.String("trip_id", id))
	middleware.RecordTripEvent(middleware.MetricTripCancelled)
	return nil
}

// Fail marks a trip as failed due to issues like breakdowns
func (s *TripService) Fail(ctx context.Context, id, reason string) error {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != trip.OrgID {
			s.logger.Warn("unauthorized trip fail attempt",
				zap.String("trip_id", id),
				zap.String("trip_tenant", trip.OrgID),
				zap.String("request_tenant", orgID))
			return fmt.Errorf("trip not found")
		}
	}

	// Valid Transitions: STARTED/PAUSED/RESUMED -> FAILED
	if trip.Status != models.TripStatusStarted && trip.Status != models.TripStatusPaused && trip.Status != models.TripStatusResumed {
		return fmt.Errorf("trip cannot be failed from status: %s", trip.Status)
	}

	updates := map[string]interface{}{
		"status":              models.TripStatusFailed,
		"cancellation_reason": reason, // Reuse this field for failure reason
		"updated_at":          time.Now(),
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		s.logger.Error("failed to fail trip", zap.String("trip_id", id), zap.Error(err))
		return fmt.Errorf("failed to fail trip: %w", err)
	}

	s.logger.Info("trip failed", zap.String("trip_id", id))
	middleware.RecordTripEvent(middleware.MetricTripFailed)
	return nil
}

// AddEvent adds an event to a trip
func (s *TripService) AddEvent(ctx context.Context, tripID, driverID string, req *models.AddEventRequest) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != trip.OrgID {
			s.logger.Warn("unauthorized trip add event attempt",
				zap.String("trip_id", tripID),
				zap.String("trip_tenant", trip.OrgID),
				zap.String("request_tenant", orgID))
			return fmt.Errorf("trip not found")
		}
	}

	if trip.DriverID == nil || *trip.DriverID != driverID {
		return fmt.Errorf("unauthorized: driver does not own this trip")
	}

	event := models.TripEvent{
		ID:          uuid.New().String(),
		Type:        req.Type,
		Timestamp:   time.Now(),
		Location:    req.Location,
		Description: req.Description,
		Photos:      req.Photos,
		CreatedBy:   driverID,
	}

	// Hash Chaining logic
	lastEvent, err := s.repo.GetLastEvent(ctx, tripID)
	if err != nil {
		s.logger.Warn("failed to get last event for hash chaining", zap.Error(err))
	}
	if lastEvent != nil {
		event.PrevHash = lastEvent.Hash
	}

	// Generate Hash: h(type + timestamp + prevHash + description)
	hashInput := fmt.Sprintf("%s|%d|%s|%s", event.Type, event.Timestamp.Unix(), event.PrevHash, event.Description)
	h := sha256.New()
	h.Write([]byte(hashInput))
	event.Hash = fmt.Sprintf("%x", h.Sum(nil))

	if err := s.repo.AddEvent(ctx, tripID, event); err != nil {
		s.logger.Error("failed to add trip event", zap.String("trip_id", tripID), zap.Error(err))
		return fmt.Errorf("failed to add event: %w", err)
	}

	s.logger.Info("trip event added", zap.String("trip_id", tripID), zap.String("event_type", req.Type))
	return nil
}

// SubmitPOD submits proof of delivery for a trip
func (s *TripService) SubmitPOD(ctx context.Context, tripID, driverID string, req *models.SubmitPODRequest) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership (CWE-639)
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal != nil {
		orgID, ok := orgIDVal.(string)
		if ok && orgID != trip.OrgID {
			s.logger.Warn("unauthorized trip pod submission attempt",
				zap.String("trip_id", tripID),
				zap.String("trip_tenant", trip.OrgID),
				zap.String("request_tenant", orgID))
			return fmt.Errorf("trip not found")
		}
	}

	if trip.DriverID == nil || *trip.DriverID != driverID {
		return fmt.Errorf("unauthorized: driver does not own this trip")
	}

	if trip.Status != models.TripStatusCompleted {
		return fmt.Errorf("POD can only be submitted for completed trips")
	}

	pod := models.PODDocument{
		ID:          uuid.New().String(),
		Type:        req.Type,
		URL:         req.URL,
		Timestamp:   time.Now(),
		UploadedBy:  driverID,
		Recipient:   req.Recipient,
		OTPVerified: req.OTPVerified,
		Signature:   req.Signature,
	}

	// Save to new table
	proof := &models.ProofOfDelivery{
		ID:     pod.ID,
		TripID: tripID,
		Type:   req.Type,
		URL:    req.URL,
		Metadata: map[string]interface{}{
			"recipient":    req.Recipient,
			"signature":    req.Signature,
			"otp_verified": req.OTPVerified,
		},
		UploadedBy: driverID,
		Verified:   false,
		CreatedAt:  time.Now(),
	}
	if err := s.repo.SaveProofOfDelivery(ctx, proof); err != nil {
		s.logger.Error("failed to save POD to new table", zap.Error(err))
		// Continue to save to legacy JSONB for backward compatibility
	}

	if err := s.repo.AddPODDocument(ctx, tripID, pod); err != nil {
		s.logger.Error("failed to submit POD", zap.String("trip_id", tripID), zap.Error(err))
		return fmt.Errorf("failed to submit POD: %w", err)
	}

	s.logger.Info("POD submitted", zap.String("trip_id", tripID), zap.String("pod_type", req.Type))
	s.publishEvent("trip.pod_submitted", proof)
	return nil
}

// GetActiveTripsForVehicle retrieves active trips for a vehicle
func (s *TripService) GetActiveTripsForVehicle(ctx context.Context, vehicleID string) ([]*models.Trip, error) {
	trips, err := s.repo.GetActiveTripsForVehicle(ctx, vehicleID)
	if err != nil {
		s.logger.Error("failed to get active trips for vehicle", zap.String("vehicle_id", vehicleID), zap.Error(err))
		return nil, fmt.Errorf("failed to get active trips: %w", err)
	}

	return trips, nil
}

// GetSummary retrieves trip analytics for a tenant
func (s *TripService) GetTripSummary(ctx context.Context, orgID string, startDate, endDate time.Time) (*models.TripSummary, error) {
	summary, err := s.repo.GetTripSummary(ctx, orgID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get trip summary", zap.String("org_id", orgID), zap.Error(err))
		return nil, fmt.Errorf("failed to get trip summary: %w", err)
	}

	return summary, nil
}

// GetDriverStats retrieves driver performance metrics
func (s *TripService) GetDriverStats(ctx context.Context, orgID string) ([]models.DriverPerformance, error) {
	return s.repo.GetDriverStats(ctx, orgID)
}

// GetTripStats retrieves daily trip statistics for the last 7 days
func (s *TripService) GetTripStats(ctx context.Context, orgID string) ([]models.DailyTripTrend, error) {
	return s.repo.GetTripStats(ctx, orgID)
}
func (s *TripService) GetStats(ctx context.Context) (*models.TripStats, error) {
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		s.logger.Error("failed to get trip stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	return stats, nil
}

// GetTelemetry retrieves telemetry points for a trip
func (s *TripService) GetTelemetry(ctx context.Context, tripID string) ([]models.TelemetryPoint, error) {
	// First get the trip to verify tenant ownership (CWE-639)
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		s.logger.Error("failed to get telemetry: trip not found", zap.String("trip_id", tripID), zap.Error(err))
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	// SECURITY FIX: Validate tenant ownership
	orgIDVal := ctx.Value(ctxkeys.OrgIDKey)
	if orgIDVal == nil {
		s.logger.Error("tenant context missing for telemetry access")
		return nil, fmt.Errorf("unauthorized")
	}
	orgID, ok := orgIDVal.(string)
	if !ok || orgID != trip.OrgID {
		s.logger.Warn("unauthorized telemetry access attempt",
			zap.String("trip_id", tripID),
			zap.String("trip_tenant", trip.OrgID),
			zap.Any("request_tenant", orgIDVal))
		return nil, fmt.Errorf("unauthorized")
	}

	return s.repo.GetTelemetry(ctx, tripID)
}
