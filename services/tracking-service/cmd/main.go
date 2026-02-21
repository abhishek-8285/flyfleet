package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/fleet-platform/shared/pkg/cache"
	"github.com/fleet-platform/shared/pkg/database"
	"github.com/fleet-platform/shared/pkg/messaging"
	sharedMiddleware "github.com/fleet-platform/shared/pkg/middleware"
	"github.com/fleet-platform/shared/pkg/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	google_grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/flyfleet/tracking-service/internal/analytics"
	"github.com/flyfleet/tracking-service/internal/config"
	trackinggrpc "github.com/flyfleet/tracking-service/internal/grpc"
	"github.com/flyfleet/tracking-service/internal/handlers"
	"github.com/flyfleet/tracking-service/internal/processor"
	"github.com/flyfleet/tracking-service/internal/repository"
	trackingservice "github.com/flyfleet/tracking-service/internal/service"
	pb "github.com/flyfleet/tracking-service/proto/tracking/v1"
)

func main() {
	// 1. Initialize Base Service (Config, Logger, DB, Graceful Shutdown)
	base, err := service.NewBase("tracking-service", service.Config{
		DatabaseURL: "", // Gets from env
		LogLevel:    "info",
		GRPCPort:    "50052",
		Port:        "8081",
	})
	if err != nil {
		fmt.Printf("Fatal: failed to initialize base service: %v\n", err)
		return
	}

	// 2. Load custom config overlays (JWT, Nats, ML URL)
	// We can bind them to base config or keep using custom config struct
	cfg := config.Load()

	base.Logger.Info("Base service connected and database pool established")
	zlogger := base.Logger.WithContext(context.Background())

	// 3. NATS connection
	mb, err := messaging.NewNATSClient(cfg.NatsURL, zlogger)
	if err != nil {
		base.Logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer mb.Close()

	// 4. Initialize Redis Cache
	redisCache, err := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, zlogger)
	if err != nil {
		base.Logger.Warn("Failed to connect to Redis, caching will be disabled", zap.Error(err))
	} else {
		defer redisCache.Close()
	}

	// 5. Initialize Domain Services & Repositories
	analyticsService := analytics.NewTelemetryAnalytics(base.DB, zlogger)

	// tracking-service uses raw PGX pool directly for some high-performance bulk operations
	pg := base.DB.(*database.Postgres)
	deviceRepo := repository.NewDeviceRepository(pg.GetPool(), zlogger)
	deviceService := trackingservice.NewDeviceService(deviceRepo, cfg.JWTSecret, nil, zlogger)
	deviceHandler := handlers.NewDeviceHandler(deviceService, zlogger)

	// Initialize Processor
	telemetryProcessor := processor.NewTelemetryProcessor(base.DB, redisCache, analyticsService, zlogger)

	// Subscribe to telemetry subjects
	subjects := []string{"telemetry.gps", "telemetry.obd", "telemetry.fuel", "telemetry.heartbeat"}
	for _, sub := range subjects {
		subject := sub // capture variable
		if err := mb.Subscribe(subject, func(data []byte) {
			msg := &nats.Msg{
				Subject: subject,
				Data:    data,
			}
			telemetryProcessor.HandleTelemetry(msg)
		}); err != nil {
			base.Logger.Error("Failed to subscribe to subject", zap.String("subject", subject), zap.Error(err))
		}
	}

	// 5. Initialize Handlers
	telemetryHandler := handlers.NewTelemetryHandler(base.DB, mb, zlogger)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, zlogger)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	// Let base service handle health endpoints on standard port (exposes /health)
	base.ServeHealth("8080") // Note: tracking HTTP acts on its own port, but health can be uniform

	// Public device routes (registration doesn't require auth)
	r.Post("/device/register", deviceHandler.Register)

	// Authenticated and Tenant-aware routes
	r.Group(func(r chi.Router) {
		r.Use(sharedMiddleware.AuthMiddleware(cfg))
		r.Use(sharedMiddleware.TenantMiddleware)

		// Device routes
		deviceHandler.RegisterRoutes(r)

		// Telemetry routes
		telemetryHandler.RegisterRoutes(r)

		// Analytics routes
		analyticsHandler.RegisterRoutes(r)
	})

	// 6. Define API and gRPC Runners
	httpRunner := func(ctx context.Context) error {
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.Port),
			Handler: r,
		}

		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			srv.Shutdown(shutdownCtx)
		}()

		base.Logger.Info("Starting tracking HTTP server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	}

	grpcRunner := func(ctx context.Context) error {
		trackingServer := trackinggrpc.NewTrackingServer(base.DB, mb, analyticsService, cfg.MLServiceURL, zlogger)
		
		grpcServerOptions := []google_grpc.ServerOption{}
		creds, err := base.GetTLSCredentials()
		if err != nil {
			base.Logger.Fatal("Failed to load TLS credentials", zap.Error(err))
		}
		if creds != nil {
			grpcServerOptions = append(grpcServerOptions, google_grpc.Creds(creds))
			base.Logger.Info("gRPC Server configured with mTLS auth")
		}
		
		grpcServer := google_grpc.NewServer(grpcServerOptions...)
		pb.RegisterTrackingServiceServer(grpcServer, trackingServer)
		reflection.Register(grpcServer)

		grpcPort := cfg.GRPCPort
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
		if err != nil {
			return fmt.Errorf("failed to listen for gRPC: %w", err)
		}

		go func() {
			<-ctx.Done()
			grpcServer.GracefulStop()
			base.Logger.Info("gRPC server stopped gracefully")
		}()

		base.Logger.Info("Starting tracking gRPC server", zap.String("port", grpcPort))
		return grpcServer.Serve(lis)
	}

	// 7. Execute context blocks
	if err := base.RunWithContext(context.Background(), httpRunner, grpcRunner); err != nil {
		base.Logger.Fatal("Service exited with error", zap.Error(err))
	}
}