package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/analytics-service/internal/handlers"
	"github.com/flyfleet/analytics-service/internal/repository"
	"github.com/flyfleet/analytics-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
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
	handler := handlers.NewAnalyticsHandler(analyticsService)

	// Setup Gin router for consistency across services
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware(logger))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "analytics-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API routes
	api := r.Group("/analytics")
	{
		// Basic Analytics (Legacy Wrapper Style)
		api.GET("/vehicles", handler.GetVehicleAnalytics)
		api.GET("/drivers", handler.GetDriverAnalytics)
		api.GET("/trips", handler.GetTripAnalytics)
		api.GET("/forecast", handler.GetForecastAnalytics)

		// Predictive Models
		api.POST("/models", handler.CreatePredictiveModel)
		api.GET("/models", handler.ListPredictiveModels)
		api.GET("/models/:id", handler.GetPredictiveModel)
		api.PATCH("/models/:id", handler.UpdatePredictiveModel)
		api.DELETE("/models/:id", handler.DeletePredictiveModel)
		api.GET("/models/version", handler.GetPredictiveModelByVersion)

		// Prediction Results
		api.POST("/results", handler.CreatePredictionResult)
		api.GET("/results", handler.ListPredictionResults)
		api.GET("/results/:id", handler.GetPredictionResult)
		api.GET("/results/model/:modelId", handler.GetPredictionResultsByModel)
		api.GET("/results/org", handler.GetPredictionResultsByOrg)

		// Analytics Jobs
		api.POST("/jobs", handler.CreateAnalyticsJob)
		api.GET("/jobs", handler.ListAnalyticsJobs)
		api.GET("/jobs/:id", handler.GetAnalyticsJob)
		api.PATCH("/jobs/:id", handler.UpdateAnalyticsJob)
		api.DELETE("/jobs/:id", handler.DeleteAnalyticsJob)

		// Fleet Analytics
		api.POST("/fleet", handler.CreateFleetAnalytics)
		api.GET("/fleet", handler.ListFleetAnalytics)
		api.GET("/fleet/:id", handler.GetFleetAnalytics)
		api.GET("/fleet/period", handler.GetFleetAnalyticsByPeriod)
		api.PATCH("/fleet/:id", handler.UpdateFleetAnalytics)
		api.DELETE("/fleet/:id", handler.DeleteFleetAnalytics)

		// Driver Behavior
		api.POST("/driver-behavior", handler.CreateDriverBehaviorAnalytics)
		api.GET("/driver-behavior", handler.ListDriverBehaviorAnalytics)
		api.GET("/driver-behavior/:id", handler.GetDriverBehaviorAnalytics)
		api.GET("/driver-behavior/driver", handler.GetDriverBehaviorAnalyticsByDriver)
		api.GET("/driver-behavior/org", handler.GetDriverBehaviorAnalyticsByOrg)
		api.PATCH("/driver-behavior/:id", handler.UpdateDriverBehaviorAnalytics)
		api.DELETE("/driver-behavior/:id", handler.DeleteDriverBehaviorAnalytics)

		// Predictive Maintenance
		api.POST("/maintenance", handler.CreatePredictiveMaintenanceAnalytics)
		api.GET("/maintenance", handler.ListPredictiveMaintenanceAnalytics)
		api.GET("/maintenance/:id", handler.GetPredictiveMaintenanceAnalytics)
		api.GET("/maintenance/vehicle", handler.GetPredictiveMaintenanceAnalyticsByVehicle)
		api.PATCH("/maintenance/:id", handler.UpdatePredictiveMaintenanceAnalytics)
		api.DELETE("/maintenance/:id", handler.DeletePredictiveMaintenanceAnalytics)
	}

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
