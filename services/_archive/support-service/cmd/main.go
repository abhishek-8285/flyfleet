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
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Middleware
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware(logger))

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "support-service",
		})
	})

	// API Routes
	api := r.Group("/api/v1/support")
	{
		// Dashboard APIs
		api.GET("/dashboard/summary", ticketHandlers.GetDashboardSummary)
		api.GET("/dashboard/recent-activity", ticketHandlers.GetRecentActivity)

		// Incident Management
		api.GET("/incidents", ticketHandlers.ListTickets)
		api.POST("/incidents", ticketHandlers.CreateTicket)
		api.GET("/incidents/:id", ticketHandlers.GetTicket)
		api.PATCH("/incidents/:id/status", ticketHandlers.UpdateTicketStatus)
		api.POST("/incidents/:id/assign", ticketHandlers.AssignTicket)
		api.POST("/incidents/:id/comment", ticketHandlers.AddComment)
		api.GET("/incidents/:id/comments", ticketHandlers.GetTicketComments)
	}

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

func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		status := c.Writer.Status()

		logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}