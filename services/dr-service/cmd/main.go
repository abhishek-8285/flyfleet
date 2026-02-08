package main

import (
	"context"
	"os"

	"github.com/flyfleet/dr-service/internal/events"
	"github.com/flyfleet/dr-service/internal/handlers"
	"github.com/flyfleet/dr-service/internal/queue"
	"github.com/flyfleet/dr-service/internal/repository"
	"github.com/flyfleet/dr-service/internal/service"
	"github.com/flyfleet/dr-service/internal/storage"
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

	// 3. DI & Infrastructure
	repo := repository.NewPostgresRepository(pool)

	// Redis / Queue
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	queueClient := queue.NewQueueClient(redisAddr)
	defer queueClient.Close()

	// Storage Provider
	var provider storage.BackupProvider
	providerType := os.Getenv("STORAGE_PROVIDER")
	switch providerType {
	case "AWS":
		provider, err = storage.NewS3Provider(context.Background(), os.Getenv("AWS_BUCKET"), os.Getenv("AWS_REGION"))
	case "AZURE":
		provider, err = storage.NewAzureProvider(os.Getenv("AZURE_CONNECTION_STRING"), os.Getenv("AZURE_CONTAINER"))
	default:
		// Default to Local
		localDir := "./backups"
		if d := os.Getenv("LOCAL_BACKUP_DIR"); d != "" {
			localDir = d
		}
		provider, err = storage.NewLocalProvider(localDir)
	}
	
	if err != nil {
		logger.Fatal("Failed to initialize storage provider", zap.Error(err))
	}

	// Event Broadcaster
	broadcaster := events.NewBroadcaster(redisAddr)
	defer broadcaster.Close()

	// Worker Server
	worker := queue.NewWorkerServer(redisAddr, repo, provider, logger, broadcaster)
	go func() {
		logger.Info("Starting Worker Server")
		if err := worker.Run(); err != nil {
			logger.Fatal("Worker server failed", zap.Error(err))
		}
	}()

	svc := service.NewDRService(repo, logger, queueClient)
	h := handlers.NewDRHandlers(svc, logger)

	// 4. Router
	r := gin.Default()
	
	// Middleware
	r.Use(func(c *gin.Context) {
		logger.Info("Request", zap.String("method", c.Request.Method), zap.String("path", c.Request.URL.Path))
		c.Next()
	})

	api := r.Group("/api/v1/admin/dr")
	{
		api.GET("/backups", h.ListBackups)
		api.POST("/backups", h.CreateBackup)
		api.GET("/backups/:id", h.GetBackup)
		api.POST("/backups/:id/verify", h.VerifyBackup)
		api.GET("/storage-pools", h.ListStoragePools)

		api.GET("/restores", h.ListRestores)
		api.POST("/restores", h.CreateRestore)
		api.GET("/restores/:id", h.GetRestore)
		api.GET("/restores/:id/logs", h.GetRestoreLogs)
		api.POST("/restores/:id/cancel", h.CancelRestore)

		// DR Testing routes
		api.GET("/tests", h.ListDRTests)
		api.POST("/tests", h.CreateDRTest)
		api.GET("/tests/metrics", h.GetDRTestMetrics)
		api.GET("/tests/settings", h.GetDRTestSettings)
		api.PUT("/tests/settings", h.UpdateDRTestSettings)
		api.GET("/tests/:id", h.GetDRTest)
		api.GET("/tests/:id/logs", h.GetDRTestLogs)
		api.POST("/tests/:id/abort", h.AbortDRTest)

		// SSE endpoint for real-time updates
		api.GET("/events", func(c *gin.Context) {
			c.Writer.Header().Set("Content-Type", "text/event-stream")
			c.Writer.Header().Set("Cache-Control", "no-cache")
			c.Writer.Header().Set("Connection", "keep-alive")
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

			pubsub := broadcaster.Subscribe(c.Request.Context())
			defer pubsub.Close()
			ch := pubsub.Channel()

			for {
				select {
				case msg := <-ch:
					c.Writer.Write([]byte("data: " + msg.Payload + "\n\n"))
					c.Writer.Flush()
				case <-c.Request.Context().Done():
					return
				}
			}
		})
	}

	// 5. Start
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	
	logger.Info("Starting Disaster Recovery Service", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
