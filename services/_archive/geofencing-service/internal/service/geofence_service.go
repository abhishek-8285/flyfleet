package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/flyfleet/geofencing-service/internal/models"
	"github.com/flyfleet/geofencing-service/internal/repository"
)

type GeofenceService struct {
	repo   *repository.GeofenceRepository
	nats   *nats.Conn
	logger *zap.Logger
}

func NewGeofenceService(repo *repository.GeofenceRepository, nats *nats.Conn, logger *zap.Logger) *GeofenceService {
	return &GeofenceService{
		repo:   repo,
		nats:   nats,
		logger: logger,
	}
}

// Create creates a new geofence
func (s *GeofenceService) Create(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, req *models.CreateGeofenceRequest) (*models.Geofence, error) {
	// Validate the request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Prepare the geofence model
	geofence := &models.Geofence{
		ID:           uuid.New(),
		OrgID:        orgID,
		Name:         req.Name,
		Type:         req.Type,
		Center:       req.Center,
		RadiusMeters: req.RadiusMeters,
		Coordinates:  req.Coordinates,
		RoutePolyline: req.RoutePolyline,
		BufferMeters: req.BufferMeters,
		EntryAlert:   req.EntryAlert,
		ExitAlert:    req.ExitAlert,
		SpeedLimit:   req.SpeedLimit,
		Vehicles:     req.Vehicles,
		Active:       true,
		CreatedBy:    userID,
		Metadata:     req.Metadata,
		ActiveHours:  req.ActiveHours,
	}

	// Save to database
	if err := s.repo.Create(ctx, geofence); err != nil {
		s.logger.Error("Failed to create geofence", zap.Error(err))
		return nil, fmt.Errorf("failed to create geofence: %w", err)
	}

	s.logger.Info("Geofence created successfully", zap.String("id", geofence.ID.String()))

	// Publish event to NATS
	event := map[string]interface{}{
		"event_type":   "geofence.created",
		"geofence_id":  geofence.ID.String(),
		"org_id":       orgID.String(),
		"name":         req.Name,
		"type":         req.Type,
		"created_by":   userID.String(),
		"timestamp":    geofence.CreatedAt.Unix(),
	}

	if err := s.nats.Publish("fleet.geofence.events", []byte(fmt.Sprintf("%v", event))); err != nil {
		s.logger.Error("Failed to publish geofence created event", zap.Error(err))
	}

	return geofence, nil
}

// GetByID retrieves a geofence by ID
func (s *GeofenceService) GetByID(ctx context.Context, id uuid.UUID) (*models.Geofence, error) {
	geofence, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get geofence", zap.String("id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to get geofence: %w", err)
	}

	return geofence, nil
}

// List retrieves geofences with pagination and filtering
func (s *GeofenceService) List(ctx context.Context, orgID uuid.UUID, req *models.ListGeofencesRequest) (*models.GeofenceListResponse, error) {
	geofences, total, err := s.repo.List(ctx, orgID, req)
	if err != nil {
		s.logger.Error("Failed to list geofences", zap.String("org_id", orgID.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to list geofences: %w", err)
	}

	response := &models.GeofenceListResponse{
		Geofences: geofences,
		Total:     total,
		Page:      req.Page,
		Limit:     req.Limit,
		HasMore:   int64((req.Page * req.Limit)) < total,
	}

	return response, nil
}

// Update updates an existing geofence
func (s *GeofenceService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateGeofenceRequest) (*models.Geofence, error) {
	// Validate the update request
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Check if geofence exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("geofence not found: %w", err)
	}

	// Update the geofence
	if err := s.repo.Update(ctx, id, req); err != nil {
		s.logger.Error("Failed to update geofence", zap.String("id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to update geofence: %w", err)
	}

	// Fetch updated geofence
	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated geofence: %w", err)
	}

	s.logger.Info("Geofence updated successfully", zap.String("id", id.String()))

	// Publish event to NATS
	event := map[string]interface{}{
		"event_type":  "geofence.updated",
		"geofence_id": id.String(),
		"updated_by":  updated.UpdatedAt,
		"timestamp":   updated.UpdatedAt.Unix(),
	}

	if err := s.nats.Publish("fleet.geofence.events", []byte(fmt.Sprintf("%v", event))); err != nil {
		s.logger.Error("Failed to publish geofence updated event", zap.Error(err))
	}

	return updated, nil
}

// Delete marks a geofence as inactive
func (s *GeofenceService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete geofence", zap.String("id", id.String()), zap.Error(err))
		return fmt.Errorf("failed to delete geofence: %w", err)
	}

	s.logger.Info("Geofence deleted successfully", zap.String("id", id.String()))

	// Publish event to NATS
	event := map[string]interface{}{
		"event_type":  "geofence.deleted",
		"geofence_id": id.String(),
		"timestamp":   time.Now().Unix(),
	}

	if err := s.nats.Publish("fleet.geofence.events", []byte(fmt.Sprintf("%v", event))); err != nil {
		s.logger.Error("Failed to publish geofence deleted event", zap.Error(err))
	}

	return nil
}

// GetEvents retrieves geofence events with filtering
func (s *GeofenceService) GetEvents(ctx context.Context, filter *models.GeofenceEventFilter) ([]models.GeofenceEvent, int64, error) {
	events, total, err := s.repo.GetEvents(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get geofence events", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get geofence events: %w", err)
	}

	return events, total, nil
}

// validateCreateRequest validates the create geofence request
func (s *GeofenceService) validateCreateRequest(req *models.CreateGeofenceRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}

	if req.Type == "" {
		return errors.New("type is required")
	}

	// Validate based on type
	switch req.Type {
	case "circle":
		if req.Center == nil {
			return errors.New("center is required for circle geofence")
		}
		if req.RadiusMeters == nil {
			return errors.New("radius_meters is required for circle geofence")
		}
		if *req.RadiusMeters < 10 || *req.RadiusMeters > 50000 {
			return errors.New("radius_meters must be between 10 and 50000")
		}
	case "polygon":
		if len(req.Coordinates) < 3 {
			return errors.New("polygon requires at least 3 coordinates")
		}
	case "route_corridor":
		if req.RoutePolyline == nil {
			return errors.New("route_polyline is required for route corridor geofence")
		}
		if req.BufferMeters == nil {
			return errors.New("buffer_meters is required for route corridor geofence")
		}
		if *req.BufferMeters < 10 || *req.BufferMeters > 1000 {
			return errors.New("buffer_meters must be between 10 and 1000")
		}
	default:
		return fmt.Errorf("invalid geofence type: %s", req.Type)
	}

	// Validate speed limit if provided
	if req.SpeedLimit != nil {
		if *req.SpeedLimit < 10 || *req.SpeedLimit > 200 {
			return errors.New("speed_limit must be between 10 and 200")
		}
	}

	return nil
}

// validateUpdateRequest validates the update geofence request
func (s *GeofenceService) validateUpdateRequest(req *models.UpdateGeofenceRequest) error {
	if req.Name != nil && len(*req.Name) > 100 {
		return errors.New("name must be less than 100 characters")
	}

	if req.Type != nil {
		validTypes := []string{"circle", "polygon", "route_corridor"}
		isValid := false
		for _, validType := range validTypes {
			if *req.Type == validType {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid geofence type: %s", *req.Type)
		}
	}

	if req.RadiusMeters != nil {
		if *req.RadiusMeters < 10 || *req.RadiusMeters > 50000 {
			return errors.New("radius_meters must be between 10 and 50000")
		}
	}

	if req.BufferMeters != nil {
		if *req.BufferMeters < 10 || *req.BufferMeters > 1000 {
			return errors.New("buffer_meters must be between 10 and 1000")
		}
	}

	if req.SpeedLimit != nil {
		if *req.SpeedLimit < 10 || *req.SpeedLimit > 200 {
			return errors.New("speed_limit must be between 10 and 200")
		}
	}

	return nil
}