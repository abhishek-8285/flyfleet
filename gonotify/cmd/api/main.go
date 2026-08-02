package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gonotify/gonotify/internal/audit"
	"github.com/gonotify/gonotify/internal/auth"
	"github.com/gonotify/gonotify/internal/consent"
	"github.com/gonotify/gonotify/internal/config"
	appcrypto "github.com/gonotify/gonotify/internal/crypto"
	"github.com/gonotify/gonotify/internal/events"
	"github.com/gonotify/gonotify/internal/inbox"
	"github.com/gonotify/gonotify/internal/observability"
	"github.com/gonotify/gonotify/internal/preference"
	"github.com/gonotify/gonotify/internal/provider"
	"github.com/gonotify/gonotify/internal/queue"
	"github.com/gonotify/gonotify/internal/realtime"
	"github.com/gonotify/gonotify/internal/server"
	"github.com/gonotify/gonotify/internal/suppression"
	"github.com/gonotify/gonotify/internal/template"
	"github.com/gonotify/gonotify/internal/tenant"
	"github.com/gonotify/gonotify/internal/user"
)

type publishEventOptions struct {
	Channels []string `json:"channels"`
}

type publishEventRequest struct {
	Event       string              `json:"event"`
	UserID      string              `json:"user_id"`
	ScheduledAt *time.Time          `json:"scheduled_at"`
	Data        map[string]any      `json:"data"`
	Options     publishEventOptions `json:"options"`
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

	bus, err := queue.NewNATSBus(cfg.NATSURL, logger)
	if err != nil {
		logger.Error("failed to create nats bus", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = bus.Close()
	}()

	realtimePublisher := realtime.NewPublisher(bus, logger)
	realtimeHandler := server.NewRealtimeHandler(bus, logger)

	eventService := events.NewService(db, logger)
	templateRepo := template.NewRepository(db)
	adminTemplateHandler := server.NewAdminTemplateHandler(templateRepo, logger)
	encryptionKey, err := appcrypto.LoadKeyFromEnv()
	if err != nil {
		logger.Warn("failed to load provider encryption key", "error", err)
	}

	providerRepo := provider.NewRepository(db, encryptionKey)
	adminProviderHandler := server.NewAdminProviderHandler(providerRepo, logger)

	userRepo := user.NewRepository(db)
	userAuthHandler := server.NewUserAuthHandler(
		userRepo,
		cfg.JWTSecret,
		cfg.UserTokenTTL,
		logger,
	)

	inboxRepo := inbox.NewRepository(db)
	inboxHandler := server.NewInboxHandler(inboxRepo, realtimePublisher, logger)

	preferenceRepo := preference.NewRepository(db)
	consentRepo := consent.NewRepository(db)
	suppressionRepo := suppression.NewRepository(db)
	auditLogger := audit.NewLogger(db)

	preferenceHandler := server.NewPreferenceHandler(preferenceRepo, consentRepo, auditLogger, logger)
	quietHoursHandler := server.NewQuietHoursHandler(preferenceRepo, auditLogger, logger)
	suppressionHandler := server.NewSuppressionHandler(suppressionRepo, auditLogger, logger)
	adminNotificationHandler := server.NewAdminNotificationHandler(db, auditLogger, logger)
		consentHandler := server.NewConsentHandler(consentRepo, auditLogger, logger)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		server.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	})

	r.Handle("/metrics", promhttp.Handler())

	r.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(server.APIKeyAuth(db, logger))

			r.With(server.RequireScope(auth.ScopeEventsWrite)).Post("/events", handlePublishEvent(eventService, logger))
			r.With(server.RequireScope(auth.ScopeUsersWrite)).Post("/auth/user-token", userAuthHandler.CreateToken)

			r.Route("/admin/templates", func(r chi.Router) {
				r.With(server.RequireScope(auth.ScopeTemplatesWrite)).Post("/", adminTemplateHandler.CreateTemplate)
				r.With(server.RequireScope(auth.ScopeTemplatesRead)).Get("/", adminTemplateHandler.ListTemplates)
				r.With(server.RequireScope(auth.ScopeTemplatesRead)).Get("/{template_id}", adminTemplateHandler.GetTemplate)
				r.With(server.RequireScope(auth.ScopeTemplatesWrite)).Post("/{template_id}/versions", adminTemplateHandler.CreateVersion)
				r.With(server.RequireScope(auth.ScopeTemplatesWrite)).Post("/{template_id}/versions/{version_id}/publish", adminTemplateHandler.PublishVersion)
			})

			r.Route("/admin/providers", func(r chi.Router) {
				r.With(server.RequireScope(auth.ScopeProvidersRead)).Get("/", adminProviderHandler.ListSystemProviders)
			})

			r.Route("/admin/provider-configs", func(r chi.Router) {
				r.With(server.RequireScope(auth.ScopeProvidersWrite)).Post("/", adminProviderHandler.CreateConfig)
				r.With(server.RequireScope(auth.ScopeProvidersRead)).Get("/", adminProviderHandler.ListConfigs)
				r.With(server.RequireScope(auth.ScopeProvidersWrite)).Post("/{configID}/enable", adminProviderHandler.EnableConfig)
				r.With(server.RequireScope(auth.ScopeProvidersWrite)).Post("/{configID}/disable", adminProviderHandler.DisableConfig)
			})

			r.Route("/admin/notifications", func(r chi.Router) {
				r.With(server.RequireScope(auth.ScopeNotificationsRead)).Get("/", adminNotificationHandler.ListDeadLetter)
				r.With(server.RequireScope(auth.ScopeNotificationsRead)).Get("/{id}", adminNotificationHandler.GetNotification)
				r.With(server.RequireScope(auth.ScopeNotificationsRead)).Get("/{id}/attempts", adminNotificationHandler.GetAttempts)
				r.With(server.RequireScope(auth.ScopeNotificationsWrite)).Post("/{id}/retry", adminNotificationHandler.Retry)
			})

			r.Route("/admin/suppressions", func(r chi.Router) {
				r.With(server.RequireScope(auth.ScopeSuppressionsRead)).Get("/", suppressionHandler.List)
				r.With(server.RequireScope(auth.ScopeSuppressionsWrite)).Post("/", suppressionHandler.Create)
				r.With(server.RequireScope(auth.ScopeSuppressionsWrite)).Delete("/{id}", suppressionHandler.Delete)
			})
		})

		r.Route("/inbox", func(r chi.Router) {
			r.Use(server.UserAuth(cfg.JWTSecret, logger))

			r.Get("/", inboxHandler.List)
			r.Patch("/{itemID}/read", inboxHandler.MarkRead)
			r.Patch("/{itemID}/unread", inboxHandler.MarkUnread)
			r.Post("/{itemID}/archive", inboxHandler.Archive)
			r.Post("/{itemID}/unarchive", inboxHandler.Unarchive)
			r.Post("/{itemID}/pin", inboxHandler.Pin)
			r.Post("/{itemID}/unpin", inboxHandler.Unpin)
		})

		r.Group(func(r chi.Router) {
			r.Use(server.UserAuth(cfg.JWTSecret, logger))

			r.Get("/preferences", preferenceHandler.Get)
			r.Patch("/preferences", preferenceHandler.Update)

			r.Get("/preferences/quiet-hours", quietHoursHandler.Get)
			r.Patch("/preferences/quiet-hours", quietHoursHandler.Update)

			r.Route("/consent", func(r chi.Router) {
				r.Use(server.UserAuth(cfg.JWTSecret, logger))

				r.Get("/", consentHandler.Get)
				r.Get("/status", consentHandler.GetStatus)
				r.Post("/", consentHandler.Record)
			})
		})

		r.Route("/realtime", func(r chi.Router) {
			r.Use(server.UserAuth(cfg.JWTSecret, logger))

			r.Get("/stream", realtimeHandler.Stream)
		})
	})

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("api server starting", "addr", cfg.HTTPAddr)

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown api server", "error", err)
	}
}

func handlePublishEvent(
	eventService *events.Service,
	logger *slog.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		var req publishEventRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			server.WriteError(
				w,
				http.StatusBadRequest,
				"invalid_request",
				"invalid request body",
			)
			return
		}

		if req.Event == "" {
			server.WriteError(
				w,
				http.StatusBadRequest,
				"validation_failed",
				"event is required",
			)
			return
		}

		tenantID, ok := tenant.FromContext(r.Context())
		if !ok {
			server.WriteError(
				w,
				http.StatusInternalServerError,
				"missing_tenant",
				"tenant not found in request context",
			)
			return
		}

		apiKey, _ := auth.APIKeyFromContext(r.Context())

		idempotencyKey := r.Header.Get("Idempotency-Key")

		resp, err := eventService.Publish(
			r.Context(),
			events.PublishRequest{
				TenantID:       tenantID,
				EventType:      req.Event,
				UserID:         req.UserID,
				IdempotencyKey: idempotencyKey,
				ScheduledAt:    req.ScheduledAt,
				Data:           req.Data,
				Channels:       req.Options.Channels,
				Metadata: map[string]any{
					"source":      "api",
					"api_key_id":  apiKey.ID,
					"environment": apiKey.Environment,
					"idempotency": idempotencyKey,
				},
			},
		)
		if err != nil {
			logger.ErrorContext(
				r.Context(),
				"failed to publish event",
				"tenant_id", tenantID,
				"event_type", req.Event,
				"error", err,
			)

			server.WriteError(
				w,
				http.StatusInternalServerError,
				"publish_failed",
				"failed to publish event",
			)
			return
		}

		server.WriteJSON(w, http.StatusAccepted, resp)
	}
}
