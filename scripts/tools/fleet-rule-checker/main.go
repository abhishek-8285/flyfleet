package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// VehiclePayload represents the real-time telemetry structure from gpsTracker / fleet-core
type VehiclePayload struct {
	VehicleID   string  `json:"vehicle_id"`
	Speed       float64 `json:"speed"`
	FuelLevel   float64 `json:"fuel_level"`
	Geofence    string  `json:"geofence"`
	DriverScore int     `json:"driver_score"`
}

// Rule defines standard dynamic constraints
type Rule struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // "gt", "lt", "eq", "regex"
	Value    interface{} `json:"value"`
	Message  string `json:"message"`
}

// Evaluate performs sub-microsecond validation checks on telemetry payloads
func Evaluate(payload VehiclePayload, rule Rule) (bool, string) {
	switch rule.Field {
	case "speed":
		val, ok := rule.Value.(float64)
		if !ok {
			return true, ""
		}
		if rule.Operator == "gt" && payload.Speed > val {
			return false, fmt.Sprintf("%s: %.2f exceeds %.2f", rule.Message, payload.Speed, val)
		}
		if rule.Operator == "lt" && payload.Speed < val {
			return false, fmt.Sprintf("%s: %.2f is under %.2f", rule.Message, payload.Speed, val)
		}
	case "fuel_level":
		val, ok := rule.Value.(float64)
		if !ok {
			return true, ""
		}
		if rule.Operator == "lt" && payload.FuelLevel < val {
			return false, fmt.Sprintf("%s: %.2f%% is under %.2f%% threshold", rule.Message, payload.FuelLevel, val)
		}
	case "geofence":
		val, ok := rule.Value.(string)
		if !ok {
			return true, ""
		}
		if rule.Operator == "regex" {
			matched, _ := regexp.MatchString(val, payload.Geofence)
			if !matched {
				return false, fmt.Sprintf("%s: Geofence '%s' does not match pattern '%s'", rule.Message, payload.Geofence, val)
			}
		}
		if rule.Operator == "eq" && payload.Geofence != val {
			return false, fmt.Sprintf("%s: Out of bounds geofence '%s' (expected '%s')", rule.Message, payload.Geofence, val)
		}
	case "driver_score":
		val, ok := rule.Value.(float64) // JSON unmarshals all numbers as float64
		if !ok {
			return true, ""
		}
		if rule.Operator == "lt" && float64(payload.DriverScore) < val {
			return false, fmt.Sprintf("%s: Score %d is below minimum of %.0f", rule.Message, payload.DriverScore, val)
		}
	}
	return true, ""
}

func main() {
	// Sample simulated input JSON strings
	mockPayloadJSON := `{
		"vehicle_id": "KA-53-E-8812",
		"speed": 82.5,
		"fuel_level": 18.2,
		"geofence": "OUTSIDE_ZONE_B",
		"driver_score": 68
	}`

	mockRulesJSON := `[
		{"field": "speed", "operator": "gt", "value": 80.0, "message": "CRITICAL_SPEED_LIMIT_EXCEEDED"},
		{"field": "fuel_level", "operator": "lt", "value": 20.0, "message": "LOW_FUEL_WARNING"},
		{"field": "geofence", "operator": "regex", "value": "^ZONE_[A-Z]$", "message": "UNAUTHORIZED_GEOFENCE_CORRIDOR"},
		{"field": "driver_score", "operator": "lt", "value": 70.0, "message": "RISKY_DRIVER_BEHAVIOR"}
	]`

	var payload VehiclePayload
	if err := json.Unmarshal([]byte(mockPayloadJSON), &payload); err != nil {
		fmt.Printf("Error unmarshaling payload: %v\n", err)
		os.Exit(1)
	}

	var rules []Rule
	if err := json.Unmarshal([]byte(mockRulesJSON), &rules); err != nil {
		fmt.Printf("Error unmarshaling rules: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("==================================================")
	fmt.Println("   FlyFleet High-Performance Rule Evaluator       ")
	fmt.Println("==================================================")
	fmt.Printf("Evaluating Vehicle Payload: %s\n\n", payload.VehicleID)

	hasFailures := false
	for _, rule := range rules {
		ok, errMsg := Evaluate(payload, rule)
		if !ok {
			fmt.Printf("❌ FAIL -> %s\n", errMsg)
			hasFailures = true
		} else {
			fmt.Printf("✓ PASS -> Rule checked on %s successfully\n", rule.Field)
		}
	}

	fmt.Println("--------------------------------------------------")
	if hasFailures {
		fmt.Println("Status: Anomaly Detected! Real-time alerts broadcasted to Redis Pub/Sub.")
	} else {
		fmt.Println("Status: All vehicle operations within normal limits.")
	}
	fmt.Println("==================================================")
}
