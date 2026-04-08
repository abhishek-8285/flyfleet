package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedis creates a Redis client from a redis:// or localhost:port URL.
func NewRedis(ctx context.Context, url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		// Fallback: treat as host:port
		opts = &redis.Options{
			Addr:     url,
			Password: "",
			DB:       0,
		}
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return rdb, nil
}

// RedisChannelForOrg returns the Redis pub/sub channel name for
// telemetry updates belonging to the given organization.
func RedisChannelForOrg(orgID string) string {
	return fmt.Sprintf("org:%s:telemetry", orgID)
}

// LiveLocationKey returns the Redis key for a vehicle's cached position.
func LiveLocationKey(vehicleID string) string {
	return fmt.Sprintf("live:loc:%s", vehicleID)
}

// GeofenceStateKey returns the Redis key tracking whether a vehicle is
// inside a specific geofence.
func GeofenceStateKey(vehicleID, geofenceID string) string {
	return fmt.Sprintf("geofence:state:%s:%s", vehicleID, geofenceID)
}

// HashFields maps a LiveLocation to Redis hash fields.
func LiveLocationToHash(lat, lng, speed, heading, fuel, odometer float64,
	ignition bool, status string, lastSeen time.Time) map[string]interface{} {
	return map[string]interface{}{
		"lat":       lat,
		"lng":       lng,
		"speed":     speed,
		"heading":   heading,
		"fuel":      fuel,
		"odometer":  odometer,
		"ignition":  b2i(ignition),
		"status":    status,
		"last_seen": lastSeen.Unix(),
	}
}

// HashFromRedis parses a Redis hash into a LiveLocation-like struct.
func ParseLiveLocationHash(vals map[string]string) (
	lat, lng, speed, heading, fuel, odometer float64,
	ignition bool, status string, lastSeenUnix int64, ok bool) {

	var err error
	if lat, err = parseFloat(vals, "lat"); err != nil {
		return
	}
	if lng, err = parseFloat(vals, "lng"); err != nil {
		return
	}
	if speed, err = parseFloat(vals, "speed"); err != nil {
		return
	}
	if heading, err = parseFloat(vals, "heading"); err != nil {
		return
	}
	if fuel, err = parseFloat(vals, "fuel"); err != nil {
		fuel = 0
	}
	if odometer, err = parseFloat(vals, "odometer"); err != nil {
		odometer = 0
	}
	ignition = vals["ignition"] == "1"
	status = vals["status"]
	lastSeenUnix, _ = parseInt(vals, "last_seen")
	ok = true
	return
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func parseFloat(m map[string]string, key string) (float64, error) {
	v := m[key]
	if v == "" {
		return 0, fmt.Errorf("missing key %s", key)
	}
	var f float64
	_, err := fmt.Sscanf(v, "%f", &f)
	return f, err
}

func parseInt(m map[string]string, key string) (int64, error) {
	v := m[key]
	if v == "" {
		return 0, fmt.Errorf("missing key %s", key)
	}
	var i int64
	_, err := fmt.Sscanf(v, "%d", &i)
	return i, err
}
