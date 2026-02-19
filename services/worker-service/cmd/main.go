package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fleet-platform/shared/pkg/async"
	"github.com/fleet-platform/shared/pkg/database"
	"github.com/fleet-platform/shared/pkg/webhook"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/flyfleet/worker-service/internal/handlers"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	// Load Configuration (Environment Variables)
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	}

	// Initialize Database
	db, err := database.NewPostgres(dbURL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Webhook Client
	webhookClient := webhook.NewWebhookClient(logger)

	// Initialize Task Handlers
	workerHandlers := handlers.NewWorkerHandlers(db, webhookClient, logger)

	// Setup Asynq Server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 20,
			Queues: map[string]int{
				"critical": 10,
				"default":  5,
				"low":      1,
			},
			Logger: logger.Sugar(),
		},
	)

	// Setup Asynq Mux
	mux := asynq.NewServeMux()
	mux.HandleFunc(async.TypeWebhookDeliver, workerHandlers.HandleWebhookDeliver)
	mux.HandleFunc(async.TypeReportGenerate, workerHandlers.HandleReportGenerate)
	mux.HandleFunc(async.TypeMaintenanceReminder, workerHandlers.HandleMaintenanceReminder)
	mux.HandleFunc(async.TypeUsageAggregation, workerHandlers.HandleUsageAggregation)
	mux.HandleFunc(async.TypeMonthlyInvoice, workerHandlers.HandleMonthlyInvoiceGeneration)

	// Start Worker
	go func() {
		logger.Info("Starting worker service", zap.String("redis", redisAddr))
		if err := srv.Run(mux); err != nil {
			logger.Fatal("Worker service crashed", zap.Error(err))
		}
	}()

	healthPort := os.Getenv("HEALTH_PORT")
	if healthPort == "" {
		healthPort = "8098"
	}
	healthSrv := &http.Server{
		Addr: ":" + healthPort,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/health" {
				http.NotFound(w, r)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	}
	go func() {
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("worker health server stopped", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker service...")
	srv.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = healthSrv.Shutdown(ctx)
	logger.Info("Worker service exited")
}
