package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// IndiaIntegrationRepositoryImpl implements the IndiaIntegrationRepository interface
type IndiaIntegrationRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewIndiaIntegrationRepository creates a new instance of IndiaIntegrationRepositoryImpl
func NewIndiaIntegrationRepository(db *pgxpool.Pool) *IndiaIntegrationRepositoryImpl {
	return &IndiaIntegrationRepositoryImpl{
		db: db,
	}
}

// Implementation methods are split across multiple files for maintainability:
// - fastag.go
// - eway_bill.go
// - vehicle_verification.go
// - driver_verification.go
// - local_permit.go
// - gst_compliance.go
// - integration_config.go
// - health.go
// - integration_log.go
// - bulk_operations.go
// - compliance_scan.go
