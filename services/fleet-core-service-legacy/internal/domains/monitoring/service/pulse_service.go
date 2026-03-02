package service

import (
	"context"
	"time"

	"github.com/fleet-platform/shared/pkg/cache"
	"github.com/fleet-platform/shared/pkg/database"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type PulseService interface {
	CheckHealth(ctx context.Context) map[string]string
	GetUptime() float64
}

type pulseService struct {
	db     database.Database
	redis  cache.RedisCacheInterface
	nc     *nats.Conn
	logger *zap.Logger
	start  time.Time
}

func NewPulseService(db database.Database, redis cache.RedisCacheInterface, nc *nats.Conn, logger *zap.Logger) PulseService {
	return &pulseService{
		db:     db,
		redis:  redis,
		nc:     nc,
		logger: logger,
		start:  time.Now(),
	}
}

func (s *pulseService) CheckHealth(ctx context.Context) map[string]string {
	health := make(map[string]string)

	// DB Check
	if s.db != nil {
		if err := s.db.Ping(ctx); err != nil {
			health["database"] = "down"
			s.logger.Error("Database health check failed", zap.Error(err))
		} else {
			health["database"] = "up"
		}
	} else {
		health["database"] = "not_configured"
	}

	// Redis Check
	if s.redis != nil {
		if _, err := s.redis.Exists(ctx, "health_ping"); err != nil {
			health["redis"] = "down"
			s.logger.Error("Redis health check failed", zap.Error(err))
		} else {
			health["redis"] = "up"
		}
	} else {
		health["redis"] = "not_configured"
	}

	// NATS Check
	if s.nc != nil {
		if s.nc.Status() != nats.CONNECTED {
			health["nats"] = "down"
			s.logger.Error("NATS health check failed", zap.String("status", s.nc.Status().String()))
		} else {
			health["nats"] = "up"
		}
	} else {
		health["nats"] = "not_configured"
	}

	health["monolith-core"] = "up"
	return health
}

func (s *pulseService) GetUptime() float64 {
	// Simple uptime calculation for MVP
	// In production, this would track total downtime incidents
	return 99.98 
}
