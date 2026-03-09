package integrations

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// ─── Circuit Breaker ──────────────────────────────────────────────────────────
//
// Per-integration circuit breaker.
//
//	CLOSED: normal operation
//	  → OPEN if 10 failures in 60 seconds
//
//	OPEN: all requests rejected immediately
//	  → HALF_OPEN after 2 minutes
//
//	HALF_OPEN: next request is a probe
//	  → CLOSED if probe succeeds
//	  → OPEN if probe fails

type CircuitState int

const (
	StateClosed   CircuitState = iota
	StateOpen     CircuitState = iota
	StateHalfOpen CircuitState = iota
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

const (
	failureThreshold  = 10              // failures to trip breaker
	failureWindow     = 60 * time.Second // window for counting failures
	cooldownDuration  = 2 * time.Minute  // OPEN → HALF_OPEN cooldown
)

type CircuitBreaker struct {
	mu           sync.RWMutex
	state        CircuitState
	failures     []time.Time // sliding window of failure timestamps
	openSince    time.Time
	logger       *zap.Logger
	integrationID string
}

func NewCircuitBreaker(integrationID string, logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		state:         StateClosed,
		integrationID: integrationID,
		logger:        logger,
	}
}

// Allow returns true if a request may proceed.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.openSince) > cooldownDuration {
			cb.state = StateHalfOpen
			cb.logger.Info("CircuitBreaker: OPEN → HALF_OPEN",
				zap.String("integration_id", cb.integrationID))
			return true
		}
		return false
	case StateHalfOpen:
		return true // allow one probe
	}
	return false
}

// RecordSuccess transitions HALF_OPEN → CLOSED or keeps CLOSED.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failures = nil
		cb.logger.Info("CircuitBreaker: HALF_OPEN → CLOSED",
			zap.String("integration_id", cb.integrationID))
	}
}

// RecordFailure records a failure and may trip the breaker.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	// Prune failures outside the window
	cutoff := now.Add(-failureWindow)
	pruned := cb.failures[:0]
	for _, t := range cb.failures {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}
	cb.failures = append(pruned, now)

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.openSince = now
		cb.logger.Warn("CircuitBreaker: HALF_OPEN → OPEN (probe failed)",
			zap.String("integration_id", cb.integrationID))
		return
	}

	if len(cb.failures) >= failureThreshold && cb.state == StateClosed {
		cb.state = StateOpen
		cb.openSince = now
		cb.logger.Warn("CircuitBreaker: CLOSED → OPEN",
			zap.String("integration_id", cb.integrationID),
			zap.Int("failures_in_window", len(cb.failures)))
	}
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// ─── Circuit Breaker Registry ─────────────────────────────────────────────────

type CircuitBreakerRegistry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
	logger   *zap.Logger
}

func NewCircuitBreakerRegistry(logger *zap.Logger) *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

func (r *CircuitBreakerRegistry) For(integrationID string) *CircuitBreaker {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cb, ok := r.breakers[integrationID]; ok {
		return cb
	}
	cb := NewCircuitBreaker(integrationID, r.logger)
	r.breakers[integrationID] = cb
	return cb
}

func (r *CircuitBreakerRegistry) OpenCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, cb := range r.breakers {
		if cb.State() == StateOpen {
			n++
		}
	}
	return n
}
