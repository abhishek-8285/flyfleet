package workers

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

// Tests only exercise initialization and a single no-op cycle since
// GetStalledIncidents requires a DB; DB interactions are covered in integration tests.

func TestNewEmergencyRecoveryWorker_Initializes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewEmergencyRecoveryWorker(nil, logger)

	if worker == nil {
		t.Fatal("Expected non-nil recovery worker")
	}
	if worker.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestRecoveryWorker_ContextCancellation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewEmergencyRecoveryWorker(nil, logger)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		cancel()
	}()

	// Should return immediately on cancel without panicking
	worker.Run(ctx)
}
