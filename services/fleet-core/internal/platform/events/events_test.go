package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEvent_Serialization(t *testing.T) {
	tenantID := uuid.New()
	eventID := uuid.New()
	timestamp := time.Now().Unix()

	event := &Event{
		EventID:        eventID,
		EventType:      "fleet.vehicle.created",
		TenantID:  tenantID,
		Source:         "vehicle-service",
		Payload:        map[string]interface{}{"id": "123"},
		Timestamp:      timestamp,
	}

	data, err := json.Marshal(event)
	assert.NoError(t, err)

	var decoded Event
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, eventID, decoded.EventID)
	assert.Equal(t, tenantID, decoded.TenantID)
	assert.Equal(t, "fleet.vehicle.created", decoded.EventType)
	assert.Equal(t, "123", decoded.Payload["id"])
}
