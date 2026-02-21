package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/analytics-service/internal/handlers"
	"github.com/flyfleet/analytics-service/internal/repository"
	"github.com/flyfleet/analytics-service/internal/service"
	analyticsgrpc "github.com/flyfleet/analytics-service/internal/grpc"
	pb "github.com/flyfleet/proto/analytics/v1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	google_grpc "google.golang.org/grpc"
)

func main() {
	// Initialize logger FIRST before any other operations
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Load configuration
	port := getEnv("PORT", "8080")
	databaseURL := getEnv("DATABASE_URL", "")

	// Create database connection pool
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Connected to database")

	// Create repository, service, and handler
	repo := repository.NewAnalyticsRepository(db)
	analyticsService := service.NewAnalyticsService(repo)
	handler := handlers.NewAnalyticsHandler(analyticsService, logger)

	// Start gRPC Server
	grpcPort := getEnv("GRPC_PORT", "50059")
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			logger.Fatal("Failed to listen for gRPC", zap.Error(err))
		}
		
		grpcServer := google_grpc.NewServer()
		aServer := analyticsgrpc.NewAnalyticsServer(analyticsService, logger)
		pb.RegisterAnalyticsServiceServer(grpcServer, aServer)
		
		logger.Info("Starting analytics gRPC server", zap.String("port", grpcPort))
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
			"service":   "analytics-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API routes
	r.Route("/analytics", func(r chi.Router) {
		// Basic Analytics
		r.Get("/vehicles", handler.GetVehicleAnalytics)
		r.Get("/drivers", handler.GetDriverAnalytics)
		r.Get("/trips", handler.GetTripAnalytics)
		r.Get("/forecast", handler.GetForecastAnalytics)

		// Predictive Models
		r.Route("/models", func(r chi.Router) {
			r.Post("/", handler.CreatePredictiveModel)
			r.Get("/", handler.ListPredictiveModels)
			r.Get("/version", handler.GetPredictiveModelByVersion)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handler.GetPredictiveModel)
				r.Patch("/", handler.UpdatePredictiveModel)
				r.Delete("/", handler.DeletePredictiveModel)
			})
		})

		// Prediction Results
		r.Route("/results", func(r chi.Router) {
			r.Post("/", handler.CreatePredictionResult)
			r.Get("/", handler.ListPredictionResults)
			r.Get("/org", handler.GetPredictionResultsByOrg)
			r.Get("/{id}", handler.GetPredictionResult)
			r.Get("/model/{modelId}", handler.GetPredictionResultsByModel)
		})

		// Analytics Jobs
		r.Route("/jobs", func(r chi.Router) {
			r.Post("/", handler.CreateAnalyticsJob)
			r.Get("/", handler.ListAnalyticsJobs)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handler.GetAnalyticsJob)
				r.Patch("/", handler.UpdateAnalyticsJob)
				r.Delete("/", handler.DeleteAnalyticsJob)
			})
		})

		// Fleet Analytics
		r.Route("/fleet", func(r chi.Router) {
			r.Post("/", handler.CreateFleetAnalytics)
			r.Get("/", handler.ListFleetAnalytics)
			r.Get("/period", handler.GetFleetAnalyticsByPeriod)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handler.GetFleetAnalytics)
				r.Patch("/", handler.UpdateFleetAnalytics)
				r.Delete("/", handler.DeleteFleetAnalytics)
			})
		})

		// Driver Behavior
		r.Route("/driver-behavior", func(r chi.Router) {
			r.Post("/", handler.CreateDriverBehaviorAnalytics)
			r.Get("/", handler.ListDriverBehaviorAnalytics)
			r.Get("/driver", handler.GetDriverBehaviorAnalyticsByDriver)
			r.Get("/org", handler.GetDriverBehaviorAnalyticsByOrg)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handler.GetDriverBehaviorAnalytics)
				r.Patch("/", handler.UpdateDriverBehaviorAnalytics)
				r.Delete("/", handler.DeleteDriverBehaviorAnalytics)
			})
		})

		// Predictive Maintenance
		r.Route("/maintenance", func(r chi.Router) {
			r.Post("/", handler.CreatePredictiveMaintenanceAnalytics)
			r.Get("/", handler.ListPredictiveMaintenanceAnalytics)
			r.Get("/vehicle", handler.GetPredictiveMaintenanceAnalyticsByVehicle)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", handler.GetPredictiveMaintenanceAnalytics)
				r.Patch("/", handler.UpdatePredictiveMaintenanceAnalytics)
				r.Delete("/", handler.DeletePredictiveMaintenanceAnalytics)
			})
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting analytics service", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down analytics service...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Analytics service exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
