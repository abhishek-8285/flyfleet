package main

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fleet-platform/shared/pkg/cache"
	"github.com/fleet-platform/shared/pkg/database"
	"github.com/fleet-platform/shared/pkg/messaging"
	"github.com/fleet-platform/shared/pkg/models"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dbURL := os.Getenv("DATABASE_URL")
	natsURL := os.Getenv("NATS_URL")
	redisAddr := os.Getenv("REDIS_ADDR")

	// 1. Database
	db, err := database.NewPostgres(dbURL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// 2. Redis for Last Known State (Pillar 2 & 6)
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisCache.Close()

	// 3. NATS Messaging (with JetStream)
	mb, err := messaging.NewNATSClient(natsURL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer mb.Close()

	js, _ := mb.JetStream()

	// 4. Create Stream if it doesn't exist
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "TELEMETRY",
		Subjects: []string{"telemetry.>"},
		Storage:  nats.FileStorage,
		MaxAge:   30 * 24 * time.Hour,
	})
	if err != nil {
		logger.Warn("Stream might already exist or failed to create", zap.Error(err))
	}

	// 5. GPS Consumer (with Metering)
	_, err = js.QueueSubscribe("telemetry.gps", "ingestion_gps_group", func(msg *nats.Msg) {
		var event struct {
			Type      string              `json:"type"`
			Timestamp string              `json:"timestamp"`
			OrgID     string              `json:"org_id"`
			VehicleID string              `json:"vehicle_id"`
			Data      models.GPSTelemetry `json:"data"`
		}

		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.Error("Failed to unmarshal GPS event", zap.Error(err))
			msg.Term()
			return
		}

		ctx := context.Background()

		// --- Pillar 6: Usage Metering ---
		go func() {
			// A. Log Heartbeat
			_ = db.WithTenant(ctx, event.OrgID, func(ctx context.Context, tx database.Transaction) error {
				return tx.Execute(ctx, `
					INSERT INTO consumption_events (time, org_id, vehicle_id, meter_code, amount)
					VALUES ($1, $2, $3, 'HEARTBEAT_COUNT', 1)`,
					time.Now(), event.OrgID, event.VehicleID)
			})

			// B. Calculate Distance if previous point exists
			lastPosKey := "last_pos:" + event.VehicleID
			prevData, _ := redisCache.Get(ctx, lastPosKey)
			if prevData != nil {
				var prev struct {
					Lat float64 `json:"lat"`
					Lon float64 `json:"lon"`
				}
				if err := json.Unmarshal(prevData, &prev); err == nil {
					dist := haversine(prev.Lat, prev.Lon, event.Data.Latitude, event.Data.Longitude)
					if dist > 0.01 { // Only log if > 10m
						_ = db.WithTenant(ctx, event.OrgID, func(ctx context.Context, tx database.Transaction) error {
							return tx.Execute(ctx, `
								INSERT INTO consumption_events (time, org_id, vehicle_id, meter_code, amount)
								VALUES ($1, $2, $3, 'DISTANCE_KM', $4)`,
								time.Now(), event.OrgID, event.VehicleID, dist)
						})
					}
				}
			}

			// Update Last Known State
			currPos, _ := json.Marshal(map[string]float64{
				"lat": event.Data.Latitude,
				"lon": event.Data.Longitude,
			})
			_ = redisCache.Set(ctx, lastPosKey, currPos, 24*time.Hour)
		}()

		// RLS storage
		err := db.WithTenant(ctx, event.OrgID, func(ctx context.Context, tx database.Transaction) error {
			return tx.Execute(ctx, `
				INSERT INTO gps_locations (
					time, vehicle_id, org_id, latitude, longitude, altitude, 
					speed, heading, accuracy, ignition, odometer, battery_level, 
					satellites, raw_data
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
				event.Data.Time, event.Data.VehicleID, event.Data.OrgID,
				event.Data.Latitude, event.Data.Longitude, event.Data.Altitude,
				event.Data.Speed, event.Data.Heading, event.Data.Accuracy,
				event.Data.Ignition, event.Data.Odometer, event.Data.Battery,
				event.Data.Satellites, event.Data.RawData,
			)
		})

		if err != nil {
			logger.Error("Failed to store GPS data", zap.Error(err))
			msg.Nak()
			return
		}

		msg.Ack()
	}, nats.Durable("gps_ingest_durable"), nats.ManualAck())

	healthPort := os.Getenv("HEALTH_PORT")
	if healthPort == "" {
		healthPort = "8099"
	}
	healthSrv := &http.Server{
		Addr: ":" + healthPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/health" {
				http.NotFound(w, r)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	}
	go func() {
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("health server stopped", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down ingestion service...")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = healthSrv.Shutdown(ctx)
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
