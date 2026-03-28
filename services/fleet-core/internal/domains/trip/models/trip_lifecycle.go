package models

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// TripEventImmutable represents an immutable, hash-chained trip event
type TripEventImmutable struct {
	ID        string          `json:"id" db:"id"`
	TripID    string          `json:"trip_id" db:"trip_id"`
	EventType string          `json:"event_type" db:"event_type"`
	Payload   json.RawMessage `json:"payload" db:"payload"`
	Hash      string          `json:"hash" db:"hash"`
	PrevHash  string          `json:"prev_hash" db:"prev_hash"`
	Timestamp time.Time       `json:"timestamp" db:"timestamp"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

// CalculateHash computes SHA-256 hash of the event
func (e *TripEventImmutable) CalculateHash() string {
	data := struct {
		TripID    string          `json:"trip_id"`
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
		PrevHash  string          `json:"prev_hash"`
		Timestamp time.Time       `json:"timestamp"`
	}{
		TripID:    e.TripID,
		EventType: e.EventType,
		Payload:   e.Payload,
		PrevHash:  e.PrevHash,
		Timestamp: e.Timestamp,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// VerifyHashChain verifies the integrity of the event chain
func (e *TripEventImmutable) VerifyHashChain(prevEvent *TripEventImmutable) bool {
	expectedPrevHash := ""
	if prevEvent != nil {
		expectedPrevHash = prevEvent.Hash
	}
	return e.PrevHash == expectedPrevHash && e.Hash == e.CalculateHash()
}

// TripEventPayload represents structured event payload
type TripEventPayload struct {
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	ActorID   string                 `json:"actor_id"`
	ActorType string                 `json:"actor_type"` // driver, system, admin, device
	Location  *Location              `json:"location,omitempty"`
	Odometer  *float64               `json:"odometer,omitempty"`
	FuelLevel *float64               `json:"fuel_level,omitempty"`
	Speed     *float64               `json:"speed,omitempty"`
	Reason    string                 `json:"reason,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	DeviceID  string                 `json:"device_id,omitempty"`
	Signature string                 `json:"signature,omitempty"` // Device signature for AIS-140
}

// TripMetrics represents real-time trip metrics
type TripMetrics struct {
	TripID              string    `json:"trip_id"`
	CurrentStatus       string    `json:"current_status"`
	DistanceCovered     float64   `json:"distance_covered"`
	DistanceRemaining   float64   `json:"distance_remaining"`
	ETAArrival          time.Time `json:"eta_arrival"`
	ETADelay            int       `json:"eta_delay_minutes"`
	AverageSpeed        float64   `json:"average_speed"`
	MaxSpeed            float64   `json:"max_speed"`
	OverspeedEvents     int       `json:"overspeed_events"`
	FuelConsumed        float64   `json:"fuel_consumed"`
	FuelEfficiency      float64   `json:"fuel_efficiency_kmpl"`
	IdlingTime          int       `json:"idling_time_minutes"`
	HaltTime            int       `json:"halt_time_minutes"`
	GeoDeviation        bool      `json:"geo_deviation"`
	GeoDeviationMeters  float64   `json:"geo_deviation_meters"`
	LastTelemetryAt     time.Time `json:"last_telemetry_at"`
	TelemetryLagSeconds int       `json:"telemetry_lag_seconds"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// TripFinancials represents financial calculations for a completed trip
type TripFinancials struct {
	TripID           string    `json:"trip_id"`
	BillableDistance float64   `json:"billable_distance_km"`
	TollCost         float64   `json:"toll_cost_inr"`
	FuelCost         float64   `json:"fuel_cost_inr"`
	DriverAllowance  float64   `json:"driver_allowance_inr"`
	SLAPenalty       float64   `json:"sla_penalty_inr"`
	Subtotal         float64   `json:"subtotal_inr"`
	TotalAmount      float64   `json:"total_amount_inr"`
	GSTAmount        float64   `json:"gst_amount_inr"`
	GSTRate          float64   `json:"gst_rate_percent"`
	CGSTAmount       float64   `json:"cgst_amount_inr"`
	SGSTAmount       float64   `json:"sgst_amount_inr"`
	IGSTAmount       float64   `json:"igst_amount_inr"`
	TDSApplicable    bool      `json:"tds_applicable"`
	TDSAmount        float64   `json:"tds_amount_inr"`
	ReverseCharge    bool      `json:"reverse_charge"`
	PlaceOfSupply    string    `json:"place_of_supply"`
	InvoiceGenerated bool      `json:"invoice_generated"`
	InvoiceID        string    `json:"invoice_id"`
	EWayBillNo       string    `json:"eway_bill_no"`
	CalculatedAt     time.Time `json:"calculated_at"`
}

// TripCompliance represents compliance status for a trip
type TripCompliance struct {
	TripID              string     `json:"trip_id"`
	PermitType          string     `json:"permit_type"`
	PermitValidUntil    *time.Time `json:"permit_valid_until"`
	PermitVerified      bool       `json:"permit_verified"`
	PUCValid            bool       `json:"puc_valid"`
	PUCValidUntil       *time.Time `json:"puc_valid_until"`
	InsuranceValid      bool       `json:"insurance_valid"`
	InsuranceValidUntil *time.Time `json:"insurance_valid_until"`
	FitnessValid        bool       `json:"fitness_valid"`
	FitnessValidUntil   *time.Time `json:"fitness_valid_until"`
	EWayBillNo          string     `json:"eway_bill_no"`
	EWayBillValid       bool       `json:"eway_bill_valid"`
	EWayBillValidUntil  *time.Time `json:"eway_bill_valid_until"`
	GSTClassification   string     `json:"gst_classification"`
	PlaceOfSupply       string     `json:"place_of_supply"`
	Interstate          bool       `json:"interstate"`
	AIS140Compliant     bool       `json:"ais140_compliant"`
	HOSCompliant        bool       `json:"hos_compliant"`
	HOSViolations       []string   `json:"hos_violations"`
	FastagLinked        bool       `json:"fastag_linked"`
	FastagBalance       float64    `json:"fastag_balance"`
	VerifiedAt          time.Time  `json:"verified_at"`
	VerifiedBy          string     `json:"verified_by"`
}

// TripSearchFilter represents search filters for trips
type TripSearchFilter struct {
	OrgID         string    `json:"org_id"`
	VehicleID     string    `json:"vehicle_id"`
	DriverID      string    `json:"driver_id"`
	Status        []string  `json:"status"`
	DateFrom      time.Time `json:"date_from"`
	DateTo        time.Time `json:"date_to"`
	AIS140Only    bool      `json:"ais140_only"`
	Interstate    *bool     `json:"interstate"`
	HasViolations *bool     `json:"has_violations"`
	HasEWayBill   *bool     `json:"has_eway_bill"`
	Page          int       `json:"page"`
	PageSize      int       `json:"page_size"`
	OrderBy       string    `json:"order_by"`
	OrderDir      string    `json:"order_dir"`
}

// TripSearchResult represents paginated search results
type TripSearchResult struct {
	Trips      []*Trip `json:"trips"`
	Total      int64   `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}

// ValidStateTransitions defines allowed state transitions
var ValidStateTransitions = map[string][]string{
	TripStatusCreated:   {TripStatusScheduled, TripStatusCancelled},
	TripStatusScheduled: {TripStatusStarted, TripStatusCancelled},
	TripStatusStarted:   {TripStatusPaused, TripStatusCompleted, TripStatusFailed},
	TripStatusPaused:    {TripStatusResumed, TripStatusFailed},
	TripStatusResumed:   {TripStatusPaused, TripStatusCompleted, TripStatusFailed},
	TripStatusCompleted: {}, // Terminal state
	TripStatusCancelled: {}, // Terminal state
	TripStatusFailed:    {}, // Terminal state
}

// IsTerminalState checks if a state is terminal
func IsTerminalState(status string) bool {
	switch status {
	case TripStatusCompleted, TripStatusCancelled, TripStatusFailed:
		return true
	default:
		return false
	}
}

// IsValidTransition checks if a state transition is valid
func IsValidTransition(fromStatus, toStatus string) bool {
	allowedTransitions, exists := ValidStateTransitions[fromStatus]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == toStatus {
			return true
		}
	}
	return false
}

// VerifyEventChain verifies a chain of immutable events
func VerifyEventChain(events []*TripEventImmutable) bool {
	if len(events) == 0 {
		return true
	}

	// First event should have empty prev_hash
	if events[0].PrevHash != "" {
		return false
	}

	// Verify each event's hash and chain
	for i, event := range events {
		// Verify hash
		if event.Hash != event.CalculateHash() {
			return false
		}

		// Verify chain (except first event)
		if i > 0 {
			if event.PrevHash != events[i-1].Hash {
				return false
			}
		}
	}

	return true
}

// CalculateEventHash calculates hash for event chaining
func CalculateEventHash(tripID, eventType string, payload interface{}, prevHash string, timestamp time.Time) string {
	data := struct {
		TripID    string      `json:"trip_id"`
		EventType string      `json:"event_type"`
		Payload   interface{} `json:"payload"`
		PrevHash  string      `json:"prev_hash"`
		Timestamp time.Time   `json:"timestamp"`
	}{
		TripID:    tripID,
		EventType: eventType,
		Payload:   payload,
		PrevHash:  prevHash,
		Timestamp: timestamp,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}
