package fallback

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/flyfleet/maintenance-service/internal/models"
)

func TestFallbackService(t *testing.T) {
	fallbackSvc := NewFallbackService()

	ctx := context.Background()
	scheduleID := uuid.New().String()
	logID := uuid.New().String()

	// Test creating and getting a schedule
	schedule := &models.MaintenanceSchedule{
		ID:          scheduleID,
		Name:        "Test Schedule",
		Type:        models.MaintenanceTypePreventive,
		CreatedAt:   time.Now(),
	}

	err := fallbackSvc.FallbackCreateSchedule(ctx, schedule)
	assert.NoError(t, err)

	retrievedSchedule, err := fallbackSvc.FallbackGetSchedule(ctx, scheduleID)
	assert.NoError(t, err)
	assert.Equal(t, schedule, retrievedSchedule)
	assert.True(t, fallbackSvc.HasSchedule(scheduleID))

	// Test creating and getting a log
	log := &models.MaintenanceLog{
		ID:          logID,
		Name:        "Test Log",
		Type:        models.MaintenanceTypePreventive,
		CreatedAt:   time.Now(),
	}

	err = fallbackSvc.FallbackCreateLog(ctx, log)
	assert.NoError(t, err)

	retrievedLog, err := fallbackSvc.FallbackGetLog(ctx, logID)
	assert.NoError(t, err)
	assert.Equal(t, log, retrievedLog)
	assert.True(t, fallbackSvc.HasLog(logID))

	// Test not found cases
	_, err = fallbackSvc.FallbackGetSchedule(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	_, err = fallbackSvc.FallbackGetLog(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test GetFallbackSchedules and GetFallbackLogs
	schedules := fallbackSvc.GetFallbackSchedules()
	assert.Len(t, schedules, 1)
	assert.Equal(t, schedule, schedules[0])

	logs := fallbackSvc.GetFallbackLogs()
	assert.Len(t, logs, 1)
	assert.Equal(t, log, logs[0])
}

func TestFallbackServiceCleanup(t *testing.T) {
	fallbackSvc := NewFallbackService()

	ctx := context.Background()
	scheduleID := uuid.New().String()
	logID := uuid.New().String()

	// Create items with old timestamps
	oldTime := time.Now().Add(-2 * time.Hour)
	schedule := &models.MaintenanceSchedule{
		ID:        scheduleID,
		Name:      "Old Schedule",
		CreatedAt: oldTime,
	}

	log := &models.MaintenanceLog{
		ID:        logID,
		Name:      "Old Log",
		CreatedAt: oldTime,
	}

	err := fallbackSvc.FallbackCreateSchedule(ctx, schedule)
	assert.NoError(t, err)
	err = fallbackSvc.FallbackCreateLog(ctx, log)
	assert.NoError(t, err)

	// Verify they exist initially
	assert.True(t, fallbackSvc.HasSchedule(scheduleID))
	assert.True(t, fallbackSvc.HasLog(logID))

	// Cleanup old entries (anything older than 30 minutes)
	fallbackSvc.CleanupOldEntries(30 * time.Minute)

	// Verify they were cleaned up
	assert.False(t, fallbackSvc.HasSchedule(scheduleID))
	assert.False(t, fallbackSvc.HasLog(logID))
}

func TestRetryOperation(t *testing.T) {
	fallbackSvc := NewFallbackService()

	// Test successful operation on first try
	attemptCount := 0
	successOp := func() error {
		attemptCount++
		return nil
	}

	err := fallbackSvc.RetryOperation(successOp, 3, 10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, 1, attemptCount)

	// Test operation that succeeds on second try
	attemptCount = 0
	partialFailureOp := func() error {
		attemptCount++
		if attemptCount == 1 {
			return assert.AnError // Use a standard error for testing
		}
		return nil
	}

	err = fallbackSvc.RetryOperation(partialFailureOp, 3, 10*time.Millisecond)
	assert.NoError(t, err)
	assert.Equal(t, 2, attemptCount)

	// Test operation that fails all retries
	attemptCount = 0
	failureOp := func() error {
		attemptCount++
		return assert.AnError
	}

	err = fallbackSvc.RetryOperation(failureOp, 3, 10*time.Millisecond)
	assert.Error(t, err)
	assert.Equal(t, 3, attemptCount)
}