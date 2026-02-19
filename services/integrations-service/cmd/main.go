package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/integrations-service/internal/config"
	"github.com/flyfleet/integrations-service/internal/database"
	"github.com/flyfleet/integrations-service/internal/handlers"
	"github.com/flyfleet/integrations-service/internal/repository"
	"github.com/flyfleet/integrations-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg := config.Load()

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	repo := repository.NewRepository(db, logger)
	svc := service.NewIntegrationsService(repo, cfg.JWTSecret, logger)
	h := handlers.NewHandlers(svc)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1/integrations", func(r chi.Router) {
		r.Use(OrgScopeMiddleware)
		r.Get("/", h.ListIntegrations)
		r.Post("/", h.CreateIntegration)
		r.Get("/logs", h.ListLogs)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				// Get individual integration logic
			})
			r.Put("/", h.UpdateIntegration)
			r.Post("/test", h.TestConnection)
		})
	})

	r.Route("/api/v1/api-clients", func(r chi.Router) {
		r.Use(OrgScopeMiddleware)
		r.Get("/", h.ListAPIClients)
		r.Post("/", h.CreateAPIClient)
		// ... other clients routes
	})

	r.Route("/api/v1/webhooks", func(r chi.Router) {
		r.Use(OrgScopeMiddleware)
		r.Get("/", h.ListWebhooks)
		r.Post("/", h.CreateWebhook)
		r.Delete("/{id}", h.DeleteWebhook)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Info("Starting integrations-service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	logger.Info("Server exiting")
}

func OrgScopeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgIDStr := r.Header.Get("X-Organization-ID")
		if orgIDStr == "" {
			orgIDStr = r.URL.Query().Get("orgId")
		}

		if orgIDStr == "" {
			http.Error(w, "MISSING_ORG_ID", http.StatusBadRequest)
			return
		}

		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			http.Error(w, "INVALID_ORG_ID", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "org_id", orgID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
