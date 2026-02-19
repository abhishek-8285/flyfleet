package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/flyfleet/driver-service/internal/handlers"
	dm "github.com/flyfleet/driver-service/internal/middleware"
	"github.com/flyfleet/driver-service/internal/repository"
	"github.com/flyfleet/driver-service/internal/service"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	port := getEnv("PORT", "8083")
	dbURL := getEnv("DATABASE_URL", "")
	natsURL := getEnv("NATS_URL", nats.DefaultURL)
	jwtSecret := getEnv("JWT_SECRET", "your-jwt-secret-key-here")

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
	driverRepo := repository.NewDriverRepository(db)
	driverService := service.NewDriverService(driverRepo, logger, service.WithNATS(nc))
	driverHandler := handlers.NewDriverHandler(driverService, logger)

	// Setup Chi router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(dm.AuthMiddleware(jwtSecret))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "driver-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Driver API endpoints
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/drivers", func(r chi.Router) {
			r.Post("/", driverHandler.CreateDriver)
			r.Get("/", driverHandler.ListDrivers)
			
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", driverHandler.GetDriver)
				r.Put("/", driverHandler.UpdateDriver)
				r.Delete("/", driverHandler.DeleteDriver)
				
				r.Post("/assign-vehicle", driverHandler.AssignVehicle)
				r.Post("/unassign-vehicle", driverHandler.UnassignVehicle)
				r.Post("/update-status", driverHandler.UpdateStatus)
				
				r.Get("/behavior-score", driverHandler.GetBehaviorScore)
				r.Post("/update-behavior", driverHandler.UpdateBehaviorScore)
				r.Get("/analytics", driverHandler.GetAnalytics)
				
				r.Post("/documents", driverHandler.UploadDocument)
				r.Get("/documents", driverHandler.ListDocuments)
				r.Delete("/documents/{doc_id}", driverHandler.DeleteDocument)
				
				r.Post("/attendance", driverHandler.MarkAttendance)
				r.Get("/attendance", driverHandler.GetAttendance)
				r.Post("/schedule", driverHandler.SetSchedule)
				r.Get("/schedule", driverHandler.GetSchedule)
			})
		})
	})

	// Start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting driver service", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down driver service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Driver service exited")
}
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to parse pagination parameters
func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(param)
	if err != nil || value < 1 {
		return defaultValue
	}
	
	return value
}
