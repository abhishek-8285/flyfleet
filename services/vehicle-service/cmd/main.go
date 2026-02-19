package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/vehicle-service/internal/handlers"
	"github.com/flyfleet/vehicle-service/internal/processor"
	"github.com/flyfleet/vehicle-service/internal/repository"
	"github.com/flyfleet/vehicle-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration from environment
	port := getEnv("PORT", "8084")
	databaseURL := getEnv("DATABASE_URL", "")

	// Initialize database connection
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(context.Background()); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	logger.Info("Connected to database successfully")

	// Initialize repository
	vehicleRepo := repository.NewVehicleRepository(db, logger)
	bulkRepo := repository.NewBulkRepository(db)

	// Initialize processor
	bulkProcessor := processor.NewBulkProcessor(bulkRepo, vehicleRepo, logger)

	// Initialize service
	vehicleService := service.NewVehicleService(vehicleRepo, logger)
	bulkService := service.NewBulkService(bulkRepo, bulkProcessor, logger)
	
	// Trip management
	tripRepo := repository.NewTripRepository(db)
	tripService := service.NewTripService(tripRepo, logger)
	tripHandlers := handlers.NewTripHandlers(tripService, logger)

	// Initialize handlers
	vehicleHandlers := handlers.NewVehicleHandlers(vehicleService, logger)
	bulkHandlers := handlers.NewBulkHandlers(bulkService, logger)

	// Setup Chi router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(loggingMiddleware(logger))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "vehicle-service",
			"timestamp": time.Now().Unix(),
		})
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Vehicle routes
		r.Route("/vehicles", func(r chi.Router) {
			r.Get("/", vehicleHandlers.ListVehicles)
			r.Post("/", vehicleHandlers.CreateVehicle)
			r.Get("/stats", vehicleHandlers.GetStats)
			
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", vehicleHandlers.GetVehicle)
				r.Put("/", vehicleHandlers.UpdateVehicle)
				r.Delete("/", vehicleHandlers.DeleteVehicle)
				r.Post("/assign-driver", vehicleHandlers.AssignDriver)
			})
		})

		// Bulk Job routes
		r.Route("/bulk-jobs", func(r chi.Router) {
			r.Post("/", bulkHandlers.CreateBulkJob)
			r.Get("/", bulkHandlers.ListBulkJobs)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", bulkHandlers.GetBulkJob)
				r.Get("/records", bulkHandlers.GetJobRecords)
				r.Post("/retry", bulkHandlers.RetryBulkJob)
			})
		})

		// Trip routes
		r.Route("/trips", func(r chi.Router) {
			r.Get("/", tripHandlers.ListTrips)
			r.Post("/", tripHandlers.CreateTrip)
			r.Get("/settings", tripHandlers.GetSettings)
			r.Put("/settings", tripHandlers.UpdateSettings)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", tripHandlers.GetTrip)
				r.Put("/", tripHandlers.UpdateTrip)
				r.Get("/telemetry", tripHandlers.GetTelemetry)
			})
		})
	})

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in goroutine
	go func() {
		logger.Info("Starting vehicle service", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down vehicle service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Vehicle service exited gracefully")
}

func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Custom response writer to capture status code
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			
			next.ServeHTTP(ww, r)
			
			duration := time.Since(start)
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.status),
				zap.Duration("duration", duration),
				zap.String("client_ip", r.RemoteAddr),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
