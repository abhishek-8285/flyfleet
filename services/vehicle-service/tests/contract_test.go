const (
	TestLimitMax = 100
	TestTelemetryLimitMax = 500
)

package contract_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

var specDoc *openapi3.T

func TestMain(m *testing.M) {
	// Load OpenAPI spec
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	var err error
	specDoc, err = loader.LoadFromFile("../api/openapi/trips.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load OpenAPI spec: %v\n", err)
		os.Exit(1)
	}

	// Clear servers so router matches paths directly (without base URL)
	specDoc.Servers = nil

	// Validate spec itself
	if err := specDoc.Validate(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "OpenAPI spec is invalid: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("OpenAPI spec loaded and validated successfully")
	os.Exit(m.Run())
}

// ========================
// Schema Validation Tests
// ========================

func TestSpec_IsValid(t *testing.T) {
	if specDoc == nil {
		t.Fatal("Spec was not loaded")
	}

	// Check all expected paths exist
	expectedPaths := []string{"/trips", "/trips/{id}", "/trips/{id}/telemetry", "/trips/settings"}
	for _, p := range expectedPaths {
		if specDoc.Paths.Find(p) == nil {
			t.Errorf("Expected path %s not found in spec", p)
		}
	}
}

func TestSpec_TripEndpoints(t *testing.T) {
	tripsPath := specDoc.Paths.Find("/trips")
	if tripsPath == nil {
		t.Fatal("/trips path not found")
	}

	// Check GET /trips
	if tripsPath.Get == nil {
		t.Error("GET /trips operation missing")
	}
	if tripsPath.Get.OperationID != "listTrips" {
		t.Errorf("GET /trips operationId mismatch: got %s", tripsPath.Get.OperationID)
	}

	// Check POST /trips
	if tripsPath.Post == nil {
		t.Error("POST /trips operation missing")
	}
	if tripsPath.Post.OperationID != "createTrip" {
		t.Errorf("POST /trips operationId mismatch: got %s", tripsPath.Post.OperationID)
	}
}

func TestSpec_TripIdEndpoints(t *testing.T) {
	tripIdPath := specDoc.Paths.Find("/trips/{id}")
	if tripIdPath == nil {
		t.Fatal("/trips/{id} path not found")
	}

	if tripIdPath.Get == nil {
		t.Error("GET /trips/{id} operation missing")
	}
	if tripIdPath.Put == nil {
		t.Error("PUT /trips/{id} operation missing")
	}
	if tripIdPath.Delete == nil {
		t.Error("DELETE /trips/{id} operation missing")
	}
}

func TestSpec_TelemetryEndpoint(t *testing.T) {
	telemetryPath := specDoc.Paths.Find("/trips/{id}/telemetry")
	if telemetryPath == nil {
		t.Fatal("/trips/{id}/telemetry path not found")
	}

	if telemetryPath.Get == nil {
		t.Error("GET /trips/{id}/telemetry operation missing")
	}
}

func TestSpec_SettingsEndpoints(t *testing.T) {
	settingsPath := specDoc.Paths.Find("/trips/settings")
	if settingsPath == nil {
		t.Fatal("/trips/settings path not found")
	}

	if settingsPath.Get == nil {
		t.Error("GET /trips/settings operation missing")
	}
	if settingsPath.Put == nil {
		t.Error("PUT /trips/settings operation missing")
	}
}

// ========================
// Request Body Schema Tests
// ========================

func TestSpec_CreateTripRequest_Schema(t *testing.T) {
	schema := specDoc.Components.Schemas["CreateTripRequest"]
	if schema == nil {
		t.Fatal("CreateTripRequest schema not found")
	}

	schemaVal := schema.Value

	// Check required fields
	required := schemaVal.Required
	expectedRequired := []string{"vehicle_id", "start_location", "end_location", "start_time"}
	for _, field := range expectedRequired {
		found := false
		for _, r := range required {
			if r == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be required in CreateTripRequest", field)
		}
	}

	// Check vehicle_id is UUID format
	vehicleIDProp := schemaVal.Properties["vehicle_id"]
	if vehicleIDProp == nil || vehicleIDProp.Value.Format != "uuid" {
		t.Error("vehicle_id should have format: uuid")
	}
}

func TestSpec_UpdateTripRequest_Schema(t *testing.T) {
	schema := specDoc.Components.Schemas["UpdateTripRequest"]
	if schema == nil {
		t.Fatal("UpdateTripRequest schema not found")
	}

	schemaVal := schema.Value

	// UpdateTripRequest should have no required fields (all optional)
	if len(schemaVal.Required) > 0 {
		t.Error("UpdateTripRequest should have no required fields")
	}

	// Check status is enum
	statusProp := schemaVal.Properties["status"]
	if statusProp != nil && statusProp.Value.Enum == nil {
		t.Error("status should reference TripStatus enum")
	}
}

func TestSpec_TripStatus_Enum(t *testing.T) {
	schema := specDoc.Components.Schemas["TripStatus"]
	if schema == nil {
		t.Fatal("TripStatus schema not found")
	}

	expectedEnums := []string{"scheduled", "in_progress", "completed", "cancelled"}
	actualEnums := schema.Value.Enum

	if len(actualEnums) != len(expectedEnums) {
		t.Errorf("Expected %d enum values, got %d", len(expectedEnums), len(actualEnums))
	}
}

// ========================
// Response Schema Tests
// ========================

func TestSpec_TripListResponse_Schema(t *testing.T) {
	schema := specDoc.Components.Schemas["TripListResponse"]
	if schema == nil {
		t.Fatal("TripListResponse schema not found")
	}

	schemaVal := schema.Value

	// Check required fields
	required := schemaVal.Required
	expectedRequired := []string{"trips", "total", "page", "limit"}
	for _, field := range expectedRequired {
		found := false
		for _, r := range required {
			if r == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be required in TripListResponse", field)
		}
	}
}

func TestSpec_ErrorResponse_Schema(t *testing.T) {
	schema := specDoc.Components.Schemas["Error"]
	if schema == nil {
		t.Fatal("Error schema not found")
	}

	schemaVal := schema.Value

	// error field must be required
	found := false
	for _, r := range schemaVal.Required {
		if r == "error" {
			found = true
			break
		}
	}
	if !found {
		t.Error("error field must be required in Error schema")
	}
}

func TestSpec_ValidationErrorResponse_Schema(t *testing.T) {
	schema := specDoc.Components.Schemas["ValidationError"]
	if schema == nil {
		t.Fatal("ValidationError schema not found")
	}

	schemaVal := schema.Value

	// Both error and validation_errors must be required
	requiredFields := []string{"error", "validation_errors"}
	for _, field := range requiredFields {
		found := false
		for _, r := range schemaVal.Required {
			if r == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be required in ValidationError", field)
		}
	}
}

// ========================
// Parameter Validation Tests
// ========================

func TestSpec_ListTrips_Parameters(t *testing.T) {
	tripsPath := specDoc.Paths.Find("/trips")
	if tripsPath == nil || tripsPath.Get == nil {
		t.Fatal("GET /trips not found")
	}

	params := tripsPath.Get.Parameters

	// Check limit parameter constraints
	for _, param := range params {
		if param.Value.Name == "limit" {
			schema := param.Value.Schema.Value
			if schema.Min == nil || *schema.Min != 1 {
				t.Error("limit should have minimum: 1")
			}
			if schema.Max == nil || *schema.Max != TestLimitMax {
				t.Error("limit should have maximum: 100")
			}
		}
		if param.Value.Name == "page" {
			schema := param.Value.Schema.Value
			if schema.Min == nil || *schema.Min != 1 {
				t.Error("page should have minimum: 1")
			}
		}
	}
}

func TestSpec_GetTelemetry_LimitParameter(t *testing.T) {
	telemetryPath := specDoc.Paths.Find("/trips/{id}/telemetry")
	if telemetryPath == nil || telemetryPath.Get == nil {
		t.Fatal("GET /trips/{id}/telemetry not found")
	}

	for _, param := range telemetryPath.Get.Parameters {
		if param.Value.Name == "limit" {
			schema := param.Value.Schema.Value
			if schema.Max == nil || *schema.Max != TestTelemetryLimitMax {
				t.Error("telemetry limit should have maximum: 500")
			}
		}
	}
}

// ========================
// Request Validation Integration Tests
// ========================

func TestRequestValidation_CreateTrip_ValidBody(t *testing.T) {
	router, err := legacy.NewRouter(specDoc)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	body := map[string]interface{}{
		"vehicle_id":     "550e8400-e29b-41d4-a716-446655440000",
		"start_location": "New York, NY",
		"end_location":   "Boston, MA",
		"start_time":     "2026-02-08T10:00:00Z",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/trips", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("Route not found: %v", err)
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}

	err = openapi3filter.ValidateRequest(context.Background(), validationInput)
	if err != nil {
		t.Errorf("Valid request failed validation: %v", err)
	}
}

func TestRequestValidation_CreateTrip_MissingRequiredField(t *testing.T) {
	router, err := legacy.NewRouter(specDoc)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	// Missing vehicle_id
	body := map[string]interface{}{
		"start_location": "New York, NY",
		"end_location":   "Boston, MA",
		"start_time":     "2026-02-08T10:00:00Z",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/trips", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("Route not found: %v", err)
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	err = openapi3filter.ValidateRequest(context.Background(), validationInput)
	if err == nil {
		t.Error("Expected validation error for missing vehicle_id, got none")
	}
}

func TestRequestValidation_CreateTrip_InvalidUUID(t *testing.T) {
	router, err := legacy.NewRouter(specDoc)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	body := map[string]interface{}{
		"vehicle_id":     "not-a-uuid",
		"start_location": "New York, NY",
		"end_location":   "Boston, MA",
		"start_time":     "2026-02-08T10:00:00Z",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/trips", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("Route not found: %v", err)
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	err = openapi3filter.ValidateRequest(context.Background(), validationInput)
	if err == nil {
		t.Error("Expected validation error for invalid UUID, got none")
	}
}

func TestRequestValidation_UpdateTrip_InvalidStatus(t *testing.T) {
	router, err := legacy.NewRouter(specDoc)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	body := map[string]interface{}{
		"status": "invalid_status",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/trips/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("Route not found: %v", err)
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	err = openapi3filter.ValidateRequest(context.Background(), validationInput)
	if err == nil {
		t.Error("Expected validation error for invalid status, got none")
	}
}

// ========================
// Response Validation Integration Tests
// ========================

func TestResponseValidation_TripListResponse(t *testing.T) {
	router, err := legacy.NewRouter(specDoc)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("Route not found: %v", err)
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	// Valid response
	response := map[string]interface{}{
		"trips": []map[string]interface{}{
			{
				"id":              "550e8400-e29b-41d4-a716-446655440000",
				"trip_id":         "TR-2026-0001",
				"vehicle_id":      "550e8400-e29b-41d4-a716-446655440001",
				"start_location":  "New York",
				"end_location":    "Boston",
				"start_time":      "2026-02-08T10:00:00Z",
				"distance":        150.5,
				"status":          "completed",
				"organization_id": "550e8400-e29b-41d4-a716-446655440002",
				"created_at":      "2026-02-08T09:00:00Z",
				"updated_at":      "2026-02-08T12:00:00Z",
			},
		},
		"total": 1,
		"page":  1,
		"limit": 20,
	}
	bodyBytes, _ := json.Marshal(response)

	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: validationInput,
		Status:                 200,
		Header:                 http.Header{"Content-Type": []string{"application/json"}},
		Body:                   io.NopCloser(bytes.NewReader(bodyBytes)),
	}

	err = openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
	if err != nil {
		t.Errorf("Valid response failed validation: %v", err)
	}
}

func TestResponseValidation_InvalidTripListResponse_MissingTotal(t *testing.T) {
	router, err := legacy.NewRouter(specDoc)
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/trips", nil)
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		t.Fatalf("Route not found: %v", err)
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	// Invalid response - missing required 'total' field
	response := map[string]interface{}{
		"trips": []map[string]interface{}{},
		"page":  1,
		"limit": 20,
	}
	bodyBytes, _ := json.Marshal(response)

	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: validationInput,
		Status:                 200,
		Header:                 http.Header{"Content-Type": []string{"application/json"}},
		Body:                   io.NopCloser(bytes.NewReader(bodyBytes)),
	}

	err = openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
	if err == nil {
		t.Error("Expected validation error for missing 'total' field, got none")
	}
}
