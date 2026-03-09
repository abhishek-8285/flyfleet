//go:build integration
// +build integration

package integrations

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestDLQManager_BackpressureActivates(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repo := &OperationalMockRepo{}
	
	dlq := NewDLQManager(repo, nil, logger)
	ctx := context.Background()
    
	// 1. Manually set flag via repo
	_ = repo.SetSystemFlag(ctx, "backpressure_paused", "true")
    
	if !dlq.IsBackpressurePaused(ctx) {
		t.Errorf("expected backpressure to be paused")
	}

	// 2. Clear flag
	_ = repo.SetSystemFlag(ctx, "backpressure_paused", "false")

	if dlq.IsBackpressurePaused(ctx) {
		t.Errorf("expected backpressure to be resumed")
	}
}
