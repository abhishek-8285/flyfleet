package queue

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeReportGenerate = "report:generate"
	TypeReportEmail    = "report:email"
)

// TaskPayloads
type ReportGeneratePayload struct {
	ExecutionID uuid.UUID
	ReportID    uuid.UUID
	UserID      uuid.UUID
}

type QueueClient struct {
	client *asynq.Client
}

func NewQueueClient(redisAddr string) *QueueClient {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &QueueClient{client: client}
}

func (c *QueueClient) EnqueueReportGeneration(executionID, reportID, userID uuid.UUID) error {
	payload, err := json.Marshal(ReportGeneratePayload{
		ExecutionID: executionID,
		ReportID:    reportID,
		UserID:      userID,
	})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeReportGenerate, payload)
	
	// MaxRetry: 3, Timeout: 30 minutes (reports can be slow)
	_, err = c.client.Enqueue(task, asynq.MaxRetry(3), asynq.Timeout(30*time.Minute))
	return err
}

func (c *QueueClient) Close() error {
	return c.client.Close()
}
