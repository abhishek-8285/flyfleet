package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type Event struct {
	EventID        uuid.UUID              `json:"event_id"`
	EventType      string                 `json:"event_type"`
	OrganizationID uuid.UUID              `json:"organization_id"`
	Source         string                 `json:"source"`
	Payload        map[string]interface{} `json:"payload"`
	Timestamp      int64                  `json:"timestamp"`
}

type EventBus interface {
	Publish(ctx context_dummy, subject string, event *Event) error
	Subscribe(subject string, handler func(event *Event)) (*nats.Subscription, error)
}

// context_dummy is just a placeholder to keep the interface Go-ish if needed,
// though NATS Publish doesn't strictly require it.
type context_dummy interface{}

type natsEventBus struct {
	nc *nats.Conn
}

func NewEventBus(nc *nats.Conn) EventBus {
	return &natsEventBus{nc: nc}
}

func (b *natsEventBus) Publish(ctx context_dummy, subject string, event *Event) error {
	if event.EventID == uuid.Nil {
		event.EventID = uuid.New()
	}
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return b.nc.Publish(subject, data)
}

func (b *natsEventBus) Subscribe(subject string, handler func(event *Event)) (*nats.Subscription, error) {
	return b.nc.Subscribe(subject, func(msg *nats.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			fmt.Printf("Error unmarshaling event: %v\n", err)
			return
		}
		handler(&event)
	})
}
