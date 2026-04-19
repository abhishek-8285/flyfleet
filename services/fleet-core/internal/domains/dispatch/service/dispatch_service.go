package service

import (
	"context"

	"github.com/google/uuid"
)

type DispatchService interface {
	Assign(ctx context.Context, orgID uuid.UUID, req AssignRequest) (*AssignResult, error)
	Redispatch(ctx context.Context, orgID uuid.UUID, req RedispatchRequest) (*RedispatchResult, error)
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

type EventBus interface {
	Publish(subject string, data interface{}) error
}

type dispatchService struct {
	eventBus EventBus
}

func NewDispatchService(eb EventBus) DispatchService {
	return &dispatchService{eventBus: eb}
}

func (s *dispatchService) Assign(ctx context.Context, orgID uuid.UUID, req AssignRequest) (*AssignResult, error) {
	tripID := uuid.New().String()
	
	if s.eventBus != nil {
		s.eventBus.Publish("trip.assigned", map[string]interface{}{
			"trip_id":     tripID,
			"shipment_id": req.ShipmentID,
			"tenant_id":      orgID,
		})
	}

	return &AssignResult{
		TripID:         tripID,
		Status:         "SCHEDULED",
		ShipmentStatus: "ASSIGNED",
	}, nil
}

func (s *dispatchService) Redispatch(ctx context.Context, orgID uuid.UUID, req RedispatchRequest) (*RedispatchResult, error) {
	return &RedispatchResult{
		NewTripID: uuid.New().String(),
		Status:    "REDISPATCHED",
	}, nil
}
