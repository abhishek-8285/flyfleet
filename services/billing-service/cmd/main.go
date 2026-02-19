package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/billing-service/internal/config"
	"github.com/flyfleet/billing-service/internal/events"
	inv_handler "github.com/flyfleet/billing-service/internal/handlers/invoice"
	"github.com/flyfleet/billing-service/internal/middleware"
	inv_repo "github.com/flyfleet/billing-service/internal/repository/postgres"
	inv_service "github.com/flyfleet/billing-service/internal/service"
	"github.com/flyfleet/billing-service/pkg/database"
	"github.com/go-chi/chi/v5"
	mw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Database Connection
	pool, err := database.Connect(ctx, database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPass,
		DBName:   cfg.DBName,
		SSLMode:  cfg.SSLMode,
	}, logger)
	if err != nil {
		logger.Fatal("failed to connect to db", zap.Error(err))
	}
	defer pool.Close()

	// NATS Publisher
	publisher, err := events.NewNATSPublisher(cfg.NATSURL, logger)
	if err != nil {
		logger.Error("failed to connect to NATS, continuing in degraded mode", zap.Error(err))
	} else {
		defer publisher.Close()
	}

	// Repositories, Services, Handlers
	repo := inv_repo.NewInvoiceRepository(pool)
	service := inv_service.NewInvoiceService(repo, publisher)
	handler := inv_handler.NewHandler(service)

	// Repositories, Services, Handlers
	r := chi.NewRouter()

	// Mandatory Middleware Chain
	r.Use(mw.RequestID)
	r.Use(mw.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Metrics)
	r.Use(middleware.Auth([]byte(cfg.JWTSecret)))

	// Register Routes
	inv_handler.RegisterRoutes(r, handler)

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	go func() {
		logger.Info("starting server", zap.Int("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
	}
	logger.Info("server exited")
}

