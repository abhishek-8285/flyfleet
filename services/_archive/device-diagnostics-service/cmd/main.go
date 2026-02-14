package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/device-diagnostics-service/internal/config"
	"github.com/flyfleet/device-diagnostics-service/internal/database"
	"github.com/flyfleet/device-diagnostics-service/internal/handlers"
	"github.com/flyfleet/device-diagnostics-service/internal/repository"
	"github.com/flyfleet/device-diagnostics-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	// Initialize logger
	var logger *zap.Logger
	var err error
	if cfg.LogLevel == "debug" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Connect to database
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize layers
	repo := repository.NewRepository(db, logger)
	svc := service.NewDeviceDiagnosticsService(repo, logger)
	h := handlers.NewHandlers(svc)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Dashboard
		r.Get("/devices/summary", h.GetSummary)
		r.Get("/devices/status", h.ListDevices)

		// Device details
		r.Get("/devices/{id}", h.GetDevice)
		r.Get("/devices/{id}/sensors", h.GetDeviceSensors)
		r.Get("/devices/{id}/logs", h.GetDeviceLogs)
		r.Get("/devices/{id}/telemetry", h.GetDeviceTelemetry)

		// Commands
		r.Post("/devices/{id}/commands", h.SendCommand)
		r.Get("/devices/{id}/commands", h.GetCommandHistory)

		// Firmware
		r.Get("/firmware", h.ListFirmware)
		r.Post("/firmware", h.CreateFirmware)
		r.Get("/firmware/rollouts", h.ListRollouts)
		r.Post("/firmware/rollouts", h.CreateRollout)
	})

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("Device Diagnostics Service starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	<-done
	logger.Info("Server stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped")
}
