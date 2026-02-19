package models

import (
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/driver/models"
)

// Customer represents a B2B client profile
type Customer struct {
	ID            string                 `json:"id" db:"id"`
	OrgID         string                 `json:"org_id" db:"org_id"`
	Name          string                 `json:"name" db:"name"`
	Email         string                 `json:"email" db:"email"`
	Phone         string                 `json:"phone,omitempty" db:"phone"`
	APIKey        string                 `json:"-" db:"api_key"`           // Never expose API Key in responses
	APISecretHash string                 `json:"-" db:"api_secret_hash"` // Never expose Secret
	Address       *models.Address        `json:"address,omitempty" db:"address"`
	BillingInfo   map[string]interface{} `json:"billing_info,omitempty" db:"billing_info"`
	Status        string                 `json:"status" db:"status"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// Shipment represents a consignment linked to a trip
type Shipment struct {
	ID                string                 `json:"id" db:"id"`
	CustomerID        string                 `json:"customer_id" db:"customer_id"`
	TripID            *string                `json:"trip_id,omitempty" db:"trip_id"`
	TrackingNumber    string                 `json:"tracking_number" db:"tracking_number"`
	Status            string                 `json:"status" db:"status"`
	Origin            *models.Location       `json:"origin" db:"origin"`
	Destination       *models.Location       `json:"destination" db:"destination"`
	CurrentLocation   *models.Location       `json:"current_location,omitempty" db:"current_location"`
	EstimatedDelivery *time.Time             `json:"estimated_delivery,omitempty" db:"estimated_delivery"`
	ActualDelivery    *time.Time             `json:"actual_delivery,omitempty" db:"actual_delivery"`
	CargoDetails      map[string]interface{} `json:"cargo_details" db:"cargo_details"`
	PODURL            string                 `json:"pod_url,omitempty" db:"pod_url"`
	PODUploadedAt     *time.Time             `json:"pod_uploaded_at,omitempty" db:"pod_uploaded_at"`
	Metadata          map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
	
	// Join fields
	CustomerName      string                 `json:"customer_name,omitempty" db:"-"`
}

// Invoice represents a billing record for a shipment
type Invoice struct {
	ID            string                 `json:"id" db:"id"`
	CustomerID    string                 `json:"customer_id" db:"customer_id"`
	ShipmentID    *string                `json:"shipment_id,omitempty" db:"shipment_id"`
	InvoiceNumber string                 `json:"invoice_number" db:"invoice_number"`
	Status        string                 `json:"status" db:"status"`
	Amount        float64                `json:"amount" db:"amount"`
	Currency      string                 `json:"currency" db:"currency"`
	IssueDate     time.Time              `json:"issue_date" db:"issue_date"`
	DueDate       time.Time              `json:"due_date" db:"due_date"`
	PaidDate      *time.Time             `json:"paid_date,omitempty" db:"paid_date"`
	PDFURL        string                 `json:"pdf_url,omitempty" db:"pdf_url"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// BookingRequest represents a customer transport request
type BookingRequest struct {
	ID                  string                 `json:"id" db:"id"`
	CustomerID          string                 `json:"customer_id" db:"customer_id"`
	RequestType         string                 `json:"request_type" db:"request_type"`
	Origin              *models.Location       `json:"origin" db:"origin"`
	Destination         *models.Location       `json:"destination" db:"destination"`
	RequestedPickupTime *time.Time             `json:"requested_pickup_time,omitempty" db:"requested_pickup_time"`
	CargoDetails        map[string]interface{} `json:"cargo_details" db:"cargo_details"`
	Status              string                 `json:"status" db:"status"`
	Notes               string                 `json:"notes,omitempty" db:"notes"`
	Metadata            map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

// ShipmentListFilter defines filters for listing shipments
type ShipmentListFilter struct {
	CustomerID string
	Status     string
	FromDate   *time.Time
	ToDate     *time.Time
	Page       int
	PageSize   int
}

// BookingRequestCreate represents the payload for creating a booking
type BookingRequestCreate struct {
	RequestType         string                 `json:"request_type" validate:"required"`
	Origin              *models.Location       `json:"origin" validate:"required"`
	Destination         *models.Location       `json:"destination" validate:"required"`
	RequestedPickupTime *time.Time             `json:"requested_pickup_time"`
	CargoDetails        map[string]interface{} `json:"cargo_details"`
	Notes               string                 `json:"notes"`
}
