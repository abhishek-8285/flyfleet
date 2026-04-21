package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"

	"github.com/flyfleet/fleet-core/internal/database"
	"github.com/flyfleet/fleet-core/internal/tracking/models"
)

const liveCacheTTL = 5 * time.Minute

// LiveCacheRepository provides Redis-backed caching for
// real-time vehicle positions.
type LiveCacheRepository struct {
	rdb *redis.Client
}

func NewLiveCacheRepository(rdb *redis.Client) *LiveCacheRepository {
	return &LiveCacheRepository{rdb: rdb}
}

// SetLiveLocation writes a vehicle's latest position to Redis with a TTL.
func (r *LiveCacheRepository) SetLiveLocation(ctx context.Context, vehicleID uuid.UUID, loc *models.LiveLocation) error {
	key := database.LiveLocationKey(vehicleID.String())
	fields := database.LiveLocationToHash(
		loc.Lat, loc.Lng, loc.Speed, loc.Heading,
		loc.FuelLevel, loc.Odometer, loc.Ignition, loc.Status, loc.LastSeen,
	)

	pipe := r.rdb.Pipeline()
	pipe.HSet(ctx, key, fields)
	pipe.Expire(ctx, key, liveCacheTTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("set live location: %w", err)
	}
	return nil
}

// GetLiveLocation reads a single vehicle's cached position.
func (r *LiveCacheRepository) GetLiveLocation(ctx context.Context, vehicleID uuid.UUID) (*models.LiveLocation, error) {
	key := database.LiveLocationKey(vehicleID.String())
	vals, err := rdbGetAll(ctx, r.rdb, key)
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, nil // cache miss
	}

	lat, lng, speed, heading, fuel, odometer, ignition, status, lastSeenUnix, ok := database.ParseLiveLocationHash(vals)
	if !ok {
		return nil, fmt.Errorf("parse live location hash: invalid data for %s", vehicleID)
	}

	return &models.LiveLocation{
		VehicleID: vehicleID,
		Lat:       lat,
		Lng:       lng,
		Speed:     speed,
		Heading:   heading,
		Ignition:  ignition,
		FuelLevel: fuel,
		Odometer:  odometer,
		LastSeen:  time.Unix(lastSeenUnix, 0),
		Status:    status,
	}, nil
}

// GetBatchLiveLocations reads cached positions for multiple vehicles
// using a Redis pipeline for efficiency.
func (r *LiveCacheRepository) GetBatchLiveLocations(ctx context.Context, vehicleIDs []uuid.UUID) (map[uuid.UUID]*models.LiveLocation, error) {
	result := make(map[uuid.UUID]*models.LiveLocation, len(vehicleIDs))

	pipe := r.rdb.Pipeline()
	cmds := make(map[string]*redis.MapStringStringCmd, len(vehicleIDs))

	for _, vid := range vehicleIDs {
		key := database.LiveLocationKey(vid.String())
		cmd := pipe.HGetAll(ctx, key)
		cmds[vid.String()] = cmd
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("pipeline exec: %w", err)
	}

	for _, vid := range vehicleIDs {
		vals := cmds[vid.String()].Val()
		if len(vals) == 0 {
			continue // cache miss — caller handles DB fallback
		}

		lat, lng, speed, heading, fuel, odometer, ignition, status, lastSeenUnix, ok := database.ParseLiveLocationHash(vals)
		if !ok {
			continue
		}

		result[vid] = &models.LiveLocation{
			VehicleID: vid,
			Lat:       lat,
			Lng:       lng,
			Speed:     speed,
			Heading:   heading,
			Ignition:  ignition,
			FuelLevel: fuel,
			Odometer:  odometer,
			LastSeen:  time.Unix(lastSeenUnix, 0),
			Status:    status,
		}
	}

	return result, nil
}

// PublishPositionUpdate publishes a position update to the org's
// Redis pub/sub channel for WebSocket fan-out.
func (r *LiveCacheRepository) PublishPositionUpdate(ctx context.Context, tenantID uuid.UUID, msg []byte) error {
	channel := database.RedisChannelForOrg(tenantID.String())
	return r.rdb.Publish(ctx, channel, msg).Err()
}

// GetGeofenceState returns whether a vehicle is currently inside
// a specific geofence (cached in Redis).
func (r *LiveCacheRepository) GetGeofenceState(ctx context.Context, vehicleID, geofenceID string) (bool, error) {
	key := database.GeofenceStateKey(vehicleID, geofenceID)
	val, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // no state tracked yet
	}
	if err != nil {
		return false, fmt.Errorf("get geofence state: %w", err)
	}
	return val == "inside", nil
}

// SetGeofenceState records whether a vehicle is inside a geofence.
func (r *LiveCacheRepository) SetGeofenceState(ctx context.Context, vehicleID, geofenceID string, inside bool) error {
	key := database.GeofenceStateKey(vehicleID, geofenceID)
	val := "outside"
	if inside {
		val = "inside"
	}
	return r.rdb.Set(ctx, key, val, 24*time.Hour).Err()
}

func rdbGetAll(ctx context.Context, rdb *redis.Client, key string) (map[string]string, error) {
	vals, err := rdb.HGetAll(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	return vals, nil
}
