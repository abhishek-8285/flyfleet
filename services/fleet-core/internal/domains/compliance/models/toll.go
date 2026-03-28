package models

import (
	"time"

	"github.com/google/uuid"
)

type TollTransaction struct {
	ID        uuid.UUID `json:"id" db:"id"`
	VehicleID uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	FastagID  string    `json:"fastag_id" db:"fastag_id"`
	TollPlaza string    `json:"toll_plaza" db:"toll_plaza"`
	Amount    float64   `json:"amount" db:"amount"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Location  string    `json:"location" db:"location"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type ReconciliationStatus string

const (
	ReconMatched  ReconciliationStatus = "MATCHED"
	ReconMismatch ReconciliationStatus = "MISMATCH"
	ReconMissing  ReconciliationStatus = "MISSING"
	ReconPending  ReconciliationStatus = "PENDING"
)

type FASTagReconciliation struct {
	ID             uuid.UUID            `json:"id" db:"id"`
	TransactionID  uuid.UUID            `json:"transaction_id" db:"transaction_id"`
	VehicleID      uuid.UUID            `json:"vehicle_id" db:"vehicle_id"`
	ExpectedAmount float64              `json:"expected_amount" db:"expected_amount"`
	ActualAmount   float64              `json:"actual_amount" db:"actual_amount"`
	Status         ReconciliationStatus `json:"status" db:"status"`
	CreatedAt      time.Time            `json:"created_at" db:"created_at"`
}
