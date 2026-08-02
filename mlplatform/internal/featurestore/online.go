package featurestore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOnlineStore implements low-latency online feature serving.
type RedisOnlineStore struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewRedisOnlineStore(addr, password string, db int) *RedisOnlineStore {
	return &RedisOnlineStore{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
			PoolSize: 50,
		}),
		prefix: "fs:",
		ttl:    24 * time.Hour,
	}
}

func (s *RedisOnlineStore) key(entityType, entityID, featureName string) string {
	return fmt.Sprintf("%s%s:%s:%s", s.prefix, entityType, entityID, featureName)
}

func (s *RedisOnlineStore) GetOnlineFeatures(ctx context.Context, entityType, entityID string, featureNames []string) (*FeatureVector, error) {
	fv := &FeatureVector{
		EntityID:   entityID,
		EntityType: entityType,
		Features:   make([]FeatureValue, 0, len(featureNames)),
	}

	keys := make([]string, len(featureNames))
	for i, name := range featureNames {
		keys[i] = s.key(entityType, entityID, name)
	}

	vals, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget: %w", err)
	}

	for _, val := range vals {
		if val == nil {
			continue
		}
		var fval FeatureValue
		if err := json.Unmarshal([]byte(val.(string)), &fval); err != nil {
			continue
		}
		fv.Features = append(fv.Features, fval)
	}

	return fv, nil
}

func (s *RedisOnlineStore) PutOnlineFeatures(ctx context.Context, fv *FeatureVector) error {
	pipe := s.client.Pipeline()

	for _, f := range fv.Features {
		data, err := json.Marshal(f)
		if err != nil {
			continue
		}
		key := s.key(fv.EntityType, fv.EntityID, f.Name)
		pipe.Set(ctx, key, data, s.ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Offline and PointInTimeJoin are delegated to the ClickHouse warehouse.
func (s *RedisOnlineStore) GetOfflineFeatures(ctx context.Context, entityType string, entityIDs []string, featureNames []string, start, end time.Time) ([]*FeatureVector, error) {
	// Delegated to warehouse.ClickHouseStore
	return nil, fmt.Errorf("use warehouse for offline queries")
}

func (s *RedisOnlineStore) WriteOfflineFeatures(ctx context.Context, fvs []*FeatureVector) error {
	return fmt.Errorf("use warehouse for offline writes")
}

func (s *RedisOnlineStore) PointInTimeJoin(ctx context.Context, entityTable string, timestamps []time.Time, featureNames []string) ([]*FeatureVector, error) {
	return nil, fmt.Errorf("use warehouse for PIT joins")
}

func (s *RedisOnlineStore) RegisterFeature(ctx context.Context, def *FeatureDefinition) error {
	// Store definitions in Redis hash for quick lookup
	data, _ := json.Marshal(def)
	return s.client.HSet(ctx, "fs:registry:"+def.EntityType, def.Name, data).Err()
}

func (s *RedisOnlineStore) ListFeatures(ctx context.Context, entityType string) ([]*FeatureDefinition, error) {
	vals, err := s.client.HGetAll(ctx, "fs:registry:"+entityType).Result()
	if err != nil {
		return nil, err
	}
	defs := make([]*FeatureDefinition, 0, len(vals))
	for _, v := range vals {
		var d FeatureDefinition
		json.Unmarshal([]byte(v), &d)
		defs = append(defs, &d)
	}
	return defs, nil
}
