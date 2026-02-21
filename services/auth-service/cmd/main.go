package main

import (
	"context"
	"fmt"
	"net"

	"github.com/fleet-platform/shared/pkg/database"
	"github.com/fleet-platform/shared/pkg/service"
	"github.com/flyfleet/auth-service/internal/grpc"
	"github.com/flyfleet/auth-service/internal/repository"
	authservice "github.com/flyfleet/auth-service/internal/service"
	pb "github.com/flyfleet/auth-service/proto/auth/v1"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	gs "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Initialize Base Service (Config, Logger, DB, Graceful Shutdown)
	base, err := service.NewBase("auth-service", service.Config{
		DatabaseURL: "",     // Will automatically pull from DATABASE_URL env
		LogLevel:    "info", // Will automatically pull from LOG_LEVEL env
		GRPCPort:    "50051",
		Port:        "8080",
	})
	if err != nil {
		fmt.Printf("Service failure: %v\n", err)
		return
	}

	// Wait, we need auth-specific configs (like JWT_SECRET).
	// We can bind it dynamically via viper on the base.Config
	base.Config.SetDefault("security.jwt_secret", "")
	base.Config.BindEnv("security.jwt_secret", "JWT_SECRET")

	// 2. Initialize Health/Ready Probes (exposes /health)
	base.ServeHealth(base.Config.GetString("server.port"))

	// 3. Initialize Domain Logic with pgxpool wrapped as sql.DB
	pg := base.DB.(*database.Postgres)
	sqlDB := stdlib.OpenDBFromPool(pg.GetPool())
	repo := repository.NewAuthRepository(sqlDB, base.Logger)
	jwtSecret := base.Config.GetString("security.jwt_secret")
	authSvc := authservice.NewAuthService(base.Logger, repo, jwtSecret)

	// 4. Register gRPC Handlers with Optional mTLS
	grpcServerOptions := []gs.ServerOption{}
	creds, err := base.GetTLSCredentials()
	if err != nil {
		base.Logger.Fatal("Failed to load TLS credentials", zap.Error(err))
	}
	if creds != nil {
		grpcServerOptions = append(grpcServerOptions, gs.Creds(creds))
		base.Logger.Info("gRPC Server configured with mTLS auth")
	}

	grpcPort := base.Config.GetString("server.grpc_port")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		base.Logger.Fatal("Failed to listen on gRPC port", zap.Error(err))
	}

	server := gs.NewServer(grpcServerOptions...)
	pb.RegisterAuthServiceServer(server, grpc.NewAuthServer(authSvc, base.Logger))
	reflection.Register(server)

	// 5. Define Runner
	grpcRunner := func(ctx context.Context) error {
		base.Logger.Info("Starting gRPC server", zap.String("port", grpcPort))
		go func() {
			<-ctx.Done()
			server.GracefulStop()
			base.Logger.Info("gRPC server stopped gracefully")
		}()
		return server.Serve(lis)
	}

	// 6. Run via ServiceBase Graceful context
	if err := base.RunWithContext(context.Background(), grpcRunner); err != nil {
		base.Logger.Fatal("Service exited with error", zap.Error(err))
	}
}
