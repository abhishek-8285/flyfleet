package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/flyfleet/governance-service/internal/handlers"
	"github.com/flyfleet/governance-service/internal/repository"
	"github.com/flyfleet/governance-service/internal/service"
)

func main() {
	// Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// DB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = ""
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Repositories, Services, Handlers
	repo := repository.NewGovernanceRepository(pool)
	svc := service.NewGovernanceService(repo, logger)
	h := handlers.NewGovernanceHandler(svc)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1/governance", func(r chi.Router) {
		r.Route("/requests", func(r chi.Router) {
			r.Post("/", h.CreateRequest)
			r.Get("/", h.ListRequests)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetRequest)
				r.Post("/approve", h.ApproveRequest)
				r.Post("/reject", h.RejectRequest)
			})
		})

		r.Get("/settings", h.GetSettings)
		r.Put("/settings", h.UpdateSettings)
	})

	// Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
