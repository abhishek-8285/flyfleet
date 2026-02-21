package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/flyfleet/fleet-core-service-new/internal/config"
	"github.com/flyfleet/fleet-core-service-new/internal/grpc"
	"github.com/flyfleet/fleet-core-service-new/internal/repository"
	"github.com/flyfleet/fleet-core-service-new/internal/service"
	pb "github.com/flyfleet/fleet-core-service-new/proto/fleet/v1"
	"net/http"

	"github.com/fleet-platform/shared/pkg/logger"
	"github.com/fleet-platform/shared/pkg/messaging"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	google_grpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()

	log, err := logger.New(logger.Config{
		Development: cfg.Environment != "production",
		Level:       zap.InfoLevel,
		OutputPaths: []string{"stdout"},
	})
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database", zap.Error(err))
	}

	js, err := messaging.NewJetStream(cfg.NatsURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS JetStream", zap.Error(err))
	}

	// Ensure FLEET stream exists for events
	subjects := []string{"fleet.vehicle.>", "fleet.driver.>", "fleet.maintenance.>", "fleet.fuel.>", "fleet.telemetry.>"}
	if err := js.CreateStream("FLEET", subjects); err != nil {
		log.Warn("Failed to ensure FLEET stream", zap.Error(err))
	}
	
	repo := repository.NewFleetRepository(db)
	svc := service.NewFleetService(log, repo, js)
	server := grpc.NewFleetServer(svc)

	grpcServer := google_grpc.NewServer()
	pb.RegisterFleetServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	go func() {
		log.Info("Starting Fleet gRPC server", zap.String("port", cfg.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// HTTP Health Check Server
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		log.Info("Starting Fleet health check server", zap.String("port", "8084"))
		if err := http.ListenAndServe(":8084", mux); err != nil {
			log.Error("Health server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down fleet-core-service (gRPC)...")
	grpcServer.GracefulStop()
	log.Info("Service exited safely")
}
