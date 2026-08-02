package streaming

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/flyfleet/mlplatform/internal/eventbus"
)

// Processor consumes events and runs transformations.
type Processor struct {
	bus         eventbus.Bus
	handlers    map[string][]eventbus.Handler
	middlewares []eventbus.Middleware
	wg          sync.WaitGroup
}

func NewProcessor(bus eventbus.Bus) *Processor {
	return &Processor{
		bus:      bus,
		handlers: make(map[string][]eventbus.Handler),
	}
}

func (p *Processor) Use(mw eventbus.Middleware) {
	p.middlewares = append(p.middlewares, mw)
}

// On registers a handler for a subject pattern.
func (p *Processor) On(subject, group string, handler eventbus.Handler) {
	p.handlers[subject] = append(p.handlers[subject], handler)
}

// Start launches all consumers.
func (p *Processor) Start(ctx context.Context) error {
	for subject, handlers := range p.handlers {
		ch, err := p.bus.Subscribe(ctx, subject, "streaming-"+subject)
		if err != nil {
			return err
		}

		for _, h := range handlers {
			handler := h
			// Wrap with middleware
			for i := len(p.middlewares) - 1; i >= 0; i-- {
				handler = p.middlewares[i](handler)
			}

			p.wg.Add(1)
			go func() {
				defer p.wg.Done()
				for evt := range ch {
					if err := handler(ctx, evt); err != nil {
						log.Error().Err(err).
							Str("subject", subject).
							Str("event_id", evt.ID).
							Msg("handler error")
					}
				}
			}()
		}
	}
	return nil
}

func (p *Processor) Wait() {
	p.wg.Wait()
}
