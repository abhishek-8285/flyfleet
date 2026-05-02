package main

import (
	"context"
	"log"
	"os"

	"github.com/flyfleet/fleet-core/internal/config"
	"github.com/flyfleet/fleet-core/internal/database"
	"go.uber.org/zap"
)

func main() {
	os.Setenv("CONFIG_PATH", "config.toml")
	cfg := config.Load()
	
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Info("Starting manual migration run", zap.String("url", cfg.DatabaseURL))

	if err := database.RunMigrations(cfg.DatabaseURL, logger); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	logger.Info("Migrations completed successfully")
}
