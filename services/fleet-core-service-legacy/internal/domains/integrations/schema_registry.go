package integrations

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// ─── Schema Registry ────────────────────────────────────────────────────────
//
// SchemaRegistry ensures events conform to expected backward-compatible schemas
// before they are allowed onto the event bus or out to external webhooks.
// (Gap 8: Event Schema Version Registry)

type SchemaRegistry interface {
	// RegisterSchema defines a new event type/version and its required fields.
	RegisterSchema(eventType, version string, requiredFields []string) error
	
	// ValidatePayload checks if a JSON payload fulfills the registered schema requirements.
	ValidatePayload(eventType, version string, payload []byte) error
}

type LocalSchemaRegistry struct {
	mu      sync.RWMutex
	schemas map[string][]string // key: "eventType:version", val: required fields (dot notation)
	logger  *zap.Logger
}

func NewLocalSchemaRegistry(logger *zap.Logger) *LocalSchemaRegistry {
	r := &LocalSchemaRegistry{
		schemas: make(map[string][]string),
		logger:  logger,
	}
	r.registerBaselines()
	return r
}

func (r *LocalSchemaRegistry) RegisterSchema(eventType, version string, requiredFields []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	key := fmt.Sprintf("%s:%s", eventType, version)
	r.schemas[key] = requiredFields
	r.logger.Info("SchemaRegistry: registered schema", 
		zap.String("event_type", eventType), 
		zap.String("version", version))
	return nil
}

func (r *LocalSchemaRegistry) ValidatePayload(eventType, version string, payload []byte) error {
	r.mu.RLock()
	key := fmt.Sprintf("%s:%s", eventType, version)
	required, ok := r.schemas[key]
	r.mu.RUnlock()

	if !ok {
		// Strict mode: if schema is unknown, reject the event to prevent pollution.
		return fmt.Errorf("schema validation failed: unknown schema %s", key)
	}

	if len(required) == 0 {
		return nil // No specific fields required
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return fmt.Errorf("schema validation failed: invalid JSON: %w", err)
	}

	var missing []string
	for _, reqField := range required {
		if !hasNestedField(parsed, strings.Split(reqField, ".")) {
			missing = append(missing, reqField)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("schema validation failed: missing required fields %v", missing)
	}

	return nil
}

func hasNestedField(m map[string]interface{}, parts []string) bool {
	if len(parts) == 0 {
		return true
	}
	val, ok := m[parts[0]]
	if !ok {
		return false
	}
	if len(parts) == 1 {
		return true
	}
	nested, ok := val.(map[string]interface{})
	if !ok {
		return false
	}
	return hasNestedField(nested, parts[1:])
}

// registerBaselines pre-populates the registry with the core FlyFleet event schemas.
func (r *LocalSchemaRegistry) registerBaselines() {
	_ = r.RegisterSchema("trip.completed", "1.0", []string{"trip_id", "driver.name", "trip.distance"})
	_ = r.RegisterSchema("trip.started", "1.0", []string{"trip_id", "driver.name", "trip.start_time"})
	_ = r.RegisterSchema("vehicle.telemetry", "1.0", []string{"vehicle.plate", "location.lat", "location.lng"})
}
