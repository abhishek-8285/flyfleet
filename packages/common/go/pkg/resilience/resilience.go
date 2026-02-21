package resilience

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sony/gobreaker"
)

// Policy defines the resilience strategy (breaker + retry)
type Policy struct {
	cb    *gobreaker.CircuitBreaker
	retry backoff.BackOff
}

// Config defines the settings for the resilience policy
type Config struct {
	Name             string
	MaxRetries       uint64
	Interval         time.Duration
	MaxInterval      time.Duration
	Timeout          time.Duration
	ErrorThreshold   uint32
	MinRequestVolume uint32
}

// NewPolicy creates a new resilience policy with circuit breaker and exponential backoff
func NewPolicy(cfg Config) *Policy {
	// Circuit Breaker settings
	cbSettings := gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: 5,
		Interval:    time.Minute,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.Requests >= cfg.MinRequestVolume &&
				float64(counts.TotalFailures)/float64(counts.Requests) >= 0.1 // 10% threshold as per user request
		},
	}

	// Exponential Backoff settings
	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = cfg.Interval
	eb.MaxInterval = cfg.MaxInterval
	eb.MaxElapsedTime = time.Minute // Stop retrying after 1 minute

	// Enforce hard cap on retries
	retryCap := backoff.WithMaxRetries(eb, cfg.MaxRetries)

	return &Policy{
		cb:    gobreaker.NewCircuitBreaker(cbSettings),
		retry: retryCap,
	}
}

// Execute wraps a function with circuit breaker and retry logic
func (p *Policy) Execute(ctx context.Context, operation func() error) error {
	return backoff.Retry(func() error {
		// Run via circuit breaker
		_, err := p.cb.Execute(func() (interface{}, error) {
			return nil, operation()
		})

		// If circuit is open, don't even bother retrying immediately
		if errors.Is(err, gobreaker.ErrOpenState) {
			return backoff.Permanent(err)
		}

		return err
	}, backoff.WithContext(p.retry, ctx))
}

// DefaultConfig returns global defaults as requested: 2s-5s timeout, 3 retries
func DefaultConfig(name string) Config {
	return Config{
		Name:             name,
		MaxRetries:       3,
		Interval:         500 * time.Millisecond,
		MaxInterval:      5 * time.Second,
		Timeout:          5 * time.Second,
		ErrorThreshold:   10, // 10% handled in ReadyToTrip
		MinRequestVolume: 5,
	}
}
