package client

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string { return e.Err.Error() }
func (e *RetryableError) Unwrap() error { return e.Err }

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func (p *RetryPolicy) Do(ctx context.Context, fn func() error) error {
	var err error
	for attempt := 0; attempt < p.MaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if _, ok := err.(*RetryableError); !ok {
			return err
		}

		if attempt == p.MaxAttempts-1 {
			break
		}

		delay := time.Duration(float64(p.BaseDelay) * math.Pow(2, float64(attempt)))
		if delay > p.MaxDelay {
			delay = p.MaxDelay
		}
		jitter := time.Duration(rand.Int63n(int64(delay) / 2))

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay + jitter):
		}
	}
	return err
}