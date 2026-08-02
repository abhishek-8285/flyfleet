package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/flyfleet/services/dispatch"
	"github.com/flyfleet/services/dispatch/server"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	mlURL := os.Getenv("MLPLATFORM_URL")
	if mlURL == "" {
		mlURL = "http://localhost:8080"
	}

	fleetCoreURL := os.Getenv("FLEET_CORE_URL")
	if fleetCoreURL == "" {
		fleetCoreURL = "http://localhost:8082"
	}

	httpAddr := os.Getenv("DISPATCH_HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8081"
	}

	grpcAddr := os.Getenv("DISPATCH_GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":9091"
	}

	emitter := dispatch.NewEventEmitter(mlURL)
	store := dispatch.NewInMemoryStore()
	dispatcher := dispatch.NewDispatcher(mlURL, fleetCoreURL, emitter, store)

	log.Info().Str("mlplatform_url", mlURL).Str("fleetcore_url", fleetCoreURL).Msg("dispatch ranker initialized")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	httpServer := server.NewHTTPServer(dispatcher, httpAddr)
	go func() {
		log.Info().Str("addr", httpAddr).Msg("dispatch HTTP server starting")
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("HTTP server failed")
			stop()
		}
	}()

	grpcServer := server.NewGRPCServer(dispatcher)
	go func() {
		log.Info().Str("addr", grpcAddr).Msg("dispatch gRPC server starting")
		if err := grpcServer.Start(grpcAddr); err != nil {
			log.Error().Err(err).Msg("gRPC server failed")
			stop()
		}
	}()

	<-ctx.Done()
	log.Info().Msg("dispatch service shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Stop(); err != nil {
		log.Error().Err(err).Msg("HTTP server shutdown failed")
	}
	grpcServer.Stop()

	<-shutdownCtx.Done()
}