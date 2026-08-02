package collector

import (
	"context"
	"time"

	"github.com/flyfleet/mlplatform/internal/eventbus"
	"github.com/flyfleet/mlplatform/internal/metrics"
	"github.com/google/uuid"
)

// Collector ingests raw events from producers and publishes to the bus.
type Collector struct {
	bus         eventbus.Bus
	metrics     *metrics.Metrics
	middlewares []eventbus.Middleware
}

func New(bus eventbus.Bus, m *metrics.Metrics) *Collector {
	return &Collector{bus: bus, metrics: m}
}

func (c *Collector) Use(mw eventbus.Middleware) {
	c.middlewares = append(c.middlewares, mw)
}

// Ingest accepts a raw event, enriches it, and publishes.
func (c *Collector) Ingest(ctx context.Context, eventType, source, subject string, data []byte) error {
	evt := &eventbus.Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Source:    source,
		Subject:   subject,
		Timestamp: time.Now().UTC(),
		Data:      data,
		Metadata:  make(map[string]string),
	}

	// Apply enrichment middleware chain.
	handler := func(ctx context.Context, e *eventbus.Event) error {
		if err := c.bus.Publish(ctx, e.Type, e); err != nil {
			return err
		}
		c.metrics.EventsIngested.WithLabelValues(e.Type, e.Source).Inc()
		return nil
	}
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}

	return handler(ctx, evt)
}
