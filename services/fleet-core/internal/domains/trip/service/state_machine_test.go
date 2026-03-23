//go:build integration
// +build integration

package service

import (
	"context"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/trip/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTripStateMachine_CanTransition(t *testing.T) {
	sm := NewTripStateMachine(zap.NewNop())

	tests := []struct {
		name     string
		from     string
		to       string
		expected bool
	}{
		{"Created to Scheduled", models.TripStatusCreated, models.TripStatusScheduled, true},
		{"Created to Cancelled", models.TripStatusCreated, models.TripStatusCancelled, true},
		{"Scheduled to Started", models.TripStatusScheduled, models.TripStatusStarted, true},
		{"Started to Paused", models.TripStatusStarted, models.TripStatusPaused, true},
		{"Paused to Resumed", models.TripStatusPaused, models.TripStatusResumed, true},
		{"Started to Completed", models.TripStatusStarted, models.TripStatusCompleted, true},
		{"Completed to Started", models.TripStatusCompleted, models.TripStatusStarted, false},
		{"Cancelled to Started", models.TripStatusCancelled, models.TripStatusStarted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.CanTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTripStateMachine_IsTerminal(t *testing.T) {
	sm := NewTripStateMachine(zap.NewNop())

	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"Completed is terminal", models.TripStatusCompleted, true},
		{"Cancelled is terminal", models.TripStatusCancelled, true},
		{"Failed is terminal", models.TripStatusFailed, true},
		{"Started is not terminal", models.TripStatusStarted, false},
		{"Paused is not terminal", models.TripStatusPaused, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.IsTerminal(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTripStateMachine_Transition_Success(t *testing.T) {
	sm := NewTripStateMachine(zap.NewNop())
	ctx := context.Background()

	req := &TransitionRequest{
		TripID:        "trip-123",
		CurrentStatus: models.TripStatusCreated,
		TargetStatus:  models.TripStatusScheduled,
		ActorID:       "driver-123",
		ActorType:     "driver",
		Reason:        "",
		Metadata:      nil,
	}

	result, err := sm.Transition(ctx, req, nil)

	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, models.TripStatusScheduled, result.NewStatus)
	assert.NotNil(t, result.Event)
	assert.NotEmpty(t, result.Event.Hash)
	assert.Empty(t, result.Event.PrevHash) // First event
}

func TestTripStateMachine_Transition_InvalidTransition(t *testing.T) {
	sm := NewTripStateMachine(zap.NewNop())
	ctx := context.Background()

	req := &TransitionRequest{
		TripID:        "trip-123",
		CurrentStatus: models.TripStatusCompleted,
		TargetStatus:  models.TripStatusStarted,
		ActorID:       "driver-123",
		ActorType:     "driver",
	}

	result, err := sm.Transition(ctx, req, nil)

	assert.NoError(t, err)
	assert.False(t, result.Success)
	assert.Error(t, result.ValidationError)
	assert.Contains(t, result.ValidationError.Error(), "invalid transition")
}

func TestTripStateMachine_Transition_WithHashChain(t *testing.T) {
	sm := NewTripStateMachine(zap.NewNop())
	ctx := context.Background()

	// First transition
	req1 := &TransitionRequest{
		TripID:        "trip-123",
		CurrentStatus: models.TripStatusCreated,
		TargetStatus:  models.TripStatusScheduled,
		ActorID:       "driver-123",
		ActorType:     "driver",
	}

	result1, err := sm.Transition(ctx, req1, nil)
	assert.NoError(t, err)
	assert.True(t, result1.Success)

	// Second transition with hash chain
	req2 := &TransitionRequest{
		TripID:        "trip-123",
		CurrentStatus: models.TripStatusScheduled,
		TargetStatus:  models.TripStatusStarted,
		ActorID:       "driver-123",
		ActorType:     "driver",
	}

	result2, err := sm.Transition(ctx, req2, result1.Event)
	assert.NoError(t, err)
	assert.True(t, result2.Success)
	assert.Equal(t, result1.Event.Hash, result2.Event.PrevHash)
}

func TestTripStateMachine_ValidateBusinessRules_Start(t *testing.T) {
	sm := NewTripStateMachine(zap.NewNop())
	ctx := context.Background()

	// Test with unverified permit
	req := &TransitionRequest{
		TripID:        "trip-123",
		CurrentStatus: models.TripStatusScheduled,
		TargetStatus:  models.TripStatusStarted,
		ActorID:       "driver-123",
		ActorType:     "driver",
		Metadata: map[string]interface{}{
			"permit_verified": false,
		},
	}

	result, err := sm.Transition(ctx, req, nil)

	assert.NoError(t, err)
	assert.False(t, result.Success)
	assert.Error(t, result.ValidationError)
	assert.Contains(t, result.ValidationError.Error(), "permit not verified")
}

func TestTripStateMachine_ValidateBusinessRules_Complete(t *testing.T) {
	sm := NewTripStateMachine(zap.NewNop())
	ctx := context.Background()

	// Test without POD documents
	req := &TransitionRequest{
		TripID:        "trip-123",
		CurrentStatus: models.TripStatusStarted,
		TargetStatus:  models.TripStatusCompleted,
		ActorID:       "driver-123",
		ActorType:     "driver",
		Metadata: map[string]interface{}{
			"pod_count": 0,
		},
	}

	result, err := sm.Transition(ctx, req, nil)

	assert.NoError(t, err)
	assert.False(t, result.Success)
	assert.Error(t, result.ValidationError)
	assert.Contains(t, result.ValidationError.Error(), "POD document required")
}

func TestTripStateMachine_ValidateBusinessRules_Cancel(t *testing.T) {
	sm := NewTripStateMachine(zap.NewNop())
	ctx := context.Background()

	// Test cancellation without reason
	req := &TransitionRequest{
		TripID:        "trip-123",
		CurrentStatus: models.TripStatusScheduled,
		TargetStatus:  models.TripStatusCancelled,
		ActorID:       "driver-123",
		ActorType:     "driver",
		Reason:        "",
	}

	result, err := sm.Transition(ctx, req, nil)

	assert.NoError(t, err)
	assert.False(t, result.Success)
	assert.Error(t, result.ValidationError)
	assert.Contains(t, result.ValidationError.Error(), "cancellation reason required")
}

func TestTripEventImmutable_CalculateHash(t *testing.T) {
	event := &models.TripEventImmutable{
		ID:        "evt-123",
		TripID:    "trip-123",
		EventType: "test_event",
		PrevHash:  "",
		Timestamp: time.Now(),
	}

	hash1 := event.CalculateHash()
	hash2 := event.CalculateHash()

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2) // Hash should be deterministic
}

func TestTripEventImmutable_VerifyHashChain(t *testing.T) {
	// Create first event
	event1 := &models.TripEventImmutable{
		ID:        "evt-1",
		TripID:    "trip-123",
		EventType: "event_1",
		PrevHash:  "",
		Timestamp: time.Now(),
	}
	event1.Hash = event1.CalculateHash()

	// Create second event with proper chain
	event2 := &models.TripEventImmutable{
		ID:        "evt-2",
		TripID:    "trip-123",
		EventType: "event_2",
		PrevHash:  event1.Hash,
		Timestamp: time.Now().Add(1 * time.Minute),
	}
	event2.Hash = event2.CalculateHash()

	// Verify chain
	assert.True(t, event2.VerifyHashChain(event1))

	// Create broken chain
	event3 := &models.TripEventImmutable{
		ID:        "evt-3",
		TripID:    "trip-123",
		EventType: "event_3",
		PrevHash:  "broken-hash",
		Timestamp: time.Now().Add(2 * time.Minute),
	}
	event3.Hash = event3.CalculateHash()

	assert.False(t, event3.VerifyHashChain(event2))
}

func TestVerifyEventChain(t *testing.T) {
	// Create valid chain
	event1 := &models.TripEventImmutable{
		ID:        "evt-1",
		TripID:    "trip-123",
		EventType: "event_1",
		PrevHash:  "",
		Timestamp: time.Now(),
	}
	event1.Hash = event1.CalculateHash()

	event2 := &models.TripEventImmutable{
		ID:        "evt-2",
		TripID:    "trip-123",
		EventType: "event_2",
		PrevHash:  event1.Hash,
		Timestamp: time.Now().Add(1 * time.Minute),
	}
	event2.Hash = event2.CalculateHash()

	events := []*models.TripEventImmutable{event1, event2}

	assert.True(t, models.VerifyEventChain(events))

	// Test empty chain
	assert.True(t, models.VerifyEventChain([]*models.TripEventImmutable{}))
}

func TestCalculateEventHash(t *testing.T) {
	payload := map[string]interface{}{
		"key": "value",
	}

	hash1 := models.CalculateEventHash("trip-123", "test_event", payload, "", time.Now())
	hash2 := models.CalculateEventHash("trip-123", "test_event", payload, "", time.Now())

	// Hashes will be different due to timestamp
	assert.NotEmpty(t, hash1)
	assert.NotEmpty(t, hash2)
}
