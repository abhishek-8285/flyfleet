package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/tracking/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	liveLocKeyPrefix = "vehicle:%s:live"
	liveLocTTL       = 5 * time.Minute
)

// LiveCacheRepository defines the Redis-backed live location contract.
type LiveCacheRepository interface {
	SetLiveLocation(ctx context.Context, loc *models.LiveLocation) error
	GetLiveLocation(ctx context.Context, vehicleID uuid.UUID) (*models.LiveLocation, error)
}

// RedisLiveCacheRepository stores live vehicle locations as Redis Hashes.
type RedisLiveCacheRepository struct {
	client *redis.Client
}

// NewRedisLiveCacheRepository creates a new live cache repository.
// Pass a *redis.Client (e.g. obtained from go-redis/v9).
func NewRedisLiveCacheRepository(client *redis.Client) *RedisLiveCacheRepository {
	return &RedisLiveCacheRepository{client: client}
}

func liveKey(vehicleID uuid.UUID) string {
	return fmt.Sprintf(liveLocKeyPrefix, vehicleID.String())
}

// SetLiveLocation writes a vehicle's current location to Redis.
// The key expires after liveLocTTL of inactivity.
func (r *RedisLiveCacheRepository) SetLiveLocation(ctx context.Context, loc *models.LiveLocation) error {
	key := liveKey(loc.VehicleID)
	fields := map[string]interface{}{
		"lat":       strconv.FormatFloat(loc.Lat, 'f', 7, 64),
		"lng":       strconv.FormatFloat(loc.Lng, 'f', 7, 64),
		"speed":     strconv.FormatFloat(loc.Speed, 'f', 2, 64),
		"heading":   strconv.Itoa(loc.Heading),
		"ignition":  boolStr(loc.Ignition),
		"last_seen": loc.LastSeen.UTC().Format(time.RFC3339),
	}
	if err := r.client.HSet(ctx, key, fields).Err(); err != nil {
		return err
	}
	return r.client.Expire(ctx, key, liveLocTTL).Err()
}

// GetLiveLocation reads a vehicle's cached live location from Redis.
// Returns nil, nil if the key does not exist (vehicle offline / expired).
func (r *RedisLiveCacheRepository) GetLiveLocation(ctx context.Context, vehicleID uuid.UUID) (*models.LiveLocation, error) {
	key := liveKey(vehicleID)
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	loc := &models.LiveLocation{VehicleID: vehicleID}
	loc.Lat, _ = strconv.ParseFloat(data["lat"], 64)
	loc.Lng, _ = strconv.ParseFloat(data["lng"], 64)
	loc.Speed, _ = strconv.ParseFloat(data["speed"], 64)
	heading, _ := strconv.Atoi(data["heading"])
	loc.Heading = heading
	loc.Ignition = data["ignition"] == "true"
	loc.LastSeen, _ = time.Parse(time.RFC3339, data["last_seen"])
	return loc, nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
