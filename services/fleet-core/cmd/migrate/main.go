package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/flyfleet/fleet-core/internal/config"
	"github.com/flyfleet/fleet-core/internal/database"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func main() {
	flag.Parse()
	command := flag.Arg(0)

	cfg := config.Load()
	logger, _ := zap.NewDevelopment()

	switch command {
	case "up":
		if err := database.RunMigrations(cfg.DatabaseURL, logger); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	case "down":
		fmt.Println("Down migration not supported for custom legacy tool. Please use SQL manually if needed.")
	default:
		fmt.Println("Usage: go run cmd/migrate/main.go [up|down]")
		os.Exit(1)
	}
}
