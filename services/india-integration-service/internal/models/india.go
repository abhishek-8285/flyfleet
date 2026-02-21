package models

import (
	"time"

	"github.com/google/uuid"
)

// FASTagTransaction represents a FASTag transaction
type FASTagTransaction struct {
	ID              string    `json:"id"`
	TagNumber       string    `json:"tag_number"`
	VehicleID       uuid.UUID `json:"vehicle_id"`
	DriverID        uuid.UUID `json:"driver_id"`
	OrganizationID  uuid.UUID `json:"organization_id"`
	TransactionID   string    `json:"transaction_id"`
	TransactionType string    `json:"transaction_type"` // debit, credit
	Amount          float64   `json:"amount"`
	TollPlaza       string    `json:"toll_plaza"`
	TollPlazaID     string    `json:"toll_plaza_id"`
	LaneID          string    `json:"lane_id"`
	TransactionTime time.Time `json:"transaction_time"`
	TransactionDate string    `json:"transaction_date"` // Format: YYYY-MM-DD
	SettlementDate  *string   `json:"settlement_date,omitempty"` // Format: YYYY-MM-DD
	Status          string    `json:"status"`          // success, failed, pending
	ErrorCode       *string   `json:"error_code,omitempty"`
	ErrorDescription *string  `json:"error_description,omitempty"`
	Reconciled      bool      `json:"reconciled"`
	ReconciliationDate *time.Time `json:"reconciliation_date,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       uuid.UUID `json:"created_by"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedBy       uuid.UUID `json:"updated_by"`
}

// EWayBill represents an E-Way Bill
type EWayBill struct {
	ID                string    `json:"id"`
	EWayBillNumber    string    `json:"eway_bill_number"`
	EWayBillDate      time.Time `json:"eway_bill_date"`
	EWayBillValidTill time.Time `json:"eway_bill_valid_till"`
	DocumentNumber    string    `json:"document_number"`
	DocumentDate      time.Time `json:"document_date"`
	DocumentType      string    `json:"document_type"` // INV, CHL, WBILL, etc.
	GoodsDescription  string    `json:"goods_description"`
	GoodsValue        float64   `json:"goods_value"`
	GoodsTaxRate      float64   `json:"goods_tax_rate"`
	GoodsUnit         string    `json:"goods_unit"`
	GoodsQuantity     float64   `json:"goods_quantity"`
	FromPlace         string    `json:"from_place"`
	FromPincode       string    `json:"from_pincode"`
	FromStateCode     int       `json:"from_state_code"`
	ToPlace           string    `json:"to_place"`
	ToPincode         string    `json:"to_pincode"`
	ToStateCode       int       `json:"to_state_code"`
	TransporterID     string    `json:"transporter_id"`
	TransporterName   string    `json:"transporter_name"`
	TransportMode     string    `json:"transport_mode"` // road, rail, air, ship
	TransportDocumentNumber string `json:"transport_document_number"`
	TransportDocumentDate string `json:"transport_document_date"`
	TransitType       string    `json:"transit_type"` // own, hired, railways
	VehicleNumber     string    `json:"vehicle_number"`
	VehicleType       string    `json:"vehicle_type"` // reg, lr, cn
	RemainingDistance float64   `json:"remaining_distance"`
	UserGSTIN         string    `json:"user_gstin"`
	RecipientGSTIN    string    `json:"recipient_gstin"`
	GeneratedBy       string    `json:"generated_by"`
	GeneratedAt       time.Time `json:"generated_at"`
	Cancelled         bool      `json:"cancelled"`
	CancelledBy       *string   `json:"cancelled_by,omitempty"`
	CancelledOn       *time.Time `json:"cancelled_on,omitempty"`
	CancelledReason   *string   `json:"cancelled_reason,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         uuid.UUID `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         uuid.UUID `json:"updated_by"`
}

// VehicleVerification represents Vahan vehicle verification response
type VehicleVerification struct {
	ID               string    `json:"id"`
	VehicleID        uuid.UUID `json:"vehicle_id"`
	OrganizationID   uuid.UUID `json:"organization_id"`
	RegistrationNumber string   `json:"registration_number"`
	Status           string    `json:"status"`           // verified, pending, failed
	VerificationDate time.Time `json:"verification_date"`
	ChassisNumber    string    `json:"chassis_number"`
	EngineNumber     string    `json:"engine_number"`
	Model            string    `json:"model"`
	Manufacturer     string    `json:"manufacturer"`
	FuelType         string    `json:"fuel_type"`
	BodyType         string    `json:"body_type"`
	CubicCapacity    int       `json:"cubic_capacity"`
	SeatingCapacity  int       `json:"seating_capacity"`
	Color            string    `json:"color"`
	ManufacturingYear int      `json:"manufacturing_year"`
	OwnerName        string    `json:"owner_name"`
	OwnerAddress     string    `json:"owner_address"`
	OwnerCity        string    `json:"owner_city"`
	OwnerDistrict    string    `json:"owner_district"`
	OwnerState       string    `json:"owner_state"`
	OwnerPIN         string    `json:"owner_pin"`
	RCStatus         string    `json:"rc_status"`        // active, inactive, suspended
	InsuranceCompany string    `json:"insurance_company"`
	InsurancePolicy  string    `json:"insurance_policy"`
	InsuranceExpiry  time.Time `json:"insurance_expiry"`
	PUCNumber        string    `json:"puc_number"`
	PUCExpiry        time.Time `json:"puc_expiry"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        uuid.UUID `json:"created_by"`
	UpdatedAt        time.Time `json:"updated_at"`
	UpdatedBy        uuid.UUID `json:"updated_by"`
}

// DriverVerification represents Sarathi driver license verification
type DriverVerification struct {
	ID              string    `json:"id"`
	DriverID        uuid.UUID `json:"driver_id"`
	OrganizationID  uuid.UUID `json:"organization_id"`
	LicenseNumber   string    `json:"license_number"`
	Status          string    `json:"status"`           // verified, pending, failed
	VerificationDate time.Time `json:"verification_date"`
	Name            string    `json:"name"`
	FatherName      string    `json:"father_name"`
	DateOfBirth     time.Time `json:"date_of_birth"`
	Gender          string    `json:"gender"`
	Address         string    `json:"address"`
	City            string    `json:"city"`
	District        string    `json:"district"`
	State           string    `json:"state"`
	PIN             string    `json:"pin"`
	LicenseClass    []string  `json:"license_class"`     // MCWG, LMV, etc.
	IssueDate       time.Time `json:"issue_date"`
	ValidityFrom    time.Time `json:"validity_from"`
	ValidityTo      time.Time `json:"validity_to"`
	LastTransaction string    `json:"last_transaction"`   // current status
	EnrollmentNumber string   `json:"enrollment_number"`
	CreatedAt       time.Time `json:"created_at"`
	CreatedBy       uuid.UUID `json:"created_by"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedBy       uuid.UUID `json:"updated_by"`
}

// LocalPermit represents local permits and authorizations
type LocalPermit struct {
	ID                string    `json:"id"`
	VehicleID         uuid.UUID `json:"vehicle_id"`
	DriverID          uuid.UUID `json:"driver_id"`
	OrganizationID    uuid.UUID `json:"organization_id"`
	PermitNumber      string    `json:"permit_number"`
	PermitType        string    `json:"permit_type"`      // stage_carriage, contract_carriage, goods_vehicle
	PermitCategory    string    `json:"permit_category"`  // ordinary, tourist, educational
	PermitSubCategory string    `json:"permit_sub_category"` // public, private
	ValidFrom         time.Time `json:"valid_from"`
	ValidTo           time.Time `json:"valid_to"`
	IssuingAuthority  string    `json:"issuing_authority"`
	Jurisdiction      string    `json:"jurisdiction"`
	RouteDetails      string    `json:"route_details"`
	NumberOfTrips     int       `json:"number_of_trips"`   // for stage carriage
	MaxLoadCapacity   float64   `json:"max_load_capacity"` // for goods vehicles
	NumberOfSeats     int       `json:"number_of_seats"`   // for passenger vehicles
	Fare              float64   `json:"fare"`              // applicable fare
	Status            string    `json:"status"`            // active, expired, cancelled, suspended
	Remarks           string    `json:"remarks"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         uuid.UUID `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         uuid.UUID `json:"updated_by"`
}

// GSTCompliance represents GST compliance status
type GSTCompliance struct {
	ID               string    `json:"id"`
	OrganizationID   uuid.UUID `json:"organization_id"`
	GSTIN            string    `json:"gstin"`
	Status           string    `json:"status"`           // active, cancelled, suspended
	LegalName        string    `json:"legal_name"`
	TradeName        string    `json:"trade_name"`
	Address          string    `json:"address"`
	CentreJurisdiction string   `json:"centre_jurisdiction"`
	StateJurisdiction string   `json:"state_jurisdiction"`
	DateOfRegistration time.Time `json:"date_of_registration"`
	DateOfCancellation *time.Time `json:"date_of_cancellation,omitempty"`
	LastUpdated      time.Time `json:"last_updated"`
	ReturnFilingFrequency string `json:"return_filing_frequency"` // monthly, quarterly, annually
	ReturnFilingStatus string   `json:"return_filing_status"`    // compliant, non_compliant
	PaymentStatus    string    `json:"payment_status"`          // compliant, non_compliant
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        uuid.UUID `json:"created_by"`
	UpdatedAt        time.Time `json:"updated_at"`
	UpdatedBy        uuid.UUID `json:"updated_by"`
}

// IntegrationConfig represents configuration for India integrations
type IntegrationConfig struct {
	ID                string    `json:"id"`
	OrganizationID    uuid.UUID `json:"organization_id"`
	ServiceType       string    `json:"service_type"`      // fastag, ewaybill, vahan, sarathi, gst
	APIEndpoint       string    `json:"api_endpoint"`
	APIKey            string    `json:"api_key"`
	Username          string    `json:"username"`
	Password          string    `json:"password"`
	ClientID          string    `json:"client_id"`
	ClientSecret      string    `json:"client_secret"`
	Environment       string    `json:"environment"`       // production, staging, sandbox
	Active            bool      `json:"active"`
	LastSyncAt        *time.Time `json:"last_sync_at,omitempty"`
	NextSyncAt        *time.Time `json:"next_sync_at,omitempty"`
	SyncInterval      int       `json:"sync_interval"`     // in minutes
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         uuid.UUID `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         uuid.UUID `json:"updated_by"`
}