//go:build integration
// +build integration

package integrations

import (
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap/zaptest"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	logger := zaptest.NewLogger(t)
	integrationID := uuid.New().String()
	
	registry := NewCircuitBreakerRegistry(logger)
	cb := registry.For(integrationID)

	// In the real system, it's 10 failures. Let's trip it forcefully.
	for i := 0; i < 11; i++ {
		cb.RecordFailure()
	}

	// Should be OPEN now
	if cb.State() != StateOpen {
		t.Errorf("expected StateOpen, got %s", cb.State().String())
	}
	if cb.Allow() {
		t.Errorf("expected cb.Allow() == false when OPEN")
	}

	// Wait until HALF_OPEN is skipped for simple tests (it's 2 min cooldown, we just test state machinery)
	// We can forcibly reset state to test RecordSuccess closing it from HALF_OPEN
	cb.mu.Lock()
	cb.state = StateHalfOpen
	cb.mu.Unlock()

	// This execution should occur (HALF_OPEN probe) and succeed
	if !cb.Allow() {
		t.Errorf("expected cb.Allow() == true when HALF_OPEN for 1 probe")
	}
	
	cb.RecordSuccess() // should lock it to CLOSED

	if cb.State() != StateClosed {
		t.Errorf("expected circuit to be CLOSED after RecordSuccess, got %v", cb.State())
	}
}
