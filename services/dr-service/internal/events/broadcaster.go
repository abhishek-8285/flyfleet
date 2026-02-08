package events

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const ChannelName = "dr:events"

type EventType string

const (
	EventBackupStatus  EventType = "BACKUP_STATUS"
	EventRestoreStatus EventType = "RESTORE_STATUS"
	EventLog           EventType = "LOG"
)

type Event struct {
	Type      EventType   `json:"type"`
	ResourceID uuid.UUID  `json:"resource_id"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

type Broadcaster struct {
	client *redis.Client
}

func NewBroadcaster(redisAddr string) *Broadcaster {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &Broadcaster{client: rdb}
}

func (b *Broadcaster) Publish(ctx context.Context, evt Event) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, ChannelName, data).Err()
}

func (b *Broadcaster) Subscribe(ctx context.Context) *redis.PubSub {
	return b.client.Subscribe(ctx, ChannelName)
}

func (b *Broadcaster) Close() error {
	return b.client.Close()
}
