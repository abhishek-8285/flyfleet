package models

import (
	"time"

	"github.com/google/uuid"
)

type FuelCardProvider string

const (
	FuelCardProviderIOCL FuelCardProvider = "IOCL"
	FuelCardProviderHPCL FuelCardProvider = "HPCL"
	FuelCardProviderBPCL FuelCardProvider = "BPCL"
)

type FuelCardTransaction struct {
	ID                    uuid.UUID        `json:"id" db:"id"`
	VehicleID             *uuid.UUID       `json:"vehicle_id,omitempty" db:"vehicle_id"`
	CardID                string           `json:"card_id" db:"card_id"`
	Provider              FuelCardProvider `json:"provider" db:"provider"`
	Station               *string          `json:"station,omitempty" db:"station"`
	Liters                float64          `json:"liters" db:"liters"`
	Amount                float64          `json:"amount" db:"amount"`
	Currency              string           `json:"currency" db:"currency"`
	TransactionTime       time.Time        `json:"transaction_time" db:"transaction_time"`
	SyncedAt              time.Time        `json:"synced_at" db:"synced_at"`
	ExternalTransactionID string           `json:"external_transaction_id" db:"external_transaction_id"`
	Metadata              map[string]any   `json:"metadata,omitempty" db:"metadata"`
	CreatedAt             time.Time        `json:"created_at" db:"created_at"`
}

type SyncFuelCardRequest struct {
	Provider FuelCardProvider `json:"provider" validate:"required"`
	CardID   string           `json:"card_id" validate:"required"`
}
