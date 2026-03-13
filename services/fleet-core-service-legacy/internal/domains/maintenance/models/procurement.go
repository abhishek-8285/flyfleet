package models

import (
	"time"

	"github.com/google/uuid"
)

type POStatus string

const (
	POStatusDraft    POStatus = "DRAFT"
	POStatusApproved POStatus = "APPROVED"
	POStatusOrdered  POStatus = "ORDERED"
	POStatusReceived POStatus = "RECEIVED"
	POStatusClosed   POStatus = "CLOSED"
)

type PurchaseOrder struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	VendorID       *uuid.UUID `json:"vendor_id,omitempty" db:"vendor_id"`
	OrderNumber    string    `json:"order_number" db:"order_number"`
	Status         POStatus  `json:"status" db:"status"`
	TotalCost      float64   `json:"total_cost" db:"total_cost"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`

	Items []PurchaseOrderItem `json:"items,omitempty"`
}

type PurchaseOrderItem struct {
	ID              uuid.UUID `json:"id" db:"id"`
	PurchaseOrderID uuid.UUID `json:"purchase_order_id" db:"purchase_order_id"`
	PartID          *uuid.UUID `json:"part_id,omitempty" db:"part_id"`
	Quantity        int       `json:"quantity" db:"quantity"`
	UnitCost        float64   `json:"unit_cost" db:"unit_cost"`
	TotalCost       float64   `json:"total_cost" db:"total_cost"`
}

type CreatePURequest struct {
	VendorID uuid.UUID `json:"vendor_id" validate:"required"`
	Items    []struct {
		PartID   uuid.UUID `json:"part_id" validate:"required"`
		Quantity int       `json:"quantity" validate:"required,min=1"`
		UnitCost float64   `json:"unit_cost" validate:"required,min=0"`
	} `json:"items" validate:"required,dive"`
}
