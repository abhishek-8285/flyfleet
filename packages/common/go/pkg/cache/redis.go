package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

type RedisCacheInterface interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	FlushAll(ctx context.Context) error
	Close() error
}

func NewRedisCache(addr, password string, db int, logger *zap.Logger) (RedisCacheInterface, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("Connected to Redis cache", zap.String("addr", addr))

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

func (rc *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		rc.logger.Error("Failed to get cache value", zap.String("key", key), zap.Error(err))
		return nil, err
	}

	return []byte(val), nil
}

func (rc *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	err := rc.client.Set(ctx, key, string(value), expiration).Err()
	if err != nil {
		rc.logger.Error("Failed to set cache value", 
			zap.String("key", key), 
			zap.Error(err))
		return err
	}

	return nil
}

func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	err := rc.client.Del(ctx, key).Err()
	if err != nil {
		rc.logger.Error("Failed to delete cache key", zap.String("key", key), zap.Error(err))
		return err
	}

	return nil
}

func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		rc.logger.Error("Failed to check cache key existence", zap.String("key", key), zap.Error(err))
		return false, err
	}

	return exists > 0, nil
}

func (rc *RedisCache) FlushAll(ctx context.Context) error {
	err := rc.client.FlushAll(ctx).Err()
	if err != nil {
		rc.logger.Error("Failed to flush cache", zap.Error(err))
		return err
	}

	rc.logger.Info("Cache flushed successfully")
	return nil
}

func (rc *RedisCache) Close() error {
	if err := rc.client.Close(); err != nil {
		rc.logger.Error("Failed to close redis connection", zap.Error(err))
		return err
	}

	rc.logger.Info("Redis connection closed")
	return nil
}