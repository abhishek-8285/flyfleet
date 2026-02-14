package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flyfleet/settings-service/internal/config"
	"github.com/flyfleet/settings-service/internal/database"
	"github.com/flyfleet/settings-service/internal/handlers"
	"github.com/flyfleet/settings-service/internal/middleware"
	"github.com/flyfleet/settings-service/internal/repository"
	"github.com/flyfleet/settings-service/internal/service"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	// Load configuration
	cfg := config.Load()

	// Initialize DB connection
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repository
	repo := repository.NewRepository(db, logger)

	// Initialize service
	svc := service.NewSettingsService(repo, logger)

	// Initialize handlers
	h := handlers.NewSettingsHandlers(svc, logger)

	// Setup routes
	router := http.NewServeMux()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Platform Settings
	router.HandleFunc("/settings/platform", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetPlatformSettings(w, r)
		case http.MethodPut:
			h.UpdatePlatformSettings(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Feature Flags
	router.HandleFunc("/settings/flags", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListFeatureFlags(w, r)
		case http.MethodPost:
			h.CreateFeatureFlag(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	// Note: For simplified routing in ServeMux without wildcards in older Go versions, 
	// specific ID routes handle their parsing. 
	// If using Go 1.22+ we could use "POST /settings/flags/{id}/toggle"
	// Here we rely on the handler parsing the path as implemented.
	router.HandleFunc("/settings/flags/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/toggle") && r.Method == http.MethodPatch {
			h.ToggleFeatureFlag(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Localization
	router.HandleFunc("/settings/localization", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetLocalizationSettings(w, r)
		case http.MethodPut:
			h.UpdateLocalizationSettings(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Templates
	router.HandleFunc("/settings/templates", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListTemplates(w, r)
		case http.MethodPost: // Assuming we might want creation, though current UI only shows list/edit
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/settings/templates/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			h.UpdateTemplate(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Notification Channels
	router.HandleFunc("/settings/channels", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListNotificationChannels(w, r)
		case http.MethodPut:
			h.UpdateNotificationChannel(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Limits
	router.HandleFunc("/settings/limits", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetSystemLimits(w, r)
		case http.MethodPut:
			h.UpdateSystemLimits(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: middleware.UserContextMiddleware(router),
	}

	// Run server in goroutine
	go func() {
		logger.Info("Starting settings service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down settings service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Settings service exited")
}
