package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	google_grpc "google.golang.org/grpc"

	tripgrpc "github.com/flyfleet/trip-service/internal/grpc"
	"github.com/flyfleet/trip-service/internal/handlers"
	"github.com/flyfleet/trip-service/internal/repository"
	"github.com/flyfleet/trip-service/internal/service"
	pb "github.com/flyfleet/proto/trip/v1"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	port := getEnv("PORT", "8082")
	dbURL := getEnv("DATABASE_URL", "")
	natsURL := getEnv("NATS_URL", nats.DefaultURL)

	// Database connection
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(context.Background()); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Connected to database")

	// NATS connection
	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()
	logger.Info("Connected to NATS")

	// Initialize components
	tripRepo := repository.NewTripRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)
	tripService := service.NewTripService(tripRepo, telemetryRepo, logger)
	tripHandler := handlers.NewTripHandler(tripService, logger)

	settingsRepo := repository.NewSettingsRepository(db)
	settingsService := service.NewSettingsService(settingsRepo, logger)
	settingsHandler := handlers.NewSettingsHandler(settingsService, logger)

	// Start gRPC Server
	grpcPort := getEnv("GRPC_PORT", "50053")
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			logger.Fatal("Failed to listen for gRPC", zap.Error(err))
		}
		
		grpcServer := google_grpc.NewServer()
		tripServer := tripgrpc.NewTripServer(tripService, logger)
		pb.RegisterTripServiceServer(grpcServer, tripServer)
		
		logger.Info("Starting trip gRPC server", zap.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Setup Chi router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "trip-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/trips", func(r chi.Router) {
			// Settings routes registered before ID route to avoid collision
			r.Get("/settings", settingsHandler.GetTripSettings)
			r.Put("/settings", settingsHandler.UpdateTripSettings)

			r.Post("/", tripHandler.CreateTrip)
			r.Get("/", tripHandler.ListTrips)
			
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", tripHandler.GetTrip)
				r.Put("/", tripHandler.UpdateTrip)
				r.Delete("/", tripHandler.DeleteTrip)
				
				r.Get("/telemetry", tripHandler.GetTripTelemetry)
				r.Post("/complete", tripHandler.CompleteTrip)
				r.Post("/cancel", tripHandler.CancelTrip)
				r.Post("/events", tripHandler.AddEvent)
				r.Post("/pod", tripHandler.SubmitPOD)
				r.Get("/optimize-route", tripHandler.OptimizeRoute)
			})
		})

		r.Route("/stats", func(r chi.Router) {
			r.Get("/drivers", tripHandler.GetDriverStats)
			r.Get("/trips", tripHandler.GetTripStats)
			r.Get("/summary", tripHandler.GetStats)
		})
	})

	// Start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting trip service", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down trip service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Trip service exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}