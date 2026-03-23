package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type DiagnosticsCache struct {
	client *redis.Client
}

func NewDiagnosticsCache(redisAddr string) *DiagnosticsCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &DiagnosticsCache{client: rdb}
}

func (c *DiagnosticsCache) SetDeviceStatus(ctx context.Context, deviceID string, status string) error {
	return c.client.Set(ctx, fmt.Sprintf("device:status:%s", deviceID), status, 1*time.Hour).Err()
}

func (c *DiagnosticsCache) GetDeviceStatus(ctx context.Context, deviceID string) (string, error) {
	return c.client.Get(ctx, fmt.Sprintf("device:status:%s", deviceID)).Result()
}

func (c *DiagnosticsCache) SetDeviceHealth(ctx context.Context, deviceID string, health string) error {
	return c.client.Set(ctx, fmt.Sprintf("device:health:%s", deviceID), health, 1*time.Hour).Err()
}

func (c *DiagnosticsCache) GetDeviceHealth(ctx context.Context, deviceID string) (string, error) {
	return c.client.Get(ctx, fmt.Sprintf("device:health:%s", deviceID)).Result()
}

func (c *DiagnosticsCache) SetFirmwareMetadata(ctx context.Context, fwID string, metadata string) error {
	return c.client.Set(ctx, fmt.Sprintf("firmware:meta:%s", fwID), metadata, 24*time.Hour).Err()
}

func (c *DiagnosticsCache) GetFirmwareMetadata(ctx context.Context, fwID string) (string, error) {
	return c.client.Get(ctx, fmt.Sprintf("firmware:meta:%s", fwID)).Result()
}

func (c *DiagnosticsCache) SetRolloutProgress(ctx context.Context, rolloutID string, progress string) error {
	return c.client.Set(ctx, fmt.Sprintf("rollout:progress:%s", rolloutID), progress, 24*time.Hour).Err()
}

func (c *DiagnosticsCache) GetRolloutProgress(ctx context.Context, rolloutID string) (string, error) {
	return c.client.Get(ctx, fmt.Sprintf("rollout:progress:%s", rolloutID)).Result()
}
