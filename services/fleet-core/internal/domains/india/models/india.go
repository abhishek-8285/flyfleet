package models

import (
	"time"

	"github.com/google/uuid"
)

// FASTagTransaction represents a FASTag transaction
type FASTagTransaction struct {
	ID                 string     `json:"id" db:"id" validate:"required"`
	TagNumber          string     `json:"tag_number" db:"tag_number" validate:"required"`
	VehicleID          uuid.UUID  `json:"vehicle_id" db:"vehicle_id" validate:"required,uuid"`
	DriverID           uuid.UUID  `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id" validate:"required,uuid"`
	TransactionID      string     `json:"transaction_id" db:"transaction_id" validate:"required"`
	TransactionType    string     `json:"transaction_type" db:"transaction_type" validate:"required,oneof=debit credit"`
	Amount             float64    `json:"amount" db:"amount" validate:"required,gt=0"`
	TollPlaza          string     `json:"toll_plaza" db:"toll_plaza" validate:"required"`
	TollPlazaID        string     `json:"toll_plaza_id" db:"toll_plaza_id" validate:"required"`
	LaneID             string     `json:"lane_id" db:"lane_id" validate:"required"`
	TransactionTime    time.Time  `json:"transaction_time" db:"transaction_time" validate:"required"`
	TransactionDate    string     `json:"transaction_date" db:"transaction_date" validate:"required"`
	SettlementDate     *string    `json:"settlement_date,omitempty" db:"settlement_date"`
	Status             string     `json:"status" db:"status" validate:"required,oneof=success failed pending"`
	ErrorCode          *string    `json:"error_code,omitempty" db:"error_code"`
	ErrorDescription   *string    `json:"error_description,omitempty" db:"error_description"`
	Reconciled         bool       `json:"reconciled" db:"reconciled"`
	ReconciliationDate *time.Time `json:"reconciliation_date,omitempty" db:"reconciliation_date"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	CreatedBy          uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy          uuid.UUID  `json:"updated_by" db:"updated_by"`
}

// EWayBill represents an E-Way Bill
type EWayBill struct {
	ID                      string     `json:"id" db:"id" validate:"required"`
	TenantID          uuid.UUID  `json:"tenant_id" db:"tenant_id" validate:"required,uuid"`
	EWayBillNumber          string     `json:"eway_bill_number" db:"eway_bill_number" validate:"required"`
	EWayBillDate            time.Time  `json:"eway_bill_date" db:"eway_bill_date" validate:"required"`
	EWayBillValidTill       time.Time  `json:"eway_bill_valid_till" db:"eway_bill_valid_till" validate:"required"`
	DocumentNumber          string     `json:"document_number" db:"document_number" validate:"required"`
	DocumentDate            time.Time  `json:"document_date" db:"document_date" validate:"required"`
	DocumentType            string     `json:"document_type" db:"document_type" validate:"required"`
	GoodsDescription        string     `json:"goods_description" db:"goods_description" validate:"required"`
	GoodsValue              float64    `json:"goods_value" db:"goods_value" validate:"required,gt=0"`
	GoodsTaxRate            float64    `json:"goods_tax_rate" db:"goods_tax_rate" validate:"gte=0"`
	GoodsUnit               string     `json:"goods_unit" db:"goods_unit" validate:"required"`
	GoodsQuantity           float64    `json:"goods_quantity" db:"goods_quantity" validate:"required,gt=0"`
	FromPlace               string     `json:"from_place" db:"from_place" validate:"required"`
	FromPincode             string     `json:"from_pincode" db:"from_pincode" validate:"required"`
	FromStateCode           int        `json:"from_state_code" db:"from_state_code" validate:"required,min=1,max=37"`
	ToPlace                 string     `json:"to_place" db:"to_place" validate:"required"`
	ToPincode               string     `json:"to_pincode" db:"to_pincode" validate:"required"`
	ToStateCode             int        `json:"to_state_code" db:"to_state_code" validate:"required,min=1,max=37"`
	TransporterID           string     `json:"transporter_id" db:"transporter_id"`
	TransporterName         string     `json:"transporter_name" db:"transporter_name"`
	TransportMode           string     `json:"transport_mode" db:"transport_mode" validate:"omitempty,oneof=road rail air ship"`
	TransportDocumentNumber string     `json:"transport_document_number" db:"transport_document_number"`
	TransportDocumentDate   string     `json:"transport_document_date" db:"transport_document_date"`
	TransitType             string     `json:"transit_type" db:"transit_type" validate:"omitempty,oneof=own hired railways"`
	VehicleNumber           string     `json:"vehicle_number" db:"vehicle_number"`
	VehicleType             string     `json:"vehicle_type" db:"vehicle_type" validate:"omitempty,oneof=reg lr cn"`
	RemainingDistance       float64    `json:"remaining_distance" db:"remaining_distance"`
	UserGSTIN               string     `json:"user_gstin" db:"user_gstin" validate:"required"`
	RecipientGSTIN          string     `json:"recipient_gstin" db:"recipient_gstin"`
	GeneratedBy             string     `json:"generated_by" db:"generated_by" validate:"required"`
	GeneratedAt             time.Time  `json:"generated_at" db:"generated_at"`
	Cancelled               bool       `json:"cancelled" db:"cancelled"`
	CancelledBy             *string    `json:"cancelled_by,omitempty" db:"cancelled_by"`
	CancelledOn             *time.Time `json:"cancelled_on,omitempty" db:"cancelled_on"`
	CancelledReason         *string    `json:"cancelled_reason,omitempty" db:"cancelled_reason"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	CreatedBy               uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy               uuid.UUID  `json:"updated_by" db:"updated_by"`
}

// VehicleVerification represents Vahan vehicle verification response
type VehicleVerification struct {
	ID                 string    `json:"id" db:"id" validate:"required"`
	VehicleID          uuid.UUID `json:"vehicle_id" db:"vehicle_id" validate:"required,uuid"`
	TenantID     uuid.UUID `json:"tenant_id" db:"tenant_id" validate:"required,uuid"`
	RegistrationNumber string    `json:"registration_number" db:"registration_number" validate:"required"`
	Status             string    `json:"status" db:"status" validate:"required,oneof=verified pending failed"`
	VerificationDate   time.Time `json:"verification_date" db:"verification_date"`
	ChassisNumber      string    `json:"chassis_number" db:"chassis_number"`
	EngineNumber       string    `json:"engine_number" db:"engine_number"`
	Model              string    `json:"model" db:"model"`
	Manufacturer       string    `json:"manufacturer" db:"manufacturer"`
	FuelType           string    `json:"fuel_type" db:"fuel_type"`
	BodyType           string    `json:"body_type" db:"body_type"`
	CubicCapacity      int       `json:"cubic_capacity" db:"cubic_capacity"`
	SeatingCapacity    int       `json:"seating_capacity" db:"seating_capacity"`
	Color              string    `json:"color" db:"color"`
	ManufacturingYear  int       `json:"manufacturing_year" db:"manufacturing_year"`
	OwnerName          string    `json:"owner_name" db:"owner_name"`
	OwnerAddress       string    `json:"owner_address" db:"owner_address"`
	OwnerCity          string    `json:"owner_city" db:"owner_city"`
	OwnerDistrict      string    `json:"owner_district" db:"owner_district"`
	OwnerState         string    `json:"owner_state" db:"owner_state"`
	OwnerPIN           string    `json:"owner_pin" db:"owner_pin"`
	RCStatus           string    `json:"rc_status" db:"rc_status" validate:"omitempty,oneof=active inactive suspended"`
	InsuranceCompany   string    `json:"insurance_company" db:"insurance_company"`
	InsurancePolicy    string    `json:"insurance_policy" db:"insurance_policy"`
	InsuranceExpiry    time.Time `json:"insurance_expiry" db:"insurance_expiry"`
	PUCNumber          string    `json:"puc_number" db:"puc_number"`
	PUCExpiry          time.Time `json:"puc_expiry" db:"puc_expiry"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	CreatedBy          uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy          uuid.UUID `json:"updated_by" db:"updated_by"`
}

// DriverVerification represents Sarathi driver license verification
type DriverVerification struct {
	ID               string    `json:"id" db:"id" validate:"required"`
	DriverID         uuid.UUID `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	TenantID   uuid.UUID `json:"tenant_id" db:"tenant_id" validate:"required,uuid"`
	LicenseNumber    string    `json:"license_number" db:"license_number" validate:"required"`
	Status           string    `json:"status" db:"status" validate:"required,oneof=verified pending failed"`
	VerificationDate time.Time `json:"verification_date" db:"verification_date"`
	Name             string    `json:"name" db:"name" validate:"required"`
	FatherName       string    `json:"father_name" db:"father_name"`
	DateOfBirth      time.Time `json:"date_of_birth" db:"date_of_birth" validate:"required"`
	Gender           string    `json:"gender" db:"gender"`
	Address          string    `json:"address" db:"address"`
	City             string    `json:"city" db:"city"`
	District         string    `json:"district" db:"district"`
	State            string    `json:"state" db:"state"`
	PIN              string    `json:"pin" db:"pin"`
	LicenseClass     []string  `json:"license_class" db:"license_class"`
	IssueDate        time.Time `json:"issue_date" db:"issue_date" validate:"required"`
	ValidityFrom     time.Time `json:"validity_from" db:"validity_from" validate:"required"`
	ValidityTo       time.Time `json:"validity_to" db:"validity_to" validate:"required"`
	LastTransaction  string    `json:"last_transaction" db:"last_transaction"`
	EnrollmentNumber string    `json:"enrollment_number" db:"enrollment_number"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	CreatedBy        uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy        uuid.UUID `json:"updated_by" db:"updated_by"`
}

// LocalPermit represents local permits and authorizations
type LocalPermit struct {
	ID                string    `json:"id" db:"id" validate:"required"`
	VehicleID         uuid.UUID `json:"vehicle_id" db:"vehicle_id" validate:"required,uuid"`
	DriverID          uuid.UUID `json:"driver_id" db:"driver_id" validate:"required,uuid"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id" validate:"required,uuid"`
	PermitNumber      string    `json:"permit_number" db:"permit_number" validate:"required"`
	PermitType        string    `json:"permit_type" db:"permit_type" validate:"required,oneof=stage_carriage contract_carriage goods_vehicle"`
	PermitCategory    string    `json:"permit_category" db:"permit_category"`
	PermitSubCategory string    `json:"permit_sub_category" db:"permit_sub_category"`
	ValidFrom         time.Time `json:"valid_from" db:"valid_from" validate:"required"`
	ValidTo           time.Time `json:"valid_to" db:"valid_to" validate:"required"`
	IssuingAuthority  string    `json:"issuing_authority" db:"issuing_authority" validate:"required"`
	Jurisdiction      string    `json:"jurisdiction" db:"jurisdiction"`
	RouteDetails      string    `json:"route_details" db:"route_details"`
	NumberOfTrips     int       `json:"number_of_trips" db:"number_of_trips"`
	MaxLoadCapacity   float64   `json:"max_load_capacity" db:"max_load_capacity"`
	NumberOfSeats     int       `json:"number_of_seats" db:"number_of_seats"`
	Fare              float64   `json:"fare" db:"fare"`
	Status            string    `json:"status" db:"status" validate:"required,oneof=active expired cancelled suspended"`
	Remarks           string    `json:"remarks" db:"remarks"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	CreatedBy         uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	UpdatedBy         uuid.UUID `json:"updated_by" db:"updated_by"`
}

// GSTCompliance represents GST compliance status
type GSTCompliance struct {
	ID                    string     `json:"id" db:"id" validate:"required"`
	TenantID        uuid.UUID  `json:"tenant_id" db:"tenant_id" validate:"required,uuid"`
	GSTIN                 string     `json:"gstin" db:"gstin" validate:"required"`
	Status                string     `json:"status" db:"status" validate:"required,oneof=active cancelled suspended"`
	LegalName             string     `json:"legal_name" db:"legal_name" validate:"required"`
	TradeName             string     `json:"trade_name" db:"trade_name"`
	Address               string     `json:"address" db:"address" validate:"required"`
	CentreJurisdiction    string     `json:"centre_jurisdiction" db:"centre_jurisdiction"`
	StateJurisdiction     string     `json:"state_jurisdiction" db:"state_jurisdiction"`
	DateOfRegistration    time.Time  `json:"date_of_registration" db:"date_of_registration" validate:"required"`
	DateOfCancellation    *time.Time `json:"date_of_cancellation,omitempty" db:"date_of_cancellation"`
	LastUpdated           time.Time  `json:"last_updated" db:"last_updated"`
	ReturnFilingFrequency string     `json:"return_filing_frequency" db:"return_filing_frequency"`
	ReturnFilingStatus    string     `json:"return_filing_status" db:"return_filing_status"`
	PaymentStatus         string     `json:"payment_status" db:"payment_status"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	CreatedBy             uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy             uuid.UUID  `json:"updated_by" db:"updated_by"`
}

// IntegrationConfig represents configuration for India integrations
type IntegrationConfig struct {
	ID             string     `json:"id" db:"id" validate:"required"`
	TenantID uuid.UUID  `json:"tenant_id" db:"tenant_id" validate:"required,uuid"`
	ServiceType    string     `json:"service_type" db:"service_type" validate:"required,oneof=fastag ewaybill vahan sarathi gst permit"`
	APIEndpoint    string     `json:"api_endpoint" db:"api_endpoint" validate:"required,url"`
	APIKey         string     `json:"-" db:"api_key"`       // SECURITY FIX: Do not expose in JSON responses
	Username       string     `json:"-" db:"username"`      // SECURITY FIX: Exclude from JSON
	Password       string     `json:"-" db:"password"`      // SECURITY FIX: Do not expose in JSON responses
	ClientID       string     `json:"-" db:"client_id"`     // SECURITY FIX: Exclude from JSON
	ClientSecret   string     `json:"-" db:"client_secret"` // SECURITY FIX: Do not expose in JSON responses
	Environment    string     `json:"environment" db:"environment" validate:"required,oneof=production staging sandbox"`
	Active         bool       `json:"active" db:"active"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty" db:"last_sync_at"`
	NextSyncAt     *time.Time `json:"next_sync_at,omitempty" db:"next_sync_at"`
	SyncInterval   int        `json:"sync_interval" db:"sync_interval"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	CreatedBy      uuid.UUID  `json:"created_by" db:"created_by"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	UpdatedBy      uuid.UUID  `json:"updated_by" db:"updated_by"`
}

// IntegrationLog represents an audit log for India integrations
type IntegrationLog struct {
	ID              string      `json:"id" db:"id"`
	ServiceType     string      `json:"service_type" db:"service_type"`
	RequestPayload  interface{} `json:"request_payload" db:"request_payload"`
	ResponsePayload interface{} `json:"response_payload" db:"response_payload"`
	Status          int         `json:"status" db:"status"`
	ErrorMessage    string      `json:"error_message" db:"error_message"`
	TenantID  *uuid.UUID  `json:"tenant_id,omitempty" db:"tenant_id"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
}

// CreateFASTagTransactionRequest represents a request to create a FASTag transaction
type CreateFASTagTransactionRequest struct {
	TagNumber       string    `json:"tag_number" validate:"required"`
	VehicleID       uuid.UUID `json:"vehicle_id" validate:"required,uuid"`
	DriverID        uuid.UUID `json:"driver_id" validate:"required,uuid"`
	TransactionID   string    `json:"transaction_id" validate:"required"`
	TransactionType string    `json:"transaction_type" validate:"required,oneof=debit credit"`
	Amount          float64   `json:"amount" validate:"required,gt=0"`
	TollPlaza       string    `json:"toll_plaza" validate:"required"`
	TollPlazaID     string    `json:"toll_plaza_id" validate:"required"`
	LaneID          string    `json:"lane_id" validate:"required"`
	TransactionTime time.Time `json:"transaction_time" validate:"required"`
	Status          string    `json:"status" validate:"required,oneof=success failed pending"`
}

// CreateEWayBillRequest represents a request to create an E-Way Bill
type CreateEWayBillRequest struct {
	EWayBillNumber   string    `json:"eway_bill_number" validate:"required"`
	DocumentNumber   string    `json:"document_number" validate:"required"`
	DocumentDate     time.Time `json:"document_date" validate:"required"`
	DocumentType     string    `json:"document_type" validate:"required"`
	GoodsDescription string    `json:"goods_description" validate:"required"`
	GoodsValue       float64   `json:"goods_value" validate:"required,gt=0"`
	GoodsQuantity    float64   `json:"goods_quantity" validate:"required,gt=0"`
	FromPlace        string    `json:"from_place" validate:"required"`
	FromPincode      string    `json:"from_pincode" validate:"required"`
	FromStateCode    int       `json:"from_state_code" validate:"required,min=1,max=37"`
	ToPlace          string    `json:"to_place" validate:"required"`
	ToPincode        string    `json:"to_pincode" validate:"required"`
	ToStateCode      int       `json:"to_state_code" validate:"required,min=1,max=37"`
	VehicleNumber    string    `json:"vehicle_number"`
	UserGSTIN        string    `json:"user_gstin" validate:"required"`
}

// VerifyVehicleRequest represents a request to verify a vehicle
type VerifyVehicleRequest struct {
	TenantID     uuid.UUID `json:"tenant_id" validate:"required,uuid"`
	VehicleID          uuid.UUID `json:"vehicle_id" validate:"required,uuid"`
	RegistrationNumber string    `json:"registration_number" validate:"required"`
}

// VerifyDriverRequest represents a request to verify a driver
type VerifyDriverRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required,uuid"`
	DriverID       uuid.UUID `json:"driver_id" validate:"required,uuid"`
	LicenseNumber  string    `json:"license_number" validate:"required"`
	DateOfBirth    string    `json:"date_of_birth" validate:"required"` // Format: YYYY-MM-DD
}
