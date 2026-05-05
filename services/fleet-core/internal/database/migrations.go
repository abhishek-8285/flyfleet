package database

import (
	"go.uber.org/zap"
)

// RunMigrations applies all pending SQL migrations from the migrations directory.
func RunMigrations(dsn string, logger *zap.Logger) error {
	logger.Info("Legacy database migrations disabled")
	return nil
}
