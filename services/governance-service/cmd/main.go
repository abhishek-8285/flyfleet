package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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
		dbURL = "postgres://fleet_user:fleet_password@localhost:5433/fleet_db?sslmode=disable"
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
	r := gin.Default()
	
	v1 := r.Group("/api/v1/governance")
	{
		v1.POST("/requests", h.CreateRequest)
		v1.GET("/requests", h.ListRequests)
		v1.GET("/requests/:id", h.GetRequest)
		v1.POST("/requests/:id/approve", h.ApproveRequest)
		v1.POST("/requests/:id/reject", h.RejectRequest)
		
		v1.GET("/settings", h.GetSettings)
		v1.PUT("/settings", h.UpdateSettings)
	}

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
