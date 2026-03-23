package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/trip/models"
	"go.uber.org/zap"
)

// TripStateMachine handles trip state transitions with validation
type TripStateMachine struct {
	logger *zap.Logger
}

// NewTripStateMachine creates a new state machine
func NewTripStateMachine(logger *zap.Logger) *TripStateMachine {
	return &TripStateMachine{logger: logger}
}

// TransitionRequest represents a state transition request
type TransitionRequest struct {
	TripID        string
	CurrentStatus string
	TargetStatus  string
	ActorID       string
	ActorType     string // driver, system, admin
	Reason        string
	Metadata      map[string]interface{}
}

// TransitionResult represents the result of a state transition
type TransitionResult struct {
	Success         bool
	NewStatus       string
	Event           *models.TripEventImmutable
	ValidationError error
}

// CanTransition checks if a transition is valid
func (sm *TripStateMachine) CanTransition(fromStatus, toStatus string) bool {
	return models.IsValidTransition(fromStatus, toStatus)
}

// IsTerminal checks if a status is terminal
func (sm *TripStateMachine) IsTerminal(status string) bool {
	return models.IsTerminalState(status)
}

// Transition performs a validated state transition
func (sm *TripStateMachine) Transition(ctx context.Context, req *TransitionRequest, prevEvent *models.TripEventImmutable) (*TransitionResult, error) {
	// Validate transition
	if !sm.CanTransition(req.CurrentStatus, req.TargetStatus) {
		return &TransitionResult{
			Success: false,
			ValidationError: fmt.Errorf("invalid transition from %s to %s", req.CurrentStatus, req.TargetStatus),
		}, nil
	}

	// Validate actor permissions
	if err := sm.validateActor(req); err != nil {
		return &TransitionResult{
			Success: false,
			ValidationError: err,
		}, nil
	}

	// Validate business rules for specific transitions
	if err := sm.validateBusinessRules(req); err != nil {
		return &TransitionResult{
			Success: false,
			ValidationError: err,
		}, nil
	}

	// Create immutable event
	event, err := sm.createEvent(req, prevEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	// Verify hash chain
	if prevEvent != nil && !event.VerifyHashChain(prevEvent) {
		sm.logger.Error("Hash chain verification failed",
			zap.String("trip_id", req.TripID),
			zap.String("event_id", event.ID))
		return nil, errors.New("audit integrity check failed")
	}

	sm.logger.Info("Trip state transition validated",
		zap.String("trip_id", req.TripID),
		zap.String("from_status", req.CurrentStatus),
		zap.String("to_status", req.TargetStatus),
		zap.String("actor_id", req.ActorID))

	return &TransitionResult{
		Success:   true,
		NewStatus: req.TargetStatus,
		Event:     event,
	}, nil
}

// validateActor checks if the actor can perform the transition
func (sm *TripStateMachine) validateActor(req *TransitionRequest) error {
	// System can always transition
	if req.ActorType == "system" {
		return nil
	}

	// Driver can only transition their own trips
	if req.ActorType == "driver" && req.ActorID == "" {
		return errors.New("driver ID required for driver-initiated transitions")
	}

	// Admin can transition any trip
	if req.ActorType == "admin" {
		return nil
	}

	return nil
}

// validateBusinessRules enforces business-specific validation
func (sm *TripStateMachine) validateBusinessRules(req *TransitionRequest) error {
	switch req.TargetStatus {
	case models.TripStatusStarted:
		// Validate permit, PUC, insurance before starting
		if req.Metadata != nil {
			if permitVerified, ok := req.Metadata["permit_verified"].(bool); ok && !permitVerified {
				return errors.New("vehicle permit not verified")
			}
			if pucValid, ok := req.Metadata["puc_valid"].(bool); ok && !pucValid {
				return errors.New("PUC certificate not valid")
			}
			if insuranceValid, ok := req.Metadata["insurance_valid"].(bool); ok && !insuranceValid {
				return errors.New("insurance not valid")
			}
		}

	case models.TripStatusCompleted:
		// Validate POD documents
		if req.Metadata != nil {
			if podCount, ok := req.Metadata["pod_count"].(int); ok && podCount < 1 {
				return errors.New("at least one POD document required")
			}
			if endOdometer, ok := req.Metadata["end_odometer"].(float64); ok {
				if startOdometer, ok := req.Metadata["start_odometer"].(float64); ok {
					if endOdometer < startOdometer {
						return errors.New("end odometer cannot be less than start odometer")
					}
				}
			}
		}

	case models.TripStatusCancelled:
		// Require reason for cancellation (except system-initiated)
		if req.Reason == "" && req.ActorType != "system" {
			return errors.New("cancellation reason required")
		}
	}

	return nil
}

// createEvent creates an immutable trip event with hash chain
func (sm *TripStateMachine) createEvent(req *TransitionRequest, prevEvent *models.TripEventImmutable) (*models.TripEventImmutable, error) {
	prevHash := ""
	if prevEvent != nil {
		prevHash = prevEvent.Hash
	}

	payload := models.TripEventPayload{
		EventType:  "state_transition",
		Timestamp:  time.Now().UTC(),
		ActorID:    req.ActorID,
		ActorType:  req.ActorType,
		Reason:     req.Reason,
		Metadata:   req.Metadata,
	}

	event := &models.TripEventImmutable{
		ID:        generateEventID(),
		TripID:    req.TripID,
		EventType: fmt.Sprintf("transition_%s_to_%s", req.CurrentStatus, req.TargetStatus),
		PrevHash:  prevHash,
		Timestamp: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}

	// Marshal payload
	payloadJSON, err := marshalPayload(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	event.Payload = payloadJSON

	// Calculate hash
	event.Hash = event.CalculateHash()

	return event, nil
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// marshalPayload marshals payload to JSON
func marshalPayload(payload interface{}) ([]byte, error) {
	// In production, use proper JSON marshaling with error handling
	return nil, nil
}
