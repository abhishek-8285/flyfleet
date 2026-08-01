package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/config"
	"github.com/gonotify/gonotify/internal/events"
	"github.com/gonotify/gonotify/internal/observability"
	"github.com/gonotify/gonotify/internal/queue"
)

func main() {
	cfg := config.Load()
	logger := observability.NewLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	bus, err := queue.NewNATSBus(cfg.NATSURL, logger)
	if err != nil {
		logger.Error("failed to create nats bus", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = bus.Close()
	}()

	relay := events.NewOutboxRelay(
		db,
		bus,
		logger,
		1*time.Second,
	)

	go func() {
		if err := relay.Run(ctx); err != nil {
			logger.Error("outbox relay failed", "error", err)
			stop()
		}
	}()

	err = bus.Subscribe(
		"gn.events.received",
		"workflow-worker",
		func(ctx context.Context, msg queue.Message) error {
			logger.InfoContext(
				ctx,
				"received event",
				"message_id", msg.ID,
				"tenant_id", msg.TenantID,
				"type", msg.Type,
				"event_type", msg.Metadata["event_type"],
				"user_id", msg.Metadata["user_id"],
				"payload", string(msg.Payload),
			)

			// Phase 1:
			// - match workflow
			// - create notification
			// - render template
			// - dispatch to provider

			return nil
		},
	)
	if err != nil {
		logger.Error("failed to subscribe to events", "error", err)
		os.Exit(1)
	}

	<-ctx.Done()
	logger.Info("worker stopped")
}
