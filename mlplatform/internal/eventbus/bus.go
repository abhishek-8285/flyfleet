package eventbus

import (
	"context"
	"time"
)

// Event is the universal envelope for all platform events.
type Event struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Source    string            `json:"source"`
	Timestamp time.Time         `json:"timestamp"`
	Subject   string            `json:"subject"`  // e.g., "driver:1234"
	Data      []byte            `json:"data"`
	Metadata  map[string]string `json:"metadata"`
	TraceID   string            `json:"trace_id"`
}

// Bus is the core pub/sub abstraction.
type Bus interface {
	// Publish sends an event to a subject/topic.
	Publish(ctx context.Context, subject string, event *Event) error

	// Subscribe returns a channel of events for the given subject.
	// Supports wildcards: "orders.*", "driver.>"
	Subscribe(ctx context.Context, subject string, group string) (<-chan *Event, error)

	// Close shuts down the connection gracefully.
	Close() error
}

// Middleware wraps event processing.
type Middleware func(next Handler) Handler

type Handler func(ctx context.Context, event *Event) error
