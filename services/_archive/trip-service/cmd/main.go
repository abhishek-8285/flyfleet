package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/flyfleet/trip-service/internal/handlers"
	"github.com/flyfleet/trip-service/internal/repository"
	"github.com/flyfleet/trip-service/internal/service"
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
	tripService := service.NewTripService(tripRepo, logger)
	tripHandler := handlers.NewTripHandler(tripService, logger)

	// Setup Gin router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware(logger))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "trip-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Trip API endpoints
	api := r.Group("/api/v1")
	{
		trips := api.Group("/trips")
		{
			trips.POST("", tripHandler.CreateTrip)
			trips.GET("", tripHandler.ListTrips)
			trips.GET("/:id", tripHandler.GetTrip)
			trips.PUT("/:id", tripHandler.UpdateTrip)
			trips.DELETE("/:id", tripHandler.DeleteTrip)
			
			// Trip lifecycle management
			trips.POST("/:id/complete", tripHandler.CompleteTrip)
			trips.POST("/:id/cancel", tripHandler.CancelTrip)
			
			// Trip events and POD
			trips.POST("/:id/events", tripHandler.AddEvent)
			trips.POST("/:id/pod", tripHandler.SubmitPOD)
			
			// Route optimization
			trips.GET("/:id/optimize-route", tripHandler.OptimizeRoute)
		}

		stats := api.Group("/stats")
		{
			stats.GET("/drivers", tripHandler.GetDriverStats)
			stats.GET("/trips", tripHandler.GetTripStats)
			stats.GET("/summary", tripHandler.GetStats)
		}
	}

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

func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		status := c.Writer.Status()

		logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)
	}
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