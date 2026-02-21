package main

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/flyfleet/reports-service/internal/handlers"
	"github.com/flyfleet/reports-service/internal/queue"
	"github.com/flyfleet/reports-service/internal/repository"
	"github.com/flyfleet/reports-service/internal/service"
	reportsgrpc "github.com/flyfleet/reports-service/internal/grpc"
	pb "github.com/flyfleet/proto/reports/v1"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	google_grpc "google.golang.org/grpc"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/flyfleet?sslmode=disable"
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		logger.Fatal("Unable to parse DB URL", zap.Error(err))
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.Fatal("Unable to connect to DB", zap.Error(err))
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	queueClient := queue.NewQueueClient(redisAddr)
	defer queueClient.Close()

	svc := service.NewReportService(repo, queueClient, logger)
	h := handlers.NewReportHandler(svc, logger)

	// Start gRPC Server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50060"
	}
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			logger.Fatal("Failed to listen for gRPC", zap.Error(err))
		}
		
		grpcServer := google_grpc.NewServer()
		rServer := reportsgrpc.NewReportsServer(svc, logger)
		pb.RegisterReportsServiceServer(grpcServer, rServer)
		
		logger.Info("Starting reports gRPC server", zap.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1/reports", func(r chi.Router) {
		h.RegisterRoutes(r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	logger.Info("Starting Reports Service", zap.String("port", port))
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
