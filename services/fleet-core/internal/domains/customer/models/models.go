package models

import (
	"time"

	"github.com/google/uuid"
)

// Customer Accounts
type Customer struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	CompanyName    string    `json:"company_name" db:"company_name"`
	GSTNumber      *string   `json:"gst_number" db:"gst_number"`
	BillingAddress []byte    `json:"billing_address" db:"billing_address"`
	ContactEmail   *string   `json:"contact_email" db:"contact_email"`
	ContactPhone   *string   `json:"contact_phone" db:"contact_phone"`
	Status         string    `json:"status" db:"status"` // ACTIVE, SUSPENDED, BLOCKED
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

type CustomerContact struct {
	ID         uuid.UUID `json:"id" db:"id"`
	CustomerID uuid.UUID `json:"customer_id" db:"customer_id"`
	Name       string    `json:"name" db:"name"`
	Email      *string   `json:"email" db:"email"`
	Phone      *string   `json:"phone" db:"phone"`
	Role       *string   `json:"role" db:"role"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type CustomerAddress struct {
	ID         uuid.UUID `json:"id" db:"id"`
	CustomerID uuid.UUID `json:"customer_id" db:"customer_id"`
	AddressType *string   `json:"address_type" db:"address_type"` // BILLING, SHIPPING, WAREHOUSE
	Address    []byte    `json:"address" db:"address"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type CustomerContract struct {
	ID             uuid.UUID `json:"id" db:"id"`
	CustomerID     uuid.UUID `json:"customer_id" db:"customer_id"`
	ContractNumber *string   `json:"contract_number" db:"contract_number"`
	ValidFrom      time.Time `json:"valid_from" db:"valid_from"`
	ValidTo        *time.Time `json:"valid_to" db:"valid_to"`
	Terms          []byte    `json:"terms" db:"terms"`
	Status         string    `json:"status" db:"status"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type CustomerCreditLimit struct {
	ID              uuid.UUID `json:"id" db:"id"`
	CustomerID      uuid.UUID `json:"customer_id" db:"customer_id"`
	CreditLimit     float64   `json:"credit_limit" db:"credit_limit"`
	AvailableCredit float64   `json:"available_credit" db:"available_credit"`
	Currency        string    `json:"currency" db:"currency"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Customer Users
type CustomerUser struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CustomerID   uuid.UUID `json:"customer_id" db:"customer_id"`
	Email        string    `json:"email" db:"email"`
	Phone        *string   `json:"phone" db:"phone"`
	Role         string    `json:"role" db:"role"` // ADMIN, OPERATIONS, FINANCE, VIEW_ONLY
	PasswordHash string    `json:"-" db:"password_hash"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type CustomerSession struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash string    `json:"token_hash" db:"token_hash"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IPAddress *string   `json:"ip_address" db:"ip_address"`
	UserAgent *string   `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CustomerAPIKey struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	CustomerID  uuid.UUID  `json:"customer_id" db:"customer_id"`
	KeyName     string     `json:"key_name" db:"key_name"`
	APIKeyHash  string     `json:"api_key_hash" db:"api_key_hash"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// Booking Requests
type BookingRequest struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	CustomerID          uuid.UUID `json:"customer_id" db:"customer_id"`
	PickupLocation      []byte    `json:"pickup_location" db:"pickup_location"`
	DropLocation        []byte    `json:"drop_location" db:"drop_location"`
	CargoType           *string   `json:"cargo_type" db:"cargo_type"`
	Weight              *float64  `json:"weight" db:"weight"`
	Volume              *float64  `json:"volume" db:"volume"`
	VehicleType         *string   `json:"vehicle_type" db:"vehicle_type"`
	RequestedPickupTime *time.Time `json:"requested_pickup_time" db:"requested_pickup_time"`
	Status              string    `json:"status" db:"status"` // REQUESTED, QUOTED, CONFIRMED, REJECTED
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

type BookingQuote struct {
	ID          uuid.UUID `json:"id" db:"id"`
	BookingID   uuid.UUID `json:"booking_id" db:"booking_id"`
	QuoteAmount float64   `json:"quote_amount" db:"quote_amount"`
	Currency    string    `json:"currency" db:"currency"`
	ValidUntil  *time.Time `json:"valid_until" db:"valid_until"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type BookingValidation struct {
	ID             uuid.UUID `json:"id" db:"id"`
	BookingID      uuid.UUID `json:"booking_id" db:"booking_id"`
	ValidationType string    `json:"validation_type" db:"validation_type"`
	Status         string    `json:"status" db:"status"` // PASSED, FAILED
	Details        *string   `json:"details" db:"details"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// Shipment
type Shipment struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CustomerID  uuid.UUID `json:"customer_id" db:"customer_id"`
	TripID      *uuid.UUID `json:"trip_id" db:"trip_id"`
	BookingID   *uuid.UUID `json:"booking_id" db:"booking_id"`
	Status      string    `json:"status" db:"status"` // BOOKED, CONFIRMED, ASSIGNED, IN_TRANSIT, AT_CHECKPOINT, DELIVERED, CANCELLED
	Origin      []byte    `json:"origin" db:"origin"`
	Destination []byte    `json:"destination" db:"destination"`
	DistanceKm        *float64   `json:"distance_km" db:"distance_km"`
	EwayBillNumber    *string    `json:"eway_bill_number" db:"eway_bill_number"`
	EwayBillExpiry    *time.Time `json:"eway_bill_expiry" db:"eway_bill_expiry"`
	EwayBillDocument  *string    `json:"eway_bill_document" db:"eway_bill_document"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

type ShipmentStop struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ShipmentID     uuid.UUID `json:"shipment_id" db:"shipment_id"`
	StopType       string    `json:"stop_type" db:"stop_type"` // PICKUP, DROPOFF, CHECKPOINT
	Sequence       int       `json:"sequence" db:"sequence"`
	Location       []byte    `json:"location" db:"location"`
	ExpectedArrival *time.Time `json:"expected_arrival" db:"expected_arrival"`
	ActualArrival  *time.Time `json:"actual_arrival" db:"actual_arrival"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type ShipmentEvent struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ShipmentID uuid.UUID `json:"shipment_id" db:"shipment_id"`
	EventType  string    `json:"event_type" db:"event_type"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	Location   []byte    `json:"location" db:"location"`
	Metadata   []byte    `json:"metadata" db:"metadata"`
}

type ShipmentStatusLog struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	ShipmentID uuid.UUID  `json:"shipment_id" db:"shipment_id"`
	OldStatus  *string    `json:"old_status" db:"old_status"`
	NewStatus  string     `json:"new_status" db:"new_status"`
	ChangedBy  *uuid.UUID `json:"changed_by" db:"changed_by"`
	ChangedAt  time.Time  `json:"changed_at" db:"changed_at"`
	Reason     *string    `json:"reason" db:"reason"`
}

// Delivery Proof
type DeliveryProof struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ShipmentID    uuid.UUID `json:"shipment_id" db:"shipment_id"`
	ReceiverName  string    `json:"receiver_name" db:"receiver_name"`
	ReceiverPhone *string   `json:"receiver_phone" db:"receiver_phone"`
	SignatureUrl  *string   `json:"signature_url" db:"signature_url"`
	PhotoUrl      *string   `json:"photo_url" db:"photo_url"`
	DeliveredAt   *time.Time `json:"delivered_at" db:"delivered_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type PODImage struct {
	ID              uuid.UUID `json:"id" db:"id"`
	DeliveryProofID uuid.UUID `json:"delivery_proof_id" db:"delivery_proof_id"`
	PhotoUrl        string    `json:"photo_url" db:"photo_url"`
	UploadedAt      time.Time `json:"uploaded_at" db:"uploaded_at"`
}

type PODSignature struct {
	ID              uuid.UUID `json:"id" db:"id"`
	DeliveryProofID uuid.UUID `json:"delivery_proof_id" db:"delivery_proof_id"`
	SignatureUrl    string    `json:"signature_url" db:"signature_url"`
	SignedBy        string    `json:"signed_by" db:"signed_by"`
	SignedAt        time.Time `json:"signed_at" db:"signed_at"`
}

// Shipment Documents
type ShipmentDocument struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ShipmentID   uuid.UUID `json:"shipment_id" db:"shipment_id"`
	DocumentType string    `json:"document_type" db:"document_type"` // LR_COPY, INVOICE, DELIVERY_NOTE, EWAY_BILL, POD
	FileUrl      string    `json:"file_url" db:"file_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type ShipmentDocumentVersion struct {
	ID            uuid.UUID `json:"id" db:"id"`
	DocumentID    uuid.UUID `json:"document_id" db:"document_id"`
	VersionNumber int       `json:"version_number" db:"version_number"`
	FileUrl       string    `json:"file_url" db:"file_url"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Pricing Engine
type PricingContract struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	CustomerID   uuid.UUID  `json:"customer_id" db:"customer_id"`
	ContractName string     `json:"contract_name" db:"contract_name"`
	ValidFrom    time.Time  `json:"valid_from" db:"valid_from"`
	ValidTo      *time.Time `json:"valid_to" db:"valid_to"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

type PricingRule struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ContractID   uuid.UUID `json:"contract_id" db:"contract_id"`
	PricingModel string    `json:"pricing_model" db:"pricing_model"` // PER_KM, PER_TRIP, PER_TON, ZONE_BASED, CONTRACT_RATE
	BaseRate     float64   `json:"base_rate" db:"base_rate"`
	Currency     string    `json:"currency" db:"currency"`
	Conditions   []byte    `json:"conditions" db:"conditions"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type PricingComponent struct {
	ID            uuid.UUID `json:"id" db:"id"`
	RuleID        uuid.UUID `json:"rule_id" db:"rule_id"`
	ComponentType string    `json:"component_type" db:"component_type"` // SURCHARGE, DISCOUNT, TAX
	Value         float64   `json:"value" db:"value"`
	IsPercentage  bool      `json:"is_percentage" db:"is_percentage"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type PricingQuote struct {
	ID          uuid.UUID `json:"id" db:"id"`
	BookingID   uuid.UUID `json:"booking_id" db:"booking_id"`
	TotalAmount float64   `json:"total_amount" db:"total_amount"`
	Currency    string    `json:"currency" db:"currency"`
	Breakdown   []byte    `json:"breakdown" db:"breakdown"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Credit System
type CustomerWallet struct {
	ID         uuid.UUID `json:"id" db:"id"`
	CustomerID uuid.UUID `json:"customer_id" db:"customer_id"`
	Balance    float64   `json:"balance" db:"balance"`
	Currency   string    `json:"currency" db:"currency"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type CreditTransaction struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	WalletID        uuid.UUID  `json:"wallet_id" db:"wallet_id"`
	TransactionType string     `json:"transaction_type" db:"transaction_type"` // CREDIT, DEBIT
	Amount          float64    `json:"amount" db:"amount"`
	Currency        string     `json:"currency" db:"currency"`
	ReferenceID     *uuid.UUID `json:"reference_id" db:"reference_id"`
	Description     *string    `json:"description" db:"description"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

type CreditLimit struct {
	ID         uuid.UUID `json:"id" db:"id"`
	CustomerID uuid.UUID `json:"customer_id" db:"customer_id"`
	MaxLimit   float64   `json:"max_limit" db:"max_limit"`
	Currency   string    `json:"currency" db:"currency"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Invoice Access
type CustomerInvoiceView struct {
	ID           uuid.UUID `json:"id" db:"id"`
	InvoiceID    uuid.UUID `json:"invoice_id" db:"invoice_id"`
	CustomerID   uuid.UUID `json:"customer_id" db:"customer_id"`
	DownloadedAt time.Time `json:"downloaded_at" db:"downloaded_at"`
}

// SLA Monitoring
type ShipmentSLARule struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CustomerID  uuid.UUID `json:"customer_id" db:"customer_id"`
	RuleName    string    `json:"rule_name" db:"rule_name"`
	Metric      string    `json:"metric" db:"metric"`
	TargetValue float64   `json:"target_value" db:"target_value"`
	Unit        string    `json:"unit" db:"unit"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type ShipmentSLAViolation struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ShipmentID  uuid.UUID `json:"shipment_id" db:"shipment_id"`
	RuleID      uuid.UUID `json:"rule_id" db:"rule_id"`
	ActualValue float64   `json:"actual_value" db:"actual_value"`
	ViolatedAt  time.Time `json:"violated_at" db:"violated_at"`
}

// Audit Logs
type CustomerAPILog struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	APIKeyID       *uuid.UUID `json:"api_key_id" db:"api_key_id"`
	Endpoint       string     `json:"endpoint" db:"endpoint"`
	Method         string     `json:"method" db:"method"`
	StatusCode     *int       `json:"status_code" db:"status_code"`
	ResponseTimeMs *int       `json:"response_time_ms" db:"response_time_ms"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type CustomerAuditLog struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	CustomerID uuid.UUID  `json:"customer_id" db:"customer_id"`
	UserID     *uuid.UUID `json:"user_id" db:"user_id"`
	EventType  string     `json:"event_type" db:"event_type"` // booking_created, tracking_viewed, invoice_downloaded, shipment_cancelled
	ResourceID *uuid.UUID `json:"resource_id" db:"resource_id"`
	Metadata   []byte     `json:"metadata" db:"metadata"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// Tracking Models
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type ShipmentTrackingData struct {
	ShipmentID      uuid.UUID       `json:"shipment_id"`
	TrackingNumber  string          `json:"tracking_number"`
	Status          string          `json:"status"`
	CurrentLocation *Location       `json:"current_location"`
	SpeedKmH        float64         `json:"speed_kmh"`
	Events          []ShipmentEvent `json:"events"`
}

// Phase 3 Extensions
type BookingStatusHistory struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	BookingID      uuid.UUID  `json:"booking_id" db:"booking_id"`
	PreviousStatus *string    `json:"previous_status" db:"previous_status"`
	NewStatus      string     `json:"new_status" db:"new_status"`
	ChangedBy      *uuid.UUID `json:"changed_by" db:"changed_by"`
	ChangedAt      time.Time  `json:"changed_at" db:"changed_at"`
	Reason         *string    `json:"reason" db:"reason"`
}

type BookingApproval struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	BookingID      uuid.UUID  `json:"booking_id" db:"booking_id"`
	ApproverUserID uuid.UUID  `json:"approver_user_id" db:"approver_user_id"`
	ApprovalStatus string     `json:"approval_status" db:"approval_status"`
	ApprovedAt     *time.Time `json:"approved_at" db:"approved_at"`
	Comments       *string    `json:"comments" db:"comments"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type BookingShipment struct {
	ID             uuid.UUID `json:"id" db:"id"`
	BookingID      uuid.UUID `json:"booking_id" db:"booking_id"`
	ShipmentID     uuid.UUID `json:"shipment_id" db:"shipment_id"`
	SequenceNumber int       `json:"sequence_number" db:"sequence_number"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type Dispute struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	CustomerID  uuid.UUID  `json:"customer_id" db:"customer_id"`
	ShipmentID  *uuid.UUID `json:"shipment_id" db:"shipment_id"`
	InvoiceID   *uuid.UUID `json:"invoice_id" db:"invoice_id"`
	Status      string     `json:"status" db:"status"` // OPEN, INVESTIGATING, RESOLVED, CLOSED
	Reason      string     `json:"reason" db:"reason"`
	Description *string    `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ResolvedAt  *time.Time `json:"resolved_at" db:"resolved_at"`
}

type DisputeMessage struct {
	ID         uuid.UUID `json:"id" db:"id"`
	DisputeID  uuid.UUID `json:"dispute_id" db:"dispute_id"`
	SenderID   uuid.UUID `json:"sender_id" db:"sender_id"`
	SenderType string    `json:"sender_type" db:"sender_type"` // CUSTOMER, ADMIN
	Message    string    `json:"message" db:"message"`
	SentAt     time.Time `json:"sent_at" db:"sent_at"`
}

type DisputeAttachment struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	DisputeID  uuid.UUID  `json:"dispute_id" db:"dispute_id"`
	MessageID  *uuid.UUID `json:"message_id" db:"message_id"`
	FileUrl    string     `json:"file_url" db:"file_url"`
	UploadedAt time.Time  `json:"uploaded_at" db:"uploaded_at"`
}
