package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gonotify/gonotify/internal/config"
	"github.com/gonotify/gonotify/internal/events"
	"github.com/gonotify/gonotify/internal/observability"
)

const bootstrapTenantID = "00000000-0000-0000-0000-000000000001"

type publishEventRequest struct {
	Event  string         `json:"event"`
	UserID string         `json:"user_id"`
	Data   map[string]any `json:"data"`
}

func main() {
	cfg := config.Load()
	logger := observability.NewLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	eventService := events.NewService(db, logger)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	r.Handle("/metrics", promhttp.Handler())

	r.Route("/v1", func(r chi.Router) {
		r.Post("/events", func(w http.ResponseWriter, r *http.Request) {
			var req publishEventRequest

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}

			if req.Event == "" {
				writeError(w, http.StatusBadRequest, "event is required")
				return
			}

			idempotencyKey := r.Header.Get("Idempotency-Key")

			resp, err := eventService.Publish(r.Context(), events.PublishRequest{
				TenantID:       bootstrapTenantID,
				EventType:      req.Event,
				UserID:         req.UserID,
				IdempotencyKey: idempotencyKey,
				Data:           req.Data,
				Metadata: map[string]any{
					"source": "api",
				},
			})
			if err != nil {
				logger.ErrorContext(
					r.Context(),
					"failed to publish event",
					"error", err,
				)

				writeError(w, http.StatusInternalServerError, "failed to publish event")
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)

			_ = json.NewEncoder(w).Encode(resp)
		})
	})

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("api server starting", "addr", cfg.HTTPAddr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown api server", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"status":  status,
			"message": message,
		},
	})
}
