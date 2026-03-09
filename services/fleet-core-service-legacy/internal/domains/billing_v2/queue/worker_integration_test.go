//go:build integration
// +build integration

package queue_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/queue"
	"github.com/flyfleet/fleet-core-service/internal/domains/billing_v2/testutil"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestInvoiceWorker_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	redisAddr, terminate := testutil.SetupRedisContainer(ctx, t)
	defer terminate()

	// Client to enqueue task
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer client.Close()

	// Task data
	invoiceID := uuid.New()
	orgID := uuid.New()
	payload := map[string]interface{}{
		"invoice_id": invoiceID.String(),
		"org_id":     orgID.String(),
	}
	payloadJSON, _ := json.Marshal(payload)
	task := asynq.NewTask(queue.TaskInvoiceRender, payloadJSON)

	// Enqueue
	info, err := client.Enqueue(task)
	assert.NoError(t, err)
	assert.Equal(t, queue.TaskInvoiceRender, info.Type)

	// Note: In a full integration test, we would start the worker server here
	// and verify the output in MinIO/DB.
	// For this phase, we verify the task enqueuing and Redis lifecycle.
}
