package queue

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeAnalyticsJobProcess      = "analytics:job_process"
	TypeAnalyticsCacheInvalidate = "analytics:cache_invalidate"
)

type AnalyticsJobPayload struct {
	JobID          string    `json:"job_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID         uuid.UUID `json:"user_id"`
}

// CacheInvalidatePayload represents the payload for a cache invalidation task
type CacheInvalidatePayload struct {
	OrganizationID uuid.UUID `json:"organization_id"`
	Pattern        string    `json:"pattern"`
	ResourceType   string    `json:"resource_type"` // "vehicle", "driver", "fleet"
	ResourceID     string    `json:"resource_id"`
}

// AnalyticsQueueClient provides a client to enqueue analytics tasks
type AnalyticsQueueClient struct {
	client *asynq.Client
}

// NewAnalyticsQueueClient creates a new AnalyticsQueueClient
func NewAnalyticsQueueClient(redisAddr string) *AnalyticsQueueClient {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &AnalyticsQueueClient{client: client}
}

func (c *AnalyticsQueueClient) EnqueueAnalyticsJob(jobID string, orgID, userID uuid.UUID) error {
	payload, err := json.Marshal(AnalyticsJobPayload{
		JobID:          jobID,
		OrganizationID: orgID,
		UserID:         userID,
	})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeAnalyticsJobProcess, payload)

	// MaxRetry: 3, Timeout: 1 hour (analytics can be very slow)
	_, err = c.client.Enqueue(task, asynq.MaxRetry(3), asynq.Timeout(1*time.Hour))
	return err
}

// EnqueueCacheInvalidate enqueues a task to invalidate cache patterns
func (c *AnalyticsQueueClient) EnqueueCacheInvalidate(orgID uuid.UUID, pattern, resourceType, resourceID string) error {
	payload, err := json.Marshal(CacheInvalidatePayload{
		OrganizationID: orgID,
		Pattern:        pattern,
		ResourceType:   resourceType,
		ResourceID:     resourceID,
	})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeAnalyticsCacheInvalidate, payload)

	// Low priority task, fast execution
	_, err = c.client.Enqueue(task, asynq.MaxRetry(3), asynq.Timeout(5*time.Minute))
	return err
}

// Close closes the underlying asynq client
func (c *AnalyticsQueueClient) Close() error {
	return c.client.Close()
}
