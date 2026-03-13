package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

type HealthChecker struct {
	db    *pgxpool.Pool
	redis *redis.Client
	nc    *nats.Conn
}

func NewHealthChecker(db *pgxpool.Pool, redisClient *redis.Client, nc *nats.Conn) *HealthChecker {
	return &HealthChecker{db: db, redis: redisClient, nc: nc}
}

func (h *HealthChecker) Status(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := make(map[string]string)
	
	// Database Check
	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			status["database"] = "unhealthy"
		} else {
			status["database"] = "healthy"
		}
	}

	// Redis Check
	if h.redis != nil {
		if err := h.redis.Ping(ctx).Err(); err != nil {
			status["redis"] = "unhealthy"
		} else {
			status["redis"] = "healthy"
		}
	}

	// NATS Check
	if h.nc != nil {
		if h.nc.Status() != nats.CONNECTED {
			status["nats"] = "unhealthy"
		} else {
			status["nats"] = "healthy"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
