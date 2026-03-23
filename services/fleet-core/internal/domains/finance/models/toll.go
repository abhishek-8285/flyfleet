package models

import (
	"time"

	"github.com/google/uuid"
)

type TollTransaction struct {
	ID                    uuid.UUID      `json:"id" db:"id"`
	VehicleID             uuid.UUID      `json:"vehicle_id" db:"vehicle_id"`
	FASTagID              string         `json:"fastag_id" db:"fastag_id"`
	TollPlaza             string         `json:"toll_plaza" db:"toll_plaza"`
	Amount                float64        `json:"amount" db:"amount"`
	Currency              string         `json:"currency" db:"currency"`
	TransactionTime       time.Time      `json:"transaction_time" db:"transaction_time"`
	ExternalTransactionID string         `json:"external_transaction_id" db:"external_transaction_id"`
	Metadata              map[string]any `json:"metadata,omitempty" db:"metadata"`
	CreatedAt             time.Time      `json:"created_at" db:"created_at"`
}

type IngestTollRequest struct {
	VehicleID             uuid.UUID `json:"vehicle_id" validate:"required,uuid"`
	FASTagID              string    `json:"fastag_id" validate:"required"`
	TollPlaza             string    `json:"toll_plaza" validate:"required"`
	Amount                float64   `json:"amount" validate:"required,gt=0"`
	TransactionTime       time.Time `json:"transaction_time" validate:"required"`
	ExternalTransactionID string    `json:"external_transaction_id"`
}
