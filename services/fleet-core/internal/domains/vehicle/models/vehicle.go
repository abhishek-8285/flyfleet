package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// VehicleStatus defines the lifecycle state of a fleet asset.
type VehicleStatus string

const (
	StatusActive         VehicleStatus = "ACTIVE"
	StatusMaintenance    VehicleStatus = "MAINTENANCE"
	StatusSuspended      VehicleStatus = "SUSPENDED"
	StatusDecommissioned VehicleStatus = "DECOMMISSIONED"
	StatusIdle           VehicleStatus = "IDLE"
	StatusProvisioning   VehicleStatus = "PROVISIONING"
	StatusRetired        VehicleStatus = "RETIRED"
)

func (s VehicleStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusMaintenance, StatusSuspended, StatusDecommissioned, StatusIdle, StatusProvisioning, StatusRetired:
		return true
	}
	return false
}

// VehicleType defines categorized vehicle classes.
type VehicleType string

const (
	TypeTruck     VehicleType = "TRUCK"
	TypeVan       VehicleType = "VAN"
	TypeHeavyHaul VehicleType = "HEAVY_HAUL"
	TypeBus       VehicleType = "BUS"
	TypeCar       VehicleType = "CAR"
	TypeBike      VehicleType = "BIKE"
)

func (t VehicleType) IsValid() bool {
	switch t {
	case TypeTruck, TypeVan, TypeHeavyHaul, TypeBus, TypeCar, TypeBike:
		return true
	}
	return false
}

// FuelType defines the primary propulsion source.
type FuelType string

const (
	FuelDiesel   FuelType = "DIESEL"
	FuelPetrol   FuelType = "PETROL"
	FuelCNG      FuelType = "CNG"
	FuelElectric FuelType = "EV"
	FuelHybrid   FuelType = "HYBRID"
)

func (f FuelType) IsValid() bool {
	switch f {
	case FuelDiesel, FuelPetrol, FuelCNG, FuelElectric, FuelHybrid:
		return true
	}
	return false
}

// OwnershipType defines asset acquisition models.
type OwnershipType string

const (
	OwnershipOwned   OwnershipType = "OWNED"
	OwnershipLeased  OwnershipType = "LEASED"
	OwnershipPartner OwnershipType = "PARTNER"
)

func (o OwnershipType) IsValid() bool {
	switch o {
	case OwnershipOwned, OwnershipLeased, OwnershipPartner:
		return true
	}
	return false
}

// GPSStatus defines real-time device connectivity.
type GPSStatus string

const (
	GPSOnline  GPSStatus = "ONLINE"
	GPSOffline GPSStatus = "OFFLINE"
	GPSStale   GPSStatus = "STALE"
)

// DeviceHealthStatus defines IoT hardware integrity.
type DeviceHealthStatus string

const (
	HealthHealthy      DeviceHealthStatus = "HEALTHY"
	HealthWarning      DeviceHealthStatus = "WARNING"
	HealthError        DeviceHealthStatus = "ERROR"
	HealthDisconnected DeviceHealthStatus = "DISCONNECTED"
)

// IgnitionStatus defines engine state reported by telemetry.
type IgnitionStatus string

const (
	IgnitionOn      IgnitionStatus = "ON"
	IgnitionOff     IgnitionStatus = "OFF"
	IgnitionUnknown IgnitionStatus = "UNKNOWN"
)

// MaintenanceStatus defines the technical readiness of the vehicle.
type MaintenanceStatus string

const (
	MaintHealthy MaintenanceStatus = "HEALTHY"
	MaintDue     MaintenanceStatus = "DUE"
	MaintOverdue MaintenanceStatus = "OVERDUE"
)

// DriverLicenseStatus defines the regulatory readiness of the assigned driver.
type DriverLicenseStatus string

const (
	LicenseValid   DriverLicenseStatus = "VALID"
	LicenseExpired DriverLicenseStatus = "EXPIRED"
	LicenseMissing DriverLicenseStatus = "MISSING"
)

// ComplianceState defines the regulatory health color code.
type ComplianceState string

const (
	ComplianceGreen  ComplianceState = "GREEN"
	ComplianceYellow ComplianceState = "YELLOW"
	ComplianceRed    ComplianceState = "RED"
)

// RiskCategory defines categorized safety/compliance risk levels.
type RiskCategory string

const (
	RiskLow    RiskCategory = "LOW"
	RiskMedium RiskCategory = "MEDIUM"
	RiskHigh   RiskCategory = "HIGH"
)

// NetworkType defines the data transmission medium.
type NetworkType string

const (
	NetCellular  NetworkType = "CELLULAR"
	NetWiFi      NetworkType = "WIFI"
	NetSatellite NetworkType = "SATELLITE"
	NetOffline   NetworkType = "OFFLINE"
)

// AuditAction defines tracked administrative operations.
type AuditAction string

const (
	ActionCreate         AuditAction = "CREATE"
	ActionUpdate         AuditAction = "UPDATE"
	ActionDelete         AuditAction = "DELETE"
	ActionStatusChange   AuditAction = "STATUS_CHANGE"
	ActionAssignDriver   AuditAction = "ASSIGN_DRIVER"
	ActionUnassignDriver AuditAction = "UNASSIGN_DRIVER"
	ActionBlacklist      AuditAction = "BLACKLIST"
	ActionDocumentUpload AuditAction = "DOCUMENT_UPLOAD"
	ActionDocumentVerify AuditAction = "DOCUMENT_VERIFY"
	ActionHardBlock      AuditAction = "HARD_BLOCK"
	ActionHardUnblock    AuditAction = "HARD_UNBLOCK"
)

// RBAC Permissions
const (
	PermFleetVehicleRead    = "fleet.vehicle.read"
	PermFleetVehicleWrite   = "fleet.vehicle.write"
	PermFleetVehicleBlock   = "fleet.vehicle.block"
	PermFleetVehicleUnblock = "fleet.vehicle.unblock"
	PermFleetVehicleAudit   = "fleet.vehicle.audit"
)

// OutboxEventType defines types of domain events for async processing.
type OutboxEventType string

const (
	OutboxVehicleCreated     OutboxEventType = "VEHICLE_CREATED"
	OutboxVehicleUpdated     OutboxEventType = "VEHICLE_UPDATED"
	OutboxComplianceBreach   OutboxEventType = "COMPLIANCE_BREACH"
	OutboxDriverAssigned     OutboxEventType = "DRIVER_ASSIGNED"
	OutboxBlacklistTriggered OutboxEventType = "BLACKLIST_TRIGGERED"
)

// VehicleAuditLog tracks changes to vehicle state for accountability.
type VehicleAuditLog struct {
	ID        uuid.UUID   `json:"id"         db:"id"`
	VehicleID uuid.UUID   `json:"vehicle_id" db:"vehicle_id"`
	TenantID  uuid.UUID   `json:"tenant_id"  db:"tenant_id"`
	UserID    uuid.UUID   `json:"user_id"    db:"user_id"`
	Action    AuditAction `json:"action"     db:"action"`
	FieldName string      `json:"field_name" db:"field_name"`
	OldValue  string      `json:"old_value"  db:"old_value"`
	NewValue  string      `json:"new_value"  db:"new_value"`
	Timestamp time.Time   `json:"timestamp"  db:"timestamp"`
}

// VehicleOutbox stores domain events for reliable async processing.
type VehicleOutbox struct {
	ID          uuid.UUID       `json:"id"          db:"id"`
	TenantID    uuid.UUID       `json:"tenant_id"   db:"tenant_id"`
	EventType   OutboxEventType `json:"event_type"  db:"event_type"`
	Payload     []byte          `json:"payload"     db:"payload"`
	CreatedAt   time.Time       `json:"created_at"  db:"created_at"`
	ProcessedAt *time.Time      `json:"processed_at" db:"processed_at"`
	ExpiresAt   time.Time       `json:"expires_at"  db:"expires_at"` // TTL Support
}

// Vehicle is the enterprise aggregate root for the vehicle domain.
type Vehicle struct {
	ID                 uuid.UUID     `json:"id"                   db:"id"`
	TenantID           uuid.UUID     `json:"tenant_id"            db:"tenant_id"`
	RegistrationNumber string        `json:"registration_number"  db:"registration_number"`
	ChassisNumber      *string       `json:"chassis_number"       db:"chassis_number"` // AES-256 encrypted
	EngineNumber       *string       `json:"engine_number"        db:"engine_number"`  // AES-256 encrypted
	VINNumber          *string       `json:"vin_number"           db:"vin_number"`     // AES-256 encrypted
	VehicleType        VehicleType   `json:"vehicle_type"         db:"vehicle_type"`
	CapacityTonnage    float64       `json:"capacity_tonnage"     db:"capacity_tonnage"`
	FuelType           FuelType      `json:"fuel_type"            db:"fuel_type"`
	FuelTankCapacity   float64       `json:"fuel_tank_capacity"   db:"fuel_tank_capacity"`
	OwnershipType      OwnershipType `json:"ownership_type"       db:"ownership_type"`
	ParentVehicleID    *uuid.UUID    `json:"parent_vehicle_id"    db:"parent_vehicle_id"`
	BranchID           *uuid.UUID    `json:"branch_id"            db:"branch_id"`
	CostCenterID       *uuid.UUID    `json:"cost_center_id"       db:"cost_center_id"`
	Status             VehicleStatus `json:"status"               db:"status"`
	GPSEnabled         bool          `json:"gps_enabled"          db:"gps_enabled"`
	GPSDeviceID        *string       `json:"gps_device_id"        db:"gps_device_id"`
	GPSDeviceType      *string       `json:"gps_device_type"      db:"gps_device_type"`
	ImmobilizerEnabled bool          `json:"immobilizer_enabled"  db:"immobilizer_enabled"`
	AIS140Compliant    bool          `json:"ais140_compliant"     db:"ais140_compliant"`
	Make               string        `json:"make"                 db:"make"`
	Model              string        `json:"model"                db:"model"`
	Year               int           `json:"year"                 db:"year"`
	Color              string        `json:"color"                db:"color"`
	Odometer           float64       `json:"odometer"             db:"odometer"`
	CurrentDriverID    *uuid.UUID    `json:"current_driver_id"    db:"current_driver_id"`
	SpeedLimit         float64       `json:"speed_limit"          db:"speed_limit"`
	IsBlacklisted      bool          `json:"is_blacklisted"       db:"is_blacklisted"`
	IsHardBlocked      bool          `json:"is_hard_blocked"      db:"is_hard_blocked"` // Prevent usage when compliance=RED
	RiskScore          float64       `json:"risk_score"           db:"risk_score"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	Version   int        `json:"version"    db:"version"`

	// Derived operational fields (calculated at runtime)
	Summary       *VehicleSummary    `json:"summary,omitempty"`
	CurrentDriver *DriverInfo        `json:"current_driver,omitempty"`
	Documents     []*VehicleDocument `json:"documents,omitempty"`
}

// Security & Masking Helpers
func (v *Vehicle) MaskedRegistration() string {
	if len(v.RegistrationNumber) < 4 {
		return v.RegistrationNumber
	}
	return v.RegistrationNumber[:2] + "****" + v.RegistrationNumber[len(v.RegistrationNumber)-2:]
}

func (v *Vehicle) MaskedVIN() string {
	if v.VINNumber == nil || *v.VINNumber == "" {
		return ""
	}
	vin := *v.VINNumber
	if len(vin) < 8 {
		return "****"
	}
	return vin[:4] + strings.Repeat("*", len(vin)-8) + vin[len(vin)-4:]
}

var (
	reRegistration = regexp.MustCompile(`^[A-Z]{2}[0-9]{1,2}[A-Z]{1,3}[0-9]{4}$`) // Generic Indian Registration Format
	reVIN          = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)
	reChassis      = regexp.MustCompile(`^[A-Z0-9]{10,25}$`)
)

// Validate performs domain-level integrity checks with cross-field enforcement.
func (v *Vehicle) Validate() error {
	// 1. Enum Validation
	if !v.Status.IsValid() {
		return fmt.Errorf("invalid vehicle status: %s", v.Status)
	}
	if !v.VehicleType.IsValid() {
		return fmt.Errorf("invalid vehicle type: %s", v.VehicleType)
	}
	if !v.FuelType.IsValid() {
		return fmt.Errorf("invalid fuel type: %s", v.FuelType)
	}
	if !v.OwnershipType.IsValid() {
		return fmt.Errorf("invalid ownership type: %s", v.OwnershipType)
	}

	// 2. Basic Identity
	if !reRegistration.MatchString(v.RegistrationNumber) {
		return fmt.Errorf("invalid registration format: expected standard Indian regional format (e.g., MH12AB1234)")
	}
	if v.VINNumber != nil && *v.VINNumber != "" && !reVIN.MatchString(*v.VINNumber) {
		return fmt.Errorf("invalid VIN format: must be 17 characters (excluding I, O, Q)")
	}
	if v.ChassisNumber != nil && *v.ChassisNumber != "" && !reChassis.MatchString(*v.ChassisNumber) {
		return fmt.Errorf("invalid chassis number format: alphanumeric 10-25 chars")
	}

	// 3. Document Integrity (Minor Gaps Fix)
	typeSeen := make(map[DocumentType]bool)
	for _, doc := range v.Documents {
		if doc.DeletedAt != nil { continue }
		if typeSeen[doc.Type] {
			return fmt.Errorf("duplicate active document type detected: %s", doc.Type)
		}
		typeSeen[doc.Type] = true
	}

	// 4. Lifecycle & Integrity
	if v.Year < 1900 || v.Year > time.Now().Year()+1 {
		return fmt.Errorf("invalid vehicle year: must be between 1900 and next year")
	}
	if v.CapacityTonnage <= 0 {
		return fmt.Errorf("capacity must be greater than zero")
	}
	if v.FuelTankCapacity <= 0 && v.FuelType != FuelElectric {
		return fmt.Errorf("fuel tank capacity must be greater than zero for non-EV assets")
	}

	// 5. Operational Cross-Checks & Compliance Hardening
	if v.AIS140Compliant && !v.GPSEnabled {
		return fmt.Errorf("AIS140 compliance requires GPS to be enabled")
	}
	if v.IsBlacklisted && v.Status != StatusSuspended {
		return fmt.Errorf("blacklisted vehicles must be in SUSPENDED status")
	}
	if v.IsHardBlocked && v.Status != StatusSuspended {
		return fmt.Errorf("hard-blocked vehicles must be in SUSPENDED status to prevent unauthorized usage")
	}
	if v.GPSEnabled && (v.GPSDeviceID == nil || *v.GPSDeviceID == "") {
		return fmt.Errorf("GPS device ID is required when GPS is enabled")
	}

	// 6. Thresholds
	if v.SpeedLimit < 0 || v.SpeedLimit > 200 {
		return fmt.Errorf("speed limit must be between 0 and 200 km/h")
	}
	if v.Odometer < 0 {
		return fmt.Errorf("odometer cannot be negative")
	}

	return nil
}

// ComplianceBreach captures specific regulatory failure details.
type ComplianceBreach struct {
	Type       DocumentType `json:"type"`
	ExpiryDate time.Time    `json:"expiry_date"`
	Status     string       `json:"status"`
}

// CheckCompliance results (Functional return to avoid silent mutation)
type ComplianceResult struct {
	IsHardBlocked bool
	Breaches      []ComplianceBreach
	Changed       bool
}

// EvaluateCompliance evaluates documents and returns a result set without mutating state.
func (v *Vehicle) EvaluateCompliance() ComplianceResult {
	now := time.Now()
	res := ComplianceResult{}
	
	for _, doc := range v.Documents {
		// Only consider active, approved documents for compliance (Minor Gaps Fix)
		if doc.DeletedAt != nil { continue }
		if doc.Status != VerificationApproved { continue }

		if doc.ExpiryDate.Before(now) {
			res.Breaches = append(res.Breaches, ComplianceBreach{
				Type:       doc.Type,
				ExpiryDate: doc.ExpiryDate,
				Status:     "EXPIRED",
			})
			res.IsHardBlocked = true
		}
	}

	// Check if this results in a change compared to current state
	if res.IsHardBlocked != v.IsHardBlocked {
		res.Changed = true
	} else if res.IsHardBlocked && v.Status != StatusSuspended {
		res.Changed = true
	}

	return res
}

// VehicleSummary provides a high-level snapshot of vehicle health and compliance.
type VehicleSummary struct {
	ComplianceStatus      ComplianceState     `json:"compliance_status"`
	ComplianceScore       int                 `json:"compliance_score"`        // 0-100 derived metric
	ComplianceBreachCount int                 `json:"compliance_breach_count"` // Historical breaches
	GPSStatus             GPSStatus           `json:"gps_status"`
	DeviceHealthStatus    DeviceHealthStatus  `json:"device_health_status"`
	IgnitionStatus        IgnitionStatus      `json:"ignition_status"`
	LastSeen              *time.Time          `json:"last_seen"`
	LastOdometerUpdate    *time.Time          `json:"last_odometer_update"`
	TelemetryDelayMs      int64               `json:"telemetry_delay_ms"`
	DataFreshnessScore    int                 `json:"data_freshness_score"` // 0-100 score based on delay
	PacketLossPercentage  float64             `json:"packet_loss_percentage"`
	TelemetryReliability  float64             `json:"telemetry_reliability"`
	AlertCount            int                 `json:"alert_count"`
	CriticalAlertCount    int                 `json:"critical_alert_count"`
	OverSpeedCount        int                 `json:"over_speed_count"`
	GeoFenceViolation     bool                `json:"geofence_violation"`
	ActiveTrip            bool                `json:"active_trip"`
	LastTripEndTime       *time.Time          `json:"last_trip_end_time"`
	NextExpiryDate        *time.Time          `json:"next_expiry_date"`
	NextExpiryType        DocumentType        `json:"next_expiry_type"`
	NextServiceDueDate    *time.Time          `json:"next_service_due_date"`
	LastServiceDate       *time.Time          `json:"last_service_date"`
	MaintenanceStatus     MaintenanceStatus   `json:"maintenance_status"`
	UtilizationScore      float64             `json:"utilization_score"`
	FuelEfficiency30d     float64             `json:"fuel_efficiency_30d"`
	IdleTimeMinutes       int                 `json:"idle_time_minutes"`
	DriverAssigned        bool                `json:"driver_assigned"`
	DriverLicenseStatus   DriverLicenseStatus `json:"driver_license_status"`
	RiskScore             float64             `json:"risk_score"`
	RiskCategory          RiskCategory        `json:"risk_category"`
	LastLocationLat       float64             `json:"last_location_lat"`
	LastLocationLng       float64             `json:"last_location_lng"`
	LastPingIP            string              `json:"last_ping_ip"`
	NetworkType           NetworkType         `json:"network_type"`
	IsHardBlocked         bool                `json:"is_hard_blocked"`
}

// GetDataFreshnessScore calculates a score based on telemetry delay.
func GetDataFreshnessScore(delayMs int64) int {
	if delayMs < 0 {
		return 0
	}
	if delayMs < 5000 {
		return 100
	} // < 5s
	if delayMs < 30000 {
		return 90
	} // < 30s
	if delayMs < 60000 {
		return 70
	} // < 1m
	if delayMs < 300000 {
		return 40
	} // < 5m
	return 0
}

// GetRiskCategory derives a categorical level from a numerical score.
func GetRiskCategory(score float64) RiskCategory {
	if score < 30 {
		return RiskLow
	}
	if score < 70 {
		return RiskMedium
	}
	return RiskHigh
}

// VehicleSearchParams holds query params for GET /vehicles/search.
type VehicleSearchParams struct {
	TenantID    *uuid.UUID
	Query       string
	Status      VehicleStatus
	VehicleType VehicleType
	FuelType    FuelType
	BranchID    *uuid.UUID
	GPSEnabled  *bool
	Page        int
	Limit       int
	ParentVehicleID *uuid.UUID `json:"parent_vehicle_id"`
}

// CreateVehicleRequest represents the payload for creating a new vehicle.
type CreateVehicleRequest struct {
	RegistrationNumber string     `json:"registration_number"`
	VehicleType        string     `json:"vehicle_type"`
	FuelType           string     `json:"fuel_type"`
	OwnershipType      string     `json:"ownership_type"`
	CapacityTonnage    *float64   `json:"capacity_tonnage"`
	FuelTankCapacity   *float64   `json:"fuel_tank_capacity"`
	ChassisNumber      *string    `json:"chassis_number"`
	EngineNumber       *string    `json:"engine_number"`
	VINNumber          *string    `json:"vin_number"`
	ParentVehicleID    *uuid.UUID `json:"parent_vehicle_id"`
	BranchID           *uuid.UUID `json:"branch_id"`
	CostCenterID       *uuid.UUID `json:"cost_center_id"`
	GPSDeviceID        *string    `json:"gps_device_id"`
	GPSEnabled         bool       `json:"gps_enabled"`
	GPSDeviceType      *string    `json:"gps_device_type"`
	ImmobilizerEnabled bool       `json:"immobilizer_enabled"`
	AIS140Compliant    bool       `json:"ais140_compliant"`
	Make               *string    `json:"make"`
	Model              *string    `json:"model"`
	Year               *int       `json:"year"`
	Color              *string    `json:"color"`
	Odometer           *int64     `json:"odometer"`
	CurrentDriverID    *uuid.UUID `json:"current_driver_id"`
	SpeedLimit         float64    `json:"speed_limit"`
	LicensePlate       *string    `json:"license_plate"` // Legacy
}

// UpdateVehicleRequest represents the payload for updating an existing vehicle.
type UpdateVehicleRequest struct {
	VehicleType        *string    `json:"vehicle_type"`
	FuelType           *string    `json:"fuel_type"`
	OwnershipType      *string    `json:"ownership_type"`
	CapacityTonnage    *float64   `json:"capacity_tonnage"`
	FuelTankCapacity   *float64   `json:"fuel_tank_capacity"`
	ChassisNumber      *string    `json:"chassis_number"`
	EngineNumber       *string    `json:"engine_number"`
	VINNumber          *string    `json:"vin_number"`
	BranchID           *uuid.UUID `json:"branch_id"`
	CostCenterID       *uuid.UUID `json:"cost_center_id"`
	GPSDeviceID        *string    `json:"gps_device_id"`
	GPSEnabled         *bool      `json:"gps_enabled"`
	GPSDeviceType      *string    `json:"gps_device_type"`
	ImmobilizerEnabled *bool      `json:"immobilizer_enabled"`
	AIS140Compliant    *bool      `json:"ais140_compliant"`
	Make               *string    `json:"make"`
	Model              *string    `json:"model"`
	Year               *int       `json:"year"`
	Color              *string    `json:"color"`
	CurrentDriverID    *uuid.UUID `json:"current_driver_id"`
	Status             *string    `json:"status"`
	Odometer           *int64     `json:"odometer"`
	Insurance          *string    `json:"insurance"`
	SpeedLimit         *float64   `json:"speed_limit"`
	ParentVehicleID    *uuid.UUID `json:"parent_vehicle_id"`
}

// VehicleListResponse is the paginated response for vehicle listings.
type VehicleListResponse struct {
	Vehicles []Vehicle `json:"vehicles"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

// VehicleStats provides aggregated counts for the fleet dashboard.
type VehicleStats struct {
	TotalVehicles   int     `json:"total_vehicles"`
	ActiveVehicles  int     `json:"active_vehicles"`
	Maintenance     int     `json:"maintenance"`
	Suspended       int     `json:"suspended"`
	Blacklisted     int     `json:"blacklisted"`
	GPSEnabledCount int     `json:"gps_enabled_count"`
	TotalDistance   float64 `json:"total_distance,omitempty"`
	TotalFuel       float64 `json:"total_fuel,omitempty"`
}


// VehicleLocation represents a coordinate point for map visualization.
type VehicleLocation struct {
	VehicleID uuid.UUID `json:"vehicle_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

// DriverInfo represents basic driver details embedded in other models.
type DriverInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type AssignDriverRequest struct {
	DriverID uuid.UUID `json:"driver_id"`
}

