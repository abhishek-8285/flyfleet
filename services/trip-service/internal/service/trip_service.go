package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/trip-service/internal/models"
	"github.com/flyfleet/trip-service/internal/repository"
)

type TripService struct {
	repo          *repository.TripRepository
	telemetryRepo *repository.TelemetryRepository
	logger        *zap.Logger
}

func NewTripService(repo *repository.TripRepository, telemetryRepo *repository.TelemetryRepository, logger *zap.Logger) *TripService {
	return &TripService{
		repo:          repo,
		telemetryRepo: telemetryRepo,
		logger:        logger,
	}
}

// Create creates a new trip
func (s *TripService) Create(ctx context.Context, orgID string, req *models.CreateTripRequest) (*models.Trip, error) {
	// Validate that vehicle and driver exist and are available
	// This would typically involve calling vehicle-service and driver-service
	// For now, we'll assume validation happens elsewhere or in middleware

	trip := &models.Trip{
		ID:                   uuid.New().String(),
		OrgID:                orgID,
		VehicleID:            req.VehicleID,
		DriverID:             &req.DriverID,
		Status:               "planned",
		Title:                req.Title,
		Description:          req.Description,
		PlannedRoute:         req.PlannedRoute,
		Waypoints:            req.Waypoints,
		EstimatedDistanceKm:  req.EstimatedDistanceKm,
		EstimatedDurationMin: req.EstimatedDurationMin,
		CargoDetails:         req.CargoDetails,
		ScheduledStart:       req.ScheduledStart,
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

	s.logger.Info("trip created", zap.String("trip_id", trip.ID))
	return trip, nil
}

// Get retrieves a trip by ID
func (s *TripService) Get(ctx context.Context, id string) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get trip", zap.String("trip_id", id), zap.Error(err))
		return nil, fmt.Errorf("trip not found: %w", err)
	}

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
	if req.EstimatedDistanceKm != nil {
		updates["estimated_distance_km"] = *req.EstimatedDistanceKm
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

	// Only allow updates to planned trips
	if existingTrip.Status != "planned" {
		return nil, fmt.Errorf("can only update planned trips")
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

	// Only allow deletion of planned trips
	if trip.Status != "planned" {
		return fmt.Errorf("can only delete planned trips")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete trip", zap.String("trip_id", id), zap.Error(err))
		return fmt.Errorf("failed to delete trip: %w", err)
	}

	s.logger.Info("trip deleted", zap.String("trip_id", id))
	return nil
}

// Start begins a trip
func (s *TripService) Start(ctx context.Context, tripID, driverID string, req *models.StartTripRequest) (*models.Trip, error) {
	// Get existing trip
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	// Validate authorization
	if trip.DriverID == nil || *trip.DriverID != driverID {
		return nil, fmt.Errorf("unauthorized: driver does not own this trip")
	}

	// Validate trip status
	if trip.Status != "planned" && trip.Status != "paused" {
		return nil, fmt.Errorf("trip cannot be started from status: %s", trip.Status)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       "in_progress",
		"started_at":   &now,
		"start_location": trip.PlannedRoute[0].Location, // First waypoint as start location
		"updated_at":   now,
	}

	if err := s.repo.Update(ctx, tripID, updates); err != nil {
		s.logger.Error("failed to start trip", zap.String("trip_id", tripID), zap.Error(err))
		return nil, fmt.Errorf("failed to start trip: %w", err)
	}

	s.logger.Info("trip started", zap.String("trip_id", tripID))

	// Return updated trip
	return s.repo.GetByID(ctx, tripID)
}

// Pause temporarily stops a trip
func (s *TripService) Pause(ctx context.Context, tripID, driverID string) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	if trip.DriverID == nil || *trip.DriverID != driverID {
		return nil, fmt.Errorf("unauthorized: driver does not own this trip")
	}

	if trip.Status != "in_progress" {
		return nil, fmt.Errorf("trip must be in progress to pause")
	}

	updates := map[string]interface{}{
		"status":     "paused",
		"updated_at": time.Now(),
	}

	if err := s.repo.Update(ctx, tripID, updates); err != nil {
		s.logger.Error("failed to pause trip", zap.String("trip_id", tripID), zap.Error(err))
		return nil, fmt.Errorf("failed to pause trip: %w", err)
	}

	s.logger.Info("trip paused", zap.String("trip_id", tripID))
	return s.repo.GetByID(ctx, tripID)
}

// Resume continues a paused trip
func (s *TripService) Resume(ctx context.Context, tripID, driverID string) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	if trip.DriverID == nil || *trip.DriverID != driverID {
		return nil, fmt.Errorf("unauthorized: driver does not own this trip")
	}

	if trip.Status != "paused" {
		return nil, fmt.Errorf("trip must be paused to resume")
	}

	updates := map[string]interface{}{
		"status":     "in_progress",
		"updated_at": time.Now(),
	}

	if err := s.repo.Update(ctx, tripID, updates); err != nil {
		s.logger.Error("failed to resume trip", zap.String("trip_id", tripID), zap.Error(err))
		return nil, fmt.Errorf("failed to resume trip: %w", err)
	}

	s.logger.Info("trip resumed", zap.String("trip_id", tripID))
	return s.repo.GetByID(ctx, tripID)
}

// Complete finishes a trip
func (s *TripService) Complete(ctx context.Context, tripID, driverID string, req *models.CompleteTripRequest) (*models.Trip, error) {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	if trip.DriverID == nil || *trip.DriverID != driverID {
		return nil, fmt.Errorf("unauthorized: driver does not own this trip")
	}

	if trip.Status != "in_progress" {
		return nil, fmt.Errorf("trip must be in progress to complete")
	}

	now := time.Now()
	duration := now.Sub(*trip.StartedAt)

	updates := map[string]interface{}{
		"status":              "completed",
		"ended_at":            &now,
		"actual_distance_km":  req.ActualDistanceKm,
		"actual_duration_min": int(duration.Minutes()),
		"end_location":        trip.PlannedRoute[len(trip.PlannedRoute)-1].Location, // Last waypoint as end location
		"fuel_consumed_liters": req.FuelConsumedLiters,
		"idling_time_minutes":  req.IdlingTimeMinutes,
		"updated_at":          now,
	}

	if err := s.repo.Update(ctx, tripID, updates); err != nil {
		s.logger.Error("failed to complete trip", zap.String("trip_id", tripID), zap.Error(err))
		return nil, fmt.Errorf("failed to complete trip: %w", err)
	}

	// Add POD documents
	for _, pod := range req.PODDocuments {
		pod.ID = uuid.New().String()
		pod.Timestamp = now
		pod.UploadedBy = driverID
		
		if err := s.repo.AddPODDocument(ctx, tripID, pod); err != nil {
			s.logger.Error("failed to add POD document", zap.String("trip_id", tripID), zap.Error(err))
			// Don't fail the completion for POD errors
		}
	}

	s.logger.Info("trip completed", zap.String("trip_id", tripID))
	return s.repo.GetByID(ctx, tripID)
}

// Cancel cancels a trip
func (s *TripService) Cancel(ctx context.Context, tripID, reason string) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("trip not found: %w", err)
	}

	// Allow cancellation of planned, paused, or in-progress trips
	if trip.Status == "completed" || trip.Status == "cancelled" || trip.Status == "deleted" {
		return fmt.Errorf("cannot cancel trip with status: %s", trip.Status)
	}

	updates := map[string]interface{}{
		"status":              "cancelled",
		"cancellation_reason": reason,
		"updated_at":          time.Now(),
	}

	if err := s.repo.Update(ctx, tripID, updates); err != nil {
		s.logger.Error("failed to cancel trip", zap.String("trip_id", tripID), zap.Error(err))
		return fmt.Errorf("failed to cancel trip: %w", err)
	}

	s.logger.Info("trip cancelled", zap.String("trip_id", tripID))
	return nil
}

// AddEvent adds an event to a trip
func (s *TripService) AddEvent(ctx context.Context, tripID, driverID string, req *models.AddEventRequest) error {
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		return fmt.Errorf("trip not found: %w", err)
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

	if trip.DriverID == nil || *trip.DriverID != driverID {
		return fmt.Errorf("unauthorized: driver does not own this trip")
	}

	if trip.Status != "completed" {
		return fmt.Errorf("POD can only be submitted for completed trips")
	}

	pod := models.PODDocument{
		ID:         uuid.New().String(),
		Type:       req.Type,
		URL:        req.URL,
		Timestamp:  time.Now(),
		UploadedBy: driverID,
		Recipient:  req.Recipient,
		OTPVerified: req.OTPVerified,
		Signature:  req.Signature,
	}

	if err := s.repo.AddPODDocument(ctx, tripID, pod); err != nil {
		s.logger.Error("failed to submit POD", zap.String("trip_id", tripID), zap.Error(err))
		return fmt.Errorf("failed to submit POD: %w", err)
	}

	s.logger.Info("POD submitted", zap.String("trip_id", tripID), zap.String("pod_type", req.Type))
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

// GetTripSummary retrieves trip analytics for an organization
func (s *TripService) GetTripSummary(ctx context.Context, orgID string, startDate, endDate time.Time) (*models.TripSummary, error) {
	summary, err := s.repo.GetTripSummary(ctx, orgID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get trip summary", zap.String("org_id", orgID), zap.Error(err))
		return nil, fmt.Errorf("failed to get trip summary: %w", err)
	}

	return summary, nil
}

// GetDriverStats retrieves driver performance metrics
func (s *TripService) GetDriverStats(ctx context.Context, orgID string) ([]map[string]interface{}, error) {
	// Mock implementation for now
	// This would typically involve complex aggregation on the database
	stats := []map[string]interface{}{
		{
			"driver_id":    "550e8400-e29b-41d4-a716-446655440001",
			"name":         "Rajesh Kumar",
			"total_distance": 1250.5,
			"safety_score":   98,
			"trips_completed": 45,
			"on_time_rate":   95.5,
		},
		{
			"driver_id":    "550e8400-e29b-41d4-a716-446655440003",
			"name":         "Vikram Singh",
			"total_distance": 980.2,
			"safety_score":   92,
			"trips_completed": 38,
			"on_time_rate":   88.0,
		},
		{
			"driver_id":    "550e8400-e29b-41d4-a716-446655440004",
			"name":         "Amit Patel",
			"total_distance": 850.0,
			"safety_score":   96,
			"trips_completed": 32,
			"on_time_rate":   99.0,
		},
	}
	return stats, nil
}

// GetTripStats retrieves daily trip statistics for the last 7 days
func (s *TripService) GetTripStats(ctx context.Context, orgID string) ([]map[string]interface{}, error) {
	// Mock implementation for 7-day trend
	stats := make([]map[string]interface{}, 7)
	now := time.Now()
	
	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, -i)
		stats[6-i] = map[string]interface{}{
			"date":           date.Format("2006-01-02"),
			"total_trips":    15 + (i % 3) * 5, // Mock variance
			"total_distance": 250.0 + float64(i*10),
			"active_time_hours": 8.5 + float64(i%2),
		}
	}
	
	return stats, nil
}
func (s *TripService) GetStats(ctx context.Context) (*models.TripStats, error) {
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		s.logger.Error("failed to get trip stats", zap.Error(err))
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	return stats, nil
}

func (s *TripService) GetTripTelemetry(ctx context.Context, tripID, orgID string, filter models.TelemetryFilter) (*models.TelemetryResponse, error) {
	// Fetch trip first to validate existence and get metadata
	trip, err := s.repo.GetByID(ctx, tripID)
	if err != nil {
		s.logger.Error("failed to get trip for telemetry", zap.String("trip_id", tripID), zap.Error(err))
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	// Validate organization ownership
	if trip.OrgID != orgID {
		return nil, fmt.Errorf("unauthorized: trip does not belong to organization")
	}

	// Fetch telemetry data
	points, err := s.telemetryRepo.GetTelemetryByTripID(ctx, tripID, orgID, filter)
	if err != nil {
		s.logger.Error("failed to fetch telemetry", zap.String("trip_id", tripID), zap.Error(err))
		return nil, fmt.Errorf("failed to fetch telemetry: %w", err)
	}

	// Prepare response
	aggregation := filter.Aggregation
	if aggregation == "" {
		aggregation = "raw"
	}

	resp := &models.TelemetryResponse{
		Data: points,
		Meta: models.TelemetryMeta{
			TripID:      trip.ID,
			VehicleID:   trip.VehicleID,
			StartTime:   trip.CreatedAt, // Or started_at if available
			EndTime:     trip.EndedAt,
			TotalPoints: len(points),
			Aggregation: aggregation,
		},
	}

	if trip.StartedAt != nil {
		resp.Meta.StartTime = *trip.StartedAt
	}

	return resp, nil
}
