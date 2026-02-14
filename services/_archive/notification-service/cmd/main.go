package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/notification-service/internal/handlers"
	"github.com/flyfleet/notification-service/internal/repository"
	"github.com/flyfleet/notification-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := struct {
		DatabaseURL string
		Port        string
	}{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = ""
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// Database connection
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	// Initialize components
	repo := repository.NewNotificationRepository(pool)
	notifService := service.NewNotificationService(repo, logger)
	notifHandlers := handlers.NewNotificationHandlers(notifService, logger)

	// Router setup
	r := gin.Default()

	// Middleware
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "notification-service"})
	})

	// API Routes
	api := r.Group("/api/v1")
	{
		// Internal/System Routes (should be protected by private network or internal auth)
		internal := api.Group("/internal/notifications")
		{
			internal.POST("", notifHandlers.SendNotification)
		}

		admin := api.Group("/admin/notifications")
		{
			admin.GET("", notifHandlers.ListLogs)
			admin.GET("/:id", notifHandlers.GetLogDetail)
			admin.POST("/:id/resend", notifHandlers.Resend)
			admin.GET("/settings", notifHandlers.GetSettings)
			admin.PUT("/settings", notifHandlers.UpdateSettings)

			// Preferences Manager
			admin.GET("/preferences", notifHandlers.ListPreferences)
			admin.GET("/preferences/:id", notifHandlers.GetUserPreference)
			admin.PUT("/preferences/:id", notifHandlers.UpdateUserPreference)
			admin.GET("/preferences/:id/audit", notifHandlers.GetPreferenceAudit)
			admin.GET("/preferences/global", notifHandlers.GetGlobalDefaults)
			admin.PUT("/preferences/global", notifHandlers.UpdateGlobalDefaults)
		}
	}

	// Server setup
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Listen and serve error", zap.Error(err))
		}
	}()

	logger.Info("Notification Service started", zap.String("port", cfg.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}
