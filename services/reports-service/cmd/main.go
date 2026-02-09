package main

import (
	"context"
	"os"

	"github.com/flyfleet/reports-service/internal/client"
	"github.com/flyfleet/reports-service/internal/handlers"
	"github.com/flyfleet/reports-service/internal/queue"
	"github.com/flyfleet/reports-service/internal/repository"
	"github.com/flyfleet/reports-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	// 1. Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 2. Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		logger.Fatal("Unable to parse DB URL", zap.Error(err))
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.Fatal("Unable to connect to DB", zap.Error(err))
	}
	defer pool.Close()

	// 3. Repository
	repo := repository.NewPostgresRepository(pool)

	// 4. Queue & Worker
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default for local dev
	}
	
	notifServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notifServiceURL == "" {
		notifServiceURL = "http://localhost:8080" // Default for local dev (adjust if needed)
	}
	notifClient := client.NewNotificationClient(notifServiceURL, logger)

	queueClient := queue.NewQueueClient(redisAddr)
	defer queueClient.Close()

	workerServer := queue.NewWorkerServer(redisAddr, repo, notifClient, logger)
	
	// Start Worker in background
	go func() {
		logger.Info("Starting Worker Server")
		if err := workerServer.Run(); err != nil {
			logger.Fatal("Worker Server failed", zap.Error(err))
		}
	}()

	// 5. Service & Handlers
	svc := service.NewReportService(repo, queueClient, logger)
	h := handlers.NewReportHandler(svc, logger)

	// 4. Router
	r := gin.Default()
	
	// Middleware
	r.Use(func(c *gin.Context) {
		logger.Info("Request", zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path))
		c.Next()
	})

	api := r.Group("/api/v1/admin/reports")
	{
		api.POST("/schedules", h.CreateReport)
		api.GET("/schedules", h.ListReports)
		api.GET("/schedules/:id", h.GetReport)
		api.PUT("/schedules/:id", h.UpdateReport)
		api.POST("/schedules/:id/pause", h.PauseReport)
		api.POST("/schedules/:id/resume", h.ResumeReport)
		
		api.GET("/executions", h.ListExecutions)
		
		api.GET("/settings", h.GetSettings)
		api.PUT("/settings", h.UpdateSettings)
	}

	// 5. Start
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	
	logger.Info("Starting Reports Service", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
