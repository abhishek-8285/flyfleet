package integrations

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ─── PayloadTransformer ───────────────────────────────────────────────────────
//
// Converts internal FlyFleet event payloads to the schema expected by each external system.
//
// Supported engines:
//   - JSONPathTransformer: dot-notation field mapping (JSONPath-lite)
//
// Example mapping:
//   "payload.driver.name" → "external.driverName"
//   "payload.trip.distance" → "external.distanceKm"
//
// Production upgrade: integrate JQ (jq-go) or CEL (cel-go) for complex expressions.

type PayloadTransformer interface {
	Transform(input interface{}, mapping map[string]string) (interface{}, error)
	ValidateSchema(input interface{}, requiredFields []string) error
}

// ─── JSONPath Transformer ─────────────────────────────────────────────────────

type JSONPathTransformer struct{}

// Transform applies dot-notation field mappings to the input object.
// mapping[source_path] = destination_path
//
//	e.g. {"payload.driver.name": "external.driverName"}
// ValidateSchema checks if the input object contains all required dot-notation fields.
// (Gap: Transformation Validation / Schema Validation)
func (t *JSONPathTransformer) ValidateSchema(input interface{}, requiredFields []string) error {
	raw, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("transformer validate: marshal input failed: %w", err)
	}
	var src map[string]interface{}
	if err := json.Unmarshal(raw, &src); err != nil {
		return fmt.Errorf("transformer validate: unmarshal input failed: %w", err)
	}

	var missing []string
	for _, reqField := range requiredFields {
		_, found := getPath(src, strings.Split(reqField, "."))
		if !found {
			missing = append(missing, reqField)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("schema validation failed, missing required fields: %v", missing)
	}
	return nil
}

func (t *JSONPathTransformer) Transform(input interface{}, mapping map[string]string) (interface{}, error) {
	// Normalize input to map[string]interface{} via JSON round-trip
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("transformer: marshal input: %w", err)
	}
	var src map[string]interface{}
	if err := json.Unmarshal(raw, &src); err != nil {
		return nil, fmt.Errorf("transformer: unmarshal input: %w", err)
	}

	dst := make(map[string]interface{})
	for srcPath, dstPath := range mapping {
		val, found := getPath(src, strings.Split(srcPath, "."))
		if !found {
			continue // missing source field: skip rather than error
		}
		setPath(dst, strings.Split(dstPath, "."), val)
	}
	return dst, nil
}

// getPath traverses a map following dot-notation path segments.
func getPath(m map[string]interface{}, parts []string) (interface{}, bool) {
	if len(parts) == 0 {
		return m, true
	}
	val, ok := m[parts[0]]
	if !ok {
		return nil, false
	}
	if len(parts) == 1 {
		return val, true
	}
	nested, ok := val.(map[string]interface{})
	if !ok {
		return nil, false
	}
	return getPath(nested, parts[1:])
}

// setPath sets a value at the given dot-notation path in a map, creating intermediate maps.
func setPath(m map[string]interface{}, parts []string, val interface{}) {
	if len(parts) == 1 {
		m[parts[0]] = val
		return
	}
	nested, ok := m[parts[0]].(map[string]interface{})
	if !ok {
		nested = make(map[string]interface{})
		m[parts[0]] = nested
	}
	setPath(nested, parts[1:], val)
}

// ─── Common ERP Field Mappings ────────────────────────────────────────────────

// SAPFieldMapping maps FlyFleet trip completion events to SAP IDOC fields.
var SAPFieldMapping = map[string]string{
	"event_type":              "IDOC.DOCTYP",
	"tenant_id":               "IDOC.SNDPRN",
	"payload.trip_id":         "IDOC.OBJKEY",
	"payload.driver.name":     "IDOC.DRIVER_NAME",
	"payload.trip.distance":   "IDOC.DIST_KM",
	"payload.trip.start_time": "IDOC.TRIP_START",
	"payload.trip.end_time":   "IDOC.TRIP_END",
}

// SalesforceFieldMapping maps FlyFleet events to Salesforce object fields.
var SalesforceFieldMapping = map[string]string{
	"payload.driver.name":   "Contact.Name",
	"payload.trip.distance": "Trip__c.Distance_Km__c",
	"payload.vehicle.plate": "Vehicle__c.License_Plate__c",
	"payload.event_time":    "Trip__c.Event_Time__c",
}
