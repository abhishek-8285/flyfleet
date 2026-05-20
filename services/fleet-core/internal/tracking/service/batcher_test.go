package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/tracking/models"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
)

func TestTelemetryBatcher_ChronologicalSorting(t *testing.T) {
	telemRepo := repository.NewMemTelemRepo()
	logger := zap.NewNop()

	// Instantiate the batcher
	batcher := NewTelemetryBatcher(telemRepo, logger, 5, 2*time.Second)
	batcher.Start()
	defer batcher.Stop()

	tenantID := uuid.Must(uuid.NewV4())
	vehicleID := uuid.Must(uuid.NewV4())
	now := time.Now().UTC()

	// Add events in OUT-OF-ORDER chronological sequence
	e1 := models.TelemetryEvent{
		EventID:    uuid.Must(uuid.NewV4()),
		VehicleID:  vehicleID,
		TenantID:   tenantID,
		RecordedAt: now.Add(10 * time.Second),
		Lat:        12.0,
		Lng:        77.0,
	}
	e2 := models.TelemetryEvent{
		EventID:    uuid.Must(uuid.NewV4()),
		VehicleID:  vehicleID,
		TenantID:   tenantID,
		RecordedAt: now.Add(-5 * time.Second),
		Lat:        13.0,
		Lng:        78.0,
	}
	e3 := models.TelemetryEvent{
		EventID:    uuid.Must(uuid.NewV4()),
		VehicleID:  vehicleID,
		TenantID:   tenantID,
		RecordedAt: now,
		Lat:        14.0,
		Lng:        79.0,
	}

	// We pass these through BulkIngest to check sorting
	events := []models.TelemetryEvent{e1, e2, e3}
	err := telemRepo.BulkIngest(context.Background(), events)
	if err != nil {
		t.Fatalf("bulk ingest: %v", err)
	}

	points := telemRepo.Points()
	if len(points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(points))
	}

	// The repository BulkIngest sorts them chronologically using our new optimization!
	if !points[0].RecordedAt.Before(points[1].RecordedAt) {
		t.Errorf("events are not sorted chronologically: points[0] = %v, points[1] = %v", points[0].RecordedAt, points[1].RecordedAt)
	}
	if !points[1].RecordedAt.Before(points[2].RecordedAt) {
		t.Errorf("events are not sorted chronologically: points[1] = %v, points[2] = %v", points[1].RecordedAt, points[2].RecordedAt)
	}
}

func TestTelemetryBatcher_Spillover_Fallback(t *testing.T) {
	telemRepo := repository.NewMemTelemRepo()
	logger := zap.NewNop()

	mr := miniredis.RunT(t)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	// Instantiate the batcher with a tiny buffer (size 1) to trigger backpressure
	batcher := NewTelemetryBatcher(telemRepo, logger, 1, 5*time.Second)
	batcher.SetRedisClient(rdb)

	tenantID := uuid.Must(uuid.NewV4())
	vehicleID := uuid.Must(uuid.NewV4())

	e := models.TelemetryEvent{
		EventID:    uuid.Must(uuid.NewV4()),
		VehicleID:  vehicleID,
		TenantID:   tenantID,
		RecordedAt: time.Now().UTC(),
		Lat:        12.0,
		Lng:        77.0,
	}

	// Fill up the buffer channel manually so that new writes trigger blocking and then spillover
	batcher.ch <- e
	batcher.ch <- e

	// Now writing via batcher.Add will trigger channel saturation
	batcher.Add(e)

	// Wait briefly for the async spillover goroutine to run
	time.Sleep(300 * time.Millisecond)

	// Check if the event spilled over to Redis
	spilloverCount, err := rdb.LLen(context.Background(), "telemetry:spillover").Result()
	if err != nil {
		t.Fatalf("failed to query redis spillover list length: %v", err)
	}

	if spilloverCount == 0 {
		t.Error("expected event to spill over to Redis list, but list is empty")
	}

	// Verify the spilled over data can be unmarshalled back
	data, err := rdb.LPop(context.Background(), "telemetry:spillover").Bytes()
	if err != nil {
		t.Fatalf("failed to pop from redis list: %v", err)
	}

	var poppedEvent models.TelemetryEvent
	if err := json.Unmarshal(data, &poppedEvent); err != nil {
		t.Fatalf("failed to unmarshal popped event: %v", err)
	}

	if poppedEvent.EventID != e.EventID {
		t.Errorf("expected EventID %s, got %s", e.EventID, poppedEvent.EventID)
	}
}
