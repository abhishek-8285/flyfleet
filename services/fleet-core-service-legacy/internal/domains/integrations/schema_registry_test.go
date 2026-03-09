//go:build integration
// +build integration

package integrations

import (
	"testing"
	"go.uber.org/zap/zaptest"
)

func TestSchemaRegistry_Validation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := NewLocalSchemaRegistry(logger)

	// Test 1: Baseline valid payload
	validPayload := []byte(`{"trip_id":"123", "driver":{"name":"alice"}, "trip":{"distance":10.5}}`)
	err := registry.ValidatePayload("trip.completed", "1.0", validPayload)
	if err != nil {
		t.Errorf("expected valid payload to pass, got error: %v", err)
	}

	// Test 2: Missing required field
	invalidPayload := []byte(`{"trip_id":"123", "driver":{"name":"alice"}}`) // missing trip.distance
	err = registry.ValidatePayload("trip.completed", "1.0", invalidPayload)
	if err == nil {
		t.Errorf("expected validation to fail for missing field")
	}

	// Test 3: Unknown schema version (strict mode rejection)
	err = registry.ValidatePayload("trip.completed", "2.0", validPayload)
	if err == nil {
		t.Errorf("expected validation to fail for unknown schema version")
	}

	// Test 4: Register new schema evolution (forward compatibility)
	err = registry.RegisterSchema("trip.completed", "1.1", []string{"trip_id", "driver.name", "trip.distance", "trip.duration"})
	if err != nil {
		t.Fatalf("failed to register evolved schema: %v", err)
	}

	// Old payload fails new schema
	err = registry.ValidatePayload("trip.completed", "1.1", validPayload)
	if err == nil {
		t.Errorf("expected old payload to fail evolved schema")
	}

	// New payload passes new schema
	newPayload := []byte(`{"trip_id":"123", "driver":{"name":"alice"}, "trip":{"distance":10.5, "duration":120}}`)
	err = registry.ValidatePayload("trip.completed", "1.1", newPayload)
	if err != nil {
		t.Errorf("expected new payload to pass evolved schema: %v", err)
	}
}
