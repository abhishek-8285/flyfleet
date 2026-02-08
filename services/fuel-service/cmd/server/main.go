package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/flyfleet/fuel-service/internal/config"
	"github.com/flyfleet/fuel-service/internal/handlers"
	"github.com/flyfleet/fuel-service/internal/repository"
	"github.com/flyfleet/fuel-service/internal/service"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to parse database URL", zap.Error(err))
	}

	db, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	fuelRepo := repository.NewFuelLogRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// Initialize service
	fuelService := service.NewFuelService(fuelRepo, settingsRepo, logger)

	// Initialize handlers
	fuelHandler := handlers.NewFuelHandler(fuelService, logger)

	// Setup router
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// API Routes
	v1 := router.Group("/api/v1")
	{
		fuel := v1.Group("/fuel-logs")
		{
			fuel.POST("", fuelHandler.CreateLog)
			fuel.GET("", fuelHandler.ListLogs)
			fuel.GET("/:id", fuelHandler.GetLog)
			fuel.PUT("/:id", fuelHandler.UpdateLog)
			fuel.DELETE("/:id", fuelHandler.DeleteLog)
			fuel.PATCH("/:id/status", fuelHandler.ChangeStatus)
		}

		settings := v1.Group("/fuel/settings")
		{
			settings.GET("", fuelHandler.GetSettings)
			settings.PUT("", fuelHandler.UpdateSettings)
		}

		v1.GET("/fuel/stats", fuelHandler.GetStats)
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		fmt.Printf("Fuel service starting on port %s\n", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down fuel service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("fuel service exiting")
}
