package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/support-service/internal/config"
	"github.com/flyfleet/support-service/internal/database"
	"github.com/flyfleet/support-service/internal/handlers"
	"github.com/flyfleet/support-service/internal/repository"
	"github.com/flyfleet/support-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
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
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Support Service", zap.String("port", cfg.Port))

	// Initialize database connection
	db, err := database.NewConnection(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close(db)

	// Initialize components
	ticketRepo := repository.NewTicketRepository(db, logger)
	ticketService := service.NewTicketService(ticketRepo, logger)
	ticketHandlers := handlers.NewTicketHandlers(ticketService, logger)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Duration(cfg.WriteTimeout) * time.Second))

	// CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(corsMiddleware.Handler)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "support-service"}`))
	})

	// API Routes
	r.Route("/api/v1/support", func(r chi.Router) {
		// Dashboard APIs
		r.Get("/dashboard/summary", ticketHandlers.GetDashboardSummary)
		r.Get("/dashboard/recent-activity", ticketHandlers.GetRecentActivity)

		// Incident Management
		r.Get("/incidents", ticketHandlers.ListTickets)
		r.Post("/incidents", ticketHandlers.CreateTicket)
		r.Get("/incidents/{id}", ticketHandlers.GetTicket)
		r.Patch("/incidents/{id}/status", ticketHandlers.UpdateTicketStatus)
		r.Post("/incidents/{id}/assign", ticketHandlers.AssignTicket)
		r.Post("/incidents/{id}/comment", ticketHandlers.AddComment)
		r.Get("/incidents/{id}/comments", ticketHandlers.GetTicketComments)
	})

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("Support Service starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	<-done
	logger.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown failed", zap.Error(err))
	}

	logger.Info("Server stopped")
}