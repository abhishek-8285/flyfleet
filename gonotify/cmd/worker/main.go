package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gonotify/gonotify/internal/config"
	"github.com/gonotify/gonotify/internal/delivery"
	"github.com/gonotify/gonotify/internal/events"
	"github.com/gonotify/gonotify/internal/observability"
	smtpprovider "github.com/gonotify/gonotify/internal/provider/smtp"
	"github.com/gonotify/gonotify/internal/queue"
	"github.com/gonotify/gonotify/internal/template"
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

	emailProvider := smtpprovider.New(smtpprovider.Config{
		Host:      cfg.SMTPHost,
		Port:      cfg.SMTPPort,
		Username:  cfg.SMTPUsername,
		Password:  cfg.SMTPPassword,
		FromEmail: cfg.SMTPFromEmail,
		FromName:  cfg.SMTPFromName,
		UseTLS:    cfg.SMTPUseTLS,
	})

	templateRepository := template.NewRepository(db)

	dispatcher := delivery.NewDispatcher(
		db,
		emailProvider,
		templateRepository,
		logger,
	)

	err = bus.Subscribe(
		"gn.events.received",
		"dispatch-worker",
		dispatcher.HandleEvent,
	)
	if err != nil {
		logger.Error("failed to subscribe to events", "error", err)
		os.Exit(1)
	}

	<-ctx.Done()
	logger.Info("worker stopped")
}
