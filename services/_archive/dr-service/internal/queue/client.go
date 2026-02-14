package queue

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeBackup  = "dr:backup"
	TypeRestore = "dr:restore"
	TypeDRTest  = "dr:test"
)

// TaskPayloads
type BackupPayload struct {
	BackupID uuid.UUID
}

type RestorePayload struct {
	RestoreID uuid.UUID
}

type DRTestPayload struct {
	TestID uuid.UUID
}

type QueueClient struct {
	client *asynq.Client
}

func NewQueueClient(redisAddr string) *QueueClient {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &QueueClient{client: client}
}

func (c *QueueClient) EnqueueBackup(backupID uuid.UUID) error {
	payload, err := json.Marshal(BackupPayload{BackupID: backupID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeBackup, payload)
	_, err = c.client.Enqueue(task, asynq.MaxRetry(5), asynq.Timeout(1*time.Hour))
	return err
}

func (c *QueueClient) EnqueueRestore(restoreID uuid.UUID) error {
	payload, err := json.Marshal(RestorePayload{RestoreID: restoreID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeRestore, payload)
	_, err = c.client.Enqueue(task, asynq.MaxRetry(3), asynq.Timeout(2*time.Hour))
	return err
}

func (c *QueueClient) EnqueueDRTest(testID uuid.UUID) error {
	payload, err := json.Marshal(DRTestPayload{TestID: testID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeDRTest, payload)
	_, err = c.client.Enqueue(task, asynq.MaxRetry(3), asynq.Timeout(3*time.Hour))
	return err
}

func (c *QueueClient) Close() error {
	return c.client.Close()
}
